<template>
  <div class="space-y-6">
    <!-- Host Overview Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-4 gap-6">
      <FCard
        v-for="host in hostsWithStats"
        :key="host.id"
        :class="[
          'cursor-pointer transition-all duration-300',
          selectedHostId === host.id ? 'ring-2 ring-primary-400' : ''
        ]"
        :border-glow="host.state === 'CONNECTED'"
        :glow-color="getHostGlowColor(host)"
        interactive
        @click="selectHost(host.id)"
      >
        <div class="space-y-4">
          <!-- Host Header -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div :class="[
                'w-3 h-3 rounded-full',
                getHostStatusColor(host)
              ]"></div>
              <div>
                <h3 class="font-semibold text-white">{{ getHostDisplayName(host) }}</h3>
                <p class="text-sm text-slate-400">{{ host.uri }}</p>
              </div>
            </div>
            
            <!-- Connection Status -->
            <div class="flex items-center gap-2">
              <div v-if="host.isConnecting" class="animate-spin w-4 h-4 border-2 border-primary-400 border-t-transparent rounded-full"></div>
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium',
                getHostStatusBadgeClass(host)
              ]">
                {{ getHostStatusText(host) }}
              </span>
            </div>
          </div>

          <!-- Host Stats (if connected) -->
          <div v-if="host.stats" class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-slate-400">VMs</div>
              <div class="text-xl font-bold text-white">{{ host.stats.vm_count || 0 }}</div>
            </div>
            <div>
              <div class="text-sm text-slate-400">CPU</div>
              <div class="text-xl font-bold text-white">{{ Math.round(host.stats.cpu_percent || 0) }}%</div>
            </div>
            <div>
              <div class="text-sm text-slate-400">Memory</div>
              <div class="text-xl font-bold text-white">
                {{ formatBytes(host.stats.memory_total - host.stats.memory_available) }} / 
                {{ formatBytes(host.stats.memory_total) }}
              </div>
            </div>
            <div>
              <div class="text-sm text-slate-400">Uptime</div>
              <div class="text-xl font-bold text-white">{{ formatUptime(host.stats.uptime) }}</div>
            </div>
          </div>

          <!-- Host Actions -->
          <div class="flex gap-2 pt-2 border-t border-white/10">
            <FButton
              v-if="host.state === 'DISCONNECTED'"
              size="sm"
              variant="primary"
              :disabled="loading.connectHost[host.id]"
              @click.stop="connectHost(host.id)"
            >
              {{ loading.connectHost[host.id] ? 'Connecting...' : 'Connect' }}
            </FButton>
            
            <FButton
              v-if="host.state === 'CONNECTED'"
              size="sm"
              variant="ghost"
              @click.stop="disconnectHost(host.id)"
            >
              Disconnect
            </FButton>
            
            <FButton
              v-if="host.state === 'CONNECTED'"
              size="sm"
              variant="accent"
              @click.stop="refreshHostData(host.id)"
            >
              Refresh
            </FButton>
            
            <FButton
              size="sm"
              variant="ghost"
              @click.stop="openHostModal(host)"
            >
              ⚙️
            </FButton>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Add Host Button -->
    <div class="flex justify-center">
      <FButton
        variant="primary"
        glow
        @click="openAddHostModal"
        :disabled="loading.addHost"
      >
        <div class="flex items-center gap-2">
          <span>{{ loading.addHost ? 'Adding...' : '+ Add Host' }}</span>
        </div>
      </FButton>
    </div>

    <!-- Selected Host Details -->
    <div v-if="selectedHost" class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-2xl font-bold text-white">{{ getHostDisplayName(selectedHost) }}</h2>
        <div class="flex gap-2">
          <FButton
            v-if="selectedHost.state === 'CONNECTED'"
            variant="accent"
            size="sm"
            @click="importAllVMs(selectedHost.id)"
            :disabled="loading.hostImportAll === selectedHost.id"
          >
            {{ loading.hostImportAll === selectedHost.id ? 'Importing...' : 'Import All VMs' }}
          </FButton>
          <FButton
            variant="ghost"
            size="sm"
            @click="refreshDiscoveredVMs(selectedHost.id)"
          >
            🔄 Refresh
          </FButton>
        </div>
      </div>

      <!-- VM Management Tabs -->
      <div class="space-y-4">
        <div class="flex gap-4 border-b border-white/10">
          <button
            v-for="tab in vmTabs"
            :key="tab.id"
            :class="[
              'px-4 py-2 font-medium transition-colors',
              activeVMTab === tab.id
                ? 'text-primary-400 border-b-2 border-primary-400'
                : 'text-slate-400 hover:text-white'
            ]"
            @click="activeVMTab = tab.id"
          >
            {{ tab.label }}
            <span v-if="tab.count !== undefined" class="ml-2 px-2 py-1 bg-white/10 rounded-full text-xs">
              {{ tab.count }}
            </span>
          </button>
        </div>

        <!-- VM Lists -->
        <div v-if="activeVMTab === 'managed'" class="space-y-4">
          <h3 class="text-lg font-semibold text-white">Managed VMs</h3>
          <div v-if="managedVMs.length === 0" class="text-center py-8 text-slate-400">
            No managed VMs found. Import some VMs to get started.
          </div>
          <VMCard
            v-for="vm in managedVMs"
            :key="vm.uuid"
            :vm="vm"
            :host-id="selectedHost.id"
            @action="handleVMAction"
          />
        </div>

        <div v-if="activeVMTab === 'discovered'" class="space-y-4">
          <h3 class="text-lg font-semibold text-white">Discovered VMs</h3>
          <div v-if="discoveredVMs.length === 0" class="text-center py-8 text-slate-400">
            No discovered VMs found.
          </div>
          <DiscoveredVMCard
            v-for="vm in discoveredVMs"
            :key="vm.uuid"
            :vm="vm"
            :host-id="selectedHost.id"
            @import="importVM"
          />
        </div>
      </div>
    </div>

    <!-- Add Host Modal -->
    <AddHostModal
      :open="modals.addHost"
      @close="closeAddHostModal"
      @submit="handleAddHost"
    />

    <!-- Host Settings Modal -->
    <HostSettingsModal
      :open="modals.hostSettings"
      :host="selectedHostForSettings"
      @close="closeHostModal"
      @save="handleSaveHost"
      @delete="handleDeleteHost"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useHostStore, useVMStore, useUIStore } from '@/stores';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import VMCard from '@/components/vm/VMCard.vue';
import DiscoveredVMCard from '@/components/vm/DiscoveredVMCard.vue';
import AddHostModal from '@/components/modals/AddHostModal.vue';
import HostSettingsModal from '@/components/modals/HostSettingsModal.vue';
import type { Host, VirtualMachine, DiscoveredVM } from '@/types';

// Store instances
const hostStore = useHostStore();
const vmStore = useVMStore();
const uiStore = useUIStore();

// Local state
const selectedHostId = ref<string | null>(null);
const activeVMTab = ref('managed');
const selectedHostForSettings = ref<Host | null>(null);

const modals = ref({
  addHost: false,
  hostSettings: false
});

// VM tabs with dynamic counts
const vmTabs = computed(() => [
  {
    id: 'managed',
    label: 'Managed VMs',
    count: managedVMs.value.length
  },
  {
    id: 'discovered',
    label: 'Discovered VMs',
    count: discoveredVMs.value.length
  }
]);

// Computed properties
const hostsWithStats = computed(() => hostStore.hostsWithStats);
const loading = computed(() => hostStore.loading);
const selectedHost = computed(() => 
  selectedHostId.value ? hostStore.getHostById(selectedHostId.value) : null
);

const managedVMs = computed(() => 
  selectedHostId.value ? vmStore.vmsByHost(selectedHostId.value) : []
);

const discoveredVMs = computed(() => 
  selectedHostId.value ? hostStore.discoveredVMs[selectedHostId.value] || [] : []
);

// Actions
const selectHost = (hostId: string): void => {
  selectedHostId.value = hostId;
  hostStore.selectHost(hostId);
  
  // Fetch VMs for this host
  if (hostStore.getHostById(hostId)?.state === 'CONNECTED') {
    vmStore.fetchVMs(hostId);
    hostStore.refreshDiscoveredVMs(hostId);
  }
};

const connectHost = async (hostId: string): Promise<void> => {
  try {
    await hostStore.connectHost(hostId);
    uiStore.addToast('Host connection initiated', 'success');
  } catch (error) {
    uiStore.addToast('Failed to connect to host', 'error');
  }
};

const disconnectHost = async (hostId: string): Promise<void> => {
  try {
    await hostStore.disconnectHost(hostId);
    uiStore.addToast('Host disconnected', 'success');
  } catch (error) {
    uiStore.addToast('Failed to disconnect host', 'error');
  }
};

const refreshHostData = async (hostId: string): Promise<void> => {
  try {
    await Promise.all([
      hostStore.fetchHostStats(hostId),
      vmStore.fetchVMs(hostId),
      hostStore.refreshDiscoveredVMs(hostId)
    ]);
    uiStore.addToast('Host data refreshed', 'success');
  } catch (error) {
    uiStore.addToast('Failed to refresh host data', 'error');
  }
};

const refreshDiscoveredVMs = async (hostId: string): Promise<void> => {
  try {
    await hostStore.refreshDiscoveredVMs(hostId);
    uiStore.addToast('Discovered VMs refreshed', 'success');
  } catch (error) {
    uiStore.addToast('Failed to refresh discovered VMs', 'error');
  }
};

const importAllVMs = async (hostId: string): Promise<void> => {
  try {
    await hostStore.importAllVMs(hostId);
    uiStore.addToast('All VMs imported successfully', 'success');
    // Refresh managed VMs
    await vmStore.fetchVMs(hostId);
  } catch (error) {
    uiStore.addToast('Failed to import VMs', 'error');
  }
};

const importVM = async (hostId: string, vmName: string): Promise<void> => {
  try {
    await vmStore.importVM(hostId, vmName);
    uiStore.addToast(`VM ${vmName} imported successfully`, 'success');
    // Refresh both lists
    await Promise.all([
      vmStore.fetchVMs(hostId),
      hostStore.refreshDiscoveredVMs(hostId)
    ]);
  } catch (error) {
    uiStore.addToast(`Failed to import VM ${vmName}`, 'error');
  }
};

const handleVMAction = async (action: string, hostId: string, vmName: string): Promise<void> => {
  try {
    switch (action) {
      case 'start':
        await vmStore.startVM(hostId, vmName);
        break;
      case 'shutdown':
        await vmStore.stopVM(hostId, vmName);
        break;
      case 'reboot':
        await vmStore.restartVM(hostId, vmName);
        break;
      // Add more actions as needed
    }
    uiStore.addToast(`VM ${action} initiated`, 'success');
  } catch (error) {
    uiStore.addToast(`Failed to ${action} VM`, 'error');
  }
};

// Modal handlers
const openAddHostModal = (): void => {
  modals.value.addHost = true;
};

const closeAddHostModal = (): void => {
  modals.value.addHost = false;
};

const openHostModal = (host: Host): void => {
  selectedHostForSettings.value = host;
  modals.value.hostSettings = true;
};

const closeHostModal = (): void => {
  modals.value.hostSettings = false;
  selectedHostForSettings.value = null;
};

const handleAddHost = async (hostData: Omit<Host, 'id'>): Promise<void> => {
  try {
    await hostStore.addHost(hostData);
    uiStore.addToast('Host added successfully', 'success');
    closeAddHostModal();
  } catch (error) {
    uiStore.addToast('Failed to add host', 'error');
  }
};

const handleSaveHost = async (hostData: Partial<Host>): Promise<void> => {
  if (!selectedHostForSettings.value) return;
  
  try {
    await hostStore.updateHost(selectedHostForSettings.value.id, hostData);
    uiStore.addToast('Host updated successfully', 'success');
    closeHostModal();
  } catch (error) {
    uiStore.addToast('Failed to update host', 'error');
  }
};

const handleDeleteHost = async (): Promise<void> => {
  if (!selectedHostForSettings.value) return;
  
  try {
    await hostStore.deleteHost(selectedHostForSettings.value.id);
    uiStore.addToast('Host deleted successfully', 'success');
    closeHostModal();
    
    // Clear selection if this was the selected host
    if (selectedHostId.value === selectedHostForSettings.value.id) {
      selectedHostId.value = null;
    }
  } catch (error) {
    uiStore.addToast('Failed to delete host', 'error');
  }
};

// Utility functions
const getHostDisplayName = (host: Host): string => {
  return host.id || 'Unknown Host';
};

const getHostStatusColor = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-400';
    case 'DISCONNECTED': return 'bg-red-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-yellow-400';
  }
};

const getHostGlowColor = (host: any): 'primary' | 'accent' | 'neon-blue' | 'neon-cyan' | 'neon-purple' => {
  switch (host.state) {
    case 'CONNECTED': return 'accent';
    case 'ERROR': return 'neon-purple';
    default: return 'primary';
  }
};

const getHostStatusBadgeClass = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-500/20 text-green-400';
    case 'DISCONNECTED': return 'bg-red-500/20 text-red-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-yellow-500/20 text-yellow-400';
  }
};

const getHostStatusText = (host: any): string => {
  if (host.isConnecting) return 'Connecting...';
  return host.state?.toLowerCase() || 'unknown';
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
};

// Lifecycle
onMounted(async () => {
  // Load initial data
  await hostStore.fetchHosts();
  
  // Select first connected host by default
  const connectedHost = hostStore.connectedHosts[0];
  if (connectedHost) {
    selectHost(connectedHost.id);
  }
});

onUnmounted(() => {
  // Cleanup if needed
});
</script>