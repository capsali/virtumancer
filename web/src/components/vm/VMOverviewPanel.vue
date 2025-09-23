<template>
  <div class="space-y-6">
    <!-- VM Information Overview -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Virtual Machine Information</h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Name</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ vmInfo.name }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">UUID</label>
          <p class="text-white bg-gray-800 p-2 rounded font-mono text-xs">{{ vmInfo.uuid }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Domain UUID</label>
          <p class="text-white bg-gray-800 p-2 rounded font-mono text-xs">{{ vmInfo.domainUuid }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">State</label>
          <span 
            :class="{
              'bg-green-500/20 text-green-400': vmInfo.state === 'ACTIVE',
              'bg-red-500/20 text-red-400': vmInfo.state === 'STOPPED',
              'bg-yellow-500/20 text-yellow-400': vmInfo.state === 'PAUSED',
              'bg-gray-500/20 text-gray-400': vmInfo.state === 'UNKNOWN'
            }"
            class="px-2 py-1 rounded text-sm font-medium"
          >
            {{ vmInfo.state }}
          </span>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Source</label>
          <span 
            :class="{
              'bg-blue-500/20 text-blue-400': vmInfo.source === 'managed',
              'bg-purple-500/20 text-purple-400': vmInfo.source === 'imported'
            }"
            class="px-2 py-1 rounded text-sm font-medium"
          >
            {{ vmInfo.source }}
          </span>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Sync Status</label>
          <span 
            :class="{
              'bg-green-500/20 text-green-400': vmInfo.syncStatus === 'SYNCED',
              'bg-orange-500/20 text-orange-400': vmInfo.syncStatus === 'DRIFTED',
              'bg-gray-500/20 text-gray-400': vmInfo.syncStatus === 'UNKNOWN'
            }"
            class="px-2 py-1 rounded text-sm font-medium"
          >
            {{ vmInfo.syncStatus }}
          </span>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Title</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ vmInfo.title || 'N/A' }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">OS Type</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ vmInfo.osType || 'N/A' }}</p>
        </div>
      </div>

      <div class="mt-4">
        <label class="block text-sm font-medium text-gray-300 mb-1">Description</label>
        <p class="text-white bg-gray-800 p-3 rounded min-h-[3rem]">{{ vmInfo.description || 'No description provided' }}</p>
      </div>

      <div class="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <label class="block text-xs font-medium text-gray-400 mb-1">Created</label>
          <p class="text-gray-300">{{ formatDate(vmInfo.createdAt) }}</p>
        </div>
        <div>
          <label class="block text-xs font-medium text-gray-400 mb-1">Updated</label>
          <p class="text-gray-300">{{ formatDate(vmInfo.updatedAt) }}</p>
        </div>
        <div>
          <label class="block text-xs font-medium text-gray-400 mb-1">Is Template</label>
          <p class="text-gray-300">{{ vmInfo.isTemplate ? 'Yes' : 'No' }}</p>
        </div>
        <div>
          <label class="block text-xs font-medium text-gray-400 mb-1">Needs Rebuild</label>
          <p :class="vmInfo.needsRebuild ? 'text-orange-400' : 'text-gray-300'">
            {{ vmInfo.needsRebuild ? 'Yes' : 'No' }}
          </p>
        </div>
      </div>
    </FCard>

    <!-- CPU Configuration -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">CPU Configuration</h3>
      
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">vCPU Count</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ vmInfo.vcpuCount }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">CPU Model</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ vmInfo.cpuModel || 'Default' }}</p>
        </div>
        <div v-if="cpuTopology">
          <label class="block text-sm font-medium text-gray-300 mb-1">Topology</label>
          <p class="text-white bg-gray-800 p-2 rounded">
            {{ cpuTopology.sockets }}S/{{ cpuTopology.cores }}C/{{ cpuTopology.threads }}T
          </p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">CPU Features</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ cpuFeatures.length }} features</p>
        </div>
      </div>

      <!-- CPU Features List -->
      <div v-if="cpuFeatures.length > 0">
        <h4 class="text-md font-medium text-white mb-3">CPU Features</h4>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 max-h-48 overflow-y-auto">
          <div
            v-for="feature in cpuFeatures"
            :key="feature.name"
            class="flex items-center justify-between p-2 bg-gray-800 rounded text-sm"
          >
            <span class="text-white">{{ feature.name }}</span>
            <span 
              :class="{
                'text-green-400': feature.policy === 'require',
                'text-blue-400': feature.policy === 'optional',
                'text-red-400': feature.policy === 'disable',
                'text-orange-400': feature.policy === 'forbid'
              }"
              class="text-xs font-medium"
            >
              {{ feature.policy }}
            </span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Memory Configuration -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Memory Configuration</h3>
      
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Max Memory</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ formatMemory(vmInfo.memoryBytes) }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Current Memory</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ formatMemory(vmInfo.currentMemory) }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Memory Configs</label>
          <p class="text-white bg-gray-800 p-2 rounded">{{ memoryConfigs.length }} configurations</p>
        </div>
      </div>

      <!-- Memory Configurations -->
      <div v-if="memoryConfigs.length > 0">
        <h4 class="text-md font-medium text-white mb-3">Memory Configurations</h4>
        <div class="space-y-3">
          <div
            v-for="config in memoryConfigs"
            :key="config.id"
            class="p-3 bg-gray-800 rounded"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="text-white font-medium">{{ config.configType }}</span>
              <span class="text-gray-400 text-sm">{{ config.mode || 'N/A' }}</span>
            </div>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
              <div v-if="config.sourceType">
                <span class="text-gray-400">Source:</span>
                <span class="text-white ml-1">{{ config.sourceType }}</span>
              </div>
              <div v-if="config.sizeKB">
                <span class="text-gray-400">Size:</span>
                <span class="text-white ml-1">{{ Math.round(config.sizeKB / 1024) }} MB</span>
              </div>
              <div>
                <span class="text-gray-400">Locked:</span>
                <span class="text-white ml-1">{{ config.locked ? 'Yes' : 'No' }}</span>
              </div>
              <div>
                <span class="text-gray-400">No Share:</span>
                <span class="text-white ml-1">{{ config.nosharepages ? 'Yes' : 'No' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Timestamps -->
    <div class="grid grid-cols-2 gap-4 text-xs text-gray-500">
      <div>
        <span class="font-medium">Created:</span> {{ formatTimestamp(vmInfo.createdAt) }}
      </div>
      <div>
        <span class="font-medium">Last Updated:</span> {{ formatTimestamp(vmInfo.updatedAt) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import FCard from '../ui/FCard.vue';

interface Props {
  vmInfo: any;
  cpuTopology: any;
  cpuFeatures: any[];
  memoryConfigs: any[];
}

const props = defineProps<Props>();

const formatMemory = (bytes: number) => {
  if (!bytes) return 'N/A';
  const mb = Math.round(bytes / (1024 * 1024));
  const gb = Math.round(mb / 1024);
  
  if (gb >= 1) {
    return `${gb} GB`;
  }
  return `${mb} MB`;
};

const formatDate = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleDateString();
};

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};
</script>