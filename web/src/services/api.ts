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
const API_BASE_URL = '/api/v1';

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
    return apiClient.get<Host>(`/hosts/${id}`);
  },

  async create(hostData: Omit<Host, 'id'>): Promise<Host> {
    return apiClient.post<Host>('/hosts', hostData);
  },

  async update(id: string, hostData: Partial<Host>): Promise<Host> {
    return apiClient.put<Host>(`/hosts/${id}`, hostData);
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

  async getDiscoveredVMs(id: string): Promise<DiscoveredVM[]> {
    return apiClient.get<DiscoveredVM[]>(`/hosts/${id}/discovered-vms`);
  },

  async importAllVMs(id: string): Promise<void> {
    return apiClient.post(`/hosts/${id}/import-all`);
  }
};

// VM API methods
export const vmApi = {
  async getAll(hostId?: string): Promise<VirtualMachine[]> {
    const endpoint = hostId ? `/vms?hostId=${hostId}` : '/vms';
    return apiClient.get<VirtualMachine[]>(endpoint);
  },

  async getById(uuid: string): Promise<VirtualMachine> {
    return apiClient.get<VirtualMachine>(`/vms/${uuid}`);
  },

  async create(vmData: Omit<VirtualMachine, 'uuid' | 'createdAt' | 'updatedAt'>): Promise<VirtualMachine> {
    return apiClient.post<VirtualMachine>('/vms', vmData);
  },

  async update(uuid: string, vmData: Partial<VirtualMachine>): Promise<VirtualMachine> {
    return apiClient.put<VirtualMachine>(`/vms/${uuid}`, vmData);
  },

  async delete(uuid: string): Promise<void> {
    return apiClient.delete(`/vms/${uuid}`);
  },

  async start(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/start`);
  },

  async stop(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/stop`);
  },

  async restart(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/restart`);
  },

  async pause(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/pause`);
  },

  async unpause(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/unpause`);
  },

  async suspend(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/suspend`);
  },

  async resume(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/resume`);
  },

  async getStats(uuid: string): Promise<VMStats> {
    return apiClient.get<VMStats>(`/vms/${uuid}/stats`);
  },

  async getHardware(uuid: string): Promise<VMHardware> {
    return apiClient.get<VMHardware>(`/vms/${uuid}/hardware`);
  },

  async updateHardware(uuid: string, hardware: Partial<VMHardware>): Promise<VMHardware> {
    return apiClient.put<VMHardware>(`/vms/${uuid}/hardware`, hardware);
  },

  async import(hostId: string, vmUuid: string): Promise<VirtualMachine> {
    return apiClient.post<VirtualMachine>(`/vms/import`, { hostId, vmUuid });
  },

  async reconcile(uuid: string): Promise<void> {
    return apiClient.post(`/vms/${uuid}/reconcile`);
  }
};

// WebSocket connection for real-time updates
export class WebSocketManager {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners = new Map<string, Set<Function>>();

  constructor(private url: string = 'ws://localhost:8080/ws') {}

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