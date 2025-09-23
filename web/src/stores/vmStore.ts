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
      const data = await vmApi.getAll(hostId);
      if (hostId) {
        // Update only VMs for this host
        vms.value = vms.value.filter(vm => vm.hostId !== hostId).concat(data);
      } else {
        vms.value = data;
      }
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
  const startVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'STARTING', () => vmApi.start(uuid));
  };

  const stopVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'STOPPING', () => vmApi.stop(uuid));
  };

  const restartVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'REBOOTING', () => vmApi.restart(uuid));
  };

  const pauseVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'PAUSING', () => vmApi.pause(uuid));
  };

  const unpauseVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'UNPAUSING', () => vmApi.unpause(uuid));
  };

  const suspendVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'SUSPENDING', () => vmApi.suspend(uuid));
  };

  const resumeVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'RESUMING', () => vmApi.resume(uuid));
  };

  const reconcileVM = async (uuid: string): Promise<void> => {
    return performVMAction(uuid, 'SCHEDULING', () => vmApi.reconcile(uuid));
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
  const fetchVMStats = async (uuid: string): Promise<void> => {
    loading.value.vmStats[uuid] = true;
    clearError('fetchVMStats');
    
    try {
      const stats = await vmApi.getStats(uuid);
      vmStats.value[uuid] = stats;
    } catch (error) {
      handleError('fetchVMStats', error);
      // Don't throw here, stats are optional
    } finally {
      loading.value.vmStats[uuid] = false;
    }
  };

  const fetchVMHardware = async (uuid: string): Promise<void> => {
    loading.value.vmHardware[uuid] = true;
    clearError('fetchVMHardware');
    
    try {
      const hardware = await vmApi.getHardware(uuid);
      vmHardware.value[uuid] = hardware;
    } catch (error) {
      handleError('fetchVMHardware', error);
      throw error;
    } finally {
      loading.value.vmHardware[uuid] = false;
    }
  };

  const updateVMHardware = async (uuid: string, hardware: Partial<VMHardware>): Promise<void> => {
    loading.value.vmHardware[uuid] = true;
    clearError('updateVMHardware');
    
    try {
      const updatedHardware = await vmApi.updateHardware(uuid, hardware);
      vmHardware.value[uuid] = updatedHardware;
    } catch (error) {
      handleError('updateVMHardware', error);
      throw error;
    } finally {
      loading.value.vmHardware[uuid] = false;
    }
  };

  const importVM = async (hostId: string, vmUuid: string): Promise<VirtualMachine> => {
    clearError('importVM');
    
    try {
      const importedVM = await vmApi.import(hostId, vmUuid);
      vms.value.push(importedVM);
      return importedVM;
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
  const fetchAllVMStats = async (): Promise<void> => {
    const activeVMList = vms.value.filter(vm => vm.state === 'ACTIVE');
    const promises = activeVMList.map(vm => fetchVMStats(vm.uuid));
    await Promise.allSettled(promises);
  };

  const stopAllVMs = async (hostId?: string): Promise<void> => {
    const targetVMs = hostId 
      ? vms.value.filter(vm => vm.hostId === hostId && vm.state === 'ACTIVE')
      : vms.value.filter(vm => vm.state === 'ACTIVE');
    
    const promises = targetVMs.map(vm => stopVM(vm.uuid));
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
    pauseVM,
    unpauseVM,
    suspendVM,
    resumeVM,
    reconcileVM,
    
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