package main

import (
	"flag"
	"os"

	log "github.com/capsali/virtumancer/internal/logging"

	"github.com/capsali/virtumancer/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := flag.String("db", "../virtumancer.db", "path to sqlite database file")
	dryRun := flag.Bool("dry-run", false, "perform a dry run without making changes")
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	if *verbose {
		log.SetLevel("verbose")
	}

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		log.Fatalf("db file not found: %s", *dbPath)
	}

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// AutoMigrate to ensure new tables exist
	if err := db.AutoMigrate(&storage.Disk{}, &storage.DiskAttachment{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	log.Infof("Starting disk schema migration...")

	// Migrate Volumes to Disks
	var volumes []storage.Volume
	if err := db.Find(&volumes).Error; err != nil {
		log.Fatalf("failed to fetch volumes: %v", err)
	}

	for _, vol := range volumes {
		var existing storage.Disk
		if err := db.Where("name = ?", vol.Name).First(&existing).Error; err == nil {
			log.Verbosef("Disk already exists for volume %s", vol.Name)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("error checking disk: %v", err)
		}
		disk := storage.Disk{
			Name:          vol.Name,
			VolumeID:      &vol.ID,
			Path:          vol.Name, // assuming path is name
			Format:        vol.Format,
			CapacityBytes: vol.CapacityBytes,
		}
		if *dryRun {
			log.Verbosef("dry-run: would create disk %+v", disk)
		} else {
			if err := db.Create(&disk).Error; err != nil {
				log.Fatalf("failed to create disk: %v", err)
			}
			log.Verbosef("Created disk %d for volume %s", disk.ID, vol.Name)
		}
	}

	// Migrate VolumeAttachments to DiskAttachments
	var volAtts []storage.VolumeAttachment
	if err := db.Preload("Volume").Find(&volAtts).Error; err != nil {
		log.Fatalf("failed to fetch volume attachments: %v", err)
	}

	for _, va := range volAtts {
		// Find the corresponding Disk
		var disk storage.Disk
		if err := db.Where("volume_id = ?", va.VolumeID).First(&disk).Error; err != nil {
			log.Fatalf("failed to find disk for volume %d: %v", va.VolumeID, err)
		}

		var existing storage.DiskAttachment
		if err := db.Where("vm_uuid = ? AND device_name = ?", va.VMUUID, va.DeviceName).First(&existing).Error; err == nil {
			log.Verbosef("DiskAttachment already exists for %s:%s", va.VMUUID, va.DeviceName)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("error checking disk attachment: %v", err)
		}

		da := storage.DiskAttachment{
			VMUUID:     va.VMUUID,
			DiskID:     disk.ID,
			DeviceName: va.DeviceName,
			BusType:    va.BusType,
			ReadOnly:   va.IsReadOnly,
		}
		if *dryRun {
			log.Verbosef("dry-run: would create disk attachment %+v", da)
		} else {
			if err := db.Create(&da).Error; err != nil {
				log.Fatalf("failed to create disk attachment: %v", err)
			}
			log.Verbosef("Created disk attachment %d for VM %s device %s", da.ID, va.VMUUID, va.DeviceName)
		}
	}

	log.Infof("Disk schema migration completed successfully")
}
