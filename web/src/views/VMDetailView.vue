<template>
  <div class="space-y-6">
    <!-- VM Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <FButton
          variant="ghost"
          size="sm"
          @click="$router.back()"
        >
          ← Back
        </FButton>
        <div>
          <h1 class="text-2xl font-bold text-white">{{ vm?.name || 'Loading...' }}</h1>
          <p class="text-slate-400">{{ vm?.description || 'VM Details' }}</p>
        </div>
      </div>
      
      <div v-if="vm" class="flex items-center gap-3">
        <div :class="[
          'w-3 h-3 rounded-full',
          getVMStatusColor(vm.state)
        ]"></div>
        <span :class="[
          'px-3 py-1 rounded-full text-sm font-medium',
          getVMStateBadgeClass(vm.state)
        ]">
          {{ (vm.state || 'UNKNOWN').toLowerCase() }}
        </span>
      </div>
    </div>

    <!-- VM Control Panel -->
    <FCard v-if="vm" class="p-6 card-glow">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-white">VM Controls</h2>
        <div v-if="vm.taskState" class="animate-pulse">
          <span class="px-2 py-1 rounded-full text-xs font-medium bg-yellow-500/20 text-yellow-400">
            {{ vm.taskState }}
          </span>
        </div>
      </div>
      
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <!-- Power Controls -->
        <FButton
          v-if="vm.state === 'STOPPED'"
          variant="primary"
          @click="handleVMAction('start')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          ▶️ Start
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('shutdown')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          ⏹️ Shutdown
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('reboot')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          🔄 Reboot
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="openConsole"
          class="flex items-center gap-2"
        >
          💻 Console
        </FButton>
      </div>

      <!-- Hardware Configuration Button -->
      <div class="mt-4 pt-4 border-t border-gray-700">
        <FButton
          variant="outline"
          @click="showExtendedHardwareModal = true"
          class="flex items-center gap-2"
        >
          🔧 Extended Hardware Configuration
        </FButton>
      </div>
    </FCard>

    <!-- VM Information Grid -->
    <div v-if="vm" class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <!-- Hardware Specs -->
      <FCard class="p-6 card-glow">
        <h3 class="text-lg font-semibold text-white mb-4">Hardware</h3>
        <div class="space-y-3">
          <div class="flex justify-between">
            <span class="text-slate-400">CPU Cores:</span>
            <span class="text-white font-medium">{{ vm.vcpuCount || 0 }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Memory:</span>
            <span class="text-white font-medium">{{ formatBytes((vm.memoryMB || 0) * 1024 * 1024) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Disk Size:</span>
            <span class="text-white font-medium">{{ vm.diskSizeGB || 0 }} GB</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">CPU Model:</span>
            <span class="text-white font-medium">{{ vm.cpuModel || 'Default' }}</span>
          </div>
        </div>
      </FCard>

      <!-- System Information -->
      <FCard class="p-6 card-glow">
        <h3 class="text-lg font-semibold text-white mb-4">System</h3>
        <div class="space-y-3">
          <div class="flex justify-between">
            <span class="text-slate-400">OS Type:</span>
            <span class="text-white font-medium">{{ vm.osType || 'Unknown' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Boot Device:</span>
            <span class="text-white font-medium">{{ vm.bootDevice || 'hd' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Network:</span>
            <span class="text-white font-medium">{{ vm.networkInterface || 'default' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">UUID:</span>
            <span class="text-white font-mono text-sm">{{ vm.uuid }}</span>
          </div>
        </div>
      </FCard>

      <!-- Status Information -->
      <FCard class="p-6 card-glow">
        <h3 class="text-lg font-semibold text-white mb-4">Status</h3>
        <div class="space-y-3">
          <div class="flex justify-between">
            <span class="text-slate-400">Current State:</span>
            <span :class="[
              'font-medium',
              vm.state === 'ACTIVE' ? 'text-green-400' : 
              vm.state === 'STOPPED' ? 'text-red-400' : 'text-yellow-400'
            ]">
              {{ vm.state }}
            </span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Libvirt State:</span>
            <span class="text-white font-medium">{{ vm.libvirtState }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Sync Status:</span>
            <span :class="[
              'font-medium',
              vm.syncStatus === 'SYNCED' ? 'text-green-400' : 
              vm.syncStatus === 'DRIFTED' ? 'text-yellow-400' : 'text-gray-400'
            ]">
              {{ vm.syncStatus }}
            </span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Source:</span>
            <span class="text-white font-medium capitalize">{{ vm.source }}</span>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Performance Stats -->
    <FCard v-if="vm && vm.state === 'ACTIVE'" class="p-6 card-glow">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-white">Performance Stats</h3>
        <div class="flex items-center gap-3">
          <FButton variant="outline" size="sm" @click="showMetricSettings = true">⚙️ Metrics</FButton>
          <div class="flex items-center gap-2 text-sm text-slate-400">
            <div class="text-sm text-slate-400">CPU Usage</div>
          </div>
          <FButton
            variant="ghost"
            size="sm"
            @click="refreshStats"
            :disabled="loadingStats"
          >
            <span v-if="!loadingStats">🔄 Refresh</span>
            <span v-else class="flex items-center gap-2">
              <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              Loading...
            </span>
          </FButton>
        </div>
      </div>
      
      <div v-if="vmStats" class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="text-center p-3 bg-white/5 rounded-lg">
          <div class="text-2xl font-bold text-primary-400">{{ cpuValue.toFixed(1) }}%</div>
          <div class="text-sm text-slate-400">CPU Usage ({{ cpuLabel }})</div>
        </div>
        <div class="text-center p-3 bg-white/5 rounded-lg">
          <div class="text-2xl font-bold text-accent-400">{{ formatBytes((vmStats.memory_mb || 0) * 1024 * 1024) }}</div>
          <div class="text-sm text-slate-400">Memory Usage</div>
        </div>
        <div class="text-center p-3 bg-white/5 rounded-lg col-span-2 md:col-span-1">
          <div class="text-sm text-slate-400">Disk I/O</div>
          <div class="text-3xl font-bold text-secondary-400">{{ formatDisk((vmStats.disk_read_kib_per_sec || 0)) }}</div>
          <div class="text-sm text-slate-400 mt-1">Read • <span class="font-medium">{{ (vmStats.disk_read_kib_per_sec || 0).toFixed(1) }} KiB/s</span></div>
          <div class="text-sm text-slate-400 mt-2">Write • <span class="font-medium">{{ (vmStats.disk_write_kib_per_sec || 0).toFixed(1) }} KiB/s</span></div>
          <div class="text-xs text-slate-500 mt-2">IOPS: R {{ (vmStats.disk_read_iops || 0).toFixed(1) }} • W {{ (vmStats.disk_write_iops || 0).toFixed(1) }}</div>
        </div>
        <div class="text-center p-3 bg-white/5 rounded-lg">
          <div class="text-sm text-slate-400">Network</div>
          <div class="text-3xl font-bold text-secondary-400">{{ formatNetwork((vmStats.network_rx_mbps || 0)) }}</div>
          <div class="text-sm text-slate-400 mt-1">RX • <span class="font-medium">{{ (vmStats.network_rx_mbps || vmStats.network_rx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</span></div>
          <div class="text-sm text-slate-400 mt-2">TX • <span class="font-medium">{{ (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</span></div>
        </div>

        <div class="text-center p-3 bg-white/5 rounded-lg">
          <div class="text-2xl font-bold text-primary-400">{{ formatUptime(vmStats.uptime || 0) }}</div>
          <div class="text-sm text-slate-400">Uptime</div>
        </div>
      </div>
      
      <div v-else class="text-center py-8 text-slate-400">
        No performance data available
      </div>
    </FCard>

    <!-- Advanced Actions -->
    <FCard v-if="vm" class="p-6 card-glow">
      <h3 class="text-lg font-semibold text-white mb-4">Advanced Actions</h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <FButton
          variant="ghost"
          @click="handleVMAction('sync')"
          :disabled="!!vm.taskState"
        >
          🔄 Sync from Libvirt
        </FButton>
        <FButton
          variant="ghost"
          @click="handleVMAction('rebuild')"
          :disabled="!!vm.taskState"
        >
          🏗️ Rebuild from DB
        </FButton>
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('forceOff')"
          :disabled="!!vm.taskState"
          class="text-orange-400 hover:bg-orange-500/10"
        >
          ⚡ Force Off
        </FButton>
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('forceReset')"
          :disabled="!!vm.taskState"
          class="text-red-400 hover:bg-red-500/10"
        >
          ⚡ Force Reset
        </FButton>
      </div>
    </FCard>

    <!-- Loading State -->
    <div v-if="!vm" class="flex items-center justify-center py-12">
      <div class="flex items-center gap-3">
        <div class="w-6 h-6 border-2 border-primary-400 border-t-transparent rounded-full animate-spin"></div>
        <span class="text-white">Loading VM details...</span>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 bg-red-500/10 border border-red-400/20 rounded-lg">
      <p class="text-red-400">{{ error }}</p>
    </div>

    <!-- Extended Hardware Configuration Modal -->
    <VMHardwareConfigModalExtended
      v-if="vm"
      :show="showExtendedHardwareModal"
      :host-id="props.hostId"
      :vm-name="vm.name"
      @close="showExtendedHardwareModal = false"
      @hardware-updated="loadVM"
    />

    <!-- Metrics Settings Modal (overlay) -->
    <MetricSettingsModal
      v-if="showMetricSettings"
      :show="showMetricSettings"
      @close="showMetricSettings = false"
      @applied="refreshStats"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useVMStore } from '@/stores/vmStore';
import { useUIStore } from '@/stores/uiStore';
import { useSettingsStore } from '@/stores/settingsStore';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import VMHardwareConfigModalExtended from '@/components/modals/VMHardwareConfigModalExtended.vue';
import MetricSettingsModal from '@/components/modals/MetricSettingsModal.vue';
import type { VirtualMachine, VMStats } from '@/types';
import { wsManager } from '@/services/api';

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const route = useRoute();
const router = useRouter();

const vmStore = useVMStore();
const uiStore = useUIStore();

// Component state
const vm = ref<VirtualMachine | null>(null);
const vmStats = ref<VMStats | null>(null);
const error = ref<string | null>(null);
const loadingStats = ref(false);
const showExtendedHardwareModal = ref(false);
// simplified CPU display: show smoothed host-normalized `cpu_percent`
const showMetricSettings = ref(false);

const settings = useSettingsStore();

function formatDisk(valueKiB: number) {
  if (settings.units.disk === 'mib') return (valueKiB/1024).toFixed(2) + ' MiB/s'
  return valueKiB.toFixed(1) + ' KiB/s'
}

function formatNetwork(valueMBps: number) {
  if (settings.units.network === 'kb') return (valueMBps*1024).toFixed(1) + ' KB/s'
  return valueMBps.toFixed(2) + ' MB/s'
}

const cpuValue = computed(() => {
  if (!vmStats.value) return 0
  const s = settings.cpuDisplayDefault
  if (s === 'guest') return (vmStats.value.cpu_percent_guest ?? vmStats.value.cpu_percent ?? 0)
  if (s === 'raw') return (vmStats.value.cpu_percent_raw ?? vmStats.value.cpu_percent_core ?? vmStats.value.cpu_percent ?? 0)
  // default host
  return (vmStats.value.cpu_percent_host ?? vmStats.value.cpu_percent ?? 0)
})

const cpuLabel = computed(() => {
  const s = settings.cpuDisplayDefault
  if (s === 'guest') return 'Guest'
  if (s === 'raw') return 'Raw %'
  return 'Host'
})

// Get VM data
const loadVM = async (): Promise<void> => {
  try {
    error.value = null;
    
    // Stop current monitoring if active
    stopStatsMonitoring();
    
    await vmStore.fetchVMs(props.hostId);
    
    // Find the VM by name
    const foundVM = vmStore.vmsByHost(props.hostId).find((v: VirtualMachine) => v.name === props.vmName);
    if (foundVM) {
      vm.value = foundVM;
      
      // Start monitoring if VM is active
      if (foundVM.state === 'ACTIVE') {
        startStatsMonitoring();
      }
    } else {
      error.value = `VM "${props.vmName}" not found`;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load VM details';
  }
};

// Refresh performance stats
const refreshStats = async (): Promise<void> => {
  if (!vm.value || vm.value.state !== 'ACTIVE') return;
  
  loadingStats.value = true;
  try {
    await vmStore.fetchVMStats(props.hostId, vm.value.name);
    // Get stats from store after fetching
    const statsKey = `${props.hostId}:${vm.value.name}`;
    vmStats.value = vmStore.vmStats[statsKey] || null;
  } catch (err) {
    console.error('Failed to load VM stats:', err);
  } finally {
    loadingStats.value = false;
  }
};

// Handle VM actions
const handleVMAction = async (action: string): Promise<void> => {
  if (!vm.value) return;
  
  try {
    error.value = null;
    
    switch (action) {
      case 'start':
        await vmStore.startVM(props.hostId, vm.value.name);
        break;
      case 'shutdown':
        await vmStore.stopVM(props.hostId, vm.value.name);
        break;
      case 'reboot':
        await vmStore.restartVM(props.hostId, vm.value.name);
        break;
      case 'forceOff':
        await vmStore.forceOffVM(props.hostId, vm.value.name);
        break;
      case 'forceReset':
        await vmStore.forceResetVM(props.hostId, vm.value.name);
        break;
      case 'sync':
        await vmStore.syncVM(props.hostId, vm.value.name);
        break;
      case 'rebuild':
        await vmStore.rebuildVM(props.hostId, vm.value.name);
        break;
    }
    
    // Refresh VM data after action
    await loadVM();
  } catch (err) {
    error.value = err instanceof Error ? err.message : `Failed to ${action} VM`;
  }
};

// Open console
const openConsole = (): void => {
  if (!vm.value) return;
  
  // Navigate to SPICE console route
  router.push(`/spice/${props.hostId}/${vm.value.name}`);
};

// Utility functions
const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-400';
    case 'STOPPED': return 'bg-red-400';
    case 'PAUSED': return 'bg-yellow-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-gray-400';
  }
};

const getVMStateBadgeClass = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500/20 text-green-400';
    case 'STOPPED': return 'bg-red-500/20 text-red-400';
    case 'PAUSED': return 'bg-yellow-500/20 text-yellow-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-gray-500/20 text-gray-400';
  }
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
  
  if (days > 0) {
    return `${days}d ${hours}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
};

// WebSocket-based stats monitoring
let isSubscribed = false;

const startStatsMonitoring = (): void => {
  if (vm.value?.state === 'ACTIVE' && !isSubscribed) {
    console.log(`Subscribing to VM stats: ${props.hostId}/${vm.value.name}`);
    
    // Connect WebSocket if not connected
    wsManager.connect().then(() => {
      // Subscribe to stats updates
      wsManager.subscribeToVMStats(props.hostId, vm.value!.name);
      isSubscribed = true;
      
      // Listen for stats updates
      wsManager.on('vm-stats-updated', handleStatsUpdate);
      
      // Also do an initial fetch
      refreshStats();
    }).catch(error => {
      console.error('Failed to connect WebSocket:', error);
      // Fallback to initial fetch only
      refreshStats();
    });
  }
};

const stopStatsMonitoring = (): void => {
  if (isSubscribed && vm.value) {
    console.log(`Unsubscribing from VM stats: ${props.hostId}/${vm.value.name}`);
    wsManager.unsubscribeFromVMStats(props.hostId, vm.value.name);
    wsManager.off('vm-stats-updated', handleStatsUpdate);
    isSubscribed = false;
  }
};

// Handle incoming WebSocket stats updates
const handleStatsUpdate = (data: any): void => {
  if (data.hostId === props.hostId && data.vmName === vm.value?.name) {
    // Update vmStats with the received data
    if (data.stats) {
      vmStats.value = data.stats;
      console.log('Received VM stats update via WebSocket:', data.stats);
    }
  }
};

// Lifecycle
onMounted(() => {
  loadVM().then(() => {
    startStatsMonitoring();
  });
});

onUnmounted(() => {
  stopStatsMonitoring();
});
</script>