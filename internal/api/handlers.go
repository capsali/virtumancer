package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/go-chi/chi/v5"
)

type APIHandler struct {
	HostService *services.HostService
}

func NewAPIHandler(hostService *services.HostService) *APIHandler {
	return &APIHandler{
		HostService: hostService,
	}
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

	if _, err := h.HostService.AddHost(host); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(host)
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

// --- VM Action Handlers ---

// StartVM handles the request to start a virtual machine.
func (h *APIHandler) StartVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName, err := url.PathUnescape(chi.URLParam(r, "vmName"))
	if err != nil {
		http.Error(w, "Invalid VM name in URL", http.StatusBadRequest)
		return
	}

	if err := h.HostService.StartVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM started successfully"})
}

// GracefulShutdownVM handles the request to gracefully shut down a virtual machine.
func (h *APIHandler) GracefulShutdownVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName, err := url.PathUnescape(chi.URLParam(r, "vmName"))
	if err != nil {
		http.Error(w, "Invalid VM name in URL", http.StatusBadRequest)
		return
	}

	if err := h.HostService.GracefulShutdownVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM graceful shutdown initiated"})
}

// GracefulRebootVM handles the request to gracefully reboot a virtual machine.
func (h *APIHandler) GracefulRebootVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName, err := url.PathUnescape(chi.URLParam(r, "vmName"))
	if err != nil {
		http.Error(w, "Invalid VM name in URL", http.StatusBadRequest)
		return
	}
	if err := h.HostService.GracefulRebootVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM graceful reboot initiated"})
}

// ForceOffVM handles the request to forcefully stop a virtual machine.
func (h *APIHandler) ForceOffVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName, err := url.PathUnescape(chi.URLParam(r, "vmName"))
	if err != nil {
		http.Error(w, "Invalid VM name in URL", http.StatusBadRequest)
		return
	}
	if err := h.HostService.ForceOffVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM force off successful"})
}

// ForceResetVM handles the request to forcefully reset a virtual machine.
func (h *APIHandler) ForceResetVM(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	vmName, err := url.PathUnescape(chi.URLParam(r, "vmName"))
	if err != nil {
		http.Error(w, "Invalid VM name in URL", http.StatusBadRequest)
		return
	}
	if err := h.HostService.ForceResetVM(hostID, vmName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM force reset successful"})
}


