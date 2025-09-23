# **Virtumancer**
Code written by AI

**The ultimate manager for libvirt, built with Go and Vue.js.**  
Virtumancer is a modern, web-based virtualization management platform designed for simplicity, performance, and power. It leverages a Go backend, a Vue.js frontend, and the battle-tested libvirt API to provide a seamless experience for managing virtual machines.

## **Core Architecture**

* **Backend**: A high-performance Go application that serves a RESTful API and handles real-time communication via WebSockets.  
* **Libvirt Connector**: Uses the pure-Go github.com/digitalocean/go-libvirt library to communicate with libvirt daemons, supporting both local and remote hosts via secure SSH tunneling.  
* **Frontend**: A reactive and intuitive user interface built with Vue.js and Tailwind CSS.  
* **Database**: A self-contained SQLite database for storing host configurations and caching VM metadata. The schema is automatically migrated on startup.

## **Current Features**

* **Multi-Host Management**: Connect to and manage multiple libvirt hosts from a single interface.  
* **Secure Connections**: First-class support for qemu+ssh URIs using native SSH tunneling for secure, agentless remote management.  
* **Host Connection Control**: Manually connect/disconnect hosts with persistent user disconnect preferences that prevent unwanted automatic reconnections.  
* **VM Lifecycle Management**: Start, stop, shutdown, reboot, and force-reset virtual machines.  
* **Real-Time Monitoring**: Live-stream CPU, memory, and I/O statistics for running VMs directly to the UI.  
* **Normalized Datastore**: VM hardware configurations are discovered and stored in a structured, relational database, enabling powerful future features.  
* **Automatic Discovery & Sync**: Automatically synchronizes the state of all VMs with the central database.

## **Step-by-Step Tutorial & Setup**

### **Prerequisites**

* **Go**: Version 1.23 or newer.  
* **Node.js**: Version 18 or newer (for frontend development).  
* **Libvirt**: A running libvirt daemon on the hosts you wish to manage.

### **Backend Setup**

1. **Clone the repository:**  
   git clone \<your-repo-url\>  
   cd virtumancer

2. **Install dependencies:**  
   go mod tidy

3. **Run the application:**  
   go run main.go

   The backend server will start, typically on http://localhost:8080. The first run will automatically create and migrate the virtumancer.db SQLite database file in the root directory.

### **Frontend Setup**

1. **Navigate to the web directory:**  
   cd web

2. **Install dependencies:**  
   npm install

3. **Run the development server:**  
   npm run dev

   The frontend will be accessible at http://localhost:5173 and will automatically proxy API requests to the backend.

### **Host Configuration for Remote Access (qemu+ssh)**

For Virtumancer to connect to a remote host, the user running the Virtumancer backend must have **passwordless SSH access** to the target host.

1. **Generate an SSH key** on the machine running Virtumancer if you don't have one:  
   ssh-keygen \-t rsa \-b 4096

2. **Copy the public key** to the remote libvirt host. Replace user and remote-host-ip.  
   ssh-copy-id user@remote-host-ip

3. **Test the connection**: You should be able to SSH into the remote host without a password.  
   ssh user@remote-host-ip

4. Add the host in the Virtumancer UI: Use the qemu+ssh URI format, for example:  
   qemu+ssh://user@remote-host-ip/system

## **Project Directory Tree**

.  
├── API.md                      \# Detailed API documentation.  
├── Database\_Schema\_Documentation.md \# Detailed database schema.  
├── README.md                   \# This file.  
├── go.mod                      \# Go module definition.  
├── go.sum                      \# Go module checksums.  
├── internal/  
│   ├── api/  
│   │   └── handlers.go         \# HTTP request handlers for the REST API.  
│   ├── console/  
│   │   └── proxy.go            \# Websocket proxy for VNC/SPICE consoles.  
│   ├── libvirt/  
│   │   └── connector.go        \# Manages connections to libvirt hosts via SSH/TCP.  
│   ├── services/  
│   │   └── host\_service.go     \# Core business logic for managing hosts and VMs.  
│   ├── storage/  
│   │   └── database.go         \# GORM models and database initialization.  
│   └── ws/  
│       ├── client.go           \# Represents a single WebSocket client.  
│       └── hub.go              \# Manages all active WebSocket clients and broadcasting.  
├── main.go                     \# Application entry point, sets up server and routes.  
├── virtumancer.db              \# SQLite database file (auto-generated).  
└── web/                        \# Vue.js frontend source code.  
    ├── public/  
    ├── src/  
    ├── index.html  
    └── package.json  

## **Documentation**

### **API Documentation**
- **API.md**: Complete REST and WebSocket API reference
- **Database_Schema_Documentation.md**: Database schema and relationships

### **Frontend Documentation**
- **Frontend_UIUX_plan.md**: UI/UX design philosophy and implementation plan
- **docs/network-topology-view-documentation.md**: Network topology visualization component
- **docs/discovered-vm-bulk-management.md**: Bulk operations for discovered VM management

### **Technical Documentation**
- **docs/technology-stack-analysis.md**: Technology choices and architecture decisions
- **docs/design-system-specification.md**: Design system and component guidelines
- **docs/db-schema-expansion-plan.md**: Database evolution planning

## **Key Features**

### **Network Topology Visualization**
Virtumancer provides a comprehensive network topology view with two visualization modes:
- **Grid View**: Host cards with detailed VM information and real-time statistics
- **Network Diagram**: Interactive SVG-based topology with circular layout and status indicators
- **Real-time Updates**: Automatic refresh based on infrastructure changes via WebSockets
- **Interactive Navigation**: Click-to-navigate between hosts and VMs

### **Discovered VM Bulk Management**
Comprehensive tools for managing multiple unmanaged VMs discovered on hosts:
- **Multi-selection Interface**: Checkbox-based selection with select-all functionality
- **Bulk Import Operations**: Import multiple VMs into management simultaneously
- **Bulk Delete Operations**: Remove multiple discovered VMs from tracking
- **Search and Filtering**: Real-time search by name or UUID with flexible sorting
- **Progress Indicators**: Visual feedback during bulk operations with error handling

### **Modern UI/UX**
- Glass morphism design with backdrop blur effects
- Neon glow states based on resource status
- Responsive layout for all device sizes
- Smooth animations and transitions
- Dark theme optimized for extended use
