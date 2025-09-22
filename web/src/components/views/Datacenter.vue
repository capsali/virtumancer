<script setup>
import { useMainStore } from '@/stores/mainStore';
import { useRouter } from 'vue-router';
import { computed } from 'vue';
import { useVmStateDisplay } from '@/composables/useVmStateDisplay';

const mainStore = useMainStore();
const router = useRouter();
const { isVmRunning } = useVmStateDisplay();

function selectHost(hostId) {
  mainStore.selectHost(hostId);
  router.push({ name: 'host-dashboard', params: { hostId } });
}

const totalVms = (host) => host.vms?.length || 0;
const runningVms = (host) => host.vms?.filter(vm => isVmRunning(vm, host)).length || 0;

const formatBytes = (bytes, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';
    if (!bytes) return 'N/A';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

const calculateCpuUsage = (host) => {
  if (!host || !host.vms || !host.info) return 0;
  const hostCores = host.info.cpu;
  if (!hostCores) return 0;
  const runningVmCores = host.vms.reduce((acc, vm) => isVmRunning(vm, host) ? acc + vm.vcpu_count : acc, 0);
  return Math.min(100, (runningVmCores / hostCores) * 100);
};

const calculateMemoryUsage = (host) => {
    if (!host || !host.info || !host.info.memory) return { percent: 0, used: 0, total: 0 };
    const totalMem = host.info.memory;
    const usedMem = host.info.memory_used;
    return {
        percent: totalMem > 0 ? (usedMem / totalMem) * 100 : 0,
        used: usedMem,
        total: totalMem
    };
}

</script>

<template>
  <div>
    <h1 class="text-3xl font-bold text-white mb-6">Datacenter Overview</h1>
    <div v-if="mainStore.isLoading.hosts && mainStore.hosts.length === 0" class="flex items-center justify-center h-64 text-gray-400">
        <svg class="animate-spin mr-3 h-8 w-8 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>Loading Hosts...</span>
    </div>
     <div v-else-if="mainStore.hosts.length === 0" class="flex items-center justify-center h-64 text-gray-500 bg-gray-900 rounded-lg">
      <p>No hosts have been added. Click "Add Host" in the sidebar to get started.</p>
    </div>
    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
      <div 
        v-for="host in mainStore.hosts" 
        :key="host.id" 
        @click="selectHost(host.id)"
        class="bg-gray-800 p-6 rounded-lg shadow-lg hover:shadow-xl hover:bg-gray-700/50 transition-all duration-200 cursor-pointer flex flex-col justify-between"
      >
        <div>
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xl font-bold text-white truncate">{{ host.id }}</h2>
              <span class="px-3 py-1 text-xs font-semibold text-green-300 bg-green-900/50 rounded-full">Connected</span>
            </div>
            <p class="text-sm text-gray-400 font-mono break-all mb-6">{{ host.uri }}</p>
        </div>

        <div class="space-y-4">
            <!-- CPU Usage -->
            <div>
                <div class="flex justify-between items-end mb-1">
                    <span class="text-sm font-medium text-gray-300">CPU</span>
                    <span class="text-xs font-mono text-gray-400">{{ calculateCpuUsage(host).toFixed(0) }}% of {{ host.info?.cpu }} Cores</span>
                </div>
                <div class="w-full bg-gray-700 rounded-full h-2.5">
                    <div class="bg-indigo-500 h-2.5 rounded-full" :style="{ width: calculateCpuUsage(host) + '%' }"></div>
                </div>
            </div>

            <!-- Memory Usage -->
            <div>
                <div class="flex justify-between items-end mb-1">
                    <span class="text-sm font-medium text-gray-300">Memory</span>
                     <span class="text-xs font-mono text-gray-400">{{ formatBytes(calculateMemoryUsage(host).used) }} / {{ formatBytes(calculateMemoryUsage(host).total) }}</span>
                </div>
                <div class="w-full bg-gray-700 rounded-full h-2.5">
                    <div class="bg-teal-500 h-2.5 rounded-full" :style="{ width: calculateMemoryUsage(host).percent + '%' }"></div>
                </div>
            </div>

            <!-- VM Count -->
             <div class="border-t border-gray-700 pt-4 mt-4">
                <div class="grid grid-cols-2 gap-4 text-center">
                    <div>
                        <p class="text-2xl font-bold text-white">{{ runningVms(host) }}</p>
                        <p class="text-xs text-gray-400 uppercase tracking-wider">Running VMs</p>
                    </div>
                    <div>
                        <p class="text-2xl font-bold text-white">{{ totalVms(host) }}</p>
                        <p class="text-xs text-gray-400 uppercase tracking-wider">Total VMs</p>
                    </div>
                </div>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

