package services

import (
	"encoding/json"
	"fmt"
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
	TaskState       storage.VMTaskState `json:"task_state"`

	// NEW: Drift detection fields
	SyncStatus   storage.SyncStatus `json:"sync_status"`
	DriftDetails string             `json:"drift_details"`
	NeedsRebuild bool               `json:"needs_rebuild"`

	// From Libvirt or DB cache
	State    storage.VMState       `json:"state"`
	Graphics libvirt.GraphicsInfo  `json:"graphics"`
	Hardware *libvirt.HardwareInfo `json:"hardware,omitempty"` // Pointer to allow for null

	// From Libvirt (live data, only in some calls)
	MaxMem  uint64 `json:"max_mem"`
	Memory  uint64 `json:"memory"`
	CpuTime uint64 `json:"cpu_time"`
	Uptime  int64  `json:"uptime"`
}

// PortAttachmentView is a transport-friendly view of a PortAttachment including
// the underlying Port and (optional) Network information for the UI.
type PortAttachmentView struct {
	ID         uint             `json:"id"`
	VMUUID     string           `json:"vm_uuid"`
	DeviceName string           `json:"device_name"`
	MACAddress string           `json:"mac_address"`
	ModelName  string           `json:"model_name"`
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
	GetHostInfo(hostID string) (*libvirt.HostInfo, error)
	AddHost(host storage.Host) (*storage.Host, error)
	RemoveHost(hostID string) error
	ConnectToAllHosts()
	GetVMsForHostFromDB(hostID string) ([]VMView, error)
	GetVMStats(hostID, vmName string) (*libvirt.VMStats, error)
	GetVMHardwareAndDetectDrift(hostID, vmName string) (*libvirt.HardwareInfo, error)
	GetPortsForHostFromDB(hostID string) ([]storage.Port, error)
	GetPortAttachmentsForVM(vmUUID string) ([]PortAttachmentView, error)
	SyncVMsForHost(hostID string)
	SyncVMFromLibvirt(hostID, vmName string) error
	RebuildVMFromDB(hostID, vmName string) error
	StartVM(hostID, vmName string) error
	ShutdownVM(hostID, vmName string) error
	RebootVM(hostID, vmName string) error
	ForceOffVM(hostID, vmName string) error
	ForceResetVM(hostID, vmName string) error
}

type HostService struct {
	db          *gorm.DB
	connector   *libvirt.Connector
	hub         *ws.Hub
	monitor     *MonitoringManager
	hostMonitor *HostMonitoringManager
}

func NewHostService(db *gorm.DB, connector *libvirt.Connector, hub *ws.Hub) *HostService {
	s := &HostService{
		db:        db,
		connector: connector,
		hub:       hub,
	}
	s.monitor = NewMonitoringManager(s)
	s.hostMonitor = NewHostMonitoringManager(s)
	return s
}

// EnsureHostConnected ensures there's an active libvirt connection for the
// given host ID. If no connection exists, it will attempt to read the host
// URI from the database and connect. This allows lazy connection on demand
// (e.g., when the UI first subscribes to stats) instead of connecting all
// hosts at startup.
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
	return nil
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

// --- Host Management ---

func (s *HostService) GetAllHosts() ([]storage.Host, error) {
	var hosts []storage.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

func (s *HostService) GetHostInfo(hostID string) (*libvirt.HostInfo, error) {
	return s.connector.GetHostInfo(hostID)
}

// GetPortsForHostFromDB returns ports scoped to a host that are not currently attached to any VM.
func (s *HostService) GetPortsForHostFromDB(hostID string) ([]storage.Port, error) {
	var ports []storage.Port
	// subquery for attached port ids
	sub := s.db.Model(&storage.PortAttachment{}).Select("port_id")
	if err := s.db.Where("host_id = ? AND id NOT IN (?)", hostID, sub).Find(&ports).Error; err != nil {
		return nil, err
	}
	return ports, nil
}

// GetPortAttachmentsForVM returns structured port attachments for a VM UUID.
func (s *HostService) GetPortAttachmentsForVM(vmUUID string) ([]PortAttachmentView, error) {
	var atts []storage.PortAttachment
	if err := s.db.Preload("Port").Where("vm_uuid = ?", vmUUID).Find(&atts).Error; err != nil {
		return nil, err
	}
	var out []PortAttachmentView
	for _, a := range atts {
		var pb storage.PortBinding
		var net *storage.Network
		if err := s.db.Preload("Network").Where("port_id = ?", a.PortID).First(&pb).Error; err == nil {
			net = &pb.Network
		}
		out = append(out, PortAttachmentView{
			ID:         a.ID,
			VMUUID:     a.VMUUID,
			DeviceName: a.DeviceName,
			MACAddress: a.MACAddress,
			ModelName:  a.ModelName,
			Ordinal:    a.Ordinal,
			Metadata:   a.Metadata,
			Port:       a.Port,
			Network:    net,
		})
	}
	return out, nil
}

func (s *HostService) AddHost(host storage.Host) (*storage.Host, error) {
	if err := s.db.Create(&host).Error; err != nil {
		return nil, fmt.Errorf("failed to save host to database: %w", err)
	}
	// Mark host as connecting in DB
	s.db.Model(&storage.Host{}).Where("id = ?", host.ID).Updates(map[string]interface{}{"task_state": storage.HostTaskStateConnecting})
	log.Verbosef("AddHost: saved host %s (uri=%s)", host.ID, host.URI)

	err := s.connector.AddHost(host)
	if err != nil {
		if delErr := s.db.Delete(&host).Error; delErr != nil {
			log.Debugf("CRITICAL: Failed to rollback host creation for %s after connection failure. DB Error: %v", host.ID, delErr)
		}
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}

	// Initial sync after adding a host
	go s.SyncVMsForHost(host.ID)

	log.Verbosef("AddHost: starting initial sync for host %s", host.ID)
	// Mark connected in DB and clear task state
	s.db.Model(&storage.Host{}).Where("id = ?", host.ID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateConnected})
	// Notify clients both that hosts changed and that this host is connected
	s.broadcastHostConnectionChanged(host.ID, true)
	s.broadcastHostsChanged()
	return &host, nil
}

func (s *HostService) RemoveHost(hostID string) error {
	// Mark as disconnecting
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": storage.HostTaskStateDisconnecting})
	if err := s.connector.RemoveHost(hostID); err != nil {
		log.Verbosef("RemoveHost: failed to disconnect from host %s during removal: %v", hostID, err)
		// If removal failed, mark error
		s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateError})
	}

	// Remove VMs and their attachment indices transactionally
	tx := s.db.Begin()
	var vms []storage.VirtualMachine
	if err := tx.Where("host_id = ?", hostID).Find(&vms).Error; err != nil {
		tx.Rollback()
		log.Verbosef("Warning: failed to query VMs for host %s: %v", hostID, err)
	} else {
		for _, vm := range vms {
			if err := tx.Where("vm_uuid = ?", vm.UUID).Delete(&storage.AttachmentIndex{}).Error; err != nil {
				tx.Rollback()
				log.Verbosef("Warning: failed to delete attachment indices for VM %s: %v", vm.UUID, err)
				break
			}
		}
		// delete VMs
		if err := tx.Where("host_id = ?", hostID).Delete(&storage.VirtualMachine{}).Error; err != nil {
			tx.Rollback()
			log.Verbosef("Warning: failed to delete VMs for host %s from database: %v", hostID, err)
		} else {
			if err := tx.Commit().Error; err != nil {
				log.Verbosef("Warning: failed to commit VM deletion transaction for host %s: %v", hostID, err)
			}
		}
	}

	if err := s.db.Where("id = ?", hostID).Delete(&storage.Host{}).Error; err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}

	// Mark disconnected and clear task state
	s.db.Model(&storage.Host{}).Where("id = ?", hostID).Updates(map[string]interface{}{"task_state": "", "state": storage.HostStateDisconnected})
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

// --- VM Management ---
func (s *HostService) GetVMsForHostFromDB(hostID string) ([]VMView, error) {
	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return nil, fmt.Errorf("could not get DB VM records for host %s: %w", hostID, err)
	}

	var vmViews []VMView
	for _, dbVM := range dbVMs {
		var graphics libvirt.GraphicsInfo // Default to false

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
			State:           dbVM.State,
			TaskState:       dbVM.TaskState,
			Graphics:        graphics,
			SyncStatus:      dbVM.SyncStatus,
			DriftDetails:    dbVM.DriftDetails,
			NeedsRebuild:    dbVM.NeedsRebuild,
		})
	}
	return vmViews, nil
}

func (s *HostService) getVMHardwareFromDB(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	var vm storage.VirtualMachine
	if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&vm).Error; err != nil {
		return nil, fmt.Errorf("could not find VM %s in database: %w", vmName, err)
	}

	var hardware libvirt.HardwareInfo

	// Retrieve and populate disks
	var diskAttachments []storage.VolumeAttachment
	s.db.Preload("Volume").Where("vm_uuid = ?", vm.UUID).Find(&diskAttachments)
	for _, da := range diskAttachments {
		hardware.Disks = append(hardware.Disks, libvirt.DiskInfo{
			Device: da.DeviceName,
			Path:   da.Volume.Name,
			Target: struct {
				Dev string `xml:"dev,attr" json:"dev"`
				Bus string `xml:"bus,attr" json:"bus"`
			}{
				Dev: da.DeviceName,
				Bus: da.BusType,
			},
			Driver: struct {
				Name string `xml:"name,attr" json:"driver_name"`
				Type string `xml:"type,attr" json:"type"`
			}{
				Type: da.Volume.Format,
			},
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
				devName := a.DeviceName
				if devName == "" {
					devName = a.Port.DeviceName
				}
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
						Bridge string `xml:"bridge,attr" json:"bridge"`
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

	return &hardware, nil
}
func (s *HostService) GetVMHardwareAndDetectDrift(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	if changed, syncErr := s.detectDriftOrIngestVM(hostID, vmName, false); syncErr != nil {
		log.Verbosef("Error during hardware sync for %s: %v", vmName, syncErr)
	} else if changed {
		s.broadcastVMsChanged(hostID)
	}

	return s.getVMHardwareFromDB(hostID, vmName)
}

func (s *HostService) SyncVMsForHost(hostID string) {
	changed, err := s.syncHostVMs(hostID)
	if err != nil {
		log.Verbosef("Error during background VM sync for host %s: %v", hostID, err)
		return
	}
	if changed {
		s.broadcastVMsChanged(hostID)
	}
}

func (s *HostService) detectDriftOrIngestVM(hostID, vmName string, isInitialSync bool) (bool, error) {
	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		// If we can't get info from libvirt, check if it's a stale DB entry
		var dbVM storage.VirtualMachine
		if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&dbVM).Error; err == nil {
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
	var changed bool
	err = tx.Where("host_id = ? AND domain_uuid = ?", hostID, vmInfo.UUID).First(&existingVM).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return false, err
	}

	// --- Case 1: New VM found, perform initial ingestion ---
	if err == gorm.ErrRecordNotFound {
		log.Infof("New VM '%s' detected on host '%s'. Performing initial ingestion.", vmName, hostID)

		// Check if this VM's UUID exists on a *different* host. This is a conflict.
		var conflictingVM storage.VirtualMachine
		err := tx.Where("domain_uuid = ? AND host_id != ?", vmInfo.UUID, hostID).First(&conflictingVM).Error

		// This is not an error, just informational. It means the VM is new to this host.
		// We log it at a debug/info level instead of treating it as a query failure.
		if err == nil {
			log.Infof("VM with DomainUUID %s was previously on host %s, now detected on %s; treating as new on this host.", vmInfo.UUID, conflictingVM.HostID, hostID)
		} else if err != gorm.ErrRecordNotFound {
			// An actual database error occurred.
			tx.Rollback()
			return false, fmt.Errorf("error checking for conflicting VM: %w", err)
		}

		newVMRecord := storage.VirtualMachine{
			HostID:      hostID,
			Name:        vmInfo.Name,
			DomainUUID:  vmInfo.UUID,
			State:       mapLibvirtStateToVMState(vmInfo.State),
			VCPUCount:   vmInfo.Vcpu,
			MemoryBytes: vmInfo.MaxMem * 1024,
			SyncStatus:  storage.StatusSynced, // New VMs are synced by definition
		}

		// If no conflict was found (err was gorm.ErrRecordNotFound), we can safely use the DomainUUID as our primary UUID.
		// Otherwise, we generate a new internal UUID to avoid primary key collision.
		if err == gorm.ErrRecordNotFound {
			newVMRecord.UUID = vmInfo.UUID
		} else {
			log.Infof("UUID conflict detected for DomainUUID %s; assigning new internal UUID.", vmInfo.UUID)
			newVMRecord.UUID = uuid.New().String()
		}

		if err := tx.Create(&newVMRecord).Error; err != nil {
			// If insertion failed due to a name conflict on (host_id, name), attempt
			// to disambiguate the VM name by appending a short suffix and retry.
			if strings.Contains(err.Error(), "UNIQUE constraint failed: virtual_machines.host_id, virtual_machines.name") {
				origName := newVMRecord.Name
				newVMRecord.Name = fmt.Sprintf("%s-%s", origName, uuid.New().String()[:8])
				log.Verbosef("Name conflict for VM '%s' on host %s — retrying insert as '%s'", origName, hostID, newVMRecord.Name)
				if ierr := tx.Create(&newVMRecord).Error; ierr != nil {
					tx.Rollback()
					return false, fmt.Errorf("failed to create VM record after disambiguating name: %w", ierr)
				}
			} else {
				tx.Rollback()
				return false, err
			}
		}
		changed = true
		existingVM = newVMRecord

		// Also ingest hardware on initial sync
		hardwareInfo, hwErr := s.connector.GetDomainHardware(hostID, vmName)
		if hwErr != nil {
			log.Verbosef("Warning: could not fetch hardware for new VM %s: %v", vmInfo.Name, hwErr)
		} else {
			if _, err := s.syncVMHardware(tx, existingVM.UUID, hostID, hardwareInfo, &vmInfo.Graphics); err != nil {
				tx.Rollback()
				return false, fmt.Errorf("failed to sync hardware for new VM: %w", err)
			}
		}
	} else { // --- Case 2: Existing VM, perform drift detection ---
		updates := make(map[string]interface{})
		driftDetails := make(map[string]map[string]interface{})

		// Always update volatile state
		newState := mapLibvirtStateToVMState(vmInfo.State)
		if existingVM.State != newState {
			updates["state"] = newState
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

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return changed, nil
}

// syncVMHardware intelligently syncs hardware state, only performing writes when necessary.
func (s *HostService) syncVMHardware(tx *gorm.DB, vmUUID string, hostID string, hardware *libvirt.HardwareInfo, graphics *libvirt.GraphicsInfo) (bool, error) {
	var changed bool = false
	if hardware != nil {
		log.Debugf("syncVMHardware: vm=%s host=%s start devices: disks=%d networks=%d", vmUUID, hostID, len(hardware.Disks), len(hardware.Networks))
	} else {
		log.Debugf("syncVMHardware: vm=%s host=%s start (no hardware)", vmUUID, hostID)
	}
	defer log.Debugf("syncVMHardware: vm=%s host=%s finished", vmUUID, hostID)

	// --- Sync Networks / Ports ---
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

	for _, net := range hardware.Networks {
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
				MACAddress: net.Mac.Address, DeviceName: net.Target.Dev, ModelName: net.Model.Type,
				HostID: hostID,
			}
			if err := tx.Create(&newPort).Error; err != nil {
				return false, err
			}

			if network.ID != 0 && newPort.ID != 0 {
				binding := storage.PortBinding{PortID: newPort.ID, NetworkID: network.ID}
				tx.Create(&binding)
			}

			// Create or reuse an attachment row linking the port to the VM.
			var existingAtt storage.PortAttachment
			attErr := tx.Where("vm_uuid = ? AND device_name = ?", vmUUID, net.Target.Dev).First(&existingAtt).Error
			if attErr == nil {
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
				}
			} else if attErr == gorm.ErrRecordNotFound {
				// Create new attachment
				newAttachment := storage.PortAttachment{VMUUID: vmUUID, PortID: newPort.ID, DeviceName: net.Target.Dev, MACAddress: net.Mac.Address, ModelName: net.Model.Type}
				if err := tx.Create(&newAttachment).Error; err != nil {
					return false, err
				}
				existingAtt = newAttachment
			} else {
				return false, attErr
			}

			// Ensure the AttachmentIndex references this attachment/device.
			// Check for an existing index by (device_type, attachment_id) first to avoid
			// violating the unique index on (device_type, attachment_id). If an index
			// by attachment_id exists, ensure it matches this VM. Otherwise, check for
			// an existing index by (device_type, device_id) which would indicate the
			// device is already allocated to some attachment.
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "port", AttachmentID: existingAtt.ID, DeviceID: existingAtt.PortID}
			var existingAlloc storage.AttachmentIndex
			// First, try to find by attachment_id (this is the unique index that failed earlier)
			aerr := tx.Where("device_type = ? AND attachment_id = ?", alloc.DeviceType, alloc.AttachmentID).First(&existingAlloc).Error
			if aerr == nil {
				// An index already exists for this attachment. Ensure it belongs to the same VM.
				if existingAlloc.VMUUID != alloc.VMUUID {
					return false, fmt.Errorf("attachment (id=%d) already indexed for VM %s (index id %d)", alloc.AttachmentID, existingAlloc.VMUUID, existingAlloc.ID)
				}
				// Already correctly indexed; nothing to do.
			} else if aerr == gorm.ErrRecordNotFound {
				// No index by attachment_id exists; ensure device_id isn't already allocated
				derr := tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).First(&existingAlloc).Error
				if derr == nil {
					// Device is already allocated to some attachment_id; conflict unless it's the same attachment
					if existingAlloc.AttachmentID != alloc.AttachmentID || existingAlloc.VMUUID != alloc.VMUUID {
						return false, fmt.Errorf("device (type=%s id=%d) already allocated to VM %s (attachment_index id %d)", alloc.DeviceType, alloc.DeviceID, existingAlloc.VMUUID, existingAlloc.ID)
					}
					// otherwise matches — nothing to do
				} else if derr == gorm.ErrRecordNotFound {
					// Safe to create a new allocation
					if err := tx.Create(&alloc).Error; err != nil {
						return false, err
					}
				} else {
					return false, derr
				}
			} else {
				return false, aerr
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

	// --- Sync Disks (Intelligent Update) ---
	var existingAttachments []storage.VolumeAttachment
	tx.Preload("Volume").Where("vm_uuid = ?", vmUUID).Find(&existingAttachments)
	existingAttachmentsMap := make(map[string]storage.VolumeAttachment)
	for _, da := range existingAttachments {
		existingAttachmentsMap[da.DeviceName] = da
	}

	for _, disk := range hardware.Disks {
		var volume storage.Volume
		tx.FirstOrCreate(&volume, storage.Volume{Name: disk.Path}, storage.Volume{
			Name: disk.Path, Format: disk.Driver.Type, Type: "DISK",
		})

		attachment, exists := existingAttachmentsMap[disk.Target.Dev]
		if exists {
			updates := make(map[string]interface{})
			if attachment.VolumeID != volume.ID {
				updates["volume_id"] = volume.ID
			}
			if attachment.BusType != disk.Target.Bus {
				updates["bus_type"] = disk.Target.Bus
			}
			if len(updates) > 0 {
				if err := tx.Model(&attachment).Updates(updates).Error; err != nil {
					return false, err
				}
				// If the volume_id changed, keep the attachment index in sync
				// No attachment_index device_id update for volumes: volumes are multi-attach.
				changed = true
			}
			delete(existingAttachmentsMap, disk.Target.Dev)
		} else {
			newAttachment := storage.VolumeAttachment{
				VMUUID: vmUUID, VolumeID: volume.ID, DeviceName: disk.Target.Dev, BusType: disk.Target.Bus,
			}
			if err := tx.Create(&newAttachment).Error; err != nil {
				return false, err
			}
			// Insert corresponding attachment index in the same transaction
			// Volumes can be multi-attached (e.g., ISOs or multi-attach volumes). To support that,
			// we do not enforce uniqueness by device_id for device_type == "volume". We still
			// record the attachment in the index for fast VM-scoped queries, but store DeviceID=0
			// to avoid conflicts with the unique (device_type, device_id) index.
			alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "volume", AttachmentID: newAttachment.ID, DeviceID: 0}
			if err := tx.Create(&alloc).Error; err != nil {
				return false, err
			}
			changed = true
		}
	}

	if len(existingAttachmentsMap) > 0 {
		var idsToDelete []uint
		for _, attachment := range existingAttachmentsMap {
			idsToDelete = append(idsToDelete, attachment.ID)
		}
		// Remove index entries first, then remove attachment rows within the same tx
		if err := tx.Where("device_type = ? AND attachment_id IN ?", "volume", idsToDelete).Delete(&storage.AttachmentIndex{}).Error; err != nil {
			return false, err
		}
		if err := tx.Where("id IN ?", idsToDelete).Delete(&storage.VolumeAttachment{}).Error; err != nil {
			return false, err
		}
		changed = true
	}

	// --- Sync Graphics (Console model) ---
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
		if err := tx.Where("vm_uuid = ? AND type = ?", vmUUID, desiredGfxType).First(&console).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				console = storage.Console{VMUUID: vmUUID, HostID: hostID, Type: desiredGfxType, ModelName: desiredGfxType}
				if err := tx.Create(&console).Error; err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}

		// Ensure attachment index references the console instance
		alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: "console", AttachmentID: console.ID, DeviceID: console.ID}
		var existingAlloc storage.AttachmentIndex
		res := tx.Where("device_type = ? AND device_id = ?", alloc.DeviceType, alloc.DeviceID).First(&existingAlloc)
		if res.Error == nil {
			if existingAlloc.AttachmentID != alloc.AttachmentID || existingAlloc.VMUUID != alloc.VMUUID {
				return false, fmt.Errorf("console (id=%d) already allocated to VM %s (attachment_index id %d)", alloc.DeviceID, existingAlloc.VMUUID, existingAlloc.ID)
			}
			// allocation already present and matching
		} else if res.Error != gorm.ErrRecordNotFound {
			return false, res.Error
		} else {
			if err := tx.Create(&alloc).Error; err != nil {
				return false, err
			}
		}
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
		changed, err := s.detectDriftOrIngestVM(hostID, vmInfo.Name, true)
		if err != nil {
			log.Verbosef("Error syncing VM %s: %v", vmInfo.Name, err)
		}
		if changed {
			overallChanged = true
		}
	}

	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return false, fmt.Errorf("could not get DB records for pruning check: %w", err)
	}

	for _, dbVM := range dbVMs {
		if _, exists := liveVMUUIDs[dbVM.DomainUUID]; !exists {
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
			if err := tx.Commit().Error; err != nil {
				log.Verbosef("Warning: failed to commit prune transaction for VM %s: %v", dbVM.Name, err)
				continue
			}
			overallChanged = true
		}
	}

	return overallChanged, nil
}

func (s *HostService) GetVMStats(hostID, vmName string) (*libvirt.VMStats, error) {
	stats := s.monitor.GetLastKnownStats(hostID, vmName)
	if stats != nil {
		return stats, nil
	}
	return s.connector.GetDomainStats(hostID, vmName)
}

// --- VM Actions ---

func (s *HostService) performVMAction(hostID, vmName string, taskState storage.VMTaskState, action func() error) error {
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
	})
}

func (s *HostService) ShutdownVM(hostID, vmName string) error {
	return s.performVMAction(hostID, vmName, storage.TaskStateStopping, func() error {
		return s.connector.ShutdownDomain(hostID, vmName)
	})
}

func (s *HostService) RebootVM(hostID, vmName string) error {
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
	})
}

func (s *HostService) ForceResetVM(hostID, vmName string) error {
	return s.performVMAction(hostID, vmName, storage.TaskStateRebooting, func() error {
		// If a rebuild is needed, this power cycle will apply the changes.
		// So, we can clear the flag.
		s.db.Model(&storage.VirtualMachine{}).Where("host_id = ? AND name = ?", hostID, vmName).Update("needs_rebuild", false)
		return s.connector.ResetDomain(hostID, vmName)
	})
}

// --- Drift and Sync Actions ---

// SyncVMFromLibvirt forces an update from the live libvirt state into the database,
// overwriting the DB record and clearing any drift status.
func (s *HostService) SyncVMFromLibvirt(hostID, vmName string) error {
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
		"SyncStatus":   storage.StatusSynced,
		"DriftDetails": "",
		"NeedsRebuild": false,
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
	// Ensure the host connection exists before starting monitoring. Do this
	// without holding the monitoring mutex because EnsureHostConnected may
	// attempt network I/O and we don't want to block other subscribe/unsubscribe operations.
	if err := m.service.EnsureHostConnected(hostID); err != nil {
		log.Debugf("Warning: could not ensure host %s connected: %v", hostID, err)
		// Continue: we'll still start monitoring which will report errors
		// when attempting to fetch stats if the host is unavailable.
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	sub, exists := m.subscriptions[key]
	if !exists {
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
		if err := client.SendMessage(ws.Message{Type: "vm-stats-updated", Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": cached}}); err != nil {
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
	m.service.hub.BroadcastMessage(ws.Message{
		Type:    "vm-stats-updated",
		Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": stats},
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

			m.service.hub.BroadcastMessage(ws.Message{
				Type:    "vm-stats-updated",
				Payload: ws.MessagePayload{"hostId": hostID, "vmName": vmName, "stats": stats},
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
	if err := m.service.EnsureHostConnected(hostID); err != nil {
		log.Printf("Warning: could not ensure host %s connected: %v", hostID, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sub, exists := m.subscriptions[hostID]
	if !exists {
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
