<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-amber-500 to-orange-500 rounded-xl flex items-center justify-center shadow-neon-amber">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
                Discovered Virtual Machines
              </h1>
              <p class="text-slate-400">Import VMs discovered from connected hosts</p>
            </div>
          </div>
          
          <div class="flex gap-3">
            <button 
              @click="refreshDiscoveredVMs"
              :disabled="loading"
              class="glass-button glass-button-secondary"
            >
              <svg class="w-5 h-5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
              {{ loading ? 'Refreshing...' : 'Refresh' }}
            </button>
            <button 
              @click="importSelectedVMs"
              :disabled="selectedVMs.length === 0"
              class="glass-button glass-button-primary"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"/>
              </svg>
              Import Selected ({{ selectedVMs.length }})
            </button>
          </div>
        </div>
      </div>
      
      <!-- Host Filter -->
      <div class="glass-panel rounded-xl p-4 mb-6 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <label class="text-sm text-slate-400">Host:</label>
              <select v-model="selectedHost" class="glass-input">
                <option value="">All Hosts</option>
                <option v-for="host in hosts" :key="host.id" :value="host.id">
                  {{ host.name || host.uri }}
                </option>
              </select>
            </div>
            
            <div class="flex items-center gap-2">
              <input
                type="checkbox"
                :checked="allSelected"
                @change="toggleSelectAll"
                class="form-checkbox h-4 w-4 text-amber-500 glass-checkbox"
              />
              <label class="text-sm text-slate-400">Select All</label>
            </div>
          </div>
          
          <div class="text-sm text-slate-400">
            {{ filteredDiscoveredVMs.length }} VMs discovered
          </div>
        </div>
      </div>
      
      <!-- Discovered VMs Table -->
      <div class="glass-panel rounded-xl border border-white/10 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-800/50 border-b border-white/10">
              <tr>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Select</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">VM Name</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Host</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">State</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">vCPUs</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Memory</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/10">
              <tr v-for="vm in filteredDiscoveredVMs" :key="vm.domain_uuid" class="hover:bg-white/5 transition-colors duration-200">
                <td class="px-6 py-4 whitespace-nowrap">
                  <input
                    type="checkbox"
                    :value="vm.domain_uuid"
                    v-model="selectedVMs"
                    class="form-checkbox h-4 w-4 text-amber-500 glass-checkbox"
                  />
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-white">{{ vm.name }}</div>
                  <div class="text-sm text-slate-400 font-mono">{{ vm.domain_uuid.substring(0, 8) }}...</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-white">{{ getHostName(vm.host_id) }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    getStateColor(vm.state)
                  ]">
                    {{ vm.state }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-white">
                  {{ vm.vcpu_count }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-white">
                  {{ formatMemory(vm.memory_bytes) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button 
                    @click="importSingleVM(vm)"
                    class="glass-button glass-button-sm glass-button-primary"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"/>
                    </svg>
                    Import
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <!-- Empty State -->
      <div v-if="filteredDiscoveredVMs.length === 0 && !loading" class="glass-panel rounded-xl p-12 border border-white/10 text-center">
        <div class="w-16 h-16 bg-amber-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
        </div>
        <h3 class="text-lg font-semibold text-white mb-2">No VMs Discovered</h3>
        <p class="text-slate-400 mb-4">No unmanaged VMs were found on the connected hosts.</p>
        <button @click="refreshDiscoveredVMs" class="glass-button glass-button-primary">
          Refresh Discovery
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useHostStore } from '@/stores/hostStore'

interface DiscoveredVM {
  id: string
  name: string
  domain_uuid: string
  host_id: string
  state: string
  vcpu_count: number
  memory_bytes: number
  imported_at?: string
}

const hostStore = useHostStore()

const discoveredVMs = ref<DiscoveredVM[]>([])
const selectedHost = ref('')
const selectedVMs = ref<string[]>([])
const loading = ref(false)

const hosts = computed(() => hostStore.hosts)

const filteredDiscoveredVMs = computed(() => {
  return discoveredVMs.value.filter(vm => {
    if (selectedHost.value && vm.host_id !== selectedHost.value) return false
    return true
  })
})

const allSelected = computed(() => {
  return filteredDiscoveredVMs.value.length > 0 && 
         selectedVMs.value.length === filteredDiscoveredVMs.value.length
})

const getHostName = (hostId: string): string => {
  const host = hosts.value.find(h => h.id === hostId)
  return host?.name || host?.uri || 'Unknown Host'
}

const getStateColor = (state: string) => {
  switch (state) {
    case 'ACTIVE':
    case 'running':
      return 'bg-green-500/20 text-green-400'
    case 'STOPPED':
    case 'shut off':
      return 'bg-slate-500/20 text-slate-400'
    case 'PAUSED':
    case 'paused':
      return 'bg-amber-500/20 text-amber-400'
    default:
      return 'bg-slate-500/20 text-slate-400'
  }
}

const formatMemory = (bytes: number): string => {
  if (!bytes) return '0 MB'
  const mb = bytes / (1024 * 1024)
  if (mb >= 1024) {
    return `${(mb / 1024).toFixed(1)} GB`
  }
  return `${Math.round(mb)} MB`
}

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedVMs.value = []
  } else {
    selectedVMs.value = filteredDiscoveredVMs.value.map(vm => vm.domain_uuid)
  }
}

const refreshDiscoveredVMs = async () => {
  try {
    loading.value = true
    // TODO: Implement proper discovered VMs API
    // For now, use placeholder data
    discoveredVMs.value = []
  } catch (error) {
    console.error('Failed to refresh discovered VMs:', error)
  } finally {
    loading.value = false
  }
}

const importSingleVM = async (vm: DiscoveredVM) => {
  try {
    // TODO: Implement VM import API
    console.log('Importing VM:', vm.name)
  } catch (error) {
    console.error('Failed to import VM:', error)
  }
}

const importSelectedVMs = async () => {
  try {
    // TODO: Implement bulk VM import API
    console.log('Importing VMs:', selectedVMs.value)
    selectedVMs.value = []
  } catch (error) {
    console.error('Failed to import selected VMs:', error)
  }
}

onMounted(() => {
  refreshDiscoveredVMs()
})
</script>