<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Networks</h1>
        <p class="text-slate-400 mt-2">Manage virtual networks and bridges</p>
      </div>
      <FButton variant="primary" size="lg">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        Create Network
      </FButton>
    </div>

    <!-- Networks Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <FCard
        v-for="network in networks"
        :key="network.id"
        class="card-glow hover:scale-105 transition-all duration-300"
        interactive
      >
        <div class="p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ network.name }}</h3>
              <span
                :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-2',
                  network.mode === 'nat' ? 'bg-blue-500/20 text-blue-400' : 'bg-green-500/20 text-green-400'
                ]"
              >
                {{ network.mode.toUpperCase() }}
              </span>
            </div>

            <div class="flex gap-2">
              <FButton variant="secondary" size="sm">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
              </FButton>

              <FButton variant="danger" size="sm">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
              </FButton>
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
      </FCard>
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