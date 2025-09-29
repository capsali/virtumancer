<template>
  <div class="space-y-6">
    <!-- Navigation -->
    <div class="flex items-center justify-between mb-4">
      <FBreadcrumbs />
    </div>

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-6">
        <div>
          <div class="flex items-center gap-3 mb-1">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center shadow-xl ring-2 ring-cyan-500/20">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"/>
              </svg>
            </div>
            <h1 class="text-3xl font-bold text-white">Storage</h1>
          </div>
          <p class="text-slate-400 text-lg">Manage storage pools, volumes, and disk attachments</p>
        </div>
      </div>
      
      <div class="flex items-center gap-4">
        <FButton
          variant="ghost"
          size="sm"
          @click="refreshStorage"
          :disabled="loading"
          class="px-3 py-2"
        >
          <span v-if="!loading" class="flex items-center gap-2">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            Refresh
          </span>
          <span v-else class="flex items-center gap-2">
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Loading...
          </span>
        </FButton>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- Storage Pools -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Storage Pools</h3>
          <p class="text-2xl font-bold text-purple-400">{{ stats.storagePools }}</p>
          <p class="text-xs text-slate-500 mt-1">Total pools</p>
        </div>
      </FCard>

      <!-- Volumes -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-teal-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Volumes</h3>
          <p class="text-2xl font-bold text-green-400">{{ stats.volumes }}</p>
          <p class="text-xs text-slate-500 mt-1">Total volumes</p>
        </div>
      </FCard>

      <!-- Disks -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3-3m0 0l-3 3m3-3v12"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Disks</h3>
          <p class="text-2xl font-bold text-blue-400">{{ stats.disks }}</p>
          <p class="text-xs text-slate-500 mt-1">Total disks</p>
        </div>
      </FCard>

      <!-- Total Storage -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-semibold text-white mb-1">Total Capacity</h3>
          <p class="text-2xl font-bold text-orange-400">{{ formatBytes(stats.totalCapacity) }}</p>
          <p class="text-xs text-slate-500 mt-1">Across all storage</p>
        </div>
      </FCard>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">
      <!-- Storage Pools -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-xl">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
              </div>
              <div>
                <h3 class="text-lg font-bold text-white">Storage Pools</h3>
                <p class="text-sm text-slate-400">Libvirt storage pools</p>
              </div>
            </div>
          </div>

          <div class="space-y-4 max-h-96 overflow-y-auto">
            <div v-if="storagePools.length === 0" class="text-center py-8">
              <svg class="w-12 h-12 text-slate-600 mx-auto mb-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 6.707 6.293a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <p class="text-slate-400">No storage pools found</p>
            </div>

            <div v-for="pool in storagePools" :key="pool.id" class="p-4 bg-slate-800/30 rounded-lg border border-slate-700/30 hover:border-slate-600/50 transition-colors">
              <div class="flex items-center justify-between mb-2">
                <h4 class="font-semibold text-white">{{ pool.name }}</h4>
                <span class="text-xs px-2 py-1 rounded-full bg-purple-500/20 text-purple-400">{{ pool.type }}</span>
              </div>
              <p class="text-sm text-slate-400 mb-2">{{ pool.path }}</p>
              <div class="flex items-center justify-between text-xs">
                <span class="text-slate-500">Capacity: {{ formatBytes(pool.capacity_bytes || 0) }}</span>
                <span class="text-slate-500">Available: {{ formatBytes((pool.capacity_bytes || 0) - (pool.allocation_bytes || 0)) }}</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Volumes -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500 to-teal-500 flex items-center justify-center shadow-xl">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                </svg>
              </div>
              <div>
                <h3 class="text-lg font-bold text-white">Volumes</h3>
                <p class="text-sm text-slate-400">Storage volumes</p>
              </div>
            </div>
          </div>

          <div class="space-y-4 max-h-96 overflow-y-auto">
            <div v-if="volumes.length === 0" class="text-center py-8">
              <svg class="w-12 h-12 text-slate-600 mx-auto mb-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 6.707 6.293a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <p class="text-slate-400">No volumes found</p>
            </div>

            <div v-for="volume in volumes" :key="volume.id" class="p-4 bg-slate-800/30 rounded-lg border border-slate-700/30 hover:border-slate-600/50 transition-colors">
              <div class="flex items-center justify-between mb-2">
                <h4 class="font-semibold text-white">{{ volume.name }}</h4>
                <span class="text-xs px-2 py-1 rounded-full bg-green-500/20 text-green-400">{{ volume.format }}</span>
              </div>
              <p class="text-sm text-slate-400 mb-2">{{ volume.type }}</p>
              <div class="flex items-center justify-between text-xs">
                <span class="text-slate-500">Pool: {{ volume.pool_name }}</span>
                <span class="text-slate-500">Size: {{ formatBytes(volume.capacity_bytes || 0) }}</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Disks and Attachments -->
    <div class="grid grid-cols-1 gap-6">
      <!-- Disk Attachments -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center shadow-xl">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3-3m0 0l-3 3m3-3v12"/>
                </svg>
              </div>
              <div>
                <h3 class="text-lg font-bold text-white">Disk Attachments</h3>
                <p class="text-sm text-slate-400">VM disk attachments and configurations</p>
              </div>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-slate-700/50">
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">VM</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">Device</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">Target</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">Format</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">Size</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-slate-300">Type</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="diskAttachments.length === 0">
                  <td colspan="6" class="text-center py-8 text-slate-400">
                    No disk attachments found
                  </td>
                </tr>
                <tr v-for="attachment in diskAttachments" :key="attachment.id" class="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                  <td class="py-3 px-4 text-sm text-white">{{ attachment.vm_name || 'N/A' }}</td>
                  <td class="py-3 px-4 text-sm text-slate-300">{{ attachment.device_name || attachment.target_dev || 'N/A' }}</td>
                  <td class="py-3 px-4 text-sm text-slate-300">{{ attachment.target_dev || 'N/A' }}</td>
                  <td class="py-3 px-4 text-sm">
                    <span class="px-2 py-1 rounded-full text-xs bg-blue-500/20 text-blue-400">
                      {{ attachment.format || 'unknown' }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-sm text-slate-300">{{ formatBytes(attachment.capacity_bytes || 0) }}</td>
                  <td class="py-3 px-4 text-sm">
                    <span class="px-2 py-1 rounded-full text-xs bg-indigo-500/20 text-indigo-400">
                      {{ attachment.device_type || 'disk' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </FCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';

// Component state
const loading = ref(false);
const storagePools = ref<any[]>([]);
const volumes = ref<any[]>([]);
const diskAttachments = ref<any[]>([]);
const stats = ref({
  storagePools: 0,
  volumes: 0,
  disks: 0,
  totalCapacity: 0
});

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Load storage data
const loadStorageData = async (): Promise<void> => {
  loading.value = true;
  
  try {
    // Fetch storage data from API endpoints
    const [poolsResponse, volumesResponse, attachmentsResponse] = await Promise.all([
      fetch('/api/v1/storage/pools'),
      fetch('/api/v1/storage/volumes'),
      fetch('/api/v1/storage/disk-attachments')
    ]);

    if (poolsResponse.ok) {
      storagePools.value = await poolsResponse.json();
    } else {
      console.error('Failed to fetch storage pools:', poolsResponse.statusText);
      storagePools.value = [];
    }

    if (volumesResponse.ok) {
      volumes.value = await volumesResponse.json();
    } else {
      console.error('Failed to fetch volumes:', volumesResponse.statusText);
      volumes.value = [];
    }

    if (attachmentsResponse.ok) {
      diskAttachments.value = await attachmentsResponse.json();
    } else {
      console.error('Failed to fetch disk attachments:', attachmentsResponse.statusText);
      diskAttachments.value = [];
    }

    // Calculate stats
    stats.value = {
      storagePools: storagePools.value.length,
      volumes: volumes.value.length,
      disks: diskAttachments.value.length,
      totalCapacity: storagePools.value.reduce((total, pool) => total + (pool.capacity_bytes || 0), 0)
    };
    
  } catch (error) {
    console.error('Failed to load storage data:', error);
    storagePools.value = [];
    volumes.value = [];
    diskAttachments.value = [];
    stats.value = { storagePools: 0, volumes: 0, disks: 0, totalCapacity: 0 };
  } finally {
    loading.value = false;
  }
};

// Refresh storage data
const refreshStorage = (): void => {
  loadStorageData();
};

// Lifecycle
onMounted(() => {
  loadStorageData();
});
</script>