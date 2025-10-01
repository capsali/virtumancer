<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header Section -->
      <div class="glass-panel rounded-2xl p-8 mb-8 border border-white/10 relative overflow-hidden">
        <!-- Background Glow Effect -->
        <div class="absolute inset-0 bg-gradient-to-r from-slate-600/10 via-blue-600/10 to-slate-600/10 opacity-50"></div>
        
        <div class="relative z-10">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-4">
              <!-- Host Icon with Glow -->
              <div class="w-16 h-16 bg-gradient-to-br from-slate-500 via-blue-500 to-slate-600 rounded-xl flex items-center justify-center shadow-neon-blue">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                </svg>
              </div>
              
              <div>
                <h1 class="text-3xl font-bold bg-gradient-to-r from-slate-400 via-blue-400 to-slate-400 bg-clip-text text-transparent">
                  Hosts Overview
                </h1>
                <p class="text-slate-400 mt-2">Manage and monitor virtualization hosts</p>
              </div>
            </div>
            
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
                Add Host
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Quick Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-blue transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Total Hosts</p>
              <p class="text-2xl font-bold text-white mt-1">{{ hostStats.total }}</p>
            </div>
            <div class="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-green transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Connected</p>
              <p class="text-2xl font-bold text-white mt-1">{{ hostStats.connected }}</p>
            </div>
            <div class="w-12 h-12 bg-green-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-amber transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Total VMs</p>
              <p class="text-2xl font-bold text-white mt-1">{{ hostStats.totalVMs }}</p>
            </div>
            <div class="w-12 h-12 bg-amber-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="glass-panel rounded-xl p-6 border border-white/10 relative group hover:shadow-glow-purple transition-all duration-300">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-slate-400 text-sm">Avg. CPU Usage</p>
              <p class="text-2xl font-bold text-white mt-1">{{ hostStats.avgCpuUsage }}%</p>
            </div>
            <div class="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Hosts Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <router-link 
          v-for="host in hosts" 
          :key="host.id" 
          :to="`/hosts/${host.id}`" 
          class="group"
        >
          <div class="glass-panel rounded-xl p-6 border border-white/10 relative overflow-hidden group-hover:shadow-glow-blue transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-blue-600/5 to-slate-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
              <div class="flex items-start justify-between mb-4">
                <div>
                  <h3 class="text-xl font-bold text-white">{{ host.name || 'Unnamed Host' }}</h3>
                  <p class="text-sm text-slate-400 mt-1 font-mono">{{ host.uri }}</p>
                </div>
                <div class="flex items-center gap-2">
                  <!-- Connection Status -->
                  <span :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    host.state === 'CONNECTED' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                  ]">
                    {{ host.state }}
                  </span>
                </div>
              </div>
              
              <!-- Host Stats Grid -->
              <div class="grid grid-cols-3 gap-4">
                <div class="text-center p-3 glass-subtle rounded-lg">
                  <div class="text-lg font-bold text-white">{{ getHostVMCount(host.id) }}</div>
                  <div class="text-xs text-slate-400">VMs</div>
                </div>
                <div class="text-center p-3 glass-subtle rounded-lg">
                  <div class="text-lg font-bold text-white">20</div> <!-- Placeholder -->
                  <div class="text-xs text-slate-400">vCPUs</div>
                </div>
                <div class="text-center p-3 glass-subtle rounded-lg">
                  <div class="text-lg font-bold text-white">96GB</div> <!-- Placeholder -->
                  <div class="text-xs text-slate-400">Memory</div>
                </div>
              </div>
            </div>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'

interface HostStats {
  total: number
  connected: number
  totalVMs: number
  avgCpuUsage: number
}

const hostStore = useHostStore()
const vmStore = useVMStore()

const hostStats = ref<HostStats>({
  total: 0,
  connected: 0,
  totalVMs: 0,
  avgCpuUsage: 0
})

const hosts = computed(() => hostStore.hosts)

const getHostVMCount = (hostId: string): number => {
  return vmStore.vms.filter(vm => vm.hostId === hostId).length
}

const loadHostStats = async () => {
  try {
    const connectedHosts = hosts.value.filter(host => host.state === 'CONNECTED')
    
    hostStats.value = {
      total: hosts.value.length,
      connected: connectedHosts.length,
      totalVMs: vmStore.vms.length,
      avgCpuUsage: 45 // TODO: Calculate real average CPU usage
    }
  } catch (error) {
    console.error('Failed to load host stats:', error)
  }
}

onMounted(() => {
  loadHostStats()
})
</script>