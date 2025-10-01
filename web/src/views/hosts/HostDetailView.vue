<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="glass-panel rounded-2xl p-6 mb-8 border border-white/10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-gradient-to-br from-slate-500 to-blue-500 rounded-xl flex items-center justify-center shadow-neon-blue">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-slate-400 to-blue-400 bg-clip-text text-transparent">
                {{ host?.name || 'Host Details' }}
              </h1>
              <p class="text-slate-400 font-mono text-sm">{{ host?.uri }}</p>
            </div>
          </div>
          
          <div class="flex gap-3">
            <span :class="[
              'inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium',
              host?.state === 'CONNECTED' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
            ]">
              {{ host?.state || 'UNKNOWN' }}
            </span>
            <router-link to="/hosts" class="glass-button glass-button-secondary">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m0 0h6a2 2 0 012 2v1M3 12h6m6 7l7-7m0 0l-7-7m0 0H3a2 2 0 00-2 2v1"/>
              </svg>
              Back to Hosts
            </router-link>
          </div>
        </div>
      </div>
      
      <!-- Content will be populated by the HostDashboard component -->
      <div class="glass-panel rounded-xl p-8 border border-white/10 text-center">
        <div class="w-16 h-16 bg-slate-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
        </div>
        <h3 class="text-lg font-semibold text-white mb-2">Host Dashboard</h3>
        <p class="text-slate-400 mb-4">Detailed host information and management will be displayed here.</p>
        <router-link :to="`/hosts/${hostId}`" class="glass-button glass-button-primary">
          View Full Dashboard
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHostStore } from '@/stores/hostStore'

interface Props {
  hostId: string
}

const props = defineProps<Props>()
const hostStore = useHostStore()

const host = computed(() => {
  return hostStore.hosts.find(h => h.id === props.hostId)
})
</script>