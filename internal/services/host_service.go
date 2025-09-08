package services

import (
	"fmt"
	"log"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"gorm.io/gorm"
	lv "libvirt.org/go/libvirt"
)

// VMView is a combination of DB data and live libvirt data for the frontend.
type VMView struct {
	// From DB
	ID          uint   `json:"db_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VCPUCount   uint   `json:"vcpu_count"`
	MemoryBytes uint64 `json:"memory_bytes"`
	IsTemplate  bool   `json:"is_template"`

	// From Libvirt
	State    lv.DomainState       `json:"state"`
	MaxMem   uint64               `json:"max_mem"`
	Memory   uint64               `json:"memory"`
	Vcpu     uint                 `json:"vcpu"`
	CpuTime  uint64               `json:"cpu_time"`
	Uptime   int64                `json:"uptime"`
	Graphics libvirt.GraphicsInfo `json:"graphics"`
}

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

	// Initial sync after adding a host
	if _, err := s.syncAndListVMs(host.ID); err != nil {
		log.Printf("Warning: failed to sync VMs on initial add for host %s: %v", host.ID, err)
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

// GetVMsForHost retrieves a merged list of VMs for the frontend.
func (s *HostService) GetVMsForHost(hostID string) ([]VMView, error) {
	// 1. Get live data from libvirt and sync DB in the process
	liveVMs, err := s.syncAndListVMs(hostID)
	if err != nil {
		return nil, fmt.Errorf("could not get live VM data for host %s: %w", hostID, err)
	}
	liveVMMap := make(map[string]libvirt.VMInfo)
	for _, vm := range liveVMs {
		liveVMMap[vm.UUID] = vm
	}

	// 2. Get DB records for the host
	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return nil, fmt.Errorf("could not get DB VM records for host %s: %w", hostID, err)
	}

	// 3. Create the unified view
	var vmViews []VMView
	for _, dbVM := range dbVMs {
		view := VMView{
			ID:          dbVM.ID,
			Name:        dbVM.Name,
			Description: dbVM.Description,
			VCPUCount:   dbVM.VCPUCount,
			MemoryBytes: dbVM.MemoryBytes,
			IsTemplate:  dbVM.IsTemplate,
		}

		// Merge live data if the VM is found
		if liveVM, ok := liveVMMap[dbVM.UUID]; ok {
			view.State = liveVM.State
			view.MaxMem = liveVM.MaxMem
			view.Memory = liveVM.Memory
			view.Vcpu = liveVM.Vcpu
			view.CpuTime = liveVM.CpuTime
			view.Uptime = liveVM.Uptime
			view.Graphics = liveVM.Graphics
		} else {
			// VM exists in DB but not in libvirt
			view.State = -1 // Using -1 as a custom state for "Not Found" or "Undefined"
		}
		vmViews = append(vmViews, view)
	}

	return vmViews, nil
}


// syncAndListVMs is the core function to get VMs from libvirt and sync with the local DB.
func (s *HostService) syncAndListVMs(hostID string) ([]libvirt.VMInfo, error) {
	liveVMs, err := s.connector.ListAllDomains(hostID)
	if err != nil {
		return nil, fmt.Errorf("service failed to list vms for host %s: %w", hostID, err)
	}

	// Sync with DB
	for _, vmInfo := range liveVMs {
		vmRecord := storage.VirtualMachine{
			HostID:      hostID,
			Name:        vmInfo.Name,
			UUID:        vmInfo.UUID,
			VCPUCount:   vmInfo.Vcpu,
			MemoryBytes: vmInfo.MaxMem * 1024, // libvirt reports MaxMem in KiB
		}

		// Use FirstOrCreate to find the existing VM by UUID or create a new one.
		// Then, update its configuration details.
		if err := s.db.Where(storage.VirtualMachine{UUID: vmInfo.UUID}).Assign(vmRecord).FirstOrCreate(&vmRecord).Error; err != nil {
			log.Printf("Warning: could not sync VM %s to database: %v", vmInfo.Name, err)
		}
	}

	// Optional: Prune DB entries for VMs that no longer exist on the host
	var liveVMUUIDs []string
	for _, vm := range liveVMs {
		liveVMUUIDs = append(liveVMUUIDs, vm.UUID)
	}
	if err := s.db.Where("host_id = ? AND uuid NOT IN ?", hostID, liveVMUUIDs).Delete(&storage.VirtualMachine{}).Error; err != nil {
		log.Printf("Warning: failed to prune old VMs for host %s: %v", hostID, err)
	}

	return liveVMs, nil
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


