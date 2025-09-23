<template>
  <FModal :show="show" @close="handleClose" size="full">
    <div class="space-y-4">
      <!-- Header -->
      <div class="border-b border-gray-700 pb-4">
        <h3 class="text-lg font-semibold text-white">
          VM Hardware Configuration - {{ vmName }}
        </h3>
      </div>

      <!-- Body -->
      <div class="flex h-[80vh]">
        <!-- Sidebar with Hardware Categories -->
        <div class="w-64 bg-gray-900/50 border-r border-gray-700 p-4 overflow-y-auto">
          <div class="space-y-2">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="{
                'bg-blue-500/20 text-blue-400 border border-blue-500/30': activeTab === tab.id,
                'hover:bg-gray-800 text-gray-300': activeTab !== tab.id
              }"
              class="w-full flex items-center justify-between p-3 rounded-lg transition-colors"
            >
              <div class="flex items-center space-x-3">
                <div 
                  :class="{
                    'bg-blue-500': tab.id === 'overview',
                    'bg-green-500': tab.id === 'storage',
                    'bg-purple-500': tab.id === 'network',
                    'bg-orange-500': tab.id === 'video',
                    'bg-red-500': tab.id === 'advanced'
                  }"
                  class="w-3 h-3 rounded-full"
                ></div>
                <span class="font-medium">{{ tab.name }}</span>
              </div>
              <span 
                v-if="getTabDataCount(tab.id) > 0" 
                class="text-xs bg-gray-700 px-2 py-1 rounded"
              >
                {{ getTabDataCount(tab.id) }}
              </span>
            </button>
          </div>

          <!-- Actions -->
          <div class="mt-6 pt-6 border-t border-gray-700 space-y-2">
            <button
              @click="refreshData"
              :disabled="isLoading"
              class="w-full p-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded text-sm transition-colors disabled:opacity-50"
            >
              {{ isLoading ? 'Refreshing...' : 'Refresh Data' }}
            </button>
            <button
              @click="exportConfiguration"
              class="w-full p-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm transition-colors"
            >
              Export Config
            </button>
          </div>
        </div>

        <!-- Main Content Area -->
        <div class="flex-1 p-6 overflow-y-auto">
          <!-- Loading State -->
          <div v-if="isLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
              <p class="text-gray-300">Loading extended hardware configuration...</p>
              <p class="text-sm text-gray-500">Fetching data from {{ Object.keys(allHardwareEntities).length }}+ database entities</p>
            </div>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-lg max-w-md">
                <p class="text-red-400 mb-4">{{ error }}</p>
                <button
                  @click="loadExtendedHardwareConfig"
                  class="px-4 py-2 bg-red-500/20 hover:bg-red-500/30 text-red-400 rounded transition-colors"
                >
                  Retry
                </button>
              </div>
            </div>
          </div>

          <!-- Content -->
          <div v-else>
            <!-- Data Summary -->
            <div class="mb-6 p-4 bg-gray-800/50 rounded-lg border border-gray-700">
              <h4 class="text-sm font-medium text-white mb-2">Hardware Configuration Summary</h4>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                <div>
                  <span class="text-gray-400">Database Entities:</span>
                  <span class="text-white ml-1">{{ Object.keys(allHardwareEntities).length }}</span>
                </div>
                <div>
                  <span class="text-gray-400">Storage Devices:</span>
                  <span class="text-white ml-1">{{ (extendedHardware?.disk_attachments?.length || 0) + (extendedHardware?.filesystem_attachments?.length || 0) }}</span>
                </div>
                <div>
                  <span class="text-gray-400">Network Interfaces:</span>
                  <span class="text-white ml-1">{{ extendedHardware?.port_attachments?.length || 0 }}</span>
                </div>
                <div>
                  <span class="text-gray-400">Last Updated:</span>
                  <span class="text-white ml-1">{{ formatTimestamp(extendedHardware?.vm_info?.updatedAt) }}</span>
                </div>
              </div>
            </div>

            <!-- Tab Content -->
            <div>
              <!-- Overview Tab -->
              <VMOverviewPanel
                v-if="activeTab === 'overview'"
                :vm-info="extendedHardware?.vm_info || {}"
                :cpu-topology="extendedHardware?.cpu_topology"
                :cpu-features="extendedHardware?.cpu_features || []"
                :memory-configs="extendedHardware?.memory_configs || []"
              />

              <!-- Storage Tab -->
              <VMStoragePanel
                v-if="activeTab === 'storage'"
                :disk-attachments="extendedHardware?.disk_attachments || []"
                :filesystem-attachments="extendedHardware?.filesystem_attachments || []"
              />

              <!-- Network Tab -->
              <VMNetworkPanel
                v-if="activeTab === 'network'"
                :port-attachments="extendedHardware?.port_attachments || []"
                :network-stats="extendedHardware?.network_stats"
              />

              <!-- Video Tab -->
              <VMVideoPanel
                v-if="activeTab === 'video'"
                :video-attachments="extendedHardware?.video_attachments || []"
                :sound-attachments="extendedHardware?.sound_attachments || []"
                :input-attachments="extendedHardware?.input_attachments || []"
                :graphics-type="extendedHardware?.vm_info?.graphics_type"
                :graphics-port="extendedHardware?.vm_info?.graphics_port"
                :graphics-listen="extendedHardware?.vm_info?.graphics_listen"
              />

              <!-- Advanced Tab -->
              <VMAdvancedPanel
                v-if="activeTab === 'advanced'"
                :controller-attachments="extendedHardware?.controller_attachments || []"
                :host-device-attachments="extendedHardware?.host_device_attachments || []"
                :tpm-attachments="extendedHardware?.tpm_attachments || []"
                :watchdog-attachments="extendedHardware?.watchdog_attachments || []"
                :rng-attachments="extendedHardware?.rng_attachments || []"
              />
            </div>

            <!-- Debug Information (Development Mode) -->
            <div v-if="showDebugInfo" class="mt-8 p-4 bg-gray-900 rounded-lg border border-gray-700">
              <h4 class="text-sm font-medium text-white mb-3">Debug Information</h4>
              <div class="space-y-2 text-xs">
                <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
                  <div v-for="(count, entity) in allHardwareEntities" :key="entity" class="flex justify-between">
                    <span class="text-gray-400">{{ entity }}:</span>
                    <span class="text-white">{{ count }}</span>
                  </div>
                </div>
              </div>
              <button
                @click="showRawData = !showRawData"
                class="mt-3 text-xs px-2 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded"
              >
                {{ showRawData ? 'Hide' : 'Show' }} Raw Data
              </button>
              <div v-if="showRawData" class="mt-3 p-3 bg-black rounded border max-h-64 overflow-y-auto">
                <pre class="text-xs text-green-400">{{ JSON.stringify(extendedHardware, null, 2) }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="border-t border-gray-700 pt-4">
        <div class="flex justify-between items-center">
          <div class="flex items-center space-x-2">
            <button
              @click="showDebugInfo = !showDebugInfo"
              class="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded"
            >
              {{ showDebugInfo ? 'Hide' : 'Show' }} Debug
            </button>
            <span v-if="extendedHardware" class="text-xs text-gray-500">
              {{ Object.keys(allHardwareEntities).length }} entities loaded
            </span>
          </div>
          <div class="flex space-x-3">
            <button
              @click="handleClose"
              class="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </FModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import FModal from '../ui/FModal.vue';
import VMOverviewPanel from '../vm/VMOverviewPanel.vue';
import VMStoragePanel from '../vm/VMStoragePanel.vue';
import VMNetworkPanel from '../vm/VMNetworkPanel.vue';
import VMVideoPanel from '../vm/VMVideoPanel.vue';
import VMAdvancedPanel from '../vm/VMAdvancedPanel.vue';
import { vmApi } from '../../services/api';
import { errorRecoveryService } from '../../services/errorRecovery';

interface Props {
  show: boolean;
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
  hardwareUpdated: [];
}>();

// State
const isLoading = ref(false);
const error = ref<string | null>(null);
const extendedHardware = ref<any>(null);
const activeTab = ref('overview');
const showDebugInfo = ref(false);
const showRawData = ref(false);

// Tabs configuration
const tabs = [
  { id: 'overview', name: 'Overview' },
  { id: 'storage', name: 'Storage' },
  { id: 'network', name: 'Network' },
  { id: 'video', name: 'Video & Audio' },
  { id: 'advanced', name: 'Advanced' }
];

// Computed properties
const allHardwareEntities = computed(() => {
  if (!extendedHardware.value) return {};
  
  const entities: Record<string, number> = {};
  
  // Count all the different types of hardware entities
  if (extendedHardware.value.vm_info) entities['VM Info'] = 1;
  if (extendedHardware.value.cpu_topology) entities['CPU Topology'] = 1;
  entities['CPU Features'] = extendedHardware.value.cpu_features?.length || 0;
  entities['Memory Configs'] = extendedHardware.value.memory_configs?.length || 0;
  entities['Disk Attachments'] = extendedHardware.value.disk_attachments?.length || 0;
  entities['Port Attachments'] = extendedHardware.value.port_attachments?.length || 0;
  entities['Video Attachments'] = extendedHardware.value.video_attachments?.length || 0;
  entities['Controller Attachments'] = extendedHardware.value.controller_attachments?.length || 0;
  entities['Host Device Attachments'] = extendedHardware.value.host_device_attachments?.length || 0;
  entities['TPM Attachments'] = extendedHardware.value.tpm_attachments?.length || 0;
  entities['Watchdog Attachments'] = extendedHardware.value.watchdog_attachments?.length || 0;
  entities['Serial Device Attachments'] = extendedHardware.value.serial_device_attachments?.length || 0;
  entities['Channel Attachments'] = extendedHardware.value.channel_attachments?.length || 0;
  entities['Filesystem Attachments'] = extendedHardware.value.filesystem_attachments?.length || 0;
  entities['Sound Attachments'] = extendedHardware.value.sound_attachments?.length || 0;
  entities['Input Attachments'] = extendedHardware.value.input_attachments?.length || 0;
  entities['RNG Attachments'] = extendedHardware.value.rng_attachments?.length || 0;
  entities['Memory Balloon Attachments'] = extendedHardware.value.memory_balloon_attachments?.length || 0;
  entities['VSock Attachments'] = extendedHardware.value.vsock_attachments?.length || 0;
  if (extendedHardware.value.boot_config) entities['Boot Config'] = 1;
  entities['Security Labels'] = extendedHardware.value.security_labels?.length || 0;
  
  return entities;
});

const getTabDataCount = (tabId: string): number => {
  if (!extendedHardware.value) return 0;
  
  switch (tabId) {
    case 'overview':
      return (extendedHardware.value.cpu_features?.length || 0) + 
             (extendedHardware.value.memory_configs?.length || 0) + 1;
    case 'storage':
      return (extendedHardware.value.disk_attachments?.length || 0) + 
             (extendedHardware.value.filesystem_attachments?.length || 0);
    case 'network':
      return extendedHardware.value.port_attachments?.length || 0;
    case 'video':
      return (extendedHardware.value.video_attachments?.length || 0) + 
             (extendedHardware.value.sound_attachments?.length || 0) + 
             (extendedHardware.value.input_attachments?.length || 0);
    case 'advanced':
      return (extendedHardware.value.controller_attachments?.length || 0) + 
             (extendedHardware.value.host_device_attachments?.length || 0) + 
             (extendedHardware.value.tpm_attachments?.length || 0) + 
             (extendedHardware.value.watchdog_attachments?.length || 0) + 
             (extendedHardware.value.rng_attachments?.length || 0);
    default:
      return 0;
  }
};

// Methods
const loadExtendedHardwareConfig = async () => {
  try {
    isLoading.value = true;
    error.value = null;
    
    const response = await vmApi.getExtendedVMHardware(props.hostId, props.vmName);
    extendedHardware.value = response;
    
    emit('hardwareUpdated');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load extended hardware configuration';
    errorRecoveryService.addError(
      err as Error,
      `load_vm_extended_hardware_${props.vmName}`,
      { hostId: props.hostId, vmName: props.vmName }
    );
  } finally {
    isLoading.value = false;
  }
};

const refreshData = () => {
  loadExtendedHardwareConfig();
};

const exportConfiguration = () => {
  if (!extendedHardware.value) return;
  
  const config = {
    vmName: props.vmName,
    hostId: props.hostId,
    timestamp: new Date().toISOString(),
    hardwareData: extendedHardware.value,
    summary: allHardwareEntities.value
  };
  
  const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${props.vmName}-extended-hardware-config.json`;
  a.click();
  URL.revokeObjectURL(url);
};

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};

const handleClose = () => {
  emit('close');
};

// Watchers
watch(() => props.show, (show) => {
  if (show) {
    loadExtendedHardwareConfig();
  }
});

// Lifecycle
onMounted(() => {
  if (props.show) {
    loadExtendedHardwareConfig();
  }
});
</script>