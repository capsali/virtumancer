package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	log "github.com/capsali/virtumancer/internal/logging"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/capsali/virtumancer/internal/storage"
)

func main() {
	dbPath := flag.String("db", "../virtumancer.db", "path to sqlite database file")
	fixDuplicates := flag.Bool("fix", false, "if set, remove duplicate attachment_indices (keep oldest unless --keep-newest is set)")
	reportCSV := flag.String("report-csv", "", "path to CSV file to write duplicate groups report")
	createIndex := flag.Bool("create-index", false, "if set, attempt to create the unique index on (device_type, device_id) after dedupe/backfill")
	yesFlag := flag.Bool("yes", false, "skip interactive confirmation prompts")
	yFlag := flag.Bool("y", false, "shorthand for -yes")
	forceFlag := flag.Bool("force", false, "alias for -yes; skip interactive confirmation prompts")
	keepNewest := flag.Bool("keep-newest", false, "when deduping, keep the newest record instead of the oldest")
	dryRun := flag.Bool("dry-run", false, "perform a dry run without making changes")
	verboseFlag := flag.Bool("verbose", false, "enable verbose logging")
	vFlag := flag.Bool("v", false, "shorthand for -verbose")
	logFile := flag.String("log-file", "", "path to file to append logs")
	flag.Parse()

	verbose := *verboseFlag || *vFlag
	if verbose {
		log.SetLevel("verbose")
	}

	// Configure log file if requested
	var logF *os.File
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open log file %s: %v", *logFile, err)
		}
		logF = f
		mw := io.MultiWriter(os.Stdout, logF)
		log.SetOutput(mw)
		defer logF.Close()
	}

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		log.Fatalf("db file not found: %s", *dbPath)
	}

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// Ensure migration for AttachmentIndex
	if err := db.AutoMigrate(&storage.AttachmentIndex{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Dedupe existing attachment_indices by (device_type, device_id) if requested
	if err := dedupeAttachmentIndices(db, *fixDuplicates, *reportCSV, *keepNewest, *dryRun, verbose); err != nil {
		log.Fatalf("dedupe failed: %v", err)
	}

	count := 0
	start := time.Now()

	// Helper to upsert an allocation row if not exists
	upsert := func(vmUUID, deviceType string, attachmentID uint, deviceID uint) error {
		var existing storage.AttachmentIndex
		res := db.Where("device_type = ? AND attachment_id = ?", deviceType, attachmentID).First(&existing)
		if res.Error == nil {
			return nil // already present
		}
		if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
			return res.Error
		}
		// For volumes, allow multi-attach: record the allocation but leave DeviceID as 0
		if deviceType == "volume" {
			deviceID = 0
		}
		alloc := storage.AttachmentIndex{VMUUID: vmUUID, DeviceType: deviceType, AttachmentID: attachmentID, DeviceID: deviceID}
		if *dryRun {
			log.Verbosef("dry-run: would insert allocation: %+v", alloc)
			return nil
		}
		return db.Create(&alloc).Error
	}

	// Discover tables with "attachment" in their name from sqlite_master
	tablesRows, err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE ?", "%attachment%").Rows()
	if err != nil {
		log.Fatalf("query sqlite_master: %v", err)
	}
	defer tablesRows.Close()

	for tablesRows.Next() {
		var tbl string
		if err := tablesRows.Scan(&tbl); err != nil {
			log.Fatalf("scan table name: %v", err)
		}

		if verbose {
			log.Verbosef("Inspecting table: %s", tbl)
		}
		// Inspect columns
		pragma := fmt.Sprintf("PRAGMA table_info(%s);", tbl)
		colRows, err := db.Raw(pragma).Rows()
		if err != nil {
			log.Fatalf("pragma table_info %s: %v", tbl, err)
		}

		deviceIDCol := ""
		hasVMUUID := false
		for colRows.Next() {
			// PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
			var cid int
			var name string
			var ctype sql.NullString
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := colRows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				colRows.Close()
				log.Fatalf("scan pragma: %v", err)
			}
			lname := strings.ToLower(name)
			if lname == "vm_uuid" {
				hasVMUUID = true
			}
			if lname == "id" {
				continue // primary id col
			}
			if strings.HasSuffix(lname, "_id") && deviceIDCol == "" {
				// prefer columns like '*device_id' or '*_id'
				deviceIDCol = name
			}
		}
		colRows.Close()

		if !hasVMUUID {
			// not a VM attachment table
			continue
		}

		// Determine device type label from table name
		deviceType := tbl
		deviceType = strings.TrimSuffix(deviceType, "_attachments")
		deviceType = strings.TrimSuffix(deviceType, "_attachment")

		// Build select list
		selectCols := "id, vm_uuid"
		if deviceIDCol != "" {
			selectCols = fmt.Sprintf("%s, %s", selectCols, deviceIDCol)
		}

		rows, err := db.Raw(fmt.Sprintf("SELECT %s FROM %s", selectCols, tbl)).Rows()
		if err != nil {
			log.Fatalf("select from %s: %v", tbl, err)
		}
		for rows.Next() {
			var id uint
			var vmUUID sql.NullString
			var deviceID sql.NullInt64
			if deviceIDCol != "" {
				if err := rows.Scan(&id, &vmUUID, &deviceID); err != nil {
					rows.Close()
					log.Fatalf("scan row from %s: %v", tbl, err)
				}
			} else {
				if err := rows.Scan(&id, &vmUUID); err != nil {
					rows.Close()
					log.Fatalf("scan row from %s: %v", tbl, err)
				}
			}

			if !vmUUID.Valid || vmUUID.String == "" {
				continue // no vm association
			}

			var devID uint
			if deviceID.Valid {
				devID = uint(deviceID.Int64)
			}

			if err := upsert(vmUUID.String, deviceType, id, devID); err != nil {
				rows.Close()
				log.Fatalf("upsert alloc for %s id %d: %v", tbl, id, err)
			}
				if verbose && !*dryRun {
					log.Verbosef("inserted allocation for vm=%s device=%s attachment_id=%d device_id=%d", vmUUID.String, deviceType, id, devID)
				}
			count++
		}
		rows.Close()
	}

	dur := time.Since(start)
	fmt.Printf("Backfill completed: %d entries in %s\n", count, dur)

	if *createIndex {
		// Check for duplicates first
		rows, err := db.Raw(`SELECT device_type, device_id, COUNT(*) as cnt FROM attachment_indices WHERE device_type != 'volume' AND device_id IS NOT NULL AND device_id != 0 GROUP BY device_type, device_id HAVING cnt > 1`).Rows()
		if err != nil {
			log.Fatalf("failed to check duplicates: %v", err)
		}
		defer rows.Close()
		hasDup := false
		for rows.Next() {
			hasDup = true
			break
		}

	if hasDup {
			// run automatically if -yes/-y/--force provided
			skipConfirm := *yesFlag || *yFlag || *forceFlag
			if skipConfirm {
				log.Verbosef("Skipping confirmation due to --yes; running dedupe and creating index...")
				if err := dedupeAttachmentIndices(db, true, *reportCSV, *keepNewest, *dryRun, verbose); err != nil {
					log.Fatalf("dedupe failed: %v", err)
				}
				if err := createUniqueDeviceIndex(db, *dryRun); err != nil {
					log.Fatalf("failed to create unique device index after dedupe: %v", err)
				}
				log.Verbosef("Unique index on (device_type, device_id) created successfully")
			} else {
				// Ask for confirmation to run dedupe automatically
				fmt.Print("Duplicates exist for (device_type, device_id). Run dedupe now and create the unique index? [y/N]: ")
				var resp string
				if _, err := fmt.Scanln(&resp); err != nil {
					log.Fatalf("failed to read input: %v", err)
				}
				resp = strings.ToLower(strings.TrimSpace(resp))
				if resp == "y" || resp == "yes" {
					log.Verbosef("Running dedupe (removing duplicates, keeping oldest)...")
					if err := dedupeAttachmentIndices(db, true, *reportCSV, *keepNewest, *dryRun, verbose); err != nil {
						log.Fatalf("dedupe failed: %v", err)
					}
					// Now create the index
					if err := createUniqueDeviceIndex(db, *dryRun); err != nil {
						log.Fatalf("failed to create unique device index after dedupe: %v", err)
					}
					log.Verbosef("Unique index on (device_type, device_id) created successfully")
				} else {
					log.Verbosef("Aborting index creation. Run the tool with -fix to remove duplicates first or rerun with -create-index and confirm.")
				}
			}
		} else {
			// No duplicates, create index directly
			if err := createUniqueDeviceIndex(db, *dryRun); err != nil {
				log.Fatalf("failed to create unique device index: %v", err)
			}
			log.Verbosef("Unique index on (device_type, device_id) created successfully")
		}
	}
}

// createUniqueDeviceIndex checks for duplicates and creates the unique index if none remain.
func createUniqueDeviceIndex(db *gorm.DB, dryRun bool) error {
	// Re-check for duplicates
	rows, err := db.Raw(`SELECT device_type, device_id, COUNT(*) as cnt FROM attachment_indices WHERE device_id IS NOT NULL AND device_id != 0 GROUP BY device_type, device_id HAVING cnt > 1`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	var hasDup bool
	for rows.Next() {
		hasDup = true
		break
	}
	if hasDup {
		return fmt.Errorf("duplicates still exist for (device_type, device_id); run dedupe with -fix before creating index")
	}

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uniq_attachment_index_device ON attachment_indices(device_type, device_id) WHERE device_type != 'volume' AND device_id IS NOT NULL;`).Error; err != nil {
		return err
	}
	return nil
}

// dedupeAttachmentIndices finds groups with the same (device_type, device_id) and
// optionally removes duplicates (keeps the lowest id).
func dedupeAttachmentIndices(db *gorm.DB, fix bool, reportPath string, keepNewest bool, dryRun bool, verbose bool) error {
	type grp struct {
		DeviceType string
		DeviceID   int64
		Cnt        int
	}

	// Exclude volume device_type since volumes can be multi-attached and we store DeviceID=0 for them.
	rows, err := db.Raw(`SELECT device_type, device_id, COUNT(*) as cnt FROM attachment_indices WHERE device_type != 'volume' AND device_id IS NOT NULL AND device_id != 0 GROUP BY device_type, device_id HAVING cnt > 1`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var groups []grp
	for rows.Next() {
		var g grp
		if err := rows.Scan(&g.DeviceType, &g.DeviceID, &g.Cnt); err != nil {
			return err
		}
		groups = append(groups, g)
	}

	if len(groups) == 0 {
		log.Infof("No duplicate attachment_indices found (by device_type+device_id)")
		return nil
	}

	if verbose {
	log.Infof("Found %d duplicate groups; fix=%v; report=%s", len(groups), fix, reportPath)
	} else {
	log.Infof("Found %d duplicate groups; fix=%v", len(groups), fix)
	}

	// Prepare CSV writer if requested
	var csvWriter *csv.Writer
	var csvFile *os.File
	if reportPath != "" {
		f, err := os.Create(reportPath)
		if err != nil {
			return fmt.Errorf("create report file: %w", err)
		}
		csvFile = f
		csvWriter = csv.NewWriter(f)
		// header
		if err := csvWriter.Write([]string{"device_type", "device_id", "count", "all_ids", "remove_ids"}); err != nil {
			csvFile.Close()
			return fmt.Errorf("write csv header: %w", err)
		}
	}

	totalRemoved := 0
	for _, g := range groups {
		var items []storage.AttachmentIndex
		if err := db.Where("device_type = ? AND device_id = ?", g.DeviceType, g.DeviceID).Order("id ASC").Find(&items).Error; err != nil {
			if csvFile != nil {
				csvFile.Close()
			}
			return err
		}
		if len(items) <= 1 {
			continue
		}
		// collect ids
		var allIDs []string
		for _, it := range items {
			allIDs = append(allIDs, strconv.FormatUint(uint64(it.ID), 10))
		}
		// choose items to remove depending on keepNewest flag
		var removeIDs []string
		var removeIDNums []uint
		if keepNewest {
			// keep the newest -> remove all but the last
			for i := 0; i < len(items)-1; i++ {
				removeIDs = append(removeIDs, strconv.FormatUint(uint64(items[i].ID), 10))
				removeIDNums = append(removeIDNums, items[i].ID)
			}
		} else {
			// keep the oldest -> remove all after the first
			for i := 1; i < len(items); i++ {
				removeIDs = append(removeIDs, strconv.FormatUint(uint64(items[i].ID), 10))
				removeIDNums = append(removeIDNums, items[i].ID)
			}
		}

		if csvWriter != nil {
			if err := csvWriter.Write([]string{g.DeviceType, strconv.FormatInt(g.DeviceID, 10), strconv.Itoa(g.Cnt), strings.Join(allIDs, ","), strings.Join(removeIDs, ",")}); err != nil {
				csvFile.Close()
				return fmt.Errorf("write csv row: %w", err)
			}
		}

		if dryRun {
			log.Verbosef("dry-run: would remove ids %v for device_type=%s device_id=%d", removeIDs, g.DeviceType, g.DeviceID)
		} else if !fix {
			log.Verbosef("Duplicate for device_type=%s device_id=%d: %d entries (would remove %d)", g.DeviceType, g.DeviceID, g.Cnt, len(removeIDs))
			continue
		}

		if !dryRun {
			tx := db.Begin()
			if err := tx.Where("id IN ?", removeIDNums).Delete(&storage.AttachmentIndex{}).Error; err != nil {
				tx.Rollback()
				if csvFile != nil {
					csvFile.Close()
				}
				return err
			}
			if err := tx.Commit().Error; err != nil {
				if csvFile != nil {
					csvFile.Close()
				}
				return err
			}
		}
		totalRemoved += len(removeIDNums)
	log.Infof("Removed %d duplicate entries for device_type=%s device_id=%d", len(removeIDNums), g.DeviceType, g.DeviceID)
	}

	if csvWriter != nil {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			csvFile.Close()
			return fmt.Errorf("flush csv: %w", err)
		}
		csvFile.Close()
	log.Infof("Duplicate report written to %s", reportPath)
	}

	if fix {
		log.Infof("Dedupe complete; total removed: %d", totalRemoved)
	}
	return nil
}
