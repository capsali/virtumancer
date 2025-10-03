<template>
  <div class="space-y-8">
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent mb-4">
        Storage Pools
      </h2>
      <p class="text-slate-400 text-lg">Manage and configure storage pool resources</p>
    </div>

    <!-- Quick Actions Bar -->
    <div class="flex justify-center gap-4">
      <button class="glass-button glass-button-primary">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        Create Storage Pool
      </button>
      <button class="glass-button glass-button-secondary">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
        </svg>
        Refresh All
      </button>
    </div>

    <!-- Storage Pools Overview Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Active Pools Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center shadow-lg shadow-emerald-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Active Pools</h3>
                <p class="text-slate-400 text-sm">Currently running pools</p>
              </div>
            </div>
            <div class="text-right">
              <div class="text-2xl font-bold text-emerald-400">{{ activePoolsCount }}</div>
              <div class="text-xs text-slate-400">of {{ totalPoolsCount }}</div>
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Pool Health</span>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
                <span class="text-sm text-emerald-400">Healthy</span>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Total Capacity</span>
              <span class="text-sm text-white">{{ formatBytes(totalCapacity) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Total Used</span>
              <span class="text-sm text-white">{{ formatBytes(totalUsed) }}</span>
            </div>
          </div>

          <div class="mt-6 h-2 bg-slate-700 rounded-full overflow-hidden">
            <div
              class="h-full bg-gradient-to-r from-emerald-500 to-emerald-600 transition-all duration-300"
              :style="{ width: `${totalPoolsCount > 0 ? (activePoolsCount / totalPoolsCount) * 100 : 0}%` }"
            ></div>
          </div>
        </div>
      </FCard>

      <!-- Pool Types Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Pool Types</h3>
                <p class="text-slate-400 text-sm">Storage technologies</p>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="text-center p-4 glass-subtle rounded-lg">
              <div class="text-xl font-bold text-blue-400">{{ poolTypeCounts.dir || 0 }}</div>
              <div class="text-xs text-slate-400">Directory</div>
            </div>
            <div class="text-center p-4 glass-subtle rounded-lg">
              <div class="text-xl font-bold text-purple-400">{{ poolTypeCounts.logical || 0 }}</div>
              <div class="text-xs text-slate-400">Logical</div>
            </div>
            <div class="text-center p-4 glass-subtle rounded-lg">
              <div class="text-xl font-bold text-amber-400">{{ poolTypeCounts.netfs || 0 }}</div>
              <div class="text-xs text-slate-400">Network</div>
            </div>
            <div class="text-center p-4 glass-subtle rounded-lg">
              <div class="text-xl font-bold text-slate-400">{{ poolTypeCounts.other || 0 }}</div>
              <div class="text-xs text-slate-400">Other</div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Individual Storage Pools -->
    <div class="space-y-6">
      <h3 class="text-2xl font-bold text-white">Pool Details</h3>

      <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <FCard
          v-for="pool in storagePools"
          :key="pool.id"
          class="card-glow hover:scale-105 transition-all duration-300"
          interactive
          @click="selectPool(pool)"
        >
          <div class="p-6">
            <div class="flex items-start justify-between mb-6">
              <div class="flex items-center gap-3">
                <div
                  class="w-12 h-12 rounded-xl flex items-center justify-center shadow-lg"
                  :class="getPoolTypeColor(pool.type)"
                >
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                  </svg>
                </div>
                <div>
                  <h3 class="text-xl font-bold text-white">{{ pool.name }}</h3>
                  <p class="text-slate-400 text-sm capitalize">{{ pool.type }} pool</p>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <span
                  :class="[
                    'inline-flex items-center px-3 py-1 rounded-full text-xs font-medium',
                    pool.state === 'active' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-500/20 text-slate-400'
                  ]"
                >
                  <div class="w-1.5 h-1.5 rounded-full mr-1.5" :class="pool.state === 'active' ? 'bg-emerald-400' : 'bg-slate-400'"></div>
                  {{ pool.state }}
                </span>
              </div>
            </div>

            <!-- Capacity Usage -->
            <div class="mb-6">
              <div class="flex justify-between text-sm mb-2">
                <span class="text-slate-400">Storage Usage</span>
                <span class="text-white">{{ formatBytes(pool.allocation_bytes) }} / {{ formatBytes(pool.capacity_bytes) }}</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-full transition-all duration-300"
                  :class="getUsageColor(pool.capacity_bytes > 0 ? pool.allocation_bytes / pool.capacity_bytes : 0)"
                  :style="{ width: `${pool.capacity_bytes > 0 ? (pool.allocation_bytes / pool.capacity_bytes) * 100 : 0}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-400 mt-1">
                <span>{{ pool.capacity_bytes > 0 ? Math.round((pool.allocation_bytes / pool.capacity_bytes) * 100) : 0 }}% used</span>
                <span>{{ formatBytes(pool.capacity_bytes - pool.allocation_bytes) }} available</span>
              </div>
            </div>

            <!-- Pool Stats Grid -->
            <div class="grid grid-cols-3 gap-4 mb-4">
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ getPoolVolumeCount(pool.id) }}</div>
                <div class="text-xs text-slate-400">Volumes</div>
              </div>
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ pool.capacity_bytes > 0 ? Math.round((pool.allocation_bytes / pool.capacity_bytes) * 100) : 0 }}%</div>
                <div class="text-xs text-slate-400">Allocated</div>
              </div>
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ pool.capacity_bytes > 0 ? Math.round(((pool.capacity_bytes - pool.allocation_bytes) / pool.capacity_bytes) * 100) : 0 }}%</div>
                <div class="text-xs text-slate-400">Available</div>
              </div>
            </div>

            <!-- Pool Path -->
            <div class="p-3 glass-subtle rounded-lg">
              <div class="text-xs text-slate-400 mb-1">Storage Path</div>
              <div class="text-sm text-white font-mono break-all">{{ pool.path }}</div>
            </div>
          </div>
        </FCard>
      </div>
    </div>

    <!-- Pool Management Actions -->
    <div v-if="selectedPool" class="glass-panel rounded-xl p-6 border border-white/10">
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
          </div>
          <div>
            <h3 class="text-lg font-bold text-white">Pool Management</h3>
            <p class="text-sm text-slate-400">Manage {{ selectedPool.name }}</p>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <button class="glass-button glass-button-secondary">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2m-9 0h10m-9 0V1m10 3V1m0 3l1 1v16a2 2 0 01-2 2H6a2 2 0 01-2-2V5l1-1z"/>
          </svg>
          Create Volume
        </button>
        <button class="glass-button glass-button-ghost">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          Refresh Pool
        </button>
        <button class="glass-button glass-button-danger">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
          </svg>
          Delete Pool
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import FCard from '@/components/ui/FCard.vue'
import { useStorageStore } from '@/stores/storageStore'
import type { StoragePool } from '@/types'

const storageStore = useStorageStore()

const storagePools = computed(() => storageStore.storagePools)
const storageVolumes = computed(() => storageStore.storageVolumes)
const selectedPool = ref<StoragePool | null>(null)

const totalPoolsCount = computed(() => storagePools.value.length)
const activePools = computed(() => storagePools.value.filter(pool => pool.state === 'running'))
const activePoolsCount = computed(() => activePools.value.length)

const totalCapacity = computed(() =>
  storagePools.value.reduce((sum, pool) => sum + pool.capacity_bytes, 0)
)

const totalUsed = computed(() =>
  storagePools.value.reduce((sum, pool) => sum + pool.allocation_bytes, 0)
)

const poolTypeCounts = computed(() => {
  const counts: Record<string, number> = {}
  storagePools.value.forEach(pool => {
    counts[pool.type] = (counts[pool.type] || 0) + 1
  })
  return counts
})

const getPoolVolumeCount = (poolId: string) => {
  // First try to match by storage_pool_id
  let count = storageVolumes.value.filter(volume => volume.storage_pool_id === poolId).length

  // If no matches, try to match by known pool name to path mappings
  if (count === 0) {
    const pool = storagePools.value.find(p => p.id === poolId)
    if (pool) {
      const pathMappings: Record<string, string> = {
        'default': '/var/lib/libvirt/images/',
        'vms-data': '/vms-data/',
        'vms': '/vms/'
      }

      const pathPrefix = pathMappings[pool.name]
      if (pathPrefix) {
        count = storageVolumes.value.filter(volume =>
          (volume.path && volume.path.startsWith(pathPrefix)) || (volume.name && volume.name.startsWith(pathPrefix))
        ).length
      }
    }
  }

  return count
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const getPoolTypeColor = (type: string): string => {
  const colors: Record<string, string> = {
    dir: 'bg-gradient-to-br from-blue-500 to-blue-600 shadow-lg shadow-blue-500/25',
    logical: 'bg-gradient-to-br from-purple-500 to-purple-600 shadow-lg shadow-purple-500/25',
    netfs: 'bg-gradient-to-br from-amber-500 to-amber-600 shadow-lg shadow-amber-500/25',
    default: 'bg-gradient-to-br from-slate-500 to-slate-600 shadow-lg shadow-slate-500/25'
  }
  return colors[type] || 'bg-gradient-to-br from-slate-500 to-slate-600 shadow-lg shadow-slate-500/25'
}

const getUsageColor = (ratio: number): string => {
  if (ratio < 0.7) return 'bg-gradient-to-r from-emerald-500 to-emerald-600'
  if (ratio < 0.9) return 'bg-gradient-to-r from-amber-500 to-amber-600'
  return 'bg-gradient-to-r from-red-500 to-red-600'
}

const selectPool = (pool: StoragePool) => {
  selectedPool.value = selectedPool.value?.id === pool.id ? null : pool
}

const loadStoragePools = async () => {
  await Promise.all([
    storageStore.fetchStoragePools(),
    storageStore.fetchStorageVolumes()
  ])
}

onMounted(() => {
  loadStoragePools()
})
</script>