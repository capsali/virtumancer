package libvirt

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/capsali/virtumancer/internal/logging"

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
	Uptime    int64                `json:"uptime"`
	DiskStats []DomainDiskStats    `json:"disk_stats"`
	NetStats  []DomainNetworkStats `json:"net_stats"`
}

// HardwareInfo holds the hardware configuration of a VM.
type HardwareInfo struct {
	// Basic VM info
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Metadata      string `json:"metadata"`
	OSType        string `json:"os_type"`
	CurrentMemory uint64 `json:"current_memory"`

	// OS Configuration
	OSConfig   *OSConfigInfo `json:"os_config,omitempty"`
	SMBIOSInfo []SMBIOSInfo  `json:"smbios_info,omitempty"`

	// CPU Configuration
	CPUInfo     *CPUConfigInfo   `json:"cpu_info,omitempty"`
	CPUFeatures []CPUFeatureInfo `json:"cpu_features,omitempty"`

	// Memory Configuration
	MemoryBacking *MemoryBackingInfo `json:"memory_backing,omitempty"`
	NUMANodes     []NUMANodeInfo     `json:"numa_nodes,omitempty"`

	// Security
	SecurityLabels []SecurityLabelInfo `json:"security_labels,omitempty"`
	LaunchSecurity *LaunchSecurityInfo `json:"launch_security,omitempty"`

	// Features
	HypervisorFeatures []HypervisorFeatureInfo `json:"hypervisor_features,omitempty"`

	// Lifecycle
	LifecycleActions *LifecycleActionInfo `json:"lifecycle_actions,omitempty"`

	// Clock
	ClockConfig *ClockInfo `json:"clock_config,omitempty"`

	// Performance
	PerfEvents []PerfEventInfo `json:"perf_events,omitempty"`

	// Existing device arrays
	Disks     []DiskInfo    `json:"disks"`
	Networks  []NetworkInfo `json:"networks"`
	Videos    []VideoInfo   `json:"videos,omitempty"`
	Consoles  []ConsoleInfo `json:"consoles,omitempty"`
	Hostdevs  []HostdevInfo `json:"hostdevs,omitempty"`
	BlockDevs []BlockDev    `json:"blockdevs,omitempty"`
	IOThreads []IOThread    `json:"iothreads,omitempty"`
	Mdevs     []MdevInfo    `json:"mdevs,omitempty"`
	Boot      []BootEntry   `json:"boot,omitempty"`
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
		File string `xml:"file,attr" json:"file"`
		Dev  string `xml:"dev,attr" json:"dev"`
	} `xml:"source" json:"source"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	ReadOnly  bool   `xml:"readonly" json:"readonly"`
	Shareable bool   `xml:"shareable" json:"shareable"`
	Target    struct {
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
	Name        string `xml:"name" json:"name"`
	UUID        string `xml:"uuid" json:"uuid"`
	Title       string `xml:"title" json:"title"`
	Description string `xml:"description" json:"description"`
	Metadata    struct {
		Content string `xml:",innerxml" json:"content"`
	} `xml:"metadata" json:"metadata"`
	OS struct {
		Type   string `xml:"type" json:"type"`
		Loader struct {
			Path      string `xml:",chardata" json:"path"`
			Type      string `xml:"type,attr" json:"type"`
			Readonly  string `xml:"readonly,attr" json:"readonly"`
			Secure    string `xml:"secure,attr" json:"secure"`
			Stateless string `xml:"stateless,attr" json:"stateless"`
		} `xml:"loader" json:"loader"`
		NVram struct {
			Path     string `xml:",chardata" json:"path"`
			Template string `xml:"template,attr" json:"template"`
			Type     string `xml:"type,attr" json:"type"`
		} `xml:"nvram" json:"nvram"`
		Bootmenu struct {
			Enable  string `xml:"enable,attr" json:"enable"`
			Timeout string `xml:"timeout,attr" json:"timeout"`
		} `xml:"bootmenu" json:"bootmenu"`
		SmBIOS struct {
			Mode string `xml:"mode,attr" json:"mode"`
		} `xml:"smbios" json:"smbios"`
		Firmware struct {
			Value string `xml:",chardata" json:"value"`
		} `xml:"firmware" json:"firmware"`
		BIOS struct {
			UsesSerial    string `xml:"useserial,attr" json:"useserial"`
			RebootTimeout string `xml:"rebootTimeout,attr" json:"rebootTimeout"`
		} `xml:"bios" json:"bios"`
	} `xml:"os" json:"os"`
	Memory struct {
		Value uint64 `xml:",chardata" json:"value"`
		Unit  string `xml:"unit,attr" json:"unit"`
	} `xml:"memory" json:"memory"`
	CurrentMemory struct {
		Value uint64 `xml:",chardata" json:"value"`
		Unit  string `xml:"unit,attr" json:"unit"`
	} `xml:"currentMemory" json:"currentMemory"`
	CPU struct {
		Mode  string `xml:"mode,attr" json:"mode"`
		Model struct {
			Name     string `xml:",chardata" json:"name"`
			Fallback string `xml:"fallback,attr" json:"fallback"`
		} `xml:"model" json:"model"`
		Topology struct {
			Sockets uint `xml:"sockets,attr" json:"sockets"`
			Cores   uint `xml:"cores,attr" json:"cores"`
			Threads uint `xml:"threads,attr" json:"threads"`
		} `xml:"topology" json:"topology"`
		Features []struct {
			Name   string `xml:"name,attr" json:"name"`
			Policy string `xml:"policy,attr" json:"policy"`
		} `xml:"feature" json:"features"`
	} `xml:"cpu" json:"cpu"`
	MemoryBacking struct {
		Hugepages struct {
			Page []struct {
				Size    uint64 `xml:"size,attr" json:"size"`
				Unit    string `xml:"unit,attr" json:"unit"`
				Nodeset string `xml:"nodeset,attr" json:"nodeset"`
			} `xml:"page" json:"page"`
		} `xml:"hugepages" json:"hugepages"`
		Nosharepages struct{} `xml:"nosharepages" json:"nosharepages"`
		Locked       struct{} `xml:"locked" json:"locked"`
		Source       struct {
			Type string `xml:"type,attr" json:"type"`
		} `xml:"source" json:"source"`
		Access struct {
			Mode string `xml:"mode,attr" json:"mode"`
		} `xml:"access" json:"access"`
	} `xml:"memoryBacking" json:"memoryBacking"`
	NUMA struct {
		Cell []NUMANodeInfo `xml:"cell" json:"cell"`
	} `xml:"numa" json:"numa"`
	Features struct {
		PAE     struct{} `xml:"pae" json:"pae"`
		ACPI    struct{} `xml:"acpi" json:"acpi"`
		APIC    struct{} `xml:"apic" json:"apic"`
		HAP     struct{} `xml:"hap" json:"hap"`
		Privnet struct{} `xml:"privnet" json:"privnet"`
		HyperV  struct {
			Mode    string `xml:"mode,attr" json:"mode"`
			Relaxed struct {
				State string `xml:"state,attr" json:"state"`
			} `xml:"relaxed" json:"relaxed"`
			VAPIC struct {
				State string `xml:"state,attr" json:"state"`
			} `xml:"vapic" json:"vapic"`
			Spinlocks struct {
				State   string `xml:"state,attr" json:"state"`
				Retries string `xml:"retries,attr" json:"retries"`
			} `xml:"spinlocks" json:"spinlocks"`
		} `xml:"hyperv" json:"hyperv"`
		KVM struct {
			Hidden struct {
				State string `xml:"state,attr" json:"state"`
			} `xml:"hidden" json:"hidden"`
			HintDedicated struct {
				State string `xml:"state,attr" json:"state"`
			} `xml:"hint-dedicated" json:"hint-dedicated"`
		} `xml:"kvm" json:"kvm"`
		PVSpinlock struct {
			State string `xml:"state,attr" json:"state"`
		} `xml:"pvspinlock" json:"pvspinlock"`
	} `xml:"features" json:"features"`
	OnPoweroff    string `xml:"on_poweroff,attr" json:"on_poweroff"`
	OnReboot      string `xml:"on_reboot,attr" json:"on_reboot"`
	OnCrash       string `xml:"on_crash,attr" json:"on_crash"`
	OnLockfailure string `xml:"on_lockfailure,attr" json:"on_lockfailure"`
	Clock         struct {
		Offset     string `xml:"offset,attr" json:"offset"`
		Timezone   string `xml:"timezone,attr" json:"timezone"`
		Basis      string `xml:"basis,attr" json:"basis"`
		Adjustment int64  `xml:"adjustment,attr" json:"adjustment"`
	} `xml:"clock" json:"clock"`
	Perf struct {
		Event []struct {
			Name  string `xml:"name,attr" json:"name"`
			State string `xml:"enabled,attr" json:"state"`
		} `xml:"event" json:"event"`
	} `xml:"perf" json:"perf"`
	Devices struct {
		Disks      []DiskInfo     `xml:"disk"`
		Interfaces []NetworkInfo  `xml:"interface"`
		Videos     []VideoInfo    `xml:"video"`
		Consoles   []ConsoleInfo  `xml:"console"`
		Hostdevs   []HostdevInfo  `xml:"hostdev"`
		BlockDevs  []BlockDev     `xml:"blockdev"`
		IOThreads  []IOThread     `xml:"iothread"`
		Mdevs      []MdevInfo     `xml:"mdev"`
		NUMANodes  []NUMANodeInfo `xml:"numa>cell"`
		Boot       []BootEntry    `xml:"boot"`
		CPU        *CPUInfo       `xml:"cpu"`
	} `xml:"devices"`
}

// VideoInfo represents a <video> entry in domain XML.
type VideoInfo struct {
	Model struct {
		Type  string `xml:"type,attr" json:"type"`
		VRAM  int    `xml:"vram,attr,omitempty" json:"vram,omitempty"`
		Heads int    `xml:"heads,attr,omitempty" json:"heads,omitempty"`
	} `xml:"model" json:"model"`
}

// ConsoleInfo represents a <console> entry (serial/graphics consoles may use <console> too).
type ConsoleInfo struct {
	Type   string `xml:"type,attr" json:"type"`
	Target struct {
		Dev string `xml:"dev,attr" json:"dev"`
	} `xml:"target" json:"target"`
}

// HostdevInfo represents a <hostdev> passthrough device (PCI/USB) in domain XML.
type HostdevInfo struct {
	Mode   string `xml:"mode,attr" json:"mode"`
	Type   string `xml:"type,attr" json:"type"`
	Source struct {
		Address struct {
			Domain   string `xml:"domain,attr" json:"domain"`
			Bus      string `xml:"bus,attr" json:"bus"`
			Slot     string `xml:"slot,attr" json:"slot"`
			Function string `xml:"function,attr" json:"function"`
		} `xml:"address" json:"address"`
	} `xml:"source" json:"source"`
}

// BlockDev is a lightweight representation of a <blockdev> element.
type BlockDev struct {
	NodeName string `xml:"node-name,attr" json:"node_name"`
	Driver   struct {
		Name string `xml:"name,attr" json:"name"`
		Type string `xml:"type,attr" json:"type"`
	} `xml:"driver" json:"driver"`
}

// IOThread represents an <iothread> element when present.
type IOThread struct {
	Name string `xml:"name,attr" json:"name"`
}

// MdevInfo represents a mediated device entry (<mdev> or <mdev:...> structures).
type MdevInfo struct {
	Type string `xml:"type,attr" json:"type"`
	UUID string `xml:"uuid,attr,omitempty" json:"uuid,omitempty"`
}

// NUMANodeInfo represents a <numa><cell>...</cell></numa> cell entry.
type NUMANodeInfo struct {
	ID       int    `xml:"id,attr" json:"id"`
	MemoryKB uint64 `xml:"memory,attr" json:"memory_kb"`
	CPUs     string `xml:"cpus" json:"cpus"`
}

// BootEntry represents <boot dev="..."/> entries.
type BootEntry struct {
	Dev   string `xml:"dev,attr" json:"dev"`
	Order int    `xml:"order,attr" json:"order"`
}

// CPUInfo is a minimal representation of <cpu> subtree for parsing features/topology.
type CPUInfo struct {
	Mode string `xml:"mode,attr" json:"mode"`
}

// OSConfigInfo represents OS configuration information.
type OSConfigInfo struct {
	Type      string        `json:"type"`
	Arch      string        `json:"arch,omitempty"`
	Machine   string        `json:"machine,omitempty"`
	BootMenu  *BootMenuInfo `json:"boot_menu,omitempty"`
	BootDev   []string      `json:"boot_dev,omitempty"`
	Init      string        `json:"init,omitempty"`
	InitArgs  []string      `json:"init_args,omitempty"`
	InitEnv   []InitEnvInfo `json:"init_env,omitempty"`
	InitDir   string        `json:"init_dir,omitempty"`
	InitUser  string        `json:"init_user,omitempty"`
	InitGroup string        `json:"init_group,omitempty"`
}

// BootMenuInfo represents boot menu configuration.
type BootMenuInfo struct {
	Enable  string `json:"enable"`
	Timeout string `json:"timeout,omitempty"`
}

// InitEnvInfo represents environment variables for init.
type InitEnvInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SMBIOSInfo represents SMBIOS configuration.
type SMBIOSInfo struct {
	Mode string `json:"mode"`
}

// CPUConfigInfo represents CPU configuration information.
type CPUConfigInfo struct {
	Mode       string           `json:"mode,omitempty"`
	Model      string           `json:"model,omitempty"`
	Match      string           `json:"match,omitempty"`
	Check      string           `json:"check,omitempty"`
	Migratable string           `json:"migratable,omitempty"`
	Topology   *CPUTopologyInfo `json:"topology,omitempty"`
	Vendor     string           `json:"vendor,omitempty"`
	VendorID   string           `json:"vendor_id,omitempty"`
}

// CPUTopologyInfo represents CPU topology configuration.
type CPUTopologyInfo struct {
	Sockets int `json:"sockets,omitempty"`
	Cores   int `json:"cores,omitempty"`
	Threads int `json:"threads,omitempty"`
}

// CPUFeatureInfo represents CPU feature configuration.
type CPUFeatureInfo struct {
	Name   string `json:"name"`
	Policy string `json:"policy,omitempty"`
}

// MemoryBackingInfo represents memory backing configuration.
type MemoryBackingInfo struct {
	HugePages    *HugePagesInfo `json:"hugepages,omitempty"`
	NoSharePages bool           `json:"nosharepages,omitempty"`
	Locked       bool           `json:"locked,omitempty"`
	Source       string         `json:"source,omitempty"`
	Access       string         `json:"access,omitempty"`
	Allocation   string         `json:"allocation,omitempty"`
	Discard      bool           `json:"discard,omitempty"`
}

// HugePagesInfo represents huge pages configuration.
type HugePagesInfo struct {
	Page []HugePageInfo `json:"page,omitempty"`
}

// HugePageInfo represents a single huge page configuration.
type HugePageInfo struct {
	Size    string `json:"size"`
	Unit    string `json:"unit,omitempty"`
	Nodeset string `json:"nodeset,omitempty"`
}

// SecurityLabelInfo represents security label configuration.
type SecurityLabelInfo struct {
	Type    string `json:"type"`
	Label   string `json:"label,omitempty"`
	Relabel string `json:"relabel,omitempty"`
}

// LaunchSecurityInfo represents launch security configuration.
type LaunchSecurityInfo struct {
	Type            string `json:"type"`
	CBitPos         string `json:"cbitpos,omitempty"`
	ReducedPhysBits string `json:"reduced_phys_bits,omitempty"`
	Policy          string `json:"policy,omitempty"`
	DHCert          string `json:"dh_cert,omitempty"`
	Session         string `json:"session,omitempty"`
}

// HypervisorFeatureInfo represents hypervisor feature configuration.
type HypervisorFeatureInfo struct {
	Name  string `json:"name"`
	State string `json:"state,omitempty"`
}

// LifecycleActionInfo represents lifecycle action configuration.
type LifecycleActionInfo struct {
	OnPoweroff    string `json:"on_poweroff,omitempty"`
	OnReboot      string `json:"on_reboot,omitempty"`
	OnCrash       string `json:"on_crash,omitempty"`
	OnLockFailure string `json:"on_lock_failure,omitempty"`
}

// ClockInfo represents clock configuration.
type ClockInfo struct {
	Offset string           `json:"offset"`
	Timers []ClockTimerInfo `json:"timers,omitempty"`
}

// ClockTimerInfo represents clock timer configuration.
type ClockTimerInfo struct {
	Name       string `json:"name"`
	Track      string `json:"track,omitempty"`
	TickPolicy string `json:"tick_policy,omitempty"`
	Frequency  string `json:"frequency,omitempty"`
	Mode       string `json:"mode,omitempty"`
	Present    string `json:"present,omitempty"`
}

// PerfEventInfo represents performance event configuration.
type PerfEventInfo struct {
	Name  string `json:"name"`
	Event string `json:"event"`
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

		log.Debugf("dialLibvirt: attempting SSH connection to %s for user %s", sshAddr, user)
		sshClient, err := sshDialWithTimeout("tcp", sshAddr, sshConfig, defaultDialTimeout)
		if err != nil {
			log.Debugf("dialLibvirt: ssh dial to %s failed: %v", sshAddr, err)
			return nil, fmt.Errorf("failed to dial SSH to %s: %w", sshAddr, err)
		}

		// Dial the libvirt socket on the remote machine through the SSH tunnel.
		remoteSocketPath := "/var/run/libvirt/libvirt-sock"
		log.Verbosef("SSH connected to %s. Dialing remote libvirt socket at %s", sshAddr, remoteSocketPath)
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
	log.Verbosef("Successfully connected to host: %s", host.ID)
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
	log.Verbosef("Disconnected from host: %s", hostID)
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
		log.Verbosef("Could not get memory stats count for host %s: %v", hostID, err)
	} else if rNparams <= 0 {
		log.Debugf("Host %s: NodeGetMemoryStats reported %d available entries", hostID, rNparams)
	} else {
		// rNparams > 0, fetch the actual entries.
		stats, _, err = l.NodeGetMemoryStats(rNparams, -1, 0)
		if err != nil {
			log.Verbosef("Failed to fetch %d memory stats for host %s: %v", rNparams, hostID, err)
		} else {
			params = nodeMemoryStatsToTypedParams(stats)
			log.Debugf("Fetched %d memory params for host %s", len(params), hostID)
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
			log.Debugf("Warning: could not get free memory for host %s: %v", hostID, ferr)
		} else {
			log.Verbosef("Host %s: NodeGetMemoryStats reported %d available entries (using NodeGetFreeMemory fallback)", hostID, rNparams)
			memoryUsed = totalMemoryBytes - freeMemory
		}
	}

	// Attempt to get host uptime via cached SSH client (quick). Non-fatal on error.
	var uptimeSec int64
	if u, err := c.getHostUptime(hostID, 60*time.Second, 3*time.Second); err == nil {
		uptimeSec = u
	} else {
		log.Verbosef("Could not get uptime for host %s: %v", hostID, err)
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
		log.Debugf("Warning: could not get memory stats count for host %s: %v", hostID, err)
	} else if rNparams <= 0 {
		log.Verbosef("Host %s: NodeGetMemoryStats reported %d available entries", hostID, rNparams)
	} else {
		stats, _, err = l.NodeGetMemoryStats(rNparams, -1, 0)
		if err != nil {
			log.Verbosef("Warning: failed to fetch %d memory stats for host %s: %v", rNparams, hostID, err)
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
			log.Verbosef("Warning: could not get free memory for host %s: %v", hostID, ferr)
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
			log.Debugf("Warning: could not get info for domain %s on host %s: %v", domain.Name, hostID, err)
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
		log.Debugf("Warning: could not parse domain UUID for %s: %v. Using raw hex.", domain.Name, err)
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
			log.Debugf("Warning: could not get block stats for device %s on VM %s: %v", disk.Target.Dev, vmName, err)
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
			log.Debugf("Warning: could not get interface stats for device %s on VM %s: %v", iface.Target.Dev, vmName, err)
			continue
		}
		netStats = append(netStats, DomainNetworkStats{
			Device:     iface.Target.Dev,
			ReadBytes:  int64(rxBytes),
			WriteBytes: int64(txBytes),
		})
	}

	var uptime int64 = -1
	if state == libvirt.DomainRunning {
		seconds, nanoseconds, err := l.DomainGetTime(domain, 0)
		if err == nil {
			uptime = int64(seconds) + int64(nanoseconds)/1_000_000_000
		}
	}

	stats := &VMStats{
		State:     state,
		Memory:    uint64(memory),
		MaxMem:    uint64(maxMem),
		Vcpu:      uint(nrVirtCPU),
		CpuTime:   cpuTime,
		Uptime:    uptime,
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
		Name:          def.Name,
		UUID:          def.UUID,
		Title:         def.Title,
		Description:   def.Description,
		Metadata:      def.Metadata.Content,
		OSType:        def.OS.Type,
		CurrentMemory: def.CurrentMemory.Value,
		Disks:         def.Devices.Disks,
		Networks:      def.Devices.Interfaces,
		Videos:        def.Devices.Videos,
		Consoles:      def.Devices.Consoles,
		Hostdevs:      def.Devices.Hostdevs,
		BlockDevs:     def.Devices.BlockDevs,
		IOThreads:     def.Devices.IOThreads,
		Mdevs:         def.Devices.Mdevs,
		NUMANodes:     def.Devices.NUMANodes,
		Boot:          def.Devices.Boot,
	}

	// Post-process disks to populate the unified 'Path' field.
	for i := range hardware.Disks {
		if hardware.Disks[i].Source.File != "" {
			hardware.Disks[i].Path = hardware.Disks[i].Source.File
		} else if hardware.Disks[i].Source.Dev != "" {
			hardware.Disks[i].Path = hardware.Disks[i].Source.Dev
		}
	}

	// Normalize NUMA CPU lists (if present) by trimming whitespace.
	for i := range hardware.NUMANodes {
		hardware.NUMANodes[i].CPUs = strings.TrimSpace(hardware.NUMANodes[i].CPUs)
	}

	// Populate OS Configuration
	if def.OS.Type != "" {
		hardware.OSConfig = &OSConfigInfo{
			Type:    def.OS.Type,
			Arch:    def.OS.Loader.Type, // This might need adjustment based on actual XML structure
			Machine: def.OS.Firmware.Value,
		}
		if def.OS.Bootmenu.Enable != "" {
			hardware.OSConfig.BootMenu = &BootMenuInfo{
				Enable:  def.OS.Bootmenu.Enable,
				Timeout: def.OS.Bootmenu.Timeout,
			}
		}
	}

	// Populate SMBIOS Info
	if def.OS.SmBIOS.Mode != "" {
		hardware.SMBIOSInfo = []SMBIOSInfo{{Mode: def.OS.SmBIOS.Mode}}
	}

	// Populate CPU Configuration
	if def.CPU.Mode != "" || def.CPU.Model.Name != "" {
		hardware.CPUInfo = &CPUConfigInfo{
			Mode:  def.CPU.Mode,
			Model: def.CPU.Model.Name,
		}
		if def.CPU.Topology.Sockets > 0 || def.CPU.Topology.Cores > 0 || def.CPU.Topology.Threads > 0 {
			hardware.CPUInfo.Topology = &CPUTopologyInfo{
				Sockets: int(def.CPU.Topology.Sockets),
				Cores:   int(def.CPU.Topology.Cores),
				Threads: int(def.CPU.Topology.Threads),
			}
		}
	}

	// Populate CPU Features
	for _, feature := range def.CPU.Features {
		hardware.CPUFeatures = append(hardware.CPUFeatures, CPUFeatureInfo{
			Name:   feature.Name,
			Policy: feature.Policy,
		})
	}

	// Populate Memory Backing
	if def.MemoryBacking.Source.Type != "" || len(def.MemoryBacking.Hugepages.Page) > 0 {
		hardware.MemoryBacking = &MemoryBackingInfo{
			Source: def.MemoryBacking.Source.Type,
			Access: def.MemoryBacking.Access.Mode,
		}
		if len(def.MemoryBacking.Hugepages.Page) > 0 {
			hardware.MemoryBacking.HugePages = &HugePagesInfo{}
			for _, page := range def.MemoryBacking.Hugepages.Page {
				hardware.MemoryBacking.HugePages.Page = append(hardware.MemoryBacking.HugePages.Page, HugePageInfo{
					Size:    strconv.FormatUint(page.Size, 10),
					Unit:    page.Unit,
					Nodeset: page.Nodeset,
				})
			}
		}
		// Set boolean flags
		hardware.MemoryBacking.NoSharePages = def.MemoryBacking.Nosharepages != struct{}{}
		hardware.MemoryBacking.Locked = def.MemoryBacking.Locked != struct{}{}
	}

	// Populate Hypervisor Features
	if def.Features.PAE != (struct{}{}) {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "pae", State: "on"})
	}
	if def.Features.ACPI != (struct{}{}) {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "acpi", State: "on"})
	}
	if def.Features.APIC != (struct{}{}) {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "apic", State: "on"})
	}
	if def.Features.HAP != (struct{}{}) {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "hap", State: "on"})
	}
	if def.Features.Privnet != (struct{}{}) {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "privnet", State: "on"})
	}
	if def.Features.PVSpinlock.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "pvspinlock", State: def.Features.PVSpinlock.State})
	}
	// Hyper-V features
	if def.Features.HyperV.Relaxed.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "hyperv_relaxed", State: def.Features.HyperV.Relaxed.State})
	}
	if def.Features.HyperV.VAPIC.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "hyperv_vapic", State: def.Features.HyperV.VAPIC.State})
	}
	if def.Features.HyperV.Spinlocks.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "hyperv_spinlocks", State: def.Features.HyperV.Spinlocks.State})
	}
	// KVM features
	if def.Features.KVM.Hidden.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "kvm_hidden", State: def.Features.KVM.Hidden.State})
	}
	if def.Features.KVM.HintDedicated.State != "" {
		hardware.HypervisorFeatures = append(hardware.HypervisorFeatures, HypervisorFeatureInfo{Name: "kvm_hint_dedicated", State: def.Features.KVM.HintDedicated.State})
	}

	// Populate Lifecycle Actions
	hardware.LifecycleActions = &LifecycleActionInfo{
		OnPoweroff:    def.OnPoweroff,
		OnReboot:      def.OnReboot,
		OnCrash:       def.OnCrash,
		OnLockFailure: def.OnLockfailure,
	}

	// Populate Clock Configuration
	if def.Clock.Offset != "" {
		hardware.ClockConfig = &ClockInfo{
			Offset: def.Clock.Offset,
		}
		// Note: Timers would need additional parsing if present in XML
	}

	// Populate Performance Events
	for _, event := range def.Perf.Event {
		hardware.PerfEvents = append(hardware.PerfEvents, PerfEventInfo{
			Name:  event.Name,
			Event: event.State, // Using State as Event for now
		})
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
	log.Debugf("Attempting to start domain %s on host %s", vmName, hostID)
	err = l.DomainCreate(domain)
	if err != nil {
		log.Errorf("Failed to start domain %s: %v", vmName, err)
		return fmt.Errorf("libvirt start failed for %s: %w", vmName, err)
	}
	log.Debugf("Successfully initiated start for domain %s", vmName)
	return nil
}

func (c *Connector) ShutdownDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	log.Debugf("Attempting to shutdown domain %s on host %s", vmName, hostID)
	err = l.DomainShutdown(domain)
	if err != nil {
		log.Errorf("Failed to shutdown domain %s: %v", vmName, err)
		return fmt.Errorf("libvirt shutdown failed for %s: %w", vmName, err)
	}
	log.Debugf("Successfully initiated shutdown for domain %s", vmName)
	return nil
}

func (c *Connector) RebootDomain(hostID, vmName string) error {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return err
	}
	log.Debugf("Attempting to reboot domain %s on host %s", vmName, hostID)
	err = l.DomainReboot(domain, 0)
	if err != nil {
		log.Errorf("Failed to reboot domain %s: %v", vmName, err)
		return fmt.Errorf("libvirt reboot failed for %s: %w", vmName, err)
	}
	log.Debugf("Successfully initiated reboot for domain %s", vmName)
	return nil
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
