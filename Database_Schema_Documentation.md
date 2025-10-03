# **Virtumancer Database Schema**

Virtumancer uses a SQLite database (virtumancer.db) with a normalized relational schema to store host configurations and cache virtual machine hardware details. The schema is managed by GORM and is automatically migrated on application startup.

## **Table Definitions**

### **hosts**

Stores the connection details for each managed libvirt host.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | A user-defined, unique ID for the host. |
| name | TEXT |  | Optional friendly name for the host |
| uri | TEXT | NOT NULL | The full libvirt connection URI. |
| state | TEXT | DEFAULT "DISCONNECTED" | The current connection state of the host (CONNECTED, DISCONNECTED, ERROR). |
| task_state | TEXT |  | The current task state if a connection/disconnection operation is in progress. |
| auto_reconnect_disabled | BOOLEAN | DEFAULT false | If true, prevents automatic reconnection to this host. Set when user manually disconnects. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **virtual_machines**

The central table for virtual machines, caching their basic state and configuration.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| host_id | TEXT | NOT NULL | Foreign key to the hosts table. |
| name | TEXT | NOT NULL | The name of the VM. |
| domain_uuid | TEXT | UNIQUE, NOT NULL | The libvirt-assigned unique ID of the VM. |
| source | TEXT | DEFAULT 'managed' | Whether the VM was created ('managed') or imported ('imported'). |
| title | TEXT |  | Short domain title. |
| description | TEXT |  | A user-defined description. |
| state | TEXT | DEFAULT 'INITIALIZED' | Intended/target state of the VM. |
| libvirt_state | TEXT | DEFAULT 'INITIALIZED' | Observed state from libvirt (UNKNOWN when disconnected). |
| task_state | TEXT |  | Transient state during operations. |
| vcpu_count | INTEGER |  | Number of virtual CPUs. |
| cpu_model | TEXT |  | The configured CPU model. |
| cpu_topology_json | TEXT |  | JSON blob for sockets, cores, threads. |
| memory_bytes | INTEGER |  | Maximum memory allocated in bytes. |
| current_memory | INTEGER |  | Current memory allocation. |
| os_type | TEXT |  | Operating system type. |
| is_template | BOOLEAN |  | Whether the VM is a template. |
| metadata | TEXT |  | Custom XML metadata. |
| sync_status | TEXT | DEFAULT 'UNKNOWN' | Sync state against libvirt. |
| drift_details | TEXT |  | JSON blob storing drift information. |
| needs_rebuild | BOOLEAN | DEFAULT false | Whether VM needs rebuild from DB state. |
| created_at | DATETIME |  | Timestamp when the VM record was created. |
| updated_at | DATETIME |  | Timestamp when the VM record was last updated. |

### **discovered_vms**

Stores VMs discovered on hosts that haven't been imported into management yet.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| host_id | TEXT | NOT NULL, INDEX | Foreign key to the hosts table. |
| name | TEXT |  | The name of the discovered VM. |
| domain_uuid | TEXT | UNIQUE, INDEX | The libvirt-assigned unique ID. |
| info_json | TEXT |  | Optional serialized domain XML/metadata. |
| last_seen_at | DATETIME |  | When this VM was last observed. |
| imported | BOOLEAN | DEFAULT false, INDEX | Whether this VM has been imported. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **Storage Management**

### **storage_pools**

Represents libvirt storage pools (LVM, directories, etc.).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| host_id | TEXT |  | Foreign key to the hosts table. |
| name | TEXT |  | The name of the storage pool. |
| uuid | TEXT | UNIQUE | The libvirt-assigned UUID. |
| type | TEXT |  | The type of pool (dir, lvm, etc.). |
| path | TEXT |  | The path to the pool. |
| state | TEXT |  | Human-friendly observed pool state (e.g. "active", "inactive", "unknown"). Stored from libvirt when available. |
| capacity_bytes | INTEGER |  | Total capacity in bytes. |
| allocation_bytes | INTEGER |  | Currently allocated bytes. |
| state | TEXT | DEFAULT 'AVAILABLE' | Stable human-friendly state for the volume (AVAILABLE, IN_USE, ERROR, UNKNOWN). |
| task_state | TEXT |  | Transient task state for operations in progress (CREATING, DELETING, MIGRATING, etc.). |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **volumes**

Represents storage volumes (virtual disks, ISOs).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| storage_pool_id | TEXT |  | Foreign key to storage_pools. |
| name | TEXT | UNIQUE | The unique short name of the storage volume. Note: Virtumancer normalizes stored volume and disk names by taking the basename and stripping the last extension (e.g. `/var/lib/libvirt/images/ubuntu-20.04.qcow2` -> `ubuntu-20.04`). The original full path (when available) is preserved in the `path` column. |
| path | TEXT |  | Original full path of the volume when available (preserved for tooling, debugging, and tooltips). |
| type | TEXT |  | The type of volume (DISK, ISO). |
| format | TEXT |  | The disk format (qcow2, raw). |
| capacity_bytes | INTEGER |  | Total capacity in bytes. |
| allocation_bytes | INTEGER |  | Currently allocated bytes. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **disks**

Enhanced disk representation with detailed metadata.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| name | TEXT |  | The name of the disk. |
| volume_id | TEXT |  | Foreign key to volumes (nullable). |
| path | TEXT |  | Raw path if not in a pool. |
| format | TEXT |  | Disk format. |
| capacity_bytes | INTEGER |  | Total capacity in bytes. |
| allocation_bytes | INTEGER |  | Currently allocated bytes. |
| physical_bytes | INTEGER |  | Physical storage footprint. |
| serial | TEXT |  | Disk serial number. |
| driver_json | TEXT |  | Driver options (cache/io/...) as JSON. |
| backing_json | TEXT |  | Backing store/layered info as JSON. |
| target_path | TEXT |  | Target device path. |
| source_format | TEXT |  | Source format if different. |
| encryption | TEXT |  | Encryption information. |
| iotune | TEXT |  | I/O tuning parameters as JSON. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |
| state | TEXT | DEFAULT 'AVAILABLE' | Stable human-friendly state for the disk (AVAILABLE, IN_USE, ERROR, UNKNOWN). |
| task_state | TEXT |  | Transient task state for disk operations (CREATING, DELETING, MIGRATING, etc.). |

### **disk_attachments**

Links disks to VMs with per-VM attachment metadata.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| disk_id | TEXT |  | Foreign key to disks. |
| device_name | TEXT |  | Device name inside guest (vda). |
| bus_type | TEXT |  | Bus type (virtio, sata, ide). |
| read_only | BOOLEAN |  | Whether disk is read-only. |
| shareable | BOOLEAN |  | Whether disk is shareable. |
| address_json | TEXT |  | PCI address or target addressing. |
| metadata | TEXT |  | Optional JSON for attachment options. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **Network Management**

### **networks**

Represents virtual networks or bridges on a host.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| host_id | TEXT | UNIQUE INDEX | Foreign key to hosts table. |
| name | TEXT | UNIQUE INDEX | The name of the network. |
| uuid | TEXT |  | The libvirt-assigned UUID. |
| bridge_name | TEXT |  | The name of the host bridge interface. |
| mode | TEXT |  | The network mode (bridged, nat, isolated). |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **ports**

Represents virtual network interfaces (vNICs).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| mac_address | TEXT | UNIQUE | The unique MAC address. |
| model_name | TEXT |  | vNIC model (virtio, e1000). |
| ip_address | TEXT |  | DHCP-assigned IP address. |
| host_id | TEXT | INDEX | Optional host scoping. |
| source_type | TEXT |  | Network source type. |
| source_ref | TEXT |  | Network name, hostdev address, etc. |
| port_group | TEXT |  | Portgroup name for OpenVSwitch/VLAN. |
| virtual_port_json | TEXT |  | Serialized virtualport subelements. |
| filter_ref_json | TEXT |  | Serialized filterref subelements. |
| vlan_tags_json | TEXT |  | Serialized VLAN tags/metadata. |
| trust_guest_rx_filters | BOOLEAN |  | Trust guest RX filters. |
| primary_vlan | INTEGER |  | Nullable primary VLAN tag. |
| address_json | TEXT |  | Optional device address. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **port_bindings**

Links ports to networks.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| port_id | TEXT |  | Foreign key to ports. |
| network_id | TEXT |  | Foreign key to networks. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **port_attachments**

Links ports to VMs with per-VM attachment metadata.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| port_id | TEXT | INDEX | Foreign key to ports. |
| host_id | TEXT | INDEX | Host that the attachment is bound to. |
| device_name | TEXT |  | Per-VM device name. |
| mac_address | TEXT |  | Per-attachment MAC override. |
| model_name | TEXT |  | Per-attachment model override. |
| ordinal | INTEGER |  | Interface order. |
| metadata | TEXT |  | Optional JSON for hotplug/options. |
| address_json | TEXT |  | Optional PCI/USB address. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **Virtual Hardware Management**

### **controllers**

Represents hardware controllers (USB, SATA, virtio-serial).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| type | TEXT |  | Controller type (usb, sata, virtio-serial). |
| model_name | TEXT |  | Controller model. |
| index | INTEGER |  | Controller index. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **controller_attachments**

Links controllers to VMs.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| controller_id | INTEGER |  | Foreign key to controllers. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **input_devices**

Represents input devices (mouse, keyboard, tablet).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| type | TEXT |  | Device type (mouse, tablet, keyboard). |
| bus | TEXT |  | Bus type (usb, ps2, virtio). |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **input_device_attachments**

Links input devices to VMs.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| input_device_id | INTEGER |  | Foreign key to input_devices. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **consoles**

Represents per-VM console instances (VNC/SPICE).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| host_id | TEXT | INDEX | Foreign key to hosts. |
| type | TEXT |  | Console type (vnc, spice). |
| model_name | TEXT |  | Graphics model. |
| listen_address | TEXT |  | Listen address. |
| port | INTEGER |  | Port number. |
| tls_port | INTEGER |  | TLS port number. |
| metadata | TEXT |  | Optional JSON blob. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **Additional Device Tables**

The schema includes many additional device tables for comprehensive hardware support:

- **sound_cards** & **sound_card_attachments**: Audio devices
- **host_devices** & **host_device_attachments**: PCI/USB passthrough
- **tpms** & **tpm_attachments**: Trusted Platform Modules
- **watchdogs** & **watchdog_attachments**: Watchdog devices
- **serial_devices** & **serial_device_attachments**: Serial ports
- **channel_devices** & **channel_device_attachments**: Communication channels
- **filesystem** & **filesystem_attachments**: Shared filesystems
- **smartcard** & **smartcard_attachments**: Smartcard devices
- **usb_redirector** & **usb_redirector_attachments**: USB redirection
- **rng_device** & **rng_device_attachments**: Random number generators
- **panic_device** & **panic_device_attachments**: Panic devices
- **vsock** & **vsock_attachments**: VirtIO sockets
- **memory_balloon** & **memory_balloon_attachments**: Memory balloons
- **shmem_device** & **shmem_device_attachments**: Shared memory
- **iommu_device** & **iommu_device_attachments**: IOMMU devices

### **Performance Monitoring**

### **block_statistics**

Real-time disk performance metrics.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| disk_attachment_id | TEXT | INDEX | Foreign key to disk_attachments. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| device_name | TEXT |  | Device name. |
| read_reqs | INTEGER |  | Number of read requests. |
| read_bytes | INTEGER |  | Bytes read. |
| write_reqs | INTEGER |  | Number of write requests. |
| write_bytes | INTEGER |  | Bytes written. |
| errors | INTEGER |  | Error count. |
| allocation | INTEGER |  | Actual disk space used. |
| capacity | INTEGER |  | Total disk capacity. |
| physical | INTEGER |  | Physical size on storage. |
| collected_at | DATETIME |  | Timestamp of metrics collection. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **network_statistics**

Real-time network performance metrics.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| port_attachment_id | TEXT | INDEX | Foreign key to port_attachments. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| device_name | TEXT |  | Device name. |
| rx_bytes | INTEGER |  | Bytes received. |
| rx_packets | INTEGER |  | Packets received. |
| rx_errs | INTEGER |  | Receive errors. |
| rx_drop | INTEGER |  | Receive drops. |
| tx_bytes | INTEGER |  | Bytes transmitted. |
| tx_packets | INTEGER |  | Packets transmitted. |
| tx_errs | INTEGER |  | Transmit errors. |
| tx_drop | INTEGER |  | Transmit drops. |
| ip_address | TEXT |  | DHCP assigned IP. |
| lease_expiry | DATETIME |  | DHCP lease expiration. |
| collected_at | DATETIME |  | Timestamp of metrics collection. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **cpu_performance, memory_performance, node_performance, device_performance**

Additional performance monitoring tables for CPU, memory, node, and device metrics.

### **Video and Graphics**

### **video_models**

Video adapter templates.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| model_name | TEXT |  | Video model name (qxl, vga, virtio). |
| vram | INTEGER |  | Video RAM in bytes. |
| heads | INTEGER |  | Number of display heads. |
| accel_3d | BOOLEAN |  | 3D acceleration support. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **video_attachments**

Links video models to VMs.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm_uuid | TEXT | INDEX | Foreign key to virtual_machines. |
| video_model_id | INTEGER |  | Foreign key to video_models. |
| monitor_index | INTEGER |  | Monitor index. |
| primary | BOOLEAN |  | Whether this is the primary display. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

### **Advanced Configuration Tables**

The schema includes numerous advanced configuration tables for comprehensive VM management:

- **OS Configuration**: `os_config`, `smbios_system_info`
- **CPU Configuration**: `cpu_features`, `cpu_topology`, `cpu_tune`
- **Memory Configuration**: `memory_config`, `memory_backing`, `numa_node`
- **Security**: `security_label`, `launch_security`
- **Hypervisor Features**: `hypervisor_feature`
- **Lifecycle**: `lifecycle_action`
- **Clock**: `clock`
- **Performance**: `perf_event`
- **I/O Tuning**: `iotune`
- **QEMU Extensions**: `qemu_arg`, `vendor_option`

### **System Management**

### **users, roles, permissions**

User management and role-based access control.

### **tasks, audit_log**

Asynchronous task tracking and audit logging.

### **settings**

Key-value configuration storage with owner scoping.

### **host_capabilities**

Discovered host capabilities and features.

### **Index Tables**

### **attachment_index**

Cross-device index for fast VM attachment queries.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | Auto-generated UUID primary key. |
| vm_uuid | TEXT | INDEX, NOT NULL | Foreign key to virtual_machines. |
| device_type | TEXT | INDEX, NOT NULL | Device type (volume, graphics, hostdevice, etc.). |
| attachment_id | TEXT | NOT NULL | Row ID in the specific attachment table. |
| device_id | TEXT | INDEX | Optional device UUID. |
| created_at | DATETIME |  | Timestamp of creation. |
| updated_at | DATETIME |  | Timestamp of last update. |

This comprehensive schema supports full libvirt domain XML representation and enables advanced VM management features including real-time monitoring, drift detection, and automated configuration management.
Represents a virtual network interface (vNIC).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm\_id | INTEGER |  | Foreign key to virtual\_machines. |
| mac\_address | TEXT | UNIQUE | The unique MAC address of the vNIC. |
| model\_name | TEXT |  | The vNIC model, e.g., virtio. |
| port\_group | TEXT |  | The portgroup name for OpenVSwitch or bridged networks. Used for network policy and VLAN configuration. |

### **port\_bindings**

A join table linking a port to a network.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| port\_id | INTEGER |  | Foreign key to ports. |
| network\_id | INTEGER |  | Foreign key to networks. |

### **graphics\_devices**

Represents a graphical console device type.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| type | TEXT |  | The type of console, e.g., vnc, spice. |
| model\_name | TEXT |  | The graphics model, e.g., qxl. |
| vram\_kib | INTEGER |  | (Future Use) Video RAM in KiB. |
| listen\_address | TEXT |  | (Future Use) The listen address. |

### **graphics\_device\_attachments**

A join table linking a virtual\_machine to a graphics\_device.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm\_id | INTEGER |  | Foreign key to virtual\_machines. |
| graphics\_device\_id | INTEGER |  | Foreign key to graphics\_devices. |


