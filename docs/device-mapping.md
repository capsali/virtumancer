# Libvirt Device → DB Mapping

This document maps canonical libvirt domain `<devices>` elements to the existing Virtumancer DB models in `internal/storage/database.go`, notes where data is already captured, and recommends priorities for additional modeling.

Status key
- **Present**: a focused model + attachment table exist in `internal/storage/database.go`.
- **Partial**: some representation exists but important attributes are not captured.
- **Missing**: no dedicated model exists; data would be lost or flattened if imported.

Priority key
- **High**: required for basic create/manage workflows (disks, NICs, consoles, hostdev).
- **Medium**: useful for lifecycle and management but not blocking VM create (controllers, channels, rng, tpm).
- **Low**: niche or advanced features (pstore, crypto backends, vhost-user specifics).

---

## Canonical libvirt device elements and mapping

- disk (file/volume/network/nvme/vhostuser)
  - Status: Partial
  - Present: `Volume`, `VolumeAttachment`
  - Gaps: libvirt `disk` has `device` (disk/cdrom/lun), `target` (dev/bus), `driver` attributes (format/queues/queue_size), `iotune`, encryption, throttling / throttlefilters, `serial`, `shareable`, and `snapshot` metadata. Current model treats storage as `Volume` and attaches to VM, but disk-specific attributes are not modeled.
  - Priority: High

- interface (type=network/bridge/hostdev/vhostuser/null/vdpa)
  - Status: Partial
  - Present: `Port`, `PortBinding`
  - Gaps: `interface` source details (network name, bridge, hostdev address, vhostuser socket), `virtualport` params, `filterref`, VLAN/tagging, `model` driver options, teaming/teaming params, `mac` with stable/currentAddress, and hotplug-related attributes.
  - Priority: High

- controller (ide/scsi/usb/virtio-serial/etc.)
  - Status: Present
  - Present: `Controller`, `ControllerAttachment`
  - Gaps: `target`/`address` subelements (PCI topology metadata), controller `driver` options and `iothread` references.
  - Priority: Medium

- console / graphics / video / framebuffers
  - Status: Partial
  - Present: `Console` (per-VM console instances)
  - Gaps: `graphics` framebuffers (listen type, tls/socket), `video` device (type/model/vram/acceleration), per-graphics `listen`/`port`/security options. `Console` models VNC/SPICE but `video` and additional display metadata are missing.
  - Priority: High (console), Medium (video/framebuffer)

- channel / serial / char devices
  - Status: Present (basic)
  - Present: `ChannelDevice`, `SerialDevice` and their attachments
  - Gaps: per-channel live `state` attributes, channel `target` deeper fields, serial `protocol` and `reconnect` options.
  - Priority: Medium

- input devices (mouse/tablet/keyboard)
  - Status: Present
  - Present: `InputDevice`, `InputDeviceAttachment`
  - Gaps: fine-grained model attributes generally small.
  - Priority: Low/Medium

- sound devices
  - Status: Present (basic)
  - Present: `SoundCard`, `SoundCardAttachment`
  - Gaps: audio backend and driver options.
  - Priority: Low

- filesystem (virtiofs/9p/share)
  - Status: Present (basic)
  - Present: `Filesystem`, `FilesystemAttachment`
  - Gaps: virtiofs socket attribute, `source` type/mode, `accessMode` and mount semantics.
  - Priority: Medium

- hostdev (usb/pci/scsi passthrough and mdev)
  - Status: Present (basic)
  - Present: `HostDevice`, `HostDeviceAttachment`
  - Gaps: PCI address breakdown, `mdev` uuid, `managed`/`mode` attributes, `driver` model, SCSI host specifics.
  - Priority: High

- usb redirector / hub devices
  - Status: Present (basic)
  - Present: `USBRedirector`, `USBRedirectorAttachment`
  - Gaps: detailed filter rules and hub topology.
  - Priority: Low

- rng device
  - Status: Present
  - Present: `RngDevice`, `RngDeviceAttachment`
  - Gaps: `backend` specifics already partially present.
  - Priority: Medium

- tpm
  - Status: Present
  - Present: `TPM`, `TPMAttachment`
  - Gaps: tpm policies and key wrapping metadata.
  - Priority: Medium

- watchdog
  - Status: Present
  - Present: `Watchdog`, `WatchdogAttachment`
  - Gaps: none critical.
  - Priority: Low

- panic device
  - Status: Present
  - Present: `PanicDevice`, `PanicDeviceAttachment`
  - Priority: Low

- rng / crypto / pstore / crypto
  - Status: Partial / Missing
  - Present: `RngDevice` (present); `Crypto`, `Pstore` are missing.
  - Priority: Low

- vsock
  - Status: Present
  - Present: `Vsock`, `VsockAttachment`
  - Priority: Low/Medium

- memory devices / virtio-mem / pmem / memoryBacking details
  - Status: Partial
  - Present: `MemoryBalloon`, `MemoryBalloonAttachment`
  - Gaps: virtio-mem, memoryBacking / hugepage / numa node fields not modeled.
  - Priority: Low/Medium

- shmem
  - Status: Present
  - Present: `ShmemDevice`, `ShmemDeviceAttachment`
  - Priority: Low

- IOMMU devices
  - Status: Present
  - Present: `IOMMUDevice`, `IOMMUDeviceAttachment`
  - Priority: Low

- vhost-user / vhostvdpa / vDPA
  - Status: Missing
  - Gaps: vhost-user socket path, reconnect options, vhostvdpa character device path, vDPA-specific capabilities.
  - Priority: Medium

- NVRAM / pstore / crypto / launchSecurity elements
  - Status: Missing
  - Priority: Low

- ACPI / NUMA / per-device address metadata
  - Status: Mostly Missing
  - Gaps: Per-device `<address>` and `<target>` element attributes (PCI slot info, controller targets, chassis, port) are not fully captured for many devices.
  - Priority: Medium

---

## Recommendations / next steps

1. Short-term (small, high value):
   - Add richer `Disk`/`DiskAttachment` or extend `VolumeAttachment` to capture disk-level attributes: `target_dev`, `target_bus`, `driver_format`, `driver_queues`, `driver_queue_size`, `iotune` JSON, `serial`, `device` type.
   - Extend `Port` / `PortBinding` to capture `source_type` (network/bridge/hostdev/vhostuser), `source_reference` (network name or hostdev address or vhost socket), `virtualport` metadata, `filterref` JSON, and VLAN tags.
   - Ensure `Console` remains canonical and add optional `Graphics`/`Video` model (if you rely on per-VM video device attributes).

2. Medium-term:
   - Add vhost-user device models (network/block/vdpa) if the environment will use vhost-user devices.
   - Expand `Filesystem` to model virtiofs socket vs mount types.
   - Capture per-device `<address>` metadata for PCI topology where relevant.

3. Longer-term:
   - Model advanced/rare elements (pstore, crypto backends, NVRAM) when needed.
   - Consider a small JSON `DeviceMetadata` blob column for low-priority attributes to avoid schema churn; promote fields to first-class columns as usage grows.

---

## Where to start in code
- `internal/storage/database.go` contains the current models and is the right place to add fields / new models.
- Follow existing naming conventions: `<Thing>` + `<Thing>Attachment` for per-VM linking, and use `AttachmentIndex` for aggregated queries.
- Keep volume behavior: volumes can be multi-attached; ensure any unique indexes on `AttachmentIndex` exclude `volume` as currently done.

---

If you want, next I'll implement a focused schema change for the top two high-priority items (disk attributes and richer interface fields) with GORM model updates and `AutoMigrate` entries, then run `go build ./...` to catch errors. Otherwise I can first open a brief PR-style patch showing the exact fields to add. 
