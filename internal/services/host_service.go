package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	golibvirt "github.com/digitalocean/go-libvirt"
	"gorm.io/gorm"
)

// VMView is a combination of DB data and live libvirt data for the frontend.
type VMView struct {
	// From DB
	ID              uint   `json:"db_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	VCPUCount       uint   `json:"vcpu_count"`
	MemoryBytes     uint64 `json:"memory_bytes"`
	IsTemplate      bool   `json:"is_template"`
	CPUModel        string `json:"cpu_model"`
	CPUTopologyJSON string `json:"cpu_topology_json"`

	// From Libvirt or DB cache
	State    golibvirt.DomainState   `json:"state"`
	Graphics libvirt.GraphicsInfo    `json:"graphics"`
	Hardware *libvirt.HardwareInfo `json:"hardware,omitempty"` // Pointer to allow for null

	// From Libvirt (live data, only in some calls)
	MaxMem  uint64 `json:"max_mem"`
	Memory  uint64 `json:"memory"`
	CpuTime uint64 `json:"cpu_time"`
	Uptime  int64  `json:"uptime"`
}

// VmSubscription holds the clients subscribed to a VM's stats and a channel to stop polling.
type VmSubscription struct {
	clients      map[*ws.Client]bool
	stop         chan struct{}
	lastKnownStats *libvirt.VMStats
	mu           sync.RWMutex
}

// MonitoringManager handles real-time VM stat subscriptions.
type MonitoringManager struct {
	mu            sync.Mutex
	subscriptions map[string]*VmSubscription // key is "hostId:vmName"
	service       *HostService               // back-reference
}

// NewMonitoringManager creates a new manager.
func NewMonitoringManager(service *HostService) *MonitoringManager {
	return &MonitoringManager{
		subscriptions: make(map[string]*VmSubscription),
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
	GetVMHardwareAndTriggerSync(hostID, vmName string) (*libvirt.HardwareInfo, error)
	SyncVMsForHost(hostID string)
	StartVM(hostID, vmName string) error
	ShutdownVM(hostID, vmName string) error
	RebootVM(hostID, vmName string) error
	ForceOffVM(hostID, vmName string) error
	ForceResetVM(hostID, vmName string) error
}

type HostService struct {
	db        *gorm.DB
	connector *libvirt.Connector
	hub       *ws.Hub
	monitor   *MonitoringManager
}

func NewHostService(db *gorm.DB, connector *libvirt.Connector, hub *ws.Hub) *HostService {
	s := &HostService{
		db:        db,
		connector: connector,
		hub:       hub,
	}
	s.monitor = NewMonitoringManager(s)
	return s
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

func (s *HostService) AddHost(host storage.Host) (*storage.Host, error) {
	if err := s.db.Create(&host).Error; err != nil {
		return nil, fmt.Errorf("failed to save host to database: %w", err)
	}

	err := s.connector.AddHost(host)
	if err != nil {
		if delErr := s.db.Delete(&host).Error; delErr != nil {
			log.Printf("CRITICAL: Failed to rollback host creation for %s after connection failure. DB Error: %v", host.ID, delErr)
		}
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}

	// Initial sync after adding a host
	go s.SyncVMsForHost(host.ID)

	s.broadcastHostsChanged()
	return &host, nil
}

func (s *HostService) RemoveHost(hostID string) error {
	if err := s.connector.RemoveHost(hostID); err != nil {
		log.Printf("Warning: failed to disconnect from host %s during removal, continuing with DB deletion: %v", hostID, err)
	}

	if err := s.db.Where("host_id = ?", hostID).Delete(&storage.VirtualMachine{}).Error; err != nil {
		log.Printf("Warning: failed to delete VMs for host %s from database: %v", hostID, err)
	}

	if err := s.db.Where("id = ?", hostID).Delete(&storage.Host{}).Error; err != nil {
		return fmt.Errorf("failed to delete host from database: %w", err)
	}

	s.broadcastHostsChanged()
	return nil
}

func (s *HostService) ConnectToAllHosts() {
	hosts, err := s.GetAllHosts()
	if err != nil {
		log.Printf("Error retrieving hosts from database on startup: %v", err)
		return
	}

	for _, host := range hosts {
		log.Printf("Attempting to connect to stored host: %s", host.ID)
		if err := s.connector.AddHost(host); err != nil {
			log.Printf("Failed to connect to host %s (%s) on startup: %v", host.ID, host.URI, err)
		} else {
			go s.SyncVMsForHost(host.ID)
		}
	}
}

// --- VM Management ---

// GetVMsForHostFromDB retrieves the merged VM list for a host purely from the database,
// for a fast initial UI load.
func (s *HostService) GetVMsForHostFromDB(hostID string) ([]VMView, error) {
	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return nil, fmt.Errorf("could not get DB VM records for host %s: %w", hostID, err)
	}

	var vmViews []VMView
	for _, dbVM := range dbVMs {
		var graphics libvirt.GraphicsInfo
		if err := json.Unmarshal([]byte(dbVM.GraphicsJSON), &graphics); err != nil {
			log.Printf("Warning: could not parse cached graphics info for VM %s: %v", dbVM.Name, err)
		}

		vmViews = append(vmViews, VMView{
			ID:              dbVM.ID,
			Name:            dbVM.Name,
			Description:     dbVM.Description,
			VCPUCount:       dbVM.VCPUCount,
			MemoryBytes:     dbVM.MemoryBytes,
			IsTemplate:      dbVM.IsTemplate,
			CPUModel:        dbVM.CPUModel,
			CPUTopologyJSON: dbVM.CPUTopologyJSON,
			State:           golibvirt.DomainState(dbVM.State),
			Graphics:        graphics,
		})
	}
	return vmViews, nil
}

// getVMHardwareFromDB retrieves the hardware info from the JSON cache in the database.
func (s *HostService) getVMHardwareFromDB(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	var dbVM storage.VirtualMachine
	if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&dbVM).Error; err != nil {
		return nil, fmt.Errorf("could not find VM %s in database: %w", vmName, err)
	}

	if dbVM.HardwareJSON == "" {
		return nil, fmt.Errorf("no cached hardware info for VM %s", vmName)
	}

	var hardware libvirt.HardwareInfo
	if err := json.Unmarshal([]byte(dbVM.HardwareJSON), &hardware); err != nil {
		return nil, fmt.Errorf("could not parse cached hardware info for VM %s: %w", vmName, err)
	}

	return &hardware, nil
}

// GetVMHardwareAndTriggerSync serves cached hardware info from the DB immediately
// and triggers a background sync with libvirt.
func (s *HostService) GetVMHardwareAndTriggerSync(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	hardware, err := s.getVMHardwareFromDB(hostID, vmName)
	if err != nil {
		log.Printf("Could not get cached hardware for %s, will attempt live sync.", vmName)
	}

	go func() {
		if changed, syncErr := s.syncSingleVM(hostID, vmName); syncErr == nil && changed {
			s.broadcastVMsChanged(hostID)
		} else if syncErr != nil {
			log.Printf("Error during background hardware sync for %s: %v", vmName, syncErr)
		}
	}()

	return hardware, err
}

// SyncVMsForHost triggers a background sync with libvirt for a specific host's VMs.
// It sends a websocket message upon completion *only if* there were changes.
func (s *HostService) SyncVMsForHost(hostID string) {
	changed, err := s.syncAndListVMs(hostID)
	if err != nil {
		log.Printf("Error during background VM sync for host %s: %v", hostID, err)
		return
	}
	if changed {
		s.broadcastVMsChanged(hostID)
	}
}

// syncSingleVM syncs the state of a single VM from libvirt to the database.
func (s *HostService) syncSingleVM(hostID, vmName string) (bool, error) {
	vmInfo, err := s.connector.GetDomainInfo(hostID, vmName)
	if err != nil {
		// If the VM is not found, it might have been deleted. Check if it exists in the DB.
		var dbVM storage.VirtualMachine
		if err := s.db.Where("host_id = ? AND name = ?", hostID, vmName).First(&dbVM).Error; err == nil {
			// VM exists in DB but not on host, so delete it.
			if err := s.db.Delete(&dbVM).Error; err != nil {
				log.Printf("Warning: failed to prune old VM %s: %v", dbVM.Name, err)
				return false, err
			}
			return true, nil
		}
		return false, fmt.Errorf("could not fetch info for VM %s on host %s: %w", vmName, hostID, err)
	}

	graphicsBytes, err := json.Marshal(vmInfo.Graphics)
	if err != nil {
		log.Printf("Warning: could not marshal graphics info for VM %s: %v", vmInfo.Name, err)
		graphicsBytes = []byte("{}")
	}

	hardwareInfo, err := s.connector.GetDomainHardware(hostID, vmName)
	if err != nil {
		log.Printf("Warning: could not fetch hardware for VM %s: %v", vmInfo.Name, err)
	}
	hardwareBytes, err := json.Marshal(hardwareInfo)
	if err != nil {
		log.Printf("Warning: could not marshal hardware info for VM %s: %v", vmInfo.Name, err)
		hardwareBytes = []byte("{}")
	}

	vmRecord := storage.VirtualMachine{
		HostID:       hostID,
		Name:         vmInfo.Name,
		UUID:         vmInfo.UUID,
		State:        int(vmInfo.State),
		VCPUCount:    vmInfo.Vcpu,
		MemoryBytes:  vmInfo.MaxMem * 1024,
		GraphicsJSON: string(graphicsBytes),
		HardwareJSON: string(hardwareBytes),
	}

	var existingVM storage.VirtualMachine
	if err := s.db.Where("uuid = ?", vmInfo.UUID).First(&existingVM).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := s.db.Create(&vmRecord).Error; err != nil {
				log.Printf("Warning: could not create VM %s in database: %v", vmInfo.Name, err)
				return false, err
			}
			return true, nil
		}
		return false, err
	}

	if existingVM.Name != vmRecord.Name ||
		existingVM.State != vmRecord.State ||
		existingVM.VCPUCount != vmRecord.VCPUCount ||
		existingVM.MemoryBytes != vmRecord.MemoryBytes ||
		existingVM.GraphicsJSON != vmRecord.GraphicsJSON ||
		existingVM.HardwareJSON != vmRecord.HardwareJSON {

		if err := s.db.Model(&existingVM).Updates(vmRecord).Error; err != nil {
			log.Printf("Warning: could not update VM %s in database: %v", vmInfo.Name, err)
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// syncAndListVMs is the core function to get VMs from libvirt and sync with the local DB.
// It returns true if any data was changed in the database.
func (s *HostService) syncAndListVMs(hostID string) (bool, error) {
	liveVMs, err := s.connector.ListAllDomains(hostID)
	if err != nil {
		return false, fmt.Errorf("service failed to list vms for host %s: %w", hostID, err)
	}

	var changed bool

	var dbVMs []storage.VirtualMachine
	if err := s.db.Where("host_id = ?", hostID).Find(&dbVMs).Error; err != nil {
		return false, fmt.Errorf("could not get DB records for comparison: %w", err)
	}
	dbVMMap := make(map[string]storage.VirtualMachine)
	for _, vm := range dbVMs {
		dbVMMap[vm.UUID] = vm
	}

	for _, vmInfo := range liveVMs {
		graphicsBytes, err := json.Marshal(vmInfo.Graphics)
		if err != nil {
			log.Printf("Warning: could not marshal graphics info for VM %s: %v", vmInfo.Name, err)
			graphicsBytes = []byte("{}")
		}

		hardwareInfo, err := s.connector.GetDomainHardware(hostID, vmInfo.Name)
		if err != nil {
			log.Printf("Warning: could not fetch hardware for VM %s: %v", vmInfo.Name, err)
		}
		hardwareBytes, err := json.Marshal(hardwareInfo)
		if err != nil {
			log.Printf("Warning: could not marshal hardware info for VM %s: %v", vmInfo.Name, err)
			hardwareBytes = []byte("{}")
		}

		vmRecord := storage.VirtualMachine{
			HostID:       hostID,
			Name:         vmInfo.Name,
			UUID:         vmInfo.UUID,
			State:        int(vmInfo.State),
			VCPUCount:    vmInfo.Vcpu,
			MemoryBytes:  vmInfo.MaxMem * 1024,
			GraphicsJSON: string(graphicsBytes),
			HardwareJSON: string(hardwareBytes),
		}

		existingVM, exists := dbVMMap[vmInfo.UUID]
		if !exists {
			if err := s.db.Create(&vmRecord).Error; err != nil {
				log.Printf("Warning: could not create VM %s in database: %v", vmInfo.Name, err)
			} else {
				changed = true
			}
		} else {
			if existingVM.Name != vmRecord.Name ||
				existingVM.State != vmRecord.State ||
				existingVM.VCPUCount != vmRecord.VCPUCount ||
				existingVM.MemoryBytes != vmRecord.MemoryBytes ||
				existingVM.GraphicsJSON != vmRecord.GraphicsJSON ||
				existingVM.HardwareJSON != vmRecord.HardwareJSON {

				if err := s.db.Model(&existingVM).Updates(vmRecord).Error; err != nil {
					log.Printf("Warning: could not update VM %s in database: %v", vmInfo.Name, err)
				} else {
					changed = true
				}
			}
		}
	}

	liveVMUUIDs := make(map[string]struct{})
	for _, vm := range liveVMs {
		liveVMUUIDs[vm.UUID] = struct{}{}
	}
	for _, dbVM := range dbVMs {
		if _, ok := liveVMUUIDs[dbVM.UUID]; !ok {
			if err := s.db.Delete(&dbVM).Error; err != nil {
				log.Printf("Warning: failed to prune old VM %s: %v", dbVM.Name, err)
			} else {
				changed = true
			}
		}
	}

	return changed, nil
}

func (s *HostService) GetVMStats(hostID, vmName string) (*libvirt.VMStats, error) {
	// First, check if there's an active subscription.
	stats := s.monitor.GetLastKnownStats(hostID, vmName)
	if stats != nil {
		return stats, nil
	}

	// If no active subscription, perform a one-time fetch.
	return s.connector.GetDomainStats(hostID, vmName)
}

// --- VM Actions ---

func (s *HostService) StartVM(hostID, vmName string) error {
	if err := s.connector.StartDomain(hostID, vmName); err != nil {
		return err
	}
	if changed, err := s.syncSingleVM(hostID, vmName); err == nil && changed {
		s.broadcastVMsChanged(hostID)
	}
	return nil
}

func (s *HostService) ShutdownVM(hostID, vmName string) error {
	if err := s.connector.ShutdownDomain(hostID, vmName); err != nil {
		return err
	}
	if changed, err := s.syncSingleVM(hostID, vmName); err == nil && changed {
		s.broadcastVMsChanged(hostID)
	}
	return nil
}

func (s *HostService) RebootVM(hostID, vmName string) error {
	if err := s.connector.RebootDomain(hostID, vmName); err != nil {
		return err
	}
	if changed, err := s.syncSingleVM(hostID, vmName); err == nil && changed {
		s.broadcastVMsChanged(hostID)
	}
	return nil
}

func (s *HostService) ForceOffVM(hostID, vmName string) error {
	if err := s.connector.DestroyDomain(hostID, vmName); err != nil {
		return err
	}
	if changed, err := s.syncSingleVM(hostID, vmName); err == nil && changed {
		s.broadcastVMsChanged(hostID)
	}
	return nil
}

func (s *HostService) ForceResetVM(hostID, vmName string) error {
	if err := s.connector.ResetDomain(hostID, vmName); err != nil {
		return err
	}
	if changed, err := s.syncSingleVM(hostID, vmName); err == nil && changed {
		s.broadcastVMsChanged(hostID)
	}
	return nil
}

// --- WebSocket Message Handling ---

func (s *HostService) HandleSubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok1 := payload["hostId"].(string)
	vmName, ok2 := payload["vmName"].(string)
	if !ok1 || !ok2 {
		log.Println("Invalid payload for vm-stats subscription")
		return
	}
	s.monitor.Subscribe(client, hostID, vmName)
}

func (s *HostService) HandleUnsubscribe(client *ws.Client, payload ws.MessagePayload) {
	hostID, ok1 := payload["hostId"].(string)
	vmName, ok2 := payload["vmName"].(string)
	if !ok1 || !ok2 {
		log.Println("Invalid payload for vm-stats unsubscription")
		return
	}
	s.monitor.Unsubscribe(client, hostID, vmName)
}

func (s *HostService) HandleClientDisconnect(client *ws.Client) {
	s.monitor.UnsubscribeClient(client)
}

// --- Monitoring Goroutine Logic ---

func (m *MonitoringManager) Subscribe(client *ws.Client, hostID, vmName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	sub, exists := m.subscriptions[key]
	if !exists {
		log.Printf("Starting monitoring for %s", key)
		sub = &VmSubscription{
			clients: make(map[*ws.Client]bool),
			stop:    make(chan struct{}),
		}
		m.subscriptions[key] = sub
		go m.pollVmStats(hostID, vmName, sub)
	}
	sub.clients[client] = true
}

func (m *MonitoringManager) Unsubscribe(client *ws.Client, hostID, vmName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", hostID, vmName)
	if sub, exists := m.subscriptions[key]; exists {
		delete(sub.clients, client)
		if len(sub.clients) == 0 {
			log.Printf("Stopping monitoring for %s", key)
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
				log.Printf("Stopping monitoring for %s due to client disconnect", key)
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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := m.service.connector.GetDomainStats(hostID, vmName)
			if err != nil {
				stats = &libvirt.VMStats{State: golibvirt.DomainShutoff}
			}

			// Update last known stats
			sub.mu.Lock()
			sub.lastKnownStats = stats
			sub.mu.Unlock()

			// Broadcast the stats update.
			m.service.hub.BroadcastMessage(ws.Message{
				Type: "vm-stats-updated",
				Payload: ws.MessagePayload{
					"hostId": hostID,
					"vmName": vmName,
					"stats":  stats,
				},
			})

			// If the VM is no longer running, stop polling it.
			if stats.State != golibvirt.DomainRunning {
				log.Printf("VM %s is not running, stopping stats polling.", vmName)
				// Unsubscribe all clients for this VM
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


