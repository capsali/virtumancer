package storage

import (
	"time"

	"gorm.io/gorm"
)

// DiscoveredVM represents a libvirt-only domain that has been observed on a host
// but not yet imported into Virtumancer's canonical VM table.
type DiscoveredVM struct {
	Base
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
// Returns true if a change was made, false otherwise.
func UpsertDiscoveredVM(db *gorm.DB, d *DiscoveredVM) (bool, error) {
	var existing []DiscoveredVM
	db.Where("host_id = ? AND domain_uuid = ?", d.HostID, d.DomainUUID).Limit(1).Find(&existing)
	if len(existing) > 0 {
		// Update only if LastSeenAt is more than 30 seconds old
		if time.Since(existing[0].LastSeenAt) > 30*time.Second {
			existing[0].Name = d.Name
			existing[0].InfoJSON = d.InfoJSON
			existing[0].LastSeenAt = time.Now()
			existing[0].Imported = d.Imported
			return true, db.Save(&existing[0]).Error
		}
		// No update needed
		return false, nil
	} else {
		d.LastSeenAt = time.Now()
		return true, db.Create(d).Error
	}
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

// BulkDeleteDiscoveredVMs removes multiple discovered VM records by their domain UUIDs.
func BulkDeleteDiscoveredVMs(db *gorm.DB, hostID string, domainUUIDs []string) error {
	if len(domainUUIDs) == 0 {
		return nil
	}
	return db.Where("host_id = ? AND domain_uuid IN ?", hostID, domainUUIDs).Delete(&DiscoveredVM{}).Error
}

// BulkMarkDiscoveredVMsImported marks multiple discovered VMs as imported.
func BulkMarkDiscoveredVMsImported(db *gorm.DB, hostID string, domainUUIDs []string) error {
	if len(domainUUIDs) == 0 {
		return nil
	}
	return db.Model(&DiscoveredVM{}).Where("host_id = ? AND domain_uuid IN ?", hostID, domainUUIDs).Updates(map[string]interface{}{"imported": true}).Error
}
