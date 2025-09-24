<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-white">Networks</h1>
        <p class="text-slate-400">Network infrastructure and topology overview</p>
      </div>
      
      <div class="flex items-center gap-3">
        <FButton
          variant="ghost"
          size="sm"
          @click="refreshTopology"
          :disabled="isLoading"
        >
          <span v-if="!isLoading">🔄 Refresh</span>
          <span v-else class="flex items-center gap-2">
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Loading...
          </span>
        </FButton>
        
        <div class="flex items-center gap-2">
          <label class="text-sm text-white">View:</label>
          <select
            v-model="viewMode"
            class="px-3 py-1 bg-white/10 border border-white/20 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
          >
            <option value="grid">Grid View</option>
            <option value="network">Network Diagram</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Network Topology Section -->
    <div class="border-l-4 border-primary-500 pl-4 mb-6">
      <h2 class="text-xl font-semibold text-white mb-2">Network Topology</h2>
      <p class="text-slate-400 text-sm">Visualize infrastructure relationships and connections</p>
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <FCard class="p-4 text-center card-glow">
        <div class="text-2xl font-bold text-primary-400">{{ totalHosts }}</div>
        <div class="text-sm text-slate-400">Total Hosts</div>
      </FCard>
      
      <FCard class="p-4 text-center card-glow">
        <div class="text-2xl font-bold text-accent-400">{{ totalVMs }}</div>
        <div class="text-sm text-slate-400">Total VMs</div>
      </FCard>
      
      <FCard class="p-4 text-center card-glow">
        <div class="text-2xl font-bold text-blue-400">{{ connectedHosts }}</div>
        <div class="text-sm text-slate-400">Connected Hosts</div>
      </FCard>
      
      <FCard class="p-4 text-center card-glow">
        <div class="text-2xl font-bold text-cyan-400">{{ activeVMs }}</div>
        <div class="text-sm text-slate-400">Active VMs</div>
      </FCard>
    </div>

    <!-- Grid View -->
    <div v-if="viewMode === 'grid'" class="space-y-6">
      <div
        v-for="host in hosts"
        :key="host.id"
        class="space-y-4"
      >
        <!-- Host Header -->
        <FCard
          class="p-4 cursor-pointer transition-all duration-300 hover:scale-[1.02] card-glow"
          interactive
          @click="$router.push(`/hosts/${host.id}`)"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div :class="[
                'w-4 h-4 rounded-full',
                getHostStatusColor(host)
              ]"></div>
              <div>
                <h3 class="text-lg font-semibold text-white">{{ getHostDisplayName(host) }}</h3>
                <p class="text-sm text-slate-400">{{ host.uri }}</p>
              </div>
            </div>
            
            <div class="flex items-center gap-4">
              <div class="text-right">
                <div class="text-sm text-slate-400">VMs</div>
                <div class="text-lg font-semibold text-white">{{ getHostVMs(host.id).length }}</div>
              </div>
              <span :class="[
                'px-3 py-1 rounded-full text-sm font-medium',
                getHostStatusBadgeClass(host)
              ]">
                {{ host.state.toLowerCase() }}
              </span>
            </div>
          </div>
        </FCard>

        <!-- Host VMs -->
        <div v-if="getHostVMs(host.id).length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 ml-8">
          <FCard
            v-for="vm in getHostVMs(host.id)"
            :key="vm.uuid"
            class="p-4 cursor-pointer transition-all duration-300 hover:scale-[1.02] card-glow"
            interactive
            @click="$router.push(`/hosts/${host.id}/vms/${vm.name}`)"
          >
            <div class="flex items-center gap-3">
              <div :class="[
                'w-3 h-3 rounded-full',
                getVMStatusColor(vm.state)
              ]"></div>
              <div class="flex-1">
                <h4 class="font-medium text-white">{{ vm.name }}</h4>
                <p class="text-xs text-slate-400">{{ vm.osType || 'Unknown OS' }}</p>
              </div>
              <div class="text-right">
                <div class="text-xs text-slate-400">{{ vm.vcpuCount }}c</div>
                <div class="text-xs text-slate-400">{{ formatBytes(vm.memoryMB * 1024 * 1024) }}</div>
              </div>
            </div>
          </FCard>
        </div>
        
        <div v-else class="ml-8 p-4 text-center text-slate-400 bg-white/5 rounded-lg border border-white/10">
          No VMs on this host
        </div>
      </div>
    </div>

    <!-- Network Diagram View -->
    <div v-if="viewMode === 'network'" class="relative">
      <FCard class="p-8 min-h-[600px] overflow-hidden card-glow">
        <div class="relative w-full h-full">
          <!-- Network topology visualization -->
          <svg
            ref="networkSvg"
            class="w-full h-full"
            viewBox="0 0 1200 800"
          >
            <!-- Host nodes -->
            <g v-for="(host, index) in hosts" :key="host.id">
              <!-- Host circle -->
              <circle
                :cx="getHostPosition(index).x"
                :cy="getHostPosition(index).y"
                :r="50 + (getHostVMs(host.id).length * 2)"
                :fill="getHostNodeColor(host)"
                :stroke="getHostStrokeColor(host)"
                stroke-width="3"
                class="cursor-pointer transition-all duration-300 hover:opacity-80"
                @click="$router.push(`/hosts/${host.id}`)"
              />
              
              <!-- Host label -->
              <text
                :x="getHostPosition(index).x"
                :y="getHostPosition(index).y - 70"
                text-anchor="middle"
                class="fill-white text-sm font-medium pointer-events-none"
              >
                {{ getHostDisplayName(host) }}
              </text>
              
              <!-- Host status -->
              <text
                :x="getHostPosition(index).x"
                :y="getHostPosition(index).y + 5"
                text-anchor="middle"
                class="fill-white text-xs opacity-75 pointer-events-none"
              >
                {{ getHostVMs(host.id).length }} VMs
              </text>
              
              <!-- VM nodes around host -->
              <g v-for="(vm, vmIndex) in getHostVMs(host.id).slice(0, 8)" :key="vm.uuid">
                <!-- Connection line -->
                <line
                  :x1="getHostPosition(index).x"
                  :y1="getHostPosition(index).y"
                  :x2="getVMPosition(index, vmIndex).x"
                  :y2="getVMPosition(index, vmIndex).y"
                  :stroke="getVMConnectionColor(vm)"
                  stroke-width="2"
                  opacity="0.6"
                />
                
                <!-- VM circle -->
                <circle
                  :cx="getVMPosition(index, vmIndex).x"
                  :cy="getVMPosition(index, vmIndex).y"
                  r="15"
                  :fill="getVMNodeColor(vm)"
                  :stroke="getVMStrokeColor(vm)"
                  stroke-width="2"
                  class="cursor-pointer transition-all duration-300 hover:opacity-80"
                  @click="$router.push(`/hosts/${host.id}/vms/${vm.name}`)"
                />
                
                <!-- VM label -->
                <text
                  :x="getVMPosition(index, vmIndex).x"
                  :y="getVMPosition(index, vmIndex).y - 25"
                  text-anchor="middle"
                  class="fill-white text-xs pointer-events-none"
                >
                  {{ vm.name.slice(0, 8) }}
                </text>
              </g>
              
              <!-- More VMs indicator -->
              <text
                v-if="getHostVMs(host.id).length > 8"
                :x="getHostPosition(index).x"
                :y="getHostPosition(index).y + 20"
                text-anchor="middle"
                class="fill-slate-400 text-xs pointer-events-none"
              >
                +{{ getHostVMs(host.id).length - 8 }} more
              </text>
            </g>
          </svg>
          
          <!-- Legend -->
          <div class="absolute bottom-4 left-4 p-4 bg-black/50 backdrop-blur-sm rounded-lg border border-white/20">
            <h4 class="text-sm font-medium text-white mb-2">Legend</h4>
            <div class="space-y-1 text-xs">
              <div class="flex items-center gap-2">
                <div class="w-3 h-3 rounded-full bg-green-400"></div>
                <span class="text-green-400">Connected/Active</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-3 h-3 rounded-full bg-red-400"></div>
                <span class="text-red-400">Disconnected/Stopped</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-3 h-3 rounded-full bg-yellow-400"></div>
                <span class="text-yellow-400">Paused/Warning</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="flex items-center justify-center py-12">
      <div class="flex items-center gap-3">
        <div class="w-6 h-6 border-2 border-primary-400 border-t-transparent rounded-full animate-spin"></div>
        <span class="text-white">Loading network topology...</span>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 bg-red-500/10 border border-red-400/20 rounded-lg">
      <p class="text-red-400">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useHostStore } from '@/stores/hostStore';
import { useVMStore } from '@/stores/vmStore';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';
import type { Host, VirtualMachine } from '@/types';

const hostStore = useHostStore();
const vmStore = useVMStore();

// Component state
const viewMode = ref<'grid' | 'network'>('grid');
const isLoading = ref(false);
const error = ref<string | null>(null);

// Data
const hosts = computed(() => hostStore.hosts);
const vms = computed(() => vmStore.vms);

// Stats
const totalHosts = computed(() => hosts.value.length);
const totalVMs = computed(() => vms.value.length);
const connectedHosts = computed(() => hosts.value.filter(h => h.state === 'CONNECTED').length);
const activeVMs = computed(() => vms.value.filter(vm => vm.state === 'ACTIVE').length);

// Get VMs for a specific host
const getHostVMs = (hostId: string): VirtualMachine[] => {
  return vmStore.vmsByHost(hostId);
};

// Load all data
const refreshTopology = async (): Promise<void> => {
  isLoading.value = true;
  error.value = null;
  
  try {
    await hostStore.fetchHosts();
    
    // Load VMs for each connected host
    const connectedHostIds = hosts.value
      .filter(h => h.state === 'CONNECTED')
      .map(h => h.id);
    
    await Promise.all(
      connectedHostIds.map(hostId => vmStore.fetchVMs(hostId))
    );
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load topology';
  } finally {
    isLoading.value = false;
  }
};

// Network diagram positioning
const getHostPosition = (index: number): { x: number; y: number } => {
  const centerX = 600;
  const centerY = 400;
  const radius = 250;
  const angle = (index * 2 * Math.PI) / hosts.value.length;
  
  return {
    x: centerX + radius * Math.cos(angle),
    y: centerY + radius * Math.sin(angle)
  };
};

const getVMPosition = (hostIndex: number, vmIndex: number): { x: number; y: number } => {
  const hostPos = getHostPosition(hostIndex);
  const vmRadius = 80;
  const vmAngle = (vmIndex * 2 * Math.PI) / 8; // Max 8 VMs shown
  
  return {
    x: hostPos.x + vmRadius * Math.cos(vmAngle),
    y: hostPos.y + vmRadius * Math.sin(vmAngle)
  };
};

// Color functions for hosts
const getHostDisplayName = (host: Host): string => {
  const uri = new URL(host.uri);
  return uri.hostname || host.uri;
};

const getHostStatusColor = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-400';
    case 'DISCONNECTED': return 'bg-red-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-gray-400';
  }
};

const getHostStatusBadgeClass = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-500/20 text-green-400';
    case 'DISCONNECTED': return 'bg-red-500/20 text-red-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-gray-500/20 text-gray-400';
  }
};

const getHostGlowColor = (host: Host): 'primary' | 'accent' | 'neon-blue' | 'neon-cyan' | 'neon-purple' => {
  switch (host.state) {
    case 'CONNECTED': return 'primary';
    case 'ERROR': return 'neon-cyan';
    default: return 'accent';
  }
};

// Network diagram colors
const getHostNodeColor = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return '#10b981'; // green-500
    case 'DISCONNECTED': return '#ef4444'; // red-500
    case 'ERROR': return '#dc2626'; // red-600
    default: return '#6b7280'; // gray-500
  }
};

const getHostStrokeColor = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return '#34d399'; // green-400
    case 'DISCONNECTED': return '#f87171'; // red-400
    case 'ERROR': return '#ef4444'; // red-500
    default: return '#9ca3af'; // gray-400
  }
};

// VM color functions
const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-400';
    case 'STOPPED': return 'bg-red-400';
    case 'PAUSED': return 'bg-yellow-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-gray-400';
  }
};

const getVMGlowColor = (vm: VirtualMachine): 'primary' | 'accent' | 'neon-blue' | 'neon-cyan' | 'neon-purple' => {
  switch (vm.state) {
    case 'ACTIVE': return 'accent';
    case 'ERROR': return 'neon-cyan';
    default: return 'primary';
  }
};

const getVMNodeColor = (vm: VirtualMachine): string => {
  switch (vm.state) {
    case 'ACTIVE': return '#10b981'; // green-500
    case 'STOPPED': return '#ef4444'; // red-500
    case 'PAUSED': return '#f59e0b'; // yellow-500
    case 'ERROR': return '#dc2626'; // red-600
    default: return '#6b7280'; // gray-500
  }
};

const getVMStrokeColor = (vm: VirtualMachine): string => {
  switch (vm.state) {
    case 'ACTIVE': return '#34d399'; // green-400
    case 'STOPPED': return '#f87171'; // red-400
    case 'PAUSED': return '#fbbf24'; // yellow-400
    case 'ERROR': return '#ef4444'; // red-500
    default: return '#9ca3af'; // gray-400
  }
};

const getVMConnectionColor = (vm: VirtualMachine): string => {
  switch (vm.state) {
    case 'ACTIVE': return '#34d399'; // green-400
    case 'STOPPED': return '#f87171'; // red-400
    case 'PAUSED': return '#fbbf24'; // yellow-400
    default: return '#9ca3af'; // gray-400
  }
};

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Lifecycle
onMounted(() => {
  refreshTopology();
});
</script>