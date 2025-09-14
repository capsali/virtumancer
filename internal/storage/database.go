package storage

import (
	log "github.com/capsali/virtumancer/internal/logging"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// VMState defines the possible stable states of a VM.
type VMState string

const (
	StateInitialized VMState = "INITIALIZED"
	StateActive      VMState = "ACTIVE"
	StatePaused      VMState = "PAUSED"
	StateSuspended   VMState = "SUSPENDED"
	StateStopped     VMState = "STOPPED"
	StateError       VMState = "ERROR"
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
	ID  string `gorm:"primaryKey" json:"id"`
	URI string `json:"uri"`
}

// VirtualMachine is Virtumancer's canonical definition of a VM's intended state.
type VirtualMachine struct {
	gorm.Model
	HostID          string `gorm:"uniqueIndex:idx_vm_host_name"`
	Name            string `gorm:"uniqueIndex:idx_vm_host_name"`
	UUID            string `gorm:"primaryKey"`  // Virtumancer's internal, guaranteed-unique UUID
	DomainUUID      string `gorm:"uniqueIndex"` // The UUID as reported by libvirt
	Description     string
	State           VMState     `gorm:"default:'INITIALIZED'"` // Stable state
	TaskState       VMTaskState // Transient state during operations
	VCPUCount       uint
	CPUModel        string
	CPUTopologyJSON string
	MemoryBytes     uint64
	OSType          string
	IsTemplate      bool
	SyncStatus      SyncStatus `gorm:"default:'UNKNOWN'"`
	DriftDetails    string     // JSON blob storing drift information
	NeedsRebuild    bool       `gorm:"default:false"`
}

// --- Storage Management ---

// StoragePool represents a libvirt storage pool (e.g., LVM, a directory).
type StoragePool struct {
	gorm.Model
	HostID          string
	Name            string
	UUID            string `gorm:"uniqueIndex"`
	Type            string
	Path            string
	CapacityBytes   uint64
	AllocationBytes uint64
}

// Volume represents a single storage volume, like a virtual disk or an ISO.
type Volume struct {
	gorm.Model
	StoragePoolID   uint
	Name            string
	Type            string // 'DISK' or 'ISO'
	Format          string
	CapacityBytes   uint64
	AllocationBytes uint64
}

// VolumeAttachment links a Volume to a VirtualMachine.
type VolumeAttachment struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	VolumeID   uint
	Volume     Volume
	DeviceName string // e.g., "vda", "hdb"
	BusType    string // e.g., "virtio", "sata", "ide"
	IsReadOnly bool
}

// --- Network Management ---

// Network represents a virtual network or bridge on a host.
type Network struct {
	gorm.Model
	HostID     string `gorm:"uniqueIndex:idx_network_host_name"`
	Name       string `gorm:"uniqueIndex:idx_network_host_name"`
	UUID       string
	BridgeName string
	Mode       string // e.g., 'bridged', 'nat', 'isolated'
}

// Port represents a virtual Network Interface Card (vNIC) belonging to a VM.
type Port struct {
	gorm.Model
	VMUUID     string `gorm:"uniqueIndex:idx_port_vm_mac"`
	MACAddress string `gorm:"uniqueIndex:idx_port_vm_mac"`
	DeviceName string // e.g. "vnet0", "eth0"
	ModelName  string // e.g., 'virtio', 'e1000'
	IPAddress  string
}

// PortBinding links a Port to a Network.
type PortBinding struct {
	gorm.Model
	PortID    uint
	Port      Port
	NetworkID uint
	Network   Network
}

// --- Virtual Hardware Management ---

// Controller represents a hardware controller within a VM (e.g., USB, SATA).
type Controller struct {
	gorm.Model
	Type      string // 'usb', 'sata', 'virtio-serial'
	ModelName string
	Index     uint
}

// ControllerAttachment links a Controller to a VirtualMachine.
type ControllerAttachment struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	ControllerID uint
}

// InputDevice represents an input device like a mouse or keyboard.
type InputDevice struct {
	gorm.Model
	Type string // 'mouse', 'tablet', 'keyboard'
	Bus  string // 'usb', 'ps2', 'virtio'
}

// InputDeviceAttachment links an InputDevice to a VirtualMachine.
type InputDeviceAttachment struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	InputDeviceID uint
}

// GraphicsDevice represented a virtual GPU and display protocol (legacy).
// It was removed in favor of per-VM `Console` instances.
// Console represents a per-VM console instance (VNC/SPICE) discovered from domain XML.
// This is a per-instance model used during migration from the older
// GraphicsDevice/GraphicsDeviceAttachment model to a first-class Console instance.
type Console struct {
	gorm.Model
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
	gorm.Model
	ModelName string // 'ich6', 'ac97'
}

// SoundCardAttachment links a SoundCard to a VirtualMachine.
type SoundCardAttachment struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	SoundCardID uint
}

// HostDevice represents a physical device on a host for passthrough.
type HostDevice struct {
	gorm.Model
	HostID      string
	Type        string // 'pci', 'usb'
	Address     string // Physical address on host
	Description string
}

// HostDeviceAttachment links a HostDevice to a VirtualMachine for passthrough.
type HostDeviceAttachment struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	HostDeviceID uint
}

// TPM represents a Trusted Platform Module device.
type TPM struct {
	gorm.Model
	ModelName   string // 'tpm-crb', 'tpm-tis'
	BackendType string // 'passthrough', 'emulator'
	BackendPath string
}

// TPMAttachment links a TPM to a VirtualMachine.
type TPMAttachment struct {
	gorm.Model
	VMUUID string `gorm:"index"`
	TPMID  uint
}

// Watchdog represents a virtual watchdog device.
type Watchdog struct {
	gorm.Model
	ModelName string // 'i6300esb'
	Action    string // 'reset', 'shutdown', 'poweroff'
}

// WatchdogAttachment links a Watchdog to a VirtualMachine.
type WatchdogAttachment struct {
	gorm.Model
	VMUUID     string `gorm:"index"`
	WatchdogID uint
}

// SerialDevice represents a serial port configuration.
type SerialDevice struct {
	gorm.Model
	Type       string // 'pty', 'tcp', 'stdio'
	TargetPort uint
	ConfigJSON string
}

// SerialDeviceAttachment links a SerialDevice to a VirtualMachine.
type SerialDeviceAttachment struct {
	gorm.Model
	VMUUID         string `gorm:"index"`
	SerialDeviceID uint
}

// ChannelDevice represents a communication channel (e.g., for guest agent).
type ChannelDevice struct {
	gorm.Model
	Type       string // 'unix', 'spicevmc'
	TargetName string // e.g., 'org.qemu.guest_agent.0'
	ConfigJSON string
}

// ChannelDeviceAttachment links a ChannelDevice to a VirtualMachine.
type ChannelDeviceAttachment struct {
	gorm.Model
	VMUUID          string `gorm:"index"`
	ChannelDeviceID uint
}

// Filesystem represents a shared filesystem for a VM.
type Filesystem struct {
	gorm.Model
	DriverType string
	SourcePath string
	TargetPath string
}

// FilesystemAttachment links a Filesystem to a VM.
type FilesystemAttachment struct {
	gorm.Model
	VMUUID       string `gorm:"index"`
	FilesystemID uint
}

// Smartcard represents a smartcard device for a VM.
type Smartcard struct {
	gorm.Model
	Type       string
	ConfigJSON string
}

// SmartcardAttachment links a Smartcard to a VM.
type SmartcardAttachment struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	SmartcardID uint
}

// USBRedirector represents a USB redirection device.
type USBRedirector struct {
	gorm.Model
	Type       string
	FilterRule string
}

// USBRedirectorAttachment links a USBRedirector to a VM.
type USBRedirectorAttachment struct {
	gorm.Model
	VMUUID          string `gorm:"index"`
	USBRedirectorID uint
}

// RngDevice represents a Random Number Generator device.
type RngDevice struct {
	gorm.Model
	ModelName   string
	BackendType string
}

// RngDeviceAttachment links an RngDevice to a VM.
type RngDeviceAttachment struct {
	gorm.Model
	VMUUID      string `gorm:"index"`
	RngDeviceID uint
}

// PanicDevice represents a panic device for a VM.
type PanicDevice struct {
	gorm.Model
	ModelName string
}

// PanicDeviceAttachment links a PanicDevice to a VM.
type PanicDeviceAttachment struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	PanicDeviceID uint
}

// Vsock represents a VirtIO socket device.
type Vsock struct {
	gorm.Model
	GuestCID uint
}

// VsockAttachment links a Vsock to a VM.
type VsockAttachment struct {
	gorm.Model
	VMUUID  string `gorm:"index"`
	VsockID uint
}

// MemoryBalloon represents a memory balloon device.
type MemoryBalloon struct {
	gorm.Model
	ModelName  string
	ConfigJSON string
}

// MemoryBalloonAttachment links a MemoryBalloon to a VM.
type MemoryBalloonAttachment struct {
	gorm.Model
	VMUUID          string `gorm:"index"`
	MemoryBalloonID uint
}

// ShmemDevice represents a shared memory device.
type ShmemDevice struct {
	gorm.Model
	Name    string
	SizeKiB uint
	Path    string
}

// ShmemDeviceAttachment links a ShmemDevice to a VM.
type ShmemDeviceAttachment struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	ShmemDeviceID uint
}

// IOMMUDevice represents an IOMMU device.
type IOMMUDevice struct {
	gorm.Model
	ModelName string
}

// IOMMUDeviceAttachment links an IOMMUDevice to a VM.
type IOMMUDeviceAttachment struct {
	gorm.Model
	VMUUID        string `gorm:"index"`
	IOMMUDeviceID uint
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

// AttachmentAllocation is an index table that maps VM attachments across device types
// to allow fast, aggregated queries ("all attachments for a VM") without scanning
// every per-device attachment table.
// AttachmentIndex provides a compact index of attachments across device types.
// The corresponding table name will be `attachment_indices` by GORM pluralization.
type AttachmentIndex struct {
	gorm.Model
	VMUUID       string `gorm:"index;not null"`
	DeviceType   string `gorm:"index;size:64;not null"` // e.g. 'volume', 'graphics', 'hostdevice'
	AttachmentID uint   `gorm:"not null"`               // row id in the specific attachment table
	DeviceID     uint   `gorm:"index"`                  // optional convenience: device's numeric id
}

// InitDB initializes and returns a GORM database instance.
func InitDB(dataSourceName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the full schema
	err = db.AutoMigrate(
		&Host{},
		&VirtualMachine{},
		&StoragePool{},
		&Volume{},
		&VolumeAttachment{},
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
		&VMSnapshot{},
		&AttachmentIndex{},
		&Console{},
		&User{},
		&Role{},
		&Permission{},
		&Task{},
		&AuditLog{},
	)
	if err != nil {
		return nil, err
	}

	// Ensure indexes / unique constraints for the attachment index for fast queries
	// and to prevent duplicate entries. Run after AutoMigrate so tables exist.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_attachment_index ON attachment_indices(device_type, attachment_id);").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_attachment_index: %v", err)
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_attachment_index_vm_uuid ON attachment_indices(vm_uuid);").Error; err != nil {
		log.Verbosef("failed to create index idx_attachment_index_vm_uuid: %v", err)
		return nil, err
	}

	// Optional: prevent the same device (by device_type + device_id) from being allocated multiple times.
	// This covers per-instance device types such as `console` (and other non-volume types).
	// Volumes can be multi-attached so exclude them from this unique constraint using a partial index.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uniq_attachment_index_device ON attachment_indices(device_type, device_id) WHERE device_type != 'volume' AND device_id IS NOT NULL;").Error; err != nil {
		log.Verbosef("failed to create unique index uniq_attachment_index_device: %v", err)
		return nil, err
	}

	return db, nil
}
