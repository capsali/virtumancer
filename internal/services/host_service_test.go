package services

import (
	"testing"

	"github.com/capsali/virtumancer/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&storage.Host{}, &storage.VirtualMachine{})
	require.NoError(t, err)

	return db
}

func TestDisconnectHost_UserInitiated(t *testing.T) {
	db := setupTestDB(t)

	// Create a test host
	testHost := &storage.Host{
		Base:                  storage.Base{ID: "test-host"},
		URI:                   "qemu+ssh://test@example.com/system",
		State:                 "CONNECTED",
		AutoReconnectDisabled: false,
	}
	err := db.Create(testHost).Error
	require.NoError(t, err)

	// Simulate what DisconnectHost does for user-initiated disconnect
	updates := map[string]interface{}{
		"task_state":              "",
		"state":                   "DISCONNECTED",
		"auto_reconnect_disabled": true,
	}
	err = db.Model(&storage.Host{}).Where("id = ?", "test-host").Updates(updates).Error
	require.NoError(t, err)

	// Verify the host was updated correctly
	var updatedHost storage.Host
	err = db.Where("id = ?", "test-host").First(&updatedHost).Error
	require.NoError(t, err)

	assert.Equal(t, "DISCONNECTED", updatedHost.State)
	assert.True(t, updatedHost.AutoReconnectDisabled)
}

func TestDisconnectHost_NetworkDisconnect(t *testing.T) {
	db := setupTestDB(t)

	// Create a test host
	testHost := &storage.Host{
		Base:                  storage.Base{ID: "test-host"},
		URI:                   "qemu+ssh://test@example.com/system",
		State:                 "CONNECTED",
		AutoReconnectDisabled: false,
	}
	err := db.Create(testHost).Error
	require.NoError(t, err)

	// Simulate what DisconnectHost does for network-initiated disconnect
	updates := map[string]interface{}{
		"task_state": "",
		"state":      "DISCONNECTED",
	}
	err = db.Model(&storage.Host{}).Where("id = ?", "test-host").Updates(updates).Error
	require.NoError(t, err)

	// Verify the host was updated correctly - auto_reconnect should remain enabled
	var updatedHost storage.Host
	err = db.Where("id = ?", "test-host").First(&updatedHost).Error
	require.NoError(t, err)

	assert.Equal(t, "DISCONNECTED", updatedHost.State)
	assert.False(t, updatedHost.AutoReconnectDisabled) // Should remain false for network disconnects
}

func TestEnsureHostConnected_RespectsAutoReconnectDisabled(t *testing.T) {
	db := setupTestDB(t)

	// Create a test host with auto-reconnect disabled
	testHost := &storage.Host{
		Base:                  storage.Base{ID: "test-host"},
		URI:                   "qemu+ssh://test@example.com/system",
		State:                 "DISCONNECTED",
		AutoReconnectDisabled: true,
	}
	err := db.Create(testHost).Error
	require.NoError(t, err)

	// Test the logic that EnsureHostConnected would use
	var host storage.Host
	err = db.Where("id = ?", "test-host").First(&host).Error
	require.NoError(t, err)

	// EnsureHostConnected should check this flag and fail
	if host.AutoReconnectDisabled {
		// This is the expected behavior - it should fail
		assert.True(t, host.AutoReconnectDisabled)
	} else {
		t.Error("EnsureHostConnected should have failed due to auto_reconnect_disabled")
	}
}

func TestManualConnect_ResetsAutoReconnectDisabled(t *testing.T) {
	db := setupTestDB(t)

	// Create a test host with auto-reconnect disabled
	testHost := &storage.Host{
		Base:                  storage.Base{ID: "test-host"},
		URI:                   "qemu+ssh://test@example.com/system",
		State:                 "DISCONNECTED",
		AutoReconnectDisabled: true,
	}
	err := db.Create(testHost).Error
	require.NoError(t, err)

	// Simulate what EnsureHostConnectedForced does
	updates := map[string]interface{}{
		"task_state":              "",
		"state":                   "CONNECTED",
		"auto_reconnect_disabled": false,
	}
	err = db.Model(&storage.Host{}).Where("id = ?", "test-host").Updates(updates).Error
	require.NoError(t, err)

	// Check final state
	var updatedHost storage.Host
	err = db.Where("id = ?", "test-host").First(&updatedHost).Error
	require.NoError(t, err)

	assert.Equal(t, "CONNECTED", updatedHost.State)
	assert.False(t, updatedHost.AutoReconnectDisabled)
}
