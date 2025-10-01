<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-emerald-500 to-slate-600 rounded-xl flex items-center justify-center shadow-neon-emerald">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-emerald-400 to-slate-400 bg-clip-text text-transparent">
                Storage Pools
              </h1>
              <p class="text-slate-400">Manage storage pool configurations</p>
            </div>
          </div>
          
          <div class="flex gap-3">
            <button class="glass-button glass-button-secondary">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
              Refresh
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
      
      <!-- Storage Pools Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div v-for="pool in storagePools" :key="pool.id" class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-emerald transition-all duration-300">
          <div class="absolute inset-0 bg-gradient-to-br from-emerald-600/5 to-slate-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
          <div class="relative z-10">
            <div class="flex items-start justify-between mb-4">
              <div>
                <h3 class="text-xl font-bold text-white">{{ pool.name }}</h3>
                <p class="text-sm text-slate-400 mt-1">{{ pool.type }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  pool.state === 'active' ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
                ]">
                  {{ pool.state }}
                </span>
                <button class="text-slate-400 hover:text-white transition-colors">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"/>
                  </svg>
                </button>
              </div>
            </div>
            
            <!-- Storage Usage Bar -->
            <div class="mb-4">
              <div class="flex justify-between text-sm mb-2">
                <span class="text-slate-400">Usage</span>
                <span class="text-white">{{ formatBytes(pool.used) }} / {{ formatBytes(pool.capacity) }}</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-2">
                <div 
                  class="bg-gradient-to-r from-emerald-500 to-emerald-400 h-2 rounded-full transition-all duration-300"
                  :style="{ width: `${(pool.used / pool.capacity) * 100}%` }"
                ></div>
              </div>
              <div class="text-xs text-slate-400 mt-1">
                {{ Math.round((pool.used / pool.capacity) * 100) }}% used
              </div>
            </div>
            
            <!-- Pool Stats -->
            <div class="grid grid-cols-3 gap-4">
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ pool.volumes }}</div>
                <div class="text-xs text-slate-400">Volumes</div>
              </div>
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ pool.allocation_percent }}%</div>
                <div class="text-xs text-slate-400">Allocated</div>
              </div>
              <div class="text-center p-3 glass-subtle rounded-lg">
                <div class="text-lg font-bold text-white">{{ pool.available_percent }}%</div>
                <div class="text-xs text-slate-400">Available</div>
              </div>
            </div>
            
            <!-- Pool Path -->
            <div class="mt-4 p-3 glass-subtle rounded-lg">
              <div class="text-xs text-slate-400 mb-1">Path</div>
              <div class="text-sm text-white font-mono">{{ pool.path }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface StoragePool {
  id: string
  name: string
  type: string
  state: string
  capacity: number
  used: number
  available: number
  allocation_percent: number
  available_percent: number
  volumes: number
  path: string
}

const storagePools = ref<StoragePool[]>([])

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const loadStoragePools = async () => {
  try {
    // TODO: Fetch real storage pools from API
    storagePools.value = [
      {
        id: '1',
        name: 'default',
        type: 'dir',
        state: 'active',
        capacity: 1000000000000, // 1TB
        used: 300000000000, // 300GB
        available: 700000000000, // 700GB
        allocation_percent: 30,
        available_percent: 70,
        volumes: 8,
        path: '/var/lib/libvirt/images'
      },
      {
        id: '2',
        name: 'ssd-pool',
        type: 'logical',
        state: 'active',
        capacity: 500000000000, // 500GB
        used: 200000000000, // 200GB
        available: 300000000000, // 300GB
        allocation_percent: 40,
        available_percent: 60,
        volumes: 4,
        path: '/dev/vg-ssd/storage'
      },
      {
        id: '3',
        name: 'backup-pool',
        type: 'netfs',
        state: 'inactive',
        capacity: 2000000000000, // 2TB
        used: 0,
        available: 2000000000000, // 2TB
        allocation_percent: 0,
        available_percent: 100,
        volumes: 0,
        path: 'nfs://backup-server/storage'
      }
    ]
  } catch (error) {
    console.error('Failed to load storage pools:', error)
  }
}

onMounted(() => {
  loadStoragePools()
})
</script>