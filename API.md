# **Virtumancer API Documentation**

Virtumancer exposes a RESTful HTTP API for management operations and a WebSocket API for real-  * **Valid actions**: start, shutdown, reboot, destroy (force off), reset (force reset).  
* **Response**: 204 No Content

### **Discovered VM Management**

#### **GET /api/v1/hosts/:hostId/discovered-vms**

* **Description**: Retrieves a list of discovered VMs on a host that are not yet imported into Virtumancer management.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK  
  \[  
    {  
      "domain_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",  
      "name": "discovered-vm-01",  
      "host_id": "kvmsrv",  
      "last_seen_at": "2023-10-27T12:00:00Z",  
      "imported": false  
    }  
  \]

#### **POST /api/v1/hosts/:hostId/vms/:vmName/import**

* **Description**: Imports a single discovered VM into Virtumancer management.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine to import.  
* **Response**: 202 Accepted

#### **POST /api/v1/hosts/:hostId/vms/import-all**

* **Description**: Imports all discovered VMs on a host into Virtumancer management.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 202 Accepted

#### **POST /api/v1/hosts/:hostId/vms/import-selected**

* **Description**: Imports selected discovered VMs by their domain UUIDs into Virtumancer management.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Request Body**:  
  {  
    "domain_uuids": \["f47ac10b-58cc-4372-a567-0e02b2c3d479", "a1b2c3d4-5678-90ab-cdef-1234567890ab"\]  
  }  
* **Response**: 202 Accepted

#### **DELETE /api/v1/hosts/:hostId/discovered-vms**

* **Description**: Removes selected discovered VMs from the database by their domain UUIDs.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Request Body**:  
  {  
    "domain_uuids": \["f47ac10b-58cc-4372-a567-0e02b2c3d479", "a1b2c3d4-5678-90ab-cdef-1234567890ab"\]  
  }  
* **Response**: 202 Accepted

### **Storage Management**

#### **GET /api/v1/storage/pools**

* **Description**: Retrieves a list of all storage pools across all hosts.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "pool-uuid",  
      "host_id": "kvmsrv",  
      "name": "default",  
      "type": "dir",  
      "path": "/var/lib/libvirt/images",  
      "capacity_bytes": 100000000000,  
      "allocation_bytes": 50000000000  
    }  
  \]

#### **GET /api/v1/storage/volumes**

* **Description**: Retrieves a list of all storage volumes across all hosts.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "vol-uuid",  
      "storage_pool_id": "pool-uuid",  
      "name": "ubuntu-20.04.qcow2",  
      "type": "DISK",  
      "format": "qcow2",  
      "capacity_bytes": 21474836480,  
      "allocation_bytes": 8589934592  
    }  
  \]

#### **GET /api/v1/storage/disk-attachments**

* **Description**: Retrieves a list of all disk attachments across all VMs.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "attachment-uuid",  
      "vm_uuid": "vm-uuid",  
      "disk_id": "disk-uuid",  
      "device_name": "vda",  
      "bus_type": "virtio",  
      "size_bytes": 21474836480  
    }  
  \]

#### **GET /api/v1/hosts/:hostId/storage/pools**

* **Description**: Retrieves storage pools for a specific host.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK (same format as global pools endpoint)

#### **GET /api/v1/hosts/:hostId/storage/volumes**

* **Description**: Retrieves storage volumes for a specific host.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK (same format as global volumes endpoint)

### **Network Management**

#### **GET /api/v1/networks**

* **Description**: Retrieves a list of all networks across all hosts.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "network-uuid",  
      "host_id": "kvmsrv",  
      "name": "default",  
      "uuid": "network-uuid",  
      "bridge_name": "virbr0",  
      "mode": "nat"  
    }  
  \]

#### **GET /api/v1/ports**

* **Description**: Retrieves a list of all network ports across all VMs.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "port-uuid",  
      "mac_address": "52:54:00:11:22:33",  
      "model_name": "virtio",  
      "host_id": "kvmsrv"  
    }  
  \]

#### **GET /api/v1/port-attachments**

* **Description**: Retrieves a list of all port attachments across all VMs.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "attachment-uuid",  
      "vm_uuid": "vm-uuid",  
      "port_id": "port-uuid",  
      "device_name": "eth0"  
    }  
  \]

#### **GET /api/v1/hosts/:hostId/networks**

* **Description**: Retrieves networks for a specific host.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK (same format as global networks endpoint)

#### **GET /api/v1/hosts/:hostId/ports**

* **Description**: Retrieves network ports for a specific host.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "port-uuid",  
      "vm_uuid": "vm-uuid",  
      "port_id": "port-uuid",  
      "device_name": "eth0",  
      "mac_address": "52:54:00:11:22:33"  
    }  
  \]

#### **GET /api/v1/hosts/:hostId/vms/:vmName/port-attachments**

* **Description**: Retrieves port attachments for a specific VM.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Response**: 200 OK (same format as global port attachments endpoint)

### **Video/Graphics Management**

#### **GET /api/v1/video/models**

* **Description**: Retrieves a list of all known video model templates.  
* **Response**: 200 OK  
  \[  
    {  
      "id": 1,  
      "model_name": "qxl",  
      "vram": 65536,  
      "heads": 1,  
      "accel_3d": false  
    }  
  \]

#### **GET /api/v1/hosts/:hostId/video/devices**

* **Description**: Retrieves physical video devices discovered on a host.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
* **Response**: 200 OK  
  \[  
    {  
      "id": 1,  
      "host_device_id": 1,  
      "vendor": "NVIDIA",  
      "model_name": "GeForce RTX 3080",  
      "uuid": "GPU-uuid"  
    }  
  \]

#### **GET /api/v1/hosts/:hostId/vms/:vmName/video-attachments**

* **Description**: Retrieves video attachments for a specific VM.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Response**: 200 OK  
  \[  
    {  
      "id": 1,  
      "vm_uuid": "vm-uuid",  
      "video_model_id": 1,  
      "monitor_index": 0,  
      "primary": true  
    }  
  \]

### **Console Management**

#### **GET /api/v1/hosts/:hostId/vms/:vmName/console**

* **Description**: Establishes a console connection for a VM (VNC/SPICE proxy).  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Response**: WebSocket upgrade or proxy connection

#### **GET /api/v1/hosts/:hostId/vms/:vmName/spice**

* **Description**: Establishes a SPICE console connection for a VM.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Response**: SPICE protocol connection

### **Dashboard & Statistics**

#### **GET /api/v1/dashboard/stats**

* **Description**: Retrieves aggregated system-wide statistics for the dashboard.  
* **Response**: 200 OK  
  {  
    "total_hosts": 3,  
    "connected_hosts": 2,  
    "total_vms": 15,  
    "active_vms": 12,  
    "total_storage_gb": 500,  
    "used_storage_gb": 250  
  }

#### **GET /api/v1/dashboard/activity**

* **Description**: Retrieves recent system activity events.  
* **Query Parameters**:  
  * limit (integer): Maximum number of activities to return (default: 10).  
* **Response**: 200 OK  
  \[  
    {  
      "id": 1,  
      "timestamp": "2023-10-27T15:30:00Z",  
      "type": "vm_started",  
      "message": "VM ubuntu-vm-01 started on host kvmsrv",  
      "host_id": "kvmsrv",  
      "vm_name": "ubuntu-vm-01"  
    }  
  \]

#### **GET /api/v1/dashboard/overview**

* **Description**: Retrieves combined dashboard data for initial page load.  
* **Response**: 200 OK  
  {  
    "stats": { ... },  
    "recent_activity": [ ... ]  
  }

### **Settings & Configuration**

#### **GET /api/v1/settings/metrics**

* **Description**: Retrieves persistent metrics settings (smoothing, units, CPU defaults).  
* **Response**: 200 OK  
  {  
    "cpu_smoothing_window": 10,  
    "memory_display_unit": "GB",  
    "default_cpu_view": "percentage"  
  }

#### **PUT /api/v1/settings/metrics**

* **Description**: Updates and persists metrics settings.  
* **Request Body**:  
  {  
    "cpu_smoothing_window": 15,  
    "memory_display_unit": "MB"  
  }  
* **Response**: 200 OK

#### **GET /api/v1/settings/metrics/runtime**

* **Description**: Retrieves the current runtime metrics settings from the host service.  
* **Response**: 200 OK (same format as persistent settings)

### **Health Check**

#### **GET /api/v1/health**

* **Description**: Basic health check endpoint.  
* **Response**: 200 OK  
  {  
    "ok": true  
  }

## **WebSocket API** updates and monitoring.

## **REST API**

Base URL: /api/v1

### **Host Management**

#### **GET /api/v1/hosts**

* **Description**: Retrieves a list of all configured hosts from the database.  
* **Response**: 200 OK  
  \[  
    {  
      "id": "kvmsrv",  
      "uri": "qemu+ssh://user@host/system",  
      "state": "CONNECTED",  
      "auto_reconnect_disabled": false,  
      "created\_at": "2023-10-27T10:00:00Z"  
    }  
  \]

#### **POST /api/v1/hosts**

* **Description**: Adds a new host, connects to it, and stores it in the database.  
* **Request Body**:  
  {  
    "id": "new-kvm-host",  
    "uri": "qemu+ssh://user@new-host/system"  
  }

* **Response**: 200 OK on success, with the created host object. 500 Internal Server Error if the connection fails.

#### **POST /api/v1/hosts/:id/connect**

* **Description**: Manually connects to a previously disconnected host. This will succeed even if auto-reconnection was previously disabled by a user disconnect.  
* **URL Parameters**:  
  * id (string): The ID of the host to connect.  
* **Response**: 200 OK on success. 500 Internal Server Error if the connection fails.

#### **POST /api/v1/hosts/:id/disconnect**

* **Description**: Disconnects from a host and marks it so that automatic reconnection is disabled. This prevents the system from automatically reconnecting to the host until manually reconnected.  
* **URL Parameters**:  
  * id (string): The ID of the host to disconnect.  
* **Response**: 200 OK on success. 500 Internal Server Error if the disconnection fails.

#### **DELETE /api/v1/hosts/:id**

* **Description**: Disconnects from a host and removes it from the database.  
* **URL Parameters**:  
  * id (string): The ID of the host to remove.  
* **Response**: 204 No Content

#### **GET /api/v1/hosts/:id/info**

* **Description**: Retrieves real-time information and statistics about a specific host (CPU, memory, etc.).  
* **URL Parameters**:  
  * id (string): The ID of the host.  
* **Response**: 200 OK  
  {  
    "hostname": "kvm-host-01",  
    "cpu": 8,  
    "memory": 16777216000,  
    "cores": 4,  
    "threads": 2  
  }

### **Virtual Machine Management**

#### **GET /api/v1/hosts/:id/vms**

* **Description**: Retrieves a list of all virtual machines on a specific host from the local database cache.  
* **URL Parameters**:  
  * id (string): The ID of the host.  
* **Response**: 200 OK  
  \[  
    {  
      "db\_id": 1,  
      "name": "ubuntu-vm-01",  
      "description": "",  
      "vcpu\_count": 2,  
      "memory\_bytes": 2147483648,  
      "state": 1,  
      "graphics": {  
        "vnc": true,  
        "spice": false  
      },
      "created\_at": "2023-10-27T10:30:00Z",
      "updated\_at": "2023-10-27T15:45:00Z"
    }  
  \]

*Note: The `created_at` and `updated_at` timestamps are now properly managed by GORM and correctly displayed in the frontend UI. These timestamps track when VM records are created and last modified in the database.*

#### **GET /api/v1/hosts/:hostId/vms/:vmName/hardware**

* **Description**: Retrieves the hardware configuration for a specific VM. This triggers a fresh sync from libvirt before returning the cached data.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Response**: 200 OK  
  {  
    "disks": \[  
      {  
        "type": "file",  
        "device": "disk",  
        "driver": { "driver\_name": "qemu", "type": "qcow2" },  
        "path": "/path/to/disk.qcow2",  
        "target": { "dev": "vda", "bus": "virtio" },
        "size\_bytes": 21474836480  
      }  
    \],

*Note: The `size_bytes` field provides the actual disk size in bytes for display purposes. The system calculates this from both current disk_attachments and legacy volume_attachments tables for comprehensive coverage.*  
    "networks": \[  
      {  
        "type": "bridge",  
        "mac": { "address": "52:54:00:11:22:33" },  
        "source": { "bridge": "br0", "portgroup": "vlan100" },  
        "model": { "model\_type": "virtio" }  
      }  
    \]

*Note: The `portgroup` field in the source object is optional and indicates the OpenVSwitch portgroup or bridge VLAN configuration for the network interface. This field is used for network policy and VLAN segmentation.*  
  }

#### **POST /api/v1/hosts/:hostId/vms/:vmName/action**

* **Description**: Performs a power action on a specific VM.  
* **URL Parameters**:  
  * hostId (string): The ID of the host.  
  * vmName (string): The name of the virtual machine.  
* **Request Body**:  
  {  
    "action": "start"  
  }

  * **Valid actions**: start, shutdown, reboot, destroy (force off), reset (force reset).  
* **Response**: 204 No Content

## **WebSocket API**

The WebSocket API is used for real-time notifications and statistics monitoring.

* **Connection URL**: /ws

### **Client-to-Server Messages**

Messages are sent as JSON objects with type and payload fields.

#### **subscribe-vm-stats**

* **Description**: Subscribes the client to real-time statistics updates for a specific VM. The server will start polling the VM and broadcasting vm-stats-updated messages.  
* **Payload**:  
  {  
    "type": "subscribe-vm-stats",  
    "payload": {  
      "hostId": "kvmsrv",  
      "vmName": "ubuntu-vm-01"  
    }  
  }

#### **unsubscribe-vm-stats**

* **Description**: Unsubscribes the client from a VM's statistics updates. If no clients are left subscribed, the server will stop polling.  
* **Payload**:  
  {  
    "type": "unsubscribe-vm-stats",  
    "payload": {  
      "hostId": "kvmsrv",  
      "vmName": "ubuntu-vm-01"  
    }  
  }

### **Server-to-Client Messages**

#### **hosts-changed**

* **Description**: Sent whenever a host is added or removed. The client should re-fetch the list of hosts via GET /api/hosts.  
* **Payload**: null

#### **vms-changed**

* **Description**: Sent whenever the list of VMs on a host has changed (e.g., a VM was added, removed, or its state changed after a power operation). The client should re-fetch the VM list for the specified host.  
* **Payload**:  
  {  
    "type": "vms-changed",  
    "payload": {  
      "hostId": "kvmsrv"  
    }  
  }

#### **vm-stats-updated**

* **Description**: Broadcast periodically to all subscribed clients for a specific VM.  
* **Payload**:  
  {  
    "type": "vm-stats-updated",  
    "payload": {  
      "hostId": "kvmsrv",  
      "vmName": "ubuntu-vm-01",  
      "stats": {  
        "state": 1,  
        "memory": 2097152,  
        "max\_mem": 2097152,  
        "vcpu": 2,  
        "cpu\_time": 1234567890,  
        "disk\_stats": \[  
          { "device": "vda", "read\_bytes": 1024, "write\_bytes": 2048 }  
        \],  
        "net\_stats": \[  
          { "device": "vnet0", "read\_bytes": 4096, "write\_bytes": 8192 }  
        \]  
      }  
    }  
  }

## **Network Topology Visualization**

The Virtumancer frontend provides a comprehensive network topology view that visualizes the infrastructure in two modes:

### **Grid View**
- Displays hosts as cards with detailed VM information
- Shows real-time statistics (total hosts, VMs, connected hosts, active VMs)
- Interactive host and VM cards with navigation links
- Status indicators with color coding

### **Network Diagram View**
- Interactive SVG-based network topology
- Circular layout with hosts positioned around the center
- VM nodes positioned around their respective hosts
- Color-coded status indicators for hosts and VMs
- Interactive elements with click navigation
- Legend showing status color mappings

### **Status Color Coding**
- **Connected/Active**: Green (#10b981)
- **Disconnected/Stopped**: Red (#ef4444)
- **Paused/Warning**: Yellow (#f59e0b)
- **Error**: Dark Red (#dc2626)

### **Data Sources**
The network topology view uses the following API endpoints:
- `GET /api/v1/hosts` - Retrieves all hosts for topology display
- `GET /api/v1/hosts/:id/vms` - Fetches VMs for each connected host

### **Real-time Updates**
The topology view automatically updates based on WebSocket events:
- `hosts-changed` - Triggers refresh of host topology data
- `vms-changed` - Updates VM display for the affected host

For detailed implementation information, see `docs/network-topology-view-documentation.md`.  

