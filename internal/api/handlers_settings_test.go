package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dummy host service implementing minimal interface used by handler
type dummyHostService struct{}

func (d *dummyHostService) GetAllHosts() ([]storage.Host, error)                   { return []storage.Host{}, nil }
func (d *dummyHostService) EnsureHostConnected(hostID string) error                { return nil }
func (d *dummyHostService) EnsureHostConnectedForced(hostID string) error          { return nil }
func (d *dummyHostService) DisconnectHost(hostID string, userInitiated bool) error { return nil }
func (d *dummyHostService) GetHostInfo(hostID string) (*libvirt.HostInfo, error)   { return nil, nil }
func (d *dummyHostService) GetHostStats(hostID string) (*libvirt.HostStats, error) { return nil, nil }
func (d *dummyHostService) AddHost(host storage.Host) (*storage.Host, error)       { return &host, nil }
func (d *dummyHostService) RemoveHost(hostID string) error                         { return nil }
func (d *dummyHostService) ConnectToAllHosts()                                     {}
func (d *dummyHostService) GetVMsForHostFromDB(hostID string) ([]services.VMView, error) {
	return nil, nil
}
func (d *dummyHostService) GetVMStats(hostID, vmName string) (*services.ProcessedVMStats, error) {
	return nil, nil
}
func (d *dummyHostService) GetVMHardwareAndDetectDrift(hostID, vmName string) (*libvirt.HardwareInfo, error) {
	return nil, nil
}
func (d *dummyHostService) UpdateVMState(hostID, vmName, state string) error { return nil }
func (d *dummyHostService) GetPortsForHostFromDB(hostID string) ([]storage.Port, error) {
	return nil, nil
}
func (d *dummyHostService) GetPortAttachmentsForVM(vmUUID string) ([]services.PortAttachmentView, error) {
	return nil, nil
}
func (d *dummyHostService) SyncVMsForHost(hostID string) {}
func (d *dummyHostService) ListDiscoveredVMs(hostID string) ([]libvirt.VMInfo, error) {
	return nil, nil
}
func (d *dummyHostService) ImportVM(hostID, vmName string) error                        { return nil }
func (d *dummyHostService) ImportAllVMs(hostID string) error                            { return nil }
func (d *dummyHostService) ImportSelectedVMs(hostID string, domainUUIDs []string) error { return nil }
func (d *dummyHostService) DeleteSelectedDiscoveredVMs(hostID string, domainUUIDs []string) error {
	return nil
}
func (d *dummyHostService) SyncVMFromLibvirt(hostID, vmName string) error        { return nil }
func (d *dummyHostService) RebuildVMFromDB(hostID, vmName string) error          { return nil }
func (d *dummyHostService) StartVM(hostID, vmName string) error                  { return nil }
func (d *dummyHostService) ShutdownVM(hostID, vmName string) error               { return nil }
func (d *dummyHostService) RebootVM(hostID, vmName string) error                 { return nil }
func (d *dummyHostService) ForceOffVM(hostID, vmName string) error               { return nil }
func (d *dummyHostService) ForceResetVM(hostID, vmName string) error             { return nil }
func (d *dummyHostService) GetDashboardStats() (*services.DashboardStats, error) { return nil, nil }
func (d *dummyHostService) GetDashboardActivity(limit int) ([]services.ActivityEntry, error) {
	return nil, nil
}
func (d *dummyHostService) HandleClientDisconnect(client *ws.Client)                           {}
func (d *dummyHostService) HandleSubscribe(client *ws.Client, payload ws.MessagePayload)       {}
func (d *dummyHostService) HandleUnsubscribe(client *ws.Client, payload ws.MessagePayload)     {}
func (d *dummyHostService) HandleHostSubscribe(client *ws.Client, payload ws.MessagePayload)   {}
func (d *dummyHostService) HandleHostUnsubscribe(client *ws.Client, payload ws.MessagePayload) {}

func setupTestHandler(t *testing.T) (*APIHandler, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	// migrate Setting table only
	if err := db.AutoMigrate(&storage.Setting{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	hub := ws.NewHub()
	connector := &libvirt.Connector{}
	h := NewAPIHandler(&dummyHostService{}, hub, db, connector)
	return h, db
}

func TestGetMetricsDefaults(t *testing.T) {
	h, _ := setupTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/settings/metrics", nil)
	h.GetMetricsSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 ok, got %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed decode: %v", err)
	}
	if out["cpuDisplayDefault"] != "host" {
		t.Fatalf("unexpected default cpuDisplayDefault: %v", out["cpuDisplayDefault"])
	}
}

func TestUpdateMetricsValidation(t *testing.T) {
	h, db := setupTestHandler(t)

	// invalid alpha > 1
	body := map[string]interface{}{"diskSmoothAlpha": 1.5}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/settings/metrics", bytes.NewReader(b))
	h.UpdateMetricsSettings(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	// valid payload
	body2 := map[string]interface{}{"diskSmoothAlpha": 0.4, "netSmoothAlpha": 0.2, "cpuSmoothAlpha": 0.1, "cpuDisplayDefault": "guest", "units": map[string]string{"disk": "mib", "network": "kb"}}
	b2, _ := json.Marshal(body2)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", "/api/v1/settings/metrics", bytes.NewReader(b2))
	h.UpdateMetricsSettings(rr2, req2)
	if rr2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 no content, got %d", rr2.Code)
	}

	// ensure persisted
	var s storage.Setting
	if err := db.Where("key = ?", "metrics:global").First(&s).Error; err != nil {
		t.Fatalf("setting not persisted: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(s.ValueJSON), &parsed); err != nil {
		t.Fatalf("invalid stored JSON: %v", err)
	}
	if parsed["cpuDisplayDefault"] != "guest" {
		t.Fatalf("stored cpuDisplayDefault mismatch: %v", parsed["cpuDisplayDefault"])
	}
}

func TestGetRuntimeMetricsIncludesCpuDisplayDefault(t *testing.T) {
	h, db := setupTestHandler(t)

	// Persist a metrics:global setting with cpuDisplayDefault = "raw"
	payload := map[string]interface{}{"cpuDisplayDefault": "raw", "cpuSmoothAlpha": 0.25}
	b, _ := json.Marshal(payload)
	s := storage.Setting{Key: "metrics:global", ValueJSON: string(b), OwnerType: "global"}
	// Upsert to avoid duplicate rows across tests using shared in-memory DB
	if err := db.Where("key = ?", "metrics:global").Assign(s).FirstOrCreate(&s).Error; err != nil {
		t.Fatalf("failed to persist setting: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/settings/metrics/runtime", nil)
	h.GetRuntimeMetricsSettings(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if out["cpuDisplayDefault"] != "raw" {
		t.Fatalf("expected cpuDisplayDefault 'raw', got %v", out["cpuDisplayDefault"])
	}
}
