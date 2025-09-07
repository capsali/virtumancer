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
	HostService *services.HostService
	Hub         *ws.Hub
	DB          *gorm.DB
	Connector   *libvirt.Connector
}

func NewAPIHandler(hostService *services.HostService, hub *ws.Hub, db *gorm.DB, connector *libvirt.Connector) *APIHandler {
	return &APIHandler{
		HostService: hostService,
		Hub:         hub,
		DB:          db,
		Connector:   connector,
	}
}

// HandleWebSocket handles websocket requests for UI updates from the peer.
func (h *APIHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(h.Hub, w, r)
}

// HandleVMConsole proxies the WebSocket connection to the VM's VNC console.
func (h *APIHandler) HandleVMConsole(w http.ResponseWriter, r *http.Request) {
	console.HandleConsole(h.DB, h.Connector, w, r)
}

// HandleSpiceConsole proxies the WebSocket connection to the VM's SPICE console.
func (h *APIHandler) HandleSpiceConsole(w http.ResponseWriter, r *http.Request) {
	console.HandleSpiceConsole(h.DB, h.Connector, w, r)
}

// HealthCheck confirms the server is running.
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// CreateHost handles adding a new host.
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

// GetHosts returns a list of all configured hosts.
func (h *APIHandler) GetHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.HostService.GetAllHosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

// DeleteHost handles removing a host.
func (h *APIHandler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if err := h.HostService.RemoveHost(hostID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetHostInfo returns statistics for a specific host.
func (h *APIHandler) GetHostInfo(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	info, err := h.HostService.GetHostInfo(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// ListVMs lists all virtual machines on a specific host.
func (h *APIHandler) ListVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vms, err := h.HostService.ListVMs(hostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
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

