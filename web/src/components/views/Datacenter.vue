<script setup>
import { useMainStore } from '@/stores/mainStore';
import { useRouter } from 'vue-router';

const mainStore = useMainStore();
const router = useRouter();

function selectHost(hostId) {
  mainStore.selectHost(hostId);
  router.push({ name: 'host-dashboard', params: { hostId } });
}

const totalVms = (host) => host.vms?.length || 0;
const runningVms = (host) => host.vms?.filter(vm => vm.state === 1).length || 0;

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
        class="bg-gray-800 p-6 rounded-lg shadow-lg hover:shadow-xl hover:bg-gray-700/50 transition-all duration-200 cursor-pointer"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-white">{{ host.id }}</h2>
          <span class="px-3 py-1 text-xs font-semibold text-green-300 bg-green-900/50 rounded-full">Connected</span>
        </div>
        <p class="text-sm text-gray-400 font-mono break-all">{{ host.uri }}</p>
        <div class="mt-6 border-t border-gray-700 pt-4">
            <div class="grid grid-cols-2 gap-4 text-center">
                <div>
                    <p class="text-2xl font-bold text-white">{{ runningVms(host) }}</p>
                    <p class="text-sm text-gray-400">Running VMs</p>
                </div>
                <div>
                    <p class="text-2xl font-bold text-white">{{ totalVms(host) }}</p>
                    <p class="text-sm text-gray-400">Total VMs</p>
                </div>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

