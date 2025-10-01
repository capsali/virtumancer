<template>
  <FCard class="card-glow hover:shadow-glow-blue transition-all duration-300" interactive>
    <div class="p-6">
      <div class="flex items-center space-x-4 mb-4">
        <!-- Icon Section -->
        <div class="flex items-center justify-center flex-shrink-0">
          <div class="w-12 h-12 rounded-2xl flex items-center justify-center shadow-xl" :class="iconBg">
            <svg class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
            </svg>
          </div>
        </div>

        <!-- Host Info Section -->
        <div class="flex-1 min-w-0">
          <div class="flex items-baseline gap-2 mb-1">
            <h3 class="text-2xl font-bold text-white truncate">{{ displayName }}</h3>
            <span class="text-sm text-slate-400 truncate">({{ host.id }})</span>
          </div>
          <p class="text-sm text-slate-500 truncate">URI: {{ host.uri }}</p>
        </div>
      </div>

      <div class="grid grid-cols-4 gap-4 mb-4">
        <div class="text-center">
          <div class="text-2xl font-bold text-blue-400">{{ vmCount }}</div>
          <div class="text-xs text-slate-400">VMs</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-green-400">{{ cpuDisplay }}</div>
          <div class="text-xs text-slate-400">CPUs</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-cyan-400">{{ memoryTotalDisplay }}</div>
          <div class="text-xs text-slate-400">Memory</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-orange-400">{{ storageTotalDisplay }}</div>
          <div class="text-xs text-slate-400">Storage</div>
        </div>
      </div>

      <!-- Full-width Progress Bars -->
      <div class="space-y-3">
        <div>
          <div class="flex justify-between text-xs mb-1">
            <span class="text-slate-400">Memory</span>
            <span class="text-slate-300">{{ memoryDisplay }}</span>
          </div>
          <div class="w-full bg-slate-700/50 rounded-full h-2">
            <div class="h-2 bg-gradient-to-r from-cyan-500 to-blue-500 rounded-full transition-all duration-500" :style="{ width: memoryUsagePercent + '%' }"></div>
          </div>
        </div>
        <div>
          <div class="flex justify-between text-xs mb-1">
            <span class="text-slate-400">vCPUs</span>
            <span class="text-slate-300">{{ vcpuDisplay }} / {{ cpuDisplay }}</span>
          </div>
          <div class="w-full bg-slate-700/50 rounded-full h-2">
            <div class="h-2 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-500" :style="{ width: vcpuUsagePercent + '%' }"></div>
          </div>
        </div>
        <div>
          <div class="flex justify-between text-xs mb-1">
            <span class="text-slate-400">Storage</span>
            <span class="text-slate-300">{{ storageDisplay }}</span>
          </div>
          <div class="w-full bg-slate-700/50 rounded-full h-2">
            <div class="h-2 bg-gradient-to-r from-orange-500 to-red-500 rounded-full transition-all duration-500" :style="{ width: storageUsagePercent + '%' }"></div>
          </div>
        </div>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FCard from '@/components/ui/FCard.vue'

interface Host {
  id: string
  name?: string
  uri?: string
  state?: string
}

interface Props {
  host: Host
  vmCount: number
  cpu?: number | null
  memory?: string | null
  vcpuUsage?: number | null
  storage?: string | null
  diskTotal?: number | null
  diskFree?: number | null
  memoryAvailable?: number | null
  memoryTotal?: number | null
}

const props = withDefaults(defineProps<Props>(), {
  cpu: null,
  memory: null,
  vcpuUsage: null,
  storage: null,
  diskTotal: null,
  diskFree: null,
  memoryAvailable: null,
  memoryTotal: null
})

const iconBg = computed(() => {
  switch (props.host.state) {
    case 'CONNECTED':
      return 'bg-gradient-to-br from-green-500 to-green-600 shadow-green-500/25'
    case 'DISCONNECTED':
      return 'bg-gradient-to-br from-gray-500 to-gray-600 shadow-gray-500/25'
    case 'ERROR':
      return 'bg-gradient-to-br from-red-500 to-red-600 shadow-red-500/25'
    default:
      return 'bg-gradient-to-br from-gray-500 to-gray-600 shadow-gray-500/25'
  }
})

const displayName = computed(() => props.host.name || props.host.uri || 'Unknown Host')

const formatBytes = (bytes: number | null | undefined): string => {
  if (bytes == null) return '—'
  const thresh = 1024
  if (Math.abs(bytes) < thresh) return bytes + ' B'
  const units = ['KB','MB','GB','TB','PB','EB','ZB','YB']
  let u = -1
  let b = Number(bytes)
  do {
    b /= thresh
    ++u
  } while (Math.abs(b) >= thresh && u < units.length - 1)
  return b.toFixed( b >= 10 || u === 0 ? 0 : 1 ) + ' ' + units[u]
}

const cpuDisplay = computed(() => props.cpu ?? '—')
const memoryTotalDisplay = computed(() => props.memory ?? '—')
const memoryDisplay = computed(() => {
  if (props.memoryTotal && props.memoryAvailable != null) {
    const used = props.memoryTotal - props.memoryAvailable
    return `${formatBytes(used)} / ${formatBytes(props.memoryTotal)}`
  }
  return props.memory ?? '—'
})
const vcpuDisplay = computed(() => props.vcpuUsage ?? '—')
const storageTotalDisplay = computed(() => props.storage ?? '—')
const storageDisplay = computed(() => {
  if (props.storage && props.diskFree != null && props.diskTotal != null) {
    const used = props.diskTotal - props.diskFree
    return `${formatBytes(used)} / ${props.storage}`
  }
  return props.storage ?? '—'
})

// Memory usage percentage
const memoryUsagePercent = computed(() => {
  if (props.memoryTotal && props.memoryAvailable != null) {
    const used = props.memoryTotal - props.memoryAvailable
    return Math.min(Math.round((used / props.memoryTotal) * 100), 100)
  }
  return 65 // fallback
})

// vCPU usage percentage
const vcpuUsagePercent = computed(() => {
  if (props.vcpuUsage && props.cpu) {
    return Math.min(Math.round((props.vcpuUsage / props.cpu) * 100), 100)
  }
  return 0
})

// Placeholder for storage usage percentage
const storageUsagePercent = computed(() => {
  if (props.diskTotal && props.diskFree) {
    const used = props.diskTotal - props.diskFree
    return Math.round((used / props.diskTotal) * 100)
  }
  return 45 // fallback
})
</script>

<style scoped>
.card-glow {
  will-change: opacity, box-shadow;
}
</style>
