<template>
  <div class="space-y-8">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent mb-4">
        Hosts Overview
      </h2>
      <p class="text-slate-400 text-lg">Manage and monitor virtualization hosts</p>
    </div>

    <!-- Add Host Button -->
    <div class="flex justify-center">
      <FButton
        variant="primary"
        @click="showAddHostModal = true"
        class="button-glow apply"
      >
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        Add New Host
      </FButton>
    </div>

    <!-- Host Statistics Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">
      <!-- Total Hosts Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostStats.total }}</div>
            <div class="text-sm text-slate-400">Total Hosts</div>
          </div>
        </div>
      </FCard>
      
      <!-- (Connected Hosts Card removed as per UI cleanup) -->
      
      <!-- Total VMs Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-amber-600 flex items-center justify-center shadow-lg shadow-amber-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostStats.totalVMs }}</div>
            <div class="text-sm text-slate-400">Total VMs</div>
          </div>
        </div>
      </FCard>
      
      <!-- Total vCPUs Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg shadow-purple-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostStats.totalVcpus }}</div>
            <div class="text-sm text-slate-400">Total vCPUs</div>
          </div>
        </div>
      </FCard>

      <!-- Total Memory Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-500 to-cyan-600 flex items-center justify-center shadow-lg shadow-cyan-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ formatBytes(hostStats.totalMemory) }}</div>
            <div class="text-sm text-slate-400">Total Memory</div>
          </div>
        </div>
      </FCard>

      <!-- Total Storage Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center shadow-lg shadow-orange-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ formatBytes(hostStats.totalStorage) }}</div>
            <div class="text-sm text-slate-400">Total Storage</div>
          </div>
        </div>
      </FCard>
    </div>
    
      <!-- Hosts Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <router-link
          v-for="host in hosts"
          :key="host.id"
          :to="`/hosts/${host.id}`"
          class="group"
        >
          <HostCardSimple
            :host="host"
            :vmCount="getHostVMCount(host.id)"
            :cpu="hostStore.hostStats[host.id]?.host_info?.cpu ?? hostStore.hostStats[host.id]?.resources?.cpu_count ?? null"
            :memory="formatBytes(hostStore.hostStats[host.id]?.memory_total ?? null)"
            :vcpuUsage="getHostVcpuUsage(host.id)"
            :storage="formatBytes(hostStore.hostStats[host.id]?.disk_total ?? null)"
            :diskTotal="hostStore.hostStats[host.id]?.disk_total ?? null"
            :diskFree="hostStore.hostStats[host.id]?.disk_free ?? null"
            :memoryAvailable="hostStore.hostStats[host.id]?.memory_available ?? null"
            :memoryTotal="hostStore.hostStats[host.id]?.memory_total ?? null"
          />
        </router-link>
  </div>
</div>

<!-- Add Host Modal -->
<AddHostModal
  :open="showAddHostModal"
  @close="showAddHostModal = false"
  @hostAdded="onHostAdded"
/>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import HostCardSimple from '@/components/host/HostCardSimple.vue'
import AddHostModal from '@/components/modals/AddHostModal.vue'

const router = useRouter()

interface HostStats {
  total: number
  connected: number
  totalVMs: number
  totalVcpus: number
  totalMemory: number
  totalStorage: number
}

const hostStore = useHostStore()
const vmStore = useVMStore()

const hostStats = ref<HostStats>({
  total: 0,
  connected: 0,
  totalVMs: 0,
  totalVcpus: 0,
  totalMemory: 0,
  totalStorage: 0
})

const showAddHostModal = ref(false)

const hosts = computed(() => hostStore.hosts)

const getHostVMCount = (hostId: string): number => {
  return vmStore.vms.filter(vm => vm.hostId === hostId).length
}

const getHostVcpuUsage = (hostId: string): number => {
  return vmStore.vms
    .filter(vm => vm.hostId === hostId && vm.state === 'ACTIVE')
    .reduce((total, vm) => total + vm.vcpu_count, 0)
}

const formatBytes = (bytes: number | null | undefined): string => {
  if (bytes == null) return '—'
  const thresh = 1024
  if (Math.abs(bytes) < thresh) return bytes + ' B'
  const units = ['KB','MB','GB','TB','PB','EB','ZB','YB']
  let u = -1
  let b = Number(bytes)
  do {
    b /= thresh
    ++u
  } while (Math.abs(b) >= thresh && u < units.length - 1)
  return b.toFixed( b >= 10 || u === 0 ? 0 : 1 ) + ' ' + units[u]
}

const loadHostStats = async () => {
  try {
    await hostStore.fetchHosts()
    const connectedHosts = hosts.value.filter(host => host.state === 'CONNECTED')
    
    // Count total VMs across all hosts
    let totalVMs = 0
    let totalVcpus = 0
    let totalMemory = 0
    let totalStorage = 0
    
    for (const host of hosts.value) {
      totalVMs += getHostVMCount(host.id)
      totalVcpus += getHostVcpuUsage(host.id)
      
      const hostStatsData = hostStore.hostStats[host.id]
      if (hostStatsData) {
        totalMemory += hostStatsData.memory_total || 0
        totalStorage += hostStatsData.disk_total || 0
      }
    }

    // Fetch per-host detailed stats where available
    for (const host of hosts.value) {
      try {
        await hostStore.fetchHostStats(host.id)
      } catch (e) {
        // ignore per-host stat errors
      }
    }
    
    hostStats.value = {
      total: hosts.value.length,
      connected: connectedHosts.length,
      totalVMs: totalVMs,
      totalVcpus: totalVcpus,
      totalMemory: totalMemory,
      totalStorage: totalStorage
    }
  } catch (error) {
    console.error('Failed to load host stats:', error)
  }
}

const onHostAdded = () => {
  showAddHostModal.value = false
  // Refresh the host stats after adding a new host
  loadHostStats()
}

onMounted(() => {
  loadHostStats()
})
</script>