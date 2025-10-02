import { defineStore } from 'pinia';
import { ref, computed, watch, readonly } from 'vue';
import type { 
  Host, 
  HostStats, 
  LoadingStates, 
  DiscoveredVM,
  DiscoveredVMWithHost,
  AppError 
} from '@/types';
import { hostApi, wsManager, ApiError } from '@/services/api';
import { errorRecoveryService } from '@/services/errorRecovery';

export const useHostStore = defineStore('hosts', () => {
  // State
  const hosts = ref<Host[]>([]);
  const selectedHostId = ref<string | null>(null);
  const hostStats = ref<Record<string, HostStats>>({});
  const hostCapabilities = ref<Record<string, any>>({});
  const discoveredVMs = ref<Record<string, DiscoveredVM[]>>({});
  const globalDiscoveredVMs = ref<DiscoveredVMWithHost[]>([]);
  const errors = ref<Record<string, AppError>>({});
  
  // Background polling
  const discoveredVMsPollingInterval = ref<number | null>(null);
  const POLLING_INTERVAL = 30000; // 30 seconds
  
  // Loading states with granular control
  const loading = ref<LoadingStates>({
    hosts: false,
    vms: false,
    addHost: false,
    vmAction: null,
    vmHardware: false,
    vmReconcile: null,
    vmImport: null,
    hostImportAll: null,
    connectHost: {},
    hostStats: {},
    hostCapabilities: {},
    globalDiscoveredVMs: false,
    refreshDiscoveredVMs: false
  });

  // Connection state tracking with debouncing
  const hostConnecting = ref<Record<string, boolean>>({});
  const visibleConnecting = ref<Record<string, boolean>>({});
  const connectDebounceTimers = ref<Record<string, ReturnType<typeof setTimeout>>>({});

  // Computed properties
  const selectedHost = computed((): Host | null => {
    if (!selectedHostId.value) return null;
    return hosts.value.find(h => h.id === selectedHostId.value) ?? null;
  });

  const connectedHosts = computed((): Host[] => {
    return hosts.value.filter(h => h && h.state === 'CONNECTED');
  });

  const disconnectedHosts = computed((): Host[] => {
    return hosts.value.filter(h => h && h.state === 'DISCONNECTED');
  });

  const allDiscoveredVMs = computed(() => {
    return globalDiscoveredVMs.value;
  });

  const errorHosts = computed((): Host[] => {
    return hosts.value.filter(h => h && h.state === 'ERROR');
  });

  const hostsWithStats = computed(() => {
    return hosts.value.filter(h => h).map(host => ({
      ...host,
      stats: hostStats.value[host.id] || null,
      discovered: discoveredVMs.value[host.id] || [],
      isConnecting: visibleConnecting.value[host.id] || false
    }));
  });

  // Actions
  // Global discovered VMs management
  const fetchGlobalDiscoveredVMs = async (): Promise<void> => {
    loading.value.globalDiscoveredVMs = true;
    clearError('fetchGlobalDiscoveredVMs');
    
    try {
      const data = await hostApi.getAllDiscoveredVMs();
      globalDiscoveredVMs.value = data;
    } catch (error) {
      handleError('fetchGlobalDiscoveredVMs', error);
      // Don't throw here to allow continued operation
    } finally {
      loading.value.globalDiscoveredVMs = false;
    }
  };

  const refreshAllDiscoveredVMs = async (): Promise<void> => {
    loading.value.refreshDiscoveredVMs = true;
    clearError('refreshAllDiscoveredVMs');
    
    try {
      await hostApi.refreshAllDiscoveredVMs();
      // Wait a moment for the refresh to propagate, then fetch updated data
      setTimeout(() => {
        fetchGlobalDiscoveredVMs();
      }, 1000);
    } catch (error) {
      handleError('refreshAllDiscoveredVMs', error);
      throw error;
    } finally {
      loading.value.refreshDiscoveredVMs = false;
    }
  };

  // Background polling management
  const startDiscoveredVMsPolling = (): void => {
    if (discoveredVMsPollingInterval.value) {
      return; // Already polling
    }
    
    // Initial fetch
    fetchGlobalDiscoveredVMs();
    
    // Set up polling
    discoveredVMsPollingInterval.value = window.setInterval(() => {
      fetchGlobalDiscoveredVMs();
    }, POLLING_INTERVAL);
  };

  const stopDiscoveredVMsPolling = (): void => {
    if (discoveredVMsPollingInterval.value) {
      clearInterval(discoveredVMsPollingInterval.value);
      discoveredVMsPollingInterval.value = null;
    }
  };

  const fetchHosts = async (): Promise<void> => {
    loading.value.hosts = true;
    clearError('fetchHosts');
    
    try {
      const data = await hostApi.getAll();
      hosts.value = data;
      
      // Start discovered VMs polling when hosts are loaded
      startDiscoveredVMsPolling();
    } catch (error) {
      handleError('fetchHosts', error);
      throw error;
    } finally {
      loading.value.hosts = false;
    }
  };

  const addHost = async (hostData: Omit<Host, 'id'>): Promise<Host> => {
    loading.value.addHost = true;
    clearError('addHost');
    
    try {
      const newHost = await hostApi.create(hostData);
      hosts.value.push(newHost);
      
      // Automatically try to connect to the new host
      if (newHost.id) {
        connectHost(newHost.id);
      }
      
      return newHost;
    } catch (error) {
      handleError('addHost', error);
      throw error;
    } finally {
      loading.value.addHost = false;
    }
  };

  const updateHost = async (id: string, updates: Partial<Host>): Promise<void> => {
    clearError('updateHost');
    
    try {
      const updatedHost = await hostApi.update(id, updates);
      const index = hosts.value.findIndex(h => h.id === id);
      if (index !== -1) {
        hosts.value[index] = updatedHost;
      }
    } catch (error) {
      handleError('updateHost', error);
      throw error;
    }
  };

  const deleteHost = async (id: string): Promise<void> => {
    clearError('deleteHost');
    
    try {
      await hostApi.delete(id);
      hosts.value = hosts.value.filter(h => h.id !== id);
      
      // Clean up related data
      delete hostStats.value[id];
      delete discoveredVMs.value[id];
      delete hostConnecting.value[id];
      delete visibleConnecting.value[id];
      
      // Clear timers
      if (connectDebounceTimers.value[id]) {
        clearTimeout(connectDebounceTimers.value[id]);
        delete connectDebounceTimers.value[id];
      }
      
      // Clear selection if this host was selected
      if (selectedHostId.value === id) {
        selectedHostId.value = null;
      }
    } catch (error) {
      handleError('deleteHost', error);
      throw error;
    }
  };

  const connectHost = async (id: string): Promise<void> => {
    loading.value.connectHost[id] = true;
    hostConnecting.value[id] = true;
    clearError('connectHost');
    
    try {
      await hostApi.connect(id);
      
      // Update host state optimistically
      const host = hosts.value.find(h => h.id === id);
      if (host) {
        host.task_state = 'CONNECTING';
      }
    } catch (error) {
      handleError('connectHost', error);
      throw error;
    } finally {
      loading.value.connectHost[id] = false;
      hostConnecting.value[id] = false;
    }
  };

  const disconnectHost = async (id: string): Promise<void> => {
    loading.value.connectHost[id] = true;
    clearError('disconnectHost');
    
    try {
      await hostApi.disconnect(id);
      
      // Update host state optimistically
      const host = hosts.value.find(h => h.id === id);
      if (host) {
        host.task_state = 'DISCONNECTING';
      }
    } catch (error) {
      handleError('disconnectHost', error);
      throw error;
    } finally {
      loading.value.connectHost[id] = false;
    }
  };

  const fetchHostStats = async (id: string): Promise<void> => {
    clearError('fetchHostStats');
    
    try {
      loading.value.hostStats[id] = true;
      console.log('Fetching host stats for:', id);
      const stats = await hostApi.getStats(id);
      console.log('Received host stats:', stats);
      hostStats.value[id] = stats;
      console.log('Host stats stored:', hostStats.value[id]);
    } catch (error) {
      console.error('Error fetching host stats:', error);
      handleError('fetchHostStats', error);
      // Don't throw here, stats are optional
    } finally {
      loading.value.hostStats[id] = false;
    }
  };

  const fetchHostCapabilities = async (id: string): Promise<void> => {
    clearError('fetchHostCapabilities');

    try {
      loading.value.hostCapabilities[id] = true;
      console.log('Fetching host capabilities for:', id);
      const capabilities = await hostApi.getHostCapabilities(id);
      console.log('Received host capabilities:', capabilities);
      hostCapabilities.value[id] = capabilities;
      console.log('Host capabilities stored:', hostCapabilities.value[id]);
    } catch (error) {
      console.error('Error fetching host capabilities:', error);
      handleError('fetchHostCapabilities', error);
      // Don't throw here, capabilities are optional
    } finally {
      loading.value.hostCapabilities[id] = false;
    }
  };

  const refreshDiscoveredVMs = async (hostId: string): Promise<DiscoveredVM[]> => {
    clearError('refreshDiscoveredVMs');
    
    try {
      const vms = await hostApi.getDiscoveredVMs(hostId);
      
      // If fetch returned empty but we have cached data, preserve it
      const currentCache = discoveredVMs.value[hostId];
      const hadPrevious = currentCache && currentCache.length > 0;
      if ((!vms || vms.length === 0) && hadPrevious) {
        console.warn(`RefreshDiscoveredVMs: got empty list for ${hostId}, preserving cache`);
        return currentCache;
      }
      
      discoveredVMs.value[hostId] = vms;
      return vms;
    } catch (error) {
      handleError('refreshDiscoveredVMs', error);
      return discoveredVMs.value[hostId] || [];
    }
  };

  const importAllVMs = async (hostId: string): Promise<void> => {
    loading.value.hostImportAll = hostId;
    clearError('importAllVMs');
    
    try {
      await hostApi.importAllVMs(hostId);
      // Refresh discovered VMs after import
      await refreshDiscoveredVMs(hostId);
    } catch (error) {
      handleError('importAllVMs', error);
      throw error;
    } finally {
      loading.value.hostImportAll = null;
    }
  };

  const importSelectedVMs = async (hostId: string, domainUUIDs: string[]): Promise<void> => {
    clearError('importSelectedVMs');
    
    try {
      await hostApi.importSelectedVMs(hostId, domainUUIDs);
      // Refresh discovered VMs after import
      await refreshDiscoveredVMs(hostId);
    } catch (error) {
      handleError('importSelectedVMs', error);
      throw error;
    }
  };

  const deleteSelectedDiscoveredVMs = async (hostId: string, domainUUIDs: string[]): Promise<void> => {
    clearError('deleteSelectedDiscoveredVMs');
    
    try {
      await hostApi.deleteSelectedDiscoveredVMs(hostId, domainUUIDs);
      // Refresh discovered VMs after deletion
      await refreshDiscoveredVMs(hostId);
    } catch (error) {
      handleError('deleteSelectedDiscoveredVMs', error);
      throw error;
    }
  };

  // Helper functions
  const selectHost = (id: string | null): void => {
    selectedHostId.value = id;
  };

  const getHostById = (id: string): Host | undefined => {
    return hosts.value.find(h => h.id === id);
  };

  const clearError = (key: string): void => {
    delete errors.value[key];
  };

  const handleError = (operation: string, error: unknown): void => {
    const appError: AppError = {
      message: error instanceof ApiError ? error.message : 'An unexpected error occurred',
      code: error instanceof ApiError ? error.code : 'UNKNOWN_ERROR',
      details: error instanceof ApiError ? error.details : error,
      timestamp: new Date()
    };
    
    errors.value[operation] = appError;
    console.error(`Host store error in ${operation}:`, appError);

    // Also add to error recovery service for enhanced handling
    errorRecoveryService.addError(
      error as Error,
      operation,
      {
        store: 'hostStore',
        selectedHostId: selectedHostId.value,
        hostCount: hosts.value.length
      }
    );
  };

  // Connection state debouncing watcher
  watch(
    () => hosts.value.map(h => ({ id: h.id, task_state: h.task_state })),
    (newHosts) => {
      const seen = new Set<string>();
      
      newHosts.forEach(({ id, task_state }) => {
        seen.add(id);
        const isConnecting = task_state?.toUpperCase() === 'CONNECTING';
        
        if (isConnecting) {
          // Already visible, nothing to do
          if (visibleConnecting.value[id]) return;
          
          // Timer already scheduled, keep it
          if (connectDebounceTimers.value[id]) return;
          
          // Schedule debounce (300ms delay to avoid flicker)
          connectDebounceTimers.value[id] = setTimeout(() => {
            visibleConnecting.value[id] = true;
            delete connectDebounceTimers.value[id];
          }, 300);
        } else {
          // Not connecting, clear immediately
          if (connectDebounceTimers.value[id]) {
            clearTimeout(connectDebounceTimers.value[id]);
            delete connectDebounceTimers.value[id];
          }
          visibleConnecting.value[id] = false;
        }
      });
      
      // Clean up any hosts that no longer exist
      Object.keys(visibleConnecting.value).forEach(id => {
        if (!seen.has(id)) {
          visibleConnecting.value[id] = false;
          if (connectDebounceTimers.value[id]) {
            clearTimeout(connectDebounceTimers.value[id]);
            delete connectDebounceTimers.value[id];
          }
        }
      });
    },
    { deep: true }
  );

  // WebSocket integration for real-time updates
  const initializeWebSocket = (): void => {
    wsManager.on('host_updated', (hostData: Host) => {
      const index = hosts.value.findIndex(h => h.id === hostData.id);
      if (index !== -1) {
        hosts.value[index] = { ...hosts.value[index], ...hostData };
      }
    });

    wsManager.on('host_stats', (data: { hostId: string; stats: HostStats }) => {
      hostStats.value[data.hostId] = data.stats;
    });

    wsManager.on('discovered_vms_updated', (data: { hostId: string; vms: DiscoveredVM[] }) => {
      discoveredVMs.value[data.hostId] = data.vms;
    });
  };

  return {
    // State
    hosts: readonly(hosts),
    selectedHostId,
    hostStats: readonly(hostStats),
    hostCapabilities: readonly(hostCapabilities),
    discoveredVMs: readonly(discoveredVMs),
    globalDiscoveredVMs: readonly(globalDiscoveredVMs),
    allDiscoveredVMs,
    loading: readonly(loading),
    errors: readonly(errors),
    
    // Computed
    selectedHost,
    connectedHosts,
    disconnectedHosts,
    errorHosts,
    hostsWithStats,
    
    // Actions
    fetchHosts,
    fetchGlobalDiscoveredVMs,
    refreshAllDiscoveredVMs,
    startDiscoveredVMsPolling,
    stopDiscoveredVMsPolling,
    addHost,
    updateHost,
    deleteHost,
    connectHost,
    disconnectHost,
    fetchHostStats,
    fetchHostCapabilities,
    refreshDiscoveredVMs,
    importAllVMs,
    importSelectedVMs,
    deleteSelectedDiscoveredVMs,
    selectHost,
    getHostById,
    clearError,
    initializeWebSocket
  };
});