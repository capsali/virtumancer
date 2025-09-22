package storage

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// DiscoveredVM represents a libvirt-only domain that has been observed on a host
// but not yet imported into Virtumancer's canonical VM table.
type DiscoveredVM struct {
	gorm.Model
	HostID     string    `gorm:"index;not null" json:"host_id"`
	Name       string    `json:"name"`
	DomainUUID string    `gorm:"size:64;index" json:"domain_uuid"`
	InfoJSON   string    `gorm:"type:text" json:"info_json"` // optional serialized domain XML / metadata
	LastSeenAt time.Time `json:"last_seen_at"`
	Imported   bool      `gorm:"default:false;index" json:"imported"`
}

// TableName forces a predictable table name so we can create indexes via SQL safely.
func (DiscoveredVM) TableName() string {
	return "discovered_vms"
}

// UpsertDiscoveredVM inserts or updates a discovered VM record, updating LastSeenAt only if it's older than 30 seconds.
func UpsertDiscoveredVM(db *gorm.DB, d *DiscoveredVM) error {
	var existing DiscoveredVM
	err := db.Where("host_id = ? AND domain_uuid = ?", d.HostID, d.DomainUUID).First(&existing).Error
	if err == nil {
		// Update only if LastSeenAt is more than 30 seconds old
		if time.Since(existing.LastSeenAt) > 30*time.Second {
			existing.Name = d.Name
			existing.InfoJSON = d.InfoJSON
			existing.LastSeenAt = time.Now()
			existing.Imported = d.Imported
			return db.Save(&existing).Error
		}
		// No update needed
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		d.LastSeenAt = time.Now()
		return db.Create(d).Error
	}
	return err
}

// ListDiscoveredVMsByHost lists non-imported discovered VMs for a host.
func ListDiscoveredVMsByHost(db *gorm.DB, hostID string) ([]DiscoveredVM, error) {
	var out []DiscoveredVM
	if err := db.Where("host_id = ? AND imported = 0", hostID).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteDiscoveredVMByDomainUUID removes a discovered VM record by its domain UUID.
func DeleteDiscoveredVMByDomainUUID(db *gorm.DB, hostID, domainUUID string) error {
	return db.Where("host_id = ? AND domain_uuid = ?", hostID, domainUUID).Delete(&DiscoveredVM{}).Error
}

// MarkDiscoveredVMImported marks a discovered VM as imported so it no longer shows
// up in the discovered list.
func MarkDiscoveredVMImported(db *gorm.DB, hostID, domainUUID string) error {
	return db.Model(&DiscoveredVM{}).Where("host_id = ? AND domain_uuid = ?", hostID, domainUUID).Updates(map[string]interface{}{"imported": true}).Error
}
