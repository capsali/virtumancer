<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed, onMounted } from 'vue';
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';

const mainStore = useMainStore();
const route = useRoute();
const router = useRouter();

const selectedHost = computed(() => {
  const hostId = route.params.hostId;
  if (!hostId) return null;
  return mainStore.hosts.find(h => h.id === hostId);
});

const hostStats = computed(() => {
    if (!selectedHost.value) return null;
    return mainStore.hostStats[selectedHost.value.id];
});

const vms = computed(() => {
    return selectedHost.value?.vms || [];
});

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
    return selectedHost.value.vms.reduce((total, vm) => total + (vm.state === 'ACTIVE' ? vm.vcpu_count : 0), 0);
});

const vCpuAllocationPercent = computed(() => {
    if (!totalCpu.value) return 0;
    return (vCpuAllocation.value / totalCpu.value) * 100;
});

const cpuUtilization = computed(() => {
    if (!hostStats.value) return 0;
    return (hostStats.value.cpu_utilization * 100).toFixed(1);
});

const selectVm = (vmName) => {
    router.push({ name: 'vm-view', params: { vmName } });
}



onMounted(() => {
    if (route.params.hostId) {
        mainStore.subscribeHostStats(route.params.hostId);
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
    const states = {
        'INITIALIZED': 'Initialized',
        'ACTIVE': 'Running', 
        'PAUSED': 'Paused', 
        'SUSPENDED': 'Suspended',
        'STOPPED': 'Stopped', 
        'ERROR': 'Error'
    };
    return states[vm.state] || 'Unknown';
};

const stateColor = (vm) => {
  if (!vm) return 'text-gray-400 bg-gray-700';
  if (vm.task_state) {
      return 'text-orange-300 bg-orange-900/50 animate-pulse';
  }
  const colors = {
    'ACTIVE': 'text-green-400 bg-green-900/50',
    'PAUSED': 'text-yellow-400 bg-yellow-900/50',
    'SUSPENDED': 'text-blue-400 bg-blue-900/50',
    'STOPPED': 'text-red-400 bg-red-900/50',
    'ERROR': 'text-red-400 bg-red-900/50 font-bold',
  };
  return colors[vm.state] || 'text-gray-400 bg-gray-700';
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

</script>

<template>
  <div v-if="selectedHost">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-3xl font-bold text-white">Host: {{ selectedHost.id }}</h1>
      <p class="text-gray-400 font-mono mt-1">{{ selectedHost.uri }}</p>
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
                    'bg-green-500': vm.state === 'ACTIVE' && !vm.task_state, 
                    'bg-red-500': (vm.state === 'STOPPED' || vm.state === 'ERROR') && !vm.task_state,
                    'bg-yellow-500': vm.state === 'PAUSED' && !vm.task_state,
                    'bg-blue-500': vm.state === 'SUSPENDED' && !vm.task_state,
                    'bg-orange-500 animate-pulse': vm.task_state,
                    'bg-gray-500': !['ACTIVE', 'STOPPED', 'ERROR', 'PAUSED', 'SUSPENDED'].includes(vm.state) && !vm.task_state
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
</template>


