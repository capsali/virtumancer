<template>
  <FCard
    class="border-blue-400/30 transition-all duration-300 card-glow"
    interactive
  >
    <div class="space-y-4">
      <!-- VM Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-4 h-4 rounded-full bg-blue-400"></div>
          <div>
            <h4 class="font-semibold text-white">{{ vm.name }}</h4>
            <p class="text-sm text-slate-400">Discovered VM</p>
          </div>
        </div>
        
        <!-- Import Status -->
        <div class="flex items-center gap-2">
          <span class="px-2 py-1 rounded-full text-xs font-medium bg-blue-500/20 text-blue-400">
            Unmanaged
          </span>
        </div>
      </div>

      <!-- VM Info -->
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-slate-400">UUID:</span>
          <span class="text-white ml-2 font-mono text-xs">{{ vm.domain_uuid?.slice(0, 8) || 'N/A' }}...</span>
        </div>
        <div>
          <span class="text-slate-400">State:</span>
          <span :class="[
            'ml-2 px-2 py-0.5 rounded-full text-xs',
            getStateClass(vm.state || 'UNKNOWN')
          ]">
            {{ (vm.state || 'UNKNOWN').toLowerCase() }}
          </span>
        </div>
        <div>
          <span class="text-slate-400">CPU:</span>
          <span class="text-white ml-2">{{ vm.vcpuCount || 0 }} cores</span>
        </div>
        <div>
          <span class="text-slate-400">Memory:</span>
          <span class="text-white ml-2">{{ formatBytes(vm.memoryMB ? vm.memoryMB * 1024 * 1024 : 0) }}</span>
        </div>
      </div>

      <!-- VM Description -->
      <div class="p-3 bg-white/5 rounded border border-white/10">
        <p class="text-sm text-slate-300">Unmanaged VM discovered on this host</p>
      </div>

      <!-- Import Actions -->
      <div class="flex gap-2 pt-2 border-t border-white/10">
        <FButton
          variant="primary"
          size="sm"
          @click="$emit('import', hostId, vm.name)"
          :disabled="importing"
        >
          <span v-if="!importing">📥 Import VM</span>
          <span v-else class="flex items-center gap-2">
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Importing...
          </span>
        </FButton>
        
        <FButton
          variant="ghost"
          size="sm"
          @click="$emit('inspect', hostId, vm.name)"
        >
          🔍 Inspect
        </FButton>
        
        <FButton
          variant="ghost"
          size="sm"
          @click="showDetails = !showDetails"
        >
          {{ showDetails ? '📄 Less' : '📄 More' }}
        </FButton>
      </div>

      <!-- Expanded Details -->
      <div v-if="showDetails" class="space-y-3 pt-3 border-t border-white/10">
        <div class="text-sm">
          <h5 class="font-medium text-white mb-2">VM Configuration</h5>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div>
              <span class="text-slate-400">OS Type:</span>
              <span class="text-white ml-2">{{ vm.osType || 'Unknown' }}</span>
            </div>
            <div>
              <span class="text-slate-400">Active:</span>
              <span class="text-white ml-2">{{ vm.isActive ? 'Yes' : 'No' }}</span>
            </div>
          </div>
        </div>

        <!-- Import Warning -->
        <div class="flex items-start gap-2 p-3 bg-yellow-500/10 border border-yellow-400/20 rounded">
          <span class="text-yellow-400 mt-0.5">⚠️</span>
          <div class="text-sm">
            <p class="text-yellow-400 font-medium">Import Notes</p>
            <p class="text-yellow-300 text-xs mt-1">
              Importing will add this VM to VirtuMancer management. 
              The VM will remain in its current state and location.
            </p>
          </div>
        </div>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import type { DiscoveredVM } from '@/types';

interface Props {
  vm: DiscoveredVM;
  hostId: string;
  importing?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  import: [hostId: string, vmName: string];
  inspect: [hostId: string, vmName: string];
}>();

const showDetails = ref(false);

// Utility functions
const getStateClass = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500/20 text-green-400';
    case 'STOPPED': return 'bg-red-500/20 text-red-400';
    case 'PAUSED': return 'bg-yellow-500/20 text-yellow-400';
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
</script>