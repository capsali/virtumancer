package api

import (
	"encoding/json"
	"net/http"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/go-chi/chi/v5"
)

// APIHandler holds dependencies for API handlers.
type APIHandler struct {
	LibvirtConnector *libvirt.Connector
	HostService      *services.HostService
}

// HealthCheck is a simple handler to confirm the API is running.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok"}
	writeJSON(w, http.StatusOK, response)
}

// --- Host Handlers ---

// AddHost handles requests to add and connect to a new libvirt host.
func (h *APIHandler) AddHost(w http.ResponseWriter, r *http.Request) {
	var newHost storage.Host
	if err := json.NewDecoder(r.Body).Decode(&newHost); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if newHost.ID == "" || newHost.URI == "" {
		writeError(w, http.StatusBadRequest, "Host 'id' and 'uri' are required fields")
		return
	}

	host, err := h.HostService.AddHost(newHost.ID, newHost.URI)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, host)
}

// ListHosts returns a list of all configured hosts from the database.
func (h *APIHandler) ListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.HostService.GetAllHosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, hosts)
}

// DeleteHost handles requests to disconnect and remove a host.
func (h *APIHandler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if hostID == "" {
		writeError(w, http.StatusBadRequest, "Host ID is required")
		return
	}

	if err := h.HostService.DeleteHost(hostID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


// --- VM Handlers ---

// ListVMs handles the request to list all virtual machines on a given host.
func (h *APIHandler) ListVMs(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if hostID == "" {
		writeError(w, http.StatusBadRequest, "Host ID is required")
		return
	}

	domains, err := h.LibvirtConnector.ListAllDomains(hostID)
	if err != nil {
		if _, ok := err.(services.ErrHostNotConnected); ok {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domains)
}

// writeJSON is a helper function for writing JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeError is a helper function for writing JSON error responses.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}


