<script setup>
import { useMainStore } from '@/stores/mainStore';
import { computed, ref, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';

const mainStore = useMainStore();
const route = useRoute();

// Network data across all hosts
const allHostPorts = ref([]);
const allPortAttachments = ref([]);

// Load network data for all hosts
const loadAllNetworkData = async () => {
  allHostPorts.value = [];
  allPortAttachments.value = [];

  // Load data for all hosts
  const hostPromises = mainStore.hosts.map(async (host) => {
    try {
      // Load unattached ports for this host
      const ports = await mainStore.fetchHostPorts(host.id);
      allHostPorts.value.push(...(ports || []).map(p => ({ ...p, hostId: host.id, hostName: host.name })));

      // Load port attachments for all VMs on this host
      if (host.vms && host.vms.length) {
        const attachmentPromises = host.vms.map(vm =>
          mainStore.fetchVmPortAttachments(host.id, vm.name)
            .then(list => (list || []).map(a => ({ ...a, vmName: vm.name, hostId: host.id, hostName: host.name })))
            .catch(() => [])
        );
        const settled = await Promise.allSettled(attachmentPromises);
        const attachments = settled.flatMap(s => (s.status === 'fulfilled' ? s.value : []));
        allPortAttachments.value.push(...attachments);
      }
    } catch (e) {
      console.error(`[NetworkView] Failed to load network data for host ${host.id}:`, e);
    }
  });

  await Promise.allSettled(hostPromises);
};

onMounted(async () => {
  // Ensure hosts are loaded
  if (mainStore.hosts.length === 0) {
    await mainStore.fetchHosts();
  }
  await loadAllNetworkData();
});

// Watch for host changes and reload network data
watch(() => mainStore.hosts, async (newHosts) => {
  if (newHosts && newHosts.length > 0) {
    await loadAllNetworkData();
  }
}, { deep: true });

// Watch for VM changes on any host and reload attachments
watch(() => mainStore.hosts.map(h => h.vms), async () => {
  await loadAllNetworkData();
}, { deep: true });
</script>

<template>
  <div class="network-view p-6">
    <div class="mb-6">
      <h1 class="text-3xl font-bold text-white mb-2">Network Management</h1>
      <p class="text-gray-400">Manage network ports and attachments across all hosts</p>
    </div>

    <!-- Unattached Ports (Global View) -->
    <div class="mb-6 bg-gray-900 rounded-lg p-4">
      <h2 class="text-xl font-semibold text-white mb-3">Unattached Ports (Port Pool)</h2>
      <div v-if="allHostPorts.length === 0" class="text-gray-400">No unattached ports found across all hosts.</div>
      <ul v-else class="space-y-2">
        <li v-for="p in allHostPorts" :key="`${p.hostId}-${p.id}`" class="bg-gray-800 p-3 rounded flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-300">
              <span class="font-medium text-blue-400">{{ p.hostName }}</span> •
              MAC: <span class="font-mono">{{ p.mac_address || p.MACAddress || p.MAC }}</span>
            </div>
            <div class="text-xs text-gray-500">Device: {{ p.device_name || p.DeviceName || '-' }} • Model: {{ p.model_name || p.ModelName || '-' }}</div>
          </div>
          <div class="text-sm text-gray-400">ID: {{ p.id }}</div>
        </li>
      </ul>
    </div>

    <!-- Port Attachments (Global View) -->
    <div class="mb-6 bg-gray-900 rounded-lg p-4">
      <h2 class="text-xl font-semibold text-white mb-3">Port Attachments Across All Hosts</h2>
      <div v-if="allPortAttachments.length === 0" class="text-gray-400">No port attachments found across all hosts.</div>
      <ul v-else class="space-y-2">
        <li v-for="att in allPortAttachments" :key="`${att.hostId}-${att.id}`" class="bg-gray-800 p-3 rounded flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-300">
              <span class="font-medium text-blue-400">{{ att.hostName }}</span> •
              Device: <span class="font-medium text-white">{{ att.device_name || att.DeviceName || '-' }}</span>
            </div>
            <div class="text-xs text-gray-500">
              VM: {{ att.vmName || att.vm_name || '-' }} •
              MAC: <span class="font-mono">{{ att.mac_address || att.MACAddress || (att.port && att.port.MACAddress) || '-' }}</span>
            </div>
          </div>
          <div class="text-sm text-gray-400">Host: {{ att.hostName }}</div>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.network-view {
  min-height: 100vh;
  background-color: #1f2937;
}
</style>