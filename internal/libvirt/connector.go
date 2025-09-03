package libvirt

import (
	"fmt"
	"log"
	"sync"

	"libvirt.org/go/libvirt"
)

// DomainInfo holds detailed, serializable information about a VM.
type DomainInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	State      string `json:"state"`
	Memory     uint64 `json:"memory_kb"`      // Current memory
	MaxMemory  uint64 `json:"max_memory_kb"`  // Maximum memory
	VCPUs      uint   `json:"vcpus"`
	Persistent bool   `json:"persistent"`
	Title      string `json:"title,omitempty"` // omitempty because it might not be set
}

// Connector manages active connections to multiple libvirt hosts.
type Connector struct {
	mu          sync.RWMex
	connections map[string]*libvirt.Connect
}

// NewConnector creates and initializes a new Connector.
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[string]*libvirt.Connect),
	}
}

// Connect establishes a new connection to a libvirt host and stores it.
// hostID is a user-defined alias for the connection.
// uri is the libvirt connection URI (e.g., "qemu+ssh://user@host/system").
func (c *Connector) Connect(hostID, uri string) (*libvirt.Connect, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.connections[hostID]; exists {
		return nil, fmt.Errorf("a connection with ID '%s' already exists", hostID)
	}

	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt at %s: %w", uri, err)
	}

	c.connections[hostID] = conn
	log.Printf("Established new libvirt connection: %s (%s)", hostID, uri)
	return conn, nil
}

// GetConnection retrieves an active connection by its ID.
func (c *Connector) GetConnection(hostID string) (*libvirt.Connect, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	conn, ok := c.connections[hostID]
	if !ok {
		return nil, fmt.Errorf("no active libvirt connection found for host ID '%s'", hostID)
	}
	return conn, nil
}

// Disconnect closes and removes a connection.
func (c *Connector) Disconnect(hostID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.connections[hostID]
	if !ok {
		return fmt.Errorf("no active libvirt connection found for host ID '%s'", hostID)
	}

	if _, err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close libvirt connection for %s: %w", hostID, err)
	}

	delete(c.connections, hostID)
	log.Printf("Closed and removed libvirt connection: %s", hostID)
	return nil
}

// ListAllDomains retrieves a list of all domains (VMs) from a connected host.
func (c *Connector) ListAllDomains(hostID string) ([]DomainInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	// List all domains, including inactive ones.
	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains for host '%s': %w", hostID, err)
	}

	var domainInfos []DomainInfo
	for _, domain := range domains {
		defer domain.Free()

		name, err := domain.GetName()
		if err != nil {
			log.Printf("Warning: could not get name for a domain on host %s: %v", hostID, err)
			continue
		}

		uuid, err := domain.GetUUIDString()
		if err != nil {
			log.Printf("Warning: could not get UUID for domain '%s' on host %s: %v", name, hostID, err)
			continue
		}

		state, _, err := domain.GetState()
		if err != nil {
			log.Printf("Warning: could not get state for domain '%s' on host %s: %v", name, hostID, err)
			continue
		}

		info, err := domain.GetInfo()
		if err != nil {
			log.Printf("Warning: could not get info for domain '%s' on host %s: %v", name, hostID, err)
			continue
		}

		id, err := domain.GetID()
		if err != nil {
			// This can happen if the domain is shutoff, GetID returns an error.
			// We'll set it to 0 as a convention for shutoff VMs.
			id = 0
		}
		
		isPersistent, err := domain.IsPersistent()
		if err != nil {
			log.Printf("Warning: could not get persistence for domain '%s' on host %s: %v", name, hostID, err)
			isPersistent = false
		}

		title, err := domain.GetMetadata(libvirt.DOMAIN_METADATA_TITLE, "")
		if err != nil {
			// Title is optional, so we don't need to log a warning if it's not found.
			title = ""
		}


		domainInfos = append(domainInfos, DomainInfo{
			ID:         id,
			Name:       name,
			UUID:       uuid,
			State:      domainStateToString(state),
			Memory:     info.Memory,
			MaxMemory:  info.MaxMem,
			VCPUs:      info.NrVirtCpu,
			Persistent: isPersistent,
			Title:      title,
		})
	}

	return domainInfos, nil
}

// Helper function to convert libvirt domain state enum to a string.
func domainStateToString(state libvirt.DomainState) string {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return "nostate"
	case libvirt.DOMAIN_RUNNING:
		return "running"
	case libvirt.DOMAIN_BLOCKED:
		return "blocked"
	case libvirt.DOMAIN_PAUSED:
		return "paused"
	case libvirt.DOMAIN_SHUTDOWN:
		return "shutdown"
	case libvirt.DOMAIN_SHUTOFF:
		return "shutoff"
	case libvirt.DOMAIN_CRASHED:
		return "crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "pmsuspended"
	default:
		return "unknown"
	}
}


