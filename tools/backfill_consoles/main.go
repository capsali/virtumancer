package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/capsali/virtumancer/internal/logging"
)

func main() {
	_ = flag.String("db", "../virtumancer.db", "path to sqlite database file")
	// deprecated: flags are unused for no-op tool
	flag.Parse()

	// Deprecated: this tool was used to migrate existing `graphics_devices` and
	// `graphics_device_attachments` into `consoles`. If you are starting with a
	// fresh database, no backfill is required. This tool is a no-op.
	fmt.Println("tools/backfill_consoles is deprecated: no-op for fresh DBs")
	log.Verbosef("backfill_consoles invoked; no action taken for fresh DBs")
	os.Exit(0)
}
