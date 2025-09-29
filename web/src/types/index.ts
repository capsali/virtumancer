// Core application types for Virtumancer

export interface Host {
  id: string;
  name?: string;
  uri: string;
  state: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  task_state?: string;
  auto_reconnect_disabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface VirtualMachine {
  uuid: string;
  name: string;
  domain_uuid: string;
  description: string;
  vcpu_count: number;
  memory_bytes: number;
  is_template: boolean;
  cpu_model: string;
  cpu_topology_json: string;
  os_type: string;
  state: VMState;
  libvirtState: VMState;
  task_state: VMTaskState;
  graphics: GraphicsInfo;
  sync_status: SyncStatus;
  drift_details: string;
  needs_rebuild: boolean;
  createdAt: string;
  updatedAt: string;
  uptime?: number;
  disk_size_gb?: number;
  network_interface?: string;
}

export interface DiscoveredVM {
  domain_uuid: string;
  name: string;
  host_id: string;
  last_seen_at: string;
  imported: boolean;
}

export interface VMStats {
  cpu_percent: number;
  cpu_percent_core?: number;
  cpu_percent_raw?: number;
  cpu_percent_guest?: number;
  cpu_percent_host?: number;
  memory_mb: number;
  disk_read_mb: number;
  disk_write_mb: number;
  disk_read_kib_per_sec: number;
  disk_write_kib_per_sec: number;
  disk_read_iops: number;
  disk_write_iops: number;
  network_rx_mb: number;
  network_tx_mb: number;
  network_rx_mbps: number;
  network_tx_mbps: number;
  uptime: number;
}

export interface VMHardware {
  name: string;
  uuid: string;
  title: string;
  description: string;
  memory: {
    value: number;
    unit: string;
  };
  currentMemory: {
    value: number;
    unit: string;
  };
  vcpu: number;
  cpu: {
    mode: string;
    model: {
      name: string;
      fallback: string;
    };
    topology: {
      sockets: number;
      cores: number;
      threads: number;
    };
  };
  os: {
    type: string;
    loader?: {
      path: string;
      type: string;
      readonly: string;
      secure: string;
    };
  };
  disks: DiskInfo[];
  networks: NetworkInfo[];
  graphics: GraphicsInfo[];
  consoles: ConsoleInfo[];
}

export interface DiskInfo {
  device: string;
  driver: {
    driver_name: string;
    type: string;
  };
  source: {
    file: string;
    dev: string;
  };
  path: string;
  name: string;
  readonly: boolean;
  shareable: boolean;
  target: {
    dev: string;
    bus: string;
  };
}

export interface NetworkInfo {
  type: string;
  mac: {
    address: string;
  };
  source: {
    bridge: string;
    network: string;
    portgroup: string;
  };
  model: {
    type: string;
  };
  target: {
    dev: string;
  };
}

export interface GraphicsInfo {
  vnc: boolean;
  spice: boolean;
}

export interface ConsoleInfo {
  type: string;
  port: number;
  host: string;
  password?: string;
}

export interface HostStats {
  totalMemoryGB: number;
  usedMemoryGB: number;
  memoryUtilization: number;
  totalCPUs: number;
  allocatedCPUs: number;
}

export type VMState = 
  | 'INITIALIZED'
  | 'ACTIVE' 
  | 'PAUSED'
  | 'SUSPENDED'
  | 'STOPPED'
  | 'ERROR'
  | 'UNKNOWN';

export type VMTaskState = 
  | 'BUILDING'
  | 'PAUSING'
  | 'UNPAUSING'
  | 'SUSPENDING'
  | 'RESUMING'
  | 'DELETING'
  | 'STOPPING'
  | 'STARTING'
  | 'REBOOTING'
  | 'REBUILDING'
  | 'POWERING_ON'
  | 'POWERING_OFF'
  | 'SCHEDULING'
  | null;

export type SyncStatus = 'UNKNOWN' | 'SYNCED' | 'DRIFTED';

export interface AppError {
  message: string;
  code?: string | number;
  details?: any;
  timestamp: Date;
}

export interface LoadingStates {
  [key: string]: boolean;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T = any> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// WebSocket message types
export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp?: string;
}

export interface VMStatsUpdate extends WebSocketMessage {
  type: 'vm-stats-updated';
  data: {
    hostId: string;
    vmName: string;
    stats: VMStats;
  };
}

export interface HostStateUpdate extends WebSocketMessage {
  type: 'host-state-changed';
  data: {
    hostId: string;
    state: string;
    task_state?: string;
  };
}

export interface VMStateUpdate extends WebSocketMessage {
  type: 'vm-state-changed';
  data: {
    hostId: string;
    vmName: string;
    state: VMState;
    task_state?: VMTaskState;
  };
}

// Form types
export interface HostFormData {
  id: string;
  name?: string;
  uri: string;
}

export interface VMFormData {
  name: string;
  description?: string;
  vcpu_count: number;
  memory_bytes: number;
  os_type?: string;
}

// UI-specific types
export interface BreadcrumbItem {
  label: string;
  to?: string;
  active?: boolean;
}

export interface ToastMessage {
  id: string;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number;
  persistent?: boolean;
}

export interface ModalConfig {
  isOpen: boolean;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  persistent?: boolean;
}
