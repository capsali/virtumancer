<template>
  <div class="space-y-6">
    <!-- Network Interfaces -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Network Interfaces</h3>
      
      <div class="mb-4 text-sm text-gray-400">
        Total Interfaces: {{ portAttachments.length }}
      </div>

      <div v-if="portAttachments.length === 0" class="text-center py-8 text-gray-400">
        No network interfaces configured
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="port in portAttachments"
          :key="port.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <!-- Interface Header -->
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-green-500': port.interfaceType === 'bridge',
                  'bg-blue-500': port.interfaceType === 'network',
                  'bg-purple-500': port.interfaceType === 'direct',
                  'bg-orange-500': port.interfaceType === 'hostdev',
                  'bg-gray-500': !port.interfaceType
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">{{ port.target || 'Unnamed Interface' }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ port.interfaceType || 'unknown' }}
              </span>
            </div>
            <div v-if="port.linkState" class="flex items-center space-x-2">
              <span 
                :class="{
                  'text-green-400': port.linkState === 'up',
                  'text-red-400': port.linkState === 'down'
                }"
                class="text-xs"
              >
                {{ port.linkState }}
              </span>
            </div>
          </div>

          <!-- Interface Details Grid -->
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 mb-3">
            <div v-if="port.source">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source</label>
              <p class="text-white text-sm bg-gray-900 p-2 rounded">{{ port.source }}</p>
            </div>
            <div v-if="port.sourceType">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source Type</label>
              <p class="text-white text-sm">{{ port.sourceType }}</p>
            </div>
            <div v-if="port.model">
              <label class="block text-xs font-medium text-gray-400 mb-1">Model</label>
              <p class="text-white text-sm">{{ port.model }}</p>
            </div>
            <div v-if="port.bootOrder">
              <label class="block text-xs font-medium text-gray-400 mb-1">Boot Order</label>
              <p class="text-white text-sm">{{ port.bootOrder }}</p>
            </div>
          </div>

          <!-- MAC Address -->
          <div v-if="port.macAddress" class="mb-3">
            <label class="block text-xs font-medium text-gray-400 mb-1">MAC Address</label>
            <p class="text-white text-sm font-mono bg-gray-900 p-2 rounded">{{ port.macAddress }}</p>
          </div>

          <!-- Driver Information -->
          <div v-if="port.driverName || port.driverQueues" class="mb-3">
            <label class="block text-xs font-medium text-gray-400 mb-2">Driver Configuration</label>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
              <div v-if="port.driverName">
                <span class="text-gray-400 text-xs">Name:</span>
                <span class="text-white text-sm ml-1">{{ port.driverName }}</span>
              </div>
              <div v-if="port.driverQueues">
                <span class="text-gray-400 text-xs">Queues:</span>
                <span class="text-white text-sm ml-1">{{ port.driverQueues }}</span>
              </div>
            </div>
          </div>

          <!-- Backend Configuration -->
          <div v-if="port.backendType || port.backendTxMode || port.backendIOEventFD" class="mb-3">
            <label class="block text-xs font-medium text-gray-400 mb-2">Backend Configuration</label>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
              <div v-if="port.backendType">
                <span class="text-gray-400 text-xs">Type:</span>
                <span class="text-white text-sm ml-1">{{ port.backendType }}</span>
              </div>
              <div v-if="port.backendTxMode">
                <span class="text-gray-400 text-xs">TX Mode:</span>
                <span class="text-white text-sm ml-1">{{ port.backendTxMode }}</span>
              </div>
              <div v-if="port.backendIOEventFD !== undefined">
                <span class="text-gray-400 text-xs">IO Event FD:</span>
                <span class="text-white text-sm ml-1">{{ port.backendIOEventFD ? 'Yes' : 'No' }}</span>
              </div>
            </div>
          </div>

          <!-- Additional Properties -->
          <div class="flex flex-wrap gap-2 text-xs">
            <span v-if="port.trustGuestRxFilters" class="px-2 py-1 bg-green-500/20 text-green-400 rounded">
              Trust Guest RX Filters
            </span>
            <span v-if="port.mtu" class="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              MTU: {{ port.mtu }}
            </span>
            <span v-if="port.alias" class="px-2 py-1 bg-purple-500/20 text-purple-400 rounded">
              Alias: {{ port.alias }}
            </span>
          </div>

          <!-- Timestamps -->
          <div v-if="port.createdAt || port.updatedAt" class="mt-3 pt-3 border-t border-gray-700 flex justify-between text-xs text-gray-500">
            <span v-if="port.createdAt">Added: {{ formatTimestamp(port.createdAt) }}</span>
            <span v-if="port.updatedAt">Modified: {{ formatTimestamp(port.updatedAt) }}</span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Network Statistics (if available) -->
    <FCard v-if="networkStats && networkStats.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Network Statistics</h3>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="stat in networkStats"
          :key="stat.interface"
          class="p-3 bg-gray-800 rounded border border-gray-700"
        >
          <h4 class="text-white font-medium mb-2">{{ stat.interface }}</h4>
          <div class="grid grid-cols-2 gap-2 text-sm">
            <div>
              <span class="text-gray-400">RX Bytes:</span>
              <span class="text-white ml-1">{{ formatBytes(stat.rxBytes) }}</span>
            </div>
            <div>
              <span class="text-gray-400">TX Bytes:</span>
              <span class="text-white ml-1">{{ formatBytes(stat.txBytes) }}</span>
            </div>
            <div>
              <span class="text-gray-400">RX Packets:</span>
              <span class="text-white ml-1">{{ stat.rxPackets?.toLocaleString() || 'N/A' }}</span>
            </div>
            <div>
              <span class="text-gray-400">TX Packets:</span>
              <span class="text-white ml-1">{{ stat.txPackets?.toLocaleString() || 'N/A' }}</span>
            </div>
          </div>
        </div>
      </div>
    </FCard>
  </div>
</template>

<script setup lang="ts">
import FCard from '../ui/FCard.vue';

interface Props {
  portAttachments: any[];
  networkStats?: any[];
}

const props = defineProps<Props>();

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B';
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
};
</script>