<template>
  <div class="space-y-8">
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent mb-4">
        Storage Overview
      </h2>
      <p class="text-slate-400 text-lg">Manage storage pools, volumes, and disk resources</p>
    </div>

    <!-- Main Storage Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Storage Pools Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="router.push('/storage/pools')">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center shadow-lg shadow-emerald-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Storage Pools</h3>
                <p class="text-slate-400 text-sm">Pool management</p>
              </div>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-white">{{ storageStats.totalPools }}</div>
              <div class="text-xs text-slate-400">Total Pools</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-emerald-400">{{ storageStats.activePools }}</div>
              <div class="text-xs text-slate-400">Active</div>
            </div>
          </div>

          <div class="mt-4 h-2 bg-slate-700 rounded-full overflow-hidden">
            <div
              class="h-full bg-gradient-to-r from-emerald-500 to-emerald-600 transition-all duration-300"
              :style="{ width: `${storageStats.totalPools > 0 ? (storageStats.activePools / storageStats.totalPools) * 100 : 0}%` }"
            ></div>
          </div>
        </div>
      </FCard>

      <!-- Volumes & Disks Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="router.push('/storage/volumes')">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Volumes & Disks</h3>
                <p class="text-slate-400 text-sm">Volume management</p>
              </div>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-white">{{ storageStats.totalVolumes }}</div>
              <div class="text-xs text-slate-400">Total Volumes</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-blue-400">{{ storageStats.attachedDisks }}</div>
              <div class="text-xs text-slate-400">Attached</div>
            </div>
          </div>

          <div class="mt-6 flex items-center justify-center">
            <div class="flex items-center space-x-1">
              <div v-for="i in 5" :key="i" class="w-2 h-2 bg-gradient-to-r from-blue-500 to-blue-600 rounded-full animate-pulse" :style="{ animationDelay: `${i * 0.1}s` }"></div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Storage Statistics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <!-- Total Capacity -->
      <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-emerald transition-all duration-300">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-400 text-sm">Total Capacity</p>
            <p class="text-2xl font-bold text-white mt-1">{{ formatBytes(storageStats.totalCapacity) }}</p>
          </div>
          <div class="w-12 h-12 bg-emerald-500/20 rounded-lg flex items-center justify-center">
            <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
            </svg>
          </div>
        </div>
      </div>

      <!-- Used Space -->
      <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-blue transition-all duration-300">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-400 text-sm">Used Space</p>
            <p class="text-2xl font-bold text-white mt-1">{{ formatBytes(storageStats.usedSpace) }}</p>
          </div>
          <div class="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
            <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
            </svg>
          </div>
        </div>
      </div>

      <!-- Available Space -->
      <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-amber transition-all duration-300">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-400 text-sm">Available Space</p>
            <p class="text-2xl font-bold text-white mt-1">{{ formatBytes(storageStats.availableSpace) }}</p>
          </div>
          <div class="w-12 h-12 bg-amber-500/20 rounded-lg flex items-center justify-center">
            <svg class="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
        </div>
      </div>

      <!-- Utilization -->
      <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-purple transition-all duration-300">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-400 text-sm">Utilization</p>
            <p class="text-2xl font-bold text-white mt-1">{{ storageStats.utilization }}%</p>
          </div>
          <div class="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center">
            <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Activity / Quick Actions -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Quick Actions -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white">Quick Actions</h3>
              <p class="text-sm text-slate-400">Common storage operations</p>
            </div>
          </div>

          <div class="space-y-3">
            <button class="w-full glass-button glass-button-primary">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
              </svg>
              Create Storage Pool
            </button>
            <button class="w-full glass-button glass-button-secondary">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2m-9 0h10m-9 0V1m10 3V1m0 3l1 1v16a2 2 0 01-2 2H6a2 2 0 01-2-2V5l1-1z"/>
              </svg>
              Create Volume
            </button>
            <button class="w-full glass-button glass-button-ghost">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
              Refresh All
            </button>
          </div>
        </div>
      </FCard>

      <!-- Storage Health -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500 to-teal-500 flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white">Storage Health</h3>
              <p class="text-sm text-slate-400">System status overview</p>
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">All Pools</span>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span class="text-sm text-green-400">Healthy</span>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Data Integrity</span>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span class="text-sm text-green-400">Verified</span>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Performance</span>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-blue-400 rounded-full animate-pulse"></div>
                <span class="text-sm text-blue-400">Optimal</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import FCard from '@/components/ui/FCard.vue'

const router = useRouter()

interface StorageStats {
  totalPools: number
  activePools: number
  totalVolumes: number
  attachedDisks: number
  totalCapacity: number
  usedSpace: number
  availableSpace: number
  utilization: number
}

const storageStats = ref<StorageStats>({
  totalPools: 0,
  activePools: 0,
  totalVolumes: 0,
  attachedDisks: 0,
  totalCapacity: 0,
  usedSpace: 0,
  availableSpace: 0,
  utilization: 0
})

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const loadStorageStats = async () => {
  try {
    // TODO: Fetch real storage statistics from API
    // For now, using mock data
    const mockData = {
      totalPools: 3,
      activePools: 3,
      totalVolumes: 12,
      attachedDisks: 8,
      totalCapacity: 2199023255552, // 2TB in bytes
      usedSpace: 1099511627776, // 1TB in bytes
      availableSpace: 1099511627776, // 1TB in bytes
      utilization: 50
    }

    storageStats.value = mockData
  } catch (error) {
    console.error('Failed to load storage stats:', error)
  }
}

onMounted(() => {
  loadStorageStats()
})
</script>