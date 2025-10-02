import type { 
  Host, 
  VirtualMachine, 
  DiscoveredVM,
  DiscoveredVMWithHost, 
  HostStats, 
  VMStats, 
  VMHardware,
  ApiResponse,
  PaginatedResponse,
  CreateVMData,
  VMTaskState,
  SyncStatus
} from '@/types';

// Import error recovery service for automatic error handling
import { errorRecoveryService } from './errorRecovery';
import { useHostStore } from '@/stores/hostStore';
import { useUIStore } from '@/stores/uiStore';

// Base API configuration
const API_BASE_URL = import.meta.env.DEV 
  ? '/api/v1'  // Use proxy in development
  : 'https://localhost:8888/api/v1';  // Direct connection in production

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public details?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// Generic API client with error handling and TypeScript support
class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, options);
  }

  // Request method with automatic error recovery integration
  private async requestWithRecovery<T>(
    endpoint: string,
    options: RequestInit = {},
    operation?: string
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        let errorCode = response.statusText;
        let errorDetails;
        
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          errorCode = errorData.code || errorCode;
          errorDetails = errorData;
        } catch {
          // If we can't parse the error response, use the default message
        }
        
        const apiError = new ApiError(
          errorMessage,
          response.status,
          errorCode,
          errorDetails
        );

        // Special handling for HOST_DISCONNECTED errors on manually disconnected hosts
        if (errorCode === 'HOST_DISCONNECTED') {
          // Try to extract host ID from operation
          const operationStr = operation || `${options.method || 'GET'} ${endpoint}`;
          const hostIdMatch = operationStr.match(/hosts\/([^\/]+)/);
          const hostId = hostIdMatch ? hostIdMatch[1] : null;
          
          if (hostId) {
            const hostStore = useHostStore();
            const host = hostStore.hosts.find(h => h.id === hostId);
            
            if (host?.auto_reconnect_disabled) {
              // Host was manually disconnected, show a toast instead of error notification
              const uiStore = useUIStore();
              uiStore.addToast('Host disconnected successfully', 'info', 10000);
              throw apiError;
            }
          }
        }

        // Add error to recovery service for automatic handling
        errorRecoveryService.addError(
          apiError,
          operation || `${options.method || 'GET'} ${endpoint}`,
          { 
            url, 
            status: response.status,
            endpoint,
            method: options.method || 'GET'
          }
        );
        
        throw apiError;
      }

      // Check if response has content before parsing JSON
      const contentType = response.headers.get('content-type');
      
      // Handle empty responses (204 No Content, empty body, etc.)
      if (response.status === 204 || !contentType?.includes('application/json')) {
        return undefined as T;
      }
      
      // Check if response body is empty
      const text = await response.text();
      if (!text.trim()) {
        return undefined as T;
      }
      
      const data = JSON.parse(text);
      return data;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      // Network or other errors
      const networkError = new ApiError(
        error instanceof Error ? error.message : 'Unknown error occurred',
        0,
        'NETWORK_ERROR',
        error
      );

      // Add network error to recovery service
      errorRecoveryService.addError(
        networkError,
        operation || `${options.method || 'GET'} ${endpoint}`,
        { 
          url, 
          endpoint,
          method: options.method || 'GET',
          networkError: true
        }
      );
      
      throw networkError;
    }
  }

  async get<T>(endpoint: string, operation?: string): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, { method: 'GET' }, operation);
  }

  async post<T>(endpoint: string, data?: any, operation?: string): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    }, operation);
  }

  async put<T>(endpoint: string, data?: any, operation?: string): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    }, operation);
  }

  async delete<T>(endpoint: string, data?: any, operation?: string): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, { 
      method: 'DELETE',
      body: data ? JSON.stringify(data) : undefined,
    }, operation);
  }

  async patch<T>(endpoint: string, data?: any, operation?: string): Promise<T> {
    return this.requestWithRecovery<T>(endpoint, {
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    }, operation);
  }
}

// Create shared API client instance
const apiClient = new ApiClient();

// Host API methods
export const hostApi = {
  async getAll(): Promise<Host[]> {
    return apiClient.get<Host[]>('/hosts', 'fetch_all_hosts');
  },

  async getById(id: string): Promise<Host> {
    return apiClient.get<Host>(`/hosts/${id}/info`, `fetch_host_${id}`);
  },

  async create(hostData: Omit<Host, 'id'>): Promise<Host> {
    return apiClient.post<Host>('/hosts', hostData, `create_host_${hostData.uri}`);
  },

  async update(id: string, updates: Partial<Host>): Promise<Host> {
    return apiClient.patch<Host>(`/hosts/${id}`, updates, `update_host_${id}`);
  },

  async delete(id: string): Promise<void> {
    return apiClient.delete(`/hosts/${id}`, undefined, `delete_host_${id}`);
  },

  async connect(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/connect`, undefined, `connect_host_${id}`);
  },

  async disconnect(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/disconnect`, undefined, `disconnect_host_${id}`);
  },

  async getStats(id: string): Promise<HostStats> {
    return apiClient.get<HostStats>(`/hosts/${id}/stats`);
  },

  async getVMs(id: string): Promise<VirtualMachine[]> {
    return apiClient.get<VirtualMachine[]>(`/hosts/${id}/vms`);
  },

  async getDiscoveredVMs(id: string): Promise<DiscoveredVM[]> {
    return apiClient.get<DiscoveredVM[]>(`/hosts/${id}/discovered-vms`);
  },

  async getAllDiscoveredVMs(): Promise<DiscoveredVMWithHost[]> {
    return apiClient.get<DiscoveredVMWithHost[]>('/discovered-vms');
  },

  async refreshAllDiscoveredVMs(): Promise<void> {
    return apiClient.post('/discovered-vms/refresh');
  },

  async importAllVMs(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/vms/import-all`);
  },

  async importSelectedVMs(id: string, domainUUIDs: string[]): Promise<void> {
    return apiClient.post(`/hosts/${id}/vms/import-selected`, { domain_uuids: domainUUIDs });
  },

  async deleteSelectedDiscoveredVMs(id: string, domainUUIDs: string[]): Promise<void> {
    return apiClient.delete(`/hosts/${id}/discovered-vms`, { domain_uuids: domainUUIDs });
  },

  async getPorts(id: string): Promise<any[]> {
    return apiClient.get<any[]>(`/hosts/${id}/ports`);
  },

  async getHostCapabilities(id: string): Promise<any> {
    return apiClient.get<any>(`/hosts/${id}/capabilities`);
  },

  async refreshHostCapabilities(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/capabilities/refresh`);
  }
};

// Backend VM response interface (snake_case from Go)
interface BackendVMResponse {
  name: string;
  uuid: string;
  domain_uuid: string;
  description: string;
  vcpu_count: number;
  memory_bytes: number;
  is_template: boolean;
  cpu_model: string;
  cpu_topology_json: string;
  os_type: string;
  task_state: string;
  sync_status: string;
  drift_details: string;
  needs_rebuild: boolean;
  state: string;
  libvirtState: string;
  graphics: {
    vnc: boolean;
    spice: boolean;
  };
  max_mem: number;
  memory: number;
  cpu_time: number;
  uptime: number;
  disk_size_gb?: number;
  network_interface?: string;
}

// Transform backend VM response to frontend VirtualMachine interface
function transformBackendVMToFrontend(backendVM: BackendVMResponse, hostId: string): VirtualMachine {
  return {
    // Core identifiers
    uuid: backendVM.uuid,
    name: backendVM.name,

    // Backend (snake_case)
    domain_uuid: backendVM.domain_uuid,
    vcpu_count: backendVM.vcpu_count,
    memory_bytes: backendVM.memory_bytes,
    os_type: backendVM.os_type,
    disk_size_gb: backendVM.disk_size_gb,
    network_interface: backendVM.network_interface,

    // Frontend-friendly aliases (camelCase)
    hostId: hostId,
    domainUuid: backendVM.domain_uuid,
    source: 'managed' as const, // VMs from this endpoint are managed (imported)
    title: backendVM.name, // Use name as title for now
    description: backendVM.description,
    state: backendVM.state as any,
    libvirtState: backendVM.libvirtState as any,
  taskState: backendVM.task_state as VMTaskState,
    vcpuCount: backendVM.vcpu_count,
    cpuModel: backendVM.cpu_model,
    memoryMB: Math.round(backendVM.memory_bytes / (1024 * 1024)), // Convert bytes to MB
    osType: backendVM.os_type,
    bootDevice: '', // Not provided by backend, set default
    diskSizeGB: backendVM.disk_size_gb || 0,
    networkInterface: backendVM.network_interface || '',
    syncStatus: backendVM.sync_status as any,
  // snake_case sync_status required by VirtualMachine
  sync_status: backendVM.sync_status as SyncStatus,
    graphics: backendVM.graphics,
    // Backend-required fields with defaults if missing
    is_template: backendVM.is_template ?? false,
    cpu_model: backendVM.cpu_model ?? 'host',
    cpu_topology_json: backendVM.cpu_topology_json ?? '{}',
  task_state: (backendVM.task_state as unknown as VMTaskState) ?? null,
    drift_details: backendVM.drift_details ?? '',
    needs_rebuild: backendVM.needs_rebuild ?? false,
    createdAt: '', // Not provided by backend, set default
    updatedAt: '', // Not provided by backend, set default
  };
}

// VM API methods
export const vmApi = {
  async getAll(hostId: string): Promise<VirtualMachine[]> {
    const backendVMs = await apiClient.get<BackendVMResponse[]>(`/hosts/${hostId}/vms`);
    return backendVMs.map(vm => transformBackendVMToFrontend(vm, hostId));
  },

  async getByHost(hostId: string): Promise<VirtualMachine[]> {
    const backendVMs = await apiClient.get<BackendVMResponse[]>(`/hosts/${hostId}/vms`);
    return backendVMs.map(vm => transformBackendVMToFrontend(vm, hostId));
  },

  async create(vmData: CreateVMData): Promise<VirtualMachine> {
    return apiClient.post<VirtualMachine>(`/hosts/${vmData.hostId}/vms`, vmData);
  },

  async update(uuid: string, updates: Partial<VirtualMachine>): Promise<VirtualMachine> {
    // Note: This might need adjustment based on actual backend API
    return apiClient.put<VirtualMachine>(`/vms/${uuid}`, updates);
  },

  async delete(uuid: string): Promise<void> {
    // Note: This might need adjustment based on actual backend API  
    return apiClient.delete(`/vms/${uuid}`);
  },

  async start(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/start`, undefined, `start_vm_${vmName}`);
  },

  async shutdown(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/shutdown`, undefined, `shutdown_vm_${vmName}`);
  },

  async reboot(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/reboot`, undefined, `reboot_vm_${vmName}`);
  },

  async forceOff(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/forceoff`, undefined, `force_off_vm_${vmName}`);
  },

  async forceReset(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/forcereset`, undefined, `force_reset_vm_${vmName}`);
  },

  async getStats(hostId: string, vmName: string): Promise<VMStats> {
    return apiClient.get<VMStats>(`/hosts/${hostId}/vms/${vmName}/stats`);
  },

  async getHardware(hostId: string, vmName: string): Promise<VMHardware> {
    return apiClient.get<VMHardware>(`/hosts/${hostId}/vms/${vmName}/hardware`);
  },

  async updateState(hostId: string, vmName: string, state: string): Promise<void> {
    return apiClient.put(`/hosts/${hostId}/vms/${vmName}/state`, { state });
  },

  async import(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/import`);
  },

  async sync(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/sync-from-libvirt`);
  },

  async rebuild(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/rebuild-from-db`);
  },

  async getPortAttachments(hostId: string, vmName: string): Promise<any[]> {
    return apiClient.get<any[]>(`/hosts/${hostId}/vms/${vmName}/port-attachments`);
  },

  async getVideoAttachments(hostId: string, vmName: string): Promise<any[]> {
    return apiClient.get<any[]>(`/hosts/${hostId}/vms/${vmName}/video-attachments`);
  },

  async updateHardware(hostId: string, vmName: string, hardwareConfig: any): Promise<{ success: boolean }> {
    return apiClient.put(`/hosts/${hostId}/vms/${vmName}/hardware`, hardwareConfig);
  },

  async getExtendedVMHardware(hostId: string, vmName: string): Promise<any> {
    return apiClient.get(`/hosts/${hostId}/vms/${vmName}/hardware/extended`);
  }
};

// WebSocket connection for real-time updates
export class WebSocketManager {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners = new Map<string, Set<Function>>();

  constructor(private url: string = import.meta.env.DEV 
    ? `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`  // Use proxy in development
    : 'wss://localhost:8888/ws'  // Direct connection in production
  ) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        resolve();
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.emit(data.type, data.payload);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      };
    });
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      this.connect().catch(() => {
        // Will retry again due to onclose handler
      });
    }, delay);
  }

  on(event: string, callback: Function): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
  }

  off(event: string, callback: Function): void {
    this.listeners.get(event)?.delete(callback);
  }

  private emit(event: string, data: any): void {
    this.listeners.get(event)?.forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error('Error in WebSocket event listener:', error);
      }
    });
  }

  send(type: string, payload: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message = { type, payload };
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', { type, payload });
    }
  }

  // VM Stats subscription methods
  subscribeToVMStats(hostId: string, vmName: string): void {
    this.send('subscribe-vm-stats', { hostId, vmName });
  }

  unsubscribeFromVMStats(hostId: string, vmName: string): void {
    this.send('unsubscribe-vm-stats', { hostId, vmName });
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
};

// Dashboard API methods
export const dashboardApi = {
  async getStats(): Promise<{
    infrastructure: {
      totalHosts: number;
      connectedHosts: number;
      totalVMs: number;
      runningVMs: number;
      stoppedVMs: number;
    };
    resources: {
      totalMemoryGB: number;
      usedMemoryGB: number;
      memoryUtilization: number;
      totalCPUs: number;
      allocatedCPUs: number;
      cpuUtilization: number;
    };
    health: {
      systemStatus: string;
      lastSync: string;
      errors: number;
      warnings: number;
    };
  }> {
    return apiClient.get('/dashboard/stats');
  },

  async getActivity(): Promise<{
    activities: Array<{
      id: string;
      type: 'vm_state_change' | 'host_connect' | 'host_disconnect' | 'system';
      message: string;
      hostId: string;
      vmUuid?: string;
      vmName?: string;
      timestamp: string;
      severity: 'info' | 'warning' | 'error';
      details?: string;
    }>;
    pagination: {
      total: number;
      page: number;
      limit: number;
    };
  }> {
    return apiClient.get('/dashboard/activity');
  },

  async getOverview(): Promise<{
    stats: any;
    activities: any[];
    timestamp: string;
  }> {
    return apiClient.get('/dashboard/overview');
  }
};

// Settings API
export const settingsApi = {
  async getMetrics(): Promise<any> {
    return apiClient.get<any>(`/settings/metrics`, `get_metrics_settings`);
  },

  async updateMetrics(payload: any): Promise<void> {
    return apiClient.put(`/settings/metrics`, payload, `update_metrics_settings`);
  }
}

// Export singleton WebSocket manager
export const wsManager = new WebSocketManager();