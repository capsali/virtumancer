<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-slate-600 rounded-xl flex items-center justify-center shadow-neon-blue">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-slate-400 bg-clip-text text-transparent">
                Volumes & Disks
              </h1>
              <p class="text-slate-400">Manage storage volumes and disk images</p>
            </div>
          </div>
          
          <div class="flex gap-3">
            <select class="glass-button glass-button-secondary bg-slate-800/50">
              <option value="all">All Pools</option>
              <option value="default">default</option>
              <option value="ssd-pool">ssd-pool</option>
            </select>
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
              Create Volume
            </button>
          </div>
        </div>
      </div>
      
      <!-- Volumes Table -->
      <div class="glass-panel rounded-xl border border-white/10 overflow-hidden">
        <div class="p-6 border-b border-white/10">
          <h2 class="text-lg font-semibold text-white">Storage Volumes</h2>
        </div>
        
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-800/30">
              <tr>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Name</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Pool</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Format</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Size</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Used By</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Status</th>
                <th class="text-left p-4 text-sm font-medium text-slate-400">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="volume in storageVolumes" :key="volume.id" class="border-t border-white/10 hover:bg-slate-800/20 transition-colors">
                <td class="p-4">
                  <div class="flex items-center gap-3">
                    <div class="w-8 h-8 bg-blue-500/20 rounded-lg flex items-center justify-center">
                      <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                      </svg>
                    </div>
                    <div>
                      <div class="font-medium text-white">{{ volume.name }}</div>
                      <div class="text-sm text-slate-400 font-mono">{{ volume.path }}</div>
                    </div>
                  </div>
                </td>
                <td class="p-4">
                  <span class="text-white">{{ volume.pool }}</span>
                </td>
                <td class="p-4">
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-500/20 text-slate-300">
                    {{ volume.format }}
                  </span>
                </td>
                <td class="p-4">
                  <span class="text-white">{{ formatBytes(volume.capacity) }}</span>
                </td>
                <td class="p-4">
                  <div v-if="volume.usedBy" class="text-white">{{ volume.usedBy }}</div>
                  <div v-else class="text-slate-400">Unused</div>
                </td>
                <td class="p-4">
                  <span :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    volume.status === 'available' ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
                  ]">
                    {{ volume.status }}
                  </span>
                </td>
                <td class="p-4">
                  <div class="flex items-center gap-2">
                    <button class="text-slate-400 hover:text-blue-400 transition-colors p-1 rounded">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
                      </svg>
                    </button>
                    <button class="text-slate-400 hover:text-amber-400 transition-colors p-1 rounded">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                      </svg>
                    </button>
                    <button class="text-slate-400 hover:text-red-400 transition-colors p-1 rounded">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface StorageVolume {
  id: string
  name: string
  pool: string
  format: string
  capacity: number
  status: string
  usedBy?: string
  path: string
}

const storageVolumes = ref<StorageVolume[]>([])

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const loadStorageVolumes = async () => {
  try {
    // TODO: Fetch real storage volumes from API
    storageVolumes.value = [
      {
        id: '1',
        name: 'ubuntu-20.04.qcow2',
        pool: 'default',
        format: 'qcow2',
        capacity: 21474836480, // 20GB
        status: 'available',
        usedBy: 'ubuntu-vm-01',
        path: '/var/lib/libvirt/images/ubuntu-20.04.qcow2'
      },
      {
        id: '2',
        name: 'centos-8-stream.qcow2',
        pool: 'default',
        format: 'qcow2',
        capacity: 32212254720, // 30GB
        status: 'available',
        usedBy: 'centos-vm-01',
        path: '/var/lib/libvirt/images/centos-8-stream.qcow2'
      },
      {
        id: '3',
        name: 'data-disk-01.qcow2',
        pool: 'ssd-pool',
        format: 'qcow2',
        capacity: 107374182400, // 100GB
        status: 'available',
        path: '/dev/vg-ssd/storage/data-disk-01.qcow2'
      },
      {
        id: '4',
        name: 'backup-image.raw',
        pool: 'default',
        format: 'raw',
        capacity: 53687091200, // 50GB
        status: 'available',
        path: '/var/lib/libvirt/images/backup-image.raw'
      }
    ]
  } catch (error) {
    console.error('Failed to load storage volumes:', error)
  }
}

onMounted(() => {
  loadStorageVolumes()
})
</script>