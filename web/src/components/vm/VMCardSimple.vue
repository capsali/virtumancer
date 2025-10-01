<template>
  <FCard class="card-glow hover:scale-105 transition-all duration-200" interactive>
    <div class="p-4" role="article" :aria-label="`VM ${vm.name || vm.uuid}`">
      <div class="flex items-start justify-between">
        <div class="flex items-center gap-3">
          <div :class="[
            'w-10 h-10 rounded-lg flex items-center justify-center',
            vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
            vm.state === 'STOPPED' ? 'bg-slate-600/20 text-slate-300' :
            vm.state === 'ERROR' ? 'bg-red-500/20 text-red-400' :
            'bg-amber-500/20 text-amber-400'
          ]" aria-hidden="true">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path d="M3 13h14v-2H3v2zm0 4h14v-2H3v2zM3 5h14V3H3v2z" />
            </svg>
          </div>
          <div class="min-w-0">
            <div class="text-sm font-semibold text-white truncate">{{ vm.name || 'Unnamed VM' }}</div>
            <div class="text-xs text-slate-400 truncate">{{ vm.os_type || 'Unknown OS' }}</div>
          </div>
        </div>
        <div class="text-right">
          <div :class="[
            'text-xs font-medium px-2 py-1 rounded-full',
            vm.state === 'ACTIVE' ? 'bg-green-500/10 text-green-300' :
            vm.state === 'STOPPED' ? 'bg-slate-700/10 text-slate-400' :
            vm.state === 'ERROR' ? 'bg-red-500/10 text-red-300' :
            'bg-amber-500/10 text-amber-300'
          ]">{{ vm.state }}</div>
        </div>
      </div>

      <div class="mt-3 grid grid-cols-3 gap-3 text-sm text-slate-400">
        <div class="flex flex-col">
          <span class="font-semibold text-white text-sm">{{ vm.vcpu_count || '—' }}</span>
          <span class="truncate">vCPUs</span>
        </div>
        <div class="flex flex-col">
          <span class="font-semibold text-white text-sm">{{ formatBytes(vm.memory_bytes || 0) }}</span>
          <span class="truncate">Memory</span>
        </div>
        <div class="flex flex-col">
          <span class="font-semibold text-white text-sm truncate">{{ getHostName(vm.hostId || '') }}</span>
          <span class="truncate">Host</span>
        </div>
      </div>

      <div class="mt-3 flex gap-2" @click.stop>
        <FButton
          size="sm"
          variant="ghost"
          :aria-label="`Start ${vm.name}`"
          class="flex-1"
          v-if="vm.state === 'STOPPED'"
          @click="$emit('action', 'start', vm)"
        >
          Start
        </FButton>

        <FButton
          size="sm"
          variant="ghost"
          :aria-label="`Stop ${vm.name}`"
          class="flex-1"
          v-else-if="vm.state === 'ACTIVE'"
          @click="$emit('action', 'stop', vm)"
        >
          Stop
        </FButton>

        <FButton
          size="sm"
          variant="outline"
          :aria-label="`Console ${vm.name}`"
          class="flex-1"
          v-if="vm.state === 'ACTIVE'"
          @click="openConsole"
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
import { useHostStore } from '@/stores';

interface Props {
  vm: VirtualMachine;
  hostId: string;
}

const props = defineProps<Props>();
const emit = defineEmits(['action']);
const router = useRouter();
const hostStore = useHostStore();

const openConsole = (e?: Event) => {
  e?.stopPropagation();
  const consoleRoute = getConsoleRoute(props.hostId, props.vm.name, props.vm);
  if (consoleRoute) router.push(consoleRoute);
};

const getHostName = (id: string) => {
  const host = hostStore.hosts.find(h => h.id === id);
  return host ? (host.name || host.uri) : 'Unknown';
};

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};
</script>

<style scoped>
.card-glow { transition: all 0.2s ease; }
</style>