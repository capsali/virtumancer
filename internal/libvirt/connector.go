package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/capsali/virtumancer/internal/storage"
	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// typedParamValueString returns a human-readable representation of a TypedParamValue
// including the discriminator (D) and the concrete Go value held in I.
func typedParamValueString(v libvirt.TypedParamValue) string {
	switch val := v.I.(type) {
	case int32:
		return fmt.Sprintf("D=%d int32=%d", v.D, val)
	case int64:
		return fmt.Sprintf("D=%d int64=%d", v.D, val)
	case uint32:
		return fmt.Sprintf("D=%d uint32=%d", v.D, val)
	case uint64:
		return fmt.Sprintf("D=%d uint64=%d", v.D, val)
	case float32:
		return fmt.Sprintf("D=%d float32=%f", v.D, val)
	case float64:
		return fmt.Sprintf("D=%d float64=%f", v.D, val)
	case string:
		return fmt.Sprintf("D=%d string=%s", v.D, val)
	default:
		return fmt.Sprintf("D=%d unknown=%v", v.D, val)
	}
}

// GraphicsInfo holds details about available graphics consoles.
type GraphicsInfo struct {
	VNC   bool `json:"vnc"`
	SPICE bool `json:"spice"`
}

// VMInfo holds basic information about a virtual machine.
type VMInfo struct {
	ID         uint32              `json:"id"`
	UUID       string              `json:"uuid"`
	Name       string              `json:"name"`
	State      libvirt.DomainState `json:"state"`
	MaxMem     uint64              `json:"max_mem"`
	Memory     uint64              `json:"memory"`
	Vcpu       uint                `json:"vcpu"`
	CpuTime    uint64              `json:"cpu_time"`
	Uptime     int64               `json:"uptime"`
	Persistent bool                `json:"persistent"`
	Autostart  bool                `json:"autostart"`
	Graphics   GraphicsInfo        `json:"graphics"`
}

// DomainDiskStats holds I/O statistics for a single disk device.
type DomainDiskStats struct {
	Device     string `json:"device"`
	ReadBytes  int64  `json:"read_bytes"`
	WriteBytes int64  `json:"write_bytes"`
}

// DomainNetworkStats holds I/O statistics for a single network interface.
type DomainNetworkStats struct {
	Device     string `json:"device"`
	ReadBytes  int64  `json:"read_bytes"`
	WriteBytes int64  `json:"write_bytes"`
}

// VMStats holds real-time statistics for a single VM.
type VMStats struct {
	State     libvirt.DomainState  `json:"state"`
	Memory    uint64               `json:"memory"`
	MaxMem    uint64               `json:"max_mem"`
	Vcpu      uint                 `json:"vcpu"`
	CpuTime   uint64               `json:"cpu_time"`
	DiskStats []DomainDiskStats    `json:"disk_stats"`
	NetStats  []DomainNetworkStats `json:"net_stats"`
}

// HardwareInfo holds the hardware configuration of a VM.
type HardwareInfo struct {
	Disks    []DiskInfo    `json:"disks"`
	Networks []NetworkInfo `json:"networks"`
}

// DiskInfo represents a virtual disk.
type DiskInfo struct {
	Type   string `xml:"type,attr" json:"type"`
	Device string `xml:"device,attr" json:"device"`
	Driver struct {
		Name string `xml:"name,attr" json:"driver_name"`
		Type string `xml:"type,attr" json:"type"`
	} `xml:"driver" json:"driver"`
	Source struct {
		File string `xml:"file,attr"`
		Dev  string `xml:"dev,attr"`
	} `xml:"source"`
	Path   string `json:"path"`
	Target struct {
		Dev string `xml:"dev,attr" json:"dev"`
		Bus string `xml:"bus,attr" json:"bus"`
	} `xml:"target" json:"target"`
}

// NetworkInfo represents a virtual network interface.
type NetworkInfo struct {
	Type string `xml:"type,attr" json:"type"`
	Mac  struct {
		Address string `xml:"address,attr" json:"address"`
	} `xml:"mac" json:"mac"`
	Source struct {
		Bridge string `xml:"bridge,attr" json:"bridge"`
	} `xml:"source" json:"source"`
	Model struct {
		Type string `xml:"type,attr" json:"type"`
	} `xml:"model" json:"model"`
	Target struct {
		Dev string `xml:"dev,attr" json:"dev"`
	} `xml:"target" json:"target"`
}

// DomainHardwareXML is used for unmarshalling hardware info from the domain XML.
type DomainHardwareXML struct {
	Devices struct {
		Disks      []DiskInfo    `xml:"disk"`
		Interfaces []NetworkInfo `xml:"interface"`
	} `xml:"devices"`
}

// HostInfo holds basic information and statistics about a hypervisor host.
type HostInfo struct {
	Hostname   string `json:"hostname"`
	CPU        uint   `json:"cpu"`
	Memory     uint64 `json:"memory"`
	MemoryUsed uint64 `json:"memory_used"`
	// Uptime seconds on the host machine. Libvirt does not provide a host
	// uptime value via the standard NodeGet* APIs; this field will be 0 unless
	// an external method (SSH /proc/uptime) is used to populate it.
	Uptime  int64 `json:"uptime"`
	Cores   uint  `json:"cores"`
	Threads uint  `json:"threads"`
}

// HostStats holds real-time statistics for a single host.
type HostStats struct {
	CPUUtilization float64 `json:"cpu_utilization"`
	MemoryUsed     uint64  `json:"memory_used"`
}

// Connector manages active connections to libvirt hosts.
type Connector struct {
	connections  map[string]*libvirt.Libvirt
	mu           sync.RWMutex
	lastCPUStats map[string][]libvirt.NodeGetCPUStats
	lastMemStats map[string]uint64
	// sshClients holds an existing *ssh.Client for hosts connected via qemu+ssh
	// so we can reuse the session for quick commands like reading /proc/uptime.
	sshClients map[string]*ssh.Client
	// uptimeCache stores a cached uptime value with a timestamp to avoid
	// executing SSH commands on every UI refresh.
	uptimeCache map[string]struct {
		uptime int64
		at     time.Time
	}
}

// NewConnector creates a new libvirt connection manager.
func NewConnector() *Connector {
	return &Connector{
		connections:  make(map[string]*libvirt.Libvirt),
		lastCPUStats: make(map[string][]libvirt.NodeGetCPUStats),
		lastMemStats: make(map[string]uint64),
		sshClients:   make(map[string]*ssh.Client),
		uptimeCache: make(map[string]struct {
			uptime int64
			at     time.Time
		}),
	}
}

// defaultDialTimeout is the conservative timeout used for network/ssh/connect
// operations during startup so a slow/unreachable host doesn't block the server.
const defaultDialTimeout = 5 * time.Second

// sshDialWithTimeout performs ssh.Dial but enforces a timeout by running the
// dial in a goroutine and selecting on a timer. This prevents long blocking
// SSH connect attempts from stalling startup.
func sshDialWithTimeout(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	type result struct {
		client *ssh.Client
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := ssh.Dial(network, addr, config)
		ch <- result{client: c, err: err}
	}()

	select {
	case r := <-ch:
		return r.client, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("ssh dial to %s timed out after %s", addr, timeout)
	}
}

// typedParamToUint64 converts a libvirt.TypedParam.Value to uint64 when possible.
// The libvirt.TypedParamValue uses a discriminated union so the concrete
// numeric type may vary depending on the platform/version.
func typedParamToUint64(v libvirt.TypedParamValue) uint64 {
	switch val := v.I.(type) {
	case int32:
		return uint64(val)
	case int64:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	case float32:
		return uint64(val)
	case float64:
		return uint64(val)
	default:
		return 0
	}
}

// getMemoryUsageFromParams attempts to derive used and total memory from
// NodeGetMemoryParameters result. It prefers an explicit "used" field if
// present. If not present, it will try to compute used = total - (free + cached).
// Returned values are in bytes.
func getMemoryUsageFromParams(params []libvirt.TypedParam, totalKiB uint64) (usedBytes uint64, ok bool) {
	var used uint64
	var free uint64
	var cached uint64

	for _, p := range params {
		name := strings.ToLower(p.Field)
		switch name {
		case "used", "actual-used":
			used = typedParamToUint64(p.Value)
		case "free", "actual-free", "available":
			free = typedParamToUint64(p.Value)
		case "cached", "cache", "buffers":
			cached = typedParamToUint64(p.Value)
		}
	}

	totalBytes := totalKiB * 1024

	// Prefer explicit 'used' when present. Interpret as KiB first (common),
	// then as bytes if KiB interpretation doesn't fit. If explicit 'used' is
	// not available, fall back to computed used = total - (free + cached).

	if used > 0 {
		// Interpret 'used' as KiB first
		usedKiBBytes := used * 1024
		if usedKiBBytes <= totalBytes {
			return usedKiBBytes, true
		}
		// If that seems too large, try interpreting 'used' as bytes
		if used <= totalBytes {
			return used, true
		}
	}

	// Compute used from free+cached when available. Try KiB interpretation first.
	if free > 0 || cached > 0 {
		effFreeKiB := free + cached
		// Compute using KiB units
		if totalBytes > effFreeKiB*1024 {
			return totalBytes - effFreeKiB*1024, true
		}
		// Fall back to bytes interpretation
		effFree := free + cached
		if totalBytes > effFree {
			return totalBytes - effFree, true
		}
		return 0, true
	}

	return 0, false
}

// paramsContainCached checks whether any of the params correspond to cached/buffers fields.
func paramsContainCached(params []libvirt.TypedParam) bool {
	for _, p := range params {
		name := strings.ToLower(p.Field)
		if name == "cached" || name == "cache" || name == "buffers" {
			return true
		}
	}
	return false
}

// nodeMemoryStatsToTypedParams converts NodeGetMemoryStats entries into TypedParam entries.
func nodeMemoryStatsToTypedParams(stats []libvirt.NodeGetMemoryStats) []libvirt.TypedParam {
	out := make([]libvirt.TypedParam, 0, len(stats))
	for _, s := range stats {
		out = append(out, libvirt.TypedParam{
			Field: s.Field,
			Value: libvirt.TypedParamValue{D: 4, I: s.Value},
		})
	}
	return out
}

// sshKeyAuth provides an AuthMethod for key-based SSH authentication
// by reading the user's default private key.
func sshKeyAuth() (ssh.AuthMethod, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}

	keyPath := filepath.Join(home, ".ssh", "id_rsa")
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key from %s: %w. Ensure SSH key-based auth is set up", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// sshTunneledConn wraps a net.Conn to ensure the underlying SSH client is also closed.
type sshTunneledConn struct {
	net.Conn
	client *ssh.Client
}

func (c *sshTunneledConn) Close() error {
	connErr := c.Conn.Close()
	clientErr := c.client.Close()
	if connErr != nil {
		return connErr
	}
	return clientErr
}

// dialLibvirt establishes a network connection based on the URI.
func dialLibvirt(uri string) (net.Conn, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	switch parsedURI.Scheme {
	case "qemu+ssh":
		user := "root" // default user
		if parsedURI.User != nil {
			user = parsedURI.User.Username()
		}

		host := parsedURI.Hostname()
		port := parsedURI.Port()
		if port == "" {
			port = "22" // default ssh port
		}
		sshAddr := fmt.Sprintf("%s:%s", host, port)

		authMethod, err := sshKeyAuth()
		if err != nil {
			return nil, fmt.Errorf("SSH key authentication setup failed: %w", err)
		}

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			// Insecure: fine for this tool where hosts are explicitly added.
			// Production systems might use a known_hosts file.
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		log.Printf("Attempting SSH connection to %s for user %s", sshAddr, user)
		sshClient, err := sshDialWithTimeout("tcp", sshAddr, sshConfig, defaultDialTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to dial SSH to %s: %w", sshAddr, err)
		}

		// Dial the libvirt socket on the remote machine through the SSH tunnel.
		remoteSocketPath := "/var/run/libvirt/libvirt-sock"
		log.Printf("SSH connected. Dialing remote libvirt socket at %s", remoteSocketPath)
		conn, err := sshClient.Dial("unix", remoteSocketPath)
		if err != nil {
			sshClient.Close()
			return nil, fmt.Errorf("failed to dial remote libvirt socket (%s) via SSH: %w", remoteSocketPath, err)
		}
		return &sshTunneledConn{
			Conn:   conn,
			client: sshClient,
		}, nil

	case "qemu+tcp":
		address := parsedURI.Host
		if !strings.Contains(address, ":") {
			address = address + ":16509" // Default libvirt tcp port
		}
		return net.DialTimeout("tcp", address, defaultDialTimeout)

	case "qemu", "qemu+unix":
		address := parsedURI.Path
		if address == "" || address == "/system" {
			address = "/var/run/libvirt/libvirt-sock"
		}
		// For unix sockets, use a short timeout by dialing via a net.Dialer with deadline.
		d := net.Dialer{Timeout: defaultDialTimeout}
		return d.Dial("unix", address)

	default:
		return nil, fmt.Errorf("unsupported scheme: %s", parsedURI.Scheme)
	}
}

// AddHost connects to a given libvirt URI and adds it to the connection pool.
func (c *Connector) AddHost(host storage.Host) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.connections[host.ID]; ok {
		return fmt.Errorf("host '%s' is already connected", host.ID)
	}

	conn, err := dialLibvirt(host.URI)
	if err != nil {
		return fmt.Errorf("failed to dial libvirt for host '%s': %w", host.ID, err)
	}

	// If this connection wraps an SSH client, capture it for reuse.
	if stc, ok := conn.(*sshTunneledConn); ok {
		if stc.client != nil {
			c.sshClients[host.ID] = stc.client
		}
	}

	l := libvirt.New(conn)
	if err := l.Connect(); err != nil {
		conn.Close() // Ensure the connection is closed on failure
		return fmt.Errorf("failed to connect to libvirt rpc for host '%s': %w", host.ID, err)
	}

	c.connections[host.ID] = l
	log.Printf("Successfully connected to host: %s", host.ID)
	return nil
}

// RemoveHost disconnects from a libvirt host and removes it from the pool.
func (c *Connector) RemoveHost(hostID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	l, ok := c.connections[hostID]
	if !ok {
		return fmt.Errorf("host '%s' not found", hostID)
	}

	if err := l.Disconnect(); err != nil {
		return fmt.Errorf("failed to close connection to host '%s': %w", hostID, err)
	}

	delete(c.connections, hostID)
	// Close and remove any stored ssh client for this host.
	if client, ok := c.sshClients[hostID]; ok {
		client.Close()
		delete(c.sshClients, hostID)
	}
	// Remove uptime cache entry as well.
	delete(c.uptimeCache, hostID)
	log.Printf("Disconnected from host: %s", hostID)
	return nil
}

// getHostUptime returns host uptime in seconds, using a cached value when recent.
// It requires that an *ssh.Client for the host is stored in c.sshClients.
func (c *Connector) getHostUptime(hostID string, ttl time.Duration, timeout time.Duration) (int64, error) {
	c.mu.RLock()
	if ent, ok := c.uptimeCache[hostID]; ok {
		if time.Since(ent.at) < ttl {
			c.mu.RUnlock()
			return ent.uptime, nil
		}
	}
	client, ok := c.sshClients[hostID]
	c.mu.RUnlock()
	if !ok || client == nil {
		return 0, fmt.Errorf("no ssh client available for host %s", hostID)
	}

	// Execute 'cat /proc/uptime' with a short timeout by using a Goroutine.
	type result struct {
		out []byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		sess, err := client.NewSession()
		if err != nil {
			ch <- result{nil, err}
			return
		}
		defer sess.Close()
		out, err := sess.Output("cat /proc/uptime")
		ch <- result{out, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return 0, fmt.Errorf("ssh session failed: %w", r.err)
		}
		// parse first float from output
		fields := strings.Fields(string(r.out))
		if len(fields) == 0 {
			return 0, fmt.Errorf("unexpected /proc/uptime output: %q", string(r.out))
		}
		secF, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse uptime: %w", err)
		}
		uptime := int64(secF)
		c.mu.Lock()
		c.uptimeCache[hostID] = struct {
			uptime int64
			at     time.Time
		}{uptime: uptime, at: time.Now()}
		c.mu.Unlock()
		return uptime, nil
	case <-time.After(timeout):
		return 0, fmt.Errorf("uptime ssh command timed out")
	}
}

// GetConnection returns the active connection for a given host ID.
func (c *Connector) GetConnection(hostID string) (*libvirt.Libvirt, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	conn, ok := c.connections[hostID]
	if !ok {
		return nil, fmt.Errorf("not connected to host '%s'", hostID)
	}
	return conn, nil
}

// GetHostInfo retrieves statistics about the host itself.
func (c *Connector) GetHostInfo(hostID string) (*HostInfo, error) {
	l, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	_, memory, cpus, _, _, _, cores, threads, err := l.NodeGetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info for host %s: %w", hostID, err)
	}

	hostname, err := l.ConnectGetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname for host %s: %w", hostID, err)
	}

	// Use NodeGetMemoryStats to retrieve memory stats (per-node, numeric fields),
	// convert them to TypedParam so existing parsing logic can be reused.
	// NodeGetMemoryStats follows a two-step pattern: first call with nparams=0
	// to discover how many entries are available, then call again with that
	// count to retrieve the entries. Some libvirt backends return 0 entries
	// on the first call but provide rNparams > 0.
	var params []libvirt.TypedParam
	stats, rNparams, err := l.NodeGetMemoryStats(0, -1, 0)
	if err != nil {
		log.Printf("Warning: could not get memory stats count for host %s: %v", hostID, err)
	} else if rNparams <= 0 {
		log.Printf("Host %s: NodeGetMemoryStats reported %d available entries", hostID, rNparams)
	} else {
		// rNparams > 0, fetch the actual entries.
		stats, _, err = l.NodeGetMemoryStats(rNparams, -1, 0)
		if err != nil {
			log.Printf("Warning: failed to fetch %d memory stats for host %s: %v", rNparams, hostID, err)
		} else {
			params = nodeMemoryStatsToTypedParams(stats)
			// params are available for parsing; debug logging removed
		}
	}

	totalMemoryBytes := uint64(memory) * 1024 // NodeGetInfo returns KiB
	var memoryUsed uint64
	if u, ok := getMemoryUsageFromParams(params, uint64(memory)); ok {
		// getMemoryUsageFromParams returns a value in bytes when ok==true.
		memoryUsed = u
	} else {
		// Final fallback: try the older NodeGetFreeMemory call and compute used as total - free.
		freeMemory, ferr := l.NodeGetFreeMemory()
		if ferr != nil {
			log.Printf("Warning: could not get free memory for host %s: %v", hostID, ferr)
			freeMemory = 0
		}
		if freeMemory > 0 {
			memoryUsed = totalMemoryBytes - freeMemory
		}
	}

	// Attempt to get host uptime via cached SSH client (quick). Non-fatal on error.
	var uptimeSec int64
	if u, err := c.getHostUptime(hostID, 60*time.Second, 3*time.Second); err == nil {
		uptimeSec = u
	} else {
		log.Printf("Warning: could not get uptime for host %s: %v", hostID, err)
	}

	return &HostInfo{
		Hostname:   hostname,
		CPU:        uint(cpus),
		Memory:     totalMemoryBytes,
		MemoryUsed: memoryUsed,
		Uptime:     uptimeSec,
		Cores:      uint(cores),
		Threads:    uint(threads),
	}, nil
}

// GetHostStats retrieves real-time statistics about the host itself.
func (c *Connector) GetHostStats(hostID string) (*HostStats, error) {
	l, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	// Get CPU stats
	// First call to get the number of parameters.
	_, nparams, err := l.NodeGetCPUStats(-1, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU stats count for host %s: %w", hostID, err)
	}

	// Second call to get the actual stats.
	cpuStats, _, err := l.NodeGetCPUStats(-1, nparams, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu stats for host %s: %w", hostID, err)
	}

	var cpuUtilization float64
	c.mu.Lock()
	defer c.mu.Unlock()

	if lastStats, ok := c.lastCPUStats[hostID]; ok {
		var totalLast, totalNow, idleLast, idleNow uint64

		lastStatsMap := make(map[string]uint64)
		for _, stat := range lastStats {
			lastStatsMap[stat.Field] = stat.Value
		}

		nowStatsMap := make(map[string]uint64)
		for _, stat := range cpuStats {
			nowStatsMap[stat.Field] = stat.Value
		}

		totalLast = lastStatsMap["kernel"] + lastStatsMap["user"] + lastStatsMap["idle"] + lastStatsMap["iowait"] + lastStatsMap["irq"] + lastStatsMap["softirq"]
		totalNow = nowStatsMap["kernel"] + nowStatsMap["user"] + nowStatsMap["idle"] + nowStatsMap["iowait"] + nowStatsMap["irq"] + nowStatsMap["softirq"]
		idleLast = lastStatsMap["idle"]
		idleNow = nowStatsMap["idle"]

		diffTotal := totalNow - totalLast
		diffIdle := idleNow - idleLast

		if diffTotal > 0 {
			cpuUtilization = 1.0 - float64(diffIdle)/float64(diffTotal)
		}
	}

	c.lastCPUStats[hostID] = cpuStats

	// Prefer an explicit 'used' value from memory parameters when possible.
	// Prefer NodeGetMemoryStats for runtime stats and reuse the same parsing helper.
	// See comment above: two-step retrieval to get available count, then entries.
	stats, rNparams, err := l.NodeGetMemoryStats(0, -1, 0)
	var params []libvirt.TypedParam
	if err != nil {
		log.Printf("Warning: could not get memory stats count for host %s: %v", hostID, err)
	} else if rNparams <= 0 {
		log.Printf("Host %s: NodeGetMemoryStats reported %d available entries", hostID, rNparams)
	} else {
		stats, _, err = l.NodeGetMemoryStats(rNparams, -1, 0)
		if err != nil {
			log.Printf("Warning: failed to fetch %d memory stats for host %s: %v", rNparams, hostID, err)
		} else {
			params = nodeMemoryStatsToTypedParams(stats)
			// params are available for parsing; debug logging removed
		}
	}

	_, totalMemory, _, _, _, _, _, _, err := l.NodeGetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info for host %s: %w", hostID, err)
	}

	totalMemoryBytes := uint64(totalMemory) * 1024
	var memoryUsed uint64
	if u, ok := getMemoryUsageFromParams(params, uint64(totalMemory)); ok {
		memoryUsed = u
	} else {
		freeMemory, ferr := l.NodeGetFreeMemory()
		if ferr != nil {
			log.Printf("Warning: could not get free memory for host %s: %v", hostID, ferr)
			freeMemory = 0
		}
		if freeMemory > 0 {
			memoryUsed = totalMemoryBytes - freeMemory
		}
	}

	return &HostStats{
		CPUUtilization: cpuUtilization,
		MemoryUsed:     memoryUsed,
	}, nil
}

// parseGraphicsFromXML extracts VNC and SPICE availability from a domain's XML definition.
func parseGraphicsFromXML(xmlDesc string) (GraphicsInfo, error) {
	type GraphicsXML struct {
		Type string `xml:"type,attr"`
		Port string `xml:"port,attr"`
	}
	type DomainDef struct {
		Graphics []GraphicsXML `xml:"devices>graphics"`
	}

	var def DomainDef
	var graphics GraphicsInfo

	if err := xml.Unmarshal([]byte(xmlDesc), &def); err != nil {
		return graphics, fmt.Errorf("failed to parse domain XML: %w", err)
	}

	for _, g := range def.Graphics {
		if g.Port != "" && g.Port != "-1" {
			switch strings.ToLower(g.Type) {
			case "vnc":
				graphics.VNC = true
			case "spice":
				graphics.SPICE = true
			}
		}
	}

	return graphics, nil
}

// ListAllDomains lists all domains (VMs) on a specific host.
func (c *Connector) ListAllDomains(hostID string) ([]VMInfo, error) {
	l, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domains, err := l.Domains()
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	var vms []VMInfo
	for _, domain := range domains {
		vmInfo, err := c.domainToVMInfo(l, domain)
		if err != nil {
			log.Printf("Warning: could not get info for domain %s on host %s: %v", domain.Name, hostID, err)
			continue
		}
		vms = append(vms, *vmInfo)
	}

	return vms, nil
}

// GetDomainInfo retrieves information for a single domain.
func (c *Connector) GetDomainInfo(hostID, vmName string) (*VMInfo, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}
	return c.domainToVMInfo(l, domain)
}

// domainToVMInfo is a helper to convert a libvirt.Domain object to our VMInfo struct.
func (c *Connector) domainToVMInfo(l *libvirt.Libvirt, domain libvirt.Domain) (*VMInfo, error) {
	stateInt, _, err := l.DomainGetState(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain state for %s: %w", domain.Name, err)
	}
	state := libvirt.DomainState(stateInt)

	_, maxMem, memory, nrVirtCPU, cpuTime, err := l.DomainGetInfo(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info for %s: %w", domain.Name, err)
	}

	var uptime int64 = -1
	if state == libvirt.DomainRunning {
		seconds, nanoseconds, err := l.DomainGetTime(domain, 0)
		if err == nil {
			uptime = int64(seconds) + int64(nanoseconds)/1_000_000_000
		}
	}

	persistent, err := l.DomainIsPersistent(domain)
	if err != nil {
		persistent = 0
	}
	autostart, err := l.DomainGetAutostart(domain)
	if err != nil {
		autostart = 0
	}
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, err
	}
	graphics, err := parseGraphicsFromXML(xmlDesc)
	if err != nil {
		return nil, err
	}

	var uuidStr string
	// The domain.UUID is a [16]byte array. We need to convert it to a slice to use uuid.FromBytes
	parsedUUID, err := uuid.FromBytes(domain.UUID[:])
	if err != nil {
		// This should not happen if libvirt provides a valid 16-byte UUID, but we handle it defensively.
		log.Printf("Warning: could not parse domain UUID for %s: %v. Using raw hex.", domain.Name, err)
		uuidStr = fmt.Sprintf("%x", domain.UUID)
	} else {
		uuidStr = parsedUUID.String()
	}

	return &VMInfo{
		ID:         uint32(domain.ID),
		UUID:       uuidStr,
		Name:       domain.Name,
		State:      state,
		MaxMem:     uint64(maxMem),
		Memory:     uint64(memory),
		Vcpu:       uint(nrVirtCPU),
		CpuTime:    cpuTime,
		Uptime:     uptime,
		Persistent: persistent == 1,
		Autostart:  autostart == 1,
		Graphics:   graphics,
	}, nil
}

// GetDomainStats retrieves real-time statistics for a single domain (VM).
func (c *Connector) GetDomainStats(hostID, vmName string) (*VMStats, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	stateInt, _, err := l.DomainGetState(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("could not get state for domain %s: %w", vmName, err)
	}
	state := libvirt.DomainState(stateInt)

	_, maxMem, memory, nrVirtCPU, cpuTime, err := l.DomainGetInfo(domain)
	if err != nil {
		return nil, fmt.Errorf("could not get info for domain %s: %w", vmName, err)
	}

	// If not running, return basic info without I/O stats
	if state != libvirt.DomainRunning {
		return &VMStats{
			State:     state,
			Memory:    0,
			MaxMem:    uint64(maxMem),
			Vcpu:      uint(nrVirtCPU),
			CpuTime:   0,
			DiskStats: []DomainDiskStats{},
			NetStats:  []DomainNetworkStats{},
		}, nil
	}

	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get XML for %s to find devices: %w", vmName, err)
	}

	var def DomainHardwareXML
	if err := xml.Unmarshal([]byte(xmlDesc), &def); err != nil {
		return nil, fmt.Errorf("failed to parse domain XML for devices: %w", err)
	}

	var diskStats []DomainDiskStats
	for _, disk := range def.Devices.Disks {
		if disk.Target.Dev == "" {
			continue
		}
		rdReq, rdBytes, wrReq, wrBytes, errs, err := l.DomainBlockStats(domain, disk.Target.Dev)
		if err != nil {
			log.Printf("Warning: could not get block stats for device %s on VM %s: %v", disk.Target.Dev, vmName, err)
			continue
		}
		_ = rdReq // Suppress unused variable warning
		_ = wrReq // Suppress unused variable warning
		_ = errs  // Suppress unused variable warning
		diskStats = append(diskStats, DomainDiskStats{
			Device:     disk.Target.Dev,
			ReadBytes:  rdBytes,
			WriteBytes: wrBytes,
		})
	}

	var netStats []DomainNetworkStats
	for _, iface := range def.Devices.Interfaces {
		if iface.Target.Dev == "" {
			continue
		}
		rxBytes, _, _, _, txBytes, _, _, _, err := l.DomainInterfaceStats(domain, iface.Target.Dev)
		if err != nil {
			log.Printf("Warning: could not get interface stats for device %s on VM %s: %v", iface.Target.Dev, vmName, err)
			continue
		}
		netStats = append(netStats, DomainNetworkStats{
			Device:     iface.Target.Dev,
			ReadBytes:  int64(rxBytes),
			WriteBytes: int64(txBytes),
		})
	}

	stats := &VMStats{
		State:     state,
		Memory:    uint64(memory),
		MaxMem:    uint64(maxMem),
		Vcpu:      uint(nrVirtCPU),
		CpuTime:   cpuTime,
		DiskStats: diskStats,
		NetStats:  netStats,
	}

	return stats, nil
}

// GetDomainHardware retrieves the hardware configuration for a single domain (VM).
func (c *Connector) GetDomainHardware(hostID, vmName string) (*HardwareInfo, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get XML for %s to read hardware: %w", vmName, err)
	}

	var def DomainHardwareXML
	if err := xml.Unmarshal([]byte(xmlDesc), &def); err != nil {
		return nil, fmt.Errorf("failed to parse domain XML for hardware: %w", err)
	}

	hardware := &HardwareInfo{
		Disks:    def.Devices.Disks,
		Networks: def.Devices.Interfaces,
	}

	// Post-process disks to populate the unified 'Path' field.
	for i := range hardware.Disks {
		if hardware.Disks[i].Source.File != "" {
			hardware.Disks[i].Path = hardware.Disks[i].Source.File
		} else if hardware.Disks[i].Source.Dev != "" {
			hardware.Disks[i].Path = hardware.Disks[i].Source.Dev
		}
	}

	return hardware, nil
}

// --- VM Actions ---

func (c *Connector) getDomainByName(hostID, vmName string) (*libvirt.Libvirt, libvirt.Domain, error) {
	l, err := c.GetConnection(hostID)
	if err != nil {
		return nil, libvirt.Domain{}, err
	}
	domain, err := l.DomainLookupByName(vmName)
	if err != nil {
		return nil, libvirt.Domain{}, fmt.Errorf("could not find VM '%s' on host '%s': %w", vmName, hostID, err)
	}
	return l, domain, nil
}

func (c *Connector) StartDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	return l.DomainCreate(domain)
}

func (c *Connector) ShutdownDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	return l.DomainShutdown(domain)
}

func (c *Connector) RebootDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	return l.DomainReboot(domain, 0)
}

func (c *Connector) DestroyDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	return l.DomainDestroy(domain)
}

func (c *Connector) ResetDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	return l.DomainReset(domain, 0)
}
