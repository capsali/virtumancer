import { defineStore } from 'pinia';
import { ref, computed, readonly } from 'vue';
import { wsManager } from '@/services/api';
import { useHostStore } from './hostStore';
import { useVMStore } from './vmStore';
import { useUIStore } from './uiStore';
import type { AppError } from '@/types';

interface AppStats {
  totalHosts: number;
  connectedHosts: number;
  totalVMs: number;
  runningVMs: number;
  lastUpdated: Date;
}

export const useAppStore = defineStore('app', () => {
  // State
  const isInitialized = ref(false);
  const isOnline = ref(navigator.onLine);
  const lastSyncTime = ref<Date | null>(null);
  const globalErrors = ref<AppError[]>([]);
  const appVersion = ref('1.0.0');
  
  // Loading states for app-wide operations
  const isInitializing = ref(false);
  const isSyncing = ref(false);

  // Get store instances
  const hostStore = useHostStore();
  const vmStore = useVMStore();
  const uiStore = useUIStore();

  // Computed properties
  const appStats = computed((): AppStats => {
    return {
      totalHosts: hostStore.hosts.length,
      connectedHosts: hostStore.connectedHosts.length,
      totalVMs: vmStore.vms.length,
      runningVMs: vmStore.activeVMs.length,
      lastUpdated: lastSyncTime.value || new Date()
    };
  });

  const connectionStatus = computed(() => {
    if (!isOnline.value) return 'offline';
    if (!isInitialized.value) return 'initializing';
    if (hostStore.connectedHosts.length === 0) return 'disconnected';
    return 'connected';
  });

  const healthStatus = computed(() => {
    const errorCount = globalErrors.value.length;
    if (errorCount === 0) return 'healthy';
    if (errorCount < 3) return 'warning';
    return 'critical';
  });

  const isReady = computed((): boolean => {
    return isInitialized.value && isOnline.value && !isInitializing.value;
  });

  // Actions
  const initialize = async (): Promise<void> => {
    if (isInitializing.value || isInitialized.value) return;
    
    isInitializing.value = true;
    
    try {
      console.log('Initializing VirtuMancer application...');
      
      // Initialize UI store first (loads preferences, sets up listeners)
      uiStore.loadPreferences();
      uiStore.initializeEventListeners();
      
      // Initialize WebSocket connections
      await initializeWebSocket();
      
      // Load initial data
      await Promise.allSettled([
        hostStore.fetchHosts(),
        vmStore.fetchVMs()
      ]);
      
      // Set up auto-refresh timers
      setupAutoRefresh();
      
      // Set up network status monitoring
      setupNetworkMonitoring();
      
      isInitialized.value = true;
      lastSyncTime.value = new Date();
      
      console.log('VirtuMancer application initialized successfully');
      
      // Show success notification
      uiStore.addToast('Application initialized successfully', 'success', 3000);
      
    } catch (error) {
      console.error('Failed to initialize application:', error);
      addGlobalError({
        message: 'Failed to initialize application',
        code: 'INIT_ERROR',
        details: error,
        timestamp: new Date()
      });
      
      uiStore.addToast('Failed to initialize application', 'error', 0);
      throw error;
    } finally {
      isInitializing.value = false;
    }
  };

  const syncData = async (force: boolean = false): Promise<void> => {
    if (isSyncing.value && !force) return;
    
    isSyncing.value = true;
    
    try {
      console.log('Syncing application data...');
      
      // Fetch latest data from all stores
      await Promise.allSettled([
        hostStore.fetchHosts(),
        vmStore.fetchVMs()
      ]);
      
      // Update stats for connected hosts
      const connectedHostIds = hostStore.connectedHosts.map(h => h.id);
      await Promise.allSettled(
        connectedHostIds.map(id => hostStore.fetchHostStats(id))
      );
      
      // Update VM stats for running VMs
      await vmStore.fetchAllVMStats();
      
      lastSyncTime.value = new Date();
      
      console.log('Data sync completed');
      
    } catch (error) {
      console.error('Data sync failed:', error);
      addGlobalError({
        message: 'Data synchronization failed',
        code: 'SYNC_ERROR',
        details: error,
        timestamp: new Date()
      });
    } finally {
      isSyncing.value = false;
    }
  };

  const initializeWebSocket = async (): Promise<void> => {
    try {
      await wsManager.connect();
      
      // Initialize WebSocket listeners for all stores
      hostStore.initializeWebSocket();
      vmStore.initializeWebSocket();
      
      // Add global WebSocket event handlers
      wsManager.on('error', (error: any) => {
        addGlobalError({
          message: 'WebSocket error occurred',
          code: 'WS_ERROR',
          details: error,
          timestamp: new Date()
        });
      });
      
      wsManager.on('disconnect', () => {
        uiStore.addToast('Lost connection to server', 'warning', 5000);
      });
      
      wsManager.on('reconnect', () => {
        uiStore.addToast('Reconnected to server', 'success', 3000);
        // Trigger data sync after reconnection
        syncData();
      });
      
    } catch (error) {
      console.warn('WebSocket connection failed, continuing without real-time updates:', error);
      // Don't throw here, app can work without WebSocket
    }
  };

  const setupAutoRefresh = (): void => {
    // Refresh data every 30 seconds
    setInterval(() => {
      if (isReady.value && !isSyncing.value) {
        syncData();
      }
    }, 30000);
    
    // Refresh stats more frequently (every 10 seconds)
    setInterval(() => {
      if (isReady.value && hostStore.connectedHosts.length > 0) {
        const connectedIds = hostStore.connectedHosts.map(h => h.id);
        Promise.allSettled(
          connectedIds.map(id => hostStore.fetchHostStats(id))
        );
      }
    }, 10000);
  };

  const setupNetworkMonitoring = (): void => {
    window.addEventListener('online', () => {
      isOnline.value = true;
      uiStore.addToast('Connection restored', 'success', 3000);
      syncData(true);
    });
    
    window.addEventListener('offline', () => {
      isOnline.value = false;
      uiStore.addToast('Connection lost - working offline', 'warning', 5000);
    });
  };

  // Error management
  const addGlobalError = (error: AppError): void => {
    globalErrors.value.push(error);
    
    // Keep only the last 50 errors
    if (globalErrors.value.length > 50) {
      globalErrors.value = globalErrors.value.slice(-50);
    }
  };

  const clearGlobalError = (index: number): void => {
    globalErrors.value.splice(index, 1);
  };

  const clearAllGlobalErrors = (): void => {
    globalErrors.value = [];
  };

  // Emergency actions
  const emergencyStop = async (): Promise<void> => {
    console.log('Emergency stop initiated');
    uiStore.addToast('Emergency stop initiated', 'warning', 0);
    
    try {
      // Stop all running VMs
      await vmStore.stopAllVMs();
      uiStore.addToast('All VMs stopped successfully', 'success', 5000);
    } catch (error) {
      console.error('Emergency stop failed:', error);
      uiStore.addToast('Emergency stop failed - some VMs may still be running', 'error', 0);
    }
  };

  const forceRefresh = async (): Promise<void> => {
    console.log('Force refresh initiated');
    uiStore.setLoading(true);
    
    try {
      await syncData(true);
      uiStore.addToast('Data refreshed successfully', 'success', 3000);
    } catch (error) {
      console.error('Force refresh failed:', error);
      uiStore.addToast('Failed to refresh data', 'error', 5000);
    } finally {
      uiStore.setLoading(false);
    }
  };

  // Cleanup function
  const cleanup = (): void => {
    console.log('Cleaning up application...');
    
    // Disconnect WebSocket
    wsManager.disconnect();
    
    // Clean up UI event listeners
    uiStore.cleanupEventListeners();
    
    // Clear timers and intervals would happen here if we tracked them
    
    isInitialized.value = false;
  };

  // Export debug information
  const getDebugInfo = () => {
    return {
      version: appVersion.value,
      initialized: isInitialized.value,
      online: isOnline.value,
      lastSync: lastSyncTime.value,
      stats: appStats.value,
      connectionStatus: connectionStatus.value,
      healthStatus: healthStatus.value,
      globalErrors: globalErrors.value.length,
      stores: {
        hosts: hostStore.hosts.length,
        vms: vmStore.vms.length,
        toasts: uiStore.toasts.length
      }
    };
  };

  return {
    // State
    isInitialized: readonly(isInitialized),
    isOnline: readonly(isOnline),
    lastSyncTime: readonly(lastSyncTime),
    globalErrors: readonly(globalErrors),
    appVersion: readonly(appVersion),
    isInitializing: readonly(isInitializing),
    isSyncing: readonly(isSyncing),
    
    // Computed
    appStats,
    connectionStatus,
    healthStatus,
    isReady,
    
    // Actions
    initialize,
    syncData,
    initializeWebSocket,
    
    // Error management
    addGlobalError,
    clearGlobalError,
    clearAllGlobalErrors,
    
    // Emergency actions
    emergencyStop,
    forceRefresh,
    
    // Utilities
    cleanup,
    getDebugInfo
  };
});