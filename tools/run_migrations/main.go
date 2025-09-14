package main

import (
	log "github.com/capsali/virtumancer/internal/logging"

	"github.com/capsali/virtumancer/internal/storage"
)

func main() {
	dbPath := "../virtumancer.db"
	db, err := storage.InitDB(dbPath)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	// Close DB if sqlite needs it (gorm DB doesn't need Close in this context)
	_ = db
	log.Infof("Migrations applied successfully")
}
