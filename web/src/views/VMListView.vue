<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Virtual Machines</h1>
        <p class="text-slate-400 mt-2">Manage and monitor all virtual machines across your infrastructure</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="text-right">
          <div class="text-2xl font-bold text-white">{{ totalVMs }}</div>
          <div class="text-sm text-slate-400">Total VMs</div>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-green-400">{{ activeVMs }}</div>
          <div class="text-sm text-slate-400">Active</div>
        </div>
      </div>
    </div>

    <!-- Filters and Search -->
    <FCard class="p-6 card-glow">
      <div class="flex flex-col md:flex-row gap-4">
        <div class="flex-1">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search virtual machines..."
            class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-500/50"
          />
        </div>
        <div class="flex gap-2">
          <select
            v-model="statusFilter"
            class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500/50"
          >
            <option value="all">All Status</option>
            <option value="ACTIVE">Active</option>
            <option value="STOPPED">Stopped</option>
            <option value="ERROR">Error</option>
          </select>
          <select
            v-model="hostFilter"
            class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500/50"
          >
            <option value="all">All Hosts</option>
            <option v-for="host in hosts" :key="host.id" :value="host.id">{{ host.uri }}</option>
          </select>
        </div>
      </div>
    </FCard>

    <!-- VM Grid -->
    <div v-if="filteredVMs.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <FCard
        v-for="vm in filteredVMs"
        :key="vm.uuid"
        class="p-6 card-glow cursor-pointer transition-all duration-300 hover:scale-[1.02]"
        @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
      >
        <div class="space-y-4">
          <!-- VM Header -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div :class="[
                'w-4 h-4 rounded-full',
                getVMStatusColor(vm.state)
              ]"></div>
              <div>
                <h3 class="text-lg font-semibold text-white">{{ vm.name }}</h3>
                <p class="text-sm text-slate-400">{{ vm.osType || 'Unknown OS' }}</p>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm text-slate-400">{{ vm.hostName || 'Unknown Host' }}</div>
            </div>
          </div>

          <!-- VM Stats -->
          <div class="grid grid-cols-2 gap-4">
            <div class="text-center">
              <div class="text-lg font-semibold text-white">{{ vm.vcpuCount || 'N/A' }}</div>
              <div class="text-xs text-slate-400">vCPUs</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-semibold text-white">{{ vm.memoryMB ? `${Math.round(vm.memoryMB / 1024)}GB` : 'N/A' }}</div>
              <div class="text-xs text-slate-400">Memory</div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex gap-2">
            <FButton
              variant="ghost"
              size="sm"
              @click.stop="handleVMAction(vm, 'start')"
              :disabled="vm.state === 'ACTIVE' || !!vm.taskState"
            >
              ▶️ Start
            </FButton>
            <FButton
              variant="ghost"
              size="sm"
              @click.stop="handleVMAction(vm, 'stop')"
              :disabled="vm.state === 'STOPPED' || !!vm.taskState"
            >
              ⏹️ Stop
            </FButton>
            <FButton
              variant="outline"
              size="sm"
              @click.stop="$router.push(`/spice/${vm.hostId}/${vm.name}`)"
              :disabled="vm.state !== 'ACTIVE'"
            >
              🖥️ Console
            </FButton>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12">
      <div class="text-6xl mb-4">🖥️</div>
      <h3 class="text-xl font-semibold text-white mb-2">No Virtual Machines Found</h3>
      <p class="text-slate-400">No virtual machines match your current filters.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'

const router = useRouter()
const hostStore = useHostStore()
const vmStore = useVMStore()

// Reactive data
const searchQuery = ref('')
const statusFilter = ref('all')
const hostFilter = ref('all')

// Computed properties
const hosts = computed(() => hostStore.hosts)

const allVMs = computed(() => {
  const vms: any[] = []
  hosts.value.forEach(host => {
    const hostVMs = vmStore.vmsByHost(host.id)
    hostVMs.forEach((vm: any) => {
      vms.push({
        ...vm,
        hostId: host.id,
        hostName: host.uri
      })
    })
  })
  return vms
})

const filteredVMs = computed(() => {
  return allVMs.value.filter(vm => {
    const matchesSearch = !searchQuery.value ||
      vm.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (vm.osType && vm.osType.toLowerCase().includes(searchQuery.value.toLowerCase()))

    const matchesStatus = statusFilter.value === 'all' || vm.state === statusFilter.value
    const matchesHost = hostFilter.value === 'all' || vm.hostId === hostFilter.value

    return matchesSearch && matchesStatus && matchesHost
  })
})

const totalVMs = computed(() => allVMs.value.length)
const activeVMs = computed(() => allVMs.value.filter(vm => vm.state === 'ACTIVE').length)

// Methods
const getVMStatusColor = (state: string) => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500'
    case 'STOPPED': return 'bg-red-500'
    case 'ERROR': return 'bg-red-600'
    default: return 'bg-yellow-500'
  }
}

const handleVMAction = async (vm: any, action: string) => {
  try {
    if (action === 'start') {
      await vmStore.startVM(vm.hostId, vm.name)
    } else if (action === 'stop') {
      await vmStore.stopVM(vm.hostId, vm.name)
    }
    // Refresh data
    await hostStore.fetchHosts()
  } catch (error) {
    console.error(`Failed to ${action} VM:`, error)
  }
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
})
</script>