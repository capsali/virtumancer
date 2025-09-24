<template>
  <div class="space-y-6">
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
    <div v-if="hosts.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <FCard
        v-for="host in hosts"
        :key="host.id"
        class="p-6 card-glow cursor-pointer transition-all duration-300 hover:scale-[1.02]"
        @click="$router.push(`/hosts/${host.id}`)"
      >
        <div class="space-y-4">
          <!-- Host Header -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div :class="[
                'w-4 h-4 rounded-full',
                getHostStatusColor(host)
              ]"></div>
              <div>
                <h3 class="text-lg font-semibold text-white">{{ host.uri }}</h3>
                <p class="text-sm text-slate-400">{{ getHostStatusText(host) }}</p>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm text-slate-400">{{ getHostVMCount(host) }} VMs</div>
            </div>
          </div>

          <!-- Host Stats -->
          <div class="grid grid-cols-2 gap-4">
            <div class="text-center">
              <div class="text-lg font-semibold text-white">{{ getHostActiveVMs(host) }}</div>
              <div class="text-xs text-slate-400">Active VMs</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-semibold text-white">{{ getHostTotalVMs(host) }}</div>
              <div class="text-xs text-slate-400">Total VMs</div>
            </div>
          </div>

          <!-- Host Actions -->
          <div class="flex gap-2">
            <FButton
              :variant="host.state === 'CONNECTED' ? 'ghost' : 'primary'"
              size="sm"
              @click.stop="handleHostAction(host, 'connect')"
              :disabled="host.task_state === 'CONNECTING' || host.task_state === 'DISCONNECTING'"
            >
              {{ host.state === 'CONNECTED' ? '✅ Connected' : '🔗 Connect' }}
            </FButton>
            <FButton
              variant="outline"
              size="sm"
              @click.stop="handleHostAction(host, 'disconnect')"
              :disabled="host.state === 'DISCONNECTED' || host.task_state === 'CONNECTING' || host.task_state === 'DISCONNECTING'"
            >
              ❌ Disconnect
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

const getHostStatusText = (host: any) => {
  if (host.task_state) {
    return host.task_state.toLowerCase()
  }
  return host.state.toLowerCase()
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