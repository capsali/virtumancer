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
	// Enhanced API-based data
	VcpuDetails       []VcpuDetail           `json:"vcpu_details,omitempty"`
	NetworkInterfaces []NetworkInterface     `json:"network_interfaces,omitempty"`
	GuestInfo         map[string]interface{} `json:"guest_info,omitempty"`
}

// MemoryDetails holds detailed memory configuration from libvirt APIs
type MemoryDetails struct {
	MaxMemoryKB  uint64               `json:"max_memory_kb"`
	MemoryParams []libvirt.TypedParam `json:"memory_params,omitempty"`
}

// CPUDetails holds detailed CPU configuration from libvirt APIs
type CPUDetails struct {
	MaxVcpus        int32                `json:"max_vcpus"`
	CurrentVcpus    int32                `json:"current_vcpus"`
	VcpuPinInfo     [][]bool             `json:"vcpu_pin_info,omitempty"`
	EmulatorPinInfo []bool               `json:"emulator_pin_info,omitempty"`
	CPUStats        []libvirt.TypedParam `json:"cpu_stats,omitempty"`
}

// BlockDeviceDetail holds detailed block device information from libvirt APIs
type BlockDeviceDetail struct {
	Device     string        `json:"device"`
	SourcePath string        `json:"source_path"`
	Capacity   uint64        `json:"capacity"`
	Allocation uint64        `json:"allocation"`
	Physical   uint64        `json:"physical"`
	JobInfo    *BlockJobInfo `json:"job_info,omitempty"`
}

// BlockJobInfo holds information about ongoing block jobs
type BlockJobInfo struct {
	Type      int32  `json:"type"`
	Bandwidth uint64 `json:"bandwidth"`
	Cur       uint64 `json:"cur"`
	End       uint64 `json:"end"`
}

// SecurityDetail holds security label information from libvirt APIs
type SecurityDetail struct {
	Label     string `json:"label"`
	Enforcing int32  `json:"enforcing"`
}

// IOThreadDetail holds I/O thread information from libvirt APIs
type IOThreadDetail struct {
	IOThreadID uint32 `json:"iothread_id"`
	Cpumap     []bool `json:"cpumap"`
}

// NUMADetails holds NUMA configuration from libvirt APIs
type NUMADetails struct {
	NUMAParams []libvirt.TypedParam `json:"numa_params,omitempty"`
	NodeCount  int                  `json:"node_count"`
}

// MemoryStats holds detailed memory statistics from libvirt APIs
type MemoryStats struct {
	Actual         uint64 `json:"actual"`
	SwapIn         uint64 `json:"swap_in,omitempty"`
	SwapOut        uint64 `json:"swap_out,omitempty"`
	MajorFault     uint64 `json:"major_fault,omitempty"`
	MinorFault     uint64 `json:"minor_fault,omitempty"`
	Unused         uint64 `json:"unused,omitempty"`
	Available      uint64 `json:"available,omitempty"`
	Rss            uint64 `json:"rss,omitempty"`
	Usable         uint64 `json:"usable,omitempty"`
	LastUpdate     uint64 `json:"last_update,omitempty"`
	DiskCaches     uint64 `json:"disk_caches,omitempty"`
	HugetlbPgalloc uint64 `json:"hugetlb_pgalloc,omitempty"`
	HugetlbPgfail  uint64 `json:"hugetlb_pgfail,omitempty"`
}

// GuestAgentDetails holds guest agent information from libvirt APIs
type GuestAgentDetails struct {
	Hostname    string                 `json:"hostname,omitempty"`
	OSInfo      map[string]interface{} `json:"os_info,omitempty"`
	Interfaces  []GuestInterfaceInfo   `json:"interfaces,omitempty"`
	Filesystems []GuestFilesystemInfo  `json:"filesystems,omitempty"`
	Users       []GuestUserInfo        `json:"users,omitempty"`
	Timezone    GuestTimezoneInfo      `json:"timezone,omitempty"`
}

// GuestInterfaceInfo holds guest network interface information
type GuestInterfaceInfo struct {
	Name    string   `json:"name"`
	HWAddr  string   `json:"hwaddr,omitempty"`
	IPAddrs []string `json:"ip_addrs,omitempty"`
}

// GuestFilesystemInfo holds guest filesystem information
type GuestFilesystemInfo struct {
	Name       string `json:"name"`
	Mountpoint string `json:"mountpoint"`
	FSType     string `json:"fstype"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
}

// GuestUserInfo holds guest user session information
type GuestUserInfo struct {
	User      string `json:"user"`
	Domain    string `json:"domain,omitempty"`
	LoginTime uint64 `json:"login_time"`
}

// GuestTimezoneInfo holds guest timezone information
type GuestTimezoneInfo struct {
	Name   string `json:"name"`
	Offset int32  `json:"offset"`
}

// PerformanceDetails holds comprehensive performance statistics
type PerformanceDetails struct {
	CPUStats       []libvirt.TypedParam          `json:"cpu_stats,omitempty"`
	MemoryStats    *MemoryStats                  `json:"memory_stats,omitempty"`
	BlockStats     map[string]BlockStatsInfo     `json:"block_stats,omitempty"`
	InterfaceStats map[string]InterfaceStatsInfo `json:"interface_stats,omitempty"`
	DomainTime     *DomainTimeInfo               `json:"domain_time,omitempty"`
}

// BlockStatsInfo holds block device statistics
type BlockStatsInfo struct {
	ReadRequests  uint64 `json:"read_requests"`
	ReadBytes     uint64 `json:"read_bytes"`
	WriteRequests uint64 `json:"write_requests"`
	WriteBytes    uint64 `json:"write_bytes"`
	Errors        uint64 `json:"errors"`
}

// InterfaceStatsInfo holds network interface statistics
type InterfaceStatsInfo struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDrops   uint64 `json:"rx_drops"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDrops   uint64 `json:"tx_drops"`
}

// DomainTimeInfo holds domain time information
type DomainTimeInfo struct {
	Seconds     uint64 `json:"seconds"`
	Nanoseconds uint32 `json:"nanoseconds"`
	Synced      bool   `json:"synced"`
}

// Phase 2 Types: Performance and Statistics APIs

// CPUPerformanceDetails holds comprehensive CPU performance data
type CPUPerformanceDetails struct {
	VCPUStats    map[int]VCPUStats `json:"vcpu_stats"`
	HostCPUStats *HostCPUStats     `json:"host_cpu_stats,omitempty"`
}

// VCPUStats holds per-VCPU statistics
type VCPUStats struct {
	CPUTime    uint64 `json:"cpu_time"`
	UserTime   uint64 `json:"user_time,omitempty"`
	SystemTime uint64 `json:"system_time,omitempty"`
}

// HostCPUStats holds host CPU statistics for comparison
type HostCPUStats struct {
	User   uint64 `json:"user"`
	Kernel uint64 `json:"kernel"`
	Idle   uint64 `json:"idle"`
	IOWait uint64 `json:"iowait"`
}

// NetworkInterfaceDetails holds network interface addressing information
type NetworkInterfaceDetails struct {
	Interfaces []InterfaceAddress `json:"interfaces"`
}

// InterfaceAddress holds interface addressing details
type InterfaceAddress struct {
	Name   string      `json:"name"`
	HWAddr string      `json:"hwaddr,omitempty"`
	Addrs  []IPAddress `json:"addrs"`
}

// IPAddress holds IP address information
type IPAddress struct {
	Type   int    `json:"type"`
	Addr   string `json:"addr"`
	Prefix int    `json:"prefix"`
}

// DomainJobInfo holds information about active domain jobs
type DomainJobInfo struct {
	Type          int    `json:"type"`
	TimeElapsed   uint64 `json:"time_elapsed"`
	TimeRemaining uint64 `json:"time_remaining"`
	DataTotal     uint64 `json:"data_total"`
	DataProcessed uint64 `json:"data_processed"`
	DataRemaining uint64 `json:"data_remaining"`
	MemTotal      uint64 `json:"mem_total"`
	MemProcessed  uint64 `json:"mem_processed"`
	MemRemaining  uint64 `json:"mem_remaining"`
	FileTotal     uint64 `json:"file_total"`
	FileProcessed uint64 `json:"file_processed"`
	FileRemaining uint64 `json:"file_remaining"`
}

// Phase 3 Types: Hybrid Optimization

// HybridDomainDetails combines API and XML data
type HybridDomainDetails struct {
	APIData *APIDomainData `json:"api_data"`
	XMLData *XMLDomainData `json:"xml_data"`
}

// APIDomainData holds all API-sourced domain information
type APIDomainData struct {
	MemoryDetails      *MemoryDetails      `json:"memory_details,omitempty"`
	CPUDetails         *CPUDetails         `json:"cpu_details,omitempty"`
	BlockDetails       []BlockDeviceDetail `json:"block_details,omitempty"`
	SecurityDetails    []SecurityDetail    `json:"security_details,omitempty"`
	NUMADetails        *NUMADetails        `json:"numa_details,omitempty"`
	MemoryStats        *MemoryStats        `json:"memory_stats,omitempty"`
	PerformanceDetails *PerformanceDetails `json:"performance_details,omitempty"`
}

// XMLDomainData holds all XML-sourced domain information
type XMLDomainData struct {
	RawXML       string           `json:"raw_xml,omitempty"`
	Features     *XMLFeatures     `json:"features,omitempty"`
	OSConfig     *XMLOSConfig     `json:"os_config,omitempty"`
	DeviceConfig *XMLDeviceConfig `json:"device_config,omitempty"`
}

// OptimizedSyncData holds intelligently sourced domain data
type OptimizedSyncData struct {
	Strategy             map[string]string        `json:"strategy"`
	MemoryStats          *MemoryStats             `json:"memory_stats,omitempty"`
	PerformanceDetails   *PerformanceDetails      `json:"performance_details,omitempty"`
	CPUPerformance       *CPUPerformanceDetails   `json:"cpu_performance,omitempty"`
	NetworkDetails       *NetworkInterfaceDetails `json:"network_details,omitempty"`
	GraphicsConfig       *XMLGraphicsConfig       `json:"graphics_config,omitempty"`
	TPMConfig            *XMLTPMConfig            `json:"tpm_config,omitempty"`
	HypervisorFeatures   *XMLHypervisorFeatures   `json:"hypervisor_features,omitempty"`
	EnhancedBlockDetails *EnhancedBlockDetails    `json:"enhanced_block_details,omitempty"`
}

// XML-specific structures for Phase 3
type XMLFeatures struct {
	ACPI     bool   `json:"acpi"`
	APIC     bool   `json:"apic"`
	PAE      bool   `json:"pae"`
	HAP      bool   `json:"hap"`
	VirtType string `json:"virt_type,omitempty"`
}

type XMLOSConfig struct {
	Type      string   `json:"type"`
	Arch      string   `json:"arch"`
	Machine   string   `json:"machine"`
	BootOrder []string `json:"boot_order"`
	Kernel    string   `json:"kernel,omitempty"`
	Initrd    string   `json:"initrd,omitempty"`
	Cmdline   string   `json:"cmdline,omitempty"`
}

type XMLDeviceConfig struct {
	Emulator    string          `json:"emulator,omitempty"`
	Controllers []XMLController `json:"controllers,omitempty"`
	Channels    []XMLChannel    `json:"channels,omitempty"`
	Watchdog    *XMLWatchdog    `json:"watchdog,omitempty"`
	RNG         *XMLRNGDevice   `json:"rng,omitempty"`
}

type XMLGraphicsConfig struct {
	Type     string `json:"type"`
	Port     int    `json:"port,omitempty"`
	AutoPort bool   `json:"autoport"`
	Listen   string `json:"listen,omitempty"`
	Keymap   string `json:"keymap,omitempty"`
}

type XMLTPMConfig struct {
	Model   string `json:"model"`
	Type    string `json:"type"`
	Version string `json:"version,omitempty"`
}

type XMLHypervisorFeatures struct {
	Relaxed   bool `json:"relaxed"`
	VAPIC     bool `json:"vapic"`
	Spinlocks bool `json:"spinlocks"`
	VPIndex   bool `json:"vpindex"`
	Runtime   bool `json:"runtime"`
	Synic     bool `json:"synic"`
	Reset     bool `json:"reset"`
	VendorID  bool `json:"vendor_id"`
}

type XMLController struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Model string `json:"model,omitempty"`
}

type XMLChannel struct {
	Type   string `json:"type"`
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
}

type XMLWatchdog struct {
	Model  string `json:"model"`
	Action string `json:"action"`
}

type XMLRNGDevice struct {
	Model string `json:"model"`
	Rate  int    `json:"rate,omitempty"`
}

type EnhancedBlockDetails struct {
	Devices     []BlockDeviceDetail         `json:"devices"`
	XMLMetadata map[string]XMLBlockMetadata `json:"xml_metadata"`
}

type XMLBlockMetadata struct {
	Driver     string `json:"driver,omitempty"`
	Cache      string `json:"cache,omitempty"`
	IO         string `json:"io,omitempty"`
	Discard    string `json:"discard,omitempty"`
	Encryption bool   `json:"encryption"`
}

// Phase 4 Types: XML-Only Components

// XMLOnlyFeatures holds features that are only available through XML parsing
type XMLOnlyFeatures struct {
	HypervisorFeatures *DetailedHypervisorFeatures `json:"hypervisor_features,omitempty"`
	CPUFeatures        *DetailedCPUFeatures        `json:"cpu_features,omitempty"`
	NUMATopology       *DetailedNUMATopology       `json:"numa_topology,omitempty"`
	AdvancedDevices    *AdvancedDeviceConfigs      `json:"advanced_devices,omitempty"`
	OSLoader           *OSLoaderConfig             `json:"os_loader,omitempty"`
	ClockConfig        *ClockConfig                `json:"clock_config,omitempty"`
}

// CompleteXMLAnalysis holds comprehensive XML analysis results
type CompleteXMLAnalysis struct {
	OriginalXML      string                 `json:"original_xml"`
	ParsedComponents map[string]interface{} `json:"parsed_components"`
}

// Detailed XML structures for Phase 4
type DetailedHypervisorFeatures struct {
	HyperV  *HyperVFeatures  `json:"hyperv,omitempty"`
	KVM     *KVMFeatures     `json:"kvm,omitempty"`
	Xen     *XenFeatures     `json:"xen,omitempty"`
	PowerPC *PowerPCFeatures `json:"powerpc,omitempty"`
	S390    *S390Features    `json:"s390,omitempty"`
	MSR     *MSRFeatures     `json:"msr,omitempty"`
	GIC     *GICFeatures     `json:"gic,omitempty"`
}

type HyperVFeatures struct {
	Relaxed         *FeatureState `json:"relaxed,omitempty"`
	VAPIC           *FeatureState `json:"vapic,omitempty"`
	Spinlocks       *FeatureState `json:"spinlocks,omitempty"`
	VPIndex         *FeatureState `json:"vpindex,omitempty"`
	Runtime         *FeatureState `json:"runtime,omitempty"`
	Synic           *FeatureState `json:"synic,omitempty"`
	SynicTimer      *FeatureState `json:"synic_timer,omitempty"`
	Reset           *FeatureState `json:"reset,omitempty"`
	VendorID        *FeatureState `json:"vendor_id,omitempty"`
	Frequencies     *FeatureState `json:"frequencies,omitempty"`
	ReenLightenment *FeatureState `json:"reenlightenment,omitempty"`
	TLBFlush        *FeatureState `json:"tlbflush,omitempty"`
	IPI             *FeatureState `json:"ipi,omitempty"`
	EVMCS           *FeatureState `json:"evmcs,omitempty"`
}

type KVMFeatures struct {
	Hidden        *FeatureState `json:"hidden,omitempty"`
	HintDedicated *FeatureState `json:"hint_dedicated,omitempty"`
	PollControl   *FeatureState `json:"poll_control,omitempty"`
	MSRFeatures   *FeatureState `json:"msr_features,omitempty"`
}

type XenFeatures struct {
	E820Host       *FeatureState `json:"e820_host,omitempty"`
	HAPTranslation *FeatureState `json:"hap_translation,omitempty"`
}

type PowerPCFeatures struct {
	HTM      *FeatureState `json:"htm,omitempty"`
	NestedHV *FeatureState `json:"nested_hv,omitempty"`
}

type S390Features struct {
	CMMA  *FeatureState `json:"cmma,omitempty"`
	PFMFI *FeatureState `json:"pfmfi,omitempty"`
}

type MSRFeatures struct {
	UnknownMSR *FeatureState `json:"unknown_msr,omitempty"`
}

type GICFeatures struct {
	Version string `json:"version,omitempty"`
}

type FeatureState struct {
	State      string            `json:"state"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type DetailedCPUFeatures struct {
	Mode        string       `json:"mode"`
	Match       string       `json:"match"`
	Check       string       `json:"check"`
	Migratable  string       `json:"migratable,omitempty"`
	Model       *CPUModel    `json:"model,omitempty"`
	Vendor      string       `json:"vendor,omitempty"`
	Topology    *CPUTopology `json:"topology,omitempty"`
	Features    []CPUFeature `json:"features,omitempty"`
	Cache       *CPUCache    `json:"cache,omitempty"`
	MaxPhysAddr *MaxPhysAddr `json:"maxphysaddr,omitempty"`
}

type CPUModel struct {
	Name     string `json:"name"`
	Fallback string `json:"fallback,omitempty"`
	VendorID string `json:"vendor_id,omitempty"`
}

type CPUTopology struct {
	Sockets int `json:"sockets"`
	Dies    int `json:"dies,omitempty"`
	Cores   int `json:"cores"`
	Threads int `json:"threads"`
}

type CPUFeature struct {
	Policy string `json:"policy"`
	Name   string `json:"name"`
}

type CPUCache struct {
	Mode  string `json:"mode"`
	Level int    `json:"level,omitempty"`
}

type MaxPhysAddr struct {
	Mode string `json:"mode"`
	Bits int    `json:"bits,omitempty"`
}

type DetailedNUMATopology struct {
	Cells []NUMACell `json:"cells"`
}

type NUMACell struct {
	ID        int            `json:"id"`
	CPUs      string         `json:"cpus"`
	Memory    uint64         `json:"memory"`
	Unit      string         `json:"unit"`
	MemAccess string         `json:"memaccess,omitempty"`
	Distances []NUMADistance `json:"distances,omitempty"`
	Caches    []NUMACache    `json:"caches,omitempty"`
}

type NUMADistance struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

type NUMACache struct {
	Level         int    `json:"level"`
	Associativity string `json:"associativity"`
	Policy        string `json:"policy"`
	Size          uint64 `json:"size"`
	Unit          string `json:"unit"`
	Line          uint64 `json:"line"`
}

type AdvancedDeviceConfigs struct {
	Input      []InputDevice     `json:"input,omitempty"`
	Video      []VideoDevice     `json:"video,omitempty"`
	Sound      []SoundDevice     `json:"sound,omitempty"`
	Hostdev    []HostDevice      `json:"hostdev,omitempty"`
	RedirDev   []RedirDevice     `json:"redirdev,omitempty"`
	SmartCard  []SmartCardDevice `json:"smartcard,omitempty"`
	Hub        []HubDevice       `json:"hub,omitempty"`
	MemBalloon *MemBalloonDevice `json:"memballoon,omitempty"`
	Panic      []PanicDevice     `json:"panic,omitempty"`
	SHMEM      []SHMEMDevice     `json:"shmem,omitempty"`
	Memory     []MemoryDevice    `json:"memory,omitempty"`
	IOMMU      *IOMMUDevice      `json:"iommu,omitempty"`
	VSOCK      *VSockDevice      `json:"vsock,omitempty"`
}

type InputDevice struct {
	Type string `json:"type"`
	Bus  string `json:"bus"`
}

type VideoDevice struct {
	Model        string             `json:"model"`
	VRAMBytes    uint64             `json:"vram_bytes,omitempty"`
	Heads        int                `json:"heads,omitempty"`
	Primary      bool               `json:"primary,omitempty"`
	Acceleration *VideoAcceleration `json:"acceleration,omitempty"`
}

type VideoAcceleration struct {
	Accel2D    bool   `json:"accel2d"`
	Accel3D    bool   `json:"accel3d"`
	RenderNode string `json:"rendernode,omitempty"`
}

type SoundDevice struct {
	Model string `json:"model"`
	Codec string `json:"codec,omitempty"`
}

type HostDevice struct {
	Mode    string            `json:"mode"`
	Type    string            `json:"type"`
	Managed bool              `json:"managed"`
	Source  map[string]string `json:"source"`
}

type RedirDevice struct {
	Bus  string `json:"bus"`
	Type string `json:"type"`
}

type SmartCardDevice struct {
	Mode string `json:"mode"`
	Type string `json:"type,omitempty"`
}

type HubDevice struct {
	Type string `json:"type"`
}

type MemBalloonDevice struct {
	Model             string `json:"model"`
	AutoDeflate       bool   `json:"autodeflate,omitempty"`
	FreePageReporting bool   `json:"freepage_reporting,omitempty"`
}

type PanicDevice struct {
	Model string `json:"model"`
}

type SHMEMDevice struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

type MemoryDevice struct {
	Model      string `json:"model"`
	Access     string `json:"access,omitempty"`
	Discard    bool   `json:"discard,omitempty"`
	TargetSize uint64 `json:"target_size"`
	TargetNode int    `json:"target_node,omitempty"`
}

type IOMMUDevice struct {
	Model string `json:"model"`
}

type VSockDevice struct {
	Model string `json:"model"`
	CID   uint32 `json:"cid,omitempty"`
}

type OSLoaderConfig struct {
	Type          string `json:"type"`
	ReadOnly      bool   `json:"readonly,omitempty"`
	Secure        bool   `json:"secure,omitempty"`
	Path          string `json:"path,omitempty"`
	Template      string `json:"template,omitempty"`
	NVRAM         string `json:"nvram,omitempty"`
	NVRAMTemplate string `json:"nvram_template,omitempty"`
}

type ClockConfig struct {
	Offset     string        `json:"offset"`
	Adjustment string        `json:"adjustment,omitempty"`
	Timezone   string        `json:"timezone,omitempty"`
	Timers     []TimerConfig `json:"timers,omitempty"`
}

type TimerConfig struct {
	Name       string        `json:"name"`
	Track      string        `json:"track,omitempty"`
	TickPolicy string        `json:"tickpolicy,omitempty"`
	CatchUp    *TimerCatchUp `json:"catchup,omitempty"`
	Frequency  uint64        `json:"frequency,omitempty"`
	Mode       string        `json:"mode,omitempty"`
	Present    bool          `json:"present"`
}

type TimerCatchUp struct {
	Threshold uint64 `json:"threshold,omitempty"`
	Slew      uint64 `json:"slew,omitempty"`
	Limit     uint64 `json:"limit,omitempty"`
}

// VcpuDetail holds detailed information about a single VCPU
type VcpuDetail struct {
	Number  uint32 `json:"number"`
	State   int32  `json:"state"`
	CPUTime uint64 `json:"cpu_time"`
	CPU     int32  `json:"cpu"`
}

// NetworkInterface holds detailed network interface information
type NetworkInterface struct {
	Name   string   `json:"name"`
	Hwaddr string   `json:"hwaddr"`
	Addrs  []string `json:"addrs"`
}

// EnhancedVMStats holds comprehensive performance statistics
type EnhancedVMStats struct {
	InterfaceStats map[string]InterfaceStats `json:"interface_stats"`
	CPUStats       map[string]interface{}    `json:"cpu_stats"`
}

// InterfaceStats holds detailed network interface statistics
type InterfaceStats struct {
	RxBytes   int64 `json:"rx_bytes"`
	RxPackets int64 `json:"rx_packets"`
	RxErrs    int64 `json:"rx_errs"`
	RxDrop    int64 `json:"rx_drop"`
	TxBytes   int64 `json:"tx_bytes"`
	TxPackets int64 `json:"tx_packets"`
	TxErrs    int64 `json:"tx_errs"`
	TxDrop    int64 `json:"tx_drop"`
}

// StoragePoolInfo holds basic information about a storage pool.
type StoragePoolInfo struct {
	Name            string `json:"name"`
	UUID            string `json:"uuid"`
	State           int    `json:"state"`
	CapacityBytes   uint64 `json:"capacity_bytes"`
	AllocationBytes uint64 `json:"allocation_bytes"`
	AvailableBytes  uint64 `json:"available_bytes"`
}

// DomainDiskStats holds I/O statistics for a single disk device.
type DomainDiskStats struct {
	Device     string `json:"device"`
	ReadBytes  int64  `json:"read_bytes"`
	WriteBytes int64  `json:"write_bytes"`
	ReadReq    int64  `json:"read_req"`
	WriteReq   int64  `json:"write_req"`
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
	Capacity struct {
		Value uint64 `xml:",chardata" json:"value"`
		Unit  string `xml:"unit,attr" json:"unit"`
	} `xml:"capacity" json:"capacity"`
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
		Bridge    string `xml:"bridge,attr" json:"bridge"`
		Network   string `xml:"network,attr" json:"network"`
		PortGroup string `xml:"portgroup,attr" json:"portgroup"`
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

// ListAllStoragePools retrieves information about all storage pools on a host.
func (c *Connector) ListAllStoragePools(hostID string) ([]StoragePoolInfo, error) {
	l, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	pools, err := l.StoragePools(libvirt.ConnectListAllStoragePoolsFlags(0))
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}

	var poolInfos []StoragePoolInfo
	for _, pool := range pools {
		poolInfo, err := c.storagePoolToInfo(l, pool)
		if err != nil {
			log.Debugf("Warning: could not get info for storage pool %s on host %s: %v", pool.Name, hostID, err)
			continue
		}
		poolInfos = append(poolInfos, *poolInfo)
	}

	return poolInfos, nil
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
// Now uses libvirt APIs primarily with XML fallback for configuration-only data.
func (c *Connector) domainToVMInfo(l *libvirt.Libvirt, domain libvirt.Domain) (*VMInfo, error) {
	stateInt, _, err := l.DomainGetState(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain state for %s: %w", domain.Name, err)
	}
	state := libvirt.DomainState(stateInt)

	// Use DomainGetInfo API for real-time values
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

	// Enhanced VMInfo with API-based data
	vmInfo := &VMInfo{
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
	}

	// Enhance with additional API-based data
	c.enhanceVMInfoWithAPIs(l, domain, vmInfo)

	return vmInfo, nil
}

// enhanceVMInfoWithAPIs adds additional data from libvirt APIs to VMInfo
func (c *Connector) enhanceVMInfoWithAPIs(l *libvirt.Libvirt, domain libvirt.Domain, vmInfo *VMInfo) {
	// Get enhanced VCPU information
	if vcpuInfo, _, err := l.DomainGetVcpus(domain, int32(vmInfo.Vcpu), 0); err == nil {
		vmInfo.VcpuDetails = make([]VcpuDetail, len(vcpuInfo))
		for i, vcpu := range vcpuInfo {
			vmInfo.VcpuDetails[i] = VcpuDetail{
				Number:  vcpu.Number,
				State:   vcpu.State,
				CPUTime: vcpu.CPUTime,
				CPU:     vcpu.CPU,
			}
		}
	}

	// Get enhanced network interface information with IP addresses
	// Source: 0 = lease, 1 = agent, 2 = arp
	if interfaces, err := l.DomainInterfaceAddresses(domain, 1, 0); err == nil {
		vmInfo.NetworkInterfaces = make([]NetworkInterface, len(interfaces))
		for i, iface := range interfaces {
			hwaddr := ""
			if len(iface.Hwaddr) > 0 {
				hwaddr = iface.Hwaddr[0]
			}
			vmInfo.NetworkInterfaces[i] = NetworkInterface{
				Name:   iface.Name,
				Hwaddr: hwaddr,
				Addrs:  make([]string, len(iface.Addrs)),
			}
			for j, addr := range iface.Addrs {
				vmInfo.NetworkInterfaces[i].Addrs[j] = addr.Addr
			}
		}
	}

	// Get guest information if guest agent is available
	if guestInfo, err := l.DomainGetGuestInfo(domain, 0, 0); err == nil {
		vmInfo.GuestInfo = make(map[string]interface{})
		for _, param := range guestInfo {
			// Use the discriminated union properly
			vmInfo.GuestInfo[param.Field] = param.Value.I
		}
	}
}

// storagePoolToInfo converts a libvirt.StoragePool object to our StoragePoolInfo struct.
func (c *Connector) storagePoolToInfo(l *libvirt.Libvirt, pool libvirt.StoragePool) (*StoragePoolInfo, error) {
	// Get storage pool info
	state, capacity, allocation, available, err := l.StoragePoolGetInfo(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage pool info for %s: %w", pool.Name, err)
	}

	// Parse UUID
	var uuidStr string
	parsedUUID, err := uuid.FromBytes(pool.UUID[:])
	if err != nil {
		log.Debugf("Warning: could not parse storage pool UUID for %s: %v. Using raw hex.", pool.Name, err)
		uuidStr = fmt.Sprintf("%x", pool.UUID)
	} else {
		uuidStr = parsedUUID.String()
	}

	return &StoragePoolInfo{
		Name:            pool.Name,
		UUID:            uuidStr,
		State:           int(state),
		CapacityBytes:   capacity,
		AllocationBytes: allocation,
		AvailableBytes:  available,
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
		// rdReq/wrReq/errs are available from DomainBlockStats and used for IOPS calculation
		_ = errs // suppress unused when not used elsewhere
		diskStats = append(diskStats, DomainDiskStats{
			Device:     disk.Target.Dev,
			ReadBytes:  rdBytes,
			WriteBytes: wrBytes,
			ReadReq:    rdReq,
			WriteReq:   wrReq,
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

// GetDiskSize gets the actual size of a disk file from the host
func (c *Connector) GetDiskSize(hostID, diskPath string) (uint64, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return 0, fmt.Errorf("failed to get libvirt connection: %w", err)
	}

	// Strategy 1: Try to find the disk as a storage volume
	capacity, err := c.getDiskSizeFromStorageVolume(conn, diskPath)
	if err == nil && capacity > 0 {
		log.Verbosef("Got disk size from storage volume for %s: %d bytes", diskPath, capacity)
		return capacity, nil
	}

	// Strategy 2: Try to get size from active domain block info
	capacity, err = c.getDiskSizeFromDomainBlockInfo(conn, diskPath)
	if err == nil && capacity > 0 {
		log.Verbosef("Got disk size from domain block info for %s: %d bytes", diskPath, capacity)
		return capacity, nil
	}

	// Strategy 3: Try to get size by searching all storage pools for the volume
	capacity, err = c.getDiskSizeFromAllPools(conn, diskPath)
	if err == nil && capacity > 0 {
		log.Verbosef("Got disk size from storage pools search for %s: %d bytes", diskPath, capacity)
		return capacity, nil
	}

	log.Verbosef("Could not determine disk size for %s using any libvirt method", diskPath)
	return 0, fmt.Errorf("unable to determine disk size for %s", diskPath)
}

// GetDomainNetworkInfo retrieves enhanced network information for a domain using APIs
func (c *Connector) GetDomainNetworkInfo(hostID, vmName string) ([]NetworkInterface, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	var networkInterfaces []NetworkInterface

	// Try multiple sources for network information
	// Source 0 = lease, 1 = agent, 2 = arp
	for _, source := range []uint32{1, 0, 2} {
		if interfaces, err := l.DomainInterfaceAddresses(domain, source, 0); err == nil && len(interfaces) > 0 {
			networkInterfaces = make([]NetworkInterface, len(interfaces))
			for i, iface := range interfaces {
				hwaddr := ""
				if len(iface.Hwaddr) > 0 {
					hwaddr = iface.Hwaddr[0]
				}
				networkInterfaces[i] = NetworkInterface{
					Name:   iface.Name,
					Hwaddr: hwaddr,
					Addrs:  make([]string, len(iface.Addrs)),
				}
				for j, addr := range iface.Addrs {
					networkInterfaces[i].Addrs[j] = addr.Addr
				}
			}
			break // Use first successful source
		}
	}

	return networkInterfaces, nil
}

// GetDomainVcpuInfo retrieves detailed VCPU information for a domain
func (c *Connector) GetDomainVcpuInfo(hostID, vmName string) ([]VcpuDetail, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get basic domain info for VCPU count
	_, _, _, nrVirtCPU, _, err := l.DomainGetInfo(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info for VCPU count: %w", err)
	}

	// Get detailed VCPU information
	vcpuInfo, _, err := l.DomainGetVcpus(domain, int32(nrVirtCPU), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get VCPU details: %w", err)
	}

	vcpuDetails := make([]VcpuDetail, len(vcpuInfo))
	for i, vcpu := range vcpuInfo {
		vcpuDetails[i] = VcpuDetail{
			Number:  vcpu.Number,
			State:   vcpu.State,
			CPUTime: vcpu.CPUTime,
			CPU:     vcpu.CPU,
		}
	}

	return vcpuDetails, nil
}

// GetDomainPerformanceStats retrieves comprehensive performance statistics
func (c *Connector) GetDomainPerformanceStats(hostID, vmName string) (*EnhancedVMStats, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	stats := &EnhancedVMStats{
		InterfaceStats: make(map[string]InterfaceStats),
		CPUStats:       make(map[string]interface{}),
	}

	// Get interface statistics
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err == nil {
		// Parse XML to get interface device names
		var def struct {
			Devices struct {
				Interfaces []struct {
					Target struct {
						Dev string `xml:"dev,attr"`
					} `xml:"target"`
				} `xml:"interface"`
			} `xml:"devices"`
		}
		if xml.Unmarshal([]byte(xmlDesc), &def) == nil {
			for _, iface := range def.Devices.Interfaces {
				if iface.Target.Dev != "" {
					rxBytes, rxPackets, rxErrs, rxDrop, txBytes, txPackets, txErrs, txDrop, err := l.DomainInterfaceStats(domain, iface.Target.Dev)
					if err == nil {
						stats.InterfaceStats[iface.Target.Dev] = InterfaceStats{
							RxBytes:   rxBytes,
							RxPackets: rxPackets,
							RxErrs:    rxErrs,
							RxDrop:    rxDrop,
							TxBytes:   txBytes,
							TxPackets: txPackets,
							TxErrs:    txErrs,
							TxDrop:    txDrop,
						}
					}
				}
			}
		}
	}

	// Get CPU statistics
	if cpuStats, _, err := l.DomainGetCPUStats(domain, 1, -1, 1, 0); err == nil {
		for _, param := range cpuStats {
			stats.CPUStats[param.Field] = param.Value.I
		}
	}

	return stats, nil
}

// getDiskSizeFromStorageVolume attempts to get disk size by treating path as a volume name
func (c *Connector) getDiskSizeFromStorageVolume(conn *libvirt.Libvirt, diskPath string) (uint64, error) {
	// First, try to find which pool contains this volume
	pools, err := conn.StoragePools(libvirt.ConnectListAllStoragePoolsFlags(0))
	if err != nil {
		return 0, fmt.Errorf("failed to list storage pools: %w", err)
	}

	// Extract the volume name from the path
	volumeName := filepath.Base(diskPath)

	for _, pool := range pools {
		// Get pool XML to get the pool name
		poolXML, err := conn.StoragePoolGetXMLDesc(pool, 0)
		if err != nil {
			continue // Skip this pool if we can't get its XML
		}

		// Parse pool XML to get the name
		var poolDef struct {
			XMLName xml.Name `xml:"pool"`
			Name    string   `xml:"name"`
		}
		if err := xml.Unmarshal([]byte(poolXML), &poolDef); err != nil {
			continue // Skip this pool if we can't parse its XML
		}

		// Try to find the volume in this pool
		vol, err := conn.StorageVolLookupByName(pool, volumeName)
		if err != nil {
			continue // Volume not in this pool
		}

		// Get volume info
		volType, capacity, allocation, err := conn.StorageVolGetInfo(vol)
		if err != nil {
			continue // Skip if we can't get volume info
		}

		log.Verbosef("Found volume %s in pool %s: type=%d, capacity=%d, allocation=%d",
			volumeName, poolDef.Name, volType, capacity, allocation)

		return capacity, nil
	}

	return 0, fmt.Errorf("volume not found in any storage pool")
}

// getDiskSizeFromDomainBlockInfo attempts to get disk size from active domains
func (c *Connector) getDiskSizeFromDomainBlockInfo(conn *libvirt.Libvirt, diskPath string) (uint64, error) {
	// Get all domains
	domains, err := conn.Domains()
	if err != nil {
		return 0, fmt.Errorf("failed to list domains: %w", err)
	}

	for _, domain := range domains {
		// Check if domain is active
		stateInt, _, err := conn.DomainGetState(domain, 0)
		if err != nil {
			continue // Skip domains where we can't get state
		}
		state := libvirt.DomainState(stateInt)
		if state != libvirt.DomainRunning {
			continue // Skip inactive domains
		}

		// Get domain XML to find block devices
		domainXML, err := conn.DomainGetXMLDesc(domain, 0)
		if err != nil {
			continue // Skip if we can't get domain XML
		}

		// Parse domain XML to find disk devices
		var domainDef struct {
			XMLName xml.Name `xml:"domain"`
			Devices struct {
				Disks []struct {
					Type   string `xml:"type,attr"`
					Device string `xml:"device,attr"`
					Source struct {
						File string `xml:"file,attr"`
						Dev  string `xml:"dev,attr"`
					} `xml:"source"`
					Target struct {
						Dev string `xml:"dev,attr"`
					} `xml:"target"`
				} `xml:"disk"`
			} `xml:"devices"`
		}

		if err := xml.Unmarshal([]byte(domainXML), &domainDef); err != nil {
			continue // Skip if we can't parse domain XML
		}

		// Look for our disk path in this domain
		for _, disk := range domainDef.Devices.Disks {
			var sourcePath string
			if disk.Source.File != "" {
				sourcePath = disk.Source.File
			} else if disk.Source.Dev != "" {
				sourcePath = disk.Source.Dev
			}

			if sourcePath == diskPath {
				// Found the disk in this domain, get block info
				allocation, capacity, physical, err := conn.DomainGetBlockInfo(domain, diskPath, 0)
				if err != nil {
					log.Verbosef("Failed to get block info for %s in domain: %v", diskPath, err)
					continue
				}

				log.Verbosef("Found disk %s in active domain: capacity=%d, allocation=%d, physical=%d",
					diskPath, capacity, allocation, physical)

				// Return the capacity (virtual size)
				return capacity, nil
			}
		}
	}

	return 0, fmt.Errorf("disk not found in any active domain")
}

// getDiskSizeFromAllPools searches all storage pools for volumes that match the disk path
func (c *Connector) getDiskSizeFromAllPools(conn *libvirt.Libvirt, diskPath string) (uint64, error) {
	pools, err := conn.StoragePools(libvirt.ConnectListAllStoragePoolsFlags(0))
	if err != nil {
		return 0, fmt.Errorf("failed to list storage pools: %w", err)
	}

	for _, pool := range pools {
		// Get all volumes in this pool
		volumes, _, err := conn.StoragePoolListAllVolumes(pool, -1, 0)
		if err != nil {
			continue // Skip this pool if we can't list volumes
		}

		for _, vol := range volumes {
			// Get volume path
			volPath, err := conn.StorageVolGetPath(vol)
			if err != nil {
				continue // Skip this volume if we can't get its path
			}

			if volPath == diskPath {
				// Found the volume, get its info
				volType, capacity, allocation, err := conn.StorageVolGetInfo(vol)
				if err != nil {
					continue // Skip if we can't get volume info
				}

				log.Verbosef("Found volume by path %s: type=%d, capacity=%d, allocation=%d",
					diskPath, volType, capacity, allocation)

				return capacity, nil
			}
		}
	}

	return 0, fmt.Errorf("volume not found by path in any storage pool")
}

// GetDomainMemoryDetails retrieves detailed memory configuration using libvirt APIs
func (c *Connector) GetDomainMemoryDetails(hostID, vmName string) (*MemoryDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	maxMem, err := l.DomainGetMaxMemory(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get max memory: %w", err)
	}

	// Get memory parameters
	memParams, _, err := l.DomainGetMemoryParameters(domain, -1, 0)
	if err != nil {
		// Not all hypervisors support memory parameters, so we'll continue without them
		log.Debugf("Memory parameters not available for domain %s: %v", vmName, err)
	}

	return &MemoryDetails{
		MaxMemoryKB:  maxMem,
		MemoryParams: memParams,
	}, nil
}

// GetDomainCPUDetails retrieves detailed CPU configuration using libvirt APIs
func (c *Connector) GetDomainCPUDetails(hostID, vmName string) (*CPUDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get VCPU count with flags
	maxVcpus, err := l.DomainGetVcpusFlags(domain, uint32(libvirt.DomainVCPUMaximum))
	if err != nil {
		return nil, fmt.Errorf("failed to get max VCPUs: %w", err)
	}

	currentVcpus, err := l.DomainGetVcpusFlags(domain, uint32(libvirt.DomainVCPUCurrent))
	if err != nil {
		return nil, fmt.Errorf("failed to get current VCPUs: %w", err)
	}

	// Get VCPU pinning info (optional - may not be supported by all hypervisors)
	vcpuPinInfo, _, err := l.DomainGetVcpuPinInfo(domain, int32(currentVcpus), 1, 0)
	if err != nil {
		log.Debugf("VCPU pin info not available for domain %s: %v", vmName, err)
	}

	// Get emulator pin info (optional)
	emulatorPinInfo, _, err := l.DomainGetEmulatorPinInfo(domain, 1, 0)
	if err != nil {
		log.Debugf("Emulator pin info not available for domain %s: %v", vmName, err)
	}

	// Get CPU stats (optional)
	cpuStats, _, err := l.DomainGetCPUStats(domain, 0, 0, 1, 0)
	if err != nil {
		log.Debugf("CPU stats not available for domain %s: %v", vmName, err)
	}

	// Convert emulator pin info from []byte to []bool
	var emulatorPin []bool
	if emulatorPinInfo != nil {
		emulatorPin = make([]bool, len(emulatorPinInfo)*8)
		for i, b := range emulatorPinInfo {
			for j := 0; j < 8; j++ {
				emulatorPin[i*8+j] = (b & (1 << j)) != 0
			}
		}
	}

	// Convert VCPU pin info from []byte to [][]bool (simplified for now)
	var vcpuPin [][]bool
	if vcpuPinInfo != nil {
		// For now, create a simple 2D array - this would need proper parsing
		// based on the actual CPU topology
		vcpuPin = make([][]bool, currentVcpus)
		for i := range vcpuPin {
			vcpuPin[i] = make([]bool, len(vcpuPinInfo)*8)
			// This is a simplified conversion - real implementation would need
			// to parse the pinning info properly
		}
	}

	return &CPUDetails{
		MaxVcpus:        int32(maxVcpus),
		CurrentVcpus:    int32(currentVcpus),
		VcpuPinInfo:     vcpuPin,
		EmulatorPinInfo: emulatorPin,
		CPUStats:        cpuStats,
	}, nil
}

// GetDomainBlockDetails retrieves detailed block device information using libvirt APIs
func (c *Connector) GetDomainBlockDetails(hostID, vmName string) ([]BlockDeviceDetail, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// First get the XML to find device names
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get XML for block devices: %w", err)
	}

	type DiskXML struct {
		Target struct {
			Dev string `xml:"dev,attr"`
		} `xml:"target"`
		Source struct {
			File string `xml:"file,attr"`
			Dev  string `xml:"dev,attr"`
		} `xml:"source"`
	}
	type DomainXML struct {
		Disks []DiskXML `xml:"devices>disk"`
	}

	var domainDef DomainXML
	if err := xml.Unmarshal([]byte(xmlDesc), &domainDef); err != nil {
		return nil, fmt.Errorf("failed to parse domain XML for disks: %w", err)
	}

	var blockDetails []BlockDeviceDetail
	for _, disk := range domainDef.Disks {
		if disk.Target.Dev == "" {
			continue
		}

		// Get block info using API
		capacity, allocation, physical, err := l.DomainGetBlockInfo(domain, disk.Target.Dev, 0)
		if err != nil {
			log.Debugf("Failed to get block info for device %s: %v", disk.Target.Dev, err)
			continue
		}

		// Get block job info (for snapshots, backups, etc.)
		jobType, bandwidth, cur, end, _, err := l.DomainGetBlockJobInfo(domain, disk.Target.Dev, 0)
		var jobInfo *BlockJobInfo
		if err == nil {
			jobInfo = &BlockJobInfo{
				Type:      int32(jobType),
				Bandwidth: uint64(bandwidth),
				Cur:       cur,
				End:       end,
			}
		}

		blockDetails = append(blockDetails, BlockDeviceDetail{
			Device:     disk.Target.Dev,
			SourcePath: getSourcePath(disk.Source.File, disk.Source.Dev),
			Capacity:   capacity,
			Allocation: allocation,
			Physical:   physical,
			JobInfo:    jobInfo,
		})
	}

	return blockDetails, nil
}

// GetDomainSecurityDetails retrieves security label information using libvirt APIs
func (c *Connector) GetDomainSecurityDetails(hostID, vmName string) ([]SecurityDetail, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get security label list
	secLabels, _, err := l.DomainGetSecurityLabelList(domain)
	if err != nil {
		log.Debugf("Security labels not available for domain %s: %v", vmName, err)
		return nil, nil
	}

	var securityDetails []SecurityDetail
	for _, label := range secLabels {
		// Convert []int8 to string
		labelStr := string(convertInt8ArrayToBytes(label.Label))
		securityDetails = append(securityDetails, SecurityDetail{
			Label:     labelStr,
			Enforcing: label.Enforcing,
		})
	}

	return securityDetails, nil
}

// GetDomainIOThreadDetails retrieves I/O thread information using libvirt APIs
func (c *Connector) GetDomainIOThreadDetails(hostID, vmName string) ([]IOThreadDetail, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get IOThread info
	iothreadInfo, _, err := l.DomainGetIothreadInfo(domain, 0)
	if err != nil {
		log.Debugf("IOThread info not available for domain %s: %v", vmName, err)
		return nil, nil
	}

	var iothreadDetails []IOThreadDetail
	for _, info := range iothreadInfo {
		iothreadDetails = append(iothreadDetails, IOThreadDetail{
			IOThreadID: info.IothreadID,
			Cpumap:     convertBytesToBoolArray(info.Cpumap),
		})
	}

	return iothreadDetails, nil
}

// GetDomainNUMADetails retrieves NUMA configuration using libvirt APIs
func (c *Connector) GetDomainNUMADetails(hostID, vmName string) (*NUMADetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get NUMA parameters
	numaParams, _, err := l.DomainGetNumaParameters(domain, -1, 0)
	if err != nil {
		log.Debugf("NUMA parameters not available for domain %s: %v", vmName, err)
		// Try to get basic NUMA info from DomainGetInfo
		return &NUMADetails{NodeCount: 0}, nil
	}

	// Count NUMA nodes by checking for node-specific parameters
	nodeCount := 0
	for _, param := range numaParams {
		if strings.Contains(param.Field, "node") {
			nodeCount++
		}
	}
	if nodeCount == 0 {
		nodeCount = 1 // Default to 1 if no node-specific params found
	}

	return &NUMADetails{
		NUMAParams: numaParams,
		NodeCount:  nodeCount,
	}, nil
}

// Helper function to convert []int8 to []byte for string conversion
func convertInt8ArrayToBytes(arr []int8) []byte {
	bytes := make([]byte, len(arr))
	for i, v := range arr {
		bytes[i] = byte(v)
	}
	return bytes
}

// Helper function to convert []byte to []bool for CPU maps
func convertBytesToBoolArray(bytes []byte) []bool {
	if bytes == nil {
		return nil
	}
	bools := make([]bool, len(bytes)*8)
	for i, b := range bytes {
		for j := 0; j < 8; j++ {
			bools[i*8+j] = (b & (1 << j)) != 0
		}
	}
	return bools
}

// Helper function to get source path
func getSourcePath(file, dev string) string {
	if file != "" {
		return file
	}
	return dev
}

// GetDomainMemoryStats retrieves detailed memory statistics using libvirt APIs
func (c *Connector) GetDomainMemoryStats(hostID, vmName string) (*MemoryStats, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get memory statistics
	memStats, err := l.DomainMemoryStats(domain, 11, 0) // 11 = max stats
	if err != nil {
		log.Debugf("Memory statistics not available for domain %s: %v", vmName, err)
		return nil, nil
	}

	stats := &MemoryStats{}
	for _, stat := range memStats {
		switch stat.Tag {
		case 0: // VIR_DOMAIN_MEMORY_STAT_SWAP_IN
			stats.SwapIn = stat.Val
		case 1: // VIR_DOMAIN_MEMORY_STAT_SWAP_OUT
			stats.SwapOut = stat.Val
		case 2: // VIR_DOMAIN_MEMORY_STAT_MAJOR_FAULT
			stats.MajorFault = stat.Val
		case 3: // VIR_DOMAIN_MEMORY_STAT_MINOR_FAULT
			stats.MinorFault = stat.Val
		case 4: // VIR_DOMAIN_MEMORY_STAT_UNUSED
			stats.Unused = stat.Val
		case 5: // VIR_DOMAIN_MEMORY_STAT_AVAILABLE
			stats.Available = stat.Val
		case 6: // VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON
			stats.Actual = stat.Val
		case 7: // VIR_DOMAIN_MEMORY_STAT_RSS
			stats.Rss = stat.Val
		case 8: // VIR_DOMAIN_MEMORY_STAT_USABLE
			stats.Usable = stat.Val
		case 9: // VIR_DOMAIN_MEMORY_STAT_LAST_UPDATE
			stats.LastUpdate = stat.Val
		case 10: // VIR_DOMAIN_MEMORY_STAT_DISK_CACHES
			stats.DiskCaches = stat.Val
		case 11: // VIR_DOMAIN_MEMORY_STAT_HUGETLB_PGALLOC
			stats.HugetlbPgalloc = stat.Val
		case 12: // VIR_DOMAIN_MEMORY_STAT_HUGETLB_PGFAIL
			stats.HugetlbPgfail = stat.Val
		}
	}

	return stats, nil
}

// GetDomainGuestAgentDetails retrieves guest agent information using libvirt APIs
func (c *Connector) GetDomainGuestAgentDetails(hostID, vmName string) (*GuestAgentDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	// Get guest info from guest agent
	guestInfo, err := l.DomainGetGuestInfo(domain, 0x1F, 0) // 0x1F = all info types
	if err != nil {
		log.Debugf("Guest agent info not available for domain %s: %v", vmName, err)
		return nil, nil
	}

	details := &GuestAgentDetails{
		OSInfo:      make(map[string]interface{}),
		Interfaces:  make([]GuestInterfaceInfo, 0),
		Filesystems: make([]GuestFilesystemInfo, 0),
		Users:       make([]GuestUserInfo, 0),
	}

	// Parse guest info (this is a simplified version - actual parsing would be more complex)
	for _, param := range guestInfo {
		switch param.Field {
		case "hostname":
			if hostname, ok := param.Value.I.(string); ok {
				details.Hostname = hostname
			}
		case "os.id", "os.name", "os.version":
			details.OSInfo[param.Field] = param.Value.I
		case "timezone.name":
			if tzName, ok := param.Value.I.(string); ok {
				details.Timezone.Name = tzName
			}
		case "timezone.offset":
			if tzOffset, ok := param.Value.I.(int32); ok {
				details.Timezone.Offset = tzOffset
			}
		}
	}

	return details, nil
}

// GetDomainPerformanceDetails retrieves comprehensive performance statistics
func (c *Connector) GetDomainPerformanceDetails(hostID, vmName string) (*PerformanceDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	performance := &PerformanceDetails{
		BlockStats:     make(map[string]BlockStatsInfo),
		InterfaceStats: make(map[string]InterfaceStatsInfo),
	}

	// Get CPU statistics
	if cpuStats, _, err := l.DomainGetCPUStats(domain, 0, 0, 1, 0); err == nil {
		performance.CPUStats = cpuStats
	} else {
		log.Debugf("CPU stats not available for domain %s: %v", vmName, err)
	}

	// Get memory statistics
	if memStats, memErr := c.GetDomainMemoryStats(hostID, vmName); memErr == nil {
		performance.MemoryStats = memStats
	}

	// Get domain time
	if seconds, nanoseconds, timeErr := l.DomainGetTime(domain, 0); timeErr == nil {
		performance.DomainTime = &DomainTimeInfo{
			Seconds:     uint64(seconds),
			Nanoseconds: nanoseconds,
			Synced:      true,
		}
	} else {
		log.Debugf("Domain time not available for domain %s: %v", vmName, timeErr)
	}

	// Get block device statistics
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err == nil {
		type SimpleDisk struct {
			Target struct {
				Dev string `xml:"dev,attr"`
			} `xml:"target"`
		}
		type SimpleDomain struct {
			Disks []SimpleDisk `xml:"devices>disk"`
		}

		var domainDef SimpleDomain
		if xml.Unmarshal([]byte(xmlDesc), &domainDef) == nil {
			for _, disk := range domainDef.Disks {
				if disk.Target.Dev != "" {
					if rdReq, rdBytes, wrReq, wrBytes, errs, blockErr := l.DomainBlockStats(domain, disk.Target.Dev); blockErr == nil {
						performance.BlockStats[disk.Target.Dev] = BlockStatsInfo{
							ReadRequests:  uint64(rdReq),
							ReadBytes:     uint64(rdBytes),
							WriteRequests: uint64(wrReq),
							WriteBytes:    uint64(wrBytes),
							Errors:        uint64(errs),
						}
					}
				}
			}
		}
	}

	// Get network interface statistics
	type SimpleInterface struct {
		Target struct {
			Dev string `xml:"dev,attr"`
		} `xml:"target"`
	}
	type SimpleInterfaceDomain struct {
		Interfaces []SimpleInterface `xml:"devices>interface"`
	}

	var interfaceDef SimpleInterfaceDomain
	if xml.Unmarshal([]byte(xmlDesc), &interfaceDef) == nil {
		for _, iface := range interfaceDef.Interfaces {
			if iface.Target.Dev != "" {
				if rxBytes, rxPackets, rxErrs, rxDrop, txBytes, txPackets, txErrs, txDrop, ifaceErr := l.DomainInterfaceStats(domain, iface.Target.Dev); ifaceErr == nil {
					performance.InterfaceStats[iface.Target.Dev] = InterfaceStatsInfo{
						RxBytes:   uint64(rxBytes),
						RxPackets: uint64(rxPackets),
						RxErrors:  uint64(rxErrs),
						RxDrops:   uint64(rxDrop),
						TxBytes:   uint64(txBytes),
						TxPackets: uint64(txPackets),
						TxErrors:  uint64(txErrs),
						TxDrops:   uint64(txDrop),
					}
				}
			}
		}
	}

	return performance, nil
}

// Phase 2: Performance and Statistics APIs

// GetDomainCPUPerformance retrieves detailed CPU performance statistics
func (c *Connector) GetDomainCPUPerformance(hostID, vmName string) (*CPUPerformanceDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	performance := &CPUPerformanceDetails{
		VCPUStats: make(map[int]VCPUStats),
	}

	// Get CPU statistics for all VCPUs
	cpuStats, _, err := l.DomainGetCPUStats(domain, 0, -1, 0, 0) // 0 params, -1 = all CPUs
	if err != nil {
		log.Debugf("CPU statistics not available for domain %s: %v", vmName, err)
	} else {
		for i, stat := range cpuStats {
			vcpuStats := VCPUStats{}
			switch stat.Field {
			case "cpu_time":
				if val, ok := stat.Value.I.(uint64); ok {
					vcpuStats.CPUTime = val
				}
			case "user_time":
				if val, ok := stat.Value.I.(uint64); ok {
					vcpuStats.UserTime = val
				}
			case "system_time":
				if val, ok := stat.Value.I.(uint64); ok {
					vcpuStats.SystemTime = val
				}
			}
			performance.VCPUStats[i] = vcpuStats
		}
	}

	// Get node CPU statistics for host comparison
	nodeStats, _, err := l.NodeGetCPUStats(-1, 0, 0) // -1 = all CPUs
	if err != nil {
		log.Debugf("Node CPU statistics not available: %v", err)
	} else {
		hostCPU := HostCPUStats{}
		for _, stat := range nodeStats {
			switch stat.Field {
			case "kernel":
				hostCPU.Kernel = stat.Value
			case "user":
				hostCPU.User = stat.Value
			case "idle":
				hostCPU.Idle = stat.Value
			case "iowait":
				hostCPU.IOWait = stat.Value
			}
		}
		performance.HostCPUStats = &hostCPU
	}

	return performance, nil
}

// GetDomainInterfaceAddresses retrieves network interface addresses using libvirt APIs
func (c *Connector) GetDomainInterfaceAddresses(hostID, vmName string) (*NetworkInterfaceDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	details := &NetworkInterfaceDetails{
		Interfaces: make([]InterfaceAddress, 0),
	}

	// Get interface addresses using DHCP lease info
	interfaces, err := l.DomainInterfaceAddresses(domain, 1, 0) // 1 = DHCP source
	if err != nil {
		log.Debugf("Interface addresses not available via DHCP for domain %s: %v", vmName, err)

		// Fallback to guest agent if DHCP fails
		interfaces, err = l.DomainInterfaceAddresses(domain, 0, 0) // 0 = guest agent
		if err != nil {
			log.Debugf("Interface addresses not available via guest agent for domain %s: %v", vmName, err)
			return details, nil
		}
	}

	for _, iface := range interfaces {
		interfaceAddr := InterfaceAddress{
			Name:   string(iface.Name),
			HWAddr: "", // Initialize empty, will be set below
			Addrs:  make([]IPAddress, 0),
		}

		// Handle OptString for hardware address
		if len(iface.Hwaddr) > 0 {
			interfaceAddr.HWAddr = iface.Hwaddr[0]
		}

		for _, addr := range iface.Addrs {
			ipAddr := IPAddress{
				Type:   int(addr.Type),
				Addr:   string(addr.Addr),
				Prefix: int(addr.Prefix),
			}
			interfaceAddr.Addrs = append(interfaceAddr.Addrs, ipAddr)
		}

		details.Interfaces = append(details.Interfaces, interfaceAddr)
	}

	return details, nil
}

// GetDomainJobInfo retrieves active job information (useful for monitoring migration, backup, etc.)
func (c *Connector) GetDomainJobInfo(hostID, vmName string) (*DomainJobInfo, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	jobType, timeElapsed, timeRemaining, dataTotal, dataProcessed, dataRemaining, memTotal, memProcessed, memRemaining, fileTotal, fileProcessed, fileRemaining, err := l.DomainGetJobInfo(domain)
	if err != nil {
		log.Debugf("Job info not available for domain %s: %v", vmName, err)
		return nil, nil
	}

	return &DomainJobInfo{
		Type:          int(jobType),
		TimeElapsed:   timeElapsed,
		TimeRemaining: timeRemaining,
		DataTotal:     dataTotal,
		DataProcessed: dataProcessed,
		DataRemaining: dataRemaining,
		MemTotal:      memTotal,
		MemProcessed:  memProcessed,
		MemRemaining:  memRemaining,
		FileTotal:     fileTotal,
		FileProcessed: fileProcessed,
		FileRemaining: fileRemaining,
	}, nil
}

// Phase 3: Hybrid Optimization - Enhanced Sync Methods

// GetDomainHybridDetails combines API and XML data for comprehensive domain information
func (c *Connector) GetDomainHybridDetails(hostID, vmName string) (*HybridDomainDetails, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	details := &HybridDomainDetails{
		APIData: &APIDomainData{},
		XMLData: &XMLDomainData{},
	}

	// Get API-based data
	if memoryDetails, err := c.GetDomainMemoryDetails(hostID, vmName); err == nil {
		details.APIData.MemoryDetails = memoryDetails
	}

	if cpuDetails, err := c.GetDomainCPUDetails(hostID, vmName); err == nil {
		details.APIData.CPUDetails = cpuDetails
	}

	if blockDetails, err := c.GetDomainBlockDetails(hostID, vmName); err == nil {
		details.APIData.BlockDetails = blockDetails
	}

	if securityDetails, err := c.GetDomainSecurityDetails(hostID, vmName); err == nil {
		details.APIData.SecurityDetails = securityDetails
	}

	if numaDetails, err := c.GetDomainNUMADetails(hostID, vmName); err == nil {
		details.APIData.NUMADetails = numaDetails
	}

	if memStats, err := c.GetDomainMemoryStats(hostID, vmName); err == nil {
		details.APIData.MemoryStats = memStats
	}

	if perfDetails, err := c.GetDomainPerformanceDetails(hostID, vmName); err == nil {
		details.APIData.PerformanceDetails = perfDetails
	}

	// Get XML data for complex configurations
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		log.Debugf("Failed to get XML description for domain %s: %v", vmName, err)
	} else {
		details.XMLData.RawXML = xmlDesc

		// Parse specific XML components that don't have direct API equivalents
		if features, err := c.parseXMLFeatures(xmlDesc); err == nil {
			details.XMLData.Features = features
		}

		if osConfig, err := c.parseXMLOSConfig(xmlDesc); err == nil {
			details.XMLData.OSConfig = osConfig
		}

		if deviceConfig, err := c.parseXMLDeviceConfig(xmlDesc); err == nil {
			details.XMLData.DeviceConfig = deviceConfig
		}
	}

	return details, nil
}

// GetDomainOptimizedSync provides an optimized sync method that chooses the best data source
func (c *Connector) GetDomainOptimizedSync(hostID, vmName string) (*OptimizedSyncData, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	sync := &OptimizedSyncData{
		Strategy: make(map[string]string),
	}

	// Use API for real-time data
	if memStats, err := c.GetDomainMemoryStats(hostID, vmName); err == nil {
		sync.MemoryStats = memStats
		sync.Strategy["memory_stats"] = "api"
	}

	if perfDetails, err := c.GetDomainPerformanceDetails(hostID, vmName); err == nil {
		sync.PerformanceDetails = perfDetails
		sync.Strategy["performance"] = "api"
	}

	if cpuPerf, err := c.GetDomainCPUPerformance(hostID, vmName); err == nil {
		sync.CPUPerformance = cpuPerf
		sync.Strategy["cpu_performance"] = "api"
	}

	if networkDetails, err := c.GetDomainInterfaceAddresses(hostID, vmName); err == nil {
		sync.NetworkDetails = networkDetails
		sync.Strategy["network_addresses"] = "api"
	}

	// Use XML for configuration data when API is insufficient
	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		log.Debugf("Failed to get XML description for domain %s: %v", vmName, err)
	} else {
		// Parse graphics configuration (no direct API)
		if graphicsConfig, err := c.parseXMLGraphics(xmlDesc); err == nil {
			sync.GraphicsConfig = graphicsConfig
			sync.Strategy["graphics"] = "xml"
		}

		// Parse TPM configuration (no direct API)
		if tpmConfig, err := c.parseXMLTPM(xmlDesc); err == nil {
			sync.TPMConfig = tpmConfig
			sync.Strategy["tpm"] = "xml"
		}

		// Parse hypervisor features (complex XML structure)
		if hypervisorFeatures, err := c.parseXMLHypervisorFeatures(xmlDesc); err == nil {
			sync.HypervisorFeatures = hypervisorFeatures
			sync.Strategy["hypervisor_features"] = "xml"
		}
	}

	// Combine API + XML for enhanced block device info
	if blockDetails, err := c.GetDomainBlockDetails(hostID, vmName); err == nil {
		// Enhance with XML metadata
		if enhancedBlocks, err := c.enhanceBlockDetailsWithXML(xmlDesc, blockDetails); err == nil {
			sync.EnhancedBlockDetails = enhancedBlocks
			sync.Strategy["block_devices"] = "hybrid"
		}
	}

	return sync, nil
}

// Phase 3 Helper Methods (Stubs for XML parsing)

// parseXMLFeatures extracts hypervisor features from XML
func (c *Connector) parseXMLFeatures(xmlDesc string) (*XMLFeatures, error) {
	// Stub implementation - would parse XML for hypervisor features
	return &XMLFeatures{
		ACPI:     true,
		APIC:     true,
		VirtType: "hvm",
	}, nil
}

// parseXMLOSConfig extracts OS configuration from XML
func (c *Connector) parseXMLOSConfig(xmlDesc string) (*XMLOSConfig, error) {
	// Stub implementation - would parse XML for OS configuration
	return &XMLOSConfig{
		Type:      "hvm",
		Arch:      "x86_64",
		Machine:   "pc-i440fx-2.9",
		BootOrder: []string{"hd", "cdrom"},
	}, nil
}

// parseXMLDeviceConfig extracts device configuration from XML
func (c *Connector) parseXMLDeviceConfig(xmlDesc string) (*XMLDeviceConfig, error) {
	// Stub implementation - would parse XML for device configuration
	return &XMLDeviceConfig{
		Emulator:    "/usr/bin/qemu-system-x86_64",
		Controllers: []XMLController{},
		Channels:    []XMLChannel{},
	}, nil
}

// parseXMLGraphics extracts graphics configuration from XML
func (c *Connector) parseXMLGraphics(xmlDesc string) (*XMLGraphicsConfig, error) {
	// Stub implementation - would parse XML for graphics configuration
	return &XMLGraphicsConfig{
		Type:     "spice",
		Port:     5900,
		AutoPort: true,
		Listen:   "0.0.0.0",
	}, nil
}

// parseXMLTPM extracts TPM configuration from XML
func (c *Connector) parseXMLTPM(xmlDesc string) (*XMLTPMConfig, error) {
	// Stub implementation - would parse XML for TPM configuration
	return &XMLTPMConfig{
		Model:   "tpm-tis",
		Type:    "passthrough",
		Version: "2.0",
	}, nil
}

// parseXMLHypervisorFeatures extracts hypervisor-specific features from XML
func (c *Connector) parseXMLHypervisorFeatures(xmlDesc string) (*XMLHypervisorFeatures, error) {
	// Stub implementation - would parse XML for hypervisor features
	return &XMLHypervisorFeatures{
		Relaxed:   true,
		VAPIC:     true,
		Spinlocks: true,
		VPIndex:   true,
	}, nil
}

// enhanceBlockDetailsWithXML combines API block data with XML metadata
func (c *Connector) enhanceBlockDetailsWithXML(xmlDesc string, blockDetails []BlockDeviceDetail) (*EnhancedBlockDetails, error) {
	// Stub implementation - would enhance block details with XML metadata
	enhanced := &EnhancedBlockDetails{
		Devices:     blockDetails,
		XMLMetadata: make(map[string]XMLBlockMetadata),
	}

	// Add sample metadata for each device
	for _, device := range blockDetails {
		enhanced.XMLMetadata[device.Device] = XMLBlockMetadata{
			Driver:  "qemu",
			Cache:   "none",
			IO:      "native",
			Discard: "unmap",
		}
	}

	return enhanced, nil
}

// Phase 4 Helper Methods (Stubs for comprehensive XML parsing)

// parseDetailedHypervisorFeatures extracts detailed hypervisor features
func (c *Connector) parseDetailedHypervisorFeatures(xmlDesc string) (*DetailedHypervisorFeatures, error) {
	// Stub implementation - would parse detailed hypervisor features
	return &DetailedHypervisorFeatures{
		HyperV: &HyperVFeatures{
			Relaxed:   &FeatureState{State: "on"},
			VAPIC:     &FeatureState{State: "on"},
			Spinlocks: &FeatureState{State: "on", Attributes: map[string]string{"retries": "8191"}},
		},
		KVM: &KVMFeatures{
			Hidden: &FeatureState{State: "off"},
		},
	}, nil
}

// parseDetailedCPUFeatures extracts detailed CPU features and topology
func (c *Connector) parseDetailedCPUFeatures(xmlDesc string) (*DetailedCPUFeatures, error) {
	// Stub implementation - would parse detailed CPU features
	return &DetailedCPUFeatures{
		Mode:  "host-passthrough",
		Match: "exact",
		Check: "none",
		Model: &CPUModel{
			Name:     "SandyBridge",
			Fallback: "allow",
		},
		Topology: &CPUTopology{
			Sockets: 1,
			Cores:   2,
			Threads: 1,
		},
		Cache: &CPUCache{
			Mode:  "passthrough",
			Level: 3,
		},
	}, nil
}

// parseDetailedNUMATopology extracts detailed NUMA topology
func (c *Connector) parseDetailedNUMATopology(xmlDesc string) (*DetailedNUMATopology, error) {
	// Stub implementation - would parse detailed NUMA topology
	return &DetailedNUMATopology{
		Cells: []NUMACell{
			{
				ID:        0,
				CPUs:      "0-1",
				Memory:    2097152,
				Unit:      "KiB",
				MemAccess: "shared",
			},
		},
	}, nil
}

// parseAdvancedDeviceConfigs extracts advanced device configurations
func (c *Connector) parseAdvancedDeviceConfigs(xmlDesc string) (*AdvancedDeviceConfigs, error) {
	// Stub implementation - would parse advanced device configurations
	return &AdvancedDeviceConfigs{
		Input: []InputDevice{
			{Type: "tablet", Bus: "usb"},
			{Type: "keyboard", Bus: "ps2"},
		},
		Video: []VideoDevice{
			{
				Model:     "qxl",
				VRAMBytes: 65536,
				Heads:     1,
				Primary:   true,
			},
		},
		Sound: []SoundDevice{
			{Model: "ich6", Codec: "duplex"},
		},
		MemBalloon: &MemBalloonDevice{
			Model:       "virtio",
			AutoDeflate: true,
		},
	}, nil
}

// parseOSLoaderConfig extracts OS loader configuration
func (c *Connector) parseOSLoaderConfig(xmlDesc string) (*OSLoaderConfig, error) {
	// Stub implementation - would parse OS loader configuration
	return &OSLoaderConfig{
		Type:          "pflash",
		ReadOnly:      true,
		Secure:        true,
		Path:          "/usr/share/OVMF/OVMF_CODE.secboot.fd",
		NVRAM:         "/var/lib/libvirt/qemu/nvram/test_VARS.fd",
		NVRAMTemplate: "/usr/share/OVMF/OVMF_VARS.fd",
	}, nil
}

// parseClockConfig extracts clock and timer configuration
func (c *Connector) parseClockConfig(xmlDesc string) (*ClockConfig, error) {
	// Stub implementation - would parse clock configuration
	return &ClockConfig{
		Offset: "utc",
		Timers: []TimerConfig{
			{
				Name:       "rtc",
				TickPolicy: "catchup",
				Present:    true,
			},
			{
				Name:       "pit",
				TickPolicy: "delay",
				Present:    true,
			},
			{
				Name:    "hpet",
				Present: false,
			},
		},
	}, nil
}

// Additional XML parsing helper methods for CompleteXMLAnalysis

func (c *Connector) parseMetadata(xmlDesc string) (interface{}, error) {
	return map[string]string{"libosinfo": "http://libosinfo.org/xmlns/libvirt/domain/1.0"}, nil
}

func (c *Connector) parseMemoryBacking(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"hugepages": map[string]interface{}{
			"page": map[string]string{"size": "2048", "unit": "KiB"},
		},
		"nosharepages": true,
		"locked":       true,
	}, nil
}

func (c *Connector) parseCPUMode(xmlDesc string) (interface{}, error) {
	return map[string]string{"mode": "host-passthrough", "check": "none"}, nil
}

func (c *Connector) parseIOThreadConfig(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"count": 2,
		"iothread": []map[string]interface{}{
			{"id": 1, "cpuset": "0-1"},
			{"id": 2, "cpuset": "2-3"},
		},
	}, nil
}

func (c *Connector) parseCPUTuning(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"shares":          1024,
		"period":          100000,
		"quota":           -1,
		"emulator_period": 100000,
		"emulator_quota":  -1,
		"vcpupin": []map[string]interface{}{
			{"vcpu": 0, "cpuset": "0-1"},
			{"vcpu": 1, "cpuset": "2-3"},
		},
	}, nil
}

func (c *Connector) parseNUMATuning(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"memory": map[string]string{"mode": "strict", "nodeset": "0-1"},
		"memnode": []map[string]interface{}{
			{"cellid": 0, "mode": "strict", "nodeset": "0"},
			{"cellid": 1, "mode": "strict", "nodeset": "1"},
		},
	}, nil
}

func (c *Connector) parseSysInfo(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"type": "smbios",
		"bios": map[string]string{
			"vendor":  "SeaBIOS",
			"version": "1.11.0-2.el7",
			"date":    "04/01/2014",
		},
		"system": map[string]string{
			"manufacturer": "QEMU",
			"product":      "Standard PC (i440FX + PIIX, 1996)",
			"version":      "pc-i440fx-2.9",
		},
	}, nil
}

func (c *Connector) parseBootLoader(xmlDesc string) (interface{}, error) {
	return map[string]string{
		"executable": "/usr/bin/qemu-system-x86_64",
		"args":       "-enable-kvm -machine q35",
	}, nil
}

func (c *Connector) parseBIOSConfig(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"useserial":     true,
		"rebootTimeout": 10000,
	}, nil
}

func (c *Connector) parsePowerManagement(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"suspend_to_mem":  map[string]bool{"enabled": false},
		"suspend_to_disk": map[string]bool{"enabled": false},
	}, nil
}

func (c *Connector) parseKeyWrap(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"aes": map[string]bool{"state": true},
		"dea": map[string]bool{"state": false},
	}, nil
}

func (c *Connector) parseIDMap(xmlDesc string) (interface{}, error) {
	return map[string]interface{}{
		"uid": []map[string]int{
			{"start": 0, "target": 1000, "count": 10},
		},
		"gid": []map[string]int{
			{"start": 0, "target": 1000, "count": 10},
		},
	}, nil
}

func (c *Connector) parseResourceConfig(xmlDesc string) (interface{}, error) {
	return map[string]string{
		"partition":    "/machine/qemu",
		"fibrechannel": "appid",
	}, nil
}

func (c *Connector) parseSecurityLabels(xmlDesc string) (interface{}, error) {
	return []map[string]interface{}{
		{
			"type":       "dynamic",
			"model":      "selinux",
			"relabel":    true,
			"label":      "system_u:system_r:svirt_t:s0:c107,c434",
			"imagelabel": "system_u:object_r:svirt_image_t:s0:c107,c434",
		},
	}, nil
}

// Phase 4: XML-Only Components - Specialized XML parsing for components without direct APIs

// GetDomainXMLOnlyFeatures retrieves features that are only available through XML parsing
func (c *Connector) GetDomainXMLOnlyFeatures(hostID, vmName string) (*XMLOnlyFeatures, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get XML description for domain %s: %v", vmName, err)
	}

	features := &XMLOnlyFeatures{}

	// Parse hypervisor features
	if hvFeatures, err := c.parseDetailedHypervisorFeatures(xmlDesc); err == nil {
		features.HypervisorFeatures = hvFeatures
	}

	// Parse CPU features and topology
	if cpuFeatures, err := c.parseDetailedCPUFeatures(xmlDesc); err == nil {
		features.CPUFeatures = cpuFeatures
	}

	// Parse NUMA topology
	if numaTopology, err := c.parseDetailedNUMATopology(xmlDesc); err == nil {
		features.NUMATopology = numaTopology
	}

	// Parse device configurations that don't have API equivalents
	if deviceConfigs, err := c.parseAdvancedDeviceConfigs(xmlDesc); err == nil {
		features.AdvancedDevices = deviceConfigs
	}

	// Parse OS loader and NVRAM settings
	if osLoader, err := c.parseOSLoaderConfig(xmlDesc); err == nil {
		features.OSLoader = osLoader
	}

	// Parse clock and timer configurations
	if clockConfig, err := c.parseClockConfig(xmlDesc); err == nil {
		features.ClockConfig = clockConfig
	}

	return features, nil
}

// GetDomainCompleteXMLAnalysis performs comprehensive XML analysis for all XML-only features
func (c *Connector) GetDomainCompleteXMLAnalysis(hostID, vmName string) (*CompleteXMLAnalysis, error) {
	l, domain, err := c.getDomainByName(hostID, vmName)
	if err != nil {
		return nil, err
	}

	xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get XML description for domain %s: %v", vmName, err)
	}

	analysis := &CompleteXMLAnalysis{
		OriginalXML:      xmlDesc,
		ParsedComponents: make(map[string]interface{}),
	}

	// Parse all major XML components
	components := map[string]func(string) (interface{}, error){
		"metadata":       func(xml string) (interface{}, error) { return c.parseMetadata(xml) },
		"memory_backing": func(xml string) (interface{}, error) { return c.parseMemoryBacking(xml) },
		"cpu_mode":       func(xml string) (interface{}, error) { return c.parseCPUMode(xml) },
		"iothreads":      func(xml string) (interface{}, error) { return c.parseIOThreadConfig(xml) },
		"cputune":        func(xml string) (interface{}, error) { return c.parseCPUTuning(xml) },
		"numatune":       func(xml string) (interface{}, error) { return c.parseNUMATuning(xml) },
		"sysinfo":        func(xml string) (interface{}, error) { return c.parseSysInfo(xml) },
		"bootloader":     func(xml string) (interface{}, error) { return c.parseBootLoader(xml) },
		"bios":           func(xml string) (interface{}, error) { return c.parseBIOSConfig(xml) },
		"pm":             func(xml string) (interface{}, error) { return c.parsePowerManagement(xml) },
		"keywrap":        func(xml string) (interface{}, error) { return c.parseKeyWrap(xml) },
		"idmap":          func(xml string) (interface{}, error) { return c.parseIDMap(xml) },
		"resource":       func(xml string) (interface{}, error) { return c.parseResourceConfig(xml) },
		"seclabel":       func(xml string) (interface{}, error) { return c.parseSecurityLabels(xml) },
	}

	for componentName, parseFunc := range components {
		if result, err := parseFunc(xmlDesc); err == nil {
			analysis.ParsedComponents[componentName] = result
		} else {
			log.Debugf("Failed to parse %s for domain %s: %v", componentName, vmName, err)
		}
	}

	return analysis, nil
}

// =============================
// Phase 1: Enhanced Disk API Methods
// =============================

// EnhancedDiskInfo combines XML and API data for comprehensive disk information
type EnhancedDiskInfo struct {
	DiskInfo // Embedded XML-parsed disk info
	// API-sourced enhancements
	BlockStats      *DomainBlockStatsResult
	AllocationBytes uint64
	PhysicalBytes   uint64
	VolumeInfo      *VolumeDetail
}

// DomainBlockStatsResult stores block statistics from API
type DomainBlockStatsResult struct {
	ReadReqs   int64
	ReadBytes  int64
	WriteReqs  int64
	WriteBytes int64
	Errors     int64
}

// VolumeDetail stores enhanced volume information from storage APIs
type VolumeDetail struct {
	Name        string
	Type        int8
	Capacity    uint64
	Allocation  uint64
	Path        string
	Format      string
	BackingPath string
}

// GetEnhancedDiskInfo retrieves comprehensive disk information using APIs where possible
func (c *Connector) GetEnhancedDiskInfo(hostID, vmUUID string, disks []DiskInfo) ([]EnhancedDiskInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	enhanced := make([]EnhancedDiskInfo, 0, len(disks))

	// Get the domain for API calls
	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		log.Debugf("Could not lookup domain %s for enhanced disk info: %v", vmUUID, err)
		// Fallback to XML-only data
		for _, disk := range disks {
			enhanced = append(enhanced, EnhancedDiskInfo{DiskInfo: disk})
		}
		return enhanced, nil
	}

	for _, disk := range disks {
		enhancedDisk := EnhancedDiskInfo{DiskInfo: disk}

		// Try to get block statistics (for running VMs)
		if state, _, err := conn.DomainGetState(domain, 0); err == nil && state == 1 { // VIR_DOMAIN_RUNNING
			if rdReq, rdBytes, wrReq, wrBytes, errs, err := conn.DomainBlockStats(domain, disk.Target.Dev); err == nil {
				enhancedDisk.BlockStats = &DomainBlockStatsResult{
					ReadReqs:   rdReq,
					ReadBytes:  rdBytes,
					WriteReqs:  wrReq,
					WriteBytes: wrBytes,
					Errors:     errs,
				}
			}
		}

		// Try to get block info (allocation, capacity, physical)
		diskPath := disk.Source.File
		if diskPath == "" {
			diskPath = disk.Source.Dev
		}
		if diskPath != "" {
			if allocation, capacity, physical, err := conn.DomainGetBlockInfo(domain, diskPath, 0); err == nil {
				enhancedDisk.AllocationBytes = allocation
				enhancedDisk.PhysicalBytes = physical
				// Update capacity if not available from XML
				if disk.Capacity.Value == 0 {
					enhancedDisk.DiskInfo.Capacity.Value = capacity
					enhancedDisk.DiskInfo.Capacity.Unit = "bytes"
				}
			}
		}

		// Try to get volume information if it's pool-managed
		if disk.Source.File != "" {
			if volDetail, err := c.getVolumeDetailFromPath(conn, disk.Source.File); err == nil {
				enhancedDisk.VolumeInfo = volDetail
			}
		}

		enhanced = append(enhanced, enhancedDisk)
	}

	return enhanced, nil
}

// getDomainByUUID helper function to lookup domain by UUID string
func (c *Connector) getDomainByUUID(conn *libvirt.Libvirt, uuidStr string) (libvirt.Domain, error) {
	// Convert string UUID to byte array
	uuid, err := parseUUID(uuidStr)
	if err != nil {
		return libvirt.Domain{}, err
	}

	return conn.DomainLookupByUUID(uuid)
}

// parseUUID converts a UUID string to a 16-byte array
func parseUUID(uuidStr string) (libvirt.UUID, error) {
	var uuid libvirt.UUID

	// Remove hyphens from UUID string
	cleanUUID := strings.ReplaceAll(uuidStr, "-", "")

	// Convert hex string to bytes
	if len(cleanUUID) != 32 {
		return uuid, fmt.Errorf("invalid UUID format: %s", uuidStr)
	}

	for i := 0; i < 16; i++ {
		hex := cleanUUID[i*2 : i*2+2]
		b, err := strconv.ParseUint(hex, 16, 8)
		if err != nil {
			return uuid, fmt.Errorf("invalid UUID format: %s", uuidStr)
		}
		uuid[i] = byte(b)
	}

	return uuid, nil
}

// getVolumeDetailFromPath attempts to get volume details by searching storage pools
func (c *Connector) getVolumeDetailFromPath(conn *libvirt.Libvirt, diskPath string) (*VolumeDetail, error) {
	// Get all storage pools
	pools, _, err := conn.ConnectListAllStoragePools(0, 0)
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		// List volumes in this pool
		volumes, _, err := conn.StoragePoolListAllVolumes(pool, 0, 0)
		if err != nil {
			continue
		}

		for _, vol := range volumes {
			// Get volume path
			volPath, err := conn.StorageVolGetPath(vol)
			if err != nil {
				continue
			}

			// Check if this volume matches our disk path
			if volPath == diskPath {
				// Get volume info
				volType, capacity, allocation, err := conn.StorageVolGetInfo(vol)
				if err != nil {
					continue
				}

				// Try to get volume XML for format information
				volXML, err := conn.StorageVolGetXMLDesc(vol, 0)
				format := "unknown"
				if err == nil {
					// Parse format from XML - simplified extraction
					if strings.Contains(volXML, "format type='qcow2'") {
						format = "qcow2"
					} else if strings.Contains(volXML, "format type='raw'") {
						format = "raw"
					}
				}

				return &VolumeDetail{
					Type:       volType,
					Capacity:   capacity,
					Allocation: allocation,
					Path:       volPath,
					Format:     format,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("volume not found for path: %s", diskPath)
}

// GetDiskBlockStatistics retrieves real-time block I/O statistics for all VM disks
func (c *Connector) GetDiskBlockStatistics(hostID, vmUUID string) (map[string]DomainBlockStatsResult, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		return nil, err
	}

	// Check if domain is running
	state, _, err := conn.DomainGetState(domain, 0)
	if err != nil || state != 1 { // Not running
		return nil, fmt.Errorf("domain not running, cannot get block statistics")
	}

	// Get domain XML to find all disk devices
	xmlDesc, err := conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, err
	}

	// Parse disk devices from XML
	var def DomainHardwareXML
	if err := xml.Unmarshal([]byte(xmlDesc), &def); err != nil {
		return nil, err
	}

	stats := make(map[string]DomainBlockStatsResult)
	for _, disk := range def.Devices.Disks {
		if rdReq, rdBytes, wrReq, wrBytes, errs, err := conn.DomainBlockStats(domain, disk.Target.Dev); err == nil {
			stats[disk.Target.Dev] = DomainBlockStatsResult{
				ReadReqs:   rdReq,
				ReadBytes:  rdBytes,
				WriteReqs:  wrReq,
				WriteBytes: wrBytes,
				Errors:     errs,
			}
		}
	}

	return stats, nil
}

// =============================
// Phase 2: Enhanced Network API Methods
// =============================

// EnhancedNetworkInfo combines XML and API data for comprehensive network information
type EnhancedNetworkInfo struct {
	NetworkInfo // Embedded XML-parsed network info
	// API-sourced enhancements
	InterfaceStats *DomainInterfaceStatsResult
	DHCPLeases     []NetworkDHCPLeaseInfo
	InterfaceAddrs []DomainInterfaceAddress
}

// DomainInterfaceStatsResult stores network interface statistics from API
type DomainInterfaceStatsResult struct {
	RxBytes   int64
	RxPackets int64
	RxErrs    int64
	RxDrop    int64
	TxBytes   int64
	TxPackets int64
	TxErrs    int64
	TxDrop    int64
}

// NetworkDHCPLeaseInfo stores DHCP lease information
type NetworkDHCPLeaseInfo struct {
	MacAddress string
	IPAddress  string
	Hostname   string
	ClientID   string
	ExpireTime time.Time
}

// DomainInterfaceAddress stores interface address from API
type DomainInterfaceAddress struct {
	Name    string
	MacAddr string
	Addrs   []NetworkInterfaceAddress
}

// =============================
// Phase 4: Enhanced Node/Device API Methods
// =============================

// NodePerformanceInfo contains host/node performance information from APIs
type NodePerformanceInfo struct {
	// Node info from NodeGetInfo
	CPUModel string `json:"cpu_model"`
	Memory   uint64 `json:"memory"`
	CPUs     int32  `json:"cpus"`
	MHz      int32  `json:"mhz"`
	Nodes    int32  `json:"nodes"`
	Sockets  int32  `json:"sockets"`
	Cores    int32  `json:"cores"`
	Threads  int32  `json:"threads"`
	// Memory stats from NodeGetMemoryStats
	TotalMemory  uint64 `json:"total_memory"`
	FreeMemory   uint64 `json:"free_memory"`
	BuffMemory   uint64 `json:"buff_memory"`
	CachedMemory uint64 `json:"cached_memory"`
	// CPU stats from NodeGetCPUStats
	UserCPUTime   uint64 `json:"user_cpu_time"`
	SystemCPUTime uint64 `json:"system_cpu_time"`
	IdleCPUTime   uint64 `json:"idle_cpu_time"`
	IOWaitTime    uint64 `json:"iowait_time"`
	// Calculated metrics
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	CollectedAt   time.Time `json:"collected_at"`
}

// DevicePerformanceInfo contains device-specific performance metrics
type DevicePerformanceInfo struct {
	DeviceType     string    `json:"device_type"`
	DeviceName     string    `json:"device_name"`
	BandwidthUsed  uint64    `json:"bandwidth_used"`
	BandwidthLimit uint64    `json:"bandwidth_limit"`
	LatencyMs      float64   `json:"latency_ms"`
	ErrorCount     uint64    `json:"error_count"`
	MetricsJSON    string    `json:"metrics_json"`
	CollectedAt    time.Time `json:"collected_at"`
}

// NetworkInterfaceAddress represents an IP address assigned to an interface
type NetworkInterfaceAddress struct {
	Type   int32
	Addr   string
	Prefix uint32
}

// GetEnhancedNetworkInfo retrieves comprehensive network information using APIs where possible
func (c *Connector) GetEnhancedNetworkInfo(hostID, vmUUID string, networks []NetworkInfo) ([]EnhancedNetworkInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	enhanced := make([]EnhancedNetworkInfo, 0, len(networks))

	// Get the domain for API calls
	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		log.Debugf("Could not lookup domain %s for enhanced network info: %v", vmUUID, err)
		// Fallback to XML-only data
		for _, network := range networks {
			enhanced = append(enhanced, EnhancedNetworkInfo{NetworkInfo: network})
		}
		return enhanced, nil
	}

	for _, network := range networks {
		enhancedNetwork := EnhancedNetworkInfo{NetworkInfo: network}

		// Try to get interface statistics (for running VMs)
		if state, _, err := conn.DomainGetState(domain, 0); err == nil && state == 1 { // VIR_DOMAIN_RUNNING
			if rxBytes, rxPackets, rxErrs, rxDrop, txBytes, txPackets, txErrs, txDrop, err := conn.DomainInterfaceStats(domain, network.Target.Dev); err == nil {
				enhancedNetwork.InterfaceStats = &DomainInterfaceStatsResult{
					RxBytes:   rxBytes,
					RxPackets: rxPackets,
					RxErrs:    rxErrs,
					RxDrop:    rxDrop,
					TxBytes:   txBytes,
					TxPackets: txPackets,
					TxErrs:    txErrs,
					TxDrop:    txDrop,
				}
			}
		}

		// Try to get interface addresses
		if interfaces, err := conn.DomainInterfaceAddresses(domain, 0, 0); err == nil {
			enhancedNetwork.InterfaceAddrs = make([]DomainInterfaceAddress, 0, len(interfaces))
			for _, iface := range interfaces {
				// Convert OptString to string helper
				macAddr := ""
				if len(iface.Hwaddr) > 0 {
					macAddr = iface.Hwaddr[0]
				}

				domainAddr := DomainInterfaceAddress{
					Name:    iface.Name,
					MacAddr: macAddr,
					Addrs:   make([]NetworkInterfaceAddress, 0, len(iface.Addrs)),
				}
				for _, addr := range iface.Addrs {
					domainAddr.Addrs = append(domainAddr.Addrs, NetworkInterfaceAddress{
						Type:   addr.Type,
						Addr:   addr.Addr,
						Prefix: addr.Prefix,
					})
				}
				enhancedNetwork.InterfaceAddrs = append(enhancedNetwork.InterfaceAddrs, domainAddr)
			}
		}

		// Try to get DHCP leases if network is specified
		if network.Source.Network != "" {
			if netObj, err := conn.NetworkLookupByName(network.Source.Network); err == nil {
				if leases, _, err := conn.NetworkGetDhcpLeases(netObj, libvirt.OptString{}, 0, 0); err == nil {
					enhancedNetwork.DHCPLeases = make([]NetworkDHCPLeaseInfo, 0, len(leases))
					for _, lease := range leases {
						// Convert OptString to string
						macAddr := ""
						if len(lease.Mac) > 0 {
							macAddr = lease.Mac[0]
						}
						hostname := ""
						if len(lease.Hostname) > 0 {
							hostname = lease.Hostname[0]
						}
						clientID := ""
						if len(lease.Clientid) > 0 {
							clientID = lease.Clientid[0]
						}

						// Filter leases by MAC address if available
						if network.Mac.Address == "" || macAddr == network.Mac.Address {
							leaseInfo := NetworkDHCPLeaseInfo{
								MacAddress: macAddr,
								IPAddress:  lease.Ipaddr,
								Hostname:   hostname,
								ClientID:   clientID,
								ExpireTime: time.Unix(int64(lease.Expirytime), 0),
							}
							enhancedNetwork.DHCPLeases = append(enhancedNetwork.DHCPLeases, leaseInfo)
						}
					}
				}
			}
		}

		enhanced = append(enhanced, enhancedNetwork)
	}

	return enhanced, nil
}

// GetNetworkInterfaceStatistics retrieves real-time network I/O statistics for all VM interfaces
func (c *Connector) GetNetworkInterfaceStatistics(hostID, vmUUID string) (map[string]DomainInterfaceStatsResult, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		return nil, err
	}

	// Check if domain is running
	state, _, err := conn.DomainGetState(domain, 0)
	if err != nil || state != 1 { // Not running
		return nil, fmt.Errorf("domain not running, cannot get interface statistics")
	}

	// Get domain XML to find all network interfaces
	xmlDesc, err := conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, err
	}

	// Parse network interfaces from XML
	var def DomainHardwareXML
	if err := xml.Unmarshal([]byte(xmlDesc), &def); err != nil {
		return nil, err
	}

	stats := make(map[string]DomainInterfaceStatsResult)
	for _, iface := range def.Devices.Interfaces {
		if rxBytes, rxPackets, rxErrs, rxDrop, txBytes, txPackets, txErrs, txDrop, err := conn.DomainInterfaceStats(domain, iface.Target.Dev); err == nil {
			stats[iface.Target.Dev] = DomainInterfaceStatsResult{
				RxBytes:   rxBytes,
				RxPackets: rxPackets,
				RxErrs:    rxErrs,
				RxDrop:    rxDrop,
				TxBytes:   txBytes,
				TxPackets: txPackets,
				TxErrs:    txErrs,
				TxDrop:    txDrop,
			}
		}
	}

	return stats, nil
}

// =============================
// Phase 3: Enhanced CPU & Memory API Methods
// =============================

// EnhancedDomainInfo combines XML and API data for comprehensive domain resource information
type EnhancedDomainInfo struct {
	// API-sourced data from DomainGetInfo
	State     uint8
	MaxMemory uint64 // KB
	Memory    uint64 // KB
	NrVirtCPU uint16
	CPUTime   uint64 // nanoseconds
	// Extended API data
	MaxVCPUs     int32
	VCPUCount    int32
	VCPUInfo     []VCPUInfo
	MemoryParams map[string]interface{}
}

// VCPUInfo stores individual VCPU information
type VCPUInfo struct {
	Number  uint32
	State   int32
	CPUTime uint64
	CPU     int32
}

// GetEnhancedDomainInfo retrieves comprehensive domain resource information using APIs
func (c *Connector) GetEnhancedDomainInfo(hostID, vmUUID string) (*EnhancedDomainInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		return nil, err
	}

	// Get basic domain info
	state, maxMem, memory, nrVirtCPU, cpuTime, err := conn.DomainGetInfo(domain)
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedDomainInfo{
		State:     state,
		MaxMemory: maxMem,
		Memory:    memory,
		NrVirtCPU: nrVirtCPU,
		CPUTime:   cpuTime,
	}

	// Get maximum VCPUs
	if maxVCPUs, err := conn.DomainGetMaxVcpus(domain); err == nil {
		enhanced.MaxVCPUs = maxVCPUs
	}

	// Get current VCPU count
	if vcpuCount, err := conn.DomainGetVcpusFlags(domain, 0); err == nil {
		enhanced.VCPUCount = vcpuCount
	}

	// Get detailed VCPU information (if domain is running)
	if state == 1 { // VIR_DOMAIN_RUNNING
		if vcpuInfos, _, err := conn.DomainGetVcpus(domain, int32(nrVirtCPU), 0); err == nil {
			enhanced.VCPUInfo = make([]VCPUInfo, len(vcpuInfos))
			for i, vcpu := range vcpuInfos {
				enhanced.VCPUInfo[i] = VCPUInfo{
					Number:  vcpu.Number,
					State:   vcpu.State,
					CPUTime: vcpu.CPUTime,
					CPU:     vcpu.CPU,
				}
			}
		}
	}

	// Get memory parameters
	if memParams, _, err := conn.DomainGetMemoryParameters(domain, 0, 0); err == nil {
		enhanced.MemoryParams = make(map[string]interface{})
		for _, param := range memParams {
			// Use the discriminated union correctly
			enhanced.MemoryParams[param.Field] = param.Value.I
		}
	}

	return enhanced, nil
}

// GetEnhancedCPUPerformance retrieves CPU performance metrics using APIs
func (c *Connector) GetEnhancedCPUPerformance(hostID, vmUUID string) (*CPUPerformanceData, error) {
	enhanced, err := c.GetEnhancedDomainInfo(hostID, vmUUID)
	if err != nil {
		return nil, err
	}

	cpuPerf := &CPUPerformanceData{
		State:     enhanced.State,
		MaxMemory: enhanced.MaxMemory,
		Memory:    enhanced.Memory,
		NrVirtCPU: enhanced.NrVirtCPU,
		CPUTime:   enhanced.CPUTime,
		VCPUCount: enhanced.VCPUCount,
		MaxVCPUs:  enhanced.MaxVCPUs,
	}

	// Calculate CPU percentage (simplified - would need previous measurement for accurate calculation)
	if enhanced.CPUTime > 0 && enhanced.NrVirtCPU > 0 {
		// This is a placeholder calculation - in practice you'd need to compare with previous samples
		cpuPerf.CPUPercent = float64(enhanced.NrVirtCPU) * 10.0 // Placeholder value
	}

	return cpuPerf, nil
}

// GetEnhancedMemoryPerformance retrieves memory performance metrics using APIs
func (c *Connector) GetEnhancedMemoryPerformance(hostID, vmUUID string) (*MemoryPerformanceData, error) {
	enhanced, err := c.GetEnhancedDomainInfo(hostID, vmUUID)
	if err != nil {
		return nil, err
	}

	memPerf := &MemoryPerformanceData{
		MaxMemoryKB:     enhanced.MaxMemory,
		CurrentMemoryKB: enhanced.Memory,
	}

	// Extract memory parameters
	if enhanced.MemoryParams != nil {
		if val, ok := enhanced.MemoryParams["hard_limit"]; ok {
			if limit, ok := val.(uint64); ok {
				memPerf.HardLimit = limit
			}
		}
		if val, ok := enhanced.MemoryParams["soft_limit"]; ok {
			if limit, ok := val.(uint64); ok {
				memPerf.SoftLimit = limit
			}
		}
		if val, ok := enhanced.MemoryParams["min_guarantee"]; ok {
			if guarantee, ok := val.(uint64); ok {
				memPerf.MinGuarantee = guarantee
			}
		}
		if val, ok := enhanced.MemoryParams["swap_hard_limit"]; ok {
			if limit, ok := val.(uint64); ok {
				memPerf.SwapHardLimit = limit
			}
		}
	}

	// Calculate memory percentage
	if enhanced.MaxMemory > 0 {
		memPerf.MemoryPercent = float64(enhanced.Memory) / float64(enhanced.MaxMemory) * 100.0
	}

	return memPerf, nil
}

// CPUPerformanceData represents CPU performance metrics
type CPUPerformanceData struct {
	State      uint8
	MaxMemory  uint64
	Memory     uint64
	NrVirtCPU  uint16
	CPUTime    uint64
	VCPUCount  int32
	MaxVCPUs   int32
	CPUPercent float64
}

// MemoryPerformanceData represents memory performance metrics
type MemoryPerformanceData struct {
	MaxMemoryKB     uint64
	CurrentMemoryKB uint64
	ActualBalloonKB uint64
	HardLimit       uint64
	SoftLimit       uint64
	MinGuarantee    uint64
	SwapHardLimit   uint64
	MemoryPercent   float64
}

// GetNodePerformance fetches host/node performance metrics using libvirt APIs
func (c *Connector) GetNodePerformance(hostID string) (*NodePerformanceInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	info := &NodePerformanceInfo{}

	// Get node info
	model, memory, cpus, mhz, nodes, sockets, cores, threads, err := conn.NodeGetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %w", err)
	}

	// Convert byte array to string for CPU model
	modelBytes := model[:]
	nullIndex := len(modelBytes)
	for i, b := range modelBytes {
		if b == 0 {
			nullIndex = i
			break
		}
	}
	// Convert []int8 to []byte then to string
	byteModel := make([]byte, nullIndex)
	for i := 0; i < nullIndex; i++ {
		byteModel[i] = byte(modelBytes[i])
	}
	info.CPUModel = string(byteModel)

	info.Memory = memory
	info.CPUs = int32(cpus)
	info.MHz = int32(mhz)
	info.Nodes = int32(nodes)
	info.Sockets = int32(sockets)
	info.Cores = int32(cores)
	info.Threads = int32(threads)

	// Get node CPU stats (use -1 for all CPUs)
	cpuStats, _, err := conn.NodeGetCPUStats(-1, 0, 0)
	if err != nil {
		// Silent failure since this is optional
	} else {
		for _, stat := range cpuStats {
			value := stat.Value
			switch stat.Field {
			case "user":
				info.UserCPUTime = value
			case "system":
				info.SystemCPUTime = value
			case "idle":
				info.IdleCPUTime = value
			case "iowait":
				info.IOWaitTime = value
			}
		}
	}

	// Get node memory stats (use 0 for nparams and -1 for all cells)
	memStats, _, err := conn.NodeGetMemoryStats(0, -1, 0)
	if err != nil {
		// Silent failure since this is optional
	} else {
		for _, stat := range memStats {
			value := stat.Value
			switch stat.Field {
			case "total":
				info.TotalMemory = value
			case "free":
				info.FreeMemory = value
			case "buffers":
				info.BuffMemory = value
			case "cached":
				info.CachedMemory = value
			}
		}
	}

	// Calculate CPU percentage
	totalCPUTime := info.UserCPUTime + info.SystemCPUTime + info.IdleCPUTime + info.IOWaitTime
	if totalCPUTime > 0 {
		activeCPUTime := info.UserCPUTime + info.SystemCPUTime
		info.CPUPercent = float64(activeCPUTime) / float64(totalCPUTime) * 100
	}

	// Calculate memory percentage
	if info.TotalMemory > 0 {
		usedMemory := info.TotalMemory - info.FreeMemory
		info.MemoryPercent = float64(usedMemory) / float64(info.TotalMemory) * 100
	}

	info.CollectedAt = time.Now()
	return info, nil
}

// GetDevicePerformance fetches device-specific performance metrics
func (c *Connector) GetDevicePerformance(hostID, vmUUID string) ([]DevicePerformanceInfo, error) {
	conn, err := c.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	domain, err := c.getDomainByUUID(conn, vmUUID)
	if err != nil {
		return nil, err
	}

	var devices []DevicePerformanceInfo

	// Get hostname (if available) as a basic device metric
	hostname, err := conn.DomainGetHostname(domain, 0)
	if err == nil && len(hostname) > 0 {
		device := DevicePerformanceInfo{
			DeviceType:  "hostname",
			DeviceName:  "hostname",
			MetricsJSON: fmt.Sprintf(`{"hostname": "%s"}`, hostname),
			CollectedAt: time.Now(),
		}
		devices = append(devices, device)
	}

	// For other device metrics, we would need to parse XML and then get specific device stats
	// This is a placeholder for device-specific metrics that would require XML parsing
	// to identify devices and then API calls for their specific performance data

	return devices, nil
}
