<template>
  <div class="space-y-8">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent mb-4">
        Storage Overview
      </h2>
      <p class="text-slate-400 text-lg">Manage storage pools, volumes, and disks</p>
    </div>

    <!-- Storage Statistics Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
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
            
            <!-- Quick Actions -->
            <div class="flex gap-3">
              <button class="glass-button glass-button-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                </svg>
                Refresh All
              </button>
              <button class="glass-button glass-button-primary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Create Pool
              </button>
            </div>
          </div>
        </div>
      </FCard>
      
      <!-- Storage Statistics -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-emerald transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Storage Pools</p>
              <p class="text-2xl font-bold text-white mt-1">{{ storageStats.totalPools }}</p>
            </div>
            <div class="w-12 h-12 bg-emerald-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-blue transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Total Volumes</p>
              <p class="text-2xl font-bold text-white mt-1">{{ storageStats.totalVolumes }}</p>
            </div>
            <div class="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-amber transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Total Capacity</p>
              <p class="text-2xl font-bold text-white mt-1">{{ formatBytes(storageStats.totalCapacity) }}</p>
            </div>
            <div class="w-12 h-12 bg-amber-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-purple transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Used Space</p>
              <p class="text-2xl font-bold text-white mt-1">{{ formatBytes(storageStats.usedSpace) }}</p>
            </div>
            <div class="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Navigation Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Storage Pools Card -->
        <router-link to="/storage/pools" class="group">
          <div class="glass-panel rounded-xl p-8 border border-white/10 relative overflow-hidden group-hover:shadow-glow-emerald transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-emerald-600/5 to-slate-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
              <div class="w-16 h-16 bg-emerald-500/20 rounded-2xl flex items-center justify-center mb-6 group-hover:shadow-glow-emerald transition-all duration-300">
                <svg class="w-8 h-8 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
              </div>
              <h3 class="text-xl font-bold text-white mb-2">Storage Pools</h3>
              <p class="text-slate-400 mb-4">Manage and configure storage pools across your infrastructure</p>
              <div class="flex items-center gap-2 text-emerald-400 text-sm">
                <span>{{ storageStats.totalPools }} pools</span>
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                </svg>
              </div>
            </div>
          </div>
        </router-link>
        
        <!-- Volumes Card -->
        <router-link to="/storage/volumes" class="group">
          <div class="glass-panel rounded-xl p-8 border border-white/10 relative overflow-hidden group-hover:shadow-glow-blue transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-blue-600/5 to-slate-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
              <div class="w-16 h-16 bg-blue-500/20 rounded-2xl flex items-center justify-center mb-6 group-hover:shadow-glow-blue transition-all duration-300">
                <svg class="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
              </div>
              <h3 class="text-xl font-bold text-white mb-2">Volumes & Disks</h3>
              <p class="text-slate-400 mb-4">Create and manage storage volumes for virtual machines</p>
              <div class="flex items-center gap-2 text-blue-400 text-sm">
                <span>{{ storageStats.totalVolumes }} volumes</span>
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                </svg>
              </div>
            </div>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'

const router = useRouter()

interface StorageStats {
  totalPools: number
  totalVolumes: number
  totalCapacity: number
  usedSpace: number
}

const storageStats = ref<StorageStats>({
  totalPools: 0,
  totalVolumes: 0,
  totalCapacity: 0,
  usedSpace: 0
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
    storageStats.value = {
      totalPools: 3,
      totalVolumes: 12,
      totalCapacity: 2199023255552, // 2TB in bytes
      usedSpace: 1099511627776 // 1TB in bytes
    }
  } catch (error) {
    console.error('Failed to load storage stats:', error)
  }
}

onMounted(() => {
  loadStorageStats()
})
</script>