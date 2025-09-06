<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const mainStore = useMainStore();
const route = useRoute();
const router = useRouter();

const selectedHost = computed(() => {
  const hostId = route.params.hostId || mainStore.selectedHostId;
  if (!hostId) {
      // Find the first host if none is selected
      return mainStore.hosts.length > 0 ? mainStore.hosts[0] : null;
  }
  return mainStore.hosts.find(h => h.id === hostId);
});

const vms = computed(() => {
    return selectedHost.value?.vms || [];
});

const selectVm = (vmName) => {
    router.push({ name: 'vm-view', params: { vmName } });
}

// Helper functions for display
const stateText = (state) => {
    const states = {
        0: 'No State', 1: 'Running', 2: 'Blocked', 3: 'Paused',
        4: 'Shutdown', 5: 'Shutoff', 6: 'Crashed', 7: 'PMSuspended',
    };
    return states[state] || 'Unknown';
};

const stateColor = (state) => {
  const colors = {
    1: 'text-green-400', 
    3: 'text-yellow-400',
    5: 'text-red-400',
  };
  return colors[state] || 'text-gray-400';
};

const formatMemory = (kb) => {
    if (kb === 0) return '0 MB';
    const mb = kb / 1024;
    if (mb < 1024) return `${mb.toFixed(0)} MB`;
    const gb = mb / 1024;
    return `${gb.toFixed(2)} GB`;
}
</script>

<template>
  <div v-if="selectedHost">
    <div class="mb-6">
      <h1 class="text-3xl font-bold text-white">Host: {{ selectedHost.id }}</h1>
      <p class="text-gray-400 font-mono mt-1">{{ selectedHost.uri }}</p>
    </div>
    
    <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
      <h2 class="text-xl font-semibold mb-4 text-white border-b border-gray-700 pb-3">Virtual Machines</h2>
      
      <div v-if="mainStore.isLoading.vms && vms.length === 0" class="flex items-center justify-center h-48 text-gray-400">
        <svg class="animate-spin mr-3 h-8 w-8 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>Loading VMs...</span>
      </div>

      <div v-else-if="vms.length === 0" class="flex items-center justify-center h-48 text-gray-500">
        <p>No Virtual Machines found on this host.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div v-for="vm in vms" :key="vm.name" 
            @click="selectVm(vm.name)"
            class="bg-gray-800 rounded-lg flex flex-col justify-between shadow-md hover:shadow-xl hover:bg-gray-700/50 transition-all duration-200 cursor-pointer">
          
          <div class="p-4 flex-grow">
            <div class="flex items-center justify-between">
              <h3 class="font-bold text-lg truncate text-white" :title="vm.name">{{ vm.name }}</h3>
              <div class="flex items-center space-x-2">
                  <span class="text-xs font-semibold px-2 py-1 rounded-full" :class="stateColor(vm.state)">●</span>
                  <span class="text-sm font-medium" :class="stateColor(vm.state)">{{ stateText(vm.state) }}</span>
              </div>
            </div>
            <div class="mt-4 space-y-2 text-sm text-gray-300">
                <div class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M12 6V4m0 16v-2M8 12a4 4 0 118 0 4 4 0 01-8 0z" /></svg>
                    <span>{{ vm.vcpu }} vCPU</span>
                </div>
                <div class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v4m0 0h-4m4 0l-5-5" /></svg>
                    <span>{{ formatMemory(vm.max_mem) }} Memory</span>
                </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="flex items-center justify-center h-full text-gray-500">
    <p>Select a host from the sidebar, or add one to get started.</p>
  </div>
</template>


