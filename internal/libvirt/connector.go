package libvirt

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/capsali/virtumancer/internal/storage"
	"libvirt.org/go/libvirt"
)

// VMInfo holds basic information about a virtual machine.
type VMInfo struct {
	ID         uint32              `json:"id"`
	Name       string              `json:"name"`
	State      libvirt.DomainState `json:"state"`
	MaxMem     uint64              `json:"max_mem"`
	Memory     uint64              `json:"memory"`
	Vcpu       uint                `json:"vcpu"`
	Persistent bool                `json:"persistent"`
	Autostart  bool                `json:"autostart"`
}

// Connector manages active connections to libvirt hosts.
type Connector struct {
	connections map[string]*libvirt.Connect
	mu          sync.RWMutex
}

// NewConnector creates a new libvirt connection manager.
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[string]*libvirt.Connect),
	}
}

// AddHost connects to a given libvirt URI and adds it to the connection pool.
func (c *Connector) AddHost(host storage.Host) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.connections[host.ID]; ok {
		return fmt.Errorf("host '%s' is already connected", host.ID)
	}

	connectURI := host.URI

	// For SSH connections, we modify the URI to bypass host key checking.
	parsedURI, err := url.Parse(host.URI)
	if err == nil && (parsedURI.Scheme == "qemu+ssh" || parsedURI.Scheme == "libssh") {
		q := parsedURI.Query()
		if q.Get("no_verify") == "" {
			q.Set("no_verify", "1")
			parsedURI.RawQuery = q.Encode()
			connectURI = parsedURI.String()
			log.Printf("Amended URI for %s to %s for non-interactive connection", host.ID, connectURI)
		}
	}

	conn, err := libvirt.NewConnect(connectURI)
	if err != nil {
		return fmt.Errorf("failed to connect to host '%s' using URI %s: %w", host.ID, connectURI, err)
	}

	c.connections[host.ID] = conn
	log.Printf("Successfully connected to host: %s", host.ID)
	return nil
}

// RemoveHost disconnects from a libvirt host and removes it from the pool.
func (c *Connector) RemoveHost(hostID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.connections[hostID]
	if !ok {
		return fmt.Errorf("host '%s' not found", hostID)
	}

	if _, err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection to host '%s': %w", hostID, err)
	}

	delete(c.connections, hostID)
	log.Printf("Disconnected from host: %s", hostID)
	return nil
}

// GetConnection returns the active connection for a given host ID.
func (c *Connector) GetConnection(hostID string) (*libvirt.Connect, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	conn, ok := c.connections[hostID]
	if !ok {
		return nil, fmt.Errorf("not connected to host '%s'", hostID)
	}
	return conn, nil
}

// ListAllDomains lists all domains (VMs) on a specific host.
func (c *Connector) ListAllDomains(hostID string) ([]VMInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	var vms []VMInfo
	for i := range domains {
		defer domains[i].Free()

		name, err := domains[i].GetName()
		if err != nil {
			continue
		}
		id, err := domains[i].GetID()
		if err != nil {
			id = 0 // Not running
		}
		state, _, err := domains[i].GetState()
		if err != nil {
			continue
		}
		info, err := domains[i].GetInfo()
		if err != nil {
			continue
		}
		isPersistent, err := domains[i].IsPersistent()
		if err != nil {
			continue
		}
		autostart, err := domains[i].GetAutostart()
		if err != nil {
			continue
		}

		vms = append(vms, VMInfo{
			ID:         uint32(id),
			Name:       name,
			State:      state,
			MaxMem:     info.MaxMem,
			Memory:     info.Memory,
			Vcpu:       uint(info.NrVirtCpu),
			Persistent: isPersistent,
			Autostart:  autostart,
		})
	}

	return vms, nil
}

// --- VM Actions ---

// getDomainByName is a helper to find a domain by its name.
func (c *Connector) getDomainByName(hostID, vmName string) (*libvirt.Domain, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}
	domain, err := conn.LookupDomainByName(vmName)
	if err != nil {
		return nil, fmt.Errorf("could not find VM '%s' on host '%s': %w", vmName, hostID, err)
	}
	return domain, nil
}

// StartDomain starts a virtual machine by name.
func (c *Connector) StartDomain(hostID, vmName string) error {
	domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	defer domain.Free()
	err = domain.Create() // This is equivalent to 'virsh start'
	if err != nil {
		return fmt.Errorf("failed to start VM '%s': %w", vmName, err)
	}
	log.Printf("Started VM '%s' on host '%s'", vmName, hostID)
	return nil
}

// GracefulShutdownDomain gracefully shuts down a virtual machine.
func (c *Connector) GracefulShutdownDomain(hostID, vmName string) error {
	domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	defer domain.Free()
	err = domain.Shutdown() // This is equivalent to 'virsh shutdown'
	if err != nil {
		return fmt.Errorf("failed to gracefully shutdown VM '%s': %w", vmName, err)
	}
	log.Printf("Gracefully shut down VM '%s' on host '%s'", vmName, hostID)
	return nil
}

// GracefulRebootDomain gracefully reboots a virtual machine.
func (c *Connector) GracefulRebootDomain(hostID, vmName string) error {
	domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	defer domain.Free()
	err = domain.Reboot(0) // This is equivalent to 'virsh reboot'
	if err != nil {
		return fmt.Errorf("failed to gracefully reboot VM '%s': %w", vmName, err)
	}
	log.Printf("Gracefully rebooted VM '%s' on host '%s'", vmName, hostID)
	return nil
}

// ForceOffDomain forces a virtual machine to stop.
func (c *Connector) ForceOffDomain(hostID, vmName string) error {
	domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	defer domain.Free()
	err = domain.Destroy() // This is equivalent to 'virsh destroy'
	if err != nil {
		return fmt.Errorf("failed to force off VM '%s': %w", vmName, err)
	}
	log.Printf("Forced off VM '%s' on host '%s'", vmName, hostID)
	return nil
}

// ForceResetDomain forces a virtual machine to reset.
func (c *Connector) ForceResetDomain(hostID, vmName string) error {
	domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	defer domain.Free()
	err = domain.Reset(0) // This is equivalent to 'virsh reset'
	if err != nil {
		return fmt.Errorf("failed to force reset VM '%s': %w", vmName, err)
	}
	log.Printf("Forced reset on VM '%s' on host '%s'", vmName, hostID)
	return nil
}


