# **Virtumancer Database Schema**

Virtumancer uses a SQLite database (virtumancer.db) with a normalized relational schema to store host configurations and cache virtual machine hardware details. The schema is managed by GORM and is automatically migrated on application startup.

## **Table Definitions**

### **hosts**

Stores the connection details for each managed libvirt host.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | TEXT | PRIMARY KEY | A user-defined, unique ID for the host. |
| uri | TEXT | NOT NULL | The full libvirt connection URI. |
| state | TEXT | DEFAULT "DISCONNECTED" | The current connection state of the host (CONNECTED, DISCONNECTED, ERROR). |
| task_state | TEXT |  | The current task state if a connection/disconnection operation is in progress. |
| auto_reconnect_disabled | BOOLEAN | DEFAULT false | If true, prevents automatic reconnection to this host. Set when user manually disconnects. |
| created\_at | DATETIME |  | Timestamp of creation. |
| updated\_at | DATETIME |  | Timestamp of last update. |

### **virtual\_machines**

The central table for virtual machines, caching their basic state and configuration.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| host\_id | TEXT | NOT NULL | Foreign key to the hosts table. |
| uuid | TEXT | UNIQUE, NOT NULL | The libvirt-assigned unique ID of the VM. |
| name | TEXT |  | The name of the VM. |
| description | TEXT |  | A user-defined description. |
| vcpu\_count | INTEGER |  | Number of virtual CPUs. |
| memory\_bytes | INTEGER |  | Maximum memory allocated in bytes. |
| state | INTEGER |  | The last known power state from libvirt. |
| is\_template | BOOLEAN |  | (Future Use) If the VM is a template. |
| cpu\_model | TEXT |  | (Future Use) The configured CPU model. |
| cpu\_topology\_json | TEXT |  | (Future Use) JSON blob for sockets, cores, threads. |
| created\_at | DATETIME |  | Timestamp when the VM record was created. |
| updated\_at | DATETIME |  | Timestamp when the VM record was last updated. |

*Note: All timestamps are managed by GORM's automatic timestamping feature and are properly displayed in the frontend UI.*

### **volumes**

Represents storage volumes (virtual disks, ISOs).

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| name | TEXT | UNIQUE | The unique name/path of the storage volume. |
| type | TEXT |  | The type of volume, e.g., DISK, ISO. |
| format | TEXT |  | The disk format, e.g., qcow2, raw. |

### **volume\_attachments** (Removed)

A join table that previously linked virtual\_machines to volumes. This table has been removed as part of the disk schema migration. All VM disk attachments now use the disk\_attachments table.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| vm\_id | INTEGER |  | Foreign key to virtual\_machines. |
| volume\_id | INTEGER |  | Foreign key to volumes. |
| device\_name | TEXT |  | The device name inside the guest, e.g., vda. |
| bus\_type | TEXT |  | The bus type, e.g., virtio, sata. |

*Note: This table was dropped after successful migration to disk_attachments.*

### **disk\_attachments** (Current)

The current table for linking virtual machines to disk storage with enhanced metadata.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key with GORM timestamps. |
| vm\_id | INTEGER |  | Foreign key to virtual\_machines. |
| volume\_id | INTEGER |  | Foreign key to volumes. |
| device\_name | TEXT |  | The device name inside the guest, e.g., vda. |
| bus\_type | TEXT |  | The bus type, e.g., virtio, sata. |
| size\_bytes | INTEGER |  | The size of the disk in bytes for display and management purposes. |
| created\_at | DATETIME |  | Timestamp when the attachment was created. |
| updated\_at | DATETIME |  | Timestamp when the attachment was last updated. |

*Note: Disk size calculation now uses only the disk_attachments table with Disk.CapacityBytes.*

### **networks**

Represents virtual networks on a host.

| Column | Type | Constraints | Description |
| :---- | :---- | :---- | :---- |
| id | INTEGER | PRIMARY KEY | Auto-incrementing primary key. |
| host\_id | TEXT | NOT NULL | Foreign key to the hosts table. |
| uuid | TEXT | UNIQUE | The libvirt-assigned UUID (can be empty). |
| name | TEXT |  | The name of the network. |
| bridge\_name | TEXT |  | The name of the host bridge interface. |
| mode | TEXT |  | The network mode, e.g., bridged. |

*Note: A UNIQUE constraint exists on the combination of (host\_id, name).*

### **ports**

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


