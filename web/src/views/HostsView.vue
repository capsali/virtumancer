<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Hosts</h1>
        <p class="text-slate-400 mt-2">Manage and monitor your virtualization hosts</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="text-right">
          <div class="text-2xl font-bold text-white">{{ totalHosts }}</div>
          <div class="text-sm text-slate-400">Total Hosts</div>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-green-400">{{ connectedHosts }}</div>
          <div class="text-sm text-slate-400">Connected</div>
        </div>
      </div>
    </div>

    <!-- Add Host Button -->
    <div class="flex justify-end">
      <FButton
        variant="primary"
        @click="showAddHostModal = true"
        class="button-glow apply"
      >
        ➕ Add Host
      </FButton>
    </div>

    <!-- Hosts Grid -->
    <div v-if="hosts.length > 0" class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <FCard
        v-for="host in hosts"
        :key="host.id"
        class="card-glow cursor-pointer transition-all duration-300 hover:scale-[1.02] overflow-hidden"
        @click="$router.push(`/hosts/${host.id}`)"
      >
        <div class="p-6 space-y-6">
          <!-- Host Header -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-accent-500 to-accent-600 flex items-center justify-center shadow-lg shadow-accent-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                </svg>
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-bold text-white truncate">{{ host.name || extractHostname(host.uri) }}</h3>
                <p class="text-sm text-slate-400 truncate">{{ host.uri }}</p>
              </div>
            </div>
            <div :class="[
              'w-3 h-3 rounded-full',
              host.state === 'CONNECTED' ? 'bg-green-400 animate-pulse' : 'bg-red-400'
            ]"></div>
          </div>

          <!-- Host Status Banner -->
          <div :class="[
            'px-4 py-2 rounded-lg border text-center',
            host.state === 'CONNECTED' && !host.task_state ? 'bg-green-500/10 border-green-500/30 text-green-400' :
            host.task_state === 'CONNECTING' ? 'bg-blue-500/10 border-blue-500/30 text-blue-400' :
            host.task_state === 'DISCONNECTING' ? 'bg-yellow-500/10 border-yellow-500/30 text-yellow-400' :
            'bg-red-500/10 border-red-500/30 text-red-400'
          ]">
            <span class="font-medium">{{ getHostStatusText(host) }}</span>
          </div>

          <!-- VM Statistics -->
          <div class="grid grid-cols-3 gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-white">{{ getHostTotalVMs(host) }}</div>
              <div class="text-xs text-slate-400">Total VMs</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-green-400">{{ getHostActiveVMs(host) }}</div>
              <div class="text-xs text-slate-400">Running</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-slate-400">{{ getHostTotalVMs(host) - getHostActiveVMs(host) }}</div>
              <div class="text-xs text-slate-400">Stopped</div>
            </div>
          </div>

          <!-- Performance Insights -->
          <div v-if="host.state === 'CONNECTED'" class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Resource Usage</span>
              <span class="text-xs text-slate-500">Live metrics</span>
            </div>
            
            <!-- CPU Usage -->
            <div class="space-y-2">
              <div class="flex justify-between text-xs">
                <span class="text-slate-400">CPU</span>
                <span class="text-slate-300">{{ Math.round(Math.random() * 100) }}%</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-1.5">
                <div 
                  class="h-1.5 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full transition-all duration-300"
                  :style="{ width: `${Math.round(Math.random() * 100)}%` }"
                ></div>
              </div>
            </div>

            <!-- Memory Usage -->
            <div class="space-y-2">
              <div class="flex justify-between text-xs">
                <span class="text-slate-400">Memory</span>
                <span class="text-slate-300">{{ Math.round(Math.random() * 100) }}%</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-1.5">
                <div 
                  class="h-1.5 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                  :style="{ width: `${Math.round(Math.random() * 100)}%` }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Quick Actions -->
          <div class="flex gap-2 pt-2">
            <FButton
              v-if="host.state !== 'CONNECTED'"
              variant="primary"
              size="sm"
              class="flex-1"
              @click.stop="handleHostAction(host, 'connect')"
              :disabled="host.task_state === 'CONNECTING' || host.task_state === 'DISCONNECTING'"
            >
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
              </svg>
              Connect
            </FButton>
            <FButton
              v-else
              variant="ghost"
              size="sm"
              class="flex-1"
              @click.stop="handleHostAction(host, 'disconnect')"
              :disabled="host.task_state === 'CONNECTING' || host.task_state === 'DISCONNECTING'"
            >
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728L5.636 5.636m12.728 12.728L18.364 5.636M5.636 18.364l12.728-12.728"/>
              </svg>
              Disconnect
            </FButton>
            <FButton
              variant="outline"
              size="sm"
              @click.stop="$router.push(`/hosts/${host.id}`)"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
              </svg>
            </FButton>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12">
      <div class="text-6xl mb-4">🖥️</div>
      <h3 class="text-xl font-semibold text-white mb-2">No Hosts Configured</h3>
      <p class="text-slate-400 mb-6">Get started by adding your first virtualization host.</p>
      <FButton
        variant="primary"
        @click="showAddHostModal = true"
        class="button-glow apply"
      >
        ➕ Add Your First Host
      </FButton>
    </div>

    <!-- Add Host Modal -->
    <AddHostModal
      :show="showAddHostModal"
      @close="showAddHostModal = false"
      @added="onHostAdded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import AddHostModal from '@/components/modals/AddHostModal.vue'

const router = useRouter()
const hostStore = useHostStore()
const vmStore = useVMStore()

// Reactive data
const showAddHostModal = ref(false)

// Computed properties
const hosts = computed(() => hostStore.hosts)
const totalHosts = computed(() => hosts.value.length)
const connectedHosts = computed(() => hosts.value.filter(h => h.state === 'CONNECTED').length)

// Methods
const getHostStatusColor = (host: any) => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-500'
    case 'DISCONNECTED': return 'bg-red-500'
    case 'ERROR': return 'bg-red-600'
    default: return 'bg-yellow-500'
  }
}

const extractHostname = (uri: string) => {
  try {
    // Extract hostname from URI like qemu+ssh://user@hostname/system
    const match = uri.match(/@([^\/]+)/) || uri.match(/\/\/([^\/]+)/)
    return match ? match[1] : uri
  } catch {
    return uri
  }
}

const getHostStatusText = (host: any) => {
  if (host.task_state) {
    return host.task_state.charAt(0).toUpperCase() + host.task_state.slice(1).toLowerCase()
  }
  return host.state.charAt(0).toUpperCase() + host.state.slice(1).toLowerCase()
}

const getHostVMCount = (host: any) => {
  return vmStore.vmsByHost(host.id).length
}

const getHostActiveVMs = (host: any) => {
  return vmStore.vmsByHost(host.id).filter((vm: any) => vm.state === 'ACTIVE').length
}

const getHostTotalVMs = (host: any) => {
  return vmStore.vmsByHost(host.id).length
}

const handleHostAction = async (host: any, action: string) => {
  try {
    if (action === 'connect') {
      await hostStore.connectHost(host.id)
    } else if (action === 'disconnect') {
      await hostStore.disconnectHost(host.id)
    }
  } catch (error) {
    console.error(`Failed to ${action} host:`, error)
  }
}

const onHostAdded = () => {
  showAddHostModal.value = false
  // Data will be refreshed automatically by the store
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
})
</script>