<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-xl flex items-center justify-center shadow-neon-blue">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
                Networks
              </h1>
              <p class="text-slate-400">Manage virtual networks and bridges</p>
            </div>
          </div>
          
          <button class="glass-button glass-button-primary">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
            </svg>
            Create Network
          </button>
        </div>
      </div>
      
      <!-- Networks Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        <div v-for="network in networks" :key="network.id" class="glass-panel rounded-xl p-6 border border-white/10 hover:shadow-glow-blue transition-all duration-300">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ network.name }}</h3>
              <span :class="[
                'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-2',
                network.mode === 'nat' ? 'bg-blue-500/20 text-blue-400' : 'bg-green-500/20 text-green-400'
              ]">
                {{ network.mode.toUpperCase() }}
              </span>
            </div>
            <div class="flex gap-2">
              <button class="glass-button glass-button-sm">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
              </button>
              <button class="glass-button glass-button-sm glass-button-danger">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
              </button>
            </div>
          </div>
          
          <div class="space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-slate-400">Bridge:</span>
              <span class="text-white font-mono">{{ network.bridge_name || 'N/A' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-slate-400">Host ID:</span>
              <span class="text-white font-mono text-xs">{{ network.host_id.substring(0, 8) }}...</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { hostApi } from '@/services/api'

interface Network {
  id: string
  name: string
  host_id: string
  bridge_name: string
  mode: string
}

const networks = ref<Network[]>([])
const loading = ref(true)

const loadNetworks = async () => {
  try {
    loading.value = true
    const response = await hostApi.getAll()
    // TODO: Implement proper network API
    // For now, use placeholder data
    networks.value = [
      {
        id: '1',
        name: 'br-vm',
        host_id: 'fffadb00-595e-4784-998c-98700b330128',
        bridge_name: 'br-vm',
        mode: 'nat'
      }
    ]
  } catch (error) {
    console.error('Failed to load networks:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadNetworks()
})
</script>