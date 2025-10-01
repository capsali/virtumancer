package storage

import (
	"time"

	log "github.com/capsali/virtumancer/internal/logging"
	"github.com/google/uuid"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Base contains common fields for all models using UUID primary keys
type Base struct {
	ID        string         `gorm:"primaryKey;type:char(36);default:(uuid())" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// VMState defines the possible stable states of a VM.
type VMState string

const (
	StateInitialized VMState = "INITIALIZED"
	StateActive      VMState = "ACTIVE"
	StatePaused      VMState = "PAUSED"
	StateSuspended   VMState = "SUSPENDED"
	StateStopped     VMState = "STOPPED"
	StateError       VMState = "ERROR"
	StateUnknown     VMState = "UNKNOWN"
)

// VMTaskState defines the possible transient states of a VM during an operation.
type VMTaskState string

const (
	TaskStateBuilding    VMTaskState = "BUILDING"
	TaskStatePausing     VMTaskState = "PAUSING"
	TaskStateUnpausing   VMTaskState = "UNPAUSING"
	TaskStateSuspending  VMTaskState = "SUSPENDING"
	TaskStateResuming    VMTaskState = "RESUMING"
	TaskStateDeleting    VMTaskState = "DELETING"
	TaskStateStopping    VMTaskState = "STOPPING"
	TaskStateStarting    VMTaskState = "STARTING"
	TaskStateRebooting   VMTaskState = "REBOOTING"
	TaskStateRebuilding  VMTaskState = "REBUILDING"
	TaskStatePoweringOn  VMTaskState = "POWERING_ON"
	TaskStatePoweringOff VMTaskState = "POWERING_OFF"
	TaskStateScheduling  VMTaskState = "SCHEDULING"
)

// SyncStatus defines the sync state of a VM's configuration against libvirt.
type SyncStatus string

const (
	StatusUnknown SyncStatus = "UNKNOWN"
	StatusSynced  SyncStatus = "SYNCED"
	StatusDrifted SyncStatus = "DRIFTED"
)

// --- Core Entities ---

// Host represents a libvirt host connection configuration.
type Host struct {
	Base
	Name string `json:"name,omitempty"` // Optional friendly name for the host
	URI  string `json:"uri"`
	// State reflects the stable connection state of the host.
	State string `gorm:"size:32;default:'DISCONNECTED'" json:"state"`
	// TaskState reflects transient work being performed on the host (e.g., connecting)
	TaskState string `gorm:"size:32" json:"task_state"`
	// AutoReconnectDisabled indicates if automatic reconnection is disabled for this host
	// (e.g., because it was manually disconnected by the user)
	AutoReconnectDisabled bool `gorm:"default:false" json:"auto_reconnect_disabled"`
}

// HostState defines allowed host states to mirror VM state behavior.
type HostState string

const (
	HostStateConnected    HostState = "CONNECTED"
	HostStateDisconnected HostState = "DISCONNECTED"
	HostStateError        HostState = "ERROR"
)

// HostTaskState defines transient host task states.
type HostTaskState string

const (
	HostTaskStateConnecting    HostTaskState = "CONNECTING"
	HostTaskStateDisconnecting HostTaskState = "DISCONNECTING"
)

// VirtualMachine is Virtumancer's canonical definition of a VM's intended state.
type VirtualMachine struct {
	Base
	HostID     string `gorm:"uniqueIndex:idx_vm_host_name" json:"hostId"`
	Name       string `gorm:"uniqueIndex:idx_vm_host_name" json:"name"`
	DomainUUID string `gorm:"uniqueIndex" json:"domainUuid"` // The UUID as reported by libvirt
	// Source indicates whether this VM was created/managed by Virtumancer
	// ('managed') or imported from libvirt ('imported'). Discovered VMs are
	// not persisted until explicitly imported.
	Source          string      `gorm:"size:32;default:'managed'" json:"source"`
	Title           string      `json:"title"` // Short domain title
	Description     string      `json:"description"`
	State           VMState     `gorm:"default:'INITIALIZED'" json:"state"`        // Intended/target state
	LibvirtState    VMState     `gorm:"default:'INITIALIZED'" json:"libvirtState"` // Observed state from libvirt (UNKNOWN when disconnected)
	TaskState       VMTaskState `json:"taskState"`                                 // Transient state during operations
	VCPUCount       uint        `json:"vcpuCount"`
	CPUModel        string      `json:"cpuModel"`
	CPUTopologyJSON string      `json:"cpuTopologyJson"`
	MemoryBytes     uint64      `json:"memoryBytes"`   // Maximum memory (maxMemory)
	CurrentMemory   uint64      `json:"currentMemory"` // Current memory allocation
	OSType          string      `json:"osType"`
	IsTemplate      bool        `json:"isTemplate"`
	Metadata        string      `gorm:"type:text" json:"metadata"` // Custom XML metadata
	SyncStatus      SyncStatus  `gorm:"default:'UNKNOWN'" json:"syncStatus"`
	DriftDetails    string      `json:"driftDetails"` // JSON blob storing drift information
	NeedsRebuild    bool        `gorm:"default:false" json:"needsRebuild"`
}

// CreateVMRequest represents the data structure for creating a new VM
type CreateVMRequest struct {
	// Basic VM configuration
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	HostID      string `json:"hostId" binding:"required"`
	OSType      string `json:"os_type,omitempty"`
	VCPUCount   uint   `json:"vcpu_count" binding:"required,min=1"`
	MemoryBytes uint64 `json:"memory_bytes" binding:"required,min=1"`
	DiskSizeGB  uint   `json:"disk_size_gb,omitempty"`

	// Network configuration
	NetworkInterface string `json:"network_interface,omitempty"`

	// Boot configuration
	BootDevice string `json:"boot_device,omitempty"`

	// CPU configuration
	CPUModel string `json:"cpu_model,omitempty"`

	// System settings
	Source       string `json:"source,omitempty"`
	SyncStatus   string `json:"sync_status,omitempty"`
	LibvirtState string `json:"libvirtState,omitempty"`
	DomainUUID   string `json:"domain_uuid,omitempty"`
	Title        string `json:"title,omitempty"`
	State        string `json:"state,omitempty"`
}

// --- Storage Management ---

// StoragePool represents a libvirt storage pool (e.g., LVM, a directory).
type StoragePool struct {
	Base
	HostID          string `json:"host_id"`
	Name            string `json:"name"`
	UUID            string `gorm:"uniqueIndex" json:"uuid"`
	Type            string `json:"type"`
	Path            string `json:"path"`
	CapacityBytes   uint64 `json:"capacity_bytes"`
	AllocationBytes uint64 `json:"allocation_bytes"`
}

// Volume represents a single storage volume, like a virtual disk or an ISO.
type Volume struct {
	Base
	StoragePoolID   string `json:"storage_pool_id"`
	Name            string `json:"name"`
	Type            string `json:"type"` // 'DISK' or 'ISO'
	Format          string `json:"format"`
	CapacityBytes   uint64 `json:"capacity_bytes"`
	AllocationBytes uint64 `json:"allocation_bytes"`
}

// --- Network Management ---

// Network represents a virtual network or bridge on a host.
type Network struct {
	Base
	HostID     string `gorm:"uniqueIndex:idx_network_host_name" json:"host_id"`
	Name       string `gorm:"uniqueIndex:idx_network_host_name" json:"name"`
	UUID       string `json:"uuid"`
	BridgeName string `json:"bridge_name"`
	Mode       string `json:"mode"` // e.g., 'bridged', 'nat', 'isolated'
}

// Port represents a virtual Network Interface Card (vNIC) belonging to a VM.
type Port struct {
	Base
	// VMUUID was removed in favor of explicit PortAttachment records.
	MACAddress string `json:"mac_address"` // canonical MAC for the resource
	// DeviceName removed from Port; device name is attachment-scoped
	ModelName           string `json:"model_name"` // e.g., 'virtio', 'e1000'
	IPAddress           string `json:"ip_address"`
	HostID              string `gorm:"index" json:"host_id"`               // optional host scoping for unattached ports
	SourceType          string `gorm:"size:32" json:"source_type"`         // 'network'|'bridge'|'hostdev'|'vhostuser'|'null'|'vdpa'
	SourceRef           string `gorm:"type:text" json:"source_ref"`        // network name, hostdev address, or vhost socket path
	PortGroup           string `gorm:"type:text" json:"port_group"`        // portgroup name for network sources
	VirtualPortJSON     string `gorm:"type:text" json:"virtual_port_json"` // serialized <virtualport> subelements
	FilterRefJSON       string `gorm:"type:text" json:"filter_ref_json"`   // serialized <filterref> subelements
	VlanTagsJSON        string `gorm:"type:text" json:"vlan_tags_json"`    // serialized VLAN tags / metadata
	TrustGuestRxFilters bool   `json:"trust_guest_rx_filters"`
	PrimaryVlan         *int   `gorm:"default:NULL" json:"primary_vlan"` // nullable primary VLAN tag
	AddressJSON         string `gorm:"type:text" json:"address_json"`    // optional device address (pci/slot/function)
}

// PortBinding links a Port to a Network.
type PortBinding struct {
	Base
	PortID    string  `json:"port_id"`
	Port      Port    `json:"port"`
	NetworkID string  `json:"network_id"`
	Network   Network `json:"network"`
}

// PortAttachment links a Port to a VirtualMachine and stores per-VM attachment metadata.
// This is intentionally separate from Port/PortBinding so ports can exist unattached
// (for provisioning / pool usage) and then be attached to VMs later.
type PortAttachment struct {
	Base
	VMUUID      string `gorm:"index" json:"vm_uuid"`
	PortID      string `gorm:"index" json:"port_id"`
	Port        Port   `json:"port"`
	HostID      string `gorm:"index" json:"host_id"` // host that the attachment is bound to (if attached)
	DeviceName  string `json:"device_name"`          // per-VM device name (overrides Port.DeviceName)
	MACAddress  string `json:"mac_address"`          // per-attachment MAC override
	ModelName   string `json:"model_name"`           // per-attachment model, if different
	Ordinal     int    `json:"ordinal"`
	Metadata    string `gorm:"type:text" json:"metadata"`     // optional JSON for hotplug / per-attachment options
	AddressJSON string `gorm:"type:text" json:"address_json"` // optional PCI/USB address for this attachment
}

// FilterRef represents a network filterref applied to a specific port/resource.
type FilterRef struct {
	Base
	PortID         string
	Name           string
	ParametersJSON string `gorm:"type:text"`
}

// VirtualPort represents the <virtualport> subtree for advanced NICs (e.g., openvswitch).
type VirtualPort struct {
	Base
	PortID     string
	Type       string
	ConfigJSON string `gorm:"type:text"`
}

// DeviceAlias records libvirt <alias name='...'> entries to map alias -> device.
type DeviceAlias struct {
	Base
	VMUUID     string `gorm:"index"`
	DeviceType string `gorm:"size:64"`
	DeviceID   string
	AliasName  string `gorm:"size:128"`
}

// --- Virtual Hardware Management ---

// Controller represents a hardware controller within a VM (e.g., USB, SATA).
type Controller struct {
	Base
	Type      string // 'usb', 'sata', 'virtio-serial'
	ModelName string
	Index     uint
}

// ControllerAttachment links a Controller to a VirtualMachine.
type ControllerAttachment struct {
	Base
	VMUUID       string `gorm:"index"`
	ControllerID string
}

// InputDevice represents an input device like a mouse or keyboard.
type InputDevice struct {
	Base
	Type string // 'mouse', 'tablet', 'keyboard'
	Bus  string // 'usb', 'ps2', 'virtio'
}

// InputDeviceAttachment links an InputDevice to a VirtualMachine.
type InputDeviceAttachment struct {
	Base
	VMUUID        string `gorm:"index"`
	InputDeviceID string
}

// GraphicsDevice represented a virtual GPU and display protocol (legacy).
// It was removed in favor of per-VM `Console` instances.
// Console represents a per-VM console instance (VNC/SPICE) discovered from domain XML.
// This is a per-instance model used during migration from the older
// GraphicsDevice/GraphicsDeviceAttachment model to a first-class Console instance.
type Console struct {
	Base
	VMUUID        string `gorm:"index"`
	HostID        string `gorm:"index"`
	Type          string // 'vnc' or 'spice'
	ModelName     string
	ListenAddress string
	Port          uint
	TLSPort       uint
	Metadata      string `gorm:"type:text"` // optional JSON blob
}

// SoundCard represents a virtual sound device.
type SoundCard struct {
	Base
	ModelName string // 'ich6', 'ac97'
}

// SoundCardAttachment links a SoundCard to a VirtualMachine.
type SoundCardAttachment struct {
	Base
	VMUUID      string `gorm:"index"`
	SoundCardID string
}

// HostDevice represents a physical device on a host for passthrough.
type HostDevice struct {
	Base
	HostID      string
	Type        string // 'pci', 'usb'
	Address     string // Physical address on host
	Description string
}

// HostDeviceAttachment links a HostDevice to a VirtualMachine for passthrough.
type HostDeviceAttachment struct {
	Base
	VMUUID       string `gorm:"index"`
	HostDeviceID string
}

// TPM represents a Trusted Platform Module device.
type TPM struct {
	Base
	ModelName   string // 'tpm-crb', 'tpm-tis'
	BackendType string // 'passthrough', 'emulator'
	BackendPath string
}

// TPMAttachment links a TPM to a VirtualMachine.
type TPMAttachment struct {
	Base
	VMUUID string `gorm:"index"`
	TPMID  string
}

// Watchdog represents a virtual watchdog device.
type Watchdog struct {
	Base
	ModelName string // 'i6300esb'
	Action    string // 'reset', 'shutdown', 'poweroff'
}

// WatchdogAttachment links a Watchdog to a VirtualMachine.
type WatchdogAttachment struct {
	Base
	VMUUID     string `gorm:"index"`
	WatchdogID string
}

// SerialDevice represents a serial port configuration.
type SerialDevice struct {
	Base
	Type       string // 'pty', 'tcp', 'stdio'
	TargetPort uint
	ConfigJSON string
}

// SerialDeviceAttachment links a SerialDevice to a VirtualMachine.
type SerialDeviceAttachment struct {
	Base
	VMUUID         string `gorm:"index"`
	SerialDeviceID string
}

// ChannelDevice represents a communication channel (e.g., for guest agent).
type ChannelDevice struct {
	Base
	Type       string // 'unix', 'spicevmc'
	TargetName string // e.g., 'org.qemu.guest_agent.0'
	ConfigJSON string
}

// ChannelDeviceAttachment links a ChannelDevice to a VirtualMachine.
type ChannelDeviceAttachment struct {
	Base
	VMUUID          string `gorm:"index"`
	ChannelDeviceID string
}

// Filesystem represents a shared filesystem for a VM.
type Filesystem struct {
	Base
	DriverType string
	SourcePath string
	TargetPath string
}

// FilesystemAttachment links a Filesystem to a VM.
type FilesystemAttachment struct {
	Base
	VMUUID       string `gorm:"index"`
	FilesystemID string
}

// Smartcard represents a smartcard device for a VM.
type Smartcard struct {
	Base
	Type       string
	ConfigJSON string
}

// SmartcardAttachment links a Smartcard to a VM.
type SmartcardAttachment struct {
	Base
	VMUUID      string `gorm:"index"`
	SmartcardID string
}

// USBRedirector represents a USB redirection device.
type USBRedirector struct {
	Base
	Type       string
	FilterRule string
}

// USBRedirectorAttachment links a USBRedirector to a VM.
type USBRedirectorAttachment struct {
	Base
	VMUUID          string `gorm:"index"`
	USBRedirectorID string
}

// RngDevice represents a Random Number Generator device.
type RngDevice struct {
	Base
	ModelName   string
	BackendType string
}

// RngDeviceAttachment links an RngDevice to a VM.
type RngDeviceAttachment struct {
	Base
	VMUUID      string `gorm:"index"`
	RngDeviceID string
}

// PanicDevice represents a panic device for a VM.
type PanicDevice struct {
	Base
	ModelName string
}

// PanicDeviceAttachment links a PanicDevice to a VM.
type PanicDeviceAttachment struct {
	Base
	VMUUID        string `gorm:"index"`
	PanicDeviceID string
}

// Vsock represents a VirtIO socket device.
type Vsock struct {
	Base
	GuestCID uint
}

// VsockAttachment links a Vsock to a VM.
type VsockAttachment struct {
	Base
	VMUUID  string `gorm:"index"`
	VsockID string
}

// MemoryBalloon represents a memory balloon device.
type MemoryBalloon struct {
	Base
	ModelName  string
	ConfigJSON string
}

// MemoryBalloonAttachment links a MemoryBalloon to a VM.
type MemoryBalloonAttachment struct {
	Base
	VMUUID          string `gorm:"index"`
	MemoryBalloonID string
}

// ShmemDevice represents a shared memory device.
type ShmemDevice struct {
	Base
	Name    string
	SizeKiB uint
	Path    string
}

// ShmemDeviceAttachment links a ShmemDevice to a VM.
type ShmemDeviceAttachment struct {
	Base
	VMUUID        string `gorm:"index"`
	ShmemDeviceID string
}

// IOMMUDevice represents an IOMMU device.
type IOMMUDevice struct {
	Base
	ModelName string
}

// IOMMUDeviceAttachment links an IOMMUDevice to a VM.
type IOMMUDeviceAttachment struct {
	Base
	VMUUID        string `gorm:"index"`
	IOMMUDeviceID string
}

// Disk represents a block device/resource. It may reference a Volume when managed
// by a storage pool, or a raw path when unmanaged.
type Disk struct {
	Base
	Name          string  `json:"name"`
	VolumeID      *string `json:"volume_id"`
	Path          string  `json:"path"`
	Format        string  `json:"format"`
	CapacityBytes uint64  `json:"capacity_bytes"`
	Serial        string  `json:"serial"`
	DriverJSON    string  `gorm:"type:text" json:"driver_json"`  // driver options (cache/io/…) as JSON
	BackingJSON   string  `gorm:"type:text" json:"backing_json"` // backingStore / layered info
	// Enhanced API-sourced fields
	VolumeType      string `json:"volume_type"`             // From StorageVolGetInfo
	AllocationBytes uint64 `json:"allocation_bytes"`        // Actual space used
	PhysicalBytes   uint64 `json:"physical_bytes"`          // Physical storage footprint
	TargetPath      string `json:"target_path"`             // Target device path
	SourceFormat    string `json:"source_format"`           // Source format if different
	Encryption      string `json:"encryption"`              // Encryption info from API
	IOTune          string `gorm:"type:text" json:"iotune"` // I/O tuning parameters as JSON
}

// DiskAttachment links a Disk (or volume) to a VM and stores per-VM metadata.
type DiskAttachment struct {
	Base
	VMUUID      string `gorm:"index" json:"vm_uuid"`
	DiskID      string `json:"disk_id"`
	Disk        Disk   `json:"disk"`        // Preloaded disk resource
	DeviceName  string `json:"device_name"` // e.g., vda
	BusType     string `json:"bus_type"`    // virtio/sata/ide
	ReadOnly    bool   `json:"read_only"`
	Shareable   bool   `json:"shareable"`
	AddressJSON string `gorm:"type:text" json:"address_json"` // PCI address or target addressing
	Metadata    string `gorm:"type:text" json:"metadata"`
}

// BlockStatistics stores real-time disk performance metrics from libvirt APIs
type BlockStatistics struct {
	Base
	DiskAttachmentID string `gorm:"index" json:"disk_attachment_id"`
	VMUUID           string `gorm:"index" json:"vm_uuid"`
	DeviceName       string `json:"device_name"`
	// Block I/O statistics from DomainGetBlockStats
	ReadReqs   uint64 `json:"read_reqs"`   // Number of read requests
	ReadBytes  uint64 `json:"read_bytes"`  // Bytes read
	WriteReqs  uint64 `json:"write_reqs"`  // Number of write requests
	WriteBytes uint64 `json:"write_bytes"` // Bytes written
	Errors     uint64 `json:"errors"`      // Error count
	// Block info from DomainGetBlockInfo
	Allocation uint64 `json:"allocation"` // Actual disk space used
	Capacity   uint64 `json:"capacity"`   // Total disk capacity
	Physical   uint64 `json:"physical"`   // Physical size on storage
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// NetworkStatistics stores real-time network performance metrics from libvirt APIs
type NetworkStatistics struct {
	Base
	PortAttachmentID string `gorm:"index" json:"port_attachment_id"`
	VMUUID           string `gorm:"index" json:"vm_uuid"`
	DeviceName       string `json:"device_name"`
	// Network I/O statistics from DomainInterfaceStats
	RxBytes   uint64 `json:"rx_bytes"`   // Bytes received
	RxPackets uint64 `json:"rx_packets"` // Packets received
	RxErrs    uint64 `json:"rx_errs"`    // Receive errors
	RxDrop    uint64 `json:"rx_drop"`    // Receive drops
	TxBytes   uint64 `json:"tx_bytes"`   // Bytes transmitted
	TxPackets uint64 `json:"tx_packets"` // Packets transmitted
	TxErrs    uint64 `json:"tx_errs"`    // Transmit errors
	TxDrop    uint64 `json:"tx_drop"`    // Transmit drops
	// DHCP information from NetworkGetDhcpLeases
	IPAddress   string    `json:"ip_address"`   // DHCP assigned IP
	LeaseExpiry time.Time `json:"lease_expiry"` // DHCP lease expiration
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// CPUPerformance stores real-time CPU performance metrics from libvirt APIs
type CPUPerformance struct {
	Base
	VMUUID string `gorm:"index" json:"vm_uuid"`
	// CPU info from DomainGetInfo
	State     uint8  `json:"state"`       // Domain state
	MaxMemory uint64 `json:"max_memory"`  // Maximum memory in KB
	Memory    uint64 `json:"memory"`      // Current memory in KB
	NrVirtCPU uint16 `json:"nr_virt_cpu"` // Number of virtual CPUs
	CPUTime   uint64 `json:"cpu_time"`    // CPU time used in nanoseconds
	// VCPU details from DomainGetVcpus
	VCPUCount int32 `json:"vcpu_count"` // Actual VCPU count
	MaxVCPUs  int32 `json:"max_vcpus"`  // Maximum VCPUs
	// Performance metrics
	CPUPercent float64 `json:"cpu_percent"` // CPU utilization percentage
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// MemoryPerformance stores real-time memory performance metrics from libvirt APIs
type MemoryPerformance struct {
	Base
	VMUUID string `gorm:"index" json:"vm_uuid"`
	// Memory info from DomainGetInfo and DomainGetMemoryParameters
	MaxMemoryKB     uint64 `json:"max_memory_kb"`     // Maximum memory in KB
	CurrentMemoryKB uint64 `json:"current_memory_kb"` // Current memory allocation in KB
	ActualBalloonKB uint64 `json:"actual_balloon_kb"` // Actual memory from balloon
	// Memory parameters from DomainGetMemoryParameters
	HardLimit     uint64 `json:"hard_limit"`      // Hard memory limit
	SoftLimit     uint64 `json:"soft_limit"`      // Soft memory limit
	MinGuarantee  uint64 `json:"min_guarantee"`   // Minimum guaranteed memory
	SwapHardLimit uint64 `json:"swap_hard_limit"` // Swap hard limit
	// Calculated metrics
	MemoryPercent float64 `json:"memory_percent"` // Memory utilization percentage
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// NodePerformance stores real-time node/host performance metrics from libvirt APIs
type NodePerformance struct {
	Base
	HostID string `gorm:"index" json:"host_id"`
	// Node info from NodeGetInfo
	CPUModel string `json:"cpu_model"` // CPU model
	Memory   uint64 `json:"memory"`    // Total memory in KB
	CPUs     int32  `json:"cpus"`      // Number of CPUs
	MHz      int32  `json:"mhz"`       // CPU frequency
	Nodes    int32  `json:"nodes"`     // Number of NUMA nodes
	Sockets  int32  `json:"sockets"`   // Number of CPU sockets
	Cores    int32  `json:"cores"`     // Cores per socket
	Threads  int32  `json:"threads"`   // Threads per core
	// Memory stats from NodeGetMemoryStats
	TotalMemory  uint64 `json:"total_memory"`  // Total memory
	FreeMemory   uint64 `json:"free_memory"`   // Free memory
	BuffMemory   uint64 `json:"buff_memory"`   // Buffered memory
	CachedMemory uint64 `json:"cached_memory"` // Cached memory
	// CPU stats from NodeGetCPUStats
	UserCPUTime   uint64 `json:"user_cpu_time"`   // User CPU time
	SystemCPUTime uint64 `json:"system_cpu_time"` // System CPU time
	IdleCPUTime   uint64 `json:"idle_cpu_time"`   // Idle CPU time
	IOWaitTime    uint64 `json:"iowait_time"`     // I/O wait time
	// Calculated metrics
	CPUPercent    float64 `json:"cpu_percent"`    // Overall CPU utilization
	MemoryPercent float64 `json:"memory_percent"` // Memory utilization
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// DevicePerformance stores device-specific performance metrics
type DevicePerformance struct {
	Base
	VMUUID     string `gorm:"index" json:"vm_uuid"`
	DeviceType string `json:"device_type"` // "video", "sound", "hostdev", etc.
	DeviceName string `json:"device_name"` // Device identifier
	// Generic performance metrics
	BandwidthUsed  uint64  `json:"bandwidth_used"`  // Used bandwidth
	BandwidthLimit uint64  `json:"bandwidth_limit"` // Bandwidth limit
	LatencyMs      float64 `json:"latency_ms"`      // Latency in milliseconds
	ErrorCount     uint64  `json:"error_count"`     // Error count
	// Device-specific JSON data
	MetricsJSON string `gorm:"type:text" json:"metrics_json"` // Device-specific metrics as JSON
	// Timestamp for metrics aging
	CollectedAt time.Time `json:"collected_at"`
}

// VideoModel represents a virtual display adapter template (shared model).
type VideoModel struct {
	gorm.Model
	ModelName string
	VRAM      uint
	Heads     int
	Accel3D   bool
}

// VideoAttachment links a Video model to a VM (monitor index / primary flag).
type VideoAttachment struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	VideoModelID uint
	VideoModel   VideoModel
	MonitorIndex int
	Primary      bool
}

// BootConfig stores loader / nvram / boot-order information for a VM.
type BootConfig struct {
	gorm.Model
	VMUUID        string `gorm:"index;unique"`
	LoaderPath    string
	LoaderType    string
	NVramPath     string
	BootOrderJSON string `gorm:"type:text"`
	SecureBoot    bool
}

// VendorOption stores vendor-specific options (e.g., qemu:) for arbitrary owners.
type VendorOption struct {
	gorm.Model
	OwnerType string `gorm:"size:64;index"` // e.g., "disk_attachment", "port"
	OwnerID   uint   `gorm:"index"`
	Namespace string `gorm:"size:64"` // e.g., "qemu"
	Key       string
	ValueJSON string `gorm:"type:text"`
}

// MediatedDevice represents an mdev type available on the host.
type MediatedDevice struct {
	gorm.Model
	TypeName   string
	Vendor     string
	DeviceID   string
	ConfigJSON string `gorm:"type:text"`
}

// MediatedDeviceAttachment links a mediated device instance to a VM.
type MediatedDeviceAttachment struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	MdevID      uint
	DeviceName  string
	AddressJSON string `gorm:"type:text"`
}

// HostCapability stores discovered host capabilities (KVM, qemu features, GPU support, etc.)
type HostCapability struct {
	gorm.Model
	HostID      string `gorm:"index"`
	Name        string
	Version     string
	DetailsJSON string `gorm:"type:text"`
}

// SRIOVPool represents an SR-IOV PF with its VF pool information.
type SRIOVPool struct {
	gorm.Model
	HostDeviceID uint
	PFAddress    string
	TotalVFs     int
	FreeVFs      int
	ConfigJSON   string `gorm:"type:text"`
}

// SRIOVFunction represents a single VF allocation from a PF.
type SRIOVFunction struct {
	gorm.Model
	HostDeviceID uint
	VFIndex      int
	Allocated    bool
	AllocVMUUID  string `gorm:"index"`
	ConfigJSON   string `gorm:"type:text"`
}

// GPUDevice represents a physical GPU (or virtual GPU) on the host.
type GPUDevice struct {
	gorm.Model
	HostDeviceID uint
	Vendor       string
	ModelName    string
	UUID         string
	ConfigJSON   string `gorm:"type:text"`
}

// --- Enhanced VM Configuration Models ---

// ResourceClass defines resource allocation templates for VMs
type ResourceClass struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Description string
	CPUCores    int
	MemoryMB    int
	StorageGB   int
	ConfigJSON  string `gorm:"type:text"` // Additional resource specifications
}

// HardwareTrait represents required hardware traits for VM placement
type HardwareTrait struct {
	gorm.Model
	VMUUID    string `gorm:"index"`
	TraitName string // 'avx2', 'ssd', 'sriov_nic', 'gpu_accel', etc.
	Required  bool
}

// PlacementPolicy defines VM placement and scheduling policies
type PlacementPolicy struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	PolicyType string // 'automatic', 'performance', 'balanced', 'power-saving'
	ConfigJSON string `gorm:"type:text"`
}

// QOSPolicy defines Quality of Service policies for storage and network
type QOSPolicy struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	ResourceType string // 'disk', 'network'
	ResourceID   string // disk ID or network interface ID
	ReadIOPS     *uint64
	WriteIOPS    *uint64
	ReadBPS      *uint64
	WriteBPS     *uint64
	ConfigJSON   string `gorm:"type:text"`
}

// GPUAttachment links a GPU device to a VM for passthrough or mediated assignments.
type GPUAttachment struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	GPUDeviceID uint
	DeviceName  string
	AddressJSON string `gorm:"type:text"`
}

// PCIeRootPort represents a PCIe root port configuration (useful for device addressing)
type PCIeRootPort struct {
	gorm.Model
	HostID     string `gorm:"index"`
	Slot       string
	Bus        string
	Function   string
	ConfigJSON string `gorm:"type:text"`
}

// MachineType captures the domain's machine type (pc/q35 etc) and architecture.
type MachineType struct {
	gorm.Model
	Name       string
	Arch       string
	Variant    string
	ConfigJSON string `gorm:"type:text"`
}

// VideoDevice represents a host-scoped physical video/GPU device.
type VideoDevice struct {
	gorm.Model
	HostDeviceID uint
	Vendor       string
	ModelName    string
	UUID         string
	ConfigJSON   string `gorm:"type:text"`
}

// VideoDeviceAttachment links a physical VideoDevice to a VM (exclusive).
type VideoDeviceAttachment struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	VideoDeviceID uint
	DeviceName    string
	AddressJSON   string `gorm:"type:text"`
}

// DiskDriverOptions provides typed disk driver options often found under <driver>.
type DiskDriverOptions struct {
	gorm.Model
	DiskID       uint
	Cache        string // 'none','writeback','writethrough'
	IO           string // 'native','threads'
	DetectZeroes string
	Discard      string
	Aio          string
	ConfigJSON   string `gorm:"type:text"`
}

// BlockDev represents a qemu blockdev object for advanced block configurations.
type BlockDev struct {
	gorm.Model
	NodeName    string `gorm:"uniqueIndex"`
	Driver      string
	FilePath    string
	Format      string
	OptionsJSON string `gorm:"type:text"`
}

// BackingStore represents layered backing store info for a disk/blockdev.
type BackingStore struct {
	gorm.Model
	DiskID     uint
	ParentNode string
	Format     string
	ConfigJSON string `gorm:"type:text"`
}

// NUMANode represents a single NUMA node attached to a VM.
type NUMANode struct {
	gorm.Model
	VMUUID   string `gorm:"index"`
	NodeID   int
	MemoryKB uint64
	CPUsJSON string `gorm:"type:text"` // list of CPU ids
}

// MemoryBacking stores memory backing details (hugepages, source) for a VM.
type MemoryBacking struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	SourceType string
	SizeKB     uint64
	ConfigJSON string `gorm:"type:text"`
}

// VFIODevice represents VFIO / PCI passthrough mapping information
type VFIODevice struct {
	gorm.Model
	HostDeviceID uint
	Group        string
	IommuGroup   string
	ConfigJSON   string `gorm:"type:text"`
}

// SCSIController represents a SCSI controller device (e.g., virtio-scsi, pci-scsi).
type SCSIController struct {
	gorm.Model
	ModelName  string
	Index      int
	Type       string // 'virtio-scsi', 'lsi', 'megasas'
	ConfigJSON string `gorm:"type:text"`
}

// SCSIControllerAttachment links a SCSIController to a VM (if controllers are modeled per-VM).
type SCSIControllerAttachment struct {
	gorm.Model
	VMUUID           string `gorm:"index"`
	SCSIControllerID uint
	AddressJSON      string `gorm:"type:text"`
}

// IOThread represents an I/O thread entity that can be pinned to vcpus and associated with block devices.
type IOThread struct {
	gorm.Model
	Name       string
	Affinity   string `gorm:"type:text"` // cpuset
	ConfigJSON string `gorm:"type:text"`
}

// IOThreadAttachment connects an IOThread to an owner (disk/blockdev).
type IOThreadAttachment struct {
	gorm.Model
	OwnerType  string `gorm:"size:64;index"`
	OwnerID    uint   `gorm:"index"`
	IOThreadID uint
}

// DeviceAddress provides a structured representation of a device PCI/USB address.
type DeviceAddress struct {
	gorm.Model
	Type     string // 'pci'|'usb' etc
	Domain   string
	Bus      string
	Slot     string
	Function string
	RawJSON  string `gorm:"type:text"`
}

// HostPCIDevice stores detailed PCI host device info (SR-IOV, capability metadata).
type HostPCIDevice struct {
	gorm.Model
	HostDeviceID  uint
	VendorID      string
	ProductID     string
	Slot          string
	Function      string
	SRIOVTotalVFs int
	SRIOVNumVFs   int
	ConfigJSON    string `gorm:"type:text"`
}

// CPUTune stores CPU tuning parameters that can be applied to a VM.
type CPUTune struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	Shares     *int
	Quota      *int
	Period     *int
	EmuThreads *int
	Vcpus      string `gorm:"type:text"` // CPU pinning set (e.g., "0-3,8-11")
}

// IOTune stores I/O throttling parameters (per-VM or per-disk).
type IOTune struct {
	gorm.Model
	OwnerType     string `gorm:"size:64;index"` // 'vm'|'disk' etc
	OwnerID       uint   `gorm:"index"`
	ReadIOPS      *int
	WriteIOPS     *int
	ReadBps       *int64
	WriteBps      *int64
	TotalBytesSec *int64
	ConfigJSON    string `gorm:"type:text"`
}

// QemuArg stores arbitrary qemu commandline/option entries attached to an owner.
type QemuArg struct {
	gorm.Model
	OwnerType string `gorm:"size:64;index"`
	OwnerID   uint   `gorm:"index"`
	Key       string
	Value     string `gorm:"type:text"`
}

// MdevType stores available mediated device types and capability metadata.
type MdevType struct {
	gorm.Model
	TypeName    string
	Description string
	Vendor      string
	ConfigJSON  string `gorm:"type:text"`
}

// --- Advanced Features ---

// VMSnapshot stores metadata about a VM snapshot.
type VMSnapshot struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	Name        string
	Description string
	ParentName  string
	State       string
	ConfigXML   string
}

// User represents a Virtumancer user account.
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	RoleID       uint
}

// Role defines a set of permissions.
type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission is a granular action that can be performed.
type Permission struct {
	gorm.Model
	Action      string `gorm:"uniqueIndex"`
	Description string
}

// Task tracks a long-running, asynchronous operation.
type Task struct {
	gorm.Model
	UserID   uint
	Type     string
	Status   string
	Progress int
	Details  string
}

// AuditLog records an event that occurred in the system.
type AuditLog struct {
	gorm.Model
	UserID     uint
	Action     string
	TargetType string
	TargetID   string
	Details    string
}

// Setting represents a simple key/value configuration entry.
// OwnerType/OwnerID allow scoping (e.g., 'user', 'host') for future extensibility.
type Setting struct {
	gorm.Model
	Key       string `gorm:"size:128;index" json:"key"`
	ValueJSON string `gorm:"type:text" json:"value_json"`
	OwnerType string `gorm:"size:64;index;default:global" json:"owner_type"`
	OwnerID   *uint  `json:"owner_id"` // nullable owner id
}

// AttachmentAllocation is an index table that maps VM attachments across device types
// to allow fast, aggregated queries ("all attachments for a VM") without scanning
// every per-device attachment table.
// AttachmentIndex provides a compact index of attachments across device types.
// The corresponding table name will be `attachment_indices` by GORM pluralization.
type AttachmentIndex struct {
	Base
	VMUUID       string  `gorm:"index;not null"`
	DeviceType   string  `gorm:"index;size:64;not null"` // e.g. 'volume', 'graphics', 'hostdevice'
	AttachmentID string  `gorm:"not null"`               // row id in the specific attachment table
	DeviceID     *string `gorm:"index"`                  // optional convenience: device's UUID (nullable for multi-attach)
}

// --- OS Configuration ---

// OSConfig stores OS-level configuration for a VM (loader, nvram, boot menu, etc.)
type OSConfig struct {
	Base
	VMUUID            string `gorm:"index;unique"`
	LoaderPath        string
	LoaderType        string // 'rom', 'pflash'
	LoaderReadonly    bool
	LoaderSecure      bool
	LoaderStateless   bool
	NVramPath         string
	NVramTemplate     string
	NVramType         string // 'file', 'block', 'network'
	BootMenuEnable    bool
	BootMenuTimeout   uint
	SMBIOSMode        string // 'emulate', 'host', 'sysinfo'
	Firmware          string // 'bios', 'efi'
	FirmwareFeatures  string `gorm:"type:text"` // JSON array of firmware features
	BIOSUsesSerial    bool
	BIOSRebootTimeout uint
}

// SMBIOSSystemInfo stores SMBIOS system information.
type SMBIOSSystemInfo struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	Type         string // 'bios', 'system', 'baseBoard', 'chassis', 'oemStrings', 'fwcfg'
	Vendor       string
	Version      string
	Serial       string
	Product      string
	Manufacturer string
	Asset        string
	SKU          string
	ConfigJSON   string `gorm:"type:text"` // Additional configuration as JSON
}

// --- CPU Configuration ---

// CPUFeature stores individual CPU features for a VM.
type CPUFeature struct {
	gorm.Model
	VMUUID  string `gorm:"index"`
	Name    string
	Policy  string // 'force', 'require', 'optional', 'disable', 'forbid'
	Default bool   // Whether this is a default feature
}

// CPUTopology stores CPU topology information (sockets, cores, threads).
type CPUTopology struct {
	gorm.Model
	VMUUID  string `gorm:"index;unique"`
	Sockets uint
	Cores   uint
	Threads uint
}

// --- Memory Configuration ---

// MemoryConfig consolidates memory backing and NUMA configuration.
// config_type determines the type: 'backing', 'numa', 'tuning'
type MemoryConfig struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	ConfigType string `gorm:"size:32"` // 'backing', 'numa', 'tuning'
	// Backing fields
	SourceType   string
	SizeKB       uint64
	Mode         string // 'shared', 'private'
	Nosharepages bool
	Locked       bool
	// NUMA fields
	NodeID   int
	MemoryKB uint64
	CPUs     string // CPU list for NUMA node
	// Tuning fields
	MinGuarantee uint64
	// Common fields
	ConfigJSON string `gorm:"type:text"` // Additional configuration
}

// --- Security Configuration ---

// SecurityLabel stores security label configuration for a VM.
type SecurityLabel struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	Type          string // 'dynamic', 'static', 'none'
	SecurityModel string `gorm:"column:model"` // 'selinux', 'apparmor', 'dac'
	Label         string
	Baselabel     string
	Relabel       bool
}

// LaunchSecurity stores launch security configuration (SEV, SEV-SNP, etc.)
type LaunchSecurity struct {
	gorm.Model
	VMUUID string `gorm:"index;unique"`
	Type   string // 'sev', 'sev-snp', 's390-pv'
	// SEV/SEV-SNP fields
	CBitPos         uint
	ReducedPhysBits uint
	Policy          uint64
	DHCert          string
	Session         string
	// SEV-SNP specific
	AuthorKey bool
	VCEK      bool
	IDBlock   string
	IDAuth    string
	HostData  string
	// S390-PV fields
	// (minimal for now, can be extended)
}

// --- Hypervisor Features ---

// HypervisorFeature stores hypervisor-specific features (KVM, Xen, Hyper-V).
type HypervisorFeature struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	Type       string // 'kvm', 'xen', 'hyperv', 'pvspinlock', etc.
	Name       string
	State      string // 'on', 'off', 'default'
	Value      string // Optional value for features that need it
	ConfigJSON string `gorm:"type:text"` // Additional configuration
}

// --- Domain Lifecycle ---

// LifecycleAction stores lifecycle event actions (on_poweroff, on_reboot, etc.)
type LifecycleAction struct {
	gorm.Model
	VMUUID        string `gorm:"index;unique"`
	OnPoweroff    string // 'destroy', 'restart', 'preserve', 'rename-restart'
	OnReboot      string // 'destroy', 'restart', 'preserve', 'rename-restart'
	OnCrash       string // 'destroy', 'restart', 'preserve', 'rename-restart', 'coredump-destroy', 'coredump-restart'
	OnLockfailure string // 'poweroff', 'restart', 'pause', 'ignore'
}

// --- Clock Configuration ---

// Clock stores clock configuration for a VM.
type Clock struct {
	gorm.Model
	VMUUID     string `gorm:"index;unique"`
	Offset     string // 'utc', 'localtime', 'timezone', 'variable'
	Timezone   string // Timezone name when offset='timezone'
	Basis      string // 'utc' or 'localtime' when offset='variable'
	Adjustment int64  // Seconds adjustment when offset='variable'
	ConfigJSON string `gorm:"type:text"` // Timer configurations
}

// --- Performance Monitoring ---

// PerfEvent stores performance monitoring event configuration.
type PerfEvent struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	Name       string
	State      string // 'on', 'off'
	ConfigJSON string `gorm:"type:text"` // Additional configuration
}

// InitDB initializes and returns a GORM database instance.
func InitDB(dataSourceName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Configure GORM logger to warn level (only log warnings and errors, not info queries)
	db.Logger = logger.Default.LogMode(logger.Warn)

	// Auto-migrate the full schema
	err = db.AutoMigrate(
		&Host{},
		&VirtualMachine{},
		&StoragePool{},
		&Volume{},
		&Network{},
		&Port{},
		&PortBinding{},
		&Controller{},
		&ControllerAttachment{},
		&InputDevice{},
		&InputDeviceAttachment{},
		// graphics device legacy models removed; Console replaces them
		&SoundCard{},
		&SoundCardAttachment{},
		&HostDevice{},
		&HostDeviceAttachment{},
		&TPM{},
		&TPMAttachment{},
		&Watchdog{},
		&WatchdogAttachment{},
		&SerialDevice{},
		&SerialDeviceAttachment{},
		&ChannelDevice{},
		&ChannelDeviceAttachment{},
		&Filesystem{},
		&FilesystemAttachment{},
		&Smartcard{},
		&SmartcardAttachment{},
		&USBRedirector{},
		&USBRedirectorAttachment{},
		&RngDevice{},
		&RngDeviceAttachment{},
		&PanicDevice{},
		&PanicDeviceAttachment{},
		&Vsock{},
		&VsockAttachment{},
		&MemoryBalloon{},
		&MemoryBalloonAttachment{},
		&ShmemDevice{},
		&ShmemDeviceAttachment{},
		&IOMMUDevice{},
		&IOMMUDeviceAttachment{},
		&PortAttachment{},
		&FilterRef{},
		&VirtualPort{},
		&DeviceAlias{},
		&Disk{},
		&DiskAttachment{},
		&BlockStatistics{},
		&NetworkStatistics{},
		&CPUPerformance{},
		&MemoryPerformance{},
		&NodePerformance{},
		&DevicePerformance{},
		&VideoModel{},
		&VideoAttachment{},
		&VideoDevice{},
		&VideoDeviceAttachment{},
		&BootConfig{},
		&VendorOption{},
		&MediatedDevice{},
		&MediatedDeviceAttachment{},
		&VMSnapshot{},
		&AttachmentIndex{},
		&Console{},
		&DeviceAddress{},
		&HostPCIDevice{},
		&CPUTune{},
		&IOTune{},
		&QemuArg{},
		&MdevType{},
		&DiskDriverOptions{},
		&BlockDev{},
		&BackingStore{},
		&NUMANode{},
		&MemoryBacking{},
		&VFIODevice{},
		&SCSIController{},
		&SCSIControllerAttachment{},
		&IOThread{},
		&IOThreadAttachment{},
		// New OS Configuration tables
		&OSConfig{},
		&SMBIOSSystemInfo{},
		// New CPU Configuration tables
		&CPUFeature{},
		&CPUTopology{},
		// New Memory Configuration (consolidated)
		&MemoryConfig{},
		// New Security tables
		&SecurityLabel{},
		&LaunchSecurity{},
		// New Hypervisor Features table
		&HypervisorFeature{},
		// New Lifecycle table
		&LifecycleAction{},
		// New Clock table
		&Clock{},
		// New Performance Monitoring table
		&PerfEvent{},
		&User{},
		&Role{},
		&Permission{},
		&Task{},
		&AuditLog{},
		&Setting{},
		&DiscoveredVM{},
		// Host Capability and SR-IOV Management
		&HostCapability{},
		&SRIOVPool{},
		&SRIOVFunction{},
		// Enhanced VM Configuration Models
		&ResourceClass{},
		&HardwareTrait{},
		&PlacementPolicy{},
		&QOSPolicy{},
	)
	if err != nil {
		return nil, err
	}

	// Ensure indexes / unique constraints for the attachment index for fast queries
	// and to prevent duplicate entries. Run after AutoMigrate so tables exist.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_attachment_index ON attachment_indices(device_type, attachment_id);").Error; err != nil {
		log.Errorf("failed to create unique index uniq_attachment_index: %v", err)
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_attachment_index_vm_uuid ON attachment_indices(vm_uuid);").Error; err != nil {
		log.Errorf("failed to create index idx_attachment_index_vm_uuid: %v", err)
		return nil, err
	}

	// Optional: prevent the same device (by device_type + device_id) from being allocated multiple times.
	// This covers per-instance device types such as `console` (and other non-volume types).
	// Volumes can be multi-attached so exclude them from this unique constraint using a partial index.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_attachment_index_device ON attachment_indices(device_type, device_id) WHERE device_type != 'volume' AND device_id IS NOT NULL;").Error; err != nil {
		log.Errorf("failed to create unique index uniq_attachment_index_device: %v", err)
		return nil, err
	}

	// Normalize old data: convert device_id == 0 to NULL so multi-attach rows
	// are represented as NULL device_id (we changed DeviceID to *uint).
	if err := db.Exec("UPDATE attachment_indices SET device_id = NULL WHERE device_id = 0").Error; err != nil {
		log.Errorf("failed to normalize attachment_indices device_id zeros: %v", err)
		return nil, err
	}

	// Index PortAttachment.HostID for efficient host-scoped attachment queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_port_attachments_host_id ON port_attachments(host_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_port_attachments_host_id: %v", err)
		return nil, err
	}

	// Create partial indexes useful for fast lookups on ports and attachments.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_port_host_mac ON ports(host_id, mac_address) WHERE host_id IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_port_host_mac: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_port_attachment_vm_dev ON port_attachments(vm_uuid, device_name) WHERE vm_uuid IS NOT NULL AND device_name IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_port_attachment_vm_dev: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_port_attachment_vm_mac ON port_attachments(vm_uuid, mac_address) WHERE vm_uuid IS NOT NULL AND mac_address IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_port_attachment_vm_mac: %v", err)
		return nil, err
	}

	// Auxiliary tables indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_filterref_port ON filter_refs(port_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_filterref_port: %v", err)
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_virtualport_port ON virtual_ports(port_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_virtualport_port: %v", err)
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_devicealias_vm ON device_aliases(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_devicealias_vm: %v", err)
		return nil, err
	}

	// New attachment indexes
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_disk_attachment_vm_dev ON disk_attachments(vm_uuid, device_name) WHERE vm_uuid IS NOT NULL AND device_name IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_disk_attachment_vm_dev: %v", err)
		return nil, err
	}

	// Discovered VMs table indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_discovered_vms_host_id ON discovered_vms(host_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_discovered_vms_host_id: %v", err)
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_discovered_vms_domain_uuid ON discovered_vms(domain_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_discovered_vms_domain_uuid: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_video_attachment_vm_monitor ON video_attachments(vm_uuid, monitor_index) WHERE vm_uuid IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_video_attachment_vm_monitor: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_bootconfig_vm_uuid ON boot_configs(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_bootconfig_vm_uuid: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vendoroption_owner ON vendor_options(owner_type, owner_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_vendoroption_owner: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_deviceaddress_type ON device_addresses(type);").Error; err != nil {
		log.Verbosef("failed to create index idx_deviceaddress_type: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_cputune_vm ON cpu_tunes(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_cputune_vm: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_iotune_owner ON io_tunes(owner_type, owner_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_iotune_owner: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_qemuarg_owner ON qemu_args(owner_type, owner_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_qemuarg_owner: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_mdevtype_name ON mdev_types(type_name);").Error; err != nil {
		log.Verbosef("failed to create index idx_mdevtype_name: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_blockdev_nodename ON block_devs(node_name);").Error; err != nil {
		log.Verbosef("failed to create index idx_blockdev_nodename: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_numa_vm ON numa_nodes(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_numa_vm: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_memorybacking_vm ON memory_backings(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_memorybacking_vm: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vfio_hostdevice ON vfio_devices(host_device_id);").Error; err != nil {
		log.Verbosef("failed to create index idx_vfio_hostdevice: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scsicontroller_model ON scsi_controllers(model_name);").Error; err != nil {
		log.Verbosef("failed to create index idx_scsicontroller_model: %v", err)
		return nil, err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_iothread_name ON io_threads(name);").Error; err != nil {
		log.Verbosef("failed to create index idx_iothread_name: %v", err)
		return nil, err
	}

	return db, nil
}
