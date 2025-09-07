package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"gorm.io/gorm"
)

type HostService struct {
	db        *gorm.DB
	connector *libvirt.Connector
	hub       *ws.Hub
}

func NewHostService(db *gorm.DB, connector *libvirt.Connector, hub *ws.Hub) *HostService {
	return &HostService{
		db:        db,
		connector: connector,
		hub:       hub,
	}
}

func (s *HostService) broadcastUpdate() {
	s.hub.BroadcastMessage([]byte(`{"type": "refresh"}`))
}

// --- Host Management ---

func (s *HostService) GetAllHosts() ([]storage.Host, error) {
	var hosts []storage.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

func (s *HostService) GetHostInfo(hostID string) (*libvirt.HostInfo, error) {
	return s.connector.GetHostInfo(hostID)
}

func (s *HostService) AddHost(host storage.Host) (*storage.Host, error) {
	if err := s.db.Create(&host).Error; err != nil {
		return nil, fmt.Errorf("failed to save host to database: %w", err)
	}

	err := s.connector.AddHost(host)
	if err != nil {
		if delErr := s.db.Delete(&host).Error; delErr != nil {
			log.Printf("CRITICAL: Failed to rollback host creation for %s after connection failure. DB Error: %v", host.ID, delErr)
		}
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}

	s.broadcastUpdate()
	return &host, nil
}

func (s *HostService) RemoveHost(hostID string) error {
	if err := s.connector.RemoveHost(hostID); err != nil {
		log.Printf("Warning: failed to disconnect from host %s during removal, continuing with DB deletion: %v", hostID, err)
	}

	// Also delete VMs associated with this host from our DB
	if err := s.db.Where("host_id = ?", hostID).Delete(&storage.VirtualMachine{}).Error; err != nil {
		log.Printf("Warning: failed to delete VMs for host %s from database: %v", hostID, err)
	}

	if err := s.db.Where("id = ?", hostID).Delete(&storage.Host{}).Error; err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}

	s.broadcastUpdate()
	return nil
}

func (s *HostService) ConnectToAllHosts() {
	hosts, err := s.GetAllHosts()
	if err != nil {
		log.Printf("Error retrieving hosts from database on startup: %v", err)
		return
	}

	for _, host := range hosts {
		log.Printf("Attempting to connect to stored host: %s", host.ID)
		if err := s.connector.AddHost(host); err != nil {
			log.Printf("Failed to connect to host %s (%s) on startup: %v", host.ID, host.URI, err)
		} else {
			// Sync VMs after successful connection
			_, err := s.syncAndListVMs(host.ID)
			if err != nil {
				log.Printf("Failed to sync VMs for host %s on startup: %v", host.ID, err)
			}
		}
	}
}

// --- VM Management ---

// syncAndListVMs is the core function to get VMs from libvirt and sync with the local DB.
func (s *HostService) syncAndListVMs(hostID string) ([]libvirt.VMInfo, error) {
	liveVMs, err := s.connector.ListAllDomains(hostID)
	if err != nil {
		return nil, fmt.Errorf("service failed to list vms for host %s: %w", hostID, err)
	}

	// Sync with DB
	for _, vmInfo := range liveVMs {
		configBytes, err := json.Marshal(vmInfo)
		if err != nil {
			log.Printf("Warning: could not marshal VM info for %s: %v", vmInfo.Name, err)
			continue
		}

		vm := storage.VirtualMachine{
			Name:       vmInfo.Name,
			HostID:     hostID,
			ConfigJSON: string(configBytes),
		}

		// Use FirstOrCreate to either find the existing VM or create a new one.
		// Then update its config.
		if err := s.db.Where(storage.VirtualMachine{Name: vmInfo.Name, HostID: hostID}).Assign(storage.VirtualMachine{ConfigJSON: string(configBytes)}).FirstOrCreate(&vm).Error; err != nil {
			log.Printf("Warning: could not sync VM %s to database: %v", vmInfo.Name, err)
		}
	}

	// Optional: Prune DB entries for VMs that no longer exist on the host
	var liveVMNames []string
	for _, vm := range liveVMs {
		liveVMNames = append(liveVMNames, vm.Name)
	}
	if err := s.db.Where("host_id = ? AND name NOT IN ?", hostID, liveVMNames).Delete(&storage.VirtualMachine{}).Error; err != nil {
		log.Printf("Warning: failed to prune old VMs for host %s: %v", hostID, err)
	}

	return liveVMs, nil
}

// ListVMsFromLibvirt gets the real-time list of VMs directly from libvirt.
func (s *HostService) ListVMsFromLibvirt(hostID string) ([]libvirt.VMInfo, error) {
	return s.syncAndListVMs(hostID)
}

// ListVMsFromDB gets the list of VMs for a host from the local database.
func (s *HostService) ListVMsFromDB(hostID string) ([]storage.VirtualMachine, error) {
	var vms []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&vms).Error; err != nil {
		return nil, err
	}
	return vms, nil
}


func (s *HostService) GetVMStats(hostID, vmName string) (*libvirt.VMStats, error) {
	stats, err := s.connector.GetDomainStats(hostID, vmName)
	if err != nil {
		return nil, fmt.Errorf("service failed to get stats for vm %s on host %s: %w", vmName, hostID, err)
	}
	return stats, nil
}

func (s *HostService) GetVMHardware(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	hardware, err := s.connector.GetDomainHardware(hostID, vmName)
	if err != nil {
		return nil, fmt.Errorf("service failed to get hardware for vm %s on host %s: %w", vmName, hostID, err)
	}
	return hardware, nil
}

func (s *HostService) StartVM(hostID, vmName string) error {
	if err := s.connector.StartDomain(hostID, vmName); err != nil {
		return err
	}
	s.broadcastUpdate()
	return nil
}

func (s *HostService) ShutdownVM(hostID, vmName string) error {
	if err := s.connector.ShutdownDomain(hostID, vmName); err != nil {
		return err
	}
	s.broadcastUpdate()
	return nil
}

func (s *HostService) RebootVM(hostID, vmName string) error {
	if err := s.connector.RebootDomain(hostID, vmName); err != nil {
		return err
	}
	s.broadcastUpdate()
	return nil
}

func (s *HostService) ForceOffVM(hostID, vmName string) error {
	if err := s.connector.DestroyDomain(hostID, vmName); err != nil {
		return err
	}
	s.broadcastUpdate()
	return nil
}

func (s *HostService) ForceResetVM(hostID, vmName string) error {
	if err := s.connector.ResetDomain(hostID, vmName); err != nil {
		return err
	}
	s.broadcastUpdate()
	return nil
}


