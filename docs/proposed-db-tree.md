# Proposed Database Tree — High-priority device schema additions

This document shows a compact tree view of the existing DB models and the planned schema additions for richer disk and interface modeling. Additive changes will keep compatibility and rely on `gorm.AutoMigrate` to add nullable columns.

Top-level (existing, abbreviated):

- Host
- VirtualMachine
  - VolumeAttachment
  - Port
  - ControllerAttachment
  - InputDeviceAttachment
  - Console
  - ... other attachments
- Volume
- Network
- AttachmentIndex

Planned changes (high-level tree view):

- VolumeAttachment (existing) — extend with:
  - ID (gorm.Model)
  - VMUUID string `gorm:"index"`
  - VolumeID uint
  - Volume Volume (rel)
  - DeviceName string // e.g., "vda"
  - BusType string // e.g., "virtio", "sata"
  - IsReadOnly bool
  - DeviceType string // new: "disk"|"cdrom"|"lun"|"snapshot"
  - TargetDev string // new: target.dev (vda, hda, nvme0n1)
  - TargetBus string // new: target.bus (virtio, sata, scsi, nvme)
  - DriverFormat string // new: driver.format (qcow2, raw)
  - DriverQueues int // new: driver.queues
  - DriverQueueSize int // new: driver.queue_size
  - IoTuneJSON string `gorm:"type:text"` // new: store iotune as JSON blob
  - Serial string // new: disk serial attribute
  - Shareable bool // new: shareable flag
  - SnapshotName string // new: snapshot subelement name

Rationale: rather than creating a separate Disk table, extending `VolumeAttachment` keeps the relationship explicit and minimizes schema churn. Multi-attach behavior for `Volume` remains unchanged.

- Port (existing) — extend with:
  - ID (gorm.Model)
  - VMUUID string `gorm:"index"`
  - MACAddress string
  - DeviceName string // e.g. vnet0
  - ModelName string // e.g. virtio, e1000
  - IPAddress string
  - SourceType string // new: 'network'|'bridge'|'hostdev'|'vhostuser'|'null'|'vdpa'
  - SourceRef string `gorm:"type:text"` // new: network name, hostdev address, or vhost socket path
  - VirtualPortJSON string `gorm:"type:text"` // new: virtualport parameters (802.1Qbh/others)
  - FilterRefJSON string `gorm:"type:text"` // new: nwfilter params
  - VlanTagsJSON string `gorm:"type:text"` // new: array of VLAN tags / tagging policy
  - TrustGuestRxFilters bool // new: trustGuestRxFilters attr

Rationale: capturing `source` and `virtualport` details is essential for correct representation of interface connectivity and to enable editing/re-creation of domain XML.

- Filesystem (extend): add `Type`/`Usage`/`SocketPath` fields for virtiofs.

- Files to edit (single patch):
  - `internal/storage/database.go` — add the new fields to `VolumeAttachment`, `Port`, and `Filesystem` structs and ensure AutoMigrate will run on next InitDB.
  - `docs/device-mapping.md` — (already created) reference updates.

Migration strategy:
- Additive changes only (nullable columns / default zero values). AutoMigrate is sufficient to add fields.
- For backfilling, provide `tools/backfill_disk_attrs` if needed to extract historical target/driver data from archived domain XML. For now, we assume fresh DB or incremental backfill later.

API/Parser changes (next step after DB changes):
- Modify import/parsing places (host sync / domain ingest) to populate new fields. Wrap writes in a DB transaction that also writes `AttachmentIndex` entries.
- For nested structures like `iotune`/`virtualport`, serialize to JSON into the newly added `*JSON` fields.

Testing plan:
- `go build ./...`
- Add unit test for `ParseDiskAndAttach` that reads a sample domain XML and writes to an in-memory sqlite DB (temp file) and asserts the `VolumeAttachment` fields are populated correctly.

If this layout looks good I will prepare the exact patch to `internal/storage/database.go` and run `go build ./...`. Would you like me to proceed to implement the changes now?