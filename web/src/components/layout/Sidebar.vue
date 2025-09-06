<script setup>
import { useUiStore } from '@/stores/uiStore';
import { useMainStore } from '@/stores/mainStore';
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';

const uiStore = useUiStore();
const mainStore = useMainStore();
const router = useRouter();

onMounted(() => {
  mainStore.initializeRealtime();
  mainStore.fetchHosts();
});

function selectHost(hostId) {
  mainStore.selectHost(hostId);
  router.push({ name: 'host-dashboard', params: { hostId } });
}

function selectVm(vm) {
  // Find the host for this VM to ensure the selectedHostId is correct
  for (const host of mainStore.hosts) {
      if (host.vms && host.vms.find(hvm => hvm.name === vm.name)) {
          mainStore.selectHost(host.id);
          break;
      }
  }
  router.push({ name: 'vm-view', params: { vmName: vm.name } });
}
</script>

<template>
  <aside 
    class="flex flex-col bg-gray-900 text-gray-300 transition-all duration-300 ease-in-out"
    :class="uiStore.isSidebarOpen ? 'w-64' : 'w-20'"
  >
    <div class="flex items-center h-16 px-6 border-b border-gray-800">
      <h1 class="text-xl font-bold text-white tracking-wider" v-show="uiStore.isSidebarOpen">
        Virtu<span class="text-indigo-400">Mancer</span>
      </h1>
      <img src="/favicon.ico" alt="Logo" class="h-8 w-8" v-show="!uiStore.isSidebarOpen">
    </div>

    <nav class="flex-1 overflow-y-auto py-4">
      <div class="px-4 mb-4">
        <button @click="uiStore.openAddHostModal" class="w-full flex items-center justify-center p-2 rounded-md bg-indigo-600 text-white hover:bg-indigo-700 transition-colors">
          <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
          <span class="ml-2" v-show="uiStore.isSidebarOpen">Add Host</span>
        </button>
      </div>
      
      <ul>
        <li v-for="host in mainStore.hosts" :key="host.id" class="px-4 mb-2">
          <div 
            @click="selectHost(host.id)" 
            class="flex items-center p-2 rounded-md cursor-pointer hover:bg-gray-700"
            :class="{ 'bg-gray-700 text-white': mainStore.selectedHostId === host.id }"
          >
            <svg class="h-6 w-6 flex-shrink-0" :class="{'text-indigo-400': mainStore.selectedHostId === host.id}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/></svg>
            <span class="ml-3 font-semibold" v-show="uiStore.isSidebarOpen">{{ host.id }}</span>
          </div>
          <ul v-if="uiStore.isSidebarOpen && host.vms && host.vms.length" class="ml-6 mt-1 space-y-1 border-l-2 border-gray-700 pl-4">
            <li v-for="vm in host.vms" :key="vm.name">
              <div @click="selectVm(vm)" class="flex items-center p-1.5 text-sm rounded-md cursor-pointer hover:bg-gray-700">
                <span class="h-2 w-2 rounded-full mr-2 flex-shrink-0" :class="{
                  'bg-green-500': vm.state === 1, 'bg-red-500': vm.state === 5,
                  'bg-yellow-500': vm.state === 3, 'bg-gray-500': ![1,3,5].includes(vm.state)
                }"></span>
                <span class="truncate">{{ vm.name }}</span>
              </div>
            </li>
          </ul>
        </li>
      </ul>
    </nav>
  </aside>
</template>


