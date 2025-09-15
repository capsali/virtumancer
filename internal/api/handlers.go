package api

import (
	"encoding/json"
	"net/http"

	"github.com/capsali/virtumancer/internal/console"
	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"github.com/go-chi/chi/v5"
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	newHost, err := h.HostService.AddHost(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newHost)
}

func (h *APIHandler) GetHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.HostService.GetAllHosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (h *APIHandler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if err := h.HostService.RemoveHost(hostID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListVMsFromLibvirt gets the unified view of VMs for a host.
func (h *APIHandler) ListVMsFromLibvirt(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")

	// Immediately get VMs from the DB for a fast response.
	vms, err := h.HostService.GetVMsForHostFromDB(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// In the background, trigger a sync from libvirt.
	// The service will broadcast a websocket update when it's done.
	go h.HostService.SyncVMsForHost(hostID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
}

// ListDiscoveredVMs lists libvirt-only VMs for a host that are not in our DB.
func (h *APIHandler) ListDiscoveredVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
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
	if err := h.HostService.ImportVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// ImportAllVMs imports all discovered VMs on a host.
func (h *APIHandler) ImportAllVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if err := h.HostService.ImportAllVMs(hostID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
		// Even if there's an error (e.g., no cache yet), we might still proceed
		// if we want to allow the background sync to populate it.
		// For now, we'll return an error if the initial fetch fails.
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hardware)
}

// ListHostPorts returns unattached ports for a host (port pool).
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ShutdownVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName := chi.URLParam(r, "vmName")
	if err := h.HostService.ShutdownVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
