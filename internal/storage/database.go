package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Host represents a libvirt host connection configuration.
type Host struct {
	ID  string `gorm:"primaryKey" json:"id"`
	URI string `json:"uri"`
}

// VirtualMachine represents a virtual machine's configuration stored in the local DB.
type VirtualMachine struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex:idx_vm_host_name" json:"name"`
	HostID       string `gorm:"uniqueIndex:idx_vm_host_name" json:"host_id"`
	ConfigJSON   string `json:"config_json"` // Storing config as a JSON string for flexibility
}

// InitDB initializes and returns a GORM database instance.
func InitDB(dataSourceName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&Host{}, &VirtualMachine{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

