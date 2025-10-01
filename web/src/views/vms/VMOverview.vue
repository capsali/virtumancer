<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header Section -->
      <div class="glass-panel rounded-2xl p-8 mb-8 border border-white/10 relative overflow-hidden">
        <!-- Background Glow Effect -->
        <div class="absolute inset-0 bg-gradient-to-r from-green-600/10 via-blue-600/10 to-green-600/10 opacity-50"></div>
        
        <div class="relative z-10">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-4">
              <!-- VM Icon with Glow -->
              <div class="w-16 h-16 bg-gradient-to-br from-green-500 via-blue-500 to-green-600 rounded-xl flex items-center justify-center shadow-neon-green">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              
              <div>
                <h1 class="text-3xl font-bold bg-gradient-to-r from-green-400 via-blue-400 to-green-400 bg-clip-text text-transparent">
                  Virtual Machines Overview
                </h1>
                <p class="text-slate-400 mt-2">Comprehensive virtual machine management and monitoring</p>
              </div>
            </div>
            
            <!-- Quick Actions -->
            <div class="flex gap-3">
              <router-link to="/vms/managed" class="glass-button glass-button-primary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
                Managed VMs
              </router-link>
              <router-link to="/vms/discovered" class="glass-button glass-button-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
                Discovered VMs
              </router-link>
              <button class="glass-button glass-button-accent" @click="showCreateVMModal = true">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Create VM
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Quick Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-green transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Total VMs</p>
              <p class="text-2xl font-bold text-white mt-1">{{ vmStats.total }}</p>
            </div>
            <div class="w-12 h-12 bg-green-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-blue transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Running VMs</p>
              <p class="text-2xl font-bold text-white mt-1">{{ vmStats.running }}</p>
            </div>
            <div class="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h6m2 5.291A8.959 8.959 0 0112 21a8.949 8.949 0 01-5-1.709V19a2 2 0 01-2-2V7a2 2 0 012-2h10a2 2 0 012 2v10a2 2 0 01-2 2v.291z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-amber transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Discovered VMs</p>
              <p class="text-2xl font-bold text-white mt-1">{{ vmStats.discovered }}</p>
            </div>
            <div class="w-12 h-12 bg-amber-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-purple transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Resource Usage</p>
              <p class="text-2xl font-bold text-white mt-1">{{ vmStats.resourceUsage }}%</p>
            </div>
            <div class="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Navigation Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Managed VMs Card -->
        <router-link to="/vms/managed" class="group">
          <div class="glass-panel rounded-xl p-8 border border-white/10 relative overflow-hidden group-hover:shadow-glow-green transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-green-600/5 to-blue-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
              <div class="w-16 h-16 bg-gradient-to-br from-green-500 to-blue-500 rounded-xl flex items-center justify-center mb-4 shadow-neon-green">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              <h3 class="text-xl font-bold text-white mb-2">Managed Virtual Machines</h3>
              <p class="text-slate-400">View and manage VMs created and controlled by Virtumancer</p>
              <div class="mt-4 text-green-400 text-sm font-medium">{{ vmStats.managed }} VMs managed</div>
            </div>
          </div>
        </router-link>
        
        <!-- Discovered VMs Card -->
        <router-link to="/vms/discovered" class="group">
          <div class="glass-panel rounded-xl p-8 border border-white/10 relative overflow-hidden group-hover:shadow-glow-amber transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-amber-600/5 to-orange-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
              <div class="w-16 h-16 bg-gradient-to-br from-amber-500 to-orange-500 rounded-xl flex items-center justify-center mb-4 shadow-neon-amber">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
              </div>
              <h3 class="text-xl font-bold text-white mb-2">Discovered Virtual Machines</h3>
              <p class="text-slate-400">Import and manage VMs discovered from connected hosts</p>
              <div class="mt-4 text-amber-400 text-sm font-medium">{{ vmStats.discovered }} VMs discovered</div>
            </div>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useVMStore } from '@/stores/vmStore'

interface VMStats {
  total: number
  running: number
  managed: number
  discovered: number
  resourceUsage: number
}

const vmStore = useVMStore()
const vmStats = ref<VMStats>({
  total: 0,
  running: 0,
  managed: 0,
  discovered: 0,
  resourceUsage: 0
})

const showCreateVMModal = ref(false)

const loadVMStats = async () => {
  try {
    // Calculate stats from VM store
    const managedVMs = vmStore.vms.filter(vm => vm.source === 'managed')
    const runningVMs = vmStore.vms.filter(vm => vm.state === 'ACTIVE')
    
    vmStats.value = {
      total: vmStore.vms.length,
      running: runningVMs.length,
      managed: managedVMs.length,
      discovered: vmStore.vms.length - managedVMs.length,
      resourceUsage: 65 // TODO: Calculate real resource usage
    }
  } catch (error) {
    console.error('Failed to load VM stats:', error)
  }
}

onMounted(() => {
  loadVMStats()
})
</script>