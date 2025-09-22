<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed, ref, watch, onMounted } from 'vue';
import { useRoute, onBeforeRouteLeave } from 'vue-router';
import VncConsole from '@/components/consoles/VncConsole.vue';
import SpiceConsole from '@/components/consoles/SpiceConsole.vue';
import { useVmStateDisplay } from '@/composables/useVmStateDisplay';

const mainStore = useMainStore();
const route = useRoute();
const { getVmDisplayState } = useVmStateDisplay();
const activeTab = ref('summary');

const vm = computed(() => {
    if (!route.params.vmName) return null;
    for (const host of mainStore.hosts) {
        const foundVm = (host.vms || []).find(v => v.name === route.params.vmName);
        if (foundVm) return foundVm;
    }
    return null;
});

const host = computed(() => {
    if (!vm.value) return null;
    return mainStore.hosts.find(h => h.vms && h.vms.some(v => v.name === vm.value.name));
});

const uuidConflict = computed(() => {
    if (!vm.value || !vm.value.uuid || !vm.value.domain_uuid) return false;
    // A conflict exists if our internal UUID is different from the domain's UUID.
    return vm.value.uuid !== vm.value.domain_uuid;
});

const stats = computed(() => {
    const s = mainStore.activeVmStats;
    if (s && host.value && vm.value && s.hostId === host.value.id && s.vmName === vm.value.name) {
        return s.stats;
    }
    // Return a default structure if no stats are available to prevent template errors
    return {
        state: -1, // Libvirt integer state
        max_mem: vm.value?.max_mem ?? 0,
        memory: 0,
        vcpu: vm.value?.vcpu ?? 0,
        cpu_time: 0,
        disk_stats: [],
        net_stats: []
    };
});

const hardware = computed(() => mainStore.activeVmHardware);

// --- Real-time Stat Calculation ---
const lastCpuTime = ref(0);
const lastCpuTimeTimestamp = ref(0);
const cpuUsagePercent = ref(0);
const lastIoStats = ref(null);
const lastIoStatsTimestamp = ref(0);
const diskRates = ref({});
const netRates = ref({});

watch(stats, (newStats) => {
    if (!newStats || newStats.state !== 1) { // state 1 is DomainRunning
        cpuUsagePercent.value = 0;
        diskRates.value = {};
        netRates.value = {};
        return;
    }
    
    const now = Date.now();

    // CPU Usage
    if (lastCpuTime.value > 0 && lastCpuTimeTimestamp.value > 0) {
        const timeDelta = now - lastCpuTimeTimestamp.value; // ms
        const cpuTimeDelta = newStats.cpu_time - lastCpuTime.value; // ns
        if (timeDelta > 0) {
            const timeDeltaNs = timeDelta * 1_000_000;
            const usage = (cpuTimeDelta / (timeDeltaNs * newStats.vcpu)) * 100;
            cpuUsagePercent.value = Math.min(Math.max(usage, 0), 100);
        }
    }
    lastCpuTime.value = newStats.cpu_time;
    lastCpuTimeTimestamp.value = now;

    // I/O Usage
    if (lastIoStats.value && lastIoStatsTimestamp.value > 0) {
        const timeDeltaSeconds = (now - lastIoStatsTimestamp.value) / 1000;
        if (timeDeltaSeconds > 0) {
            // Disk Rates
            const newDiskRates = {};
            (newStats.disk_stats || []).forEach(current => {
                const previous = (lastIoStats.value.disk_stats || []).find(p => p.device === current.device);
                if (previous) {
                    newDiskRates[current.device] = {
                        read: (current.read_bytes - previous.read_bytes) / timeDeltaSeconds,
                        write: (current.write_bytes - previous.write_bytes) / timeDeltaSeconds,
                    };
                }
            });
            diskRates.value = newDiskRates;

            // Network Rates
            const newNetRates = {};
             (newStats.net_stats || []).forEach(current => {
                const previous = (lastIoStats.value.net_stats || []).find(p => p.device === current.device);
                if (previous) {
                    newNetRates[current.device] = {
                        read: (current.read_bytes - previous.read_bytes) / timeDeltaSeconds,
                        write: (current.write_bytes - previous.write_bytes) / timeDeltaSeconds,
                    };
                }
            });
            netRates.value = newNetRates;
        }
    }
    lastIoStats.value = JSON.parse(JSON.stringify(newStats)); // Deep copy for next calculation
    lastIoStatsTimestamp.value = now;
});


const memoryUsagePercent = computed(() => {
    if (!stats.value || !stats.value.max_mem || stats.value.state !== 1) return 0;
    return (stats.value.memory / stats.value.max_mem) * 100;
});

const isTaskActive = computed(() => !!vm.value?.task_state);
const isReconcileLoading = computed(() => !!mainStore.isLoading.vmReconcile);

const driftDetails = computed(() => {
    if (vm.value?.sync_status !== 'DRIFTED' || !vm.value.drift_details) return null;
    try {
        return JSON.parse(vm.value.drift_details);
    } catch (e) {
        console.error("Failed to parse drift details:", e);
        return { "error": "Could not parse details." };
    }
});

// --- Helper functions ---
const stateText = (vm) => {
    if (!vm) return 'Unknown';
    if (vm.task_state) {
        const task = vm.task_state.toLowerCase().replace(/_/g, ' ');
        // Capitalize first letter
        return task.charAt(0).toUpperCase() + task.slice(1);
    }
    const displayState = getVmDisplayState(vm, host.value);
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
  const displayState = getVmDisplayState(vm, host.value);
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

const formatMemory = (kb) => {
    if (!kb || kb === 0) return '0 MB';
    const mb = kb / 1024;
    if (mb < 1024) return `${mb.toFixed(0)} MB`;
    const gb = mb / 1024;
    return `${gb.toFixed(2)} GB`;
};

const formatUptime = (seconds) => {
    if (seconds <= 0) return 'N/A';
    let d = Math.floor(seconds / (3600*24));
    let h = Math.floor(seconds % (3600*24) / 3600);
    let m = Math.floor(seconds % 3600 / 60);
    let s = Math.floor(seconds % 60);
    
    let dDisplay = d > 0 ? d + (d == 1 ? " day, " : " days, ") : "";
    let hDisplay = h > 0 ? h + (h == 1 ? " hr, " : " hrs, ") : "";
    let mDisplay = m > 0 ? m + (m == 1 ? " min, " : " mins, ") : "";
    let sDisplay = s + (s == 1 ? " sec" : " secs");
    if (d > 0) return dDisplay + hDisplay + mDisplay;
    if (h > 0) return hDisplay + mDisplay;
    if (m > 0) return mDisplay + sDisplay;
    return sDisplay;
}

const formatBps = (bytes) => {
    if (!bytes || bytes < 0) bytes = 0;
    if (bytes === 0) return '0 B/s';
    const k = 1024;
    const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s', 'TB/s'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// --- Lifecycle & Data Fetching ---
watch(activeTab, async (newTab) => {
    if (newTab === 'hardware' && vm.value && host.value && !hardware.value) {
        await mainStore.fetchVmHardware(host.value.id, vm.value.name);
    }
    if (newTab === 'hardware' && vm.value && host.value) {
        // Load port attachments for this VM
        await loadPortAttachments();
    }
});

// Track previous VM for proper unsubscribe
let previousVmName = null;
let previousHostId = null;

onMounted(() => {
    const vmName = route.params.vmName;
    if (vmName) {
        // Find the host for this VM
        const vmHost = mainStore.hosts.find(h => h.vms && h.vms.some(v => v.name === vmName));
        if (vmHost) {
            mainStore.subscribeToVmStats(vmHost.id, vmName);
            // Sync VM state from libvirt to ensure we have latest observed state
            mainStore.syncVmFromLibvirt(vmHost.id, vmName);
            previousVmName = vmName;
            previousHostId = vmHost.id;
            // preload attachments
            loadPortAttachments();
        }
    }
});

// Watch for route changes to vmName
watch(
    () => route.params.vmName,
    async (newVmName, oldVmName) => {
        if (previousVmName && previousHostId) {
            mainStore.unsubscribeFromVmStats(previousHostId, previousVmName);
        }
        // Clear any previously loaded hardware so the hardware tab will fetch fresh data
        mainStore.activeVmHardware = null;
        if (newVmName) {
            // Find the host for this VM
            const vmHost = mainStore.hosts.find(h => h.vms && h.vms.some(v => v.name === newVmName));
            if (vmHost) {
                mainStore.subscribeToVmStats(vmHost.id, newVmName);
                // Sync VM state from libvirt to ensure we have latest observed state
                mainStore.syncVmFromLibvirt(vmHost.id, newVmName);
                previousVmName = newVmName;
                previousHostId = vmHost.id;
                // If the hardware tab is active, immediately fetch hardware for the newly-selected VM
                if (activeTab.value === 'hardware') {
                    await mainStore.fetchVmHardware(vmHost.id, newVmName);
                    // Also reload the port attachments for this VM so MACs and device names refresh
                    await loadPortAttachments();
                }
            }
        }
    },
    { immediate: false }
);

onBeforeRouteLeave((to, from, next) => {
    // Unsubscribe from VM stats when leaving this route
    if (previousHostId && previousVmName) {
        mainStore.unsubscribeFromVmStats(previousHostId, previousVmName);
    }
    next();
});

const portAttachments = ref([]);
const loadPortAttachments = async () => {
    if (!host.value || !vm.value) return;
    // Fetch VM-scoped attachments
    portAttachments.value = await mainStore.fetchVmPortAttachments(host.value.id, vm.value.name);
}

</script>

<template>
  <div v-if="vm && host" class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between mb-2">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-bold text-white">{{ vm.name }}</h1>
        <span 
          class="text-sm font-semibold px-3 py-1 rounded-full"
          :class="stateColor(vm)"
        >
          {{ stateText(vm) }}
        </span>
      </div>
      <div class="flex items-center space-x-2">
         <button :disabled="isTaskActive || isReconcileLoading" v-if="getVmDisplayState(vm, host).status === 'STOPPED'" @click="mainStore.startVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-green-600 hover:bg-green-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">Start</button>
         <template v-if="getVmDisplayState(vm, host).status === 'ACTIVE' || getVmDisplayState(vm, host).status === 'RUNNING'">
            <button :disabled="isTaskActive || isReconcileLoading" @click="mainStore.gracefulShutdownVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-yellow-600 hover:bg-yellow-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">Shutdown</button>
            <button :disabled="isTaskActive || isReconcileLoading" @click="mainStore.gracefulRebootVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">Reboot</button>
            <button :disabled="isTaskActive || isReconcileLoading" @click="mainStore.forceOffVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">Force Off</button>
            <button :disabled="isTaskActive || isReconcileLoading" @click="mainStore.forceResetVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-red-800 hover:bg-red-900 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">Force Reset</button>
         </template>
      </div>
    </div>

    <!-- Needs Rebuild Warning -->
    <div v-if="vm.needs_rebuild" class="mb-4 p-4 bg-blue-900/50 border border-blue-700 text-blue-300 rounded-lg">
        <div class="flex items-start">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mr-3 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            <div>
                <h3 class="font-bold">Rebuild Pending</h3>
                <p class="text-sm mt-1">This VM has pending configuration changes. The changes will be applied the next time the VM is started, rebooted, or reset.</p>
            </div>
        </div>
    </div>

    <!-- Drift Detection Warning -->
    <div v-if="vm.sync_status === 'DRIFTED'" class="mb-4 p-4 bg-orange-900/50 border border-orange-700 text-orange-300 rounded-lg">
        <div class="flex items-start">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mr-3 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            <div>
                <h3 class="font-bold">Configuration Drift Detected</h3>
                <p class="text-sm mt-1">The live state of this VM in libvirt does not match the configuration stored in the Virtumancer database.</p>
                <div v-if="driftDetails" class="mt-2 text-xs font-mono bg-black/20 p-2 rounded">
                    <h4 class="font-bold mb-1">Drift Details:</h4>
                    <pre class="whitespace-pre-wrap">{{ JSON.stringify(driftDetails, null, 2) }}</pre>
                </div>
                <div class="mt-4 flex items-center space-x-4">
                    <button 
                        @click="mainStore.syncVmFromLibvirt(host.id, vm.name)"
                        :disabled="isReconcileLoading"
                        class="px-3 py-1.5 text-xs font-semibold text-white bg-green-600 hover:bg-green-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">
                        <span v-if="mainStore.isLoading.vmReconcile === `${vm.name}:sync-from-libvirt`">Syncing...</span>
                        <span v-else>Sync from Live VM</span>
                    </button>
                     <button 
                        @click="mainStore.rebuildVmFromDb(host.id, vm.name)"
                        :disabled="isReconcileLoading"
                        class="px-3 py-1.5 text-xs font-semibold text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors disabled:bg-gray-600 disabled:cursor-not-allowed">
                        <span v-if="mainStore.isLoading.vmReconcile === `${vm.name}:rebuild-from-db`">Rebuilding...</span>
                        <span v-else>Rebuild from Database</span>
                    </button>
                </div>
                 <p class="text-xs mt-2 text-orange-400">
                    <b>Sync from Live VM:</b> Updates the database to match the current live state.<br/>
                    <b>Rebuild from Database:</b> Flags the VM to be reconfigured based on the database on the next power cycle.
                </p>
            </div>
        </div>
    </div>
    
    <!-- UUID Conflict Warning -->
    <div v-if="uuidConflict" class="mb-4 p-4 bg-yellow-900/50 border border-yellow-700 text-yellow-300 rounded-lg">
        <div class="flex items-start">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mr-3 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            <div>
                <h3 class="font-bold">UUID Conflict Detected</h3>
                <p class="text-sm mt-1">This VM's libvirt UUID (`{{ vm.domain_uuid }}`) is already in use by another VM in Virtumancer. To avoid conflicts, Virtumancer has assigned a new internal UUID (`{{ vm.uuid }}`).</p>
                <p class="text-sm mt-2">It is highly recommended to update the VM's actual UUID in libvirt to match the internal one.</p>
                <button class="mt-3 px-3 py-1.5 text-xs font-semibold text-white bg-yellow-600 hover:bg-yellow-700 rounded-md transition-colors">
                    Update UUID in Libvirt
                </button>
                <p class="text-xs mt-1 text-yellow-400">Note: This action will require the VM to be stopped and started.</p>
            </div>
        </div>
    </div>


    <!-- Tab Navigation -->
    <div class="border-b border-gray-700">
      <nav class="-mb-px flex space-x-8" aria-label="Tabs">
        <button @click="activeTab = 'summary'" :class="[activeTab === 'summary' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-gray-400 hover:text-gray-200 hover:border-gray-500', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm']">Summary</button>
        <button @click="activeTab = 'console'" :class="[activeTab === 'console' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-gray-400 hover:text-gray-200 hover:border-gray-500', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm']">Console</button>
        <button @click="activeTab = 'hardware'" :class="[activeTab === 'hardware' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-gray-400 hover:text-gray-200 hover:border-gray-500', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm']">Hardware</button>
        <button @click="activeTab = 'snapshots'" :class="[activeTab === 'snapshots' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-gray-400 hover:text-gray-200 hover:border-gray-500', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm']">Snapshots</button>
      </nav>
    </div>

    <!-- Tab Content -->
    <div class="flex-grow pt-6 overflow-y-auto">
      <!-- Summary Tab -->
      <div v-if="activeTab === 'summary'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
            <h3 class="text-xl font-semibold mb-4 text-white">Core Performance</h3>
            <div class="space-y-6">
                <div>
                    <div class="flex justify-between items-baseline">
                        <label class="text-sm font-medium text-gray-400">CPU Usage</label>
                        <span class="text-sm font-semibold text-white">{{ cpuUsagePercent.toFixed(2) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                        <div class="bg-indigo-500 h-2.5 rounded-full" :style="{ width: cpuUsagePercent + '%' }"></div>
                    </div>
                </div>
                <div>
                     <div class="flex justify-between items-baseline">
                        <label class="text-sm font-medium text-gray-400">Memory Usage</label>
                         <span class="text-sm font-semibold text-white">{{ formatMemory(stats?.memory) }} / {{ formatMemory(stats?.max_mem) }}</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                        <div class="bg-teal-500 h-2.5 rounded-full" :style="{ width: memoryUsagePercent + '%' }"></div>
                    </div>
                </div>
            </div>
        </div>
        <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
            <h3 class="text-xl font-semibold mb-4 text-white">I/O Performance</h3>
            <div class="space-y-4">
                <div>
                    <h4 class="text-sm font-medium text-gray-400 mb-2">Disks</h4>
                    <div v-if="stats?.disk_stats?.length" class="space-y-2">
                        <div v-for="disk in stats.disk_stats" :key="disk.device">
                            <p class="text-xs font-mono text-gray-300">{{ disk.device }}</p>
                            <div class="text-sm flex justify-between">
                                <span>Read: {{ formatBps(diskRates[disk.device]?.read) }}</span>
                                <span>Write: {{ formatBps(diskRates[disk.device]?.write) }}</span>
                            </div>
                        </div>
                    </div>
                    <p v-else class="text-sm text-gray-500">No disk devices found.</p>
                </div>
                 <div>
                    <h4 class="text-sm font-medium text-gray-400 mb-2">Network</h4>
                     <div v-if="stats?.net_stats?.length" class="space-y-2">
                        <div v-for="net in stats.net_stats" :key="net.device">
                            <p class="text-xs font-mono text-gray-300">{{ net.device }}</p>
                            <div class="text-sm flex justify-between">
                                <span>Rx: {{ formatBps(netRates[net.device]?.read) }}</span>
                                <span>Tx: {{ formatBps(netRates[net.device]?.write) }}</span>
                            </div>
                        </div>
                    </div>
                    <p v-else class="text-sm text-gray-500">No network interfaces found.</p>
                </div>
            </div>
        </div>
        <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
          <h3 class="text-xl font-semibold mb-4 text-white">Details</h3>
          <dl class="space-y-4">
            <div> <dt class="text-sm font-medium text-gray-400">Host</dt> <dd class="mt-1 text-lg text-gray-200">{{ host.id }}</dd> </div>
            <div> <dt class="text-sm font-medium text-gray-400">Uptime</dt> <dd class="mt-1 text-lg text-gray-200">{{ formatUptime(vm.uptime) }}</dd> </div>
             <div> <dt class="text-sm font-medium text-gray-400">vCPUs</dt> <dd class="mt-1 text-lg text-gray-200">{{ vm.vcpu }}</dd> </div>
            <div> <dt class="text-sm font-medium text-gray-400">Memory</dt> <dd class="mt-1 text-lg text-gray-200">{{ formatMemory(vm.max_mem) }}</dd> </div>
            <div> <dt class="text-sm font-medium text-gray-400">Internal UUID</dt> <dd class="mt-1 text-xs font-mono text-gray-200">{{ vm.uuid }}</dd> </div>
          </dl>
        </div>
      </div>

      <!-- Console Tab -->
      <div v-if="activeTab === 'console'" class="h-full w-full">
         <div v-if="vm.state !== 'ACTIVE'" class="flex items-center justify-center h-full text-gray-500 bg-gray-900 rounded-lg">
            <p>Console is only available when the VM is running.</p>
         </div>
         <div v-else class="h-full w-full bg-black rounded-lg overflow-hidden">
            <VncConsole v-if="vm.graphics.vnc" :host-id="host.id" :vm-name="vm.name" />
            <SpiceConsole v-else-if="vm.graphics.spice" :host-id="host.id" :vm-name="vm.name" />
            <div v-else class="flex items-center justify-center h-full text-gray-500">
                <p>No supported console type (VNC or SPICE) is configured for this VM.</p>
            </div>
         </div>
      </div>
      
      <!-- Hardware Tab -->
       <div v-if="activeTab === 'hardware'" class="space-y-8">
            <div v-if="mainStore.isLoading.vmHardware" class="flex items-center justify-center h-48 text-gray-400">
                <svg class="animate-spin mr-3 h-8 w-8 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>Loading Hardware...</span>
            </div>
            <div v-else-if="hardware">
                <!-- Storage Devices -->
                <div class="bg-gray-900 rounded-lg shadow-lg">
                    <h3 class="text-xl font-semibold text-white p-4">Storage</h3>
                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-gray-700">
                            <thead class="bg-gray-800">
                                <tr>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Device</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Bus</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Source</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Format</th>
                                </tr>
                            </thead>
                            <tbody class="bg-gray-900 divide-y divide-gray-800">
                                <tr v-for="disk in hardware.disks" :key="disk.target.dev">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">{{ disk.target?.dev || 'N/A' }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ disk.target?.bus || 'N/A' }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300 font-mono break-all">{{ disk.source?.file || disk.name || 'N/A' }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ disk.driver?.type || 'N/A' }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- Video Devices -->
                <div class="bg-gray-900 rounded-lg shadow-lg">
                    <h3 class="text-xl font-semibold text-white p-4">Video / GPU</h3>
                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-gray-700">
                            <thead class="bg-gray-800">
                                <tr>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Model</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">VRAM</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Heads</th>
                                </tr>
                            </thead>
                            <tbody class="bg-gray-900 divide-y divide-gray-800">
                                <tr v-for="(video, index) in hardware.videos" :key="index">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">{{ video.model?.type || 'N/A' }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ video.model?.vram || 'N/A' }} MB</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ video.model?.heads || 'N/A' }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- Boot Order -->
                <div class="bg-gray-900 rounded-lg shadow-lg">
                    <h3 class="text-xl font-semibold text-white p-4">Boot Order</h3>
                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-gray-700">
                            <thead class="bg-gray-800">
                                <tr>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Order</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Device</th>
                                </tr>
                            </thead>
                            <tbody class="bg-gray-900 divide-y divide-gray-800">
                                <tr v-for="boot in hardware.boot" :key="boot.order">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">{{ boot?.order || 'N/A' }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ boot?.dev || 'N/A' }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- Network / Ports -->
                <div class="bg-gray-900 rounded-lg shadow-lg">
                    <h3 class="text-xl font-semibold text-white p-4">Network / Ports</h3>
                    <div class="p-4 text-sm text-gray-300">
                        <h4 class="font-semibold mb-2">Attached Ports</h4>
                        <div v-if="!portAttachments || portAttachments.length === 0" class="text-gray-400 mb-4">No port attachments found for this VM.</div>
                        <div v-else class="overflow-x-auto mb-4">
                            <table class="min-w-full divide-y divide-gray-700">
                                <thead class="bg-gray-800">
                                    <tr>
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Device</th>
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">MAC</th>
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Model</th>
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Network</th>
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Host</th>
                                    </tr>
                                </thead>
                                <tbody class="bg-gray-900 divide-y divide-gray-800">
                                    <tr v-for="att in portAttachments" :key="att.id">
                                        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">{{ att.device_name || att.DeviceName || '-' }}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300 font-mono">{{ att.mac_address || att.MACAddress || (att.port && att.port.MACAddress) || '-' }}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ att.model_name || att.ModelName || (att.port && att.port.model_name) || '-' }}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ att.network?.bridge_name || att.Network?.BridgeName || (att.network && att.network.bridge_name) || '-' }}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{{ att.host_id || '-' }}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>

                        <!-- Unattached host ports removed from VM view; host-scoped list is available in Host Dashboard -->
                    </div>
                </div>
            </div>
            <!-- Video Devices -->
            <div class="bg-gray-900 rounded-lg shadow-lg">
                <h3 class="text-xl font-semibold text-white p-4">Video / GPU</h3>
                <div class="p-4 text-sm text-gray-300">
                    <p class="mb-2">Video models attached to this VM:</p>
                    <ul class="list-disc pl-6">
                        <li v-for="va in mainStore.fetchVmVideoAttachments(host.id, vm.name)" :key="va.id">
                            Model: {{ va.video_model?.model_name || va.VideoModel?.ModelName }} — Monitor: {{ va.monitor_index }} — Primary: {{ va.primary }}
                        </li>
                    </ul>
                    <p class="mt-4 mb-2">Physical video devices on host:</p>
                    <ul class="list-disc pl-6">
                        <li v-for="vd in mainStore.fetchHostVideoDevices(host.id)" :key="vd.id">
                            {{ vd.model_name || vd.ModelName }} ({{ vd.vendor }})
                        </li>
                    </ul>
                </div>
            </div>
             <div v-if="!mainStore.isLoading.vmHardware && !hardware" class="flex items-center justify-center h-48 text-gray-500 bg-gray-900 rounded-lg">
                <p>Could not load hardware information.</p>
            </div>
       </div>
       
       <!-- Snapshots Tab Placeholder -->
        <div v-if="activeTab === 'snapshots'" class="flex items-center justify-center h-full text-gray-500 bg-gray-900 rounded-lg">
            <p>Snapshot management will be implemented here.</p>
       </div>

    </div>
  </div>

  <div v-else class="flex items-center justify-center h-full text-gray-500">
    <p>Select a VM from the sidebar to view details, or the VM is still loading.</p>
  </div>
</template>



