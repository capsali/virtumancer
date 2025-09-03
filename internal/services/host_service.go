package services

import (
	"fmt"
	"log"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
)

// ErrHostNotConnected is a custom error for when a host isn't connected.
type ErrHostNotConnected struct {
	HostID string
}

func (e ErrHostNotConnected) Error() string {
	return fmt.Sprintf("no active libvirt connection found for host ID '%s'", e.HostID)
}

// HostService orchestrates operations between the database and the libvirt connector.
type HostService struct {
	db        *storage.DB
	connector *libvirt.Connector
}

// NewHostService creates a new HostService.
func NewHostService(db *storage.DB, connector *libvirt.Connector) *HostService {
	return &HostService{
		db:        db,
		connector: connector,
	}
}

// ConnectAllHosts loads all hosts from the DB and attempts to connect to them.
func (s *HostService) ConnectAllHosts() error {
	hosts, err := s.db.GetAllHosts()
	if err != nil {
		return fmt.Errorf("could not retrieve hosts from database: %w", err)
	}

	log.Printf("Found %d saved hosts. Attempting to connect...", len(hosts))
	for _, host := range hosts {
		if _, err := s.connector.Connect(host.ID, host.URI); err != nil {
			log.Printf("WARNING: Failed to auto-connect to host '%s' (%s): %v", host.ID, host.URI, err)
		}
	}
	return nil
}

// GetAllHosts retrieves all hosts from the database.
func (s *HostService) GetAllHosts() ([]storage.Host, error) {
	return s.db.GetAllHosts()
}

// AddHost saves a host to the DB and then establishes a libvirt connection.
func (s *HostService) AddHost(id, uri string) (*storage.Host, error) {
	newHost := &storage.Host{ID: id, URI: uri}
	if err := s.db.AddHost(newHost); err != nil {
		return nil, fmt.Errorf("failed to save host to database: %w", err)
	}

	if _, err := s.connector.Connect(id, uri); err != nil {
		return newHost, fmt.Errorf("host saved to DB, but connection failed: %w", err)
	}

	return newHost, nil
}

// DeleteHost disconnects from a host and then removes it from the database.
func (s *HostService) DeleteHost(hostID string) error {
	if err := s.connector.Disconnect(hostID); err != nil {
		log.Printf("Warning: could not disconnect from host '%s' (it may have been offline): %v", hostID, err)
	}

	if err := s.db.DeleteHost(hostID); err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}

	return nil
}


