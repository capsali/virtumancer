<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed, ref, onMounted, watch, defineProps, nextTick } from 'vue';
import ConfirmModal from '@/components/modals/ConfirmModal.vue';
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';
import { useVmStateDisplay } from '@/composables/useVmStateDisplay';

const mainStore = useMainStore();
const route = useRoute();
const router = useRouter();
const { getVmDisplayState, isVmRunning } = useVmStateDisplay();
const props = defineProps({ hostId: { type: String, default: null } });

const selectedHost = computed(() => {
  const hostId = route.params.hostId;
  if (!hostId) return null;
  return mainStore.hosts.find(h => h.id === hostId);
});

const hostStats = computed(() => {
    if (!selectedHost.value) return null;
    return mainStore.hostStats[selectedHost.value.id];
});

const hostId = computed(() => selectedHost.value?.id || props.hostId || route.params.hostId || '');

const vms = computed(() => {
    return selectedHost.value?.vms || [];
});

const discoveredVMs = computed(() => mainStore.discoveredByHost[route.params.hostId] || []);
const showConfirm = ref(false);
const confirmPayload = ref({});
const confirmLoading = ref(false);

// Connect/disconnect confirmation modal state
const showConnectConfirm = ref(false);
const connectConfirmPayload = ref({});
const connectConfirmLoading = ref(false);

const openConfirmAll = () => {
  const hostId = selectedHost.value?.id || props.hostId || route.params.hostId;
  confirmPayload.value = { type: 'all', hostId };
  showConfirm.value = true;
};

const openConfirmOne = (vmName) => {
  const hostId = selectedHost.value?.id || props.hostId || route.params.hostId;
  confirmPayload.value = { type: 'one', hostId, vmName };
  showConfirm.value = true;
};

const handleConfirm = async () => {
    try {
    console.log('[HostDashboard] handleConfirm invoked, payload=', confirmPayload.value);
    if (!confirmPayload.value) return;
    confirmLoading.value = true;
    if (confirmPayload.value.type === 'all') {
      console.log('[HostDashboard] calling importAllVMs for host', confirmPayload.value.hostId);
      mainStore.addToast(`Starting import of all discovered VMs on ${confirmPayload.value.hostId}`, 'success');
      await mainStore.importAllVMs(confirmPayload.value.hostId);
    } else if (confirmPayload.value.type === 'one') {
      console.log('[HostDashboard] calling importVm for', confirmPayload.value.vmName, 'on host', confirmPayload.value.hostId);
      mainStore.addToast(`Starting import of ${confirmPayload.value.vmName}`, 'success');
      await mainStore.importVm(confirmPayload.value.hostId, confirmPayload.value.vmName);
    }
  } catch (e) {
    console.error('Import failed', e);
  } finally {
    confirmLoading.value = false;
    showConfirm.value = false;
    confirmPayload.value = {};
  }
};

const handleCancel = () => {
  showConfirm.value = false;
  confirmPayload.value = {};
};

const openConnectConfirm = (action) => {
  console.log('[HostDashboard] openConnectConfirm called with action:', action);
  const hid = selectedHost.value?.id || props.hostId || route.params.hostId;
  console.log('[HostDashboard] hostId:', hid, 'selectedHost:', selectedHost.value);
  connectConfirmPayload.value = { action, hostId: hid };
  showConnectConfirm.value = true;
  console.log('[HostDashboard] showConnectConfirm set to true');
};

const connectAction = computed(() => {
  if (!selectedHost.value) return 'connect';
  return (selectedHost.value.state === 'CONNECTED' || (selectedHost.value.info && selectedHost.value.info.connected)) ? 'disconnect' : 'connect';
});

const handleHeaderConnectClick = (ev) => {
  console.log('[HostDashboard] handleHeaderConnectClick called, connectAction:', connectAction.value, 'selectedHost:', selectedHost.value);
  openConnectConfirm(connectAction.value);
};

const handleConnectConfirm = async () => {
  if (!connectConfirmPayload.value) return;
  connectConfirmLoading.value = true;
  const { action, hostId } = connectConfirmPayload.value;
  // proceed with requested action
  try {
    if (action === 'connect') {
  await mainStore.connectHost(hostId);
    } else if (action === 'disconnect') {
  await mainStore.disconnectHost(hostId);
    }
  } catch (e) {
    console.error('Connect/Disconnect action failed', e);
  } finally {
    // Refresh discovered VMs after connect/disconnect
    mainStore.refreshDiscoveredVMs(hostId).catch(() => {});
    connectConfirmLoading.value = false;
    showConnectConfirm.value = false;
    connectConfirmPayload.value = {};
  }
};

const handleConnectCancel = () => {
  showConnectConfirm.value = false;
  connectConfirmPayload.value = {};
};

// (removed pointer debug handler)

onMounted(async () => {
  if (route.params.hostId) {
    // Ensure hosts are loaded
    if (mainStore.hosts.length === 0) {
      await mainStore.fetchHosts();
    }
  // Prime centralized discovered cache for this host
  mainStore.refreshDiscoveredVMs(route.params.hostId).catch(() => {});
  // Mark this host as selected in the central store so subscription/connect logic can use it
  if (route.params.hostId) mainStore.selectHost(route.params.hostId);
  }
});

const loadHostPortsAndAttachments = async (hostId) => {
  if (!hostId) {
    return;
  }
  // Network data moved to NetworkView - no longer needed here
}

watch(() => route.params.hostId, async (newId, oldId) => {
  if (newId === oldId) return;
  // Update central selected host so store knows which host is active
  if (newId) mainStore.selectHost(newId);
  // Prime/refresh centralized cache
  mainStore.refreshDiscoveredVMs(newId).catch(() => {});
  // (re)subscribe to host stats for the new host
  if (newId) {
    if (!selectedHost.value) {
      mainStore.fetchHosts().then(() => {
        if (newId) mainStore.subscribeHostStats(newId);
      }).catch(err => console.error('[HostDashboard] fetchHosts failed', err));
    } else {
      mainStore.subscribeHostStats(newId);
    }
  }
});

// Refresh discovered VM list when host VMs change
// Debug: log when discoveredVMs updates and after DOM updates
watch(discoveredVMs, (nv) => {
  nextTick(() => console.log('[HostDashboard] discoveredVMs rendered, count=', (nv && nv.length) || 0));
}, { immediate: true });

const totalMemory = computed(() => {
    return selectedHost.value?.info?.memory || 0;
});

const usedMemory = computed(() => {
    return hostStats.value?.memory_used || selectedHost.value?.info?.memory_used || 0;
});

const memoryUsagePercent = computed(() => {
    if (!totalMemory.value) return 0;
    return (usedMemory.value / totalMemory.value) * 100;
});

const totalCpu = computed(() => {
    return selectedHost.value?.info?.cpu || 0;
});

const vCpuAllocation = computed(() => {
    if (!selectedHost.value || !selectedHost.value.vms) return 0;
    return selectedHost.value.vms.reduce((total, vm) => total + (isVmRunning(vm, selectedHost.value) ? vm.vcpu_count : 0), 0);
});

const vCpuAllocationPercent = computed(() => {
    if (!totalCpu.value) return 0;
    return (vCpuAllocation.value / totalCpu.value) * 100;
});

const cpuUtilization = computed(() => {
    if (!hostStats.value) return 0;
    return (hostStats.value.cpu_utilization * 100).toFixed(1);
});

const vmsWithDrift = computed(() => {
    if (!selectedHost.value || !selectedHost.value.vms) return [];
    return selectedHost.value.vms.filter(vm => getVmDisplayState(vm, selectedHost.value).hasDrift);
});

const selectVm = (vmName) => {
    router.push({ name: 'vm-view', params: { vmName } });
}

const revertVmState = async (vm) => {
    // Revert: Change libvirt state to match intended DB state
    const displayState = getVmDisplayState(vm, selectedHost.value);
    if (displayState.intendedState === 'ACTIVE' || displayState.intendedState === 'RUNNING') {
        await mainStore.startVm(selectedHost.value.id, vm.name);
    } else if (displayState.intendedState === 'STOPPED') {
        await mainStore.gracefulShutdownVm(selectedHost.value.id, vm.name);
    }
    // Refresh the VM data
    await mainStore.fetchHosts();
};

const acceptVmState = async (vm) => {
    // Accept: Update DB state to match libvirt state
    const displayState = getVmDisplayState(vm, selectedHost.value);
    try {
        const response = await fetch(`/api/v1/hosts/${selectedHost.value.id}/vms/${vm.name}/state`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                state: displayState.observedState
            })
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        // Refresh the VM data
        await mainStore.fetchHosts();
        mainStore.addToast(`Accepted current state for ${vm.name}`, 'success');
    } catch (error) {
        console.error('Failed to accept VM state:', error);
        mainStore.addToast(`Failed to accept state for ${vm.name}`, 'error');
    }
};



onMounted(() => {
  console.log('[HostDashboard] mounted, route hostId=', route.params.hostId);
  if (route.params.hostId) {
    if (!selectedHost.value) {
      console.log('[HostDashboard] selectedHost not found, fetching hosts to populate data');
      mainStore.fetchHosts().then(() => {
        if (route.params.hostId) mainStore.subscribeHostStats(route.params.hostId);
      }).catch(err => console.error('[HostDashboard] fetchHosts failed', err));
    } else {
      mainStore.subscribeHostStats(route.params.hostId);
    }
  }
});

onBeforeRouteLeave((to, from, next) => {
    // Unsubscribe from host stats when leaving this route
    if (from.params.hostId) {
        mainStore.unsubscribeHostStats(from.params.hostId);
    }
    next();
});


// Helper functions for display
const stateText = (vm) => {
    if (!vm) return 'Unknown';
    if (vm.task_state) {
        const task = vm.task_state.toLowerCase().replace(/_/g, ' ');
        // Capitalize first letter
        return task.charAt(0).toUpperCase() + task.slice(1);
    }
    const displayState = getVmDisplayState(vm, selectedHost.value);
    if (displayState.hasDrift) {
        return `${displayState.status} (Drift)`;
    }
    if (displayState.status === 'UNKNOWN') {
        return `Unknown (${displayState.lastKnownState})`;
    }
    const states = {
        'INITIALIZED': 'Initialized',
        'ACTIVE': 'Running',
        'RUNNING': 'Running',
        'PAUSED': 'Paused',
        'SUSPENDED': 'Suspended',
        'STOPPED': 'Stopped',
        'ERROR': 'Error'
    };
    return states[displayState.status] || displayState.status;
};

const stateColor = (vm) => {
  if (!vm) return 'text-gray-400 bg-gray-700';
  if (vm.task_state) {
      return 'text-orange-300 bg-orange-900/50 animate-pulse';
  }
  const displayState = getVmDisplayState(vm, selectedHost.value);
  if (displayState.hasDrift) {
      return 'text-red-400 bg-red-900/50 font-bold';
  }
  const colors = {
    'green': 'text-green-400 bg-green-900/50',
    'yellow': 'text-yellow-400 bg-yellow-900/50',
    'blue': 'text-blue-400 bg-blue-900/50',
    'red': 'text-red-400 bg-red-900/50',
    'gray': 'text-gray-400 bg-gray-700',
  };
  return colors[displayState.color] || 'text-gray-400 bg-gray-700';
};

// Host state helpers (mirror VM helpers)
const hostStateText = (host) => {
    if (!host) return 'Unknown';
    if (host.task_state) {
        const task = host.task_state.toLowerCase().replace(/_/g, ' ');
        return task.charAt(0).toUpperCase() + task.slice(1);
    }
    const states = {
        'CONNECTED': 'Connected',
        'DISCONNECTED': 'Disconnected',
        'ERROR': 'Error'
    };
    if (host.state && states[host.state]) return states[host.state];
    return host && host.info ? (host.connected ? 'Connected' : 'Disconnected') : (host.connected ? 'Connected' : 'Disconnected');
};

const hostStateColor = (host) => {
  if (!host) return 'text-gray-400 bg-gray-700';
  if (host.task_state) return 'text-orange-300 bg-orange-900/50 animate-pulse';
  const colors = {
    'CONNECTED': 'text-green-400 bg-green-900/50',
    'DISCONNECTED': 'text-gray-400 bg-gray-700',
    'ERROR': 'text-red-400 bg-red-900/50 font-bold'
  };
  if (host.state && colors[host.state]) return colors[host.state];
  return host.connected ? 'text-green-400 bg-green-900/50' : 'text-gray-400 bg-gray-700';
};

const formatMemory = (kb) => {
    if (kb === 0) return '0 MB';
    if (!kb) return 'N/A';
    const mb = kb / 1024;
    if (mb < 1024) return `${mb.toFixed(0)} MB`;
    const gb = mb / 1024;
    return `${gb.toFixed(2)} GB`;
};

const formatBytes = (bytes, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';
    if (!bytes) return 'N/A';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

const formatUptime = (sec) => {
  if (sec === null || sec === undefined) return 'N/A';
  if (sec <= 0) return '0s';
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const mins = Math.floor((sec % 3600) / 60);
  const seconds = Math.floor(sec % 60);
  if (days > 0) return `${days}d ${hours}h ${mins}m`;
  if (hours > 0) return `${hours}h ${mins}m ${seconds}s`;
  if (mins > 0) return `${mins}m ${seconds}s`;
  return `${seconds}s`;
};

</script>

<template>
  <div v-if="selectedHost">
  <!-- Header -->
  <div class="mb-6 relative z-20">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <h1 class="text-3xl font-bold text-white">Host: {{ hostId }}</h1>
          <span class="text-sm font-semibold px-3 py-1 rounded-full" :class="hostStateColor(selectedHost)">{{ hostStateText(selectedHost) }}</span>
        </div>
        <div class="flex items-center">
          <p class="text-gray-400 font-mono mr-4">{{ selectedHost.uri }}</p>
          <button
            type="button"
            @click="handleHeaderConnectClick"
            :disabled="!selectedHost || showConnectConfirm || (mainStore.isLoading.connectHost && mainStore.isLoading.connectHost[hostId])"
            :aria-disabled="!selectedHost || showConnectConfirm || (mainStore.isLoading.connectHost && mainStore.isLoading.connectHost[hostId]) ? 'true' : 'false'"
            :aria-busy="(mainStore.isLoading.connectHost && mainStore.isLoading.connectHost[hostId]) ? 'true' : 'false'"
            :class="selectedHost && (selectedHost.state === 'CONNECTED' || (selectedHost.info && selectedHost.info.connected)) ? 'px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed' : 'px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed'"
          >
            <span v-if="mainStore.isLoading.connectHost && mainStore.isLoading.connectHost[hostId]" class="inline-block animate-spin w-4 h-4 border-2 border-white rounded-full border-t-transparent mr-2" aria-hidden="true"></span>
            {{ selectedHost && (selectedHost.state === 'CONNECTED' || (selectedHost.info && selectedHost.info.connected)) ? 'Disconnect' : 'Connect' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Drift Warnings -->
    <div v-if="vmsWithDrift.length > 0" class="mb-6 bg-red-900/20 border border-red-500/50 rounded-lg p-4">
      <div class="flex items-start">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
          </svg>
        </div>
        <div class="ml-3 flex-1">
          <h3 class="text-sm font-medium text-red-400">
            Configuration Drift Detected
          </h3>
          <div class="mt-2 text-sm text-red-300">
            <p>{{ vmsWithDrift.length }} VM{{ vmsWithDrift.length > 1 ? 's have' : ' has' }} state drift. The intended state doesn't match the observed state from libvirt.</p>
            <ul class="mt-2 space-y-1">
              <li v-for="vm in vmsWithDrift" :key="vm.uuid" class="flex items-center justify-between">
                <span>{{ vm.name }}: Intended {{ getVmDisplayState(vm, selectedHost).intendedState }}, Observed {{ getVmDisplayState(vm, selectedHost).observedState }}</span>
                <div class="flex space-x-2">
                  <button @click="revertVmState(vm)" class="text-xs bg-red-600 hover:bg-red-700 px-2 py-1 rounded text-white">Revert</button>
                  <button @click="acceptVmState(vm)" class="text-xs bg-gray-600 hover:bg-gray-700 px-2 py-1 rounded text-white">Accept</button>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Discovered VMs (not yet managed) -->
    <div class="mt-6 bg-gray-900 rounded-lg p-4">
        <div class="flex items-center justify-between">
        <h2 class="text-xl font-semibold text-white mb-3">Discovered VMs (Not Managed)</h2>
        <div>
          <button
            :disabled="mainStore.isLoading.hostImportAll === `host:${hostId}:import-all`"
            @click.prevent="openConfirmAll"
            :aria-disabled="mainStore.isLoading.hostImportAll === `host:${hostId}:import-all` ? 'true' : 'false'"
            :aria-busy="mainStore.isLoading.hostImportAll === `host:${hostId}:import-all` ? 'true' : 'false'"
            :aria-label="`Import all discovered VMs on host ${hostId}`"
            :class="[
              'px-3 py-1 rounded text-white mr-2',
              mainStore.isLoading.hostImportAll === `host:${hostId}:import-all` ? 'bg-indigo-600 opacity-50 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700'
            ]"
          >
            <span v-if="mainStore.isLoading.hostImportAll === `host:${hostId}:import-all`" class="inline-block animate-spin w-4 h-4 border-2 border-white rounded-full border-t-transparent mr-1" aria-hidden="true"></span>
            Import All
          </button>
        </div>
      </div>
      <div v-if="discoveredVMs.length === 0" class="text-gray-400">No discovered VMs found for this host.</div>
          <ul v-else class="space-y-2">
        <li v-for="d in discoveredVMs" :key="d.uuid" class="bg-gray-800 p-3 rounded flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-300">Name: <span class="font-medium text-white">{{ d.name }}</span></div>
            <div class="text-xs text-gray-500">UUID: <span class="font-mono">{{ d.uuid }}</span></div>
          </div>
          <div class="flex items-center gap-2">
            <button
              :disabled="mainStore.isLoading.vmImport === `${d.name}:import`"
              @click.prevent="() => openConfirmOne(d.name)"
              :aria-disabled="mainStore.isLoading.vmImport === `${d.name}:import` ? 'true' : 'false'"
              :aria-busy="mainStore.isLoading.vmImport === `${d.name}:import` ? 'true' : 'false'"
              :aria-label="`Import discovered VM ${d.name} on host ${hostId}`"
              :class="[
                'px-3 py-1 rounded text-white',
                mainStore.isLoading.vmImport === `${d.name}:import` ? 'bg-green-600 opacity-50 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'
              ]"
            >
              <span v-if="mainStore.isLoading.vmImport === `${d.name}:import`" class="inline-block animate-spin w-4 h-4 border-2 border-white rounded-full border-t-transparent mr-1" aria-hidden="true"></span>
              Import
            </button>
          </div>
        </li>
      </ul>
    </div>
    
    <!-- Summary Section -->
    <div class="mb-6">
        <h2 class="text-xl font-semibold text-white mb-4">Summary</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <!-- vCPU Usage -->
            <div class="bg-gray-800 p-4 rounded-lg">
                <h3 class="text-sm font-medium text-gray-400">vCPU Allocation</h3>
                <p class="text-2xl font-semibold text-white mt-1">{{ vCpuAllocation }} / {{ totalCpu }} Cores</p>
                <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                    <div class="bg-indigo-500 h-2.5 rounded-full" :style="{ width: vCpuAllocationPercent + '%' }"></div>
                </div>
            </div>
            
            <!-- CPU Utilization -->
            <div class="bg-gray-800 p-4 rounded-lg">
                <h3 class="text-sm font-medium text-gray-400">CPU Utilization</h3>
                <p class="text-2xl font-semibold text-white mt-1">{{ cpuUtilization }}%</p>
                <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                    <div class="bg-green-500 h-2.5 rounded-full" :style="{ width: cpuUtilization + '%' }"></div>
                </div>
            </div>
            <!-- Memory Usage -->
            <div class="bg-gray-800 p-4 rounded-lg">
                <h3 class="text-sm font-medium text-gray-400">Memory Usage</h3>
                <p class="text-2xl font-semibold text-white mt-1">{{ formatBytes(usedMemory) }} / {{ formatBytes(totalMemory) }}</p>
                <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                    <div class="bg-teal-500 h-2.5 rounded-full" :style="{ width: memoryUsagePercent + '%' }"></div>
                </div>
            </div>
            <!-- Other Host Info -->
             <div class="bg-gray-800 p-4 rounded-lg">
                <h3 class="text-sm font-medium text-gray-400">Hostname</h3>
                <p class="text-2xl font-semibold text-white mt-1 truncate">{{ selectedHost.info?.hostname || 'Loading...' }}</p>
            </div>
       <div class="bg-gray-800 p-4 rounded-lg">
        <h3 class="text-sm font-medium text-gray-400">Uptime</h3>
        <p class="text-2xl font-semibold text-white mt-1">{{ formatUptime(selectedHost.info?.uptime) }}</p>
      </div>
             <div class="bg-gray-800 p-4 rounded-lg">
                <h3 class="text-sm font-medium text-gray-400">Cores / Threads</h3>
                <p class="text-2xl font-semibold text-white mt-1">{{ selectedHost.info?.cores || 'N/A' }} / {{ selectedHost.info?.threads || 'N/A' }}</p>
            </div>
        </div>
    </div>

    <!-- VM List Section -->
    <div class="bg-gray-900 rounded-lg">
      <h2 class="text-xl font-semibold text-white p-4">Virtual Machines</h2>
      
      <div v-if="mainStore.isLoading.vms && vms.length === 0" class="flex items-center justify-center h-48 text-gray-400">
        <svg class="animate-spin mr-3 h-8 w-8 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>Loading VMs...</span>
      </div>

      <div v-else-if="vms.length === 0" class="flex items-center justify-center h-48 text-gray-500 bg-gray-800/50 rounded-lg m-4">
        <p>No Virtual Machines found on this host.</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-700">
          <thead class="bg-gray-800">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Name</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">State</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">vCPUs</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Memory</th>
            </tr>
          </thead>
          <tbody class="bg-gray-900 divide-y divide-gray-800">
            <tr v-for="vm in vms" :key="vm.uuid" @click="selectVm(vm.name)" class="hover:bg-gray-800 cursor-pointer transition-colors duration-150">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="h-2.5 w-2.5 rounded-full mr-3 flex-shrink-0" :class="{
                    'bg-green-500': getVmDisplayState(vm, selectedHost).color === 'green',
                    'bg-red-500': getVmDisplayState(vm, selectedHost).color === 'red',
                    'bg-yellow-500': getVmDisplayState(vm, selectedHost).color === 'yellow',
                    'bg-blue-500': getVmDisplayState(vm, selectedHost).color === 'blue',
                    'bg-gray-500': getVmDisplayState(vm, selectedHost).color === 'gray',
                    'bg-orange-500 animate-pulse': vm.task_state,
                  }"></div>
                  <div class="text-sm font-medium text-white">{{ vm.name }}</div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full" :class="stateColor(vm)">
                  {{ stateText(vm) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ vm.vcpu_count }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ formatBytes(vm.memory_bytes) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
  <div v-else class="flex items-center justify-center h-full text-gray-500">
    <p>Select a host from the sidebar to view details.</p>
  </div>
  <ConfirmModal v-if="showConfirm" :title="confirmPayload.type === 'all' ? 'Import all VMs?' : 'Import VM?'" :message="confirmPayload.type === 'all' ? 'Import all discovered VMs on this host into management. This will create DB records for each VM.' : `Import VM \'${confirmPayload.vmName || ''}\' into management?`" confirmText="Import" cancelText="Cancel" :loading="confirmLoading" @confirm="handleConfirm" @cancel="handleCancel" />
  <ConfirmModal v-if="showConnectConfirm" :title="connectConfirmPayload.action === 'disconnect' ? 'Disconnect host?' : 'Connect host?'" :message="connectConfirmPayload.action === 'disconnect' ? 'Disconnecting will close the libvirt connection for this host. Open consoles and live stats may stop working.' : 'Attempt to establish a libvirt connection to this host now?'" confirmText="Confirm" cancelText="Cancel" :loading="connectConfirmLoading" @confirm="handleConnectConfirm" @cancel="handleConnectCancel" />


</template>

