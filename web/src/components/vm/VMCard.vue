<template>
  <FCard
    :class="[
      'transition-all duration-300',
      vm.state === 'ACTIVE' ? 'border-green-400/30' : 
      vm.state === 'ERROR' ? 'border-red-400/30' : 'border-white/10'
    ]"
    :border-glow="vm.state === 'ACTIVE'"
    :glow-color="vm.state === 'ACTIVE' ? 'accent' : 'primary'"
    interactive
  >
    <div class="space-y-4">
      <!-- VM Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div :class="[
            'w-4 h-4 rounded-full',
            getVMStatusColor(vm.state)
          ]"></div>
          <div>
            <h4 class="font-semibold text-white">{{ vm.name }}</h4>
            <p class="text-sm text-slate-400">{{ vm.description || 'No description' }}</p>
          </div>
        </div>
        
        <!-- VM State -->
        <div class="flex items-center gap-2">
          <div v-if="vm.taskState" class="animate-pulse">
            <span :class="[
              'px-2 py-1 rounded-full text-xs font-medium',
              'bg-yellow-500/20 text-yellow-400'
            ]">
              {{ vm.taskState }}
            </span>
          </div>
          <span :class="[
            'px-2 py-1 rounded-full text-xs font-medium',
            getVMStateBadgeClass(vm.state)
          ]">
            {{ vm.state.toLowerCase() }}
          </span>
        </div>
      </div>

      <!-- VM Specs -->
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-slate-400">CPU:</span>
          <span class="text-white ml-2">{{ vm.vcpuCount || 0 }} cores</span>
        </div>
        <div>
          <span class="text-slate-400">Memory:</span>
          <span class="text-white ml-2">{{ formatBytes(vm.memoryMB ? vm.memoryMB * 1024 * 1024 : 0) }}</span>
        </div>
        <div>
          <span class="text-slate-400">OS:</span>
          <span class="text-white ml-2">{{ vm.osType || 'Unknown' }}</span>
        </div>
        <div>
          <span class="text-slate-400">UUID:</span>
          <span class="text-white ml-2 font-mono text-xs">{{ vm.uuid.slice(0, 8) }}...</span>
        </div>
      </div>

      <!-- VM Actions -->
      <div class="flex gap-2 pt-2 border-t border-white/10">
        <!-- Power Controls -->
        <FButton
          v-if="vm.state === 'STOPPED' || vm.state === 'PAUSED'"
          size="sm"
          variant="primary"
          @click="$emit('action', 'start', hostId, vm.name)"
          :disabled="!!vm.taskState"
        >
          ▶️ Start
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          size="sm"
          variant="ghost"
          @click="$emit('action', 'shutdown', hostId, vm.name)"
          :disabled="!!vm.taskState"
        >
          ⏹️ Shutdown
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          size="sm"
          variant="ghost"
          @click="$emit('action', 'reboot', hostId, vm.name)"
          :disabled="!!vm.taskState"
        >
          🔄 Reboot
        </FButton>

        <!-- Advanced Actions Dropdown -->
        <div class="relative">
          <FButton
            size="sm"
            variant="ghost"
            @click="showAdvanced = !showAdvanced"
          >
            ⚙️
          </FButton>
          
          <!-- Dropdown Menu -->
          <div
            v-if="showAdvanced"
            class="absolute right-0 top-full mt-2 w-48 glass-medium rounded-lg border border-white/20 shadow-lg z-10"
            @click.stop
          >
            <div class="p-2 space-y-1">
              <button
                v-if="vm.state === 'ACTIVE'"
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="handleAdvancedAction('forceOff')"
              >
                ⚡ Force Off
              </button>
              <button
                v-if="vm.state === 'ACTIVE'"
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="handleAdvancedAction('forceReset')"
              >
                ⚡ Force Reset
              </button>
              <button
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="handleAdvancedAction('sync')"
              >
                🔄 Sync from Libvirt
              </button>
              <button
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="handleAdvancedAction('rebuild')"
              >
                🏗️ Rebuild from DB
              </button>
              <button
                v-if="vm.state === 'ACTIVE'"
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="openConsole"
              >
                💻 Console
              </button>
              <button
                class="w-full text-left px-3 py-2 text-sm text-white hover:bg-white/10 rounded"
                @click="viewDetails"
              >
                📊 Details
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Sync Status Warning -->
      <div v-if="vm.syncStatus === 'DRIFTED'" class="flex items-center gap-2 p-2 bg-yellow-500/10 border border-yellow-400/20 rounded">
        <span class="text-yellow-400">⚠️</span>
        <span class="text-sm text-yellow-400">Configuration drift detected</span>
        <FButton
          size="sm"
          variant="ghost"
          @click="handleAdvancedAction('sync')"
          class="ml-auto"
        >
          Sync Now
        </FButton>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { ref, defineEmits } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import type { VirtualMachine } from '@/types';

interface Props {
  vm: VirtualMachine;
  hostId: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  action: [action: string, hostId: string, vmName: string];
}>();

const showAdvanced = ref(false);

// Utility functions
const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-400';
    case 'STOPPED': return 'bg-red-400';
    case 'PAUSED': return 'bg-yellow-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-gray-400';
  }
};

const getVMStateBadgeClass = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500/20 text-green-400';
    case 'STOPPED': return 'bg-red-500/20 text-red-400';
    case 'PAUSED': return 'bg-yellow-500/20 text-yellow-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
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

const handleAdvancedAction = (action: string): void => {
  showAdvanced.value = false;
  emit('action', action, props.hostId, props.vm.name);
};

const openConsole = (): void => {
  showAdvanced.value = false;
  // Open console in new window
  const consoleUrl = `https://localhost:8888/hosts/${props.hostId}/vms/${props.vm.name}/console`;
  window.open(consoleUrl, '_blank', 'width=800,height=600');
};

const viewDetails = (): void => {
  showAdvanced.value = false;
  // Navigate to VM details page
  console.log('View VM details:', props.vm);
};

// Close dropdown when clicking outside
document.addEventListener('click', () => {
  showAdvanced.value = false;
});
</script>