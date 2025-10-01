<template>
  <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="navigateToDetails">
    <div class="p-6">
      <!-- Header Section -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div :class="[
            'w-12 h-12 rounded-xl flex items-center justify-center shadow-lg',
            vm.state === 'ACTIVE' ? 'bg-gradient-to-br from-green-500 to-emerald-600 shadow-green-500/25' :
            vm.state === 'STOPPED' ? 'bg-gradient-to-br from-slate-500 to-slate-600 shadow-slate-500/25' :
            vm.state === 'ERROR' ? 'bg-gradient-to-br from-red-500 to-red-600 shadow-red-500/25' :
            'bg-gradient-to-br from-amber-500 to-orange-600 shadow-amber-500/25'
          ]">
            <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
            </svg>
          </div>
          <div>
            <h3 class="text-xl font-bold text-white">{{ vm.name || 'Unnamed VM' }}</h3>
            <p class="text-slate-400 text-sm">{{ vm.os_type || 'Virtual Machine' }}</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <!-- Status Badge -->
          <span :class="[
            'px-2 py-1 rounded-full text-xs font-medium',
            vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
            vm.state === 'STOPPED' ? 'bg-slate-500/20 text-slate-400' :
            vm.state === 'ERROR' ? 'bg-red-500/20 text-red-400' :
            'bg-amber-500/20 text-amber-400'
          ]">
            {{ vm.state }}
          </span>
          <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
          </svg>
        </div>
      </div>
      
      <!-- Resource Stats Grid -->
      <div class="grid grid-cols-2 gap-4 mb-6">
        <div class="text-center">
          <div class="text-2xl font-bold text-white">{{ vm.vcpu_count || 'N/A' }}</div>
          <div class="text-xs text-slate-400">vCPUs</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-white">{{ formatBytes(vm.memory_bytes || 0) }}</div>
          <div class="text-xs text-slate-400">Memory</div>
        </div>
      </div>
      
      <!-- Status Indicator -->
      <div class="flex items-center justify-center">
        <div :class="[
          'w-3 h-3 rounded-full mr-2',
          vm.state === 'ACTIVE' ? 'bg-green-400 animate-pulse' :
          vm.state === 'STOPPED' ? 'bg-slate-400' :
          vm.state === 'ERROR' ? 'bg-red-400 animate-pulse' :
          'bg-amber-400'
        ]"></div>
        <span class="text-sm text-slate-400">
          {{ 
            vm.state === 'ACTIVE' ? 'VM is running' :
            vm.state === 'STOPPED' ? 'VM is stopped' :
            vm.state === 'ERROR' ? 'VM has error' :
            'VM status unknown'
          }}
        </span>
      </div>

      <!-- Quick Actions -->
      <div class="mt-6 flex gap-2" @click.stop>
        <FButton
          v-if="vm.state === 'STOPPED'"
          @click="$emit('action', 'start', vm)"
          variant="primary"
          size="sm"
          class="flex-1"
        >
          Start
        </FButton>
        <FButton
          v-else-if="vm.state === 'ACTIVE'"
          @click="$emit('action', 'stop', vm)"
          variant="secondary"
          size="sm"
          class="flex-1"
        >
          Stop
        </FButton>
        <FButton
          v-if="vm.state === 'ACTIVE'"
          @click="openConsole"
          variant="outline"
          size="sm"
          class="flex-1"
        >
          Console
        </FButton>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import { getConsoleRoute } from '@/utils/console';
import type { VirtualMachine } from '@/types';

interface Props {
  vm: VirtualMachine;
  hostId: string;
}

const props = defineProps<Props>();
const emit = defineEmits(['action', 'details']);
const router = useRouter();

const openConsole = () => {
  const consoleRoute = getConsoleRoute(props.hostId, props.vm.name, props.vm);
  if (consoleRoute) {
    router.push(consoleRoute);
  }
};

const navigateToDetails = () => {
  router.push(`/hosts/${props.hostId}/vms/${props.vm.name}`);
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
</script>

<style scoped>
.card-glow {
  transition: all 0.3s ease;
}

.card-glow:hover {
  box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
}
</style>