<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed, ref, watch, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import VncConsole from '@/components/consoles/VncConsole.vue';
import SpiceConsole from '@/components/consoles/SpiceConsole.vue';

const mainStore = useMainStore();
const route = useRoute();
const activeTab = ref('summary');

let pollInterval = null;

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

const stats = computed(() => mainStore.activeVmStats);

// --- CPU Usage Calculation ---
const lastCpuTime = ref(0);
const lastCpuTimeTimestamp = ref(0);
const cpuUsagePercent = ref(0);

watch(stats, (newStats) => {
    if (!newStats || newStats.state !== 1) {
        cpuUsagePercent.value = 0;
        return;
    }
    if (lastCpuTime.value > 0 && lastCpuTimeTimestamp.value > 0) {
        const timeDelta = Date.now() - lastCpuTimeTimestamp.value; // in ms
        const cpuTimeDelta = newStats.cpu_time - lastCpuTime.value; // in ns

        if (timeDelta > 0) {
            // Convert timeDelta to nanoseconds
            const timeDeltaNs = timeDelta * 1_000_000;
            // Calculate usage percentage
            const usage = (cpuTimeDelta / (timeDeltaNs * newStats.vcpu)) * 100;
            cpuUsagePercent.value = Math.min(Math.max(usage, 0), 100);
        }
    }
    lastCpuTime.value = newStats.cpu_time;
    lastCpuTimeTimestamp.value = Date.now();
});


const memoryUsagePercent = computed(() => {
    if (!stats.value || !stats.value.max_mem || stats.value.state !== 1) return 0;
    return (stats.value.memory / stats.value.max_mem) * 100;
});


// --- Helper functions ---
const stateText = (state) => {
    const states = { 0: 'No State', 1: 'Running', 2: 'Blocked', 3: 'Paused', 4: 'Shutdown', 5: 'Shutoff', 6: 'Crashed', 7: 'PMSuspended' };
    return states[state] || 'Unknown';
};

const stateColor = (state) => {
  const colors = { 1: 'text-green-400 bg-green-900/50', 3: 'text-yellow-400 bg-yellow-900/50', 5: 'text-red-400 bg-red-900/50' };
  return colors[state] || 'text-gray-400 bg-gray-700';
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

// --- Lifecycle hooks for polling ---

// Watch for the active VM to change. This handles both initial load and navigation between VMs.
watch(vm, (newVm) => {
    // Clear any existing interval
    clearInterval(pollInterval);
    pollInterval = null;

    // Reset local component state
    activeTab.value = 'summary';
    lastCpuTime.value = 0;
    lastCpuTimeTimestamp.value = 0;
    cpuUsagePercent.value = 0;
    mainStore.activeVmStats = null;

    // If we have a new valid VM, start polling for its stats
    if (newVm && host.value) {
        mainStore.fetchVmStats(host.value.id, newVm.name);
        pollInterval = setInterval(() => {
            mainStore.fetchVmStats(host.value.id, newVm.name);
        }, 2000); // Poll every 2 seconds
    }
}, { immediate: true }); // immediate: true ensures this runs on component mount

onUnmounted(() => {
    clearInterval(pollInterval);
    mainStore.activeVmStats = null; // Clear stats when leaving view entirely
});

</script>

<template>
  <div v-if="vm && host" class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-bold text-white">{{ vm.name }}</h1>
        <span 
          class="text-sm font-semibold px-3 py-1 rounded-full"
          :class="stateColor(vm.state)"
        >
          {{ stateText(vm.state) }}
        </span>
      </div>
      <div class="flex items-center space-x-2">
         <button v-if="vm.state === 5" @click="mainStore.startVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-green-600 hover:bg-green-700 rounded-md transition-colors">Start</button>
         <template v-if="vm.state === 1">
            <button @click="mainStore.gracefulShutdownVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-yellow-600 hover:bg-yellow-700 rounded-md transition-colors">Shutdown</button>
            <button @click="mainStore.gracefulRebootVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors">Reboot</button>
            <button @click="mainStore.forceOffVm(host.id, vm.name)" class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors">Force Off</button>
         </template>
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
        <!-- Performance Section -->
        <div class="bg-gray-900 p-6 rounded-lg shadow-lg lg:col-span-2">
            <h3 class="text-xl font-semibold mb-4 text-white">Performance</h3>
            <div class="space-y-6">
                <!-- CPU Usage -->
                <div>
                    <div class="flex justify-between items-baseline">
                        <label class="text-sm font-medium text-gray-400">CPU Usage</label>
                        <span class="text-sm font-semibold text-white">{{ cpuUsagePercent.toFixed(2) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2.5 mt-2">
                        <div class="bg-indigo-500 h-2.5 rounded-full" :style="{ width: cpuUsagePercent + '%' }"></div>
                    </div>
                </div>
                <!-- Memory Usage -->
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
        <!-- Details Section -->
        <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
          <h3 class="text-xl font-semibold mb-4 text-white">Details</h3>
          <dl class="space-y-4">
            <div>
              <dt class="text-sm font-medium text-gray-400">Host</dt>
              <dd class="mt-1 text-lg text-gray-200">{{ host.id }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-400">Uptime</dt>
              <dd class="mt-1 text-lg text-gray-200">{{ formatUptime(vm.uptime) }}</dd>
            </div>
             <div>
              <dt class="text-sm font-medium text-gray-400">vCPUs</dt>
              <dd class="mt-1 text-lg text-gray-200">{{ vm.vcpu }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-400">Memory</dt>
              <dd class="mt-1 text-lg text-gray-200">{{ formatMemory(vm.max_mem) }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Console Tab -->
      <div v-if="activeTab === 'console'" class="h-full w-full">
         <div v-if="vm.state !== 1" class="flex items-center justify-center h-full text-gray-500 bg-gray-900 rounded-lg">
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
      
      <!-- Hardware Tab Placeholder -->
       <div v-if="activeTab === 'hardware'" class="flex items-center justify-center h-full text-gray-500 bg-gray-900 rounded-lg">
            <p>Hardware management will be implemented here.</p>
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


