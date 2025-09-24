// Core data types used across the application
export interface Host {
  id: string;
  name?: string;
  uri: string;
  state: HostState;
  task_state?: HostTaskState;
  auto_reconnect_disabled: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface VirtualMachine {
  uuid: string;
  hostId: string;
  name: string;
  domainUuid: string;
  source: 'managed' | 'imported';
  title: string;
  description: string;
  state: VMState;
  libvirtState: VMState;
  taskState?: VMTaskState;
  vcpuCount: number;
  cpuModel: string;
  memoryMB: number;
  osType: string;
  bootDevice: string;
  diskSizeGB: number;
  networkInterface: string;
  syncStatus: SyncStatus;
  createdAt: string;
  updatedAt: string;
}

export interface DiscoveredVM {
  uuid: string;
  name: string;
  domain_uuid: string;
  host_id: string;
  info_json?: string;
  last_seen_at: string;
  imported: boolean;
  // Additional computed fields
  state?: VMState;
  vcpuCount?: number;
  memoryMB?: number;
  osType?: string;
  isActive?: boolean;
}

export interface HostStats {
  cpu_percent: number;
  memory_total: number;
  memory_available: number;
  disk_total: number;
  disk_free: number;
  uptime: number;
  vm_count: number;
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
  disk_read_kib_per_sec?: number;
  disk_write_kib_per_sec?: number;
  disk_read_iops?: number;
  disk_write_iops?: number;
  network_rx_mbps?: number;
  network_tx_mbps?: number;
  network_rx_mb: number;
  network_tx_mb: number;
  uptime: number;
}

export interface VMHardware {
  // CPU Configuration
  vcpus: number;
  cpu_model?: string;
  cpu_topology?: {
    sockets: number;
    cores: number;
    threads: number;
  };
  cpu_features?: Array<{
    name: string;
    policy: string;
  }>;
  
  // Memory Configuration
  memory_mb: number;
  memory_bytes?: number;
  current_memory?: number;
  memory_backing?: {
    mode: string;
    sourceType: string;
    locked: boolean;
    nosharepages: boolean;
  };
  
  // Storage
  disks: VMDisk[];
  
  // Network
  networks: VMNetwork[];
  
  // Video/Graphics
  video_devices?: Array<{
    id?: string;
    model: string;
    vram: number;
    heads: number;
    accel3d: boolean;
  }>;
  
  // Advanced Hardware
  controllers?: Array<{
    id: string;
    type: string;
    model: string;
  }>;
  host_devices?: Array<{
    id: string;
    name: string;
    address: string;
  }>;
  tpm?: {
    enabled: boolean;
    version: string;
    backend: string;
  };
}

export interface VMDisk {
  id?: string;
  device: string;
  deviceName?: string;
  type: string;
  busType?: string;
  size_gb: number;
  capacityGB?: number;
  format: string;
  target: string;
  readOnly?: boolean;
  shareable?: boolean;
}

export interface VMNetwork {
  id?: string;
  interface: string;
  type: string;
  source: string;
  sourceRef?: string;
  sourceType?: string;
  mac: string;
  macAddress?: string;
  model: string;
  modelName?: string;
}

export interface ToastMessage {
  id: number;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  timeout: number;
}

// Enums
export type HostState = 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
export type HostTaskState = 'CONNECTING' | 'DISCONNECTING';

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
  | 'SCHEDULING';

export type SyncStatus = 'UNKNOWN' | 'SYNCED' | 'DRIFTED';

// Loading states
export interface LoadingStates {
  hosts: boolean;
  vms: boolean;
  addHost: boolean;
  vmAction: string | null;
  vmHardware: boolean;
  vmReconcile: string | null;
  vmImport: string | null;
  hostImportAll: string | null;
  connectHost: Record<string, boolean>;
}

// API Response types
export interface ApiResponse<T> {
  data: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

// Error handling
export interface AppError {
  message: string;
  code?: string;
  details?: any;
  timestamp: Date;
}

// UI State types
export interface UIPreferences {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'auto';
  colorScheme: 'blue' | 'purple' | 'cyan' | 'neon';
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