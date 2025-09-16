<script setup>
import { useUiStore } from '@/stores/uiStore';
import { useMainStore } from '@/stores/mainStore';
import { onMounted, ref } from 'vue';
import { watch } from 'vue';
import ConfirmModal from '@/components/modals/ConfirmModal.vue';
import { useRouter } from 'vue-router';

const uiStore = useUiStore();
const mainStore = useMainStore();
const router = useRouter();

const expandedHosts = ref({});
const showConfirm = ref(false);
const confirmPayload = ref({});
const confirmLoading = ref(false);

const openConfirmAll = (hostId) => {
  confirmPayload.value = { type: 'all', hostId };
  showConfirm.value = true;
};

const openConfirmOne = (hostId, vmName) => {
  confirmPayload.value = { type: 'one', hostId, vmName };
  showConfirm.value = true;
};

const handleConfirm = async () => {
    try {
    console.log('[Sidebar] handleConfirm invoked, payload=', confirmPayload.value);
    if (!confirmPayload.value) return;
    confirmLoading.value = true;
    if (confirmPayload.value.type === 'all') {
      console.log('[Sidebar] calling importAllVMs for host', confirmPayload.value.hostId);
      mainStore.addToast(`Starting import of all discovered VMs on ${confirmPayload.value.hostId}`, 'success');
      await mainStore.importAllVMs(confirmPayload.value.hostId);
    } else if (confirmPayload.value.type === 'one') {
      console.log('[Sidebar] calling importVm for', confirmPayload.value.vmName, 'on host', confirmPayload.value.hostId);
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

onMounted(() => {
  mainStore.initializeRealtime();
  mainStore.fetchHosts().then(() => {
    // Automatically expand all hosts on load
    mainStore.hosts.forEach(host => {
      expandedHosts.value[host.id] = true;
      // prime discovered cache for each host
      mainStore.refreshDiscoveredVMs(host.id).catch(() => {});
    });
  });
});

// When hosts list changes, ensure discovered cache updates for new hosts
watch(() => mainStore.hosts, (nv, ov) => {
  nv.forEach(host => {
    if (!mainStore.discoveredByHost || !mainStore.discoveredByHost[host.id]) {
      mainStore.refreshDiscoveredVMs(host.id).catch(() => {});
    }
  });
}, { deep: true });

function selectDatacenter() {
    mainStore.selectHost(null);
    router.push({ name: 'datacenter' });
}

function selectHost(hostId) {
  mainStore.selectHost(hostId);
  router.push({ name: 'host-dashboard', params: { hostId } });
}

function selectVm(vm) {
  for (const host of mainStore.hosts) {
      if (host.vms && host.vms.find(hvm => hvm.name === vm.name)) {
          mainStore.selectHost(host.id);
          break;
      }
  }
  router.push({ name: 'vm-view', params: { vmName: vm.name } });
}

function toggleHost(hostId) {
    expandedHosts.value[hostId] = !expandedHosts.value[hostId];
}

const runningVmsCount = (host) => {
    return host.vms ? host.vms.filter(vm => vm.state === 'ACTIVE').length : 0;
}

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
    return host.connected ? 'Connected' : 'Disconnected';
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
</script>

<template>
  <aside 
    class="flex flex-col bg-gray-900 text-gray-300 transition-all duration-300 ease-in-out"
    :class="uiStore.isSidebarOpen ? 'w-72' : 'w-20'"
  >
    <div class="flex items-center h-16 px-6 border-b border-gray-800 flex-shrink-0">
      <h1 class="text-xl font-bold text-white tracking-wider" v-show="uiStore.isSidebarOpen">
        Virtu<span class="text-indigo-400">Mancer</span>
      </h1>
       <svg v-show="!uiStore.isSidebarOpen" class="h-8 w-8 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
      </svg>
    </div>

    <div class="flex-shrink-0 px-4 py-4">
      <button @click="uiStore.openAddHostModal" class="w-full flex items-center justify-center p-2 rounded-md bg-indigo-600 text-white hover:bg-indigo-700 transition-colors">
        <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
        <span class="ml-2" v-show="uiStore.isSidebarOpen">Add Host</span>
      </button>
    </div>
    
    <nav class="flex-1 overflow-y-auto">
      <ul class="px-4">
        <!-- Datacenter Root -->
        <li class="mb-2">
            <div @click="selectDatacenter" class="flex items-center p-2 rounded-md cursor-pointer hover:bg-gray-700" :class="{ 'bg-gray-700 text-white': !mainStore.selectedHostId }">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" /></svg>
                <span class="ml-3 font-semibold" v-show="uiStore.isSidebarOpen">Datacenter</span>
            </div>
        </li>

        <!-- Hosts -->
        <li v-for="host in mainStore.hosts" :key="host.id" class="mb-1">
          <div 
            class="flex items-center p-2 rounded-md cursor-pointer hover:bg-gray-700"
            :class="{ 'bg-gray-700 text-white': mainStore.selectedHostId === host.id }"
            @click="selectHost(host.id)"
          >
            <button @click.stop="toggleHost(host.id)" v-show="uiStore.isSidebarOpen" class="mr-2 focus:outline-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 transition-transform" :class="{'rotate-90': expandedHosts[host.id]}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </button>
            <svg class="h-6 w-6 flex-shrink-0" :class="{'text-indigo-400': mainStore.selectedHostId === host.id}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/></svg>
            <span class="ml-3 font-semibold truncate" v-show="uiStore.isSidebarOpen">{{ host.id }}</span>
      <div v-show="uiStore.isSidebarOpen" class="ml-2 mr-auto flex items-center space-x-2">
      <span v-if="mainStore.hostConnecting && mainStore.hostConnecting[host.id]" class="text-xs px-2 py-0.5 rounded-full bg-yellow-500 text-white">connecting...</span>
      <span v-else class="text-xs px-2 py-0.5 rounded-full" :class="hostStateColor(host)">{{ hostStateText(host) }}</span>
        </div>
             <span v-if="uiStore.isSidebarOpen && host.vms" class="ml-auto text-xs font-mono bg-gray-800 px-2 py-0.5 rounded-full">
              {{ runningVmsCount(host) }}/{{ host.vms.length }}
            </span>
          </div>

          <!-- VMs grouped: Managed then Discovered -->
          <div v-if="uiStore.isSidebarOpen && expandedHosts[host.id]" class="mt-1 ml-6 border-l-2 border-gray-700 pl-4">
            <div v-if="host.vms && host.vms.length" class="mb-2">
              <div class="text-xs text-gray-400 uppercase font-semibold mb-1">Managed</div>
              <ul class="space-y-1">
                <li v-for="vm in host.vms" :key="vm.name">
                  <div @click="selectVm(vm)" class="flex items-center p-1.5 text-sm rounded-md cursor-pointer hover:bg-gray-700" :class="{'bg-gray-700/50': $route.params.vmName === vm.name}">
                    <span class="h-2 w-2 rounded-full mr-2 flex-shrink-0" :class="{
                      'bg-green-500': vm.state === 'ACTIVE' && !vm.task_state, 
                      'bg-red-500': (vm.state === 'STOPPED' || vm.state === 'ERROR') && !vm.task_state,
                      'bg-yellow-500': vm.state === 'PAUSED' && !vm.task_state,
                      'bg-blue-500': vm.state === 'SUSPENDED' && !vm.task_state,
                      'bg-orange-500 animate-pulse': vm.task_state,
                      'bg-gray-500': !['ACTIVE', 'STOPPED', 'ERROR', 'PAUSED', 'SUSPENDED'].includes(vm.state) && !vm.task_state
                    }"></span>
                    <span class="truncate">{{ vm.name }}</span>
                  </div>
                </li>
              </ul>
            </div>

            <div v-if="mainStore.discoveredByHost[host.id] && mainStore.discoveredByHost[host.id].length" class="mb-2">
                <div class="flex items-center justify-between">
                <div class="text-xs text-gray-400 uppercase font-semibold mb-1">Discovered</div>
                <div>
                  <button
                    :disabled="mainStore.isLoading.hostImportAll === `host:${host.id}:import-all`"
                    @click.stop.prevent="() => openConfirmAll(host.id)"
                    :aria-disabled="mainStore.isLoading.hostImportAll === `host:${host.id}:import-all` ? 'true' : 'false'"
                    :aria-busy="mainStore.isLoading.hostImportAll === `host:${host.id}:import-all` ? 'true' : 'false'"
                    :aria-label="`Import all discovered VMs on host ${host.id}`"
                    :class="[
                      'text-xs px-2 py-0.5 rounded text-white',
                      mainStore.isLoading.hostImportAll === `host:${host.id}:import-all` ? 'bg-indigo-600 opacity-50 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700'
                    ]"
                  >
                    <span v-if="mainStore.isLoading.hostImportAll === `host:${host.id}:import-all`" class="inline-block animate-spin w-3 h-3 border-2 border-white rounded-full border-t-transparent mr-1" aria-hidden="true"></span>
                    Import All
                  </button>
                </div>
              </div>
              <ul class="space-y-1">
                <li v-for="d in mainStore.discoveredByHost[host.id]" :key="d.uuid">
                  <div class="flex items-center justify-between p-1.5 text-sm rounded-md">
                    <div class="flex items-center cursor-pointer" @click="selectVm({ name: d.name })">
                      <svg class="h-3 w-3 mr-2 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3"/></svg>
                      <span class="truncate">{{ d.name }}</span>
                    </div>
                    <div>
                      <button
                        :disabled="mainStore.isLoading.vmImport === `${d.name}:import`"
                        @click.stop.prevent="() => openConfirmOne(host.id, d.name)"
                        :aria-disabled="mainStore.isLoading.vmImport === `${d.name}:import` ? 'true' : 'false'"
                        :aria-busy="mainStore.isLoading.vmImport === `${d.name}:import` ? 'true' : 'false'"
                        :aria-label="`Import discovered VM ${d.name} on host ${host.id}`"
                        :class="[
                          'text-xs px-2 py-0.5 rounded text-white',
                          mainStore.isLoading.vmImport === `${d.name}:import` ? 'bg-green-600 opacity-50 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'
                        ]"
                      >
                        <span v-if="mainStore.isLoading.vmImport === `${d.name}:import`" class="inline-block animate-spin w-3 h-3 border-2 border-white rounded-full border-t-transparent mr-1" aria-hidden="true"></span>
                        Import
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </li>
      </ul>
    </nav>
    <div class="flex-shrink-0 p-4 border-t border-gray-800">
        <div v-show="uiStore.isSidebarOpen">
            <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Status</h3>
            <div class="mt-2 text-sm text-gray-300">
                <span>{{ mainStore.hosts.length }} Host(s)</span>
                <span class="mx-2">|</span>
                <span>{{ mainStore.totalVms }} VM(s)</span>
            </div>
        </div>
        <div v-show="!uiStore.isSidebarOpen" class="text-center">
             <div class="text-xl font-bold">{{ mainStore.hosts.length }}</div>
             <div class="text-xs text-gray-400">Hosts</div>
        </div>
    </div>
  <ConfirmModal v-if="showConfirm" :title="confirmPayload.type === 'all' ? 'Import all VMs?' : 'Import VM?'" :message="confirmPayload.type === 'all' ? 'Import all discovered VMs on this host into management. This will create DB records for each VM.' : `Import VM '${confirmPayload.vmName || ''}' into management?`" confirmText="Import" cancelText="Cancel" :loading="confirmLoading" @confirm="handleConfirm" @cancel="handleCancel" />
  </aside>
</template>



