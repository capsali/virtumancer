# DB Schema Expansion Plan — Full Libvirt Device Coverage

Date: 2025-09-15
Author: virtumancer engineering

Goal
- Store a canonical representation of all libvirt domain XML elements in the database so VMs can be created from DB state and drift detected. Preserve the existing pattern of device resource tables (host-scoped or global) and separate per-VM attachment tables where appropriate.

Scope
- Inventory what we already have in `internal/storage/database.go`.
- Map libvirt-supported domain XML device types and related sections to database constructs.
- Propose concrete schema additions (tables + attachment tables) and per-table fields.
- Describe migration approach, indexes, API considerations, frontend impacts, and tests.

**Current Inventory (from `internal/storage/database.go`)
- Core VM: `VirtualMachine` (fields: HostID, Name, UUID, DomainUUID, VCPUCount, CPUModel, CPUTopologyJSON, MemoryBytes, OSType, SyncStatus, etc.)
- Storage: `StoragePool`, `Volume`, `VolumeAttachment`
- Networks: `Network`, `Port`, `PortBinding`, `PortAttachment`
- Controllers: `Controller`, `ControllerAttachment`
- Input devices: `InputDevice`, `InputDeviceAttachment`
- Console/Graphics: `Console` (VM-level: VNC/SPICE)
- Sound: `SoundCard`, `SoundCardAttachment`
- Host passthrough: `HostDevice`, `HostDeviceAttachment`
- TPM: `TPM`, `TPMAttachment`
- Watchdog: `Watchdog`, `WatchdogAttachment`
- Serial: `SerialDevice`, `SerialDeviceAttachment`
- Channel: `ChannelDevice`, `ChannelDeviceAttachment`
- Filesystem passthrough: `Filesystem`, `FilesystemAttachment`
- Smartcard: `Smartcard`, `SmartcardAttachment`
- USBRedirector: `USBRedirector`, `USBRedirectorAttachment`
- RNG: `RngDevice`, `RngDeviceAttachment`
- PanicDevice: `PanicDevice`, `PanicDeviceAttachment`
- Vsock: `Vsock`, `VsockAttachment`
- MemoryBalloon: `MemoryBalloon`, `MemoryBalloonAttachment`
- Shmem: `ShmemDevice`, `ShmemDeviceAttachment`
- IOMMU: `IOMMUDevice`, `IOMMUDeviceAttachment`
- Snapshots: `VMSnapshot`
- Index: `AttachmentIndex` (cross-device index used by UI/queries)

Notes: Many of the common libvirt device types are present already. Several areas are represented with JSON blobs (`VirtualPortJSON`, `VirtualMachine.CPUTopologyJSON`) where domain XML sub-elements are complex.

**Libvirt domain XML (common sections/devices) — quick overview**
- Domain-level: name, uuid, metadata, memory, currentMemory, vcpu, os (type/arch/boot), features (acpi/apic/hpet), cpu model/topology, clock, on_poweroff/on_reboot/on_crash, cputune, resource constraints
- Firmware / Boot: loader, nvram, bootmenu
- Devices:
  - disks: <disk type='file|block'> with <driver name type='qcow2|raw'>, <source file dev>, <target dev bus>, <backingStore>, serial, address
  - interfaces: <interface type='network|bridge|direct|hostdev|vhostuser|macvtap|vepa'> with <mac>, <source>, <model>, <address>, <filterref>, <virtualport>
  - hostdev: PCI/USB passthrough with <source><address> or vendor/product
  - controllers: sata/usb/ide/virtio-serial, with index/address
  - input devices: keyboard/tablet/mouse with bus
  - graphics/consoles: vnc/spice with listen/port/tls, and QEMU spice-specific options
  - video: model name (qxl, vga, virtio), ram, vram, heads
  - sound: model
  - rng: backend/dev
  - tpm: backend/emu
  - watchdog
  - serial / console / pty / tcp
  - channel (spicevmc, unix) for guest agent
  - hostdev pci and usb
  - iommu/passthrough (vfio), mediated devices (mdev)
  - filesystem passthrough (9p)
  - usb redirection and usb host devices
  - vsock
  - memory ballooning
  - device address mapping (PCI slot addressing)
  - qemu:commandline and other vendor-specific subtrees

**Gap Analysis — device types missing or under-modelled**
(From inventory vs libvirt list)
- Disk backingStore, disk driver options (cache, io, detect_zeroes, discard), disk serial and address are not fully modeled in `Volume`/`VolumeAttachment` (we have `Volume.Format` and `Volume.Name`). Need explicit driver options and backing store metadata.
- Disk addressing (PCI addresses for block devices) and disk specific flags (readonly, shareable) only partially present via `VolumeAttachment`.
- NIC: `Port`/`PortAttachment` covers many fields but lacks support for `filterref`, `virtualport` (structure), `vhostuser` options, and advanced source types like `direct`/`macvtap`. `Port` already has `VirtualPortJSON`, `FilterRefJSON` — these are placeholders; consider structured tables for `FilterRef` and `VirtualPort`.
- Video device: there is no `Video`/`VideoAttachment` table to represent `<video>` model settings (ram/vram/heads). Add `Video` + `VideoAttachment`.
- Firmware/boot: No `Firmware`/`Loader`/`BootOrder` tables; only `VirtualMachine.OSType`. Add `Firmware` or `BootConfig` table.
- CPU tuning & features: `CPUTopologyJSON` and `CPUModel` exist but a richer schema (`CPU`, `CPUTune`, `FeatureFlags`) might be useful for generation.
- QEMU specific options: `qemu:commandline`, `qemu:guest-agent` settings not captured; propose an `Extension` table or `VendorOption` JSON fields on attachments.
- Addressing: PCI addresses, slot/bus/domain/function used by many devices (hostdev, disk, controllers) are not fully modeled. Add an `Address` struct/table or fields on device/attachment.
- mdev (mediated device) entries absent — add `Mdev` and `MdevAttachment`.
- vhost-user specifics for interfaces and disks: consider `VhostUser` table or fields on `Port`.
- Audio/video option sets: QEMU args like `<backingStore>`, `<io>`, `<cache>` — add `DiskDriver` JSON or explicit typed fields.
- Serial/console advanced config: `target`, `protocol` currently modeled generically; ok but could capture more fields.

**Design Principles / Conventions**
- Preserve the resource/attachment pattern: for each device type D, prefer `D` (resource) and `DAttachment` where a resource can be shared or exists independently of a single VM.
- For truly per-VM-only devices (e.g., `Video`, `Console`, `Channel`, `VolumeAttachment`), keep only attachment rows referencing `VMUUID` and the resource if useful for dedup.
- Use explicit columns for commonly queried attributes (MAC, DeviceName, ModelName, DeviceID, Host address fields) and reserve `ConfigJSON`/`DriverJSON` columns for vendor-specific or rarely-queried nested data.
- Add `AttachmentIndex` entries on creation/update in the same transaction to preserve UI fast-queries.
- Use partial unique indexes where business rules require uniqueness (e.g., `port_attachments(vm_uuid, device_name)` already exists).

**Proposed New Tables / Fields (high level)**
- Disk enhancements
  - `Disk` (resource) — fields: `Name`, `StoragePoolID`, `Path`, `Format`, `CapacityBytes`, `Serial`, `DriverJSON` (cache/io/format options), `BackedByVolumeID` (nullable)
  - `DiskAttachment` (per-VM) — `VMUUID`, `DiskID`, `DeviceName`, `BusType`, `ReadOnly`, `Shareable`, `AddressJSON` (PCI address), `SnapshotPolicyJSON`
  - Keep current `Volume` for pool-managed volumes; `Disk` can reference `Volume` or a raw file path.

- Network enhancements
  - `FilterRef` table (optional) — `Name`, `ParametersJSON` and `FilterRefAttachment` or simply keep `FilterRefJSON` on `Port`.
  - `VirtualPort` structured table (optional) — fields per libvirt `<virtualport type='...'>` or use `VirtualPortJSON` kept as-is.
  - `VhostUser` settings on `Port` as `VhostUserJSON`.

- Video
  - `Video` (resource) — `ModelName`, `VRAM`, `Heads`, `Accel3D` flags
  - `VideoAttachment` — `VMUUID`, `VideoID`, `MonitorIndex`, `Primary` boolean

- Firmware / Boot
  - `BootConfig` — `VMUUID`, `LoaderPath`, `LoaderType`, `NVramPath`, `BootOrderJSON`, `SecureBoot` flag

- CPU & Features
  - `CPU` (resource) — `Model`, `TopologyJSON`, `FeaturesJSON`
  - `CPUTune` (attachment) — `VMUUID`, `Shares`, `Quota`, `Period`, `EmulatorThreads` and `CPUSet` (affinity)

- Addressing & PCI
  - `DeviceAddress` (reusable struct embedded or column JSON) — `Domain`, `Bus`, `Slot`, `Function`, `Type` (pci/usb)
  - Add `AddressJSON` to attachments that require it (DiskAttachment, PortAttachment, HostDeviceAttachment, ControllerAttachment)

- mdev / mediated devices
  - `MediatedDevice` + `MediatedDeviceAttachment` — tracks vendor-specific mdev types and configs

- vhost-user specifics
  - `VhostUser` resource table or `VhostUserJSON` columns on `Port`/`Disk`.

- QEMU vendor extensions
  - `VendorOption` generic table with `OwnerType`/`OwnerID`/`Namespace`/`Key`/`ValueJSON` to store `qemu:` or other vendor-specific options without schema churn.

- Other missing devices (one-to-one attachments exist or small tables to add):
  - `Video` (see above)
  - `MediatedDevice`/`Mdev`
  - `HostPCIDevice` detail (improved `HostDevice` with vendor/product ids and address parsed)
  - `Firmware`/`BootConfig` (see above)
  - `DeviceAlias` mapping (some UIs show alias names)

**Example Proposed Structs (sketch)**
- Disk
```go
type Disk struct {
  gorm.Model
  Name string
  VolumeID *uint // if managed by Volume
  Path string
  Format string
  CapacityBytes uint64
  Serial string
  DriverJSON string // driver options
}

type DiskAttachment struct {
  gorm.Model
  VMUUID string `gorm:"index"`
  DiskID uint
  DeviceName string
  BusType string
  ReadOnly bool
  Shareable bool
  AddressJSON string
}
```

- Video
```go
type Video struct {
  gorm.Model
  ModelName string
  VRAM uint
  Heads int
}

type VideoAttachment struct {
  gorm.Model
  VMUUID string `gorm:"index"`
  VideoID uint
  MonitorIndex int
}
```

- BootConfig
```go
type BootConfig struct {
  gorm.Model
  VMUUID string `gorm:"index;unique"`
  LoaderPath string
  LoaderType string
  NVramPath string
  BootOrderJSON string
}
```

- VendorOption
```go
type VendorOption struct {
  gorm.Model
  OwnerType string `gorm:"size:64;index"` // e.g., "disk_attachment", "port"
  OwnerID uint `gorm:"index"`
  Namespace string `gorm:"size:64"` // e.g., "qemu"
  Key string
  ValueJSON string
}
```

**Migration Strategy**
- Add new structs to `internal/storage/database.go` with `gorm` tags as above.
- Use `InitDB`/`AutoMigrate` to create new tables. Ensure `AutoMigrate` is idempotent and safe.
- For fields that are new but required for the `AttachmentIndex` unique constraints, use nullable columns and create/update `AttachmentIndex` entries in the same transaction when populating attachments.
- Back up `virtumancer.db` before running migrations (copy to `.bak`). Provide a CLI task in `tools/run_migrations` (exists) to run `InitDB` against the DB.
- Backfill strategies:
  - `Disk`: create `Disk` rows for existing `Volume` entries and set `DiskAttachment` to reference new rows where necessary.
  - `BootConfig`: create empty rows or infer loader from existing VM metadata if possible.
  - `Video`: create default `Video` rows for all VMs with a single monitor or none.
- Add checks that the `AttachmentIndex` unique indexes are present (existing code already ensures them in `InitDB`).

**Indexes & Constraints**
- For each attachment table, add partial unique constraints where appropriate, e.g.:
  - `disk_attachments(vm_uuid, device_name)` partial unique
  - `video_attachments(vm_uuid, monitor_index)` unique
  - `port_attachments(vm_uuid, device_name)` already exists
- Keep `AttachmentIndex` semantics: insert index entry in same transaction as attachment creation.
- Add indexes on commonly queried columns (`VMUUID`, `DeviceID`, `DeviceName`).

**API & Frontend Considerations**
- Update backend API handlers to expose new device endpoints and to include attachments in `GetVMHardwareAndDetectDrift` and `getVMHardwareFromDB`.
- Frontend changes: extend `web/src/stores/mainStore.js` and components in `web/src/components` to show Video, Disk driver options, Boot/firmware, CPU features, and advanced NIC options.
- Add UI flows for creating `Disk` resources from pool-backed `Volume`s and for attaching mediated devices.

**Testing & Validation**
- Unit tests for migration: verify new tables created and indexes exist.
- Integration test: create a VM with representative domain XML (disks, multiple NICs, hostdev, tpm, rng, console) and ensure DB rows match the XML and that `AttachmentIndex` entries are correct.
- Regression tests around attachment upsert to ensure the previous UNIQUE constraint errors do not reappear.

**Rollout Plan**
1. Add new structs and `AutoMigrate` via `internal/storage/database.go` and `tools/run_migrations`.
2. Add code paths in `syncVMHardware` and ingestion to populate new attachments (start with disk driver options, video, bootconfig, device addresses).
3. Backfill data with a migration script: create `Disk` rows for `Volume`s and populate `DiskAttachment` for existing `VolumeAttachment` rows.
4. Extend API and frontend incrementally (expose new devices read-only first, then add create flows).
5. Run integration tests and perform manual verification with representative libvirt hosts.

**Acceptance Criteria**
- All libvirt domain XML device types listed in the plan have a representation in the DB (either resource table, attachment table, or `VendorOption` fallback).
- `GetVMHardwareAndDetectDrift` returns hardware structures that include the new devices and match libvirt data.
- Creating or syncing VMs does not produce unique constraint violations; `AttachmentIndex` entries are maintained and de-duplicated.
- UI can display the new device types (read-only) for a VM after migration.

**Next Steps (short-term)**
- Agree on the proposed table list and the representation for complex nested elements (`virtualport`, `driver` options): JSON vs structured tables.
- Implement the highest-value additions first: `Disk`/`DiskAttachment` (capture driver/backingStore), `Video`/`VideoAttachment`, `BootConfig`.
- Implement backfill scripts for `Disk` → `Volume` mapping and `DiskAttachment` population.
- Iterate: add `MediatedDevice`, vendor options, and fine-grained NIC options.

---

Appendix A — Quick mapping table (libvirt element -> DB object)
- `<disk>` -> `Disk` (+ `DiskAttachment`) / `Volume` for pool volumes
- `<interface>` -> `Port` (+ `PortAttachment`)
- `<controller>` -> `Controller` (+ `ControllerAttachment`)
- `<input>` -> `InputDevice` (+ `InputDeviceAttachment`)
- `<graphics>`/`<console>` -> `Console` (existing) and `AttachmentIndex` entry
- `<video>` -> `Video` (+ `VideoAttachment`) [NEW]
- `<tpm>` -> `TPM` (+ `TPMAttachment`) (existing)
- `<rng>` -> `RngDevice` (+ `RngDeviceAttachment`) (existing)
- `<watchdog>` -> `Watchdog` (+ `WatchdogAttachment`) (existing)
- `<filesystem>` -> `Filesystem` (+ `FilesystemAttachment`) (existing)
- `<hostdev>` -> `HostDevice` (+ `HostDeviceAttachment`]
- `<mdev>` -> `MediatedDevice` (+ `MediatedDeviceAttachment`) [NEW]
- `<boot>` / `<os>` / `<loader>` -> `BootConfig` [NEW]
- vendor-specific (`qemu:`) -> `VendorOption` [NEW]


If you approve this plan I will:
- add a draft implementation to `internal/storage/database.go` with the new structs for `Disk`, `DiskAttachment`, `Video`, `VideoAttachment`, `BootConfig`, `VendorOption`, and `MediatedDevice`;
- add tests and migration/backfill scripts under `tools/` to populate the new tables from existing data;
- then run the migration locally (with DB backup) and present the verification results.



