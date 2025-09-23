<template>
  <div class="space-y-6">
    <!-- Video Devices -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Video Devices</h3>
      
      <div class="mb-4 text-sm text-gray-400">
        Total Devices: {{ videoAttachments.length }}
      </div>

      <div v-if="videoAttachments.length === 0" class="text-center py-8 text-gray-400">
        No video devices configured
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="video in videoAttachments"
          :key="video.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-purple-500': video.model === 'vga',
                  'bg-blue-500': video.model === 'cirrus',
                  'bg-green-500': video.model === 'vmvga',
                  'bg-orange-500': video.model === 'qxl',
                  'bg-red-500': video.model === 'virtio',
                  'bg-gray-500': !video.model
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">Video Device {{ video.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ video.model || 'unknown' }}
              </span>
            </div>
            <div v-if="video.primary" class="text-xs px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              Primary
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 mb-3">
            <div v-if="video.vram">
              <label class="block text-xs font-medium text-gray-400 mb-1">VRAM</label>
              <p class="text-white text-sm">{{ Math.round(video.vram / 1024) }} MB</p>
            </div>
            <div v-if="video.heads">
              <label class="block text-xs font-medium text-gray-400 mb-1">Heads</label>
              <p class="text-white text-sm">{{ video.heads }}</p>
            </div>
            <div v-if="video.ram">
              <label class="block text-xs font-medium text-gray-400 mb-1">RAM</label>
              <p class="text-white text-sm">{{ Math.round(video.ram / 1024) }} MB</p>
            </div>
            <div v-if="video.vgamem">
              <label class="block text-xs font-medium text-gray-400 mb-1">VGA Memory</label>
              <p class="text-white text-sm">{{ Math.round(video.vgamem / 1024) }} MB</p>
            </div>
          </div>

          <div class="flex flex-wrap gap-2 text-xs">
            <span v-if="video.acceleration3d" class="px-2 py-1 bg-green-500/20 text-green-400 rounded">
              3D Acceleration
            </span>
            <span v-if="video.accelerationAccel2d" class="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              2D Acceleration
            </span>
          </div>

          <div v-if="video.createdAt || video.updatedAt" class="mt-3 pt-3 border-t border-gray-700 flex justify-between text-xs text-gray-500">
            <span v-if="video.createdAt">Added: {{ formatTimestamp(video.createdAt) }}</span>
            <span v-if="video.updatedAt">Modified: {{ formatTimestamp(video.updatedAt) }}</span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Sound Devices -->
    <FCard v-if="soundAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Sound Devices</h3>
      
      <div class="space-y-4">
        <div
          v-for="sound in soundAttachments"
          :key="sound.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div class="w-3 h-3 rounded-full bg-green-500"></div>
              <h4 class="text-white font-medium">Sound Device {{ sound.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ sound.model || 'unknown' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="sound.alias">
              <label class="block text-xs font-medium text-gray-400 mb-1">Alias</label>
              <p class="text-white text-sm">{{ sound.alias }}</p>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Input Devices -->
    <FCard v-if="inputAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Input Devices</h3>
      
      <div class="space-y-4">
        <div
          v-for="input in inputAttachments"
          :key="input.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-blue-500': input.type === 'mouse',
                  'bg-green-500': input.type === 'keyboard',
                  'bg-purple-500': input.type === 'tablet',
                  'bg-gray-500': !input.type
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">{{ input.type || 'Input Device' }} {{ input.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ input.bus || 'unknown' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="input.source">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source</label>
              <p class="text-white text-sm bg-gray-900 p-2 rounded font-mono text-xs">{{ input.source }}</p>
            </div>
            <div v-if="input.alias">
              <label class="block text-xs font-medium text-gray-400 mb-1">Alias</label>
              <p class="text-white text-sm">{{ input.alias }}</p>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Graphics Configuration -->
    <FCard v-if="hasGraphicsConfig" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Graphics Configuration</h3>
      
      <div class="p-4 bg-gray-800 rounded-lg border border-gray-700">
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1">Type</label>
            <p class="text-white text-sm">{{ graphicsType || 'N/A' }}</p>
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1">Port</label>
            <p class="text-white text-sm">{{ graphicsPort || 'Auto' }}</p>
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1">Listen</label>
            <p class="text-white text-sm">{{ graphicsListen || 'Local' }}</p>
          </div>
        </div>
      </div>
    </FCard>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import FCard from '../ui/FCard.vue';

interface Props {
  videoAttachments: any[];
  soundAttachments: any[];
  inputAttachments: any[];
  graphicsType?: string;
  graphicsPort?: string | number;
  graphicsListen?: string;
}

const props = defineProps<Props>();

const hasGraphicsConfig = computed(() => {
  return props.graphicsType || props.graphicsPort || props.graphicsListen;
});

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};
</script>