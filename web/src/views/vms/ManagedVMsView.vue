<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-green-500 to-blue-500 rounded-xl flex items-center justify-center shadow-neon-green">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-green-400 to-blue-400 bg-clip-text text-transparent">
                Managed Virtual Machines
              </h1>
              <p class="text-slate-400">VMs created and managed by Virtumancer</p>
            </div>
          </div>
          
          <div class="flex gap-3">
            <button class="glass-button glass-button-secondary">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
              Refresh
            </button>
            <button class="glass-button glass-button-primary" @click="showCreateVMModal = true">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
              </svg>
              Create VM
            </button>
          </div>
        </div>
      </div>
      
      <!-- Filter Bar -->
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
              <label class="text-sm text-slate-400">State:</label>
              <select v-model="selectedState" class="glass-input">
                <option value="">All States</option>
                <option value="ACTIVE">Active</option>
                <option value="STOPPED">Stopped</option>
                <option value="PAUSED">Paused</option>
              </select>
            </div>
          </div>
          
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
            </svg>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search VMs..."
              class="glass-input"
            />
          </div>
        </div>
      </div>
      
      <!-- VMs Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        <div v-for="vm in filteredVMs" :key="vm.uuid" class="glass-panel rounded-xl p-6 border border-white/10 hover:shadow-glow-green transition-all duration-300">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ vm.name }}</h3>
              <p class="text-sm text-slate-400 mt-1">{{ vm.description || 'No description' }}</p>
            </div>
            <div class="flex gap-2">
              <!-- State Badge -->
              <span :class="[
                'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                getStateColor(vm.state)
              ]">
                {{ vm.state }}
              </span>
            </div>
          </div>
          
          <!-- VM Stats -->
          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="text-center p-3 glass-subtle rounded-lg">
              <div class="text-lg font-bold text-white">{{ vm.vcpu_count || vm.vcpuCount }}</div>
              <div class="text-xs text-slate-400">vCPUs</div>
            </div>
            <div class="text-center p-3 glass-subtle rounded-lg">
              <div class="text-lg font-bold text-white">{{ formatMemory(vm.memory_bytes) }}</div>
              <div class="text-xs text-slate-400">Memory</div>
            </div>
          </div>
          
          <!-- Actions -->
          <div class="flex gap-2">
            <button 
              v-if="vm.state === 'STOPPED'"
              @click="startVM(vm)"
              class="flex-1 glass-button glass-button-primary glass-button-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h6m2 5.291A8.959 8.959 0 0112 21a8.949 8.949 0 01-5-1.709V19a2 2 0 01-2-2V7a2 2 0 012-2h10a2 2 0 012 2v10a2 2 0 01-2 2v.291z"/>
              </svg>
              Start
            </button>
            <button 
              v-else-if="vm.state === 'ACTIVE'"
              @click="stopVM(vm)"
              class="flex-1 glass-button glass-button-danger glass-button-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
              </svg>
              Stop
            </button>
            
            <router-link 
              :to="`/hosts/${vm.hostId}/vms/${vm.name}`"
              class="glass-button glass-button-secondary glass-button-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
              </svg>
              Details
            </router-link>
          </div>
        </div>
      </div>
      
      <!-- Create VM Modal -->
      <CreateVMModalEnhanced
        v-if="showCreateVMModal && selectedHostForCreate"
        :open="showCreateVMModal"
        :hostId="selectedHostForCreate"
        @update:open="showCreateVMModal = $event"
        @vmCreated="handleVMCreated"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useVMStore } from '@/stores/vmStore'
import { useHostStore } from '@/stores/hostStore'
import CreateVMModalEnhanced from '@/components/modals/CreateVMModalEnhanced.vue'
import type { VirtualMachine } from '@/types'

const vmStore = useVMStore()
const hostStore = useHostStore()

const selectedHost = ref('')
const selectedState = ref('')
const searchQuery = ref('')
const showCreateVMModal = ref(false)

const hosts = computed(() => hostStore.hosts)
const selectedHostForCreate = computed(() => hosts.value[0]?.id || '')

const filteredVMs = computed(() => {
  return vmStore.vms.filter(vm => {
    // Only show managed VMs
    if (vm.source !== 'managed') return false
    
    // Host filter
    if (selectedHost.value && vm.hostId !== selectedHost.value) return false
    
    // State filter
    if (selectedState.value && vm.state !== selectedState.value) return false
    
    // Search filter
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      return vm.name.toLowerCase().includes(query) ||
             (vm.description || '').toLowerCase().includes(query)
    }
    
    return true
  })
})

const getStateColor = (state: string) => {
  switch (state) {
    case 'ACTIVE':
      return 'bg-green-500/20 text-green-400'
    case 'STOPPED':
      return 'bg-slate-500/20 text-slate-400'
    case 'PAUSED':
      return 'bg-amber-500/20 text-amber-400'
    case 'ERROR':
      return 'bg-red-500/20 text-red-400'
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

const startVM = async (vm: VirtualMachine) => {
  try {
    await vmStore.startVM(vm.hostId!, vm.name)
  } catch (error) {
    console.error('Failed to start VM:', error)
  }
}

const stopVM = async (vm: VirtualMachine) => {
  try {
    await vmStore.stopVM(vm.hostId!, vm.name)
  } catch (error) {
    console.error('Failed to stop VM:', error)
  }
}

const handleVMCreated = (vm: VirtualMachine) => {
  showCreateVMModal.value = false
  // VM will be automatically added to the store by the create action
}

onMounted(() => {
  // Load VMs for all hosts
  hosts.value.forEach(host => {
    vmStore.fetchVMs(host.id)
  })
})
</script>