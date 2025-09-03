package libvirt

import (
	"fmt"
	"sync"

	"libvirt.org/go/libvirt"
)

// VMInfo holds basic information about a virtual machine.
type VMInfo struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	MaxMem    uint64 `json:"maxMem"`
	IsManaged bool   `json:"isManaged"`
}

// Connector manages a pool of connections to libvirt hosts.
type Connector struct {
	connections map[string]*libvirt.Connect
	mu          sync.RWMutex
}

// NewConnector creates and initializes a new libvirt connector.
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[string]*libvirt.Connect),
	}
}

// AddHost establishes a new connection to a libvirt host and adds it to the pool.
func (c *Connector) AddHost(hostID, uri string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.connections[hostID]; exists {
		return fmt.Errorf("host '%s' already connected", hostID)
	}

	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return fmt.Errorf("failed to connect to host '%s' at %s: %w", hostID, uri, err)
	}

	c.connections[hostID] = conn
	fmt.Printf("Successfully connected to host '%s'\n", hostID)
	return nil
}

// RemoveHost disconnects from a libvirt host and removes it from the pool.
func (c *Connector) RemoveHost(hostID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, exists := c.connections[hostID]
	if !exists {
		return fmt.Errorf("host '%s' not found", hostID)
	}

	if err := conn.Close(); err != nil {
		// Log the error but don't prevent removal
		fmt.Printf("Warning: error while closing connection to host '%s': %v\n", hostID, err)
	}

	delete(c.connections, hostID)
	fmt.Printf("Disconnected from host '%s'\n", hostID)
	return nil
}

// ListAllDomains retrieves a list of all virtual machines from a specific host.
func (c *Connector) ListAllDomains(hostID string) ([]VMInfo, error) {
	c.mu.RLock()
	conn, exists := c.connections[hostID]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("host '%s' not connected", hostID)
	}

	// Flag includes all domains, active and inactive
	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains on host '%s': %w", hostID, err)
	}

	var vms []VMInfo
	for i := range domains {
		// It's important to free the domain handle after use
		defer domains[i].Free()

		name, err := domains[i].GetName()
		if err != nil {
			fmt.Printf("Warning: failed to get name for a domain on host '%s': %v\n", hostID, err)
			continue
		}

		state, _, err := domains[i].GetState()
		if err != nil {
			fmt.Printf("Warning: failed to get state for domain '%s' on host '%s': %v\n", name, hostID, err)
			continue
		}

		info, err := domains[i].GetInfo()
		if err != nil {
			fmt.Printf("Warning: failed to get info for domain '%s' on host '%s': %v\n", name, hostID, err)
			continue
		}

		id := domains[i].GetID() // Returns -1 for inactive domains, which is fine as it becomes ^uint32(0)

		isManaged, err := domains[i].IsManagedSave()
		if err != nil {
			fmt.Printf("Warning: failed to check managed save state for domain '%s' on host '%s': %v\n", name, hostID, err)
			continue
		}

		vms = append(vms, VMInfo{
			ID:        id,
			Name:      name,
			State:     mapDomainStateToString(state),
			MaxMem:    info.MaxMem,
			IsManaged: isManaged,
		})
	}

	return vms, nil
}

// mapDomainStateToString converts libvirt domain state enum to a human-readable string.
func mapDomainStateToString(state libvirt.DomainState) string {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return "No State"
	case libvirt.DOMAIN_RUNNING:
		return "Running"
	case libvirt.DOMAIN_BLOCKED:
		return "Blocked"
	case libvirt.DOMAIN_PAUSED:
		return "Paused"
	case libvirt.DOMAIN_SHUTDOWN:
		return "Shutdown"
	case libvirt.DOMAIN_SHUTOFF:
		return "Shutoff"
	case libvirt.DOMAIN_CRASHED:
		return "Crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "Suspended"
	default:
		return "Unknown"
	}
}


