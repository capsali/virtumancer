package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/capsali/virtumancer/internal/logging"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VMView is a combination of DB data and live libvirt data for the frontend.
type VMView struct {
	// From DB
	Name            string              `json:"name"`
	UUID            string              `json:"uuid"`
	DomainUUID      string              `json:"domain_uuid"`
	Description     string              `json:"description"`
	VCPUCount       uint                `json:"vcpu_count"`
	MemoryBytes     uint64              `json:"memory_bytes"`
	IsTemplate      bool                `json:"is_template"`
	CPUModel        string              `json:"cpu_model"`
	CPUTopologyJSON string              `json:"cpu_topology_json"`
	OSType          string              `json:"os_type"`
	TaskState       storage.VMTaskState `json:"task_state"`

	// NEW: Drift detection fields
	SyncStatus   storage.SyncStatus `json:"sync_status"`
	DriftDetails string             `json:"drift_details"`
	NeedsRebuild bool               `json:"needs_rebuild"`

	// Timestamps from gorm.Model
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// From Libvirt or DB cache
	State        storage.VMState       `json:"state"`
	LibvirtState storage.VMState       `json:"libvirtState"` // Observed state from libvirt
	Graphics     libvirt.GraphicsInfo  `json:"graphics"`
	Hardware     *libvirt.HardwareInfo `json:"hardware,omitempty"` // Pointer to allow for null

	// From Libvirt (live data, only in some calls)
	MaxMem  uint64 `json:"max_mem"`
	Memory  uint64 `json:"memory"`
	CpuTime uint64 `json:"cpu_time"`
	Uptime  int64  `json:"uptime"`

	// Computed fields for frontend display
	DiskSizeGB       float64 `json:"disk_size_gb,omitempty"`
	NetworkInterface string  `json:"network_interface,omitempty"`
}

// ProcessedVMStats holds frontend-friendly VM statistics with calculated metrics
type ProcessedVMStats struct {
	// cpu_percent: percentage of host total (smoothed)
	CPUPercent float64 `json:"cpu_percent"`
	// cpu_percent_core: percentage relative to a single core (can exceed 100)
	CPUPercentCore float64 `json:"cpu_percent_core,omitempty"`
	// pcentbase (raw): sum vcpu cpu-time delta normalized to one CPU (un-normalized)
	CPUPercentRaw float64 `json:"cpu_percent_raw,omitempty"`
	// guest-normalized percent (virt-manager's cpuGuestPercent)
	CPUPercentGuest float64 `json:"cpu_percent_guest,omitempty"`
	// host-normalized percent (virt-manager's cpuHostPercent)
	CPUPercentHost float64 `json:"cpu_percent_host,omitempty"`
	MemoryMB       float64 `json:"memory_mb"`
	DiskReadMB     float64 `json:"disk_read_mb"`
	DiskWriteMB    float64 `json:"disk_write_mb"`
	// Disk read/write rates in KiB/s
	DiskReadKiBPerSec  float64 `json:"disk_read_kib_per_sec"`
	DiskWriteKiBPerSec float64 `json:"disk_write_kib_per_sec"`
	// Disk IOPS (requests/sec summed across devices)
	DiskReadIOPS  float64 `json:"disk_read_iops"`
	DiskWriteIOPS float64 `json:"disk_write_iops"`
	NetworkRxMB   float64 `json:"network_rx_mb"`
	NetworkTxMB   float64 `json:"network_tx_mb"`
	NetworkRxMBps float64 `json:"network_rx_mbps"`
	NetworkTxMBps float64 `json:"network_tx_mbps"`
	Uptime        int64   `json:"uptime"`
}

// DashboardStats holds aggregated system-wide statistics for the dashboard.
type DashboardStats struct {
	Infrastructure struct {
		TotalHosts     int `json:"totalHosts"`
		ConnectedHosts int `json:"connectedHosts"`
		TotalVMs       int `json:"totalVMs"`
		RunningVMs     int `json:"runningVMs"`
		StoppedVMs     int `json:"stoppedVMs"`
	} `json:"infrastructure"`
	Resources struct {
		TotalMemoryGB     float64 `json:"totalMemoryGB"`
		UsedMemoryGB      float64 `json:"usedMemoryGB"`
		MemoryUtilization float64 `json:"memoryUtilization"`
		TotalCPUs         int     `json:"totalCPUs"`
		AllocatedCPUs     int     `json:"allocatedCPUs"`
		CPUUtilization    float64 `json:"cpuUtilization"`
	} `json:"resources"`
	Health struct {
		SystemStatus string `json:"systemStatus"`
		LastSync     string `json:"lastSync"`
		Errors       int    `json:"errors"`
		Warnings     int    `json:"warnings"`
	} `json:"health"`
}

// ActivityEntry represents a system activity event for the dashboard.
type ActivityEntry struct {
	ID        string `json:"id"`
	Type      string `json:"type"`      // 'vm_state_change', 'host_connect', 'host_disconnect', etc.
	Message   string `json:"message"`   // Human-readable description
	HostID    string `json:"hostId"`    // Host involved in the activity
	VMUUID    string `json:"vmUuid"`    // VM UUID if relevant
	VMName    string `json:"vmName"`    // VM name if relevant
	Timestamp string `json:"timestamp"` // ISO timestamp
	Severity  string `json:"severity"`  // 'info', 'warning', 'error'
	Details   string `json:"details"`   // Optional additional context
}

// PortAttachmentView is a transport-friendly view of a PortAttachment including
// the underlying Port and (optional) Network information for the UI.
type PortAttachmentView struct {
	ID         uint             `json:"id"`
	VMUUID     string           `json:"vm_uuid"`
	DeviceName string           `json:"device_name"`
	MACAddress string           `json:"mac_address"`
	ModelName  string           `json:"model_name"`
	HostID     string           `json:"host_id"`
	Ordinal    int              `json:"ordinal"`
	Metadata   string           `json:"metadata"`
	Port       storage.Port     `json:"port"`
	Network    *storage.Network `json:"network,omitempty"`
}

// VmSubscription holds the clients subscribed to a VM's stats and a channel to stop polling.
type VmSubscription struct {
	clients        map[*ws.Client]bool
	stop           chan struct{}
	lastKnownStats *libvirt.VMStats
	mu             sync.RWMutex
}

// HostSubscription holds the clients subscribed to a Host's stats and a channel to stop polling.
type HostSubscription struct {
	clients        map[*ws.Client]bool
	stop           chan struct{}
	lastKnownStats *libvirt.HostStats
	mu             sync.RWMutex
}

// MonitoringManager handles real-time VM stat subscriptions.
type MonitoringManager struct {
	mu            sync.Mutex
	subscriptions map[string]*VmSubscription // key is "hostId:vmName"
	service       *HostService               // back-reference
}

// HostMonitoringManager handles real-time VM stat subscriptions.
type HostMonitoringManager struct {
	mu            sync.Mutex
	subscriptions map[string]*HostSubscription // key is "hostId"
	service       *HostService                 // back-reference
}

// NewMonitoringManager creates a new manager.
func NewMonitoringManager(service *HostService) *MonitoringManager {
	return &MonitoringManager{
		subscriptions: make(map[string]*VmSubscription),
		service:       service,
	}
}

// NewHostMonitoringManager creates a new manager.
func NewHostMonitoringManager(service *HostService) *HostMonitoringManager {
	return &HostMonitoringManager{
		subscriptions: make(map[string]*HostSubscription),
		service:       service,
	}
}

type HostServiceProvider interface {
	ws.InboundMessageHandler
	GetAllHosts() ([]storage.Host, error)
	EnsureHostConnected(hostID string) error
	EnsureHostConnectedForced(hostID string) error
	DisconnectHost(hostID string, userInitiated bool) error
	GetHostInfo(hostID string) (*libvirt.HostInfo, error)
	GetHostStats(hostID string) (*libvirt.HostStats, error)
	AddHost(host storage.Host) (*storage.Host, error)
	RemoveHost(hostID string) error
	ConnectToAllHosts()
	GetVMsForHostFromDB(hostID string) ([]VMView, error)
	GetVMStats(hostID, vmName string) (*ProcessedVMStats, error)
	GetVMHardwareAndDetectDrift(hostID, vmName string) (*libvirt.HardwareInfo, error)
	UpdateVMState(hostID, vmName, state string) error
	GetPortsForHostFromDB(hostID string) ([]storage.Port, error)
	GetPortAttachmentsForVM(vmUUID string) ([]PortAttachmentView, error)
	SyncVMsForHost(hostID string)
	// Discovery and import helpers
	ListDiscoveredVMs(hostID string) ([]libvirt.VMInfo, error)
	ImportVM(hostID, vmName string) error
	ImportAllVMs(hostID string) error
	ImportSelectedVMs(hostID string, domainUUIDs []string) error
	DeleteSelectedDiscoveredVMs(hostID string, domainUUIDs []string) error
	SyncVMFromLibvirt(hostID, vmName string) error
	RebuildVMFromDB(hostID, vmName string) error
	StartVM(hostID, vmName string) error
	ShutdownVM(hostID, vmName string) error
	RebootVM(hostID, vmName string) error
	ForceOffVM(hostID, vmName string) error
	ForceResetVM(hostID, vmName string) error
	// Dashboard methods
	GetDashboardStats() (*DashboardStats, error)
	GetDashboardActivity(limit int) ([]ActivityEntry, error)
}

type HostService struct {
	db              *gorm.DB
	connector       *libvirt.Connector
	hub             *ws.Hub
	monitor         *MonitoringManager
	hostMonitor     *HostMonitoringManager
	syncMutex       sync.Map // map[string]*sync.Mutex for per-host sync locking
	lastSync        sync.Map // map[string]time.Time for last sync time
	vmPollers       sync.Map // map[string]chan struct{} for stopping VM state polling
	prevCpuSamples  sync.Map // key: "hostID:vmName" -> struct{cpuTime uint64; at time.Time}
	prevDiskSamples sync.Map // key: "hostID:vmName" -> struct{readBytes int64; writeBytes int64; readReq int64; writeReq int64; at time.Time}
	prevNetSamples  sync.Map // key: "hostID:vmName" -> struct{rxBytes int64; txBytes int64; at time.Time}
	hostCores       sync.Map // key: hostID -> uint (number of cores)
	// smoothing state: store last smoothed host-normalized percent per vm
	cpuSmoothStore sync.Map // key: "hostID:vmName" -> float64
	// disk smoothing store: key: "hostID:vmName" -> struct{read float64; write float64; readIOPS float64; writeIOPS float64}
	diskSmoothStore sync.Map
	// network smoothing store: key: "hostID:vmName" -> struct{rx float64; tx float64}
	netSmoothStore sync.Map
	// smoothing alpha (0..1) - higher = more responsive, lower = smoother
	cpuSmoothAlpha  float64
	netSmoothAlpha  float64
	diskSmoothAlpha float64
}

func NewHostService(db *gorm.DB, connector *libvirt.Connector, hub *ws.Hub) *HostService {
	s := &HostService{
		db:        db,
		connector: connector,
		hub:       hub,
	}
	s.monitor = NewMonitoringManager(s)
	s.hostMonitor = NewHostMonitoringManager(s)
	// default smoothing alpha
	s.cpuSmoothAlpha = 0.3
	// default network smoothing alpha (more responsive)
	s.netSmoothAlpha = 0.6
	// default disk smoothing alpha
	s.diskSmoothAlpha = 0.3
	return s
}

// ReloadMetricsSettings reads persisted global metrics settings from the database
// and applies them to the HostService smoothing configuration. It is safe to
// call at startup and after settings changes.
func (s *HostService) ReloadMetricsSettings() error {
	var st storage.Setting
	if err := s.db.Where("key = ?", "metrics:global").First(&st).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No settings persisted; keep defaults
			return nil
		}
		return err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(st.ValueJSON), &payload); err != nil {
		return err
	}

	if v, ok := payload["diskSmoothAlpha"]; ok {
		if num, ok2 := v.(float64); ok2 {
			if num < 0 {
				num = 0
			}
			if num > 1 {
				num = 1
			}
			s.diskSmoothAlpha = num
		}
	}
	if v, ok := payload["netSmoothAlpha"]; ok {
		if num, ok2 := v.(float64); ok2 {
			if num < 0 {
				num = 0
			}
			if num > 1 {
				num = 1
			}
			s.netSmoothAlpha = num
		}
	}
	if v, ok := payload["cpuSmoothAlpha"]; ok {
		if num, ok2 := v.(float64); ok2 {
			if num < 0 {
				num = 0
			}
			if num > 1 {
				num = 1
			}
			s.cpuSmoothAlpha = num
		}
	}
	// Note: cpu smoothing alpha is not currently persisted separately; keep default
	return nil
}

// GetRuntimeSmoothing returns the currently active smoothing alpha values.
func (s *HostService) GetRuntimeSmoothing() (cpu, disk, net float64) {
	return s.cpuSmoothAlpha, s.diskSmoothAlpha, s.netSmoothAlpha
}

// EnsureHostConnected ensures there's an active libvirt connection for the
// given host ID. If no connection exists, it will attempt to read the host
// URI from the database and connect. This allows lazy connection on demand
// (e.g., when the UI first subscribes to stats) instead of connecting all
// hosts at startup.
// This respects the AutoReconnectDisabled flag and will not reconnect
// hosts that were manually disconnected by the user.
func (s *HostService) EnsureHostConnected(hostID string) error {
	log.Debugf("EnsureHostConnected: checking connection for host %s", hostID)
	if _, err := s.connector.GetConnection(hostID); err == nil {
		log.Debugf("EnsureHostConnected: host %s already connected", hostID)
		return nil // already connected
	}

	var host storage.Host
	if err := s.db.Where("id = ?", hostID).First(&host).Error; err != nil {
		return fmt.Errorf("could not find host %s in database: %w", hostID, err)
	}

	// Check if auto-reconnection is disabled for this host
	if host.AutoReconnectDisabled {
		log.Debugf("EnsureHostConnected: auto-reconnection disabled for host %s", hostID)
		return fmt.Errorf("host %s has auto-reconnection disabled", hostID)
	}

	log.Verbosef("EnsureHostConnected: connecting to host %s (uri=%s)", hostID, host.URI)
	// Mark host as connecting in DB so UI and API can reflect transient state
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": storage.HostTaskStateConnecting, "state": storage.HostStateDisconnected})
	if err := s.connector.AddHost(host); err != nil {
		log.Debugf("EnsureHostConnected: failed to connect to host %s: %v", hostID, err)
		// Set task_state to empty and mark as error
		s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateError})
		return fmt.Errorf("failed to connect to host %s: %w", hostID, err)
	}
	log.Verbosef("EnsureHostConnected: connection established for host %s", hostID)
	// Clear task state and mark connected
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateConnected})
	// Notify clients that the host is now connected
	s.broadcastHostConnectionChanged(hostID, true)
	s.broadcastHostsChanged()

	// Start VM state polling for this host
	s.startVMPolling(hostID)

	// Sync VMs for the newly connected host
	go s.SyncVMsForHost(hostID)

	return nil
}

// EnsureHostConnectedForced ensures there's an active libvirt connection for the
// given host ID, ignoring the AutoReconnectDisabled flag. This should be used
// for manual connect requests from the user.
func (s *HostService) EnsureHostConnectedForced(hostID string) error {
	log.Debugf("EnsureHostConnectedForced: forcing connection for host %s", hostID)
	if _, err := s.connector.GetConnection(hostID); err == nil {
		log.Debugf("EnsureHostConnectedForced: host %s already connected", hostID)
		return nil // already connected
	}

	var host storage.Host
	if err := s.db.Where("id = ?", hostID).First(&host).Error; err != nil {
		return fmt.Errorf("could not find host %s in database: %w", hostID, err)
	}

	log.Verbosef("EnsureHostConnectedForced: connecting to host %s (uri=%s)", hostID, host.URI)
	// Mark host as connecting in DB so UI and API can reflect transient state
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": storage.HostTaskStateConnecting, "state": storage.HostStateDisconnected})
	if err := s.connector.AddHost(host); err != nil {
		log.Debugf("EnsureHostConnectedForced: failed to connect to host %s: %v", hostID, err)
		// Set task_state to empty and mark as error
		s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateError})
		return fmt.Errorf("failed to connect to host %s: %w", hostID, err)
	}
	log.Verbosef("EnsureHostConnectedForced: connection established for host %s", hostID)
	// Clear task state, mark connected, and enable auto-reconnection
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{
		"task_state":              "",
		"state":                   storage.HostStateConnected,
		"auto_reconnect_disabled": false,
	})
	// Notify clients that the host is now connected
	s.broadcastHostConnectionChanged(hostID, true)
	s.broadcastHostsChanged()

	// Sync VMs for the newly connected host
	go s.SyncVMsForHost(hostID)

	return nil
}

// DisconnectHost disconnects an active libvirt connection for the host and
// updates DB state. It's safe to call even if no connection exists.
// userInitiated indicates if this disconnect was requested by the user
// (in which case auto-reconnection will be disabled).
func (s *HostService) DisconnectHost(hostID string, userInitiated bool) error {
	log.Debugf("DisconnectHost: disconnecting host %s (userInitiated=%v)", hostID, userInitiated)
	// If there's no connection, return nil
	if _, err := s.connector.GetConnection(hostID); err != nil {
		log.Debugf("DisconnectHost: no active connection for host %s", hostID)
		return nil
	}

	if err := s.connector.RemoveHost(hostID); err != nil {
		// Mark host state as error
		s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateError})
		return fmt.Errorf("failed to disconnect host %s: %w", hostID, err)
	}

	// Mark host as disconnected and set auto-reconnect flag
	updates := map[string]interface{}{
		"task_state": "",
		"state":      storage.HostStateDisconnected,
	}
	if userInitiated {
		updates["auto_reconnect_disabled"] = true
	}
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(updates)

	// Set all VMs for this host to unknown libvirt state since we can't observe them
	if err := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ?", hostID).Update("libvirt_state", storage.StateUnknown).Error; err != nil {
		log.Warnf("Failed to update VM libvirt states to unknown for host %s: %v", hostID, err)
	}

	// Stop all monitoring for this host
	s.monitor.StopHostMonitoring(hostID)
	s.hostMonitor.StopHostMonitoring(hostID)
	s.stopVMPolling(hostID)

	log.Infof("Host %s disconnected successfully (userInitiated=%v)", hostID, userInitiated)
	s.broadcastHostConnectionChanged(hostID, false)
	s.broadcastHostsChanged()
	return nil
}

// AddHost creates a new Host record and triggers a background connection attempt.
func (s *HostService) AddHost(host storage.Host) (*storage.Host, error) {
	if host.ID == "" {
		host.ID = uuid.New().String()
	}

	if err := s.db.Create(&host).Error; err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Notify clients that hosts changed
	s.broadcastHostsChanged()

	// Attempt to connect in background so API returns fast. EnsureHostConnected
	// will update DB task_state/state appropriately.
	go func(id string) {
		if err := s.EnsureHostConnected(id); err != nil {
			log.Verbosef("AddHost: background connect failed for host %s: %v", id, err)
		}
	}(host.ID)

	return &host, nil
}

func (s *HostService) broadcastHostsChanged() {
	s.hub.BroadcastMessage(ws.Message{Type: "hosts-changed"})
}

func (s *HostService) broadcastVMsChanged(hostID string) {
	s.hub.BroadcastMessage(ws.Message{
		Type:    "vms-changed",
		Payload: ws.MessagePayload{"hostId": hostID},
	})
}

func (s *HostService) broadcastDiscoveredVMsChanged(hostID string) {
	s.hub.BroadcastMessage(ws.Message{
		Type:    "discovered-vms-changed",
		Payload: ws.MessagePayload{"hostId": hostID},
	})
}

func (s *HostService) broadcastHostConnectionChanged(hostID string, connected bool) {
	s.hub.BroadcastMessage(ws.Message{
		Type:    "host-connection-changed",
		Payload: ws.MessagePayload{"hostId": hostID, "connected": connected},
	})
}

// mapLibvirtStateToVMState translates libvirt's integer state to our internal string state.
func mapLibvirtStateToVMState(state golibvirt.DomainState) storage.VMState {
	switch state {
	case golibvirt.DomainRunning:
		return storage.StateActive
	case golibvirt.DomainPaused:
		return storage.StatePaused
	case golibvirt.DomainPmsuspended:
		return storage.StateSuspended
	case golibvirt.DomainShutdown, golibvirt.DomainShutoff, golibvirt.DomainCrashed:
		return storage.StateStopped
	default:
		return storage.StateError
	}
}

// ensureAttachmentIndex tries to create an AttachmentIndex. If creation fails
// due to a UNIQUE constraint (possible because a soft-deleted row exists or
// a concurrent transaction inserted it), it will attempt to reconcile by
// inspecting live and soft-deleted rows and restoring/updating as needed.
func (s *HostService) ensureAttachmentIndex(tx *gorm.DB, alloc storage.AttachmentIndex) error {
	maxAttempts := 3
	// If we have an attachment id, check for existing or soft-deleted rows
	// before attempting to create. This prevents immediate UNIQUE constraint
	// failures in common cases (e.g., video attachments) where the
	// attachment row is created before the index is reconciled.
	if alloc.AttachmentID != 0 {
		var existing []storage.AttachmentIndex
		tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&existing)
		if len(existing) > 0 {
			if existing[0].VMUUID == alloc.VMUUID {
				return nil
			}
			return fmt.Errorf("attachment (id=%d) already indexed for VM %s (index id %d)", alloc.AttachmentID, existing[0].VMUUID, existing[0].ID)
		}

		var soft []storage.AttachmentIndex
		tx.Unscoped().Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&soft)
		if len(soft) > 0 {
			if soft[0].DeletedAt.Valid {
				if uerr := tx.Unscoped().Model(&soft[0]).Updates(map[string]interface{}{"vm_uuid": alloc.VMUUID, "device_id": alloc.DeviceID, "deleted_at": nil}).Error; uerr != nil {
					return fmt.Errorf("failed to restore soft-deleted attachment_index for attachment %v: %w", alloc.AttachmentID, uerr)
				}
				return nil
			}
			return fmt.Errorf("attachment_index exists for attachment %v but could not be used", alloc.AttachmentID)
		}
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := tx.Create(&alloc).Error; err == nil {
			return nil
		} else {
			// Only attempt reconciliation on UNIQUE constraint failures
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return err
			}

			// 1) Check live by attachment_id
			var byAttach storage.AttachmentIndex
			if r := tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).First(&byAttach); r.Error == nil {
				if byAttach.VMUUID == alloc.VMUUID {
					return nil
				}
				return fmt.Errorf("attachment (id=%d) already indexed for VM %s (index id %d)", alloc.AttachmentID, byAttach.VMUUID, byAttach.ID)
			} else if r.Error != nil && r.Error != gorm.ErrRecordNotFound {
				return r.Error
			}

			// 2) Restore soft-deleted by attachment_id
			var softAttach storage.AttachmentIndex
			if u := tx.Unscoped().Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).First(&softAttach); u.Error == nil {
				if softAttach.DeletedAt.Valid {
					if uerr := tx.Unscoped().Model(&softAttach).Updates(map[string]interface{}{"vm_uuid": alloc.VMUUID, "device_id": alloc.DeviceID, "deleted_at": nil}).Error; uerr != nil {
						return fmt.Errorf("failed to restore soft-deleted attachment_index for attachment %v: %w", alloc.AttachmentID, uerr)
					}
					return nil
				}
				return fmt.Errorf("attachment_index exists for attachment %v but could not be used", alloc.AttachmentID)
			} else if u.Error != nil && u.Error != gorm.ErrRecordNotFound {
				return u.Error
			}

			// 3) Check/restore by device_id when present
			if alloc.DeviceID != nil {
				var byDevice storage.AttachmentIndex
				if r := tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).First(&byDevice); r.Error == nil {
					if byDevice.AttachmentID == alloc.AttachmentID && byDevice.VMUUID == alloc.VMUUID {
						return nil
					}
					return fmt.Errorf("device (type=%s id=%v) already allocated to VM %s (attachment_index id %d)", alloc.DeviceType, alloc.DeviceID, byDevice.VMUUID, byDevice.ID)
				} else if r.Error != nil && r.Error != gorm.ErrRecordNotFound {
					return r.Error
				}

				var softDevice storage.AttachmentIndex
				if u := tx.Unscoped().Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).First(&softDevice); u.Error == nil {
					if softDevice.DeletedAt.Valid {
						if uerr := tx.Unscoped().Model(&softDevice).Updates(map[string]interface{}{"vm_uuid": alloc.VMUUID, "attachment_id": alloc.AttachmentID, "deleted_at": nil}).Error; uerr != nil {
							return fmt.Errorf("failed to restore soft-deleted attachment_index for device %v: %w", alloc.DeviceID, uerr)
						}
						return nil
					}
					return fmt.Errorf("unexpected state: attachment_index exists but insert failed for device %v", alloc.DeviceID)
				} else if u.Error != nil && u.Error != gorm.ErrRecordNotFound {
					return u.Error
				}
			}
		}

		// brief retry delay for transient races
		if attempt < maxAttempts {
			time.Sleep(25 * time.Millisecond)
			continue
		}

		// final attempt to surface the error
		return tx.Create(&alloc).Error
	}

	return fmt.Errorf("ensureAttachmentIndex: exhausted retries")
}

// RemoveHost removes a host from the system: disconnects the connector and
// removes associated VMs and attachment indices transactionally.
func (s *HostService) RemoveHost(hostID string) error {
	// Mark as disconnecting in DB
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": storage.HostTaskStateDisconnecting})

	if err := s.connector.RemoveHost(hostID); err != nil {
		log.Verbosef("RemoveHost: failed to disconnect from host %s during removal: %v", hostID, err)
		s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateError})
	}

	// Remove VMs and their attachment indices transactionally
	tx := s.db.Begin()
	var vms []storage.VirtualMachine
	if err := tx.Where("host_id = ?", hostID).Find(&vms).Error; err != nil {
		tx.Rollback()
		log.Verbosef("Warning: failed to query VMs for host %s: %v", hostID, err)
		return err
	}

	for _, vm := range vms {
		if err := tx.Where("vm_uuid = ?", vm.UUID).Delete(&storage.AttachmentIndex{}).Error; err != nil {
			tx.Rollback()
			log.Verbosef("Warning: failed to delete attachment indices for VM %s: %v", vm.Name, err)
			return err
		}
		if err := tx.Delete(&vm).Error; err != nil {
			tx.Rollback()
			log.Verbosef("Warning: failed to delete VM %s: %v", vm.Name, err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Verbosef("Warning: failed to commit host removal transaction for %s: %v", hostID, err)
		return err
	}

	// Finally remove the host row and clear task state
	if err := s.db.Where("id = ?", hostID).Delete(&storage.Host{}).Error; err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}

	// Broadcast host removal and disconnected status
	s.broadcastHostConnectionChanged(hostID, false)
	s.broadcastHostsChanged()
	return nil
}

func (s *HostService) ConnectToAllHosts() {
	hosts, err := s.GetAllHosts()
	if err != nil {
		log.Verbosef("Error retrieving hosts from database on startup: %v", err)
		return
	}

	for _, host := range hosts {
		log.Verbosef("Attempting to connect to stored host: %s", host.ID)
		if err := s.connector.AddHost(host); err != nil {
			log.Verbosef("Failed to connect to host %s (%s) on startup: %v", host.ID, host.URI, err)
		} else {
			go s.SyncVMsForHost(host.ID)
		}
	}
}

// GetAllHosts returns all stored hosts.
func (s *HostService) GetAllHosts() ([]storage.Host, error) {
	var hosts []storage.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// AutoConnectHosts connects to all hosts that were previously connected
// (i.e., have state = CONNECTED in the database). This should be called
// on backend startup to restore connection state.
func (s *HostService) AutoConnectHosts() error {
	var hosts []storage.Host
	if err := s.db.Where("state = ?", storage.HostStateConnected).Find(&hosts).Error; err != nil {
		return fmt.Errorf("failed to query connected hosts: %w", err)
	}

	for _, host := range hosts {
		log.Infof("Auto-connecting to previously connected host %s (%s)", host.ID, host.URI)
		if err := s.EnsureHostConnected(host.ID); err != nil {
			log.Errorf("Failed to auto-connect to host %s: %v", host.ID, err)
			// Continue with other hosts even if one fails
		}
	}
	return nil
}

// GetHostInfo returns host information from the connector, ensuring the host
// is connected first.
func (s *HostService) GetHostInfo(hostID string) (*libvirt.HostInfo, error) {
	if err := s.EnsureHostConnected(hostID); err != nil {
		return nil, err
	}
	return s.connector.GetHostInfo(hostID)
}

// GetHostStats returns real-time statistics for a host.
func (s *HostService) GetHostStats(hostID string) (*libvirt.HostStats, error) {
	if err := s.EnsureHostConnected(hostID); err != nil {
		return nil, err
	}
	return s.connector.GetHostStats(hostID)
}

// GetPortsForHostFromDB returns ports scoped to the given host (the host's
// port pool / unattached ports).
func (s *HostService) GetPortsForHostFromDB(hostID string) ([]storage.Port, error) {
	var ports []storage.Port
	if err := s.db.Where("host_id = ?", hostID).Find(&ports).Error; err != nil {
		return nil, err
	}
	return ports, nil
}

// GetPortAttachmentsForVM returns port attachments for a VM UUID with the
// underlying Port and (optional) Network preloaded into the view.
func (s *HostService) GetPortAttachmentsForVM(vmUUID string) ([]PortAttachmentView, error) {
	var atts []storage.PortAttachment
	if err := s.db.Preload("Port").Where("vm_uuid = ?", vmUUID).Find(&atts).Error; err != nil {
		return nil, err
	}

	var out []PortAttachmentView
	for _, a := range atts {
		var binding storage.PortBinding
		var network *storage.Network
		if err := s.db.Preload("Network").Where("port_id = ?", a.PortID).First(&binding).Error; err == nil {
			network = &binding.Network
		}

		modelType := a.ModelName
		if modelType == "" {
			modelType = a.Port.ModelName
		}
		mac := a.MACAddress
		if mac == "" {
			mac = a.Port.MACAddress
		}

		out = append(out, PortAttachmentView{
			ID:         a.ID,
			VMUUID:     a.VMUUID,
			DeviceName: a.DeviceName,
			MACAddress: mac,
			ModelName:  modelType,
			HostID:     a.HostID,
			Ordinal:    a.Ordinal,
			Metadata:   a.Metadata,
			Port:       a.Port,
			Network:    network,
		})
	}
	return out, nil
}

// --- VM Management ---
func (s *HostService) GetVMsForHostFromDB(hostID string) ([]VMView, error) {
	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return nil, fmt.Errorf("could not get DB VM records for host %s: %w", hostID, err)
	}

	// Check if host is connected to populate live data
	// Temporarily disabled for performance
	// hostConnected := false
	// if _, err := s.connector.GetConnection(hostID); err == nil {
	//     hostConnected = true
	// }

	var vmViews = make([]VMView, 0)
	for _, dbVM := range dbVMs {
		var graphics libvirt.GraphicsInfo // Default to false
		var liveData libvirt.VMInfo

		if dbVM.State == storage.StateActive {
			var console storage.Console
			err := s.db.Where("vm_uuid = ?", dbVM.UUID).First(&console).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				log.Verbosef("Error querying console for running VM %s: %v", dbVM.Name, err)
			} else if err == nil {
				graphics.VNC = strings.ToLower(console.Type) == "vnc"
				graphics.SPICE = strings.ToLower(console.Type) == "spice"
			}
		}

		// Populate live data if host is connected
		// Temporarily disabled for performance - uptime can be fetched separately if needed
		/*
			if hostConnected {
				if vmInfo, err := s.connector.GetDomainInfo(hostID, dbVM.Name); err == nil {
					liveData = libvirt.VMInfo{
						Uptime: vmInfo.Uptime,
					}
				}
			}
		*/

		// Calculate disk size and get primary network interface
		diskSizeGB := s.calculateVMDiskSize(dbVM.UUID)
		networkInterface := s.getPrimaryNetworkInterface(dbVM.UUID)

		vmViews = append(vmViews, VMView{
			Name:            dbVM.Name,
			UUID:            dbVM.UUID,
			DomainUUID:      dbVM.DomainUUID,
			Description:     dbVM.Description,
			VCPUCount:       dbVM.VCPUCount,
			MemoryBytes:     dbVM.MemoryBytes,
			IsTemplate:      dbVM.IsTemplate,
			CPUModel:        dbVM.CPUModel,
			CPUTopologyJSON: dbVM.CPUTopologyJSON,
			OSType:          dbVM.OSType,
			State:           dbVM.State,
			LibvirtState:    dbVM.LibvirtState,
			TaskState:       dbVM.TaskState,
			Graphics:        graphics,
			SyncStatus:      dbVM.SyncStatus,
			DriftDetails:    dbVM.DriftDetails,
			NeedsRebuild:    dbVM.NeedsRebuild,
			// Timestamps
			CreatedAt: dbVM.CreatedAt,
			UpdatedAt: dbVM.UpdatedAt,
			// Live data
			Uptime: liveData.Uptime,
			// Computed fields
			DiskSizeGB:       diskSizeGB,
			NetworkInterface: networkInterface,
		})
	}
	return vmViews, nil
}

// calculateVMDiskSize calculates the total disk size in GB for a VM based on disk attachments
func (s *HostService) calculateVMDiskSize(vmUUID string) float64 {
	var totalSizeBytes uint64 = 0

	// Get disk attachments (current schema)
	var diskAttachments []storage.DiskAttachment
	if err := s.db.Preload("Disk").Where("vm_uuid = ?", vmUUID).Find(&diskAttachments).Error; err != nil {
		log.Verbosef("Failed to get disk attachments for VM %s: %v", vmUUID, err)
		return 0
	}

	for _, attachment := range diskAttachments {
		totalSizeBytes += attachment.Disk.CapacityBytes
	}

	// Convert bytes to GB
	return float64(totalSizeBytes) / (1024 * 1024 * 1024)
}

// getPrimaryNetworkInterface gets the primary network interface/bridge name for a VM
func (s *HostService) getPrimaryNetworkInterface(vmUUID string) string {
	var portAttachments []storage.PortAttachment
	if err := s.db.Preload("Port").Where("vm_uuid = ?", vmUUID).Find(&portAttachments).Error; err != nil {
		log.Verbosef("Failed to get port attachments for VM %s: %v", vmUUID, err)
		return ""
	}

	if len(portAttachments) == 0 {
		return ""
	}

	// Get the network for the first port attachment
	var portBinding storage.PortBinding
	if err := s.db.Preload("Network").Where("port_id = ?", portAttachments[0].PortID).First(&portBinding).Error; err != nil {
		log.Verbosef("Failed to get port binding for VM %s: %v", vmUUID, err)
		return ""
	}

	if portBinding.Network.BridgeName != "" {
		return portBinding.Network.BridgeName
	}

	return portBinding.Network.Name
}

func (s *HostService) getVMHardwareFromDB(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	var vm storage.VirtualMachine
	if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&vm).Error; err != nil {
		return nil, fmt.Errorf("could not find VM %s in database: %w", vmName, err)
	}

	var hardware libvirt.HardwareInfo

	// Retrieve and populate disks
	var diskAttachments []storage.DiskAttachment
	s.db.Preload("Disk").Where("vm_uuid = ?", vm.UUID).Find(&diskAttachments)
	for _, da := range diskAttachments {
		path := da.Disk.Path
		if da.Disk.VolumeID != nil {
			// If it references a volume, get the volume name
			var vol storage.Volume
			if err := s.db.First(&vol, da.Disk.VolumeID).Error; err == nil {
				path = vol.Name
			}
		}
		var driver struct {
			Name string `xml:"name,attr" json:"driver_name"`
			Type string `xml:"type,attr" json:"type"`
		}
		if da.Disk.DriverJSON != "" {
			json.Unmarshal([]byte(da.Disk.DriverJSON), &driver)
		} else {
			driver.Type = da.Disk.Format
		}
		hardware.Disks = append(hardware.Disks, libvirt.DiskInfo{
			Device: da.DeviceName,
			Source: struct {
				File string `xml:"file,attr" json:"file"`
				Dev  string `xml:"dev,attr" json:"dev"`
			}{
				File: path,
			},
			Path:     path,
			Name:     da.Disk.Name,
			ReadOnly: da.ReadOnly,
			Target: struct {
				Dev string `xml:"dev,attr" json:"dev"`
				Bus string `xml:"bus,attr" json:"bus"`
			}{
				Dev: da.DeviceName,
				Bus: da.BusType,
			},
			Driver: driver,
		})
	}

	// Retrieve and populate networks using PortAttachment records.
	var attachments []storage.PortAttachment
	if err := s.db.Preload("Port").Where("vm_uuid = ?", vm.UUID).Find(&attachments).Error; err == nil {
		for _, a := range attachments {
			var binding storage.PortBinding
			if err := s.db.Preload("Network").Where("port_id = ?", a.PortID).First(&binding).Error; err == nil {
				modelType := a.ModelName
				if modelType == "" {
					modelType = a.Port.ModelName
				}
				// DeviceName is attachment-scoped. If attachment doesn't have it, leave empty.
				devName := a.DeviceName
				// MAC can fall back to the Port resource canonical MAC
				mac := a.MACAddress
				if mac == "" {
					mac = a.Port.MACAddress
				}

				hardware.Networks = append(hardware.Networks, libvirt.NetworkInfo{
					Mac: struct {
						Address string `xml:"address,attr" json:"address"`
					}{
						Address: mac,
					},
					Source: struct {
						Bridge    string `xml:"bridge,attr" json:"bridge"`
						Network   string `xml:"network,attr" json:"network"`
						PortGroup string `xml:"portgroup,attr" json:"portgroup"`
					}{
						Bridge: binding.Network.BridgeName,
					},
					Model: struct {
						Type string `xml:"type,attr" json:"type"`
					}{
						Type: modelType,
					},
					Target: struct {
						Dev string `xml:"dev,attr" json:"dev"`
					}{
						Dev: devName,
					},
				})
			}
		}
	}

	// Retrieve and populate videos
	var videoAttachments []storage.VideoAttachment
	s.db.Preload("VideoModel").Where("vm_uuid = ?", vm.UUID).Find(&videoAttachments)
	for _, va := range videoAttachments {
		hardware.Videos = append(hardware.Videos, libvirt.VideoInfo{
			Model: struct {
				Type  string `xml:"type,attr" json:"type"`
				VRAM  int    `xml:"vram,attr,omitempty" json:"vram,omitempty"`
				Heads int    `xml:"heads,attr,omitempty" json:"heads,omitempty"`
			}{
				Type:  va.VideoModel.ModelName,
				VRAM:  int(va.VideoModel.VRAM),
				Heads: va.VideoModel.Heads,
			},
		})
	}

	var bootConfig storage.BootConfig
	if err := s.db.Where("vm_uuid = ?", vm.UUID).First(&bootConfig).Error; err == nil {
		var bootDevices []libvirt.BootEntry
		if err := json.Unmarshal([]byte(bootConfig.BootOrderJSON), &bootDevices); err == nil {
			hardware.Boot = bootDevices
		}
	}

	return &hardware, nil
}
func (s *HostService) GetVMHardwareAndDetectDrift(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	// Check if host is connected
	if _, err := s.connector.GetConnection(hostID); err != nil {
		return nil, fmt.Errorf("host %s is not connected", hostID)
	}

	if changed, syncErr := s.detectDriftOrIngestVM(hostID, vmName, false); syncErr != nil {
		log.Verbosef("Error during hardware sync for %s: %v", vmName, syncErr)
	} else if changed {
		s.broadcastVMsChanged(hostID)
	}

	return s.getVMHardwareFromDB(hostID, vmName)
}

func (s *HostService) SyncVMsForHost(hostID string) {
	// Prevent syncs too frequently (within 30 seconds)
	now := time.Now()
	if last, ok := s.lastSync.Load(hostID); ok {
		if now.Sub(last.(time.Time)) < 30*time.Second {
			log.Debugf("Skipping sync for host %s, last sync was %v ago", hostID, now.Sub(last.(time.Time)))
			return
		}
	}
	s.lastSync.Store(hostID, now)

	// Prevent concurrent syncs for the same host
	mu, _ := s.syncMutex.LoadOrStore(hostID, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

	_, err := s.syncHostVMs(hostID)
	if err != nil {
		log.Verbosef("Error during background VM sync for host %s: %v", hostID, err)
		return
	}

	_, err = s.syncHostStoragePools(hostID)
	if err != nil {
		log.Verbosef("Error during background storage pool sync for host %s: %v", hostID, err)
		return
	}
	// Note: syncHostVMs already broadcasts vms-changed and discovered-vms-changed if changed
}

// startVMPolling starts background polling of VM states for a connected host
func (s *HostService) startVMPolling(hostID string) {
	// Stop any existing poller for this host
	if stopChan, exists := s.vmPollers.Load(hostID); exists {
		close(stopChan.(chan struct{}))
	}

	// Create a stop channel for this poller
	stopChan := make(chan struct{})
	s.vmPollers.Store(hostID, stopChan)

	go func() {
		ticker := time.NewTicker(30 * time.Second) // Poll every 30 seconds
		defer ticker.Stop()

		log.Debugf("Started VM state polling for host %s", hostID)

		// Do an initial poll immediately
		if err := s.pollVMStates(hostID); err != nil {
			log.Verbosef("Error in initial VM state poll for host %s: %v", hostID, err)
		}

		for {
			select {
			case <-stopChan:
				log.Debugf("Stopped VM state polling for host %s", hostID)
				return
			case <-ticker.C:
				// Check if host is still connected
				if _, err := s.connector.GetConnection(hostID); err != nil {
					log.Debugf("Host %s no longer connected, stopping VM polling", hostID)
					return
				}

				// Poll VM states
				if err := s.pollVMStates(hostID); err != nil {
					log.Verbosef("Error polling VM states for host %s: %v", hostID, err)
				}
			}
		}
	}()
}

// stopVMPolling stops background polling for a host
func (s *HostService) stopVMPolling(hostID string) {
	if stopChan, exists := s.vmPollers.LoadAndDelete(hostID); exists {
		close(stopChan.(chan struct{}))
	}
}

// pollVMStates polls current VM states from libvirt and updates libvirtState in DB
func (s *HostService) pollVMStates(hostID string) error {
	// Get all VMs for this host from DB
	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return fmt.Errorf("failed to get VMs for host %s: %w", hostID, err)
	}

	var changed bool
	for _, dbVM := range dbVMs {
		// Get current state from libvirt
		vmInfo, err := s.connector.GetDomainInfo(hostID, dbVM.Name)
		if err != nil {
			// VM might not exist anymore, set libvirtState to UNKNOWN
			if dbVM.LibvirtState != storage.StateUnknown {
				s.db.Model(&dbVM).Update("libvirt_state", storage.StateUnknown)
				changed = true
			}
			continue
		}

		// Update libvirtState if it changed
		newLibvirtState := mapLibvirtStateToVMState(vmInfo.State)
		if dbVM.LibvirtState != newLibvirtState {
			if err := s.db.Model(&dbVM).Update("libvirt_state", newLibvirtState).Error; err != nil {
				log.Verbosef("Failed to update libvirtState for VM %s: %v", dbVM.Name, err)
				continue
			}
			changed = true
			log.Debugf("Updated libvirtState for VM %s on host %s: %s -> %s", dbVM.Name, hostID, dbVM.LibvirtState, newLibvirtState)
		}
	}

	if changed {
		s.broadcastVMsChanged(hostID)
	}

	return nil
}

// currently present in our database (i.e., discovered but not managed).
func (s *HostService) ListDiscoveredVMs(hostID string) ([]libvirt.VMInfo, error) {
	// Serve discovered VMs from the DB for fast, deterministic UI responses.
	// Background: if a fresh libvirt sync is desired, the API handler can
	// call SyncVMsForHost or the client can call with a ?refresh=true flag.
	var discs []storage.DiscoveredVM
	if err := s.db.Where("host_id = ? AND imported = 0", hostID).Find(&discs).Error; err != nil {
		return nil, err
	}

	var out []libvirt.VMInfo
	for _, d := range discs {
		out = append(out, libvirt.VMInfo{UUID: d.DomainUUID, Name: d.Name})
	}

	// Trigger a background sync to refresh discovered-vms from libvirt.
	// This will update the discovered_vms table and broadcast a vms-changed event
	// when new rows are ingested/removed.
	go func(id string) {
		// Prevent concurrent syncs for the same host
		mu, _ := s.syncMutex.LoadOrStore(id, &sync.Mutex{})
		mu.(*sync.Mutex).Lock()
		defer mu.(*sync.Mutex).Unlock()

		_, err := s.syncHostVMs(id)
		if err != nil {
			log.Verbosef("background syncHostVMs failed for host %s: %v", id, err)
		}
	}(hostID)

	return out, nil
}

// ImportVM imports a single discovered VM into the database by name.
func (s *HostService) ImportVM(hostID, vmName string) error {
	log.Infof("ImportVM started - hostID: %s, vmName: %s", hostID, vmName)

	// Prevent concurrent syncs for the same host
	mu, _ := s.syncMutex.LoadOrStore(hostID, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

	// Get domain info to obtain UUID for marking imported
	log.Verbosef("Getting domain info for VM %s on host %s", vmName, hostID)
	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		log.Errorf("Failed to get domain info for VM %s on host %s: %v", vmName, hostID, err)
		return err
	}
	log.Verbosef("Successfully got domain info for VM %s: UUID=%s, State=%v", vmName, vmInfo.UUID, vmInfo.State)

	log.Verbosef("Starting VM ingestion for %s from libvirt", vmName)
	changed, err := s.ingestVMFromLibvirt(hostID, vmName)
	if err != nil {
		log.Errorf("Failed to ingest VM %s from libvirt: %v", vmName, err)
		return err
	}
	log.Verbosef("VM ingestion completed for %s, changed: %v", vmName, changed)

	if changed {
		// Mark as imported in discovered_vms
		log.Verbosef("Marking discovered VM %s (UUID: %s) as imported", vmName, vmInfo.UUID)
		if err := storage.MarkDiscoveredVMImported(s.db, hostID, vmInfo.UUID); err != nil {
			log.Verbosef("Warning: failed to mark discovered VM %s as imported: %v", vmName, err)
		}
		// Notify clients that the managed VM list changed for this host
		s.broadcastVMsChanged(hostID)
		s.broadcastDiscoveredVMsChanged(hostID)
	}
	return nil
}

// ImportAllVMs imports all discovered VMs on the host.
func (s *HostService) ImportAllVMs(hostID string) error {
	// Get discovered VMs from database instead of all libvirt domains
	discoveredVMs, err := storage.ListDiscoveredVMsByHost(s.db, hostID)
	if err != nil {
		return fmt.Errorf("failed to list discovered VMs: %w", err)
	}
	anyImported := false
	for _, discVM := range discoveredVMs {
		// Acquire mutex for each individual import to reduce contention
		mu, _ := s.syncMutex.LoadOrStore(hostID, &sync.Mutex{})
		mu.(*sync.Mutex).Lock()
		changed, ierr := s.ingestVMFromLibvirt(hostID, discVM.Name)
		mu.(*sync.Mutex).Unlock()

		if ierr != nil {
			log.Verbosef("ImportAllVMs: failed to import VM %s on host %s: %v", discVM.Name, hostID, ierr)
		} else if changed {
			anyImported = true
			// Mark as imported in discovered_vms
			if err := storage.MarkDiscoveredVMImported(s.db, hostID, discVM.DomainUUID); err != nil {
				log.Verbosef("Warning: failed to mark discovered VM %s as imported: %v", discVM.Name, err)
			}
		}
	}
	if anyImported {
		s.broadcastVMsChanged(hostID)
		s.broadcastDiscoveredVMsChanged(hostID)
	}
	return nil
}

// ImportSelectedVMs imports a list of discovered VMs by their domain UUIDs.
func (s *HostService) ImportSelectedVMs(hostID string, domainUUIDs []string) error {
	if len(domainUUIDs) == 0 {
		return nil
	}

	// Get discovered VMs by UUIDs
	var discoveredVMs []storage.DiscoveredVM
	if err := s.db.Where("host_id = ? AND domain_uuid IN ? AND imported = 0", hostID, domainUUIDs).Find(&discoveredVMs).Error; err != nil {
		return fmt.Errorf("failed to get selected discovered VMs: %w", err)
	}

	anyImported := false
	importedUUIDs := make([]string, 0, len(discoveredVMs))

	for _, discVM := range discoveredVMs {
		// Acquire mutex for each individual import to reduce contention
		mu, _ := s.syncMutex.LoadOrStore(hostID, &sync.Mutex{})
		mu.(*sync.Mutex).Lock()
		changed, ierr := s.ingestVMFromLibvirt(hostID, discVM.Name)
		mu.(*sync.Mutex).Unlock()

		if ierr != nil {
			log.Verbosef("ImportSelectedVMs: failed to import VM %s on host %s: %v", discVM.Name, hostID, ierr)
		} else if changed {
			anyImported = true
			importedUUIDs = append(importedUUIDs, discVM.DomainUUID)
		}
	}

	// Mark imported VMs in bulk
	if len(importedUUIDs) > 0 {
		if err := storage.BulkMarkDiscoveredVMsImported(s.db, hostID, importedUUIDs); err != nil {
			log.Verbosef("Warning: failed to mark selected discovered VMs as imported: %v", err)
		}
	}

	if anyImported {
		s.broadcastVMsChanged(hostID)
		s.broadcastDiscoveredVMsChanged(hostID)
	}
	return nil
}

// DeleteSelectedDiscoveredVMs removes discovered VMs from the database by their domain UUIDs.
func (s *HostService) DeleteSelectedDiscoveredVMs(hostID string, domainUUIDs []string) error {
	if len(domainUUIDs) == 0 {
		return nil
	}

	if err := storage.BulkDeleteDiscoveredVMs(s.db, hostID, domainUUIDs); err != nil {
		return fmt.Errorf("failed to delete selected discovered VMs: %w", err)
	}

	s.broadcastDiscoveredVMsChanged(hostID)
	return nil
}

// UpdateVMState updates the intended state of a VM in the database
func (s *HostService) UpdateVMState(hostID, vmName, state string) error {
	// Validate the state
	var validState storage.VMState
	switch state {
	case "INITIALIZED":
		validState = storage.StateInitialized
	case "ACTIVE", "RUNNING":
		validState = storage.StateActive
	case "PAUSED":
		validState = storage.StatePaused
	case "SUSPENDED":
		validState = storage.StateSuspended
	case "STOPPED":
		validState = storage.StateStopped
	case "ERROR":
		validState = storage.StateError
	case "UNKNOWN":
		validState = storage.StateUnknown
	default:
		return fmt.Errorf("invalid state: %s", state)
	}

	// Update the VM state in database
	result := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("state", validState)
	if result.Error != nil {
		return fmt.Errorf("failed to update VM state: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("VM %s not found on host %s", vmName, hostID)
	}

	// Broadcast the change
	s.broadcastVMsChanged(hostID)
	return nil
}

// ingestVMFromLibvirt performs the actual DB creation/restore for a VM found
// in libvirt. It encapsulates the previous ingestion logic that used to live
// in detectDriftOrIngestVM.
func (s *HostService) ingestVMFromLibvirt(hostID, vmName string) (bool, error) {
	log.Verbosef("ingestVMFromLibvirt started - hostID: %s, vmName: %s", hostID, vmName)

	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		log.Errorf("Failed to get domain info in ingestVMFromLibvirt for %s: %v", vmName, err)
		return false, err
	}
	log.Verbosef("Got domain info in ingestVMFromLibvirt: %s (UUID: %s)", vmName, vmInfo.UUID)

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic during VM ingestion for %s: %v", vmName, r)
			tx.Rollback()
		}
	}()

	// Attempt to restore soft-deleted VM with same domain_uuid
	log.Verbosef("Checking for soft-deleted VMs with UUID %s", vmInfo.UUID)
	var softVMs []storage.VirtualMachine
	tx.Unscoped().Where("domain_uuid = ?", vmInfo.UUID).Limit(1).Find(&softVMs)
	if len(softVMs) > 0 {
		log.Verbosef("Found soft-deleted VM to restore: %s", softVMs[0].Name)
		softVM := softVMs[0]
		if softVM.DeletedAt.Valid {
			softVM.HostID = hostID
			softVM.Name = vmInfo.Name
			softVM.DomainUUID = vmInfo.UUID
			softVM.State = mapLibvirtStateToVMState(vmInfo.State)
			softVM.VCPUCount = vmInfo.Vcpu
			softVM.MemoryBytes = vmInfo.MaxMem * 1024
			softVM.SyncStatus = storage.StatusSynced
			if uerr := tx.Unscoped().Model(&softVM).Updates(map[string]interface{}{
				"host_id":       softVM.HostID,
				"name":          softVM.Name,
				"domain_uuid":   softVM.DomainUUID,
				"state":         softVM.State,
				"libvirt_state": softVM.State,
				"v_cpu_count":   softVM.VCPUCount,
				"memory_bytes":  softVM.MemoryBytes,
				"sync_status":   softVM.SyncStatus,
				"deleted_at":    nil,
			}).Error; uerr != nil {
				tx.Rollback()
				return false, fmt.Errorf("failed to restore soft-deleted VM row during ingestion: %w", uerr)
			}
			// Also ingest hardware for the restored VM
			if hw, hwErr := s.connector.GetDomainHardware(hostID, vmName); hwErr == nil {
				if _, err := s.syncVMHardware(tx, softVM.UUID, hostID, hw, &vmInfo.Graphics); err != nil {
					tx.Rollback()
					return false, fmt.Errorf("failed to sync hardware for restored VM during ingestion: %w", err)
				}

				// Update hardware details on the restored VM record
				updates := make(map[string]interface{})

				// Update OS type
				if hw.OSType != "" {
					updates["os_type"] = hw.OSType
				}

				// Update CPU model from hardware info
				if hw.CPUInfo != nil && hw.CPUInfo.Model != "" {
					updates["cpu_model"] = hw.CPUInfo.Model
				} else if hw.CPUInfo != nil && hw.CPUInfo.Mode != "" {
					// If no specific model, use the mode (e.g., "host-passthrough")
					updates["cpu_model"] = hw.CPUInfo.Mode
				}

				// Apply updates if any
				if len(updates) > 0 {
					if err := tx.Model(&softVM).Updates(updates).Error; err != nil {
						log.Verbosef("Warning: failed to update hardware details for restored VM %s: %v", vmName, err)
					}
				}
			}
			if cerr := tx.Commit().Error; cerr != nil {
				return false, cerr
			}
			return true, nil
		}
	}

	// Check conflict where domain_uuid exists on different host
	log.Verbosef("Checking for domain UUID conflicts for %s", vmInfo.UUID)
	var conflicting []storage.VirtualMachine
	tx.Where("domain_uuid = ? AND host_id != ?", vmInfo.UUID, hostID).Limit(1).Find(&conflicting)
	if len(conflicting) > 0 {
		log.Errorf("Domain UUID conflict for %s: already exists on host %s", vmInfo.UUID, conflicting[0].HostID)
		tx.Rollback()
		return false, fmt.Errorf("VM with domain UUID %s already exists on different host %s", vmInfo.UUID, conflicting[0].HostID)
	}

	log.Verbosef("Creating new VM record for %s (UUID: %s)", vmName, vmInfo.UUID)
	newVM := storage.VirtualMachine{
		HostID:       hostID,
		Name:         vmInfo.Name,
		DomainUUID:   vmInfo.UUID,
		State:        mapLibvirtStateToVMState(vmInfo.State),
		LibvirtState: mapLibvirtStateToVMState(vmInfo.State),
		VCPUCount:    vmInfo.Vcpu,
		MemoryBytes:  vmInfo.MaxMem * 1024,
		SyncStatus:   storage.StatusSynced,
		Source:       "managed",
	}
	var existingVMs []storage.VirtualMachine
	tx.Where("domain_uuid = ?", vmInfo.UUID).Limit(1).Find(&existingVMs)
	if len(existingVMs) == 0 {
		newVM.UUID = vmInfo.UUID
	} else {
		newVM.UUID = uuid.New().String()
	}

	if ierr := tx.Create(&newVM).Error; ierr != nil {
		if strings.Contains(ierr.Error(), "UNIQUE constraint failed: virtual_machines.host_id, virtual_machines.name") {
			origName := newVM.Name
			newVM.Name = fmt.Sprintf("%s-%s", origName, uuid.New().String()[:8])
			if ierr2 := tx.Create(&newVM).Error; ierr2 != nil {
				tx.Rollback()
				return false, ierr2
			}
		} else {
			tx.Rollback()
			return false, ierr
		}
	}

	// Ingest hardware for the new VM
	if hw, hwErr := s.connector.GetDomainHardware(hostID, vmName); hwErr == nil {
		if _, err := s.syncVMHardware(tx, newVM.UUID, hostID, hw, &vmInfo.Graphics); err != nil {
			tx.Rollback()
			return false, fmt.Errorf("failed to sync hardware during ingestion: %w", err)
		}

		// Update VM record with hardware details
		updates := make(map[string]interface{})

		// Update OS type
		if hw.OSType != "" {
			newVM.OSType = hw.OSType
			updates["os_type"] = hw.OSType
		}

		// Update CPU model from hardware info
		if hw.CPUInfo != nil && hw.CPUInfo.Model != "" {
			newVM.CPUModel = hw.CPUInfo.Model
			updates["cpu_model"] = hw.CPUInfo.Model
		} else if hw.CPUInfo != nil && hw.CPUInfo.Mode != "" {
			// If no specific model, use the mode (e.g., "host-passthrough")
			newVM.CPUModel = hw.CPUInfo.Mode
			updates["cpu_model"] = hw.CPUInfo.Mode
		}

		// Apply updates if any
		if len(updates) > 0 {
			if err := tx.Model(&newVM).Updates(updates).Error; err != nil {
				log.Verbosef("Warning: failed to update hardware details for VM %s: %v", vmName, err)
			}
		}
	}

	if cerr := tx.Commit().Error; cerr != nil {
		return false, cerr
	}
	return true, nil
}

func (s *HostService) detectDriftOrIngestVM(hostID, vmName string, isInitialSync bool) (bool, error) {
	_ = isInitialSync // Parameter reserved for future use to distinguish initial sync behavior
	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		// If we can't get info from libvirt, first check whether the connector
		// actually has an active connection to the host. If the host is not
		// connected, treat this as a transient error and do NOT prune the VM.
		if _, connErr := s.connector.GetConnection(hostID); connErr != nil {
			// Host not connected — skip pruning and report the underlying error.
			return false, fmt.Errorf("could not fetch info for VM %s on host %s: %w", vmName, hostID, err)
		}

		// Host is connected but the domain lookup failed — this likely means
		// the VM truly no longer exists in libvirt, so prune stale DB entries.
		var dbVM storage.VirtualMachine
		if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&dbVM).Error; err == nil {
			// Don't prune VMs that were created very recently (within last 5 minutes)
			// to avoid pruning VMs that were just imported but may not be immediately
			// visible to libvirt due to timing issues
			if time.Since(dbVM.CreatedAt) < 5*time.Minute {
				log.Verbosef("Skipping pruning recently created VM %s (created %v ago)", vmName, time.Since(dbVM.CreatedAt))
				return false, fmt.Errorf("could not fetch info for VM %s on host %s: %w", vmName, hostID, err)
			}
			log.Verbosef("Pruning VM %s from database as it's no longer in libvirt.", vmName)
			tx := s.db.Begin()
			if err := tx.Where("vm_uuid = ?", dbVM.UUID).Delete(&storage.AttachmentIndex{}).Error; err != nil {
				tx.Rollback()
				log.Verbosef("Warning: failed to delete attachment indices for VM %s: %v", dbVM.Name, err)
				return false, err
			}
			if err := tx.Delete(&dbVM).Error; err != nil {
				tx.Rollback()
				log.Verbosef("Warning: failed to prune old VM %s: %v", dbVM.Name, err)
				return false, err
			}
			if err := tx.Commit().Error; err != nil {
				log.Verbosef("Warning: failed to commit prune transaction for VM %s: %v", dbVM.Name, err)
				return false, err
			}
			return true, nil // A change occurred (deletion)
		}
		return false, fmt.Errorf("could not fetch info for VM %s on host %s: %w", vmName, hostID, err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingVM storage.VirtualMachine
	var existingVMs []storage.VirtualMachine
	var changed bool
	tx.Where("host_id = ? AND domain_uuid = ?", hostID, vmInfo.UUID).Limit(1).Find(&existingVMs)

	if len(existingVMs) == 0 {
		log.Infof("Discovered new VM '%s' on host '%s' (not managed). To import, call ImportVM or ImportAllVMs.", vmName, hostID)
		tx.Rollback()
		return false, nil
	} else { // --- Case 2: Existing VM, perform drift detection ---
		existingVM = existingVMs[0]
		updates := make(map[string]interface{})
		driftDetails := make(map[string]map[string]interface{})

		// Always update observed state from libvirt
		newLibvirtState := mapLibvirtStateToVMState(vmInfo.State)
		if existingVM.LibvirtState != newLibvirtState {
			updates["libvirt_state"] = newLibvirtState
			changed = true
		}

		// Sync intended state with libvirt state if they differ and VM is not in a task state
		// This handles cases where async operations (shutdown/reboot) have completed
		if existingVM.TaskState == "" && existingVM.State != newLibvirtState {
			log.Debugf("Syncing intended state for %s: %s -> %s (no task running)", vmInfo.Name, existingVM.State, newLibvirtState)
			updates["state"] = newLibvirtState
			changed = true
		}

		// Compare configurations for drift
		if existingVM.Name != vmInfo.Name {
			driftDetails["name"] = map[string]interface{}{"db": existingVM.Name, "live": vmInfo.Name}
		}
		if existingVM.VCPUCount != vmInfo.Vcpu {
			driftDetails["vcpu"] = map[string]interface{}{"db": existingVM.VCPUCount, "live": vmInfo.Vcpu}
		}
		if existingVM.MemoryBytes != (vmInfo.MaxMem * 1024) {
			driftDetails["memory"] = map[string]interface{}{"db": existingVM.MemoryBytes, "live": vmInfo.MaxMem * 1024}
		}

		if len(driftDetails) > 0 {
			if existingVM.SyncStatus != storage.StatusDrifted {
				updates["sync_status"] = storage.StatusDrifted
				changed = true
			}
			driftJSON, _ := json.Marshal(driftDetails)
			updates["drift_details"] = string(driftJSON)
		} else {
			// If there's no drift, ensure the drift flags are cleared
			if existingVM.SyncStatus == storage.StatusDrifted {
				updates["sync_status"] = storage.StatusSynced
				updates["drift_details"] = ""
				changed = true
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(&existingVM).Updates(updates).Error; err != nil {
				tx.Rollback()
				return false, err
			}
		}
	}

	// If a soft-deleted VM row exists with the same DomainUUID, restore it instead
	var softVMs []storage.VirtualMachine
	tx.Unscoped().Where("domain_uuid = ?", vmInfo.UUID).Limit(1).Find(&softVMs)
	if len(softVMs) > 0 {
		softVM := softVMs[0]
		// If the row is soft-deleted, revive it and update fields to match live VM
		if softVM.DeletedAt.Valid {
			log.Infof("Found soft-deleted VM with DomainUUID %s; restoring record for host %s.", vmInfo.UUID, hostID)
			softVM.HostID = hostID
			softVM.Name = vmInfo.Name
			softVM.DomainUUID = vmInfo.UUID
			softVM.State = mapLibvirtStateToVMState(vmInfo.State)
			softVM.VCPUCount = vmInfo.Vcpu
			softVM.MemoryBytes = vmInfo.MaxMem * 1024
			softVM.SyncStatus = storage.StatusSynced
			// Clear deleted_at by saving with Unscoped
			if uerr := tx.Unscoped().Model(&softVM).Updates(map[string]interface{}{
				"host_id":       softVM.HostID,
				"name":          softVM.Name,
				"domain_uuid":   softVM.DomainUUID,
				"state":         softVM.State,
				"libvirt_state": softVM.State,
				"v_cpu_count":   softVM.VCPUCount,
				"memory_bytes":  softVM.MemoryBytes,
				"sync_status":   softVM.SyncStatus,
				"deleted_at":    nil,
			}).Error; uerr != nil {
				tx.Rollback()
				return false, fmt.Errorf("failed to restore soft-deleted VM row: %w", uerr)
			}
			changed = true
			existingVM = softVM
			// Also ingest hardware on initial sync
			hardwareInfo, hwErr := s.connector.GetDomainHardware(hostID, vmName)
			if hwErr != nil {
				log.Verbosef("Warning: could not fetch hardware for restored VM %s: %v", vmInfo.Name, hwErr)
			} else {
				if _, err := s.syncVMHardware(tx, existingVM.UUID, hostID, hardwareInfo, &vmInfo.Graphics); err != nil {
					tx.Rollback()
					return false, fmt.Errorf("failed to sync hardware for restored VM: %w", err)
				}
			}
			// We're done with restoration: commit transaction and return
			if cerr := tx.Commit().Error; cerr != nil {
				return false, cerr
			}
			return changed, nil
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return changed, nil
}

// syncVMHardware intelligently syncs hardware state, only performing writes when necessary.
// syncVMSecurityLabels handles security label configuration synchronization for a VM
func (s *HostService) syncVMSecurityLabels(tx *gorm.DB, vmUUID string, securityLabels []libvirt.SecurityLabelInfo) (bool, error) {
	changed := false

	for _, label := range securityLabels {
		relabel := false
		if label.Relabel == "yes" {
			relabel = true
		}
		secLabel := storage.SecurityLabel{
			VMUUID:  vmUUID,
			Type:    label.Type,
			Label:   label.Label,
			Relabel: relabel,
		}
		var existingLabels []storage.SecurityLabel
		tx.Where("vm_uuid = ? AND type = ?", vmUUID, label.Type).Limit(1).Find(&existingLabels)
		if len(existingLabels) == 0 {
			if err := tx.Create(&secLabel).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingLabels[0]).Updates(map[string]interface{}{
				"label":   secLabel.Label,
				"relabel": secLabel.Relabel,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMLaunchSecurity handles launch security configuration synchronization for a VM
func (s *HostService) syncVMLaunchSecurity(tx *gorm.DB, vmUUID string, launchSecurity *libvirt.LaunchSecurityInfo) (bool, error) {
	changed := false

	if launchSecurity != nil {
		launchSec := storage.LaunchSecurity{
			VMUUID: vmUUID,
			Type:   launchSecurity.Type,
		}
		// Convert string fields to appropriate types
		if cbitpos, err := strconv.ParseUint(launchSecurity.CBitPos, 10, 32); err == nil {
			launchSec.CBitPos = uint(cbitpos)
		}
		if reducedBits, err := strconv.ParseUint(launchSecurity.ReducedPhysBits, 10, 32); err == nil {
			launchSec.ReducedPhysBits = uint(reducedBits)
		}
		if policy, err := strconv.ParseUint(launchSecurity.Policy, 10, 64); err == nil {
			launchSec.Policy = policy
		}
		launchSec.DHCert = launchSecurity.DHCert
		launchSec.Session = launchSecurity.Session

		var existingLaunch []storage.LaunchSecurity
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingLaunch)
		if len(existingLaunch) == 0 {
			if err := tx.Create(&launchSec).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingLaunch[0]).Updates(map[string]interface{}{
				"type":              launchSec.Type,
				"cbitpos":           launchSec.CBitPos,
				"reduced_phys_bits": launchSec.ReducedPhysBits,
				"policy":            launchSec.Policy,
				"dh_cert":           launchSec.DHCert,
				"session":           launchSec.Session,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMHypervisorFeatures handles hypervisor feature configuration synchronization for a VM
func (s *HostService) syncVMHypervisorFeatures(tx *gorm.DB, vmUUID string, hypervisorFeatures []libvirt.HypervisorFeatureInfo) (bool, error) {
	changed := false

	for _, feature := range hypervisorFeatures {
		hvFeature := storage.HypervisorFeature{
			VMUUID: vmUUID,
			Name:   feature.Name,
			State:  feature.State,
		}
		var existingHV []storage.HypervisorFeature
		tx.Where("vm_uuid = ? AND name = ?", vmUUID, feature.Name).Limit(1).Find(&existingHV)
		if len(existingHV) == 0 {
			if err := tx.Create(&hvFeature).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingHV[0]).Update("state", hvFeature.State).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMLifecycleActions handles lifecycle action configuration synchronization for a VM
func (s *HostService) syncVMLifecycleActions(tx *gorm.DB, vmUUID string, lifecycleActions *libvirt.LifecycleActionInfo) (bool, error) {
	changed := false

	if lifecycleActions != nil {
		lifecycle := storage.LifecycleAction{
			VMUUID:        vmUUID,
			OnPoweroff:    lifecycleActions.OnPoweroff,
			OnReboot:      lifecycleActions.OnReboot,
			OnCrash:       lifecycleActions.OnCrash,
			OnLockfailure: lifecycleActions.OnLockFailure,
		}

		var existingLifecycle []storage.LifecycleAction
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingLifecycle)
		if len(existingLifecycle) == 0 {
			if err := tx.Create(&lifecycle).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingLifecycle[0]).Updates(map[string]interface{}{
				"on_poweroff":    lifecycle.OnPoweroff,
				"on_reboot":      lifecycle.OnReboot,
				"on_crash":       lifecycle.OnCrash,
				"on_lockfailure": lifecycle.OnLockfailure,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMClockConfig handles clock configuration synchronization for a VM
func (s *HostService) syncVMClockConfig(tx *gorm.DB, vmUUID string, clockConfig *libvirt.ClockInfo) (bool, error) {
	changed := false

	if clockConfig != nil {
		clock := storage.Clock{
			VMUUID: vmUUID,
			Offset: clockConfig.Offset,
		}
		if len(clockConfig.Timers) > 0 {
			timersJSON, _ := json.Marshal(clockConfig.Timers)
			clock.ConfigJSON = string(timersJSON)
		}

		var existingClock []storage.Clock
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingClock)
		if len(existingClock) == 0 {
			if err := tx.Create(&clock).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingClock[0]).Updates(map[string]interface{}{
				"offset":      clock.Offset,
				"config_json": clock.ConfigJSON,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMPerfEvents handles performance event configuration synchronization for a VM
func (s *HostService) syncVMPerfEvents(tx *gorm.DB, vmUUID string, perfEvents []libvirt.PerfEventInfo) (bool, error) {
	changed := false

	for _, event := range perfEvents {
		perfEvent := storage.PerfEvent{
			VMUUID: vmUUID,
			Name:   event.Name,
			State:  event.Event, // Event field contains the state ('on'/'off')
		}

		var existingPerf []storage.PerfEvent
		tx.Where("vm_uuid = ? AND name = ?", vmUUID, event.Name).Limit(1).Find(&existingPerf)
		if len(existingPerf) == 0 {
			if err := tx.Create(&perfEvent).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingPerf[0]).Update("state", perfEvent.State).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMOSConfig handles operating system configuration synchronization for a VM
func (s *HostService) syncVMOSConfig(tx *gorm.DB, vmUUID string, osConfig *libvirt.OSConfigInfo) (bool, error) {
	changed := false

	if osConfig != nil {
		osConfigData := storage.OSConfig{
			VMUUID:         vmUUID,
			LoaderPath:     osConfig.Init, // This might need adjustment based on actual XML structure
			LoaderType:     osConfig.Arch,
			BootMenuEnable: osConfig.BootMenu != nil && osConfig.BootMenu.Enable == "yes",
		}
		if osConfig.BootMenu != nil {
			if timeout, err := strconv.Atoi(osConfig.BootMenu.Timeout); err == nil {
				osConfigData.BootMenuTimeout = uint(timeout)
			}
		}
		if osConfig.Type != "" {
			osConfigData.Firmware = osConfig.Type
		}

		var existingOS []storage.OSConfig
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingOS)
		if len(existingOS) == 0 {
			if err := tx.Create(&osConfigData).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			// Update existing
			if err := tx.Model(&existingOS[0]).Updates(map[string]interface{}{
				"loader_path":       osConfigData.LoaderPath,
				"loader_type":       osConfigData.LoaderType,
				"boot_menu_enable":  osConfigData.BootMenuEnable,
				"boot_menu_timeout": osConfigData.BootMenuTimeout,
				"firmware":          osConfigData.Firmware,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMSMBIOS handles SMBIOS information synchronization for a VM
func (s *HostService) syncVMSMBIOS(tx *gorm.DB, vmUUID string, smbiosInfo []libvirt.SMBIOSInfo) (bool, error) {
	changed := false

	if len(smbiosInfo) > 0 {
		for _, smbios := range smbiosInfo {
			smbiosConfig := storage.SMBIOSSystemInfo{
				VMUUID: vmUUID,
				Type:   "smbios", // Default type
			}
			// Store mode in ConfigJSON for now
			configJSON, _ := json.Marshal(map[string]string{"mode": smbios.Mode})
			smbiosConfig.ConfigJSON = string(configJSON)

			var existingSMBIOS []storage.SMBIOSSystemInfo
			tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingSMBIOS)
			if len(existingSMBIOS) == 0 {
				if err := tx.Create(&smbiosConfig).Error; err != nil {
					return false, err
				}
				changed = true
			} else {
				if err := tx.Model(&existingSMBIOS[0]).Update("config_json", smbiosConfig.ConfigJSON).Error; err != nil {
					return false, err
				}
				changed = true
			}
		}
	}

	return changed, nil
}

// syncVMCPUFeatures handles CPU feature flags synchronization for a VM
func (s *HostService) syncVMCPUFeatures(tx *gorm.DB, vmUUID string, cpuFeatures []libvirt.CPUFeatureInfo) (bool, error) {
	changed := false

	for _, feature := range cpuFeatures {
		cpuFeature := storage.CPUFeature{
			VMUUID: vmUUID,
			Name:   feature.Name,
			Policy: feature.Policy,
		}
		var existingFeatures []storage.CPUFeature
		tx.Where("vm_uuid = ? AND name = ?", vmUUID, feature.Name).Limit(1).Find(&existingFeatures)
		if len(existingFeatures) == 0 {
			if err := tx.Create(&cpuFeature).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingFeatures[0]).Update("policy", cpuFeature.Policy).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMCPUTopology handles CPU topology configuration synchronization for a VM
func (s *HostService) syncVMCPUTopology(tx *gorm.DB, vmUUID string, cpuInfo *libvirt.CPUConfigInfo) (bool, error) {
	changed := false

	if cpuInfo != nil && cpuInfo.Topology != nil {
		cpuTopology := storage.CPUTopology{
			VMUUID:  vmUUID,
			Sockets: uint(cpuInfo.Topology.Sockets),
			Cores:   uint(cpuInfo.Topology.Cores),
			Threads: uint(cpuInfo.Topology.Threads),
		}
		var existingTopology []storage.CPUTopology
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&existingTopology)
		if len(existingTopology) == 0 {
			if err := tx.Create(&cpuTopology).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingTopology[0]).Updates(map[string]interface{}{
				"sockets": cpuTopology.Sockets,
				"cores":   cpuTopology.Cores,
				"threads": cpuTopology.Threads,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMMemoryConfig handles memory configuration synchronization for a VM
func (s *HostService) syncVMMemoryConfig(tx *gorm.DB, vmUUID string, memoryBacking *libvirt.MemoryBackingInfo) (bool, error) {
	changed := false

	if memoryBacking != nil {
		memConfig := storage.MemoryConfig{
			VMUUID:       vmUUID,
			ConfigType:   "backing",
			SourceType:   memoryBacking.Source,
			Nosharepages: memoryBacking.NoSharePages,
			Locked:       memoryBacking.Locked,
		}
		if memoryBacking.HugePages != nil {
			hugePagesJSON, _ := json.Marshal(memoryBacking.HugePages.Page)
			memConfig.ConfigJSON = string(hugePagesJSON)
		}

		var existingMem []storage.MemoryConfig
		tx.Where("vm_uuid = ? AND config_type = ?", vmUUID, "backing").Limit(1).Find(&existingMem)
		if len(existingMem) == 0 {
			if err := tx.Create(&memConfig).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			if err := tx.Model(&existingMem[0]).Updates(map[string]interface{}{
				"source_type":  memConfig.SourceType,
				"nosharepages": memConfig.Nosharepages,
				"locked":       memConfig.Locked,
				"config_json":  memConfig.ConfigJSON,
			}).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMHostdevs handles host device (PCI/USB passthrough) synchronization for a VM
func (s *HostService) syncVMHostdevs(tx *gorm.DB, vmUUID, hostID string, hostdevs []libvirt.HostdevInfo) (bool, error) {
	changed := false

	for _, hd := range hostdevs {
		addr := fmt.Sprintf("%s:%s:%s.%s", hd.Source.Address.Domain, hd.Source.Address.Bus, hd.Source.Address.Slot, hd.Source.Address.Function)
		// find or create HostDevice by host and address
		var hdResource storage.HostDevice
		var hdList []storage.HostDevice
		tx.Where("host_id = ? AND address = ?", hostID, addr).Limit(1).Find(&hdList)
		if len(hdList) == 0 {
			hdResource = storage.HostDevice{HostID: hostID, Type: hd.Type, Address: addr}
			if err := tx.Create(&hdResource).Error; err != nil {
				return false, err
			}
		} else {
			hdResource = hdList[0]
		}

		// ensure attachment exists
		var hda storage.HostDeviceAttachment
		var hdaList []storage.HostDeviceAttachment
		tx.Where("vm_uuid = ? AND host_device_id = ?", vmUUID, hdResource.ID).Limit(1).Find(&hdaList)
		if len(hdaList) == 0 {
			hda = storage.HostDeviceAttachment{VMUUID: vmUUID, HostDeviceID: hdResource.ID}
			if err := tx.Create(&hda).Error; err != nil {
				return false, err
			}
			// create attachment index idempotently
			devID3 := hdResource.ID
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "hostdevice", AttachmentID: hda.ID, DeviceID: &devID3}
			var existingAllocs []storage.AttachmentIndex
			tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&existingAllocs)
			if len(existingAllocs) == 0 {
				var deviceAllocs3 []storage.AttachmentIndex
				tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&deviceAllocs3)
				if len(deviceAllocs3) == 0 {
					// No live allocation — check for a soft-deleted allocation and restore it
					var softAllocs []storage.AttachmentIndex
					tx.Unscoped().Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&softAllocs)
					if len(softAllocs) > 0 {
						softAlloc := softAllocs[0]
						if softAlloc.DeletedAt.Valid {
							if uerr2 := tx.Unscoped().Model(&softAlloc).Updates(map[string]interface{}{
								"vm_uuid":       alloc.VMUUID,
								"attachment_id": alloc.AttachmentID,
								"deleted_at":    nil,
							}).Error; uerr2 != nil {
								return false, fmt.Errorf("failed to restore soft-deleted attachment_index for hostdevice device %v: %w", alloc.DeviceID, uerr2)
							}
						} else {
							return false, fmt.Errorf("unexpected state: attachment_index exists but was not returned by query for device %v", alloc.DeviceID)
						}
					} else {
						// No existing allocation at all; create one via helper
						if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
							return false, err
						}
					}
				}
			}
			changed = true
		} else {
			hda = hdaList[0]
		}
	}

	return changed, nil
}

// syncVMBlockDevs handles QEMU blockdev nodes synchronization for a VM
func (s *HostService) syncVMBlockDevs(tx *gorm.DB, blockdevs []libvirt.BlockDev) (bool, error) {
	changed := false

	for _, bd := range blockdevs {
		var b storage.BlockDev
		var bList []storage.BlockDev
		tx.Where("node_name = ?", bd.NodeName).Limit(1).Find(&bList)
		if len(bList) == 0 {
			b = storage.BlockDev{NodeName: bd.NodeName, Driver: bd.Driver.Name, Format: bd.Driver.Type}
			if err := tx.Create(&b).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			b = bList[0]
			// could update fields if necessary in future
		}
	}

	return changed, nil
}

// syncVMNUMANodes handles NUMA topology synchronization for a VM
func (s *HostService) syncVMNUMANodes(tx *gorm.DB, vmUUID string, numaNodes []libvirt.NUMANodeInfo) (bool, error) {
	changed := false

	var existingNUMA []storage.NUMANode
	if err := tx.Where("vm_uuid = ?", vmUUID).Find(&existingNUMA).Error; err != nil {
		return false, err
	}
	existingNUMAByID := make(map[int]storage.NUMANode)
	for _, n := range existingNUMA {
		existingNUMAByID[n.NodeID] = n
	}
	for _, nn := range numaNodes {
		if existing, ok := existingNUMAByID[nn.ID]; ok {
			updates := make(map[string]interface{})
			if existing.MemoryKB != nn.MemoryKB {
				updates["memory_kb"] = nn.MemoryKB
			}
			if existing.CPUsJSON != nn.CPUs {
				updates["cpus_json"] = nn.CPUs
			}
			if len(updates) > 0 {
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return false, err
				}
				changed = true
			}
			delete(existingNUMAByID, nn.ID)
		} else {
			newN := storage.NUMANode{VMUUID: vmUUID, NodeID: nn.ID, MemoryKB: nn.MemoryKB, CPUsJSON: nn.CPUs}
			if err := tx.Create(&newN).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}
	if len(existingNUMAByID) > 0 {
		var idsToDelete []uint
		for _, d := range existingNUMAByID {
			idsToDelete = append(idsToDelete, d.ID)
		}
		if err := tx.Where("id IN ?", idsToDelete).Delete(&storage.NUMANode{}).Error; err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}

// syncVMIOThreads handles I/O threads synchronization for a VM
func (s *HostService) syncVMIOThreads(tx *gorm.DB, iothreads []libvirt.IOThread) (bool, error) {
	changed := false

	for _, it := range iothreads {
		var thr storage.IOThread
		var thrList []storage.IOThread
		tx.Where("name = ?", it.Name).Limit(1).Find(&thrList)
		if len(thrList) == 0 {
			thr = storage.IOThread{Name: it.Name}
			if err := tx.Create(&thr).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			thr = thrList[0]
		}
	}

	return changed, nil
}

// syncVMMdevs handles mediated device synchronization for a VM
func (s *HostService) syncVMMdevs(tx *gorm.DB, vmUUID string, mdevs []libvirt.MdevInfo) (bool, error) {
	changed := false

	for _, m := range mdevs {
		// create or find MediatedDevice by type+device id
		var md storage.MediatedDevice
		var mdList []storage.MediatedDevice
		tx.Where("type_name = ? AND device_id = ?", m.Type, m.UUID).Limit(1).Find(&mdList)
		if len(mdList) == 0 {
			md = storage.MediatedDevice{TypeName: m.Type, DeviceID: m.UUID}
			if err := tx.Create(&md).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			md = mdList[0]
		}
		// ensure attachment exists
		var mda storage.MediatedDeviceAttachment
		var mdaList []storage.MediatedDeviceAttachment
		tx.Where("vm_uuid = ? AND mdev_id = ?", vmUUID, md.ID).Limit(1).Find(&mdaList)
		if len(mdaList) == 0 {
			mda = storage.MediatedDeviceAttachment{VMUUID: vmUUID, MdevID: md.ID}
			if err := tx.Create(&mda).Error; err != nil {
				return false, err
			}
			devID4 := md.ID
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "mdev", AttachmentID: mda.ID, DeviceID: &devID4}
			var existingAllocs []storage.AttachmentIndex
			tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&existingAllocs)
			if len(existingAllocs) > 0 {
				existingAlloc := existingAllocs[0]
				if existingAlloc.VMUUID != alloc.VMUUID {
					return false, fmt.Errorf("attachment (id=%d) already indexed for VM %s (index id %d)", alloc.AttachmentID, existingAlloc.VMUUID, existingAlloc.ID)
				}
				// already indexed correctly
			} else {
				var deviceAllocs4 []storage.AttachmentIndex
				tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&deviceAllocs4)
				if len(deviceAllocs4) == 0 {
					if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
						return false, err
					}
				}
			}
		} else {
			mda = mdaList[0]
		}
		changed = true
	}

	return changed, nil
}

// syncVMBootConfig handles boot configuration synchronization for a VM
func (s *HostService) syncVMBootConfig(tx *gorm.DB, vmUUID string, boot []libvirt.BootEntry) (bool, error) {
	changed := false

	if len(boot) > 0 {
		bootJSON, _ := json.Marshal(boot)
		var bc storage.BootConfig
		var bcList []storage.BootConfig
		tx.Where("vm_uuid = ?", vmUUID).Limit(1).Find(&bcList)
		if len(bcList) == 0 {
			bc = storage.BootConfig{VMUUID: vmUUID, BootOrderJSON: string(bootJSON)}
			if err := tx.Create(&bc).Error; err != nil {
				return false, err
			}
			changed = true
		} else {
			bc = bcList[0]
			if bc.BootOrderJSON != string(bootJSON) {
				if err := tx.Model(&bc).Updates(map[string]interface{}{"boot_order_json": string(bootJSON)}).Error; err != nil {
					return false, err
				}
				changed = true
			}
		}
	}

	return changed, nil
}

// syncVMVideos handles video device synchronization for a VM
func (s *HostService) syncVMVideos(tx *gorm.DB, vmUUID string, videos []libvirt.VideoInfo) (bool, error) {
	changed := false

	var existingVideoAttachments []storage.VideoAttachment
	if err := tx.Preload("VideoModel").Where("vm_uuid = ?", vmUUID).Find(&existingVideoAttachments).Error; err != nil {
		return false, err
	}
	existingVideoByIndex := make(map[int]storage.VideoAttachment)
	for _, va := range existingVideoAttachments {
		existingVideoByIndex[va.MonitorIndex] = va
	}

	for _, v := range videos {
		modelType := v.Model.Type
		// Ensure VideoModel resource exists
		var vid storage.VideoModel
		var vids []storage.VideoModel
		tx.Where("model_name = ?", modelType).Limit(1).Find(&vids)
		if len(vids) == 0 {
			vid = storage.VideoModel{ModelName: modelType, VRAM: uint(v.Model.VRAM), Heads: v.Model.Heads}
			if err := tx.Create(&vid).Error; err != nil {
				return false, err
			}
		} else {
			vid = vids[0]
		}

		// Use monitor index 0 as default if none provided
		mi := 0
		att, exists := existingVideoByIndex[mi]
		if exists {
			updates := make(map[string]interface{})
			if att.VideoModelID != vid.ID {
				updates["video_model_id"] = vid.ID
			}
			if !att.Primary {
				updates["primary"] = true
			}
			if len(updates) > 0 {
				if err := tx.Model(&att).Updates(updates).Error; err != nil {
					return false, err
				}
			}
			// ensure attachment index exists
			// Videos represent logical models, not exclusive host devices.
			// Treat them like volumes for indexing (DeviceID=nil) so many VMs
			// can reference the same Video model without causing uniqueness conflicts.
			var nilDevID *uint = nil
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "video", AttachmentID: att.ID, DeviceID: nilDevID}
			var existingAllocs []storage.AttachmentIndex
			tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&existingAllocs)
			if len(existingAllocs) == 0 {
				var deviceAllocs []storage.AttachmentIndex
				tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&deviceAllocs)
				if len(deviceAllocs) == 0 {
					// No live allocation — create one via helper
					if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
						return false, err
					}
				}
			}
			changed = true
		} else {
			// Before creating an attachment, ensure the video device isn't already
			// allocated to another VM. If it is, log and skip attaching this device
			// instead of failing the whole sync.
			// Check whether any attachment_index already claims this device id (only meaningful for exclusive devices)
			var existingAllocs []storage.AttachmentIndex
			tx.Where("device_type = ? AND device_id = ?", "video", vid.ID).Limit(1).Find(&existingAllocs)
			if len(existingAllocs) > 0 {
				existingAlloc := existingAllocs[0]
				if existingAlloc.VMUUID != vmUUID {
					log.Verbosef("video device id=%d already allocated to VM %s (attachment_index id %d); skipping attach for VM %s", vid.ID, existingAlloc.VMUUID, existingAlloc.ID, vmUUID)
					// don't attach, continue to next video
					continue
				}
				// if it belongs to same VM, proceed to ensure attachment exists below
			}

			newAtt := storage.VideoAttachment{VMUUID: vmUUID, VideoModelID: vid.ID, MonitorIndex: mi, Primary: true}
			if err := tx.Create(&newAtt).Error; err != nil {
				return false, err
			}
			// Use DeviceID=nil for video model multi-attach semantics.
			var nilDevID2 *uint = nil
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "video", AttachmentID: newAtt.ID, DeviceID: nilDevID2}
			var existingAllocs2 []storage.AttachmentIndex
			tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).Limit(1).Find(&existingAllocs2)
			if len(existingAllocs2) > 0 {
				existingAlloc2 := existingAllocs2[0]
				if existingAlloc2.VMUUID != alloc.VMUUID {
					return false, fmt.Errorf("attachment (id=%d) already indexed for VM %s (index id %d)", alloc.AttachmentID, existingAlloc2.VMUUID, existingAlloc2.ID)
				}
			} else {
				var deviceAllocs2 []storage.AttachmentIndex
				tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&deviceAllocs2)
				if len(deviceAllocs2) == 0 {
					// No live allocation — check for a soft-deleted allocation and restore it
					var softAllocs []storage.AttachmentIndex
					tx.Unscoped().Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&softAllocs)
					if len(softAllocs) > 0 {
						softAlloc := softAllocs[0]
						if softAlloc.DeletedAt.Valid {
							if uerr2 := tx.Unscoped().Model(&softAlloc).Updates(map[string]interface{}{
								"vm_uuid":       alloc.VMUUID,
								"attachment_id": alloc.AttachmentID,
								"deleted_at":    nil,
							}).Error; uerr2 != nil {
								return false, fmt.Errorf("failed to restore soft-deleted attachment_index for video device %v: %w", alloc.DeviceID, uerr2)
							}
						} else {
							return false, fmt.Errorf("unexpected state: attachment_index exists but was not returned by query for device %v", alloc.DeviceID)
						}
					} else {
						// No existing allocation at all; create one via helper
						if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
							return false, err
						}
					}
				} else if len(deviceAllocs2) > 0 {
					existingAlloc2 := deviceAllocs2[0]
					if existingAlloc2.AttachmentID != alloc.AttachmentID || existingAlloc2.VMUUID != alloc.VMUUID {
						return false, fmt.Errorf("device (type=%s id=%d) already allocated to VM %s (attachment_index id %d)", alloc.DeviceType, alloc.DeviceID, existingAlloc2.VMUUID, existingAlloc2.ID)
					}
				}
			}
			changed = true
		}
	}

	return changed, nil
}

// syncVMConsole handles graphics/console synchronization for a VM
func (s *HostService) syncVMConsole(tx *gorm.DB, vmUUID, hostID string, graphics *libvirt.GraphicsInfo) (bool, error) {
	changed := false

	var desiredGfxType string
	if graphics.VNC {
		desiredGfxType = "vnc"
	} else if graphics.SPICE {
		desiredGfxType = "spice"
	}

	if desiredGfxType == "" {
		// No console desired: remove any Console rows + attachment index for this VM
		if err := tx.Where("device_type = ? AND vm_uuid = ?", "console", vmUUID).Delete(&storage.AttachmentIndex{}).Error; err != nil {
			return false, err
		}
		if err := tx.Where("vm_uuid = ? AND type = ?", vmUUID, desiredGfxType).Delete(&storage.Console{}).Error; err != nil {
			return false, err
		}
		changed = true
	} else {
		// Create or reuse a Console instance for this VM+type
		var console storage.Console
		var consoles []storage.Console
		tx.Where("vm_uuid = ? AND type = ?", vmUUID, desiredGfxType).Limit(1).Find(&consoles)
		if len(consoles) == 0 {
			console = storage.Console{VMUUID: vmUUID, HostID: hostID, Type: desiredGfxType, ModelName: desiredGfxType}
			if err := tx.Create(&console).Error; err != nil {
				return false, err
			}
		} else {
			console = consoles[0]
		}

		// Ensure attachment index references the console instance
		devID2 := console.ID
		alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "console", AttachmentID: console.ID, DeviceID: &devID2}
		var existingAlloc storage.AttachmentIndex
		// First try to find a live allocation by device_id
		var existingAllocs []storage.AttachmentIndex
		tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&existingAllocs)
		if len(existingAllocs) > 0 {
			existingAlloc = existingAllocs[0]
			if existingAlloc.AttachmentID != alloc.AttachmentID || existingAlloc.VMUUID != alloc.VMUUID {
				return false, fmt.Errorf("console (id=%d) already allocated to VM %s (attachment_index id %d)", alloc.DeviceID, existingAlloc.VMUUID, existingAlloc.ID)
			}
			// allocation already present and matching
		} else {
			// No live allocation found — check for a soft-deleted one and restore it
			var softAllocs []storage.AttachmentIndex
			tx.Unscoped().Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).Limit(1).Find(&softAllocs)
			if len(softAllocs) > 0 {
				softAlloc := softAllocs[0]
				if softAlloc.DeletedAt.Valid {
					// Restore and update to point at the current attachment and vm
					if uerr2 := tx.Unscoped().Model(&softAlloc).Updates(map[string]interface{}{
						"vm_uuid":       alloc.VMUUID,
						"attachment_id": alloc.AttachmentID,
						"deleted_at":    nil,
					}).Error; uerr2 != nil {
						return false, fmt.Errorf("failed to restore soft-deleted attachment_index for device %v: %w", alloc.DeviceID, uerr2)
					}
					// restored
				} else {
					// softAlloc exists but not deleted — unexpected since first query didn't find it
					return false, fmt.Errorf("unexpected state: attachment_index exists but was not returned by query for device %v", alloc.DeviceID)
				}
			} else {
				// No existing allocation at all; create one via helper
				if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
					return false, err
				}
			}
		}
		changed = true
	}

	return changed, nil
}

// syncVMDisks handles disk synchronization for a VM
func (s *HostService) syncVMDisks(tx *gorm.DB, vmUUID, hostID string, disks []libvirt.DiskInfo) (bool, error) {
	changed := false

	// Fetch existing disk attachments for this VM
	var existingDiskAttachments []storage.DiskAttachment
	tx.Preload("Disk").Where("vm_uuid = ?", vmUUID).Find(&existingDiskAttachments)
	existingDiskAttachmentsMap := make(map[string]storage.DiskAttachment)
	for _, da := range existingDiskAttachments {
		existingDiskAttachmentsMap[da.DeviceName] = da
	}

	for _, disk := range disks {
		// First, handle Volume if it's pool-managed (has a pool path)
		var volume *storage.Volume
		if disk.Source.File != "" {
			// Assume file-based disk; create Volume if not exists
			var vol storage.Volume
			tx.FirstOrCreate(&vol, storage.Volume{Name: disk.Source.File}, storage.Volume{
				Name: disk.Source.File, Format: disk.Driver.Type, Type: "DISK",
			})
			volume = &vol
		}

		// Create or update Disk resource
		var diskRes storage.Disk
		diskName := disk.Source.File
		if diskName == "" {
			diskName = disk.Source.Dev // block device
		}
		if diskName == "" {
			diskName = fmt.Sprintf("disk-%s", disk.Target.Dev) // fallback
		}
		// For driver options, serialize to JSON
		driverJSON, _ := json.Marshal(map[string]interface{}{
			"name": disk.Driver.Name,
			"type": disk.Driver.Type,
		})
		updates := make(map[string]interface{})
		if volume != nil {
			updates["volume_id"] = volume.ID
		}
		path := disk.Source.File
		if path == "" {
			path = disk.Source.Dev
		}
		updates["path"] = path
		updates["format"] = disk.Driver.Type
		updates["driver_json"] = string(driverJSON)

		// Extract capacity if available
		if disk.Capacity.Value > 0 {
			var capacityBytes uint64
			switch disk.Capacity.Unit {
			case "bytes":
				capacityBytes = disk.Capacity.Value
			case "KB", "KiB":
				capacityBytes = disk.Capacity.Value * 1024
			case "MB", "MiB":
				capacityBytes = disk.Capacity.Value * 1024 * 1024
			case "GB", "GiB":
				capacityBytes = disk.Capacity.Value * 1024 * 1024 * 1024
			case "TB", "TiB":
				capacityBytes = disk.Capacity.Value * 1024 * 1024 * 1024 * 1024
			default:
				// Default to bytes if unit is not specified or unknown
				capacityBytes = disk.Capacity.Value
			}
			updates["capacity_bytes"] = capacityBytes
		} else {
			// If capacity is not available in domain XML, try to get it from libvirt
			var diskPath string
			if disk.Source.File != "" {
				diskPath = disk.Source.File
			} else if disk.Source.Dev != "" {
				diskPath = disk.Source.Dev
			}

			if diskPath != "" {
				if diskSize, err := s.connector.GetDiskSize(hostID, diskPath); err == nil && diskSize > 0 {
					updates["capacity_bytes"] = diskSize
				}
			}
		}

		var diskResList []storage.Disk
		tx.Where("name = ?", diskName).Limit(1).Find(&diskResList)
		if len(diskResList) == 0 {
			diskRes = storage.Disk{Name: diskName}
			for k, v := range updates {
				switch k {
				case "volume_id":
					if vid, ok := v.(uint); ok {
						diskRes.VolumeID = &vid
					}
				case "path":
					diskRes.Path = v.(string)
				case "format":
					diskRes.Format = v.(string)
				case "driver_json":
					diskRes.DriverJSON = v.(string)
				case "capacity_bytes":
					if cb, ok := v.(uint64); ok {
						diskRes.CapacityBytes = cb
					}
				}
			}
			if err := tx.Create(&diskRes).Error; err != nil {
				return false, err
			}
		} else {
			diskRes = diskResList[0]
			// Update existing
			if err := tx.Model(&diskRes).Updates(updates).Error; err != nil {
				return false, err
			}
		}

		attachment, exists := existingDiskAttachmentsMap[disk.Target.Dev]
		if exists {
			updates := make(map[string]interface{})
			if attachment.DiskID != diskRes.ID {
				updates["disk_id"] = diskRes.ID
			}
			if attachment.BusType != disk.Target.Bus {
				updates["bus_type"] = disk.Target.Bus
			}
			if len(updates) > 0 {
				if err := tx.Model(&attachment).Updates(updates).Error; err != nil {
					return false, err
				}
				changed = true
			}
			delete(existingDiskAttachmentsMap, disk.Target.Dev)
		} else {
			newAttachment := storage.DiskAttachment{
				VMUUID: vmUUID, DiskID: diskRes.ID, DeviceName: disk.Target.Dev, BusType: disk.Target.Bus, ReadOnly: disk.ReadOnly, Shareable: disk.Shareable,
			}
			if err := tx.Create(&newAttachment).Error; err != nil {
				return false, err
			}

			// For shareable or readonly disks, allow multi-attach (DeviceID=nil)
			// For exclusive disks, prevent double attach
			var deviceID *uint
			if disk.Shareable || disk.ReadOnly {
				deviceID = nil
			} else {
				// Check if already attached to another VM
				var existingAllocs []storage.AttachmentIndex
				tx.Where("device_type = ? AND device_id = ?", "disk", diskRes.ID).Limit(1).Find(&existingAllocs)
				if len(existingAllocs) > 0 {
					existingAlloc := existingAllocs[0]
					if existingAlloc.VMUUID != vmUUID {
						log.Verbosef("disk device id=%d already allocated to VM %s (attachment_index id %d); skipping attach for VM %s", diskRes.ID, existingAlloc.VMUUID, existingAlloc.ID, vmUUID)
						// don't attach, continue to next disk
						continue
					}
				}
				devID := diskRes.ID
				deviceID = &devID
			}

			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "disk", AttachmentID: newAttachment.ID, DeviceID: deviceID}
			if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
				return false, err
			}
			changed = true
		}
	}

	// Clean up any stale disk attachments
	if len(existingDiskAttachmentsMap) > 0 {
		var idsToDelete []uint
		for _, attachment := range existingDiskAttachmentsMap {
			idsToDelete = append(idsToDelete, attachment.ID)
		}
		// Remove index entries first, then remove attachment rows within the same tx
		if err := tx.Where("device_type = ? AND attachment_id IN ?", "disk", idsToDelete).Delete(&storage.AttachmentIndex{}).Error; err != nil {
			return false, err
		}
		if err := tx.Where("id IN ?", idsToDelete).Delete(&storage.DiskAttachment{}).Error; err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}

// syncVMNetworks handles network/port synchronization for a VM
func (s *HostService) syncVMNetworks(tx *gorm.DB, vmUUID string, hostID string, networks []libvirt.NetworkInfo) (bool, error) {
	var changed bool = false

	// Fetch existing attachments for this VM (if any) and map by MAC
	var existingPortAttachments []storage.PortAttachment
	if err := tx.Preload("Port").Where("vm_uuid = ?", vmUUID).Find(&existingPortAttachments).Error; err != nil {
		return false, err
	}
	existingByMAC := make(map[string]storage.PortAttachment)
	existingByDev := make(map[string]storage.PortAttachment)
	for _, a := range existingPortAttachments {
		if a.MACAddress != "" {
			existingByMAC[a.MACAddress] = a
		}
		if a.DeviceName != "" {
			existingByDev[a.DeviceName] = a
		}
	}

	for _, net := range networks {
		// Prefer matching by device name (vm-scoped unique). If not present, fall back to MAC address.
		att, exists := existingByDev[net.Target.Dev]
		if !exists && net.Mac.Address != "" {
			att, exists = existingByMAC[net.Mac.Address]
		}

		updates := make(map[string]interface{})
		if !exists {
			// Create or find network
			var network storage.Network
			networkUUID := uuid.NewSHA1(uuid.Nil, []byte(fmt.Sprintf("%s:%s", hostID, net.Source.Bridge)))
			tx.FirstOrCreate(&network, storage.Network{UUID: networkUUID.String()}, storage.Network{
				HostID: hostID, Name: net.Source.Bridge, BridgeName: net.Source.Bridge, Mode: "bridged", UUID: networkUUID.String(),
			})

			// Create the port resource (unattached) and then attach it to the VM
			newPort := storage.Port{
				MACAddress: net.Mac.Address,
				ModelName:  net.Model.Type,
				HostID:     hostID,
				PortGroup:  net.Source.PortGroup,
			}
			if err := tx.Create(&newPort).Error; err != nil {
				return false, err
			}

			if network.ID != 0 && newPort.ID != 0 {
				binding := storage.PortBinding{PortID: newPort.ID, NetworkID: network.ID}
				tx.Create(&binding)
			}

			// Create or reuse an attachment row linking the port to the VM.
			var existingAtts []storage.PortAttachment
			tx.Where("vm_uuid = ? AND device_name = ?", vmUUID, net.Target.Dev).Limit(1).Find(&existingAtts)
			var existingAtt storage.PortAttachment
			if len(existingAtts) > 0 {
				existingAtt = existingAtts[0]
				// Attachment exists; ensure it references the correct port and metadata
				updates := make(map[string]interface{})
				if existingAtt.PortID != newPort.ID {
					updates["port_id"] = newPort.ID
				}
				if net.Mac.Address != "" && existingAtt.MACAddress != net.Mac.Address {
					updates["mac_address"] = net.Mac.Address
				}
				if net.Model.Type != "" && existingAtt.ModelName != net.Model.Type {
					updates["model_name"] = net.Model.Type
				}
				if existingAtt.HostID == "" && hostID != "" {
					updates["host_id"] = hostID
				}
				if len(updates) > 0 {
					if err := tx.Model(&existingAtt).Updates(updates).Error; err != nil {
						return false, err
					}
					// Refresh existingAtt.PortID if changed
					if v, ok := updates["port_id"]; ok {
						if id, ok2 := v.(uint); ok2 {
							existingAtt.PortID = id
						}
					}
					changed = true
				}
			} else {
				// Create new attachment
				existingAtt = storage.PortAttachment{
					VMUUID:     vmUUID,
					PortID:     newPort.ID,
					DeviceName: net.Target.Dev,
					MACAddress: net.Mac.Address,
					ModelName:  net.Model.Type,
					HostID:     hostID,
				}
				if err := tx.Create(&existingAtt).Error; err != nil {
					return false, err
				}
				changed = true
			}

			// Ensure attachment index
			portID := existingAtt.PortID
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "port", AttachmentID: existingAtt.ID, DeviceID: &portID}
			if err := s.ensureAttachmentIndex(tx, alloc); err != nil {
				return false, err
			}
			changed = true
		} else {
			// Attachment exists, check for updates on attachment-level and port-level
			if att.DeviceName != net.Target.Dev {
				updates["device_name"] = net.Target.Dev
			}
			modelType := att.ModelName
			if modelType == "" {
				modelType = att.Port.ModelName
			}
			if modelType != net.Model.Type {
				updates["model_name"] = net.Model.Type
			}

			if len(updates) > 0 {
				if err := tx.Model(&att).Updates(updates).Error; err != nil {
					return false, err
				}
				changed = true
			}
			// Remove from map so it's not deleted later
			delete(existingByMAC, net.Mac.Address)
		}
	}

	// Any attachments left are stale and should be removed along with their ports
	if len(existingByMAC) > 0 {
		var portIDsToDelete []uint
		var attachmentIDs []uint
		for _, att := range existingByMAC {
			portIDsToDelete = append(portIDsToDelete, att.PortID)
			attachmentIDs = append(attachmentIDs, att.ID)
		}
		// Remove binding rows, attachments, and port resources
		tx.Where("port_id IN ?", portIDsToDelete).Delete(&storage.PortBinding{})
		tx.Where("id IN ?", attachmentIDs).Delete(&storage.PortAttachment{})
		tx.Where("id IN ?", portIDsToDelete).Delete(&storage.Port{})
		// Also remove any attachment index entries for these attachments
		tx.Where("device_type = ? AND attachment_id IN ?", "port", attachmentIDs).Delete(&storage.AttachmentIndex{})
		changed = true
	}

	return changed, nil
}

func (s *HostService) syncVMHardware(tx *gorm.DB, vmUUID string, hostID string, hardware *libvirt.HardwareInfo, graphics *libvirt.GraphicsInfo) (bool, error) {
	var changed bool = false
	if hardware != nil {
		log.Debugf("syncVMHardware: vm=%s host=%s start devices: disks=%d networks=%d", vmUUID, hostID, len(hardware.Disks), len(hardware.Networks))
	} else {
		log.Debugf("syncVMHardware: vm=%s host=%s start (no hardware)", vmUUID, hostID)
	}
	defer log.Debugf("syncVMHardware: vm=%s host=%s finished", vmUUID, hostID)

	// Get VM name for enhanced API calls
	var vm storage.VirtualMachine
	if err := tx.Where("uuid = ?", vmUUID).First(&vm).Error; err != nil {
		log.Debugf("Failed to find VM for enhanced sync: %v", err)
		// Continue with standard sync if we can't get the VM name
	}

	// --- Enhanced API Data Collection (only if we have VM name) ---
	var memoryDetails *libvirt.MemoryDetails
	var cpuDetails *libvirt.CPUDetails
	var blockDetails []libvirt.BlockDeviceDetail
	var securityDetails []libvirt.SecurityDetail
	var iothreadDetails []libvirt.IOThreadDetail

	if vm.Name != "" {
		// Get enhanced memory details using direct libvirt APIs
		if memDetails, err := s.connector.GetDomainMemoryDetails(hostID, vm.Name); err == nil {
			memoryDetails = memDetails
			log.Debugf("Enhanced memory details retrieved for VM %s", vm.Name)
		} else {
			log.Debugf("Failed to get enhanced memory details for VM %s: %v", vm.Name, err)
		}

		// Get enhanced CPU details using direct libvirt APIs
		if cpuDet, err := s.connector.GetDomainCPUDetails(hostID, vm.Name); err == nil {
			cpuDetails = cpuDet
			log.Debugf("Enhanced CPU details retrieved for VM %s", vm.Name)
		} else {
			log.Debugf("Failed to get enhanced CPU details for VM %s: %v", vm.Name, err)
		}

		// Get enhanced block device details using direct libvirt APIs
		if blockDet, err := s.connector.GetDomainBlockDetails(hostID, vm.Name); err == nil {
			blockDetails = blockDet
			log.Debugf("Enhanced block details retrieved for VM %s", vm.Name)
		} else {
			log.Debugf("Failed to get enhanced block details for VM %s: %v", vm.Name, err)
		}

		// Get enhanced security details using direct libvirt APIs
		if secDet, err := s.connector.GetDomainSecurityDetails(hostID, vm.Name); err == nil {
			securityDetails = secDet
			log.Debugf("Enhanced security details retrieved for VM %s", vm.Name)
		} else {
			log.Debugf("Failed to get enhanced security details for VM %s: %v", vm.Name, err)
		}

		// Get enhanced IOThread details using direct libvirt APIs
		if iothreadDet, err := s.connector.GetDomainIOThreadDetails(hostID, vm.Name); err == nil {
			iothreadDetails = iothreadDet
			log.Debugf("Enhanced IOThread details retrieved for VM %s", vm.Name)
		} else {
			log.Debugf("Failed to get enhanced IOThread details for VM %s: %v", vm.Name, err)
		}
	}

	// --- Sync Networks / Ports ---
	if netChanged, err := s.syncVMNetworks(tx, vmUUID, hostID, hardware.Networks); err != nil {
		return false, err
	} else if netChanged {
		changed = true
	}

	// Sync disks (potentially enhanced with block device details)
	if diskChanged, err := s.syncVMDisks(tx, vmUUID, hostID, hardware.Disks); err != nil {
		return false, err
	} else if diskChanged {
		changed = true
	}

	if consoleChanged, err := s.syncVMConsole(tx, vmUUID, hostID, graphics); err != nil {
		return false, err
	} else if consoleChanged {
		changed = true
	}

	if videoChanged, err := s.syncVMVideos(tx, vmUUID, hardware.Videos); err != nil {
		return false, err
	} else if videoChanged {
		changed = true
	}

	if bootChanged, err := s.syncVMBootConfig(tx, vmUUID, hardware.Boot); err != nil {
		return false, err
	} else if bootChanged {
		changed = true
	}

	if hostdevChanged, err := s.syncVMHostdevs(tx, vmUUID, hostID, hardware.Hostdevs); err != nil {
		return false, err
	} else if hostdevChanged {
		changed = true
	}

	if blockdevChanged, err := s.syncVMBlockDevs(tx, hardware.BlockDevs); err != nil {
		return false, err
	} else if blockdevChanged {
		changed = true
	}

	if numaChanged, err := s.syncVMNUMANodes(tx, vmUUID, hardware.NUMANodes); err != nil {
		return false, err
	} else if numaChanged {
		changed = true
	}

	if iothreadChanged, err := s.syncVMIOThreads(tx, hardware.IOThreads); err != nil {
		return false, err
	} else if iothreadChanged {
		changed = true
	}

	if mdevChanged, err := s.syncVMMdevs(tx, vmUUID, hardware.Mdevs); err != nil {
		return false, err
	} else if mdevChanged {
		changed = true
	}

	if osConfigChanged, err := s.syncVMOSConfig(tx, vmUUID, hardware.OSConfig); err != nil {
		return false, err
	} else if osConfigChanged {
		changed = true
	}

	if smbiosChanged, err := s.syncVMSMBIOS(tx, vmUUID, hardware.SMBIOSInfo); err != nil {
		return false, err
	} else if smbiosChanged {
		changed = true
	}

	if cpuFeaturesChanged, err := s.syncVMCPUFeatures(tx, vmUUID, hardware.CPUFeatures); err != nil {
		return false, err
	} else if cpuFeaturesChanged {
		changed = true
	}

	if cpuTopologyChanged, err := s.syncVMCPUTopology(tx, vmUUID, hardware.CPUInfo); err != nil {
		return false, err
	} else if cpuTopologyChanged {
		changed = true
	}

	if memoryChanged, err := s.syncVMMemoryConfig(tx, vmUUID, hardware.MemoryBacking); err != nil {
		return false, err
	} else if memoryChanged {
		changed = true
	}

	if securityChanged, err := s.syncVMSecurityLabels(tx, vmUUID, hardware.SecurityLabels); err != nil {
		return false, err
	} else if securityChanged {
		changed = true
	}

	if launchSecurityChanged, err := s.syncVMLaunchSecurity(tx, vmUUID, hardware.LaunchSecurity); err != nil {
		return false, err
	} else if launchSecurityChanged {
		changed = true
	}

	if hypervisorFeaturesChanged, err := s.syncVMHypervisorFeatures(tx, vmUUID, hardware.HypervisorFeatures); err != nil {
		return false, err
	} else if hypervisorFeaturesChanged {
		changed = true
	}

	// Log enhanced data collection summary
	if vm.Name != "" {
		log.Debugf("Enhanced API collection summary for VM %s: memory=%v, cpu=%v, block_devs=%d, security=%d, iothreads=%d",
			vm.Name, memoryDetails != nil, cpuDetails != nil, len(blockDetails), len(securityDetails), len(iothreadDetails))
	}

	if lifecycleChanged, err := s.syncVMLifecycleActions(tx, vmUUID, hardware.LifecycleActions); err != nil {
		return false, err
	} else if lifecycleChanged {
		changed = true
	}

	if clockChanged, err := s.syncVMClockConfig(tx, vmUUID, hardware.ClockConfig); err != nil {
		return false, err
	} else if clockChanged {
		changed = true
	}

	if perfEventsChanged, err := s.syncVMPerfEvents(tx, vmUUID, hardware.PerfEvents); err != nil {
		return false, err
	} else if perfEventsChanged {
		changed = true
	}

	return changed, nil
}

func (s *HostService) syncHostVMs(hostID string) (bool, error) {
	liveVMs, err := s.connector.ListAllDomains(hostID)
	if err != nil {
		return false, fmt.Errorf("service failed to list vms for host %s: %w", hostID, err)
	}

	var overallChanged bool

	liveVMUUIDs := make(map[string]struct{})
	for _, vmInfo := range liveVMs {
		liveVMUUIDs[vmInfo.UUID] = struct{}{}
		// Check if VM is already in managed DB
		var existing []storage.VirtualMachine
		s.db.Where("host_id = ? AND domain_uuid = ?", hostID, vmInfo.UUID).Limit(1).Find(&existing)
		if len(existing) > 0 {
			// Already managed, sync drift
			changed, err := s.detectDriftOrIngestVM(hostID, vmInfo.Name, true)
			if err != nil {
				log.Verbosef("Error syncing VM %s: %v", vmInfo.Name, err)
			}
			if changed {
				overallChanged = true
			}
		} else {
			// Not managed, upsert to discovered_vms
			disc := storage.DiscoveredVM{
				HostID:     hostID,
				Name:       vmInfo.Name,
				DomainUUID: vmInfo.UUID,
				InfoJSON:   "", // can populate later if needed
			}
			if changed, err := storage.UpsertDiscoveredVM(s.db, &disc); err != nil {
				log.Verbosef("Error upserting discovered VM %s: %v", vmInfo.Name, err)
			} else if changed {
				// Set overallChanged to notify frontend of discovered VM changes
				overallChanged = true
			}
		}
	}

	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return false, fmt.Errorf("could not get DB records for pruning check: %w", err)
	}

	for _, dbVM := range dbVMs {
		if _, exists := liveVMUUIDs[dbVM.DomainUUID]; !exists {
			// Don't prune VMs that were created very recently (within last 5 minutes)
			// to avoid pruning VMs that were just imported but may not be immediately
			// visible to libvirt due to timing issues
			if time.Since(dbVM.CreatedAt) < 5*time.Minute {
				log.Verbosef("Skipping pruning recently created VM %s (created %v ago)", dbVM.Name, time.Since(dbVM.CreatedAt))
				continue
			}
			// Ensure connector is connected to the host; if not, skip pruning as the
			// missing VM may be due to a transient connection issue.
			if _, connErr := s.connector.GetConnection(hostID); connErr != nil {
				log.Verbosef("Skipping pruning VM %s because host %s is not connected: %v", dbVM.Name, hostID, connErr)
				continue
			}
			log.Verbosef("Pruning VM %s (UUID: %s) from database as it's no longer in libvirt.", dbVM.Name, dbVM.UUID)
			tx := s.db.Begin()
			if err := tx.Where("vm_uuid = ?", dbVM.UUID).Delete(&storage.AttachmentIndex{}).Error; err != nil {
				tx.Rollback()
				log.Verbosef("Warning: failed to delete attachment indices for VM %s: %v", dbVM.Name, err)
				continue
			}
			if err := tx.Delete(&dbVM).Error; err != nil {
				tx.Rollback()
				log.Verbosef("Warning: failed to prune old VM %s: %v", dbVM.Name, err)
				continue
			}
			// Also remove from discovered_vms if present (shouldn't be, but clean up)
			if err := tx.Where("host_id = ? AND domain_uuid = ?", hostID, dbVM.DomainUUID).Delete(&storage.DiscoveredVM{}).Error; err != nil {
				log.Verbosef("Warning: failed to delete discovered VM for %s: %v", dbVM.Name, err)
			}
			if err := tx.Commit().Error; err != nil {
				log.Verbosef("Warning: failed to commit prune transaction for VM %s: %v", dbVM.Name, err)
				continue
			}
			overallChanged = true
		}
	}

	// Prune discovered VMs that are no longer in libvirt
	var dbDiscoveredVMs []storage.DiscoveredVM
	if err := s.db.Where("host_id = ? AND imported = 0", hostID).Find(&dbDiscoveredVMs).Error; err != nil {
		return false, fmt.Errorf("could not get discovered VM records for pruning check: %w", err)
	}

	for _, dbDiscoveredVM := range dbDiscoveredVMs {
		if _, exists := liveVMUUIDs[dbDiscoveredVM.DomainUUID]; !exists {
			log.Verbosef("Pruning discovered VM %s (UUID: %s) from database as it's no longer in libvirt.", dbDiscoveredVM.Name, dbDiscoveredVM.DomainUUID)
			if err := s.db.Where("host_id = ? AND domain_uuid = ?", hostID, dbDiscoveredVM.DomainUUID).Delete(&storage.DiscoveredVM{}).Error; err != nil {
				log.Verbosef("Warning: failed to prune discovered VM %s: %v", dbDiscoveredVM.Name, err)
			} else {
				overallChanged = true
			}
		}
	}

	if overallChanged {
		s.broadcastVMsChanged(hostID)
		s.broadcastDiscoveredVMsChanged(hostID)
	}

	return overallChanged, nil
}

// syncHostStoragePools synchronizes storage pools from libvirt to the database.
func (s *HostService) syncHostStoragePools(hostID string) (bool, error) {
	livePools, err := s.connector.ListAllStoragePools(hostID)
	if err != nil {
		return false, fmt.Errorf("service failed to list storage pools for host %s: %w", hostID, err)
	}

	var overallChanged bool

	// Create a map of live storage pool UUIDs
	livePoolUUIDs := make(map[string]struct{})
	for _, poolInfo := range livePools {
		livePoolUUIDs[poolInfo.UUID] = struct{}{}

		// Check if storage pool is already in the database
		var existing []storage.StoragePool
		s.db.Where("host_id = ? AND uuid = ?", hostID, poolInfo.UUID).Limit(1).Find(&existing)
		if len(existing) == 0 {
			// Create new storage pool
			newPool := storage.StoragePool{
				HostID:          hostID,
				Name:            poolInfo.Name,
				UUID:            poolInfo.UUID,
				Type:            "unknown", // TODO: get from libvirt
				Path:            "",        // TODO: get from libvirt
				CapacityBytes:   poolInfo.CapacityBytes,
				AllocationBytes: poolInfo.AllocationBytes,
			}
			if err := s.db.Create(&newPool).Error; err != nil {
				log.Verbosef("Error creating storage pool %s: %v", poolInfo.Name, err)
			} else {
				log.Verbosef("Created storage pool %s", poolInfo.Name)
				overallChanged = true
			}
		} else {
			// Update existing storage pool
			updates := make(map[string]interface{})
			if existing[0].CapacityBytes != poolInfo.CapacityBytes {
				updates["capacity_bytes"] = poolInfo.CapacityBytes
			}
			if existing[0].AllocationBytes != poolInfo.AllocationBytes {
				updates["allocation_bytes"] = poolInfo.AllocationBytes
			}
			if len(updates) > 0 {
				if err := s.db.Model(&existing[0]).Updates(updates).Error; err != nil {
					log.Verbosef("Error updating storage pool %s: %v", poolInfo.Name, err)
				} else {
					log.Verbosef("Updated storage pool %s", poolInfo.Name)
					overallChanged = true
				}
			}
		}
	}

	// Prune storage pools that are no longer in libvirt
	var dbPools []storage.StoragePool
	if err := s.db.Where("host_id = ?", hostID).Find(&dbPools).Error; err != nil {
		return false, fmt.Errorf("could not get storage pool records for pruning check: %w", err)
	}

	for _, dbPool := range dbPools {
		if _, exists := livePoolUUIDs[dbPool.UUID]; !exists {
			log.Verbosef("Pruning storage pool %s (UUID: %s) from database as it's no longer in libvirt.", dbPool.Name, dbPool.UUID)
			if err := s.db.Delete(&dbPool).Error; err != nil {
				log.Verbosef("Warning: failed to prune storage pool %s: %v", dbPool.Name, err)
			} else {
				overallChanged = true
			}
		}
	}

	return overallChanged, nil
}

func (s *HostService) GetVMStats(hostID, vmName string) (*ProcessedVMStats, error) {

	// Check if host is connected
	if _, err := s.connector.GetConnection(hostID); err != nil {
		return nil, fmt.Errorf("host %s is not connected", hostID)
	}

	// Get raw stats from monitoring or libvirt
	var rawStats *libvirt.VMStats
	cachedStats := s.monitor.GetLastKnownStats(hostID, vmName)
	if cachedStats != nil {
		rawStats = cachedStats
	} else {
		var err error
		rawStats, err = s.connector.GetDomainStats(hostID, vmName)
		if err != nil {
			return nil, fmt.Errorf("failed to get VM stats: %w", err)
		}
	}

	// Transform raw stats to processed frontend-friendly format
	return s.transformVMStats(hostID, vmName, rawStats)
}

// transformVMStats converts raw libvirt stats to frontend-friendly processed stats
func (s *HostService) transformVMStats(hostID, vmName string, rawStats *libvirt.VMStats) (*ProcessedVMStats, error) {
	if rawStats == nil {
		return &ProcessedVMStats{}, nil
	}

	// Calculate total disk I/O (sum across all devices)
	var totalDiskReadBytes, totalDiskWriteBytes int64
	var totalDiskReadReqs, totalDiskWriteReqs int64
	for _, disk := range rawStats.DiskStats {
		totalDiskReadBytes += disk.ReadBytes
		totalDiskWriteBytes += disk.WriteBytes
		totalDiskReadReqs += disk.ReadReq
		totalDiskWriteReqs += disk.WriteReq
	}

	// Calculate total network I/O (sum across all interfaces)
	var totalNetworkRxBytes, totalNetworkTxBytes int64
	for _, net := range rawStats.NetStats {
		totalNetworkRxBytes += net.ReadBytes
		totalNetworkTxBytes += net.WriteBytes
	}

	// Calculate CPU percentages (raw pcentbase, core-relative, guest-normalized, host-normalized)
	corePercent, hostPercent, guestPercent, rawPercent := s.calculateCPUPercent(hostID, vmName, rawStats.CpuTime, rawStats.Vcpu)

	// Calculate uptime (prefer raw stats uptime when available)
	var uptime int64
	if rawStats.Uptime >= 0 {
		uptime = rawStats.Uptime
	} else {
		uptime = s.calculateVMUptime(hostID, vmName)
	}

	// Convert memory from KB to MB
	memoryMB := float64(rawStats.Memory) / 1024.0

	// Expose raw/guest/host percents and apply EWMA smoothing to host-normalized percent for legacy `cpu_percent` field
	smoothedKey := fmt.Sprintf("%s:%s", hostID, vmName)
	var smoothed float64
	if v, ok := s.cpuSmoothStore.Load(smoothedKey); ok {
		prev := v.(float64)
		smoothed = s.cpuSmoothAlpha*hostPercent + (1.0-s.cpuSmoothAlpha)*prev
	} else {
		smoothed = hostPercent
	}
	s.cpuSmoothStore.Store(smoothedKey, smoothed)

	// Calculate disk rates and IOPS using previous sample if available
	diskKey := fmt.Sprintf("%s:%s:disk", hostID, vmName)
	var diskReadKiBPerSec, diskWriteKiBPerSec, diskReadIOPS, diskWriteIOPS float64
	now := time.Now()
	if v, ok := s.prevDiskSamples.Load(diskKey); ok {
		prev := v.(struct {
			readBytes  int64
			writeBytes int64
			readReq    int64
			writeReq   int64
			at         time.Time
		})
		deltaT := now.Sub(prev.at).Seconds()
		if deltaT > 0.05 { // ignore too-small intervals to avoid division blowups
			// compute deltas, guarding against counter wrap or negative values
			deltaReadBytes64 := totalDiskReadBytes - prev.readBytes
			deltaWriteBytes64 := totalDiskWriteBytes - prev.writeBytes
			deltaReadReq64 := totalDiskReadReqs - prev.readReq
			deltaWriteReq64 := totalDiskWriteReqs - prev.writeReq

			if deltaReadBytes64 < 0 {
				deltaReadBytes64 = 0
			}
			if deltaWriteBytes64 < 0 {
				deltaWriteBytes64 = 0
			}
			if deltaReadReq64 < 0 {
				deltaReadReq64 = 0
			}
			if deltaWriteReq64 < 0 {
				deltaWriteReq64 = 0
			}

			deltaReadBytes := float64(deltaReadBytes64)
			deltaWriteBytes := float64(deltaWriteBytes64)
			deltaReadReq := float64(deltaReadReq64)
			deltaWriteReq := float64(deltaWriteReq64)

			// bytes/sec -> KiB/s
			rawReadKiB := (deltaReadBytes / deltaT) / 1024.0
			rawWriteKiB := (deltaWriteBytes / deltaT) / 1024.0

			// IOPS = requests/sec
			rawReadIOPS := deltaReadReq / deltaT
			rawWriteIOPS := deltaWriteReq / deltaT

			// early guard: if there were no actual byte/request deltas, produce zero
			if deltaReadBytes64 == 0 {
				rawReadKiB = 0
				rawReadIOPS = 0
			}
			if deltaWriteBytes64 == 0 {
				rawWriteKiB = 0
				rawWriteIOPS = 0
			}

			// Clamp suspiciously large raw rates before seeding smoothing store
			maxKiB := 1e7 // ~10 GB/s in KiB/s
			if rawReadKiB < 0 {
				rawReadKiB = 0
			}
			if rawWriteKiB < 0 {
				rawWriteKiB = 0
			}
			if rawReadKiB > maxKiB*10 {
				// treat as invalid/outlier (possible counter wrap) -> ignore this sample
				rawReadKiB = 0
				rawReadIOPS = 0
			}
			if rawWriteKiB > maxKiB*10 {
				rawWriteKiB = 0
				rawWriteIOPS = 0
			}

			// Get previous smooth value or initialize to zeros (do NOT seed with raw to avoid large outliers)
			var prevSmooth struct {
				read      float64
				write     float64
				readIOPS  float64
				writeIOPS float64
			}
			if vsm, ok := s.diskSmoothStore.Load(diskKey); ok {
				prevSmooth = vsm.(struct {
					read      float64
					write     float64
					readIOPS  float64
					writeIOPS float64
				})
			} else {
				prevSmooth = struct {
					read      float64
					write     float64
					readIOPS  float64
					writeIOPS float64
				}{read: 0, write: 0, readIOPS: 0, writeIOPS: 0}
			}

			alpha := s.diskSmoothAlpha
			diskReadKiBPerSec = alpha*rawReadKiB + (1.0-alpha)*prevSmooth.read
			diskWriteKiBPerSec = alpha*rawWriteKiB + (1.0-alpha)*prevSmooth.write
			diskReadIOPS = alpha*rawReadIOPS + (1.0-alpha)*prevSmooth.readIOPS
			diskWriteIOPS = alpha*rawWriteIOPS + (1.0-alpha)*prevSmooth.writeIOPS

			// Update smoothing store with clamped/smoothed values
			// Ensure non-negative and clamp to sane upper limit
			if diskReadKiBPerSec < 0 {
				diskReadKiBPerSec = 0
			}
			if diskWriteKiBPerSec < 0 {
				diskWriteKiBPerSec = 0
			}
			if diskReadKiBPerSec > maxKiB {
				diskReadKiBPerSec = maxKiB
			}
			if diskWriteKiBPerSec > maxKiB {
				diskWriteKiBPerSec = maxKiB
			}

			s.diskSmoothStore.Store(diskKey, struct {
				read      float64
				write     float64
				readIOPS  float64
				writeIOPS float64
			}{read: diskReadKiBPerSec, write: diskWriteKiBPerSec, readIOPS: diskReadIOPS, writeIOPS: diskWriteIOPS})
		}
	}
	// Store current disk sample for next delta
	s.prevDiskSamples.Store(diskKey, struct {
		readBytes  int64
		writeBytes int64
		readReq    int64
		writeReq   int64
		at         time.Time
	}{readBytes: totalDiskReadBytes, writeBytes: totalDiskWriteBytes, readReq: totalDiskReadReqs, writeReq: totalDiskWriteReqs, at: now})

	// Calculate network rates (MB/s) with smoothing and guards
	netKey := fmt.Sprintf("%s:%s:net", hostID, vmName)
	var netRxMBps, netTxMBps float64
	if v, ok := s.prevNetSamples.Load(netKey); ok {
		prev := v.(struct {
			rxBytes int64
			txBytes int64
			at      time.Time
		})
		deltaT := now.Sub(prev.at).Seconds()
		if deltaT > 0.05 {
			deltaRx := totalNetworkRxBytes - prev.rxBytes
			deltaTx := totalNetworkTxBytes - prev.txBytes
			if deltaRx < 0 {
				deltaRx = 0
			}
			if deltaTx < 0 {
				deltaTx = 0
			}

			rawRxMBps := (float64(deltaRx) / deltaT) / (1024.0 * 1024.0)
			rawTxMBps := (float64(deltaTx) / deltaT) / (1024.0 * 1024.0)

			// if no activity, produce zero
			if deltaRx == 0 {
				rawRxMBps = 0
			}
			if deltaTx == 0 {
				rawTxMBps = 0
			}

			// smoothing

			var prevNetSmooth struct {
				rx float64
				tx float64
			}
			if vsm, ok := s.netSmoothStore.Load(netKey); ok {
				ps := vsm.(struct {
					rx float64
					tx float64
				})
				prevNetSmooth.rx = ps.rx
				prevNetSmooth.tx = ps.tx
			} else {
				prevNetSmooth = struct {
					rx float64
					tx float64
				}{rx: 0, tx: 0}
			}
			alpha := s.netSmoothAlpha
			netRxMBps = alpha*rawRxMBps + (1.0-alpha)*prevNetSmooth.rx
			netTxMBps = alpha*rawTxMBps + (1.0-alpha)*prevNetSmooth.tx

			// clamp
			if netRxMBps < 0 {
				netRxMBps = 0
			}
			if netTxMBps < 0 {
				netTxMBps = 0
			}
			maxMBps := 10_000.0 // very high cap (10 GB/s)
			if netRxMBps > maxMBps {
				netRxMBps = maxMBps
			}
			if netTxMBps > maxMBps {
				netTxMBps = maxMBps
			}

			// store back into netSmoothStore at netKey for reuse
			s.netSmoothStore.Store(netKey, struct {
				rx float64
				tx float64
			}{rx: netRxMBps, tx: netTxMBps})
		}
	}
	s.prevNetSamples.Store(netKey, struct {
		rxBytes int64
		txBytes int64
		at      time.Time
	}{rxBytes: totalNetworkRxBytes, txBytes: totalNetworkTxBytes, at: now})

	return &ProcessedVMStats{
		CPUPercent:         smoothed,
		CPUPercentCore:     corePercent,
		CPUPercentRaw:      rawPercent,
		CPUPercentGuest:    guestPercent,
		CPUPercentHost:     hostPercent,
		MemoryMB:           memoryMB,
		DiskReadMB:         float64(totalDiskReadBytes) / (1024 * 1024),  // Convert bytes to MB
		DiskWriteMB:        float64(totalDiskWriteBytes) / (1024 * 1024), // Convert bytes to MB
		DiskReadKiBPerSec:  diskReadKiBPerSec,
		DiskWriteKiBPerSec: diskWriteKiBPerSec,
		DiskReadIOPS:       diskReadIOPS,
		DiskWriteIOPS:      diskWriteIOPS,
		NetworkRxMB:        float64(totalNetworkRxBytes) / (1024 * 1024), // Convert bytes to MB
		NetworkTxMB:        float64(totalNetworkTxBytes) / (1024 * 1024), // Convert bytes to MB
		NetworkRxMBps:      netRxMBps,
		NetworkTxMBps:      netTxMBps,
		Uptime:             uptime,
	}, nil
}

// calculateCPUPercent calculates CPU percentage based on previous CPU time sample
func (s *HostService) calculateCPUPercent(hostID, vmName string, currentCpuTime uint64, guestVcpus uint) (float64, float64, float64, float64) {
	key := fmt.Sprintf("%s:%s", hostID, vmName)
	now := time.Now()

	// Load previous sample
	if v, ok := s.prevCpuSamples.Load(key); ok {
		prev := v.(struct {
			cpuTime uint64
			at      time.Time
		})
		// Protect against clock skew / zero interval
		deltaNs := int64(currentCpuTime) - int64(prev.cpuTime)
		deltaT := now.Sub(prev.at).Seconds()
		// Update previous sample
		s.prevCpuSamples.Store(key, struct {
			cpuTime uint64
			at      time.Time
		}{cpuTime: currentCpuTime, at: now})

		if deltaNs <= 0 || deltaT <= 0 {
			return 0.0, 0.0, 0.0, 0.0
		}

		// cpu time is in nanoseconds (domain cumulative CPU ns across vcpus)
		// pcentbase (raw) = (deltaCpuNs * 100) / (deltaT * 1e9)
		rawPercent := (float64(deltaNs) * 100.0) / (deltaT * 1e9)

		// corePercent: interpret as rawPercent divided by guest vcpus? Keep original meaning (per-core equivalent)
		corePercent := rawPercent

		// Fetch host cores (cache)
		var hostCores uint = 1
		if v, ok := s.hostCores.Load(hostID); ok {
			hostCores = v.(uint)
		} else {
			// Attempt to fetch host cores and cache
			if info, err := s.connector.GetHostInfo(hostID); err == nil {
				hostCores = info.Cores
				s.hostCores.Store(hostID, hostCores)
			}
		}

		// Compute host-normalized and guest-normalized percentages per virt-manager
		hostPercent := 0.0
		if hostCores > 0 {
			hostPercent = rawPercent / float64(hostCores)
		}
		guestPercent := 0.0
		if guestVcpus > 0 {
			guestPercent = rawPercent / float64(guestVcpus)
		}

		// Clamp to [0,100]
		if hostPercent < 0 {
			hostPercent = 0
		}
		if hostPercent > 100 {
			hostPercent = 100
		}
		if guestPercent < 0 {
			guestPercent = 0
		}
		if guestPercent > 100 {
			guestPercent = 100
		}

		return corePercent, hostPercent, guestPercent, rawPercent
	}

	// No previous sample; store current and return zeros
	s.prevCpuSamples.Store(key, struct {
		cpuTime uint64
		at      time.Time
	}{cpuTime: currentCpuTime, at: now})
	return 0.0, 0.0, 0.0, 0.0
}

// calculateVMUptime calculates VM uptime in seconds
func (s *HostService) calculateVMUptime(hostID, vmName string) int64 {
	// Try to use monitoring cached uptime if available
	if stats := s.monitor.GetLastKnownStats(hostID, vmName); stats != nil {
		if stats.Uptime >= 0 {
			return stats.Uptime
		}
	}

	// As a fallback, try to fetch fresh domain info which includes uptime
	if vmInfo, err := s.connector.GetDomainInfo(hostID, vmName); err == nil {
		if vmInfo.Uptime >= 0 {
			return vmInfo.Uptime
		}
	}
	return 0
}

// --- VM Actions ---

func (s *HostService) performVMAction(hostID, vmName string, taskState storage.VMTaskState, action func() error, intendedState ...storage.VMState) error {
	// Check if host is connected
	if _, err := s.connector.GetConnection(hostID); err != nil {
		return fmt.Errorf("host %s is not connected", hostID)
	}

	// Set task state
	if err := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("task_state", taskState).Error; err != nil {
		return fmt.Errorf("failed to set task state for %s: %w", vmName, err)
	}
	s.broadcastVMsChanged(hostID)

	// Perform action
	if err := action(); err != nil {
		// Clear task state on failure
		s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("task_state", "")
		s.broadcastVMsChanged(hostID)
		return err
	}

	// Update intended state if provided
	if len(intendedState) > 0 {
		if err := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("state", intendedState[0]).Error; err != nil {
			log.Verbosef("Warning: failed to update intended state for %s: %v", vmName, err)
		}
	}

	// After a successful action, re-run drift detection.
	// This will update the power state and clear any drift flags if the action resolved them.
	if changed, syncErr := s.detectDriftOrIngestVM(hostID, vmName, false); syncErr != nil {
		log.Verbosef("Warning: failed to sync VM %s after %s action: %v", vmName, taskState, syncErr)
	} else if changed {
		s.broadcastVMsChanged(hostID)
	}

	// Clear task state on success
	if err := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("task_state", "").Error; err != nil {
		log.Verbosef("Warning: failed to clear task state for %s: %v", vmName, err)
	}
	s.broadcastVMsChanged(hostID)

	return nil
}

func (s *HostService) StartVM(hostID, vmName string) error {
	return s.performVMAction(hostID, vmName, storage.TaskStateStarting, func() error {
		// If a rebuild is needed, this power cycle will apply the changes.
		// So, we can clear the flag.
		s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("needs_rebuild", false)
		return s.connector.StartDomain(hostID, vmName)
	}, storage.StateActive)
}

func (s *HostService) ShutdownVM(hostID, vmName string) error {
	// Don't set intended state immediately for shutdown - it's async
	// Let polling detect the actual state change
	return s.performVMAction(hostID, vmName, storage.TaskStateStopping, func() error {
		return s.connector.ShutdownDomain(hostID, vmName)
	})
}

func (s *HostService) RebootVM(hostID, vmName string) error {
	// Don't set intended state immediately for reboot - it's async
	// Let polling detect the actual state change
	return s.performVMAction(hostID, vmName, storage.TaskStateRebooting, func() error {
		// If a rebuild is needed, this power cycle will apply the changes.
		// So, we can clear the flag.
		s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("needs_rebuild", false)
		return s.connector.RebootDomain(hostID, vmName)
	})
}

func (s *HostService) ForceOffVM(hostID, vmName string) error {
	return s.performVMAction(hostID, vmName, storage.TaskStatePoweringOff, func() error {
		return s.connector.DestroyDomain(hostID, vmName)
	}, storage.StateStopped)
}

func (s *HostService) ForceResetVM(hostID, vmName string) error {
	return s.performVMAction(hostID, vmName, storage.TaskStateRebooting, func() error {
		// If a rebuild is needed, this power cycle will apply the changes.
		// So, we can clear the flag.
		s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("needs_rebuild", false)
		return s.connector.ResetDomain(hostID, vmName)
	}, storage.StateActive)
}

// --- Drift and Sync Actions ---

// SyncVMFromLibvirt forces an update from the live libvirt state into the database,
// overwriting the DB record and clearing any drift status.
func (s *HostService) SyncVMFromLibvirt(hostID, vmName string) error {
	// Check if host is connected
	if _, err := s.connector.GetConnection(hostID); err != nil {
		return fmt.Errorf("host %s is not connected", hostID)
	}

	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		return fmt.Errorf("could not fetch info for VM %s on host %s: %w", vmName, hostID, err)
	}

	hardwareInfo, err := s.connector.GetDomainHardware(hostID, vmName)
	if err != nil {
		log.Verbosef("Warning: could not fetch hardware for VM %s during manual sync: %v", vmInfo.Name, err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var vmToUpdate storage.VirtualMachine
	if err := tx.Where("host_id = ? AND name = ?", hostID, vmName).First(&vmToUpdate).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not find VM %s in database to sync: %w", vmName, err)
	}

	// Update the main VM record
	updates := map[string]interface{}{
		"Name":         vmInfo.Name,
		"VCPUCount":    vmInfo.Vcpu,
		"MemoryBytes":  vmInfo.MaxMem * 1024,
		"LibvirtState": mapLibvirtStateToVMState(vmInfo.State), // Update observed state
		"SyncStatus":   storage.StatusSynced,
		"DriftDetails": "",
		"NeedsRebuild": false,
	}
	if hardwareInfo != nil {
		if hardwareInfo.OSType != "" {
			updates["OSType"] = hardwareInfo.OSType
		}
		// Update CPU model from hardware info
		if hardwareInfo.CPUInfo != nil && hardwareInfo.CPUInfo.Model != "" {
			updates["CPUModel"] = hardwareInfo.CPUInfo.Model
		} else if hardwareInfo.CPUInfo != nil && hardwareInfo.CPUInfo.Mode != "" {
			// If no specific model, use the mode (e.g., "host-passthrough")
			updates["CPUModel"] = hardwareInfo.CPUInfo.Mode
		}
	}
	if err := tx.Model(&vmToUpdate).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Sync hardware
	if hardwareInfo != nil {
		if _, err := s.syncVMHardware(tx, vmToUpdate.UUID, hostID, hardwareInfo, &vmInfo.Graphics); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to sync hardware during manual sync: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	s.broadcastVMsChanged(hostID)
	return nil
}

// RebuildVMFromDB flags a VM as needing a rebuild. The actual rebuild would
// happen on the next power cycle or via a more complex process.
func (s *HostService) RebuildVMFromDB(hostID, vmName string) error {
	log.Infof("Flagging VM %s for rebuild; changes will be applied on next power cycle.", vmName)
	if err := s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("needs_rebuild", true).Error; err != nil {
		return fmt.Errorf("failed to set needs_rebuild flag for %s: %w", vmName, err)
	}

	// In a real implementation, we would generate an XML from our DB state and
	// redefine the domain in libvirt. For now, we just set the flag.
	// Example:
	// xml, err := s.generateXMLFromDB(hostID, vmName)
	// if err != nil { return err }
	// err = s.connector.RedefineDomain(hostID, xml)
	// if err != nil { return err }

	s.broadcastVMsChanged(hostID)
	return nil
}

// --- WebSocket Message Handling ---

func (s *HostService) HandleSubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok1 := payload["hostId"].(string)
	vmName, ok2 := payload["vmName"].(string)
	if !ok1 || !ok2 {
		log.Verbosef("Invalid payload for vm-stats subscription: %+v", payload)
		return
	}
	s.monitor.Subscribe(client, hostID, vmName)
}

func (s *HostService) HandleUnsubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok1 := payload["hostId"].(string)
	vmName, ok2 := payload["vmName"].(string)
	if !ok1 || !ok2 {
		log.Verbosef("Invalid payload for vm-stats unsubscription: %+v", payload)
		return
	}
	s.monitor.Unsubscribe(client, hostID, vmName)
}

func (s *HostService) HandleHostSubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok := payload["hostId"].(string)
	if !ok {
		log.Verbosef("Invalid payload for host-stats subscription: %+v", payload)
		return
	}
	s.hostMonitor.Subscribe(client, hostID)
}

func (s *HostService) HandleHostUnsubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok := payload["hostId"].(string)
	if !ok {
		log.Verbosef("Invalid payload for host-stats unsubscription: %+v", payload)
		return
	}
	s.hostMonitor.Unsubscribe(client, hostID)
}

func (s *HostService) HandleClientDisconnect(client *ws.Client) {
	s.monitor.UnsubscribeClient(client)
	s.hostMonitor.UnsubscribeClient(client)
}

// --- Monitoring Goroutine Logic ---

func (m *MonitoringManager) Subscribe(client *ws.Client, hostID, vmName string) {
	// NOTE: Previously we attempted to call EnsureHostConnected here to lazily
	// establish a libvirt connection when a client subscribed. That had the
	// side-effect of setting DB task_state to CONNECTING on each subscribe and
	// caused the UI to flash a transient "connecting..." status when a user
	// merely clicked a host in the sidebar. The libvirt connection should be
	// managed separately (e.g. on host add or via a background reconnect loop),
	// so avoid attempting to connect here to prevent spurious UI state changes.

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	sub, exists := m.subscriptions[key]
	if !exists {
		// Check if host is connected before starting monitoring
		if _, err := m.service.connector.GetConnection(hostID); err != nil {
			log.Verbosef("Skipping VM monitoring for %s: host not connected", key)
			// Don't start monitoring if host is not connected
			return
		}

		log.Verbosef("Starting monitoring for %s", key)
		sub = &VmSubscription{
			clients: make(map[*ws.Client]bool),
			stop:    make(chan struct{}),
		}
		m.subscriptions[key] = sub
		go m.pollVmStats(hostID, vmName, sub)
	}
	sub.clients[client] = true
	// If we already have cached stats for this VM, send them immediately so the
	// client doesn't have to wait for the first live fetch. Otherwise, send a
	// warming message while the first poll is in progress.
	sub.mu.RLock()
	cached := sub.lastKnownStats
	sub.mu.RUnlock()
	if cached != nil {
		// Transform cached raw stats to ProcessedVMStats for frontend consistency
		var processedStats *ProcessedVMStats
		if cached.State == -1 {
			// Error case - send basic error stats
			processedStats = &ProcessedVMStats{
				CPUPercent:  0,
				MemoryMB:    0,
				DiskReadMB:  0,
				DiskWriteMB: 0,
				NetworkRxMB: 0,
				NetworkTxMB: 0,
				Uptime:      0,
			}
		} else {
			processedStats, _ = m.service.transformVMStats(hostID, vmName, cached)
		}
		if err := client.SendMessage(ws.Message{Type: "vm-stats-updated", Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": processedStats}}); err != nil {
			// Non-fatal: client might be slow or disconnected.
		}
	} else {
		warmingMsg := ws.Message{Type: "vm-stats-warming", Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName}}
		if err := client.SendMessage(warmingMsg); err != nil {
			// Non-fatal: client might be slow or disconnected.
		}
	}
}

func (m *MonitoringManager) Unsubscribe(client *ws.Client, hostID, vmName string) {
	// Do not call EnsureHostConnected here. Attempting to connect on every
	// subscribe caused the server to set transient task_state values which
	// the UI then displayed as "connecting..." when users clicked hosts/VMs.
	// Connection lifecycle should be handled by host add and background
	// reconnection logic instead.
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	if sub, exists := m.subscriptions[key]; exists {
		delete(sub.clients, client)
		if len(sub.clients) == 0 {
			log.Verbosef("Stopping monitoring for %s", key)
			close(sub.stop)
			delete(m.subscriptions, key)
		}
	}
}

func (m *MonitoringManager) UnsubscribeClient(client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, sub := range m.subscriptions {
		if _, ok := sub.clients[client]; ok {
			delete(sub.clients, client)
			if len(sub.clients) == 0 {
				log.Verbosef("Stopping monitoring for %s due to client disconnect", key)
				close(sub.stop)
				delete(m.subscriptions, key)
			}
		}
	}
}

func (m *MonitoringManager) StopHostMonitoring(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop monitoring for all VMs on this host
	for key, sub := range m.subscriptions {
		if strings.HasPrefix(key, hostID+":") {
			log.Verbosef("Stopping VM monitoring for %s due to host disconnect", key)
			close(sub.stop)
			delete(m.subscriptions, key)
		}
	}
}

func (m *MonitoringManager) GetLastKnownStats(hostID, vmName string) *libvirt.VMStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	if sub, exists := m.subscriptions[key]; exists {
		sub.mu.RLock()
		defer sub.mu.RUnlock()
		return sub.lastKnownStats
	}
	return nil
}

func (m *MonitoringManager) pollVmStats(hostID, vmName string, sub *VmSubscription) {
	// Perform an immediate fetch to provide instant feedback, then poll on a ticker.
	log.Debugf("pollVmStats: immediate fetch for %s:%s", hostID, vmName)
	stats, err := m.service.connector.GetDomainStats(hostID, vmName)
	if err != nil {
		stats = &libvirt.VMStats{State: -1}
	}
	sub.mu.Lock()
	sub.lastKnownStats = stats
	sub.mu.Unlock()

	// Transform raw stats to ProcessedVMStats for frontend consistency
	var processedStats *ProcessedVMStats
	if stats.State == -1 {
		// Error case - send basic error stats
		processedStats = &ProcessedVMStats{
			CPUPercent:  0,
			MemoryMB:    0,
			DiskReadMB:  0,
			DiskWriteMB: 0,
			NetworkRxMB: 0,
			NetworkTxMB: 0,
			Uptime:      0,
		}
	} else {
		processedStats, _ = m.service.transformVMStats(hostID, vmName, stats)
	}

	m.service.hub.BroadcastMessage(ws.Message{
		Type:    "vm-stats-updated",
		Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": processedStats},
	})

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := m.service.connector.GetDomainStats(hostID, vmName)
			if err != nil {
				stats = &libvirt.VMStats{State: -1} // Use an invalid state to signal error
			}

			sub.mu.Lock()
			sub.lastKnownStats = stats
			sub.mu.Unlock()

			// Transform raw stats to ProcessedVMStats for frontend consistency
			var processedStats *ProcessedVMStats
			if stats.State == -1 {
				// Error case - send basic error stats
				processedStats = &ProcessedVMStats{
					CPUPercent:  0,
					MemoryMB:    0,
					DiskReadMB:  0,
					DiskWriteMB: 0,
					NetworkRxMB: 0,
					NetworkTxMB: 0,
					Uptime:      0,
				}
			} else {
				processedStats, _ = m.service.transformVMStats(hostID, vmName, stats)
			}

			m.service.hub.BroadcastMessage(ws.Message{
				Type:    "vm-stats-updated",
				Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": processedStats},
			})

			statsState := mapLibvirtStateToVMState(stats.State)
			if statsState != storage.StateActive {
				log.Verbosef("VM %s is not running (state: %s), stopping stats polling.", vmName, statsState)
				m.mu.Lock()
				delete(m.subscriptions, fmt.Sprintf("%s:%s", hostID, vmName))
				m.mu.Unlock()
				return
			}
		case <-sub.stop:
			return
		}
	}
}

// --- Host Monitoring Goroutine Logic ---

func (m *HostMonitoringManager) Subscribe(client *ws.Client, hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub, exists := m.subscriptions[hostID]
	if !exists {
		// Check if host is connected before starting monitoring
		if _, err := m.service.connector.GetConnection(hostID); err != nil {
			log.Verbosef("Skipping host monitoring for %s: host not connected", hostID)
			// Don't start monitoring if host is not connected
			return
		}

		log.Verbosef("Starting host monitoring for %s", hostID)
		sub = &HostSubscription{
			clients: make(map[*ws.Client]bool),
			stop:    make(chan struct{}),
		}
		m.subscriptions[hostID] = sub
		go m.pollHostStats(hostID, sub)
	}
	sub.clients[client] = true
	// If we already have cached host stats, send them immediately to the new
	// client. Otherwise send a warming message while the initial poll runs.
	sub.mu.RLock()
	cachedHost := sub.lastKnownStats
	sub.mu.RUnlock()
	if cachedHost != nil {
		if err := client.SendMessage(ws.Message{Type: "host-stats-updated", Payload: ws.MessagePayload{"hostId": hostID, "stats": cachedHost}}); err != nil {
			// Non-fatal: client might be slow or disconnected.
		}
	} else {
		warmingMsg := ws.Message{Type: "host-stats-warming", Payload: ws.MessagePayload{"hostId": hostID}}
		if err := client.SendMessage(warmingMsg); err != nil {
			// Non-fatal: client might be slow or disconnected.
		}
	}
}

func (m *HostMonitoringManager) Unsubscribe(client *ws.Client, hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, exists := m.subscriptions[hostID]; exists {
		delete(sub.clients, client)
		if len(sub.clients) == 0 {
			log.Verbosef("Stopping host monitoring for %s", hostID)
			close(sub.stop)
			delete(m.subscriptions, hostID)
		}
	}
}

func (m *HostMonitoringManager) UnsubscribeClient(client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for hostID, sub := range m.subscriptions {
		if _, ok := sub.clients[client]; ok {
			delete(sub.clients, client)
			if len(sub.clients) == 0 {
				log.Verbosef("Stopping host monitoring for %s due to client disconnect", hostID)
				close(sub.stop)
				delete(m.subscriptions, hostID)
			}
		}
	}
}

func (m *HostMonitoringManager) StopHostMonitoring(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, exists := m.subscriptions[hostID]; exists {
		log.Verbosef("Stopping host monitoring for %s due to host disconnect", hostID)
		close(sub.stop)
		delete(m.subscriptions, hostID)
	}
}

// --- Dashboard Methods ---

// GetDashboardStats aggregates system-wide statistics for the dashboard.
func (s *HostService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get all hosts
	hosts, err := s.GetAllHosts()
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}

	stats.Infrastructure.TotalHosts = len(hosts)

	// Count connected hosts and get their info
	var connectedHosts []storage.Host
	var totalMemoryBytes uint64
	var totalCPUs int

	for _, host := range hosts {
		if _, err := s.connector.GetConnection(host.ID); err == nil {
			connectedHosts = append(connectedHosts, host)

			// Get host info for resource calculations
			if hostInfo, err := s.GetHostInfo(host.ID); err == nil {
				totalMemoryBytes += hostInfo.Memory
				totalCPUs += int(hostInfo.CPU)
			}
		}
	}

	stats.Infrastructure.ConnectedHosts = len(connectedHosts)

	// Get all VMs across all hosts
	var allVMs []VMView
	var totalUsedMemory uint64
	var totalAllocatedCPUs int

	for _, host := range connectedHosts {
		vms, err := s.GetVMsForHostFromDB(host.ID)
		if err != nil {
			log.Verbosef("Failed to get VMs for host %s: %v", host.ID, err)
			continue
		}
		allVMs = append(allVMs, vms...)

		// Calculate used resources
		for _, vm := range vms {
			totalUsedMemory += vm.MemoryBytes
			totalAllocatedCPUs += int(vm.VCPUCount)
		}
	}

	stats.Infrastructure.TotalVMs = len(allVMs)

	// Count running and stopped VMs
	for _, vm := range allVMs {
		switch vm.State {
		case storage.StateActive:
			stats.Infrastructure.RunningVMs++
		case storage.StateStopped:
			stats.Infrastructure.StoppedVMs++
		}
	}

	// Calculate resource statistics
	stats.Resources.TotalMemoryGB = float64(totalMemoryBytes) / (1024 * 1024 * 1024)
	stats.Resources.UsedMemoryGB = float64(totalUsedMemory) / (1024 * 1024 * 1024)
	if totalMemoryBytes > 0 {
		stats.Resources.MemoryUtilization = float64(totalUsedMemory) / float64(totalMemoryBytes) * 100
	}

	stats.Resources.TotalCPUs = totalCPUs
	stats.Resources.AllocatedCPUs = totalAllocatedCPUs
	if totalCPUs > 0 {
		stats.Resources.CPUUtilization = float64(totalAllocatedCPUs) / float64(totalCPUs) * 100
	}

	// Health status
	stats.Health.SystemStatus = "healthy"
	if len(connectedHosts) == 0 {
		stats.Health.SystemStatus = "warning"
	}
	if len(connectedHosts) < len(hosts)/2 {
		stats.Health.SystemStatus = "critical"
	}

	stats.Health.LastSync = time.Now().Format(time.RFC3339)
	stats.Health.Errors = 0
	stats.Health.Warnings = 0

	return stats, nil
}

// GetDashboardActivity generates recent activity entries for the dashboard.
func (s *HostService) GetDashboardActivity(limit int) ([]ActivityEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	var activities []ActivityEntry

	// Get all hosts for activity generation
	hosts, err := s.GetAllHosts()
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}

	// Generate host connection activities
	for _, host := range hosts {
		if _, err := s.connector.GetConnection(host.ID); err == nil {
			activity := ActivityEntry{
				ID:        fmt.Sprintf("host-%s-connected", host.ID),
				Type:      "host_connect",
				Message:   fmt.Sprintf("Host %s connected", host.URI),
				HostID:    host.ID,
				Timestamp: time.Now().Add(-time.Duration(len(activities)*2) * time.Minute).Format(time.RFC3339),
				Severity:  "info",
				Details:   "Hypervisor host is online and ready",
			}
			activities = append(activities, activity)
		}
	}

	// Generate VM activities for running VMs
	for _, host := range hosts {
		if _, err := s.connector.GetConnection(host.ID); err == nil {
			vms, err := s.GetVMsForHostFromDB(host.ID)
			if err != nil {
				continue
			}

			for _, vm := range vms {
				if vm.State == storage.StateActive {
					activity := ActivityEntry{
						ID:        fmt.Sprintf("vm-%s-running", vm.UUID),
						Type:      "vm_state_change",
						Message:   fmt.Sprintf("VM '%s' is running", vm.Name),
						HostID:    host.ID,
						VMUUID:    vm.UUID,
						VMName:    vm.Name,
						Timestamp: time.Now().Add(-time.Duration(len(activities)*3) * time.Minute).Format(time.RFC3339),
						Severity:  "info",
						Details:   fmt.Sprintf("Virtual machine on %s", host.URI),
					}
					activities = append(activities, activity)
				}

				// Limit activities per host to avoid too many entries
				if len(activities) >= limit*2 {
					break
				}
			}
		}

		// Limit to avoid too many activities
		if len(activities) >= limit*2 {
			break
		}
	}

	// Sort by timestamp (newest first) and limit
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

func (m *HostMonitoringManager) pollHostStats(hostID string, sub *HostSubscription) {
	// Perform an immediate fetch so the UI receives data quickly.
	stats, err := m.service.connector.GetHostStats(hostID)
	if err != nil {
		log.Debugf("Error getting host stats for %s (initial): %v", hostID, err)
	} else {
		sub.mu.Lock()
		sub.lastKnownStats = stats
		sub.mu.Unlock()
		m.service.hub.BroadcastMessage(ws.Message{Type: "host-stats-updated", Payload: ws.MessagePayload{"hostId": hostID, "stats": stats}})
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := m.service.connector.GetHostStats(hostID)
			if err != nil {
				log.Debugf("Error getting host stats for %s: %v", hostID, err)
				// We don't stop polling here, as the host might just be temporarily unavailable
				continue
			}

			sub.mu.Lock()
			sub.lastKnownStats = stats
			sub.mu.Unlock()

			m.service.hub.BroadcastMessage(ws.Message{Type: "host-stats-updated", Payload: ws.MessagePayload{"hostId": hostID, "stats": stats}})

		case <-sub.stop:
			return
		}
	}
}
