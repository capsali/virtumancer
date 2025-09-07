# Virtumancer Database Documentation

This document outlines the database schema used by Virtumancer. The application uses a single SQLite database file (`virtumancer.db`) to store persistent configuration data.

## Philosophy

The database is designed to be the "source of truth" for user-defined configurations, such as the list of managed hosts and the intended configuration of virtual machines. Real-time state and performance metrics are always fetched directly from the libvirt hosts and are not stored in this database. This hybrid approach ensures that our configuration is persistent and manageable while the operational data is always live and accurate.

## Tables

### `hosts`

This table stores the connection information for each libvirt host that Virtumancer manages.

| Column | Type | Constraints | Description | 
| --- | --- | --- | --- |
| `id` | `TEXT` | `PRIMARY KEY` | The user-defined unique identifier for the host (e.g., "proxmox-1"). |
| `uri` | `TEXT` | | The full libvirt connection URI (e.g., "qemu+ssh://root@192.168.1.10/system"). |

### `virtual_machines`

This table stores a local cache of the configuration for each virtual machine. This record serves as the basis for future features like VM editing and configuration drift detection. The records in this table are automatically synchronized with the state reported by libvirt upon host connection and refresh events.

| Column | Type | Constraints | Description | 
| --- | --- | --- | --- |
| `id` | `INTEGER` | `PRIMARY KEY` | Auto-incrementing primary key provided by GORM. |
| `created_at` | `DATETIME` | | Timestamp of when the record was created (managed by GORM). |
| `updated_at` | `DATETIME` | | Timestamp of the last update to the record (managed by GORM). |
| `deleted_at` | `DATETIME` | `INDEX` | Timestamp for soft deletes (managed by GORM). |
| `name` | `TEXT` | `UNIQUE INDEX` | The name of the virtual machine. This is part of a composite unique index with `host_id`. |
| `host_id` | `TEXT` | `UNIQUE INDEX` | The `id` of the host this VM belongs to (foreign key to `hosts.id`). Part of a composite unique index with `name`. |
| `config_json` | `TEXT` | | A JSON string containing a snapshot of the VM's configuration (`VMInfo` struct) as retrieved from libvirt. |
