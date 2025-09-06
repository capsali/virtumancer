Virtumancer Frontend UI/UX Plan

This document outlines the design philosophy, layout, and feature implementation plan for the Virtumancer web interface. Our goal is to create an experience that is more intuitive, efficient, and aesthetically pleasing than existing solutions like Proxmox and vSphere.
1. Design Philosophy & Core Principles

    Clarity and Intuitiveness: The UI should be easy to navigate for both new and experienced users. Information should be presented clearly, and actions should be predictable.

    Information Density: Like Proxmox, we need to display a lot of information compactly, but we will use modern design techniques (e.g., better spacing, typography, progressive disclosure) to avoid feeling cluttered.

    Responsiveness & Speed: The interface must be fast and responsive. Actions should feel instantaneous, leveraging the real-time capabilities of our WebSocket backend.

    Modern Aesthetics: We will use a clean, dark-themed interface with consistent iconography, typography, and spacing. Tailwind CSS will be our utility-first framework to achieve this.

    Context-Aware Actions: The UI will only present actions that are relevant to the selected resource's current state (e.g., only show the "Start" button for a stopped VM).

2. Overall Layout

The UI will be a classic three-pane layout, which is proven for this type of application.

    Left Pane (Collapsible Sidebar):

        Purpose: Hierarchical navigation of all managed resources.

        Structure: A tree view displaying a "Datacenter" root, followed by each connected Host. Under each Host, we'll list the VMs.

        Interactivity: Clicking on a host or VM will update the Main Content Pane to show the relevant dashboard or details. The sidebar will be collapsible to maximize content space.

    Top Pane (Header):

        Purpose: Global actions, notifications, and user information.

        Content:

            Sidebar toggle button.

            Breadcrumb navigation showing the current location (e.g., Datacenter > server-01 > vm-101).

            A global search bar (future feature).

            A notification icon/area for alerts and long-running task status.

            User/authentication status.

    Center Pane (Main Content Area):

        Purpose: Display detailed information and management options for the resource selected in the sidebar.

        Content: This area will be tab-based to organize information cleanly. For example, when a VM is selected, the tabs could be:

            Summary: Key stats (CPU, RAM, Disk I/O, Network), guest info, and primary lifecycle actions.

            Console: Embedded VNC/SPICE console.

            Hardware: Virtual hardware configuration (disks, NICs, etc.).

            Snapshots: Manage VM snapshots.

            Logs: Display logs for the specific VM.

3. Phased Implementation Plan

We will build the UI in logical phases, starting with the core functionality and iterating to add more advanced features.
Phase 1: Foundation & Core VM Management (Current State -> Near Future)

    Objective: Solidify the existing layout and enhance the core VM management experience.

    Tasks:

        Refine Sidebar:

            Implement a proper tree-view structure (Datacenter -> Host -> VMs).

            Add visual cues for VM state (e.g., green icon for running, red for stopped).

            Ensure the selected item is clearly highlighted.

        Enhance Host Dashboard (HostDashboard.vue):

            Display summary statistics for the host (CPU usage, memory usage, storage overview).

            The VM list should be a more detailed table or card view with key info (State, vCPUs, Memory).

        Improve VM View (VmView.vue):

            Implement the tabbed interface (Summary, Console, etc.).

            The Summary tab should show real-time performance graphs (CPU, Memory, Network I/O). We can start with simple value displays and add graphs later.

            Make the lifecycle action buttons (Start, Stop, etc.) context-aware and provide visual feedback (e.g., loading spinners).

Phase 2: Resource Creation & Editing

    Objective: Allow users to create and modify resources through the UI.

    Tasks:

        Create VM Wizard: A multi-step modal or dedicated page to guide users through creating a new virtual machine.

            Step 1: General (Host, Name, OS Type).

            Step 2: CPU & Memory.

            Step 3: Storage (Create/select virtual disk).

            Step 4: Network (Select bridge/network).

            Step 5: Confirmation.

        Edit VM Hardware: An interface within the VM's "Hardware" tab to add, remove, or modify virtual disks, network interfaces, etc.

Phase 3: Advanced Features

    Objective: Introduce more complex management features.

    Tasks:

        Storage Management: A dedicated view to manage storage pools on each host (e.g., LVM, ZFS, Directories).

        Network Management: A view for managing virtual networks and bridges.

        Snapshot Management: A full-featured snapshot manager within the VM view (create, delete, revert).

        User Roles & Permissions: (Future) Integrate an authentication system and define roles.

This plan provides a clear roadmap. The next immediate step is to begin implementing Phase 1, starting with refining the sidebar and the host dashboard.
