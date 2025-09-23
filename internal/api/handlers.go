package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/capsali/virtumancer/internal/logging"

	"github.com/capsali/virtumancer/internal/console"
	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIHandler struct {
	HostService services.HostServiceProvider
	Hub         *ws.Hub
	DB          *gorm.DB
	Connector   *libvirt.Connector
}

func NewAPIHandler(hostService services.HostServiceProvider, hub *ws.Hub, db *gorm.DB, connector *libvirt.Connector) *APIHandler {
	return &APIHandler{
		HostService: hostService,
		Hub:         hub,
		DB:          db,
		Connector:   connector,
	}
}

func (h *APIHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(h.Hub, h.HostService, w, r)
}

func (h *APIHandler) HandleVMConsole(w http.ResponseWriter, r *http.Request) {
	console.HandleConsole(h.DB, h.Connector, w, r)
}

func (h *APIHandler) HandleSpiceConsole(w http.ResponseWriter, r *http.Request) {
	console.HandleSpiceConsole(h.DB, h.Connector, w, r)
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *APIHandler) CreateHost(w http.ResponseWriter, r *http.Request) {
	var host storage.Host
	if err := json.NewDecoder(r.Body).Decode(&host); err != nil {
		apiErr := NewAPIError(ErrorCodeValidation, "Invalid request body", "Failed to parse JSON request")
		WriteError(w, apiErr, http.StatusBadRequest)
		return
	}

	// Validate required fields - only URI is required, ID will be generated if not provided
	if host.URI == "" {
		apiErr := NewAPIError(ErrorCodeValidation, "Missing required fields", "Host URI is required")
		WriteError(w, apiErr, http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if host.ID == "" {
		host.ID = uuid.New().String()
	}

	newHost, err := h.HostService.AddHost(host)
	if err != nil {
		h.HandleError(w, err, "create_host")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newHost)
}

func (h *APIHandler) GetHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.HostService.GetAllHosts()
	if err != nil {
		h.HandleError(w, err, "fetch_all_hosts")
		return
	}

	// Augment host list with a connection flag so the UI can render connection status.
	type hostWithStatus struct {
		storage.Host
		Connected bool `json:"connected"`
	}

	out := make([]hostWithStatus, 0, len(hosts))
	for _, host := range hosts {
		// Consider the host connected if the connector has an active connection.
		connected := true
		if _, err := h.Connector.GetConnection(host.ID); err != nil {
			connected = false
		}
		out = append(out, hostWithStatus{Host: host, Connected: connected})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *APIHandler) GetHostInfo(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	// Return host info when connected and include a connected flag so the UI
	// can always show connection state without treating missing info as an error.
	connected := true
	if _, err := h.Connector.GetConnection(hostID); err != nil {
		connected = false
	}

	var info *libvirt.HostInfo
	if connected {
		if hi, err := h.HostService.GetHostInfo(hostID); err == nil {
			info = hi
		}
	}

	resp := map[string]interface{}{"connected": connected, "info": info}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ConnectHost triggers a connection attempt for the given host id.
func (h *APIHandler) ConnectHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if !h.ValidateRequest(w, r, "hostID") {
		return
	}

	if err := h.HostService.EnsureHostConnectedForced(hostID); err != nil {
		h.HandleError(w, err, "connect_host")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisconnectHost requests a disconnect for the given host id.
func (h *APIHandler) DisconnectHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if !h.ValidateRequest(w, r, "hostID") {
		return
	}

	if err := h.HostService.DisconnectHost(hostID, true); err != nil {
		h.HandleError(w, err, "disconnect_host")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if !h.ValidateRequest(w, r, "hostID") {
		return
	}

	if err := h.HostService.RemoveHost(hostID); err != nil {
		h.HandleError(w, err, "delete_host")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListVMsFromLibvirt gets the unified view of VMs for a host.
func (h *APIHandler) ListVMsFromLibvirt(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")

	// Get VMs from the DB for a fast response.
	vms, err := h.HostService.GetVMsForHostFromDB(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
}

// ListDiscoveredVMs lists libvirt-only VMs for a host that are not in our DB.
func (h *APIHandler) ListDiscoveredVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	// If we don't have an active libvirt connection for this host, return an empty list.
	// Discovered-VMs is a lightweight UI-only fetch; a disconnected host shouldn't produce a 500.
	if _, err := h.Connector.GetConnection(hostID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	vms, err := h.HostService.ListDiscoveredVMs(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
}

// ImportVM imports a single discovered VM into the DB.
func (h *APIHandler) ImportVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")

	log.Infof("ImportVM request received - hostID: %s, vmName: %s", hostID, vmName)

	if err := h.HostService.ImportVM(hostID, vmName); err != nil {
		log.Errorf("ImportVM failed - hostID: %s, vmName: %s, error: %v", hostID, vmName, err)
		h.HandleError(w, err, "import_vm")
		return
	}

	log.Infof("ImportVM completed successfully - hostID: %s, vmName: %s", hostID, vmName)

	// Return a JSON response instead of empty body
	response := map[string]interface{}{
		"success": true,
		"message": "VM imported successfully",
		"vm_name": vmName,
		"host_id": hostID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// ImportAllVMs imports all discovered VMs on a host.
func (h *APIHandler) ImportAllVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if err := h.HostService.ImportAllVMs(hostID); err != nil {
		h.HandleError(w, err, "import_all_vms")
		return
	}

	// Return a JSON response
	response := map[string]interface{}{
		"success": true,
		"message": "All VMs imported successfully",
		"host_id": hostID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// ImportSelectedVMs imports selected discovered VMs by their domain UUIDs.
func (h *APIHandler) ImportSelectedVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")

	var req struct {
		DomainUUIDs []string `json:"domain_uuids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.DomainUUIDs) == 0 {
		http.Error(w, "No domain UUIDs provided", http.StatusBadRequest)
		return
	}

	if err := h.HostService.ImportSelectedVMs(hostID, req.DomainUUIDs); err != nil {
		h.HandleError(w, err, "import_selected_vms")
		return
	}

	// Return a JSON response
	response := map[string]interface{}{
		"success":        true,
		"message":        "Selected VMs imported successfully",
		"host_id":        hostID,
		"imported_count": len(req.DomainUUIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// DeleteSelectedDiscoveredVMs removes selected discovered VMs from the database.
func (h *APIHandler) DeleteSelectedDiscoveredVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")

	var req struct {
		DomainUUIDs []string `json:"domain_uuids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.DomainUUIDs) == 0 {
		http.Error(w, "No domain UUIDs provided", http.StatusBadRequest)
		return
	}

	if err := h.HostService.DeleteSelectedDiscoveredVMs(hostID, req.DomainUUIDs); err != nil {
		h.HandleError(w, err, "delete_discovered_vms")
		return
	}

	// Return a JSON response
	response := map[string]interface{}{
		"success":       true,
		"message":       "Selected discovered VMs deleted successfully",
		"host_id":       hostID,
		"deleted_count": len(req.DomainUUIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) GetVMStats(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	stats, err := h.HostService.GetVMStats(hostID, vmName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *APIHandler) GetVMHardware(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	hardware, err := h.HostService.GetVMHardwareAndDetectDrift(hostID, vmName)
	if err != nil {
		h.HandleError(w, err, fmt.Sprintf("get_vm_hardware_%s", vmName))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hardware)
}

// GetVMExtendedHardware retrieves comprehensive hardware configuration from all related database entities
func (h *APIHandler) GetVMExtendedHardware(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")

	log.Printf("Getting extended hardware configuration for VM %s on host %s", vmName, hostID)

	// Get the VM first
	var vm storage.VirtualMachine
	if err := h.DB.Where("host_id = ? AND name = ?", hostID, vmName).First(&vm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "VM not found", http.StatusNotFound)
			return
		}
		log.Printf("Error finding VM: %v", err)
		http.Error(w, "Failed to find VM", http.StatusInternalServerError)
		return
	}

	// Create comprehensive hardware response
	response := map[string]interface{}{
		"vm_info": vm,
	}

	// Load CPU topology
	var cpuTopology storage.CPUTopology
	if err := h.DB.Where("vm_id = ?", vm.ID).First(&cpuTopology).Error; err == nil {
		response["cpu_topology"] = cpuTopology
	}

	// Load CPU features
	var cpuFeatures []storage.CPUFeature
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&cpuFeatures).Error; err == nil {
		response["cpu_features"] = cpuFeatures
	}

	// Load memory configurations
	var memoryConfigs []storage.MemoryConfig
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&memoryConfigs).Error; err == nil {
		response["memory_configs"] = memoryConfigs
	}

	// Load disk attachments
	var diskAttachments []storage.DiskAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&diskAttachments).Error; err == nil {
		response["disk_attachments"] = diskAttachments
	}

	// Load port attachments
	var portAttachments []storage.PortAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&portAttachments).Error; err == nil {
		response["port_attachments"] = portAttachments
	}

	// Load video attachments
	var videoAttachments []storage.VideoAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&videoAttachments).Error; err == nil {
		response["video_attachments"] = videoAttachments
	}

	// Load controller attachments
	var controllerAttachments []storage.ControllerAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&controllerAttachments).Error; err == nil {
		response["controller_attachments"] = controllerAttachments
	}

	// Load host device attachments
	var hostDeviceAttachments []storage.HostDeviceAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&hostDeviceAttachments).Error; err == nil {
		response["host_device_attachments"] = hostDeviceAttachments
	}

	// Load TPM attachments
	var tpmAttachments []storage.TPMAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&tpmAttachments).Error; err == nil {
		response["tpm_attachments"] = tpmAttachments
	}

	// Load watchdog attachments
	var watchdogAttachments []storage.WatchdogAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&watchdogAttachments).Error; err == nil {
		response["watchdog_attachments"] = watchdogAttachments
	}

	// Load serial device attachments
	var serialDeviceAttachments []storage.SerialDeviceAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&serialDeviceAttachments).Error; err == nil {
		response["serial_device_attachments"] = serialDeviceAttachments
	}

	// Load filesystem attachments
	var filesystemAttachments []storage.FilesystemAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&filesystemAttachments).Error; err == nil {
		response["filesystem_attachments"] = filesystemAttachments
	}

	// Load RNG attachments
	var rngAttachments []storage.RngDeviceAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&rngAttachments).Error; err == nil {
		response["rng_attachments"] = rngAttachments
	}

	// Load memory balloon attachments
	var memoryBalloonAttachments []storage.MemoryBalloonAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&memoryBalloonAttachments).Error; err == nil {
		response["memory_balloon_attachments"] = memoryBalloonAttachments
	}

	// Load VSock attachments
	var vsockAttachments []storage.VsockAttachment
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&vsockAttachments).Error; err == nil {
		response["vsock_attachments"] = vsockAttachments
	}

	// Load boot configuration
	var bootConfig storage.BootConfig
	if err := h.DB.Where("vm_id = ?", vm.ID).First(&bootConfig).Error; err == nil {
		response["boot_config"] = bootConfig
	}

	// Load security labels
	var securityLabels []storage.SecurityLabel
	if err := h.DB.Where("vm_id = ?", vm.ID).Find(&securityLabels).Error; err == nil {
		response["security_labels"] = securityLabels
	}

	log.Printf("Extended hardware configuration loaded for VM %s with %d entity types", vmName, len(response))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding extended hardware response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
} // ListHostPorts returns unattached ports for a host (port pool).
func (h *APIHandler) ListHostPorts(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	ports, err := h.HostService.GetPortsForHostFromDB(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ports)
}

// ListVideoModels returns all known VideoModel templates.
func (h *APIHandler) ListVideoModels(w http.ResponseWriter, r *http.Request) {
	var models []storage.VideoModel
	if err := h.DB.Find(&models).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// ListHostVideoDevices returns physical video devices discovered on the host.
func (h *APIHandler) ListHostVideoDevices(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	var devices []storage.VideoDevice
	if err := h.DB.Where("host_device_id IN (SELECT id FROM host_devices WHERE host_id = ?)", hostID).Find(&devices).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// ListVMVideoAttachments returns video attachments for a VM by name.
func (h *APIHandler) ListVMVideoAttachments(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	// Resolve VM to UUID
	var vm storage.VirtualMachine
	if err := h.DB.Where("host_id = ? AND name = ?", hostID, vmName).First(&vm).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var atts []storage.VideoAttachment
	if err := h.DB.Preload("VideoModel").Where("vm_uuid = ?", vm.UUID).Find(&atts).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(atts)
}

// ListVMPortAttachments returns port attachments for a VM by name (looks up VM UUID).
func (h *APIHandler) ListVMPortAttachments(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	// Resolve VM to UUID
	var vm storage.VirtualMachine
	if err := h.DB.Where("host_id = ? AND name = ?", hostID, vmName).First(&vm).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	atts, err := h.HostService.GetPortAttachmentsForVM(vm.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(atts)
}

// --- VM Actions ---

func (h *APIHandler) StartVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.StartVM(hostID, vmName); err != nil {
		h.HandleError(w, err, fmt.Sprintf("start_vm_%s", vmName))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ShutdownVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.ShutdownVM(hostID, vmName); err != nil {
		h.HandleError(w, err, fmt.Sprintf("shutdown_vm_%s", vmName))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) RebootVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.RebootVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ForceOffVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.ForceOffVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ForceResetVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.ForceResetVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) SyncVMLive(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.SyncVMFromLibvirt(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) RebuildVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.RebuildVMFromDB(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateVMState updates the intended state of a VM in the database to match the provided state
func (h *APIHandler) UpdateVMState(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")

	var req struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.HostService.UpdateVMState(hostID, vmName, req.State); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetHostStats returns statistics for a specific host
func (h *APIHandler) GetHostStats(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")

	// Get basic host info first
	hostInfo, err := h.HostService.GetHostInfo(hostID)
	if err != nil {
		h.HandleError(w, err, fmt.Sprintf("get_host_info_%s", hostID))
		return
	}

	// Get VMs for this host to calculate VM statistics
	vms, err := h.HostService.GetVMsForHostFromDB(hostID)
	if err != nil {
		log.Printf("Warning: failed to get VMs for host %s: %v", hostID, err)
		vms = []services.VMView{} // Continue with empty VMs if this fails
	}

	// Calculate VM state counts
	vmCounts := map[string]int{
		"ACTIVE":  0,
		"STOPPED": 0,
		"PAUSED":  0,
		"ERROR":   0,
		"UNKNOWN": 0,
	}

	for _, vm := range vms {
		vmCounts[string(vm.State)]++
	}

	// Build response with host info and VM statistics
	stats := map[string]interface{}{
		"host_info": hostInfo,
		"vm_counts": vmCounts,
		"total_vms": len(vms),
		"resources": map[string]interface{}{
			"memory_bytes": hostInfo.Memory,
			"cpu_count":    hostInfo.CPU,
			"hostname":     hostInfo.Hostname,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// --- Dashboard Endpoints ---

// GetDashboardStats returns aggregated system-wide statistics.
func (h *APIHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.HostService.GetDashboardStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetDashboardActivity returns recent system activity events.
func (h *APIHandler) GetDashboardActivity(w http.ResponseWriter, r *http.Request) {
	// Get limit from query parameters, default to 10
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activities, err := h.HostService.GetDashboardActivity(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"activities": activities,
		"pagination": map[string]interface{}{
			"total": len(activities),
			"page":  1,
			"limit": limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDashboardOverview returns combined dashboard data for initial page load.
func (h *APIHandler) GetDashboardOverview(w http.ResponseWriter, r *http.Request) {
	stats, err := h.HostService.GetDashboardStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	activities, err := h.HostService.GetDashboardActivity(5)
	if err != nil {
		// Don't fail the whole request for activities, just log and continue
		log.Printf("Warning: failed to get activities: %v", err)
		activities = []services.ActivityEntry{} // empty array
	}

	response := map[string]interface{}{
		"stats":      stats,
		"activities": activities,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
