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

  // Optional camelCase aliases used across the frontend for convenience
  hostId?: string;
  domainUuid?: string;
  vcpuCount?: number;
  memoryBytes?: number; // added alias for backend memory_bytes
  currentMemory?: number;
  memoryMB?: number;
  cpuModel?: string;
  osType?: string;
  taskState?: VMTaskState;
  syncStatus?: SyncStatus;
  source?: string;
  title?: string;
  bootDevice?: string;
  diskSizeGB?: number;
  networkInterface?: string;
}

export interface DiscoveredVM {
  domain_uuid: string;
  name: string;
  host_id: string;
  last_seen_at: string;
  imported: boolean;
  // Optional properties for UI compatibility when displaying in VM lists
  uuid?: string;
  state?: VMState;
  vcpuCount?: number;
  memoryMB?: number;
  osType?: string;
  isActive?: boolean;
}

export interface DiscoveredVMWithHost {
  uuid: string;
  name: string;
  host_id: string;
  host_name: string;
  // Optional VM info properties
  state?: VMState;
  max_mem?: number;
  memory?: number;
  vcpu?: number;
  cpu_time?: number;
  uptime?: number;
  persistent?: boolean;
  autostart?: boolean;
  graphics?: GraphicsInfo;
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
  vcpus: number;
  cpu_model: string;
  cpu_topology: {
    sockets: number;
    cores: number;
    threads: number;
  };
  cpu_features: any[];
  memory_bytes: number;
  current_memory: number;
  memory_mb?: number; // Fallback property
  memory_backing: any;
  disks: HardwareDiskInfo[];
  networks: HardwareNetworkInfo[];
  video_devices: any[];
  controllers: any[];
  host_devices: any[];
  tpm: {
    enabled: boolean;
    version: string;
    backend: string;
  };
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

export interface HardwareDiskInfo {
  id?: string;
  deviceName: string;
  device?: string; // Fallback property
  target?: string; // Fallback property
  busType: string;
  type?: string; // Fallback property
  capacityGB: number;
  size_gb?: number; // Fallback property
  format: string;
  readOnly: boolean;
  shareable: boolean;
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

export interface HardwareNetworkInfo {
  id?: string;
  macAddress: string;
  mac?: string; // Fallback property
  modelName: string;
  model?: string; // Fallback property
  sourceType: string;
  type?: string; // Fallback property
  sourceRef: string;
  source?: string; // Fallback property
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
  cpu_percent: number;
  memory_total: number;
  memory_available: number;
  disk_total: number;
  disk_free: number;
  uptime: number;
  vm_count: number;
  host_info: any;
  vm_counts: Record<string, number>;
  total_vms: number;
  resources: any;
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
  hosts: boolean;
  vms: boolean;
  addHost: boolean;
  vmAction: boolean | null;
  vmHardware: boolean;
  vmReconcile: boolean | null;
  vmImport: boolean | null;
  hostImportAll: string | null;
  connectHost: Record<string, boolean>;
  hostStats: Record<string, boolean>;
  hostCapabilities: Record<string, boolean>;
  [key: string]: boolean | null | string | Record<string, boolean>;
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

// Data shape used when creating a VM from the UI. Uses snake_case to match backend.
export interface CreateVMData extends VMFormData {
  hostId: string;
  disk_size_gb?: number;
  network_interface?: string;
  boot_device?: string;
  cpu_model?: string;
  source?: string;
  sync_status?: string;
  libvirtState?: string;
  domain_uuid?: string;
  title?: string;
  state?: string;
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

export interface UIPreferences {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'auto';
  colorScheme: string;
  reducedMotion: boolean;
  particleEffects: boolean;
  glowEffects: boolean;
}

export interface ViewState {
  currentView: string;
  breadcrumbs: string[];
  filters: Record<string, any>;
  sortBy: string;
  sortOrder: 'asc' | 'desc';
}
