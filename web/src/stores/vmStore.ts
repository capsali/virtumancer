import { defineStore } from 'pinia';
import { ref, computed, readonly } from 'vue';
import type { 
  VirtualMachine, 
  VMStats, 
  VMHardware,
  AppError,
  VMTaskState 
} from '@/types';
import { vmApi, wsManager, ApiError } from '@/services/api';
import { errorRecoveryService } from '@/services/errorRecovery';

export const useVMStore = defineStore('virtualMachines', () => {
  // State
  const vms = ref<VirtualMachine[]>([]);
  const selectedVMId = ref<string | null>(null);
  const vmStats = ref<Record<string, VMStats>>({});
  const vmHardware = ref<Record<string, VMHardware>>({});
  const errors = ref<Record<string, AppError>>({});
  
  // Loading states
  const loading = ref({
    vms: false,
    vmStats: {} as Record<string, boolean>,
    vmHardware: {} as Record<string, boolean>,
    vmActions: {} as Record<string, VMTaskState | null>
  });

  // Computed properties
  const selectedVM = computed((): VirtualMachine | null => {
    if (!selectedVMId.value) return null;
    return vms.value.find(vm => vm.uuid === selectedVMId.value) ?? null;
  });

  const vmsByHost = computed(() => {
    return (hostId: string): VirtualMachine[] => {
      return vms.value.filter(vm => vm.hostId === hostId);
    };
  });

  const activeVMs = computed((): VirtualMachine[] => {
    return vms.value.filter(vm => vm.state === 'ACTIVE');
  });

  const stoppedVMs = computed((): VirtualMachine[] => {
    return vms.value.filter(vm => vm.state === 'STOPPED');
  });

  const errorVMs = computed((): VirtualMachine[] => {
    return vms.value.filter(vm => vm.state === 'ERROR');
  });

  const vmWithStats = computed(() => {
    return (vmId: string) => {
      const vm = vms.value.find(v => v.uuid === vmId);
      if (!vm) return null;
      
      return {
        ...vm,
        stats: vmStats.value[vmId] || null,
        hardware: vmHardware.value[vmId] || null,
        isLoading: loading.value.vmActions[vmId] || false
      };
    };
  });

  const vmsByState = computed(() => {
    return vms.value.reduce((acc, vm) => {
      if (!acc[vm.state]) {
        acc[vm.state] = [];
      }
      acc[vm.state]!.push(vm);
      return acc;
    }, {} as Record<string, VirtualMachine[]>);
  });

  // Actions
  const fetchVMs = async (hostId?: string): Promise<void> => {
    loading.value.vms = true;
    clearError('fetchVMs');
    
    try {
      if (!hostId) {
        // If no hostId provided, clear VMs and return
        vms.value = [];
        return;
      }
      
      const data = await vmApi.getAll(hostId);
      // Update only VMs for this host
      vms.value = vms.value.filter(vm => vm.hostId !== hostId).concat(data);
    } catch (error) {
      handleError('fetchVMs', error);
      throw error;
    } finally {
      loading.value.vms = false;
    }
  };

  const createVM = async (vmData: Omit<VirtualMachine, 'uuid' | 'createdAt' | 'updatedAt'>): Promise<VirtualMachine> => {
    clearError('createVM');
    
    try {
      const newVM = await vmApi.create(vmData);
      vms.value.push(newVM);
      return newVM;
    } catch (error) {
      handleError('createVM', error);
      throw error;
    }
  };

  const updateVM = async (uuid: string, updates: Partial<VirtualMachine>): Promise<void> => {
    clearError('updateVM');
    
    try {
      const updatedVM = await vmApi.update(uuid, updates);
      const index = vms.value.findIndex(vm => vm.uuid === uuid);
      if (index !== -1) {
        vms.value[index] = updatedVM;
      }
    } catch (error) {
      handleError('updateVM', error);
      throw error;
    }
  };

  const deleteVM = async (uuid: string): Promise<void> => {
    clearError('deleteVM');
    
    try {
      await vmApi.delete(uuid);
      vms.value = vms.value.filter(vm => vm.uuid !== uuid);
      
      // Clean up related data
      delete vmStats.value[uuid];
      delete vmHardware.value[uuid];
      delete loading.value.vmStats[uuid];
      delete loading.value.vmHardware[uuid];
      delete loading.value.vmActions[uuid];
      
      // Clear selection if this VM was selected
      if (selectedVMId.value === uuid) {
        selectedVMId.value = null;
      }
    } catch (error) {
      handleError('deleteVM', error);
      throw error;
    }
  };

  // VM Control Actions
  const startVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'STARTING', () => vmApi.start(hostId, vmName));
  };

  const stopVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'STOPPING', () => vmApi.shutdown(hostId, vmName));
  };

  const restartVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'REBOOTING', () => vmApi.reboot(hostId, vmName));
  };

  const forceOffVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'STOPPING', () => vmApi.forceOff(hostId, vmName));
  };

  const forceResetVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'REBOOTING', () => vmApi.forceReset(hostId, vmName));
  };

  const syncVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'SCHEDULING', () => vmApi.sync(hostId, vmName));
  };

  const rebuildVM = async (hostId: string, vmName: string): Promise<void> => {
    return performVMAction(`${hostId}:${vmName}`, 'REBUILDING', () => vmApi.rebuild(hostId, vmName));
  };

  // Helper function for VM actions
  const performVMAction = async (
    uuid: string, 
    taskState: VMTaskState, 
    action: () => Promise<void>
  ): Promise<void> => {
    loading.value.vmActions[uuid] = taskState;
    clearError('vmAction');
    
    // Optimistically update task state
    const vm = vms.value.find(v => v.uuid === uuid);
    if (vm) {
      vm.taskState = taskState;
    }
    
    try {
      await action();
    } catch (error) {
      // Revert optimistic update on error
      if (vm) {
        vm.taskState = undefined;
      }
      handleError('vmAction', error);
      throw error;
    } finally {
      loading.value.vmActions[uuid] = null;
    }
  };

  // Data fetching actions
  const fetchVMStats = async (hostId: string, vmName: string): Promise<void> => {
    const key = `${hostId}:${vmName}`;
    loading.value.vmStats[key] = true;
    clearError('fetchVMStats');
    
    try {
      const stats = await vmApi.getStats(hostId, vmName);
      vmStats.value[key] = stats;
    } catch (error) {
      handleError('fetchVMStats', error);
      // Don't throw here, stats are optional
    } finally {
      loading.value.vmStats[key] = false;
    }
  };

  const fetchVMHardware = async (hostId: string, vmName: string): Promise<void> => {
    const key = `${hostId}:${vmName}`;
    loading.value.vmHardware[key] = true;
    clearError('fetchVMHardware');
    
    try {
      const hardware = await vmApi.getHardware(hostId, vmName);
      vmHardware.value[key] = hardware;
    } catch (error) {
      handleError('fetchVMHardware', error);
      throw error;
    } finally {
      loading.value.vmHardware[key] = false;
    }
  };

  const updateVMHardware = async (hostId: string, vmName: string, hardware: Partial<VMHardware>): Promise<void> => {
    const key = `${hostId}:${vmName}`;
    loading.value.vmHardware[key] = true;
    clearError('updateVMHardware');
    
    try {
      // For now, just store locally until backend supports this
      console.log('Hardware update requested:', { hostId, vmName, hardware });
      // TODO: Implement when backend supports hardware updates
    } catch (error) {
      handleError('updateVMHardware', error);
      throw error;
    } finally {
      loading.value.vmHardware[key] = false;
    }
  };

  const importVM = async (hostId: string, vmName: string): Promise<void> => {
    clearError('importVM');
    
    try {
      await vmApi.import(hostId, vmName);
      // Refresh VMs list after import
      await fetchVMs(hostId);
    } catch (error) {
      handleError('importVM', error);
      throw error;
    }
  };

  // Helper functions
  const selectVM = (uuid: string | null): void => {
    selectedVMId.value = uuid;
  };

  const getVMById = (uuid: string): VirtualMachine | undefined => {
    return vms.value.find(vm => vm.uuid === uuid);
  };

  const getVMsByState = (state: string): VirtualMachine[] => {
    return vms.value.filter(vm => vm.state === state);
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
    console.error(`VM store error in ${operation}:`, appError);

    // Also add to error recovery service for enhanced handling
    errorRecoveryService.addError(
      error as Error,
      operation,
      {
        store: 'vmStore',
        selectedVMId: selectedVMId.value,
        vmCount: vms.value.length
      }
    );
  };

  // WebSocket integration for real-time updates
  const initializeWebSocket = (): void => {
    wsManager.on('vm_updated', (vmData: VirtualMachine) => {
      const index = vms.value.findIndex(vm => vm.uuid === vmData.uuid);
      if (index !== -1) {
        vms.value[index] = { ...vms.value[index], ...vmData };
      }
    });

    wsManager.on('vm_created', (vmData: VirtualMachine) => {
      const exists = vms.value.some(vm => vm.uuid === vmData.uuid);
      if (!exists) {
        vms.value.push(vmData);
      }
    });

    wsManager.on('vm_deleted', (data: { uuid: string }) => {
      vms.value = vms.value.filter(vm => vm.uuid !== data.uuid);
      delete vmStats.value[data.uuid];
      delete vmHardware.value[data.uuid];
    });

    wsManager.on('vm_stats', (data: { uuid: string; stats: VMStats }) => {
      vmStats.value[data.uuid] = data.stats;
    });
  };

  // Bulk operations
  const fetchAllVMStats = async (hostId?: string): Promise<void> => {
    const activeVMList = hostId 
      ? vms.value.filter(vm => vm.hostId === hostId && vm.state === 'ACTIVE')
      : vms.value.filter(vm => vm.state === 'ACTIVE');
    
    const promises = activeVMList.map(vm => fetchVMStats(vm.hostId, vm.name));
    await Promise.allSettled(promises);
  };

  const stopAllVMs = async (hostId?: string): Promise<void> => {
    const targetVMs = hostId 
      ? vms.value.filter(vm => vm.hostId === hostId && vm.state === 'ACTIVE')
      : vms.value.filter(vm => vm.state === 'ACTIVE');
    
    const promises = targetVMs.map(vm => stopVM(vm.hostId, vm.name));
    await Promise.allSettled(promises);
  };

  return {
    // State
    vms: readonly(vms),
    selectedVMId,
    vmStats: readonly(vmStats),
    vmHardware: readonly(vmHardware),
    loading: readonly(loading),
    errors: readonly(errors),
    
    // Computed
    selectedVM,
    vmsByHost,
    activeVMs,
    stoppedVMs,
    errorVMs,
    vmWithStats,
    vmsByState,
    
    // Actions
    fetchVMs,
    createVM,
    updateVM,
    deleteVM,
    
    // VM Control
    startVM,
    stopVM,
    restartVM,
    forceOffVM,
    forceResetVM,
    syncVM,
    rebuildVM,
    
    // Data fetching
    fetchVMStats,
    fetchVMHardware,
    updateVMHardware,
    fetchAllVMStats,
    
    // Import/Export
    importVM,
    
    // Bulk operations
    stopAllVMs,
    
    // Utilities
    selectVM,
    getVMById,
    getVMsByState,
    clearError,
    initializeWebSocket
  };
});