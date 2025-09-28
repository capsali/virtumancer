<template>
  <FModal :show="show" @close="$emit('close')" size="xl">
    <template #header>
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-violet-500 flex items-center justify-center shadow-lg">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
          </svg>
        </div>
        <div>
          <h2 class="text-2xl font-bold text-white">{{ vm.name }}</h2>
          <p class="text-slate-400">Complete Virtual Machine Details</p>
        </div>
      </div>
    </template>

    <div class="space-y-6">
      <!-- Basic Information -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <FCard class="p-6 card-glow">
          <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            Basic Information
          </h3>
          <div class="space-y-4">
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Name:</span>
              <span class="col-span-2 text-white">{{ vm.name }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">UUID:</span>
              <span class="col-span-2 text-white font-mono text-xs break-all">{{ vm.uuid }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">State:</span>
              <span class="col-span-2">
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium',
                  vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
                  vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-400' :
                  vm.state === 'ERROR' ? 'bg-red-600/20 text-red-300' :
                  'bg-yellow-500/20 text-yellow-400'
                ]">
                  {{ vm.state }}
                </span>
              </span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">OS Type:</span>
              <span class="col-span-2 text-white">{{ vm.osType || 'Unknown' }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Description:</span>
              <span class="col-span-2 text-white">{{ vm.description || 'No description' }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Source:</span>
              <span class="col-span-2 text-white capitalize">{{ vm.source || 'Unknown' }}</span>
            </div>
          </div>
        </FCard>

        <FCard class="p-6 card-glow">
          <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
            </svg>
            Hardware Configuration
          </h3>
          <div class="space-y-4">
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">CPU Cores:</span>
              <span class="col-span-2 text-white">{{ vm.vcpuCount || 0 }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Memory:</span>
              <span class="col-span-2 text-white">{{ formatBytes((vm.memoryMB || 0) * 1024 * 1024) }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Disk Size:</span>
              <span class="col-span-2 text-white">{{ vm.diskSizeGB || 0 }} GB</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">CPU Model:</span>
              <span class="col-span-2 text-white">{{ vm.cpuModel || 'Default' }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Boot Device:</span>
              <span class="col-span-2 text-white">{{ vm.bootDevice || 'hd' }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Network:</span>
              <span class="col-span-2 text-white">{{ vm.networkInterface || 'default' }}</span>
            </div>
          </div>
        </FCard>
      </div>

      <!-- System Status -->
      <FCard class="p-6 card-glow">
        <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <svg class="w-5 h-5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          System Status
        </h3>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div class="space-y-4">
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Libvirt State:</span>
              <span class="col-span-2 text-white">{{ vm.libvirtState || 'Unknown' }}</span>
            </div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Sync Status:</span>
              <span class="col-span-2">
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium',
                  vm.syncStatus === 'SYNCED' ? 'bg-green-500/20 text-green-400' : 
                  vm.syncStatus === 'DRIFTED' ? 'bg-yellow-500/20 text-yellow-400' : 
                  'bg-gray-500/20 text-gray-400'
                ]">
                  {{ vm.syncStatus || 'Unknown' }}
                </span>
              </span>
            </div>
          </div>
          <div class="space-y-4">
            <div class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Task State:</span>
              <span class="col-span-2 text-white">{{ vm.taskState || 'None' }}</span>
            </div>
            <div v-if="vm.createdAt" class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Created:</span>
              <span class="col-span-2 text-white">{{ formatDate(vm.createdAt) }}</span>
            </div>
          </div>
          <div class="space-y-4">
            <div v-if="vm.updatedAt" class="grid grid-cols-3 gap-4 text-sm">
              <span class="text-slate-400 font-medium">Updated:</span>
              <span class="col-span-2 text-white">{{ formatDate(vm.updatedAt) }}</span>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Graphics and Console -->
      <FCard v-if="vm.graphics" class="p-6 card-glow">
        <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <svg class="w-5 h-5 text-purple-400" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
          </svg>
          Graphics & Console
        </h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div v-if="vm.graphics.spice" class="text-sm">
              <span class="text-slate-400 font-medium">SPICE Console:</span>
              <span class="ml-2 text-green-400">Available</span>
              <div class="mt-2 pl-4 border-l-2 border-green-400/30 text-xs text-slate-300">
                <div>Protocol: SPICE</div>
                <div>Status: Enabled</div>
              </div>
            </div>
            <div v-if="vm.graphics.vnc" class="text-sm">
              <span class="text-slate-400 font-medium">VNC Console:</span>
              <span class="ml-2 text-green-400">Available</span>
              <div class="mt-2 pl-4 border-l-2 border-green-400/30 text-xs text-slate-300">
                <div>Protocol: VNC</div>
                <div>Status: Enabled</div>
              </div>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Raw VM Data (for debugging) -->
      <details class="group">
        <summary class="cursor-pointer text-slate-400 hover:text-white transition-colors text-sm font-medium">
          <span class="group-open:rotate-90 inline-block transition-transform">▶</span>
          Raw VM Data (Debug)
        </summary>
        <FCard class="mt-4 p-4 card-glow">
          <pre class="text-xs text-slate-300 overflow-auto max-h-60 whitespace-pre-wrap">{{ JSON.stringify(vm, null, 2) }}</pre>
        </FCard>
      </details>
    </div>

    <template #footer>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <FButton
            variant="accent"
            @click="$emit('edit-hardware')"
            class="flex items-center gap-2"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            </svg>
            Edit Hardware
          </FButton>
        </div>
        <FButton
          variant="ghost"
          @click="$emit('close')"
        >
          Close
        </FButton>
      </div>
    </template>
  </FModal>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import FModal from '@/components/ui/FModal.vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import type { VirtualMachine } from '@/types';

interface Props {
  show: boolean;
  vm: VirtualMachine;
}

const props = defineProps<Props>();

defineEmits<{
  close: []
  'edit-hardware': []
}>();

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString: string): string => {
  try {
    return new Date(dateString).toLocaleString();
  } catch {
    return 'Invalid Date';
  }
};
</script>