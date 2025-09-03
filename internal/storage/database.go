package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Host represents a libvirt connection configuration stored in the database.
type Host struct {
	ID        string `gorm:"primaryKey" json:"id"`
	URI       string `gorm:"not null" json:"uri"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// DB provides access to the database.
type DB struct {
	*gorm.DB
}

// NewDB initializes the database connection and auto-migrates schemas.
func NewDB(dsn string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema for the Host model.
	if err := db.AutoMigrate(&Host{}); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// GetAllHosts retrieves all hosts from the database.
func (db *DB) GetAllHosts() ([]Host, error) {
	var hosts []Host
	if err := db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// AddHost adds a new host to the database.
func (db *DB) AddHost(host *Host) error {
	return db.Create(host).Error
}

// DeleteHost removes a host from the database by its ID.
func (db *DB) DeleteHost(hostID string) error {
	return db.Delete(&Host{}, "id = ?", hostID).Error
}

