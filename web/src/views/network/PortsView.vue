<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-xl flex items-center justify-center shadow-neon-cyan">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-cyan-400 to-blue-400 bg-clip-text text-transparent">
                Network Ports
              </h1>
              <p class="text-slate-400">Monitor network port connections and attachments</p>
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
              Create Port
            </button>
          </div>
        </div>
      </div>
      
      <!-- Ports Table -->
      <div class="glass-panel rounded-xl border border-white/10 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-800/50 border-b border-white/10">
              <tr>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Port ID</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Network</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">VM</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">MAC Address</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Status</th>
                <th class="px-6 py-4 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/10">
              <tr v-for="port in ports" :key="port.id" class="hover:bg-white/5 transition-colors duration-200">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-mono text-white">{{ port.id.substring(0, 8) }}...</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-white">{{ port.network }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-white">{{ port.vm || 'Unattached' }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-mono text-slate-300">{{ port.mac_address }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    port.status === 'active' ? 'bg-green-500/20 text-green-400' : 'bg-slate-500/20 text-slate-400'
                  ]">
                    {{ port.status }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex gap-2">
                    <button class="glass-button glass-button-sm">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                      </svg>
                    </button>
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

interface Port {
  id: string
  network: string
  vm?: string
  mac_address: string
  status: 'active' | 'inactive'
}

const ports = ref<Port[]>([])
const loading = ref(true)

const loadPorts = async () => {
  try {
    loading.value = true
    // TODO: Implement proper ports API
    ports.value = [
      {
        id: 'port-001',
        network: 'br-vm',
        vm: 'test-vm-001',
        mac_address: '52:54:00:12:34:56',
        status: 'active'
      },
      {
        id: 'port-002',
        network: 'br-vm',
        vm: 'test-vm-002',
        mac_address: '52:54:00:12:34:57',
        status: 'active'
      },
      {
        id: 'port-003',
        network: 'br-vm',
        mac_address: '52:54:00:12:34:58',
        status: 'inactive'
      }
    ]
  } catch (error) {
    console.error('Failed to load ports:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPorts()
})
</script>