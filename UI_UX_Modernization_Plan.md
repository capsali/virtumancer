VirtuMancer UI/UX Modernization Plan

This document outlines a plan to evolve the VirtuMancer web interface by defining a unique, modern identity that prioritizes speed, efficiency, and a superior user experience, drawing lessons from—but not merely replicating—industry standards like Proxomox and vSphere.
1. Defining the VirtuMancer Design Philosophy

Instead of blending existing UIs, we will establish our own: "Command-First, Context-Aware."

    Command-First: The primary way for an experienced user to interact with the system should be through a powerful, fast, and intelligent Command Palette. Most actions (start/stop VM, create snapshot, open console) should be achievable without touching the mouse. This caters to power users who value speed above all else.

    Context-Aware: The graphical interface should intelligently adapt to the user's current selection, presenting the most relevant information and actions in a clean, uncluttered way. The GUI serves as a visual confirmation and a discovery tool for less common tasks, while the Command Palette handles the high-frequency operations.

2. Competitive Analysis (Re-evaluation)
Proxmox VE

    Insight: Its strength is not its appearance, but its information-dense, predictable tree structure. We will retain this hierarchical view as a core navigation element.

    Our Takeaway: Keep the tree, but make it smarter. Add live status badges, context menus, and make it filterable via the Command Palette.

VMware vSphere

    Insight: Its strength is guiding users through complex, infrequent tasks with wizards and clear layouts.

    Our Takeaway: For multi-step operations like VM creation or storage configuration, we will adopt a similar guided, card-based wizard approach.

3. The New VirtuMancer Layout
A. The Universal Command Palette (Ctrl+K)

This is the centerpiece of our new design. It will be a floating modal that can be summoned from anywhere.

    Capabilities:

        Fuzzy Search: Instantly find any Host or VM by name.

        Action Execution: Type commands like start web-server-01, snapshot db-server, console mail-gw.

        Navigation: Typing host-01 > summary will navigate directly to that view.

B. Main Navigation (Left Pane)

    A minimalist, collapsible sidebar containing the navigation tree.

    The tree will display Hosts and their nested Virtual Machines.

    Each item will feature a live status indicator.

    The entire tree will be searchable and filterable, with its results driven by the Command Palette.

C. Main Content (Right Pane)

This area remains context-aware but will be redesigned for clarity and actionability.

    Host Dashboard:

        Header: Clean title with primary host actions (e.g., "Add VM," "Reboot Host").

        Grid of Live Metric Cards: Large, easy-to-read charts for CPU, Memory, IO, and Network, updating in real-time.

        VM Table: A streamlined, searchable table of VMs on the host, focusing only on critical data: Name, Status, CPU/Mem Usage, IP Address. Actions will be in a context menu to reduce clutter.

    VM Management View:

        Header: Prominently displays the VM name, its status, and primary controls (Start, Stop, Reboot, Console menu).

        Tabbed Interface: The tabs remain, but the content will be organized into clean, well-spaced cards.

            Summary: A dashboard of cards for this specific VM, including live resource charts, guest agent info, and notes.

            Console: This tab will house the embedded VNC or SPICE console, providing direct, in-app access to the VM's screen without needing a new browser tab.

            Hardware: A clean, editable list of virtual hardware (CPU, Memory, Disks, NICs).

            Snapshots: A user-friendly manager for creating, viewing, and reverting snapshots.

            Options: Settings like boot order and autostart behavior.

4. Redefined Modernization Features

    Command Palette: As described above, this is now our core interaction model.

    Adaptive Information Density: The UI will be clean by default, but we will provide a "density" toggle that reduces whitespace and padding, bringing it closer to the Proxmox feel for users who prefer it.

    Keyboard-First Navigation: In addition to the Command Palette, users will be able to navigate the entire UI using arrow keys, tabs, and shortcuts.

    Integrated Real-Time Notifications: All actions will trigger toast notifications, which will also be logged in a persistent "Tasks" panel for later review.

5. Next Steps

Implement the foundational layout: the main application shell, the new collapsible sidebar, and, most importantly, the initial version of the Command Palette.
