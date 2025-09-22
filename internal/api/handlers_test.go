package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAPITest(t *testing.T) (*APIHandler, *gorm.DB) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&storage.Host{}, &storage.VirtualMachine{})
	require.NoError(t, err)

	// Create mock services
	hub := ws.NewHub()
	go hub.Run()

	// Create a minimal host service for testing (using NewHostService constructor)
	hostService := services.NewHostService(db, nil, hub)

	// Create API handler
	apiHandler := &APIHandler{
		HostService: hostService,
		DB:          db,
		Hub:         hub,
	}

	return apiHandler, db
}

func TestHealthCheckEndpoint(t *testing.T) {
	apiHandler, _ := setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	apiHandler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["ok"])
}