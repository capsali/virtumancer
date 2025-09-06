<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const mainStore = useMainStore();
const route = useRoute();

const selectedHost = computed(() => {
  const hostId = route.params.hostId || mainStore.selectedHostId;
  if (!hostId) return null;
  return mainStore.hosts.find(h => h.id === hostId);
});
</script>

<template>
  <div v-if="selectedHost">
    <h1 class="text-3xl font-bold mb-4">Host: {{ selectedHost.id }}</h1>
    <p class="text-gray-400 mb-6">URI: {{ selectedHost.uri }}</p>
    
    <!-- Placeholder for host details, stats, and VM list -->
    <div class="bg-gray-900 p-6 rounded-lg shadow-lg">
      <h2 class="text-xl font-semibold mb-4">Virtual Machines</h2>
      <div v-if="mainStore.isLoading.vms" class="text-center text-gray-400">Loading VMs...</div>
      <ul v-else-if="selectedHost.vms && selectedHost.vms.length">
        <li v-for="vm in selectedHost.vms" :key="vm.name" class="border-b border-gray-700 last:border-b-0 py-2">
          {{ vm.name }} - {{ vm.state }}
        </li>
      </ul>
      <p v-else class="text-gray-500">No VMs found on this host.</p>
    </div>
  </div>
  <div v-else class="flex items-center justify-center h-full text-gray-500">
    <p>Select a host from the sidebar to view details.</p>
  </div>
</template>


