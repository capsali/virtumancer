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
      <div class="flex flex-col gap-4">
        <!-- Search and Filters Row -->
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search virtual machines..."
              class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus-glow transition-all duration-200"
            />
          </div>
          <div class="flex gap-2">
            <select
              v-model="statusFilter"
              class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
            >
              <option value="all">All Status</option>
              <option value="ACTIVE">Active</option>
              <option value="STOPPED">Stopped</option>
              <option value="ERROR">Error</option>
            </select>
            <select
              v-model="hostFilter"
              class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
            >
              <option value="all">All Hosts</option>
              <option v-for="host in hosts" :key="host.id" :value="host.id">{{ host.uri }}</option>
            </select>
          </div>
        </div>

        <!-- View Toggle Row -->
        <div class="flex justify-between items-center">
          <div class="text-sm text-slate-400">
            Showing {{ filteredVMs.length }} of {{ totalVMs }} virtual machines
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm text-slate-400 mr-2">View:</span>
            <div class="flex bg-slate-800/50 rounded-lg p-1 border border-slate-600/50">
              <button
                @click="viewMode = 'grid'"
                :class="[
                  'px-3 py-1.5 rounded-md text-sm transition-all duration-200 flex items-center gap-2',
                  viewMode === 'grid'
                    ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                ]"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
                Cards
              </button>
              <button
                @click="viewMode = 'list'"
                :class="[
                  'px-3 py-1.5 rounded-md text-sm transition-all duration-200 flex items-center gap-2',
                  viewMode === 'list'
                    ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                ]"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
                </svg>
                List
              </button>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- VM Grid View -->
    <div v-if="filteredVMs.length > 0 && viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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

    <!-- VM List View -->
    <div v-if="filteredVMs.length > 0 && viewMode === 'list'" class="space-y-2">
      <FCard class="card-glow">
        <!-- Table Header -->
        <div class="grid grid-cols-12 gap-4 px-6 py-3 border-b border-slate-700/50 text-sm font-medium text-slate-300">
          <div class="col-span-3">Name</div>
          <div class="col-span-2">Host</div>
          <div class="col-span-1">Status</div>
          <div class="col-span-1">vCPUs</div>
          <div class="col-span-2">Memory</div>
          <div class="col-span-3">Actions</div>
        </div>
        
        <!-- Table Rows -->
        <div class="divide-y divide-slate-700/30">
          <div
            v-for="vm in filteredVMs"
            :key="vm.uuid"
            class="grid grid-cols-12 gap-4 px-6 py-4 hover:bg-slate-800/30 cursor-pointer transition-colors duration-200"
            @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
          >
            <!-- Name Column -->
            <div class="col-span-3 flex items-center gap-3">
              <div :class="[
                'w-3 h-3 rounded-full',
                getVMStatusColor(vm.state)
              ]"></div>
              <div>
                <div class="text-white font-medium">{{ vm.name }}</div>
                <div class="text-xs text-slate-400">{{ vm.osType || 'Unknown OS' }}</div>
              </div>
            </div>
            
            <!-- Host Column -->
            <div class="col-span-2 flex items-center">
              <div class="text-slate-300">{{ vm.hostName || 'Unknown Host' }}</div>
            </div>
            
            <!-- Status Column -->
            <div class="col-span-1 flex items-center">
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium',
                vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
                vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-400' :
                vm.state === 'ERROR' ? 'bg-red-600/20 text-red-300' :
                'bg-yellow-500/20 text-yellow-400'
              ]">
                {{ vm.state }}
              </span>
            </div>
            
            <!-- vCPUs Column -->
            <div class="col-span-1 flex items-center">
              <div class="text-slate-300">{{ vm.vcpuCount || 'N/A' }}</div>
            </div>
            
            <!-- Memory Column -->
            <div class="col-span-2 flex items-center">
              <div class="text-slate-300">{{ vm.memoryMB ? `${Math.round(vm.memoryMB / 1024)}GB` : 'N/A' }}</div>
            </div>
            
            <!-- Actions Column -->
            <div class="col-span-3 flex items-center gap-2">
              <FButton
                variant="ghost"
                size="sm"
                @click.stop="handleVMAction(vm, 'start')"
                :disabled="vm.state === 'ACTIVE' || !!vm.taskState"
                class="text-xs"
              >
                ▶️
              </FButton>
              <FButton
                variant="ghost"
                size="sm"
                @click.stop="handleVMAction(vm, 'stop')"
                :disabled="vm.state === 'STOPPED' || !!vm.taskState"
                class="text-xs"
              >
                ⏹️
              </FButton>
              <FButton
                variant="outline"
                size="sm"
                @click.stop="$router.push(`/spice/${vm.hostId}/${vm.name}`)"
                :disabled="vm.state !== 'ACTIVE'"
                class="text-xs"
              >
                🖥️
              </FButton>
            </div>
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
const viewMode = ref('grid')

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
    await fetchVMsForAllHosts()
  } catch (error) {
    console.error(`Failed to ${action} VM:`, error)
  }
}

const fetchVMsForAllHosts = async () => {
  const hostPromises = hosts.value.map(host => vmStore.fetchVMs(host.id))
  await Promise.allSettled(hostPromises)
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
  await fetchVMsForAllHosts()
})
</script>