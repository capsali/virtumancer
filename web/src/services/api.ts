import type { 
  Host, 
  VirtualMachine, 
  DiscoveredVM, 
  HostStats, 
  VMStats, 
  VMHardware,
  ApiResponse,
  PaginatedResponse 
} from '@/types';

// Base API configuration
const API_BASE_URL = import.meta.env.DEV 
  ? '/api/v1'  // Use proxy in development
  : 'https://localhost:8888/api/v1';  // Direct connection in production

class ApiError extends Error {
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
        let errorDetails;
        
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          errorDetails = errorData;
        } catch {
          // If we can't parse the error response, use the default message
        }
        
        throw new ApiError(
          errorMessage,
          response.status,
          response.statusText,
          errorDetails
        );
      }

      const data = await response.json();
      return data;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      // Network or other errors
      throw new ApiError(
        error instanceof Error ? error.message : 'Unknown error occurred',
        0,
        'NETWORK_ERROR',
        error
      );
    }
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }

  async patch<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    });
  }
}

// Create shared API client instance
const apiClient = new ApiClient();

// Host API methods
export const hostApi = {
  async getAll(): Promise<Host[]> {
    return apiClient.get<Host[]>('/hosts');
  },

  async getById(id: string): Promise<Host> {
    return apiClient.get<Host>(`/hosts/${id}/info`);
  },

  async create(hostData: Omit<Host, 'id'>): Promise<Host> {
    return apiClient.post<Host>('/hosts', hostData);
  },

  async update(id: string, updates: Partial<Host>): Promise<Host> {
    return apiClient.put<Host>(`/hosts/${id}`, updates);
  },

  async delete(id: string): Promise<void> {
    return apiClient.delete(`/hosts/${id}`);
  },

  async connect(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/connect`);
  },

  async disconnect(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/disconnect`);
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

  async importAllVMs(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/vms/import-all`);
  },

  async getPorts(id: string): Promise<any[]> {
    return apiClient.get<any[]>(`/hosts/${id}/ports`);
  }
};

// VM API methods
export const vmApi = {
  async getAll(hostId: string): Promise<VirtualMachine[]> {
    return apiClient.get<VirtualMachine[]>(`/hosts/${hostId}/vms`);
  },

  async getByHost(hostId: string): Promise<VirtualMachine[]> {
    return apiClient.get<VirtualMachine[]>(`/hosts/${hostId}/vms`);
  },

  async create(vmData: Omit<VirtualMachine, 'uuid' | 'createdAt' | 'updatedAt'>): Promise<VirtualMachine> {
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
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/start`);
  },

  async shutdown(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/shutdown`);
  },

  async reboot(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/reboot`);
  },

  async forceOff(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/forceoff`);
  },

  async forceReset(hostId: string, vmName: string): Promise<void> {
    return apiClient.post(`/hosts/${hostId}/vms/${vmName}/forcereset`);
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

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}

// Export singleton WebSocket manager
export const wsManager = new WebSocketManager();

// Export ApiError for error handling in stores
export { ApiError };