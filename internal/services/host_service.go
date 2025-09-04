package services

import (
	"fmt"
	"log"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"gorm.io/gorm"
)

// HostService provides business logic for host management.
type HostService struct {
	db        *gorm.DB
	connector *libvirt.Connector
}

// NewHostService creates a new HostService.
func NewHostService(db *gorm.DB, connector *libvirt.Connector) *HostService {
	return &HostService{db: db, connector: connector}
}

// GetAllHosts retrieves all hosts from the database.
func (s *HostService) GetAllHosts() ([]storage.Host, error) {
	var hosts []storage.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// AddHost adds a new host to the database and connects to it.
func (s *HostService) AddHost(host storage.Host) (*storage.Host, error) {
	if err := s.db.Create(&host).Error; err != nil {
		return nil, fmt.Errorf("failed to save host to database: %w", err)
	}

	err := s.connector.AddHost(host)
	if err != nil {
		// If connection fails, we should roll back the database change.
		if delErr := s.db.Delete(&host).Error; delErr != nil {
			log.Printf("CRITICAL: Failed to rollback host creation for %s after connection failure. Please remove manually. DB Error: %v", host.ID, delErr)
		}
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}

	return &host, nil
}

// RemoveHost removes a host from the database and disconnects from it.
func (s *HostService) RemoveHost(hostID string) error {
	// Disconnect from the host first.
	if err := s.connector.RemoveHost(hostID); err != nil {
		// We log a warning but proceed to delete from the DB anyway,
		// in case the host was just offline.
		log.Printf("Warning: failed to disconnect from host %s during removal, continuing with DB deletion: %v", hostID, err)
	}

	// Delete the host from the database.
	if err := s.db.Where("id = ?", hostID).Delete(&storage.Host{}).Error; err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}
	return nil
}

// ConnectToAllHosts attempts to connect to all hosts stored in the database on startup.
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
		}
	}
}

// ListVMs retrieves the list of VMs for a specific host.
func (s *HostService) ListVMs(hostID string) ([]libvirt.VMInfo, error) {
	vms, err := s.connector.ListAllDomains(hostID)
	if err != nil {
		return nil, fmt.Errorf("service failed to list vms for host %s: %w", hostID, err)
	}
	return vms, nil
}

// StartVM starts a virtual machine on a specific host.
func (s *HostService) StartVM(hostID, vmName string) error {
	return s.connector.StartDomain(hostID, vmName)
}

// GracefulShutdownVM gracefully shuts down a virtual machine on a specific host.
func (s *HostService) GracefulShutdownVM(hostID, vmName string) error {
	return s.connector.GracefulShutdownDomain(hostID, vmName)
}

// GracefulRebootVM gracefully reboots a virtual machine on a specific host.
func (s *HostService) GracefulRebootVM(hostID, vmName string) error {
	return s.connector.GracefulRebootDomain(hostID, vmName)
}

// ForceOffVM forces a virtual machine to stop on a specific host.
func (s *HostService) ForceOffVM(hostID, vmName string) error {
	return s.connector.ForceOffDomain(hostID, vmName)
}

// ForceResetVM forces a virtual machine to reset on a specific host.
func (s *HostService) ForceResetVM(hostID, vmName string) error {
	return s.connector.ForceResetDomain(hostID, vmName)
}


