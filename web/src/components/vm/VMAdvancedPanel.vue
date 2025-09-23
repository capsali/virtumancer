<template>
  <div class="space-y-6">
    <!-- Controllers -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Controllers</h3>
      
      <div class="mb-4 text-sm text-gray-400">
        Total Controllers: {{ controllerAttachments.length }}
      </div>

      <div v-if="controllerAttachments.length === 0" class="text-center py-8 text-gray-400">
        No controllers configured
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="controller in controllerAttachments"
          :key="controller.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-blue-500': controller.type === 'ide',
                  'bg-green-500': controller.type === 'scsi',
                  'bg-purple-500': controller.type === 'virtio-scsi',
                  'bg-orange-500': controller.type === 'usb',
                  'bg-red-500': controller.type === 'sata',
                  'bg-gray-500': !controller.type
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">{{ controller.type || 'Controller' }} {{ controller.index || controller.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ controller.model || 'default' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 mb-3">
            <div v-if="controller.index !== undefined">
              <label class="block text-xs font-medium text-gray-400 mb-1">Index</label>
              <p class="text-white text-sm">{{ controller.index }}</p>
            </div>
            <div v-if="controller.ports">
              <label class="block text-xs font-medium text-gray-400 mb-1">Ports</label>
              <p class="text-white text-sm">{{ controller.ports }}</p>
            </div>
            <div v-if="controller.vectors">
              <label class="block text-xs font-medium text-gray-400 mb-1">Vectors</label>
              <p class="text-white text-sm">{{ controller.vectors }}</p>
            </div>
            <div v-if="controller.queues">
              <label class="block text-xs font-medium text-gray-400 mb-1">Queues</label>
              <p class="text-white text-sm">{{ controller.queues }}</p>
            </div>
          </div>

          <div class="flex flex-wrap gap-2 text-xs mb-3">
            <span v-if="controller.master" class="px-2 py-1 bg-yellow-500/20 text-yellow-400 rounded">
              Master
            </span>
            <span v-if="controller.alias" class="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              Alias: {{ controller.alias }}
            </span>
          </div>

          <div v-if="controller.createdAt || controller.updatedAt" class="mt-3 pt-3 border-t border-gray-700 flex justify-between text-xs text-gray-500">
            <span v-if="controller.createdAt">Added: {{ formatTimestamp(controller.createdAt) }}</span>
            <span v-if="controller.updatedAt">Modified: {{ formatTimestamp(controller.updatedAt) }}</span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Host Devices -->
    <FCard class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Host Device Attachments</h3>
      
      <div class="mb-4 text-sm text-gray-400">
        Total Devices: {{ hostDeviceAttachments.length }}
      </div>

      <div v-if="hostDeviceAttachments.length === 0" class="text-center py-8 text-gray-400">
        No host devices attached
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="device in hostDeviceAttachments"
          :key="device.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div 
                :class="{
                  'bg-green-500': device.mode === 'subsystem',
                  'bg-blue-500': device.mode === 'capabilities',
                  'bg-purple-500': device.mode === 'block',
                  'bg-orange-500': device.mode === 'char',
                  'bg-gray-500': !device.mode
                }"
                class="w-3 h-3 rounded-full"
              ></div>
              <h4 class="text-white font-medium">Host Device {{ device.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ device.mode || 'unknown' }}
              </span>
            </div>
            <div v-if="device.managed" class="text-xs px-2 py-1 bg-green-500/20 text-green-400 rounded">
              Managed
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3 mb-3">
            <div v-if="device.sourceType">
              <label class="block text-xs font-medium text-gray-400 mb-1">Source Type</label>
              <p class="text-white text-sm">{{ device.sourceType }}</p>
            </div>
            <div v-if="device.sourceVendor">
              <label class="block text-xs font-medium text-gray-400 mb-1">Vendor</label>
              <p class="text-white text-sm">{{ device.sourceVendor }}</p>
            </div>
            <div v-if="device.sourceProduct">
              <label class="block text-xs font-medium text-gray-400 mb-1">Product</label>
              <p class="text-white text-sm">{{ device.sourceProduct }}</p>
            </div>
          </div>

          <div v-if="device.sourceDomain || device.sourceBus || device.sourceSlot || device.sourceFunction" class="mb-3">
            <label class="block text-xs font-medium text-gray-400 mb-2">PCI Address</label>
            <p class="text-white text-sm font-mono bg-gray-900 p-2 rounded">
              {{ device.sourceDomain || '0000' }}:{{ device.sourceBus || '00' }}:{{ device.sourceSlot || '00' }}.{{ device.sourceFunction || '0' }}
            </p>
          </div>

          <div class="flex flex-wrap gap-2 text-xs">
            <span v-if="device.alias" class="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">
              Alias: {{ device.alias }}
            </span>
            <span v-if="device.bootOrder" class="px-2 py-1 bg-purple-500/20 text-purple-400 rounded">
              Boot Order: {{ device.bootOrder }}
            </span>
          </div>

          <div v-if="device.createdAt || device.updatedAt" class="mt-3 pt-3 border-t border-gray-700 flex justify-between text-xs text-gray-500">
            <span v-if="device.createdAt">Added: {{ formatTimestamp(device.createdAt) }}</span>
            <span v-if="device.updatedAt">Modified: {{ formatTimestamp(device.updatedAt) }}</span>
          </div>
        </div>
      </div>
    </FCard>

    <!-- TPM Devices -->
    <FCard v-if="tpmAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">TPM Devices</h3>
      
      <div class="space-y-4">
        <div
          v-for="tpm in tpmAttachments"
          :key="tpm.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div class="w-3 h-3 rounded-full bg-red-500"></div>
              <h4 class="text-white font-medium">TPM Device {{ tpm.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ tpm.model || 'tpm-tis' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="tpm.backendType">
              <label class="block text-xs font-medium text-gray-400 mb-1">Backend Type</label>
              <p class="text-white text-sm">{{ tpm.backendType }}</p>
            </div>
            <div v-if="tpm.backendVersion">
              <label class="block text-xs font-medium text-gray-400 mb-1">Version</label>
              <p class="text-white text-sm">{{ tpm.backendVersion }}</p>
            </div>
            <div v-if="tpm.alias">
              <label class="block text-xs font-medium text-gray-400 mb-1">Alias</label>
              <p class="text-white text-sm">{{ tpm.alias }}</p>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Watchdog Devices -->
    <FCard v-if="watchdogAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Watchdog Devices</h3>
      
      <div class="space-y-4">
        <div
          v-for="watchdog in watchdogAttachments"
          :key="watchdog.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div class="w-3 h-3 rounded-full bg-orange-500"></div>
              <h4 class="text-white font-medium">Watchdog {{ watchdog.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ watchdog.model || 'unknown' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="watchdog.action">
              <label class="block text-xs font-medium text-gray-400 mb-1">Action</label>
              <p class="text-white text-sm">{{ watchdog.action }}</p>
            </div>
            <div v-if="watchdog.alias">
              <label class="block text-xs font-medium text-gray-400 mb-1">Alias</label>
              <p class="text-white text-sm">{{ watchdog.alias }}</p>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- RNG Devices -->
    <FCard v-if="rngAttachments.length > 0" class="p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Random Number Generator Devices</h3>
      
      <div class="space-y-4">
        <div
          v-for="rng in rngAttachments"
          :key="rng.id"
          class="p-4 bg-gray-800 rounded-lg border border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center space-x-3">
              <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
              <h4 class="text-white font-medium">RNG Device {{ rng.id }}</h4>
              <span class="text-xs px-2 py-1 bg-gray-700 text-gray-300 rounded">
                {{ rng.model || 'virtio' }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <div v-if="rng.backendModel">
              <label class="block text-xs font-medium text-gray-400 mb-1">Backend Model</label>
              <p class="text-white text-sm">{{ rng.backendModel }}</p>
            </div>
            <div v-if="rng.backendSource">
              <label class="block text-xs font-medium text-gray-400 mb-1">Backend Source</label>
              <p class="text-white text-sm bg-gray-900 p-1 rounded font-mono text-xs">{{ rng.backendSource }}</p>
            </div>
            <div v-if="rng.alias">
              <label class="block text-xs font-medium text-gray-400 mb-1">Alias</label>
              <p class="text-white text-sm">{{ rng.alias }}</p>
            </div>
          </div>

          <div v-if="rng.rate" class="mt-3">
            <label class="block text-xs font-medium text-gray-400 mb-1">Rate</label>
            <p class="text-white text-sm">{{ rng.rate }} bytes/second</p>
          </div>
        </div>
      </div>
    </FCard>
  </div>
</template>

<script setup lang="ts">
import FCard from '../ui/FCard.vue';

interface Props {
  controllerAttachments: any[];
  hostDeviceAttachments: any[];
  tpmAttachments: any[];
  watchdogAttachments: any[];
  rngAttachments: any[];
}

const props = defineProps<Props>();

const formatTimestamp = (dateString: string) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};
</script>