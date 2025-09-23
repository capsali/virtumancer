// Core data types used across the application
export interface Host {
  id: string;
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
  state: VMState;
  vcpuCount: number;
  memoryMB: number;
  // Additional fields from libvirt discovery
  osType?: string;
  isActive: boolean;
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
  memory_mb: number;
  disk_read_mb: number;
  disk_write_mb: number;
  network_rx_mb: number;
  network_tx_mb: number;
  uptime: number;
}

export interface VMHardware {
  vcpus: number;
  memory_mb: number;
  disks: VMDisk[];
  networks: VMNetwork[];
}

export interface VMDisk {
  device: string;
  type: string;
  size_gb: number;
  format: string;
  target: string;
}

export interface VMNetwork {
  interface: string;
  type: string;
  source: string;
  mac: string;
  model: string;
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