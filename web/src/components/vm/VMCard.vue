<template>
  <FCard
    class="transition-all duration-300 card-glow"
    :class="vm.state === 'ACTIVE' ? 'border-green-400/30' : vm.state === 'ERROR' ? 'border-red-400/30' : 'border-white/10'"
    interactive
  >
    <div class="space-y-4">
      <!-- VM Header with Quick Stats -->
      <div class="flex items-center justify-between" @click="!isExpanded && toggleExpanded()">
        <div class="flex items-center gap-3">
          <div :class="['w-4 h-4 rounded-full', getVMStatusColor(vm.state)]"></div>
          <div>
            <h4 class="font-semibold text-white">{{ vm.name }}</h4>
            <p v-if="!isExpanded" class="text-xs text-slate-400">
              {{ vm.vcpuCount || 0 }}C • {{ formatBytes(vm.memoryMB ? vm.memoryMB * 1024 * 1024 : 0) }} • {{ vm.state.toLowerCase() }}
            </p>
            <p v-else class="text-sm text-slate-400">{{ vm.description || 'No description' }}</p>
          </div>
        </div>
        
        <!-- VM State and Controls -->
        <div class="flex items-center gap-2">
          <!-- Quick Actions for Collapsed State -->
          <template v-if="!isExpanded">
            <FButton
              v-if="vm.state === 'ACTIVE'"
              size="sm"
              variant="ghost"
              @click.stop="openConsole"
              class="px-2"
              title="Open Console"
            >
              💻
            </FButton>
            <FButton
              v-if="vm.state === 'STOPPED' || vm.state === 'PAUSED'"
              size="sm"
              variant="primary"
              @click.stop="$emit('action', 'start', hostId, vm.name)"
              :disabled="!!vm.taskState"
              class="px-2"
              title="Start VM"
            >
              ▶️
            </FButton>
            <FButton
              size="sm"
              variant="ghost"
              @click.stop="viewDetails"
              class="px-2"
              title="View Details"
            >
              📊
            </FButton>
          </template>
          
          <div v-if="vm.taskState" class="animate-pulse">
            <span class="px-2 py-1 rounded-full text-xs font-medium bg-yellow-500/20 text-yellow-400">
              {{ vm.taskState }}
            </span>
          </div>
          <span :class="['px-2 py-1 rounded-full text-xs font-medium', getVMStateBadgeClass(vm.state)]">
            {{ vm.state.toLowerCase() }}
          </span>
          
          <!-- Expand/Collapse Button -->
          <FButton
            size="sm"
            variant="ghost"
            @click.stop="toggleExpanded"
            class="px-2"
            :title="isExpanded ? 'Collapse' : 'Expand'"
          >
            <svg 
              class="w-4 h-4 transition-transform duration-200" 
              :class="{ 'rotate-180': isExpanded }"
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </FButton>
        </div>
      </div>

      <!-- Expanded Content -->
      <div v-if="isExpanded" class="space-y-4">
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
        <div class="flex flex-wrap gap-2 pt-2 border-t border-white/10" @click.stop>
          <!-- Power Controls -->
          <FButton
            v-if="vm.state === 'STOPPED' || vm.state === 'PAUSED'"
            size="sm"
            variant="primary"
            @click="$emit('action', 'start', hostId, vm.name)"
            :disabled="!!vm.taskState"
          >
            Start VM
          </FButton>
          
          <FButton
            v-if="vm.state === 'ACTIVE'"
            size="sm"
            variant="secondary"
            @click="$emit('action', 'suspend', hostId, vm.name)"
            :disabled="!!vm.taskState"
          >
            Suspend
          </FButton>
          
          <FButton
            v-if="vm.state === 'ACTIVE'"
            size="sm"
            variant="ghost"
            @click="$emit('action', 'shutdown', hostId, vm.name)"
            :disabled="!!vm.taskState"
          >
            Shutdown
          </FButton>
          
          <FButton
            v-if="vm.state === 'ACTIVE'"
            size="sm"
            variant="danger"
            @click="$emit('action', 'destroy', hostId, vm.name)"
            :disabled="!!vm.taskState"
          >
            Force Stop
          </FButton>
          
          <!-- Console & Management -->
          <FButton
            v-if="vm.state === 'ACTIVE'"
            size="sm"
            variant="primary"
            @click="openConsole"
          >
            Open Console
          </FButton>
          
          <FButton
            size="sm"
            variant="ghost"
            @click="viewDetails"
          >
            View Details
          </FButton>
        </div>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { ref, defineEmits } from 'vue';
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

const isExpanded = ref(false);

const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value;
};

const openConsole = () => {
  const consoleRoute = getConsoleRoute(props.hostId, props.vm.name, props.vm);
  if (consoleRoute) {
    router.push(consoleRoute);
  }
};

const viewDetails = () => {
  router.push({ name: 'vm-details', params: { hostId: props.hostId, uuid: props.vm.uuid } });
};

const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-400';
    case 'STOPPED': return 'bg-gray-400';
    case 'PAUSED': return 'bg-yellow-400';
    case 'ERROR': return 'bg-red-400';
    default: return 'bg-gray-400';
  }
};

const getVMStateBadgeClass = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500/20 text-green-400';
    case 'STOPPED': return 'bg-gray-500/20 text-gray-400';
    case 'PAUSED': return 'bg-yellow-500/20 text-yellow-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-gray-500/20 text-gray-400';
  }
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