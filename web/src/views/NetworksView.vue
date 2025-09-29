<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Networks</h1>
        <p class="text-slate-400 mt-2">Libvirt networks, ports, and virtual network infrastructure</p>
      </div>
      <div class="flex items-center gap-4">
        <FButton
          variant="ghost"
          size="sm"
          @click="refreshNetworks"
          :disabled="loading"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          {{ loading ? 'Refreshing...' : 'Refresh' }}
        </FButton>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- Networks -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Networks</h3>
          <p class="text-2xl font-bold text-cyan-400">{{ stats.networks }}</p>
          <p class="text-xs text-slate-500 mt-1">Virtual networks</p>
        </div>
      </FCard>

      <!-- Ports -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-teal-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Ports</h3>
          <p class="text-2xl font-bold text-green-400">{{ stats.ports }}</p>
          <p class="text-xs text-slate-500 mt-1">Network ports</p>
        </div>
      </FCard>

      <!-- Port Attachments -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM21 5a2 2 0 00-2-2h-4a2 2 0 00-2 2v12a4 4 0 004 4h4a4 4 0 004-4V5z"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Attachments</h3>
          <p class="text-2xl font-bold text-purple-400">{{ stats.attachments }}</p>
          <p class="text-xs text-slate-500 mt-1">Active connections</p>
        </div>
      </FCard>

      <!-- Active Connections -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Bandwidth</h3>
          <p class="text-2xl font-bold text-orange-400">{{ formatBytes(stats.totalBandwidth) }}/s</p>
          <p class="text-xs text-slate-500 mt-1">Total throughput</p>
        </div>
      </FCard>
    </div>

    <!-- Networks Grid -->
    <div class="space-y-6">
      <div v-if="networks.length === 0" class="text-center py-12">
        <svg class="w-16 h-16 text-slate-600 mx-auto mb-4" fill="currentColor" viewBox="0 0 20 20">
          <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
        </svg>
        <p class="text-slate-400 text-lg">No networks found</p>
        <p class="text-slate-500 text-sm mt-2">Networks will appear here once hosts are connected and synced</p>
      </div>

      <div v-for="network in networks" :key="network.id" class="space-y-4">
        <!-- Network Card -->
        <FCard class="card-glow">
          <div class="p-6">
            <!-- Network Header -->
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center shadow-xl">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                  </svg>
                </div>
                <div>
                  <h3 class="text-xl font-bold text-white">{{ network.name }}</h3>
                  <p class="text-sm text-slate-400">{{ network.mode }} network on {{ getHostDisplayName(network.host_id) }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <span :class="[
                  'px-3 py-1 rounded-full text-xs font-medium',
                  network.mode === 'bridge' ? 'bg-green-500/20 text-green-400' :
                  network.mode === 'nat' ? 'bg-blue-500/20 text-blue-400' :
                  network.mode === 'isolated' ? 'bg-yellow-500/20 text-yellow-400' :
                  'bg-gray-500/20 text-gray-400'
                ]">
                  {{ network.mode }}
                </span>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click="toggleNetworkExpanded(network.id)"
                  class="text-xs"
                >
                  {{ expandedNetworks.has(network.id) ? 'Collapse' : 'Expand' }}
                  <svg 
                    class="w-4 h-4 ml-2 transition-transform duration-200"
                    :class="{ 'rotate-180': expandedNetworks.has(network.id) }"
                    fill="none" 
                    stroke="currentColor" 
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                  </svg>
                </FButton>
              </div>
            </div>

            <!-- Network Basic Info -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div class="space-y-2">
                <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">UUID</span>
                <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                  <span class="text-slate-300 text-sm font-mono">{{ network.uuid || 'N/A' }}</span>
                </div>
              </div>
              <div class="space-y-2">
                <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Bridge Name</span>
                <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                  <span class="text-white font-mono">{{ network.bridge_name || 'N/A' }}</span>
                </div>
              </div>
              <div class="space-y-2">
                <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Connected Ports</span>
                <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                  <span class="text-blue-400 font-bold">{{ getNetworkPorts(network.id).length }}</span>
                </div>
              </div>
            </div>

            <!-- Expanded Network Details -->
            <div v-if="expandedNetworks.has(network.id)" class="mt-6 pt-6 border-t border-slate-700/50">
              <!-- Ports Section -->
              <div class="space-y-4">
                <h4 class="text-lg font-semibold text-white flex items-center gap-2">
                  <svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                  </svg>
                  Network Ports
                </h4>

                <div class="grid grid-cols-1 gap-4">
                  <div v-if="getNetworkPorts(network.id).length === 0" class="text-center py-8">
                    <svg class="w-12 h-12 text-slate-600 mx-auto mb-4" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                      <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
                    </svg>
                    <p class="text-slate-400">No ports on this network</p>
                  </div>

                  <div v-for="port in getNetworkPorts(network.id)" :key="port.id" class="border border-slate-700/50 rounded-lg">
                    <!-- Port Header -->
                    <div class="p-4 bg-slate-800/30 border-b border-slate-700/50">
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                          <div class="w-8 h-8 rounded-lg bg-green-500/20 flex items-center justify-center">
                            <svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                            </svg>
                          </div>
                          <div>
                            <h5 class="font-semibold text-white">Port {{ port.id }}</h5>
                            <p class="text-sm text-slate-400">MAC: {{ port.mac_address || 'Auto-generated' }}</p>
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          <span class="text-xs px-2 py-1 rounded-full bg-green-500/20 text-green-400">
                            {{ port.model_name || 'virtio' }}
                          </span>
                          <FButton
                            variant="ghost"
                            size="xs"
                            @click="togglePortExpanded(port.id)"
                            class="p-1"
                          >
                            <svg 
                              class="w-3 h-3 transition-transform duration-200"
                              :class="{ 'rotate-180': expandedPorts.has(port.id) }"
                              fill="none" 
                              stroke="currentColor" 
                              viewBox="0 0 24 24"
                            >
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                            </svg>
                          </FButton>
                        </div>
                      </div>
                    </div>

                    <!-- Port Details -->
                    <div v-if="expandedPorts.has(port.id)" class="p-4 space-y-4">
                      <!-- Port Configuration -->
                      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <h6 class="text-sm font-medium text-slate-300 mb-3">Port Configuration</h6>
                          <div class="space-y-2 text-sm">
                            <div class="flex justify-between">
                              <span class="text-slate-400">MAC Address:</span>
                              <span class="text-white font-mono">{{ port.mac_address || 'Auto' }}</span>
                            </div>
                            <div class="flex justify-between">
                              <span class="text-slate-400">Model:</span>
                              <span class="text-white">{{ port.model_name || 'virtio' }}</span>
                            </div>
                            <div class="flex justify-between">
                              <span class="text-slate-400">Port Group:</span>
                              <span class="text-white">{{ port.port_group || 'None' }}</span>
                            </div>
                          </div>
                        </div>
                        <div>
                          <h6 class="text-sm font-medium text-slate-300 mb-3">Port Statistics</h6>
                          <div class="space-y-2 text-sm">
                            <div class="flex justify-between">
                              <span class="text-slate-400">RX Bytes:</span>
                              <span class="text-white">{{ formatBytes(port.rx_bytes || 0) }}</span>
                            </div>
                            <div class="flex justify-between">
                              <span class="text-slate-400">TX Bytes:</span>
                              <span class="text-white">{{ formatBytes(port.tx_bytes || 0) }}</span>
                            </div>
                            <div class="flex justify-between">
                              <span class="text-slate-400">Status:</span>
                              <span class="text-green-400">Active</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      <!-- Port Attachments -->
                      <div>
                        <h6 class="text-sm font-medium text-slate-300 mb-3">VM Attachments</h6>
                        <div class="space-y-2">
                          <div v-if="getPortAttachments(port.id).length === 0" class="text-center py-4 text-slate-400 text-sm">
                            No VM attachments
                          </div>
                          <div v-for="attachment in getPortAttachments(port.id)" :key="attachment.id" 
                               class="flex items-center justify-between p-3 bg-slate-900/50 rounded-lg border border-slate-700/50">
                            <div class="flex items-center gap-3">
                              <div class="w-6 h-6 rounded-lg bg-blue-500/20 flex items-center justify-center">
                                <svg class="w-3 h-3 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                                </svg>
                              </div>
                              <div>
                                <p class="text-sm font-medium text-white">{{ getVMName(attachment.vm_uuid) }}</p>
                                <p class="text-xs text-slate-400">Device: {{ attachment.device_name || 'Default' }}</p>
                              </div>
                            </div>
                            <div class="text-right">
                              <span class="text-xs px-2 py-1 rounded-full bg-blue-500/20 text-blue-400">
                                Attached
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </FCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';

// Component state
const loading = ref(false);
const networks = ref<any[]>([]);
const ports = ref<any[]>([]);
const portAttachments = ref<any[]>([]);
const vms = ref<any[]>([]);
const expandedNetworks = ref(new Set<number>());
const expandedPorts = ref(new Set<number>());

const stats = computed(() => ({
  networks: networks.value.length,
  ports: ports.value.length,
  attachments: portAttachments.value.length,
  totalBandwidth: ports.value.reduce((total, port) => total + (port.rx_bytes || 0) + (port.tx_bytes || 0), 0)
}));

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

const getHostDisplayName = (hostId: string): string => {
  return hostId || 'Unknown Host';
};

const getNetworkPorts = (networkId: number) => {
  return ports.value.filter(port => port.network_id === networkId);
};

const getPortAttachments = (portId: number) => {
  return portAttachments.value.filter(att => att.port_id === portId);
};

const getVMName = (vmUuid: string): string => {
  const vm = vms.value.find(v => v.uuid === vmUuid);
  return vm ? vm.name : 'Unknown VM';
};

const toggleNetworkExpanded = (networkId: number): void => {
  if (expandedNetworks.value.has(networkId)) {
    expandedNetworks.value.delete(networkId);
  } else {
    expandedNetworks.value.add(networkId);
  }
};

const togglePortExpanded = (portId: number): void => {
  if (expandedPorts.value.has(portId)) {
    expandedPorts.value.delete(portId);
  } else {
    expandedPorts.value.add(portId);
  }
};

// Load networks data
const loadNetworksData = async (): Promise<void> => {
  loading.value = true;
  
  try {
    // Fetch networks data from API endpoints
    // Note: These endpoints would need to be implemented in the backend
    try {
      const networksResponse = await fetch('/api/v1/networks');
      if (networksResponse.ok) {
        networks.value = await networksResponse.json();
      } else {
        console.warn('Networks API returned error');
        networks.value = [];
      }
    } catch {
      console.warn('Networks API not available, using placeholder data');
      networks.value = [];
    }

    try {
      const portsResponse = await fetch('/api/v1/ports');
      if (portsResponse.ok) {
        ports.value = await portsResponse.json();
      } else {
        console.warn('Ports API returned error');
        ports.value = [];
      }
    } catch {
      console.warn('Ports API not available, using placeholder data');
      ports.value = [];
    }

    try {
      const attachmentsResponse = await fetch('/api/v1/port-attachments');
      if (attachmentsResponse.ok) {
        portAttachments.value = await attachmentsResponse.json();
      } else {
        console.warn('Port attachments API returned error');
        portAttachments.value = [];
      }
    } catch {
      console.warn('Port attachments API not available, using placeholder data');
      portAttachments.value = [];
    }

    // Also fetch VMs for display names
    try {
      const vmResponse = await fetch('/api/v1/vms');
      if (vmResponse.ok) {
        vms.value = await vmResponse.json();
      }
    } catch (error) {
      console.warn('Failed to fetch VMs:', error);
    }
    
  } catch (error) {
    console.error('Failed to load networks data:', error);
    networks.value = [];
    ports.value = [];
    portAttachments.value = [];
  } finally {
    loading.value = false;
  }
};

// Refresh networks data
const refreshNetworks = (): void => {
  loadNetworksData();
};

// Lifecycle
onMounted(() => {
  loadNetworksData();
});
</script>