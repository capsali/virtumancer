<template>
  <div class="space-y-6">
    <!-- Storage Overview -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Storage Devices</h3>
      
      <div class="mb-4 text-sm text-gray-400">
        Total Devices: {{ diskAttachments.length }}
      </div>

      <div v-if="diskAttachments.length === 0" class="text-center py-8 text-gray-400">
        No storage devices configured
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="disk in diskAttachments"
          :key="disk.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <!-- Disk Header -->
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-blue-500': disk.deviceType === 'disk',
                  'bg-purple-500': disk.deviceType === 'cdrom',
                  'bg-green-500': disk.deviceType === 'floppy',
                  'bg-gray-500': !disk.deviceType
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">{{ disk.target || 'Unnamed Device' }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ disk.deviceType || 'unknown' }}
              </span>
            </div>
            <div class="flex items-center space-x-2">
              <span 
                :class="{
                  'text-green-400': disk.readonly === false,
                  'text-orange-400': disk.readonly === true
                }"
                class="text-xs"
              >
                {{ disk.readonly ? 'Read-Only' : 'Read-Write' }}
              </span>
              <span 
                :class="{
                  'text-green-400': disk.shareable === true,
                  'text-gray-400': disk.shareable === false
                }"
                class="text-xs"
              >
                {{ disk.shareable ? 'Shareable' : 'Exclusive' }}
              </span>
            </div>
          </div>

          <!-- Disk Details Grid -->
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 mb-3">
            <div v-if="disk.source">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source</label>
              <p class="text-white text-sm bg-gray-900 p-2 rounded font-mono text-xs">{{ disk.source }}</p>
            </div>
            <div v-if="disk.sourceType">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source Type</label>
              <p class="text-white text-sm">{{ disk.sourceType }}</p>
            </div>
            <div v-if="disk.bus">
              <label class="block text-xs font-medium text-gray-400 mb-1">Bus</label>
              <p class="text-white text-sm">{{ disk.bus }}</p>
            </div>
            <div v-if="disk.bootOrder">
              <label class="block text-xs font-medium text-gray-400 mb-1">Boot Order</label>
              <p class="text-white text-sm">{{ disk.bootOrder }}</p>
            </div>
          </div>

          <!-- Driver Information -->
          <div v-if="disk.driverName || disk.driverType || disk.driverCache" class="mb-3">
            <label class="block text-xs font-medium text-gray-400 mb-2">Driver Configuration</label>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
              <div v-if="disk.driverName">
                <span class="text-gray-400 text-xs">Name:</span>
                <span class="text-white text-sm ml-1">{{ disk.driverName }}</span>
              </div>
              <div v-if="disk.driverType">
                <span class="text-gray-400 text-xs">Type:</span>
                <span class="text-white text-sm ml-1">{{ disk.driverType }}</span>
              </div>
              <div v-if="disk.driverCache">
                <span class="text-gray-400 text-xs">Cache:</span>
                <span class="text-white text-sm ml-1">{{ disk.driverCache }}</span>
              </div>
            </div>
          </div>

          <!-- Additional Properties -->
          <div class="flex flex-wrap gap-2 text-xs">
            <span v-if="disk.removable" class="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              Removable
            </span>
            <span v-if="disk.encrypted" class="px-2 py-1 bg-green-500/20 text-green-400 rounded">
              Encrypted
            </span>
            <span v-if="disk.discard" class="px-2 py-1 bg-purple-500/20 text-purple-400 rounded">
              Discard: {{ disk.discard }}
            </span>
            <span v-if="disk.detectZeroes" class="px-2 py-1 bg-yellow-500/20 text-yellow-400 rounded">
              Detect Zeroes: {{ disk.detectZeroes }}
            </span>
          </div>

          <!-- Timestamps -->
          <div v-if="disk.createdAt || disk.updatedAt" class="mt-3 pt-3 border-t border-gray-700 flex justify-between text-xs text-gray-500">
            <span v-if="disk.createdAt">Added: {{ formatTimestamp(disk.createdAt) }}</span>
            <span v-if="disk.updatedAt">Modified: {{ formatTimestamp(disk.updatedAt) }}</span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Filesystem Attachments -->
    <FCard v-if="filesystemAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Filesystem Attachments</h3>
      
      <div class="space-y-4">
        <div
          v-for="fs in filesystemAttachments"
          :key="fs.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-white font-medium">{{ fs.target || 'Unnamed Filesystem' }}</h4>
            <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
              {{ fs.type || 'unknown' }}
            </span>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="fs.source">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source</label>
              <p class="text-white text-sm bg-gray-900 p-2 rounded font-mono text-xs">{{ fs.source }}</p>
            </div>
            <div v-if="fs.accessMode">
              <label class="block text-xs font-medium text-gray-400 mb-1">Access Mode</label>
              <p class="text-white text-sm">{{ fs.accessMode }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-400 mb-1">Read-Only</label>
              <p class="text-white text-sm">{{ fs.readonly ? 'Yes' : 'No' }}</p>
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
  diskAttachments: any[];
  filesystemAttachments: any[];
}

const props = defineProps<Props>();

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};
</script>