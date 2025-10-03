<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Storage Volumes</h1>
        <p class="text-slate-400 mt-2">Manage and monitor your storage volumes</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="text-right">
          <div class="text-2xl font-bold text-white">{{ filteredVolumes.length }}</div>
          <div class="text-sm text-slate-400">Total Volumes</div>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-emerald-400">{{ availableVolumesCount }}</div>
          <div class="text-sm text-slate-400">Available</div>
        </div>
        <FButton
          @click="openCreateVolumeModal"
          variant="primary"
          size="lg"
          class="ml-6 px-4 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
          title="Create new storage volume"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
        </FButton>
      </div>
    </div>

    <!-- Search & Filters Card -->
    <FCard class="card-glow">
      <div class="p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <h2 class="text-lg font-semibold text-white">Search & Filters</h2>
            <div v-if="hasActiveFilters" class="px-2 py-1 bg-blue-500/20 text-blue-400 text-xs rounded-full border border-blue-500/30">
              {{ activeFiltersCount }} active
            </div>
          </div>
          <div class="text-sm text-slate-400">
            {{ filteredVolumes.length }} of {{ totalVolumes }} volumes
          </div>
        </div>

        <!-- Search Bar with integrated Filter Button -->
        <div class="relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search volumes... (try 'pool:default', 'format:qcow2', 'size:>10GB')"
            class="w-full px-4 py-3 pr-20 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus-glow transition-all duration-200"
          />
          <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
            <button
              @click.stop="toggleFilterDropdown"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                showFilterDropdown ? 'text-blue-400 bg-blue-500/20' : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              ]"
              title="Filters"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clip-rule="evenodd" />
              </svg>
            </button>
            <button
              @click="showFilterHelp = !showFilterHelp"
              class="p-2 text-slate-400 hover:text-white transition-colors rounded-lg hover:bg-slate-700/50"
              title="Help"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Filter Help -->
        <div v-if="showFilterHelp" class="mt-3 p-3 bg-slate-700/50 rounded-lg text-sm text-slate-300">
          <div class="font-medium mb-2">Filter Syntax:</div>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div><code class="text-blue-400">pool:default</code> - Filter by pool</div>
            <div><code class="text-blue-400">format:qcow2</code> - Filter by format</div>
            <div><code class="text-blue-400">size:>10GB</code> - Size greater than 10GB</div>
            <div><code class="text-blue-400">status:available</code> - Filter by status</div>
          </div>
        </div>

        <!-- Expandable Filter Section -->
        <div v-if="showFilterDropdown" class="mt-4 space-y-4 p-4 bg-slate-800/30 rounded-lg border border-slate-600/30 animate-slideDown">
          <!-- Quick Filters -->
          <div>
            <h4 class="text-sm font-medium text-slate-300 mb-3">Quick Filters</h4>
            <div class="grid grid-cols-4 gap-2">
              <button
                @click="applyQuickFilter('status:available')"
                class="px-3 py-2 text-sm bg-slate-700/50 hover:bg-slate-600/50 text-slate-300 hover:text-white rounded-lg transition-all duration-200 border border-slate-600/30"
              >
                Available
              </button>
              <button
                @click="applyQuickFilter('status:in-use')"
                class="px-3 py-2 text-sm bg-slate-700/50 hover:bg-slate-600/50 text-slate-300 hover:text-white rounded-lg transition-all duration-200 border border-slate-600/30"
              >
                In Use
              </button>
              <button
                @click="applyQuickFilter('format:qcow2')"
                class="px-3 py-2 text-sm bg-slate-700/50 hover:bg-slate-600/50 text-slate-300 hover:text-white rounded-lg transition-all duration-200 border border-slate-600/30"
              >
                QCOW2
              </button>
              <button
                @click="applyQuickFilter('format:raw')"
                class="px-3 py-2 text-sm bg-slate-700/50 hover:bg-slate-600/50 text-slate-300 hover:text-white rounded-lg transition-all duration-200 border border-slate-600/30"
              >
                RAW
              </button>
            </div>
          </div>

          <!-- Advanced Filters -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Pool Filter -->
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Storage Pool</label>
              <select
                v-model="poolFilter"
                class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
              >
                <option value="">All Pools</option>
                <option v-for="pool in availablePools" :key="pool" :value="pool">{{ pool }}</option>
              </select>
            </div>

            <!-- Format Filter -->
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Format</label>
              <select
                v-model="formatFilter"
                class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
              >
                <option value="">All Formats</option>
                <option value="qcow2">QCOW2</option>
                <option value="raw">RAW</option>
              </select>
            </div>
          </div>

          <!-- Filter Actions -->
          <div class="flex gap-3 pt-4 border-t border-slate-600/50">
            <button
              @click.stop="clearAllFilters"
              class="flex-1 px-4 py-2 text-sm text-slate-400 hover:text-white transition-all rounded-lg hover:bg-slate-700 border border-slate-600"
            >
              Clear Filters
            </button>
            <button
              @click.stop="showFilterDropdown = false"
              class="flex-1 px-4 py-2 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-all"
            >
              Apply Filters
            </button>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Volume List Card -->
    <FCard class="card-glow">
      <div class="border-b border-slate-700/50">
        <div class="flex items-center justify-between px-6 pt-6 pb-4">
          <h2 class="text-lg font-semibold text-white">Storage Volumes ({{ filteredVolumes.length }})</h2>

          <!-- View Controls -->
          <div class="flex items-center gap-4">
            <!-- Items per page -->
            <div class="relative">
              <button
                @click.stop="toggleItemsDropdown"
                class="px-3 py-2 bg-slate-800/70 border border-slate-600/50 rounded-lg text-white text-sm focus:outline-none min-w-[70px] flex items-center justify-between gap-2"
              >
                <span>{{ itemsPerPage === 'all' ? 'All' : itemsPerPage }}</span>
                <svg class="w-3 h-3 text-slate-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
                </svg>
              </button>

              <div v-if="showItemsDropdown" v-click-away="() => showItemsDropdown = false" class="absolute top-full mt-2 right-0 w-20 bg-slate-800/90 border border-slate-600/50 rounded-lg shadow-xl z-50">
                <div class="py-1">
                  <button
                    v-for="option in itemsPerPageOptions"
                    :key="option.value"
                    @click.stop="setItemsPerPage(option.value)"
                    :class="[
                      'w-full text-center px-3 py-2 text-sm transition-colors',
                      itemsPerPage === option.value ? 'text-blue-400 bg-blue-500/20' : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
                    ]"
                  >
                    {{ option.label }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Sorting Header -->
      <div class="grid grid-cols-12 gap-4 px-5 py-2 bg-slate-800/20 rounded-lg border border-slate-600/20 backdrop-blur-sm text-sm font-medium text-slate-300 mx-6 mt-4">
        <div class="col-span-1"></div> <!-- Status column -->
        <button
          @click="handleSort('name')"
          class="col-span-3 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none text-left"
        >
          Name {{ getSortIcon('name') }}
        </button>
        <button
          @click="handleSort('pool')"
          class="col-span-2 hidden md:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
        >
          Pool {{ getSortIcon('pool') }}
        </button>
        <button
          @click="handleSort('status')"
          class="col-span-2 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
        >
          Status {{ getSortIcon('status') }}
        </button>
        <button
          @click="handleSort('capacity')"
          class="col-span-2 hidden lg:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
        >
          Size {{ getSortIcon('capacity') }}
        </button>
        <div class="col-span-2 text-center">
          Actions
        </div>
      </div>

      <!-- Volume List Items -->
      <div class="px-6 pb-6">
        <div
          v-for="volume in sortedPaginatedVolumes"
          :key="volume.id"
          class="group relative bg-slate-800/30 hover:bg-slate-700/40 rounded-lg p-4 border border-slate-600/20 hover:border-slate-500/40 transition-all duration-300 cursor-pointer backdrop-blur-sm mt-2"
          @click="selectVolume(volume)"
        >
          <!-- Grid Layout to Match Header -->
          <div class="grid grid-cols-12 gap-4 items-center">
            <!-- Status Orb -->
            <div class="col-span-1">
              <div :class="[
                'w-3 h-3 rounded-full transition-all duration-300',
                volume.status === 'available' ? 'bg-emerald-400 animate-pulse shadow-lg shadow-emerald-400/50' :
                volume.status === 'in-use' ? 'bg-blue-400 animate-pulse shadow-lg shadow-blue-400/50' :
                'bg-slate-400'
              ]"></div>
            </div>

            <!-- Volume Name & Format -->
            <div class="col-span-3 min-w-0">
              <h3 class="text-base font-semibold text-white truncate group-hover:text-blue-300 transition-colors">
                {{ volume.name || 'Unnamed Volume' }}
              </h3>
              <p class="text-xs text-slate-400 truncate">
                {{ volume.format.toUpperCase() }}
              </p>
            </div>

            <!-- Pool Info -->
            <div class="col-span-2 hidden md:flex items-center gap-2 text-slate-300 min-w-0">
              <svg class="w-3 h-3 text-slate-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
              </svg>
              <span class="text-sm truncate">{{ volume.pool }}</span>
            </div>

            <!-- Status Badge -->
            <div class="col-span-2 flex items-center">
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium backdrop-blur-sm border',
                volume.status === 'available' ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' :
                volume.status === 'in-use' ? 'bg-blue-500/20 text-blue-300 border-blue-500/30' :
                'bg-slate-500/20 text-slate-300 border-slate-500/30'
              ]">
                {{ volume.status }}
              </span>
            </div>

            <!-- Size Information -->
            <div class="col-span-2 hidden lg:flex items-center gap-3 text-xs text-slate-400">
              <div class="flex items-center gap-1">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span>{{ formatBytes(volume.capacity) }}</span>
              </div>
            </div>

            <!-- Action Buttons Column -->
            <div class="col-span-2 flex items-center justify-center gap-1" @click.stop>
              <FButton
                variant="ghost"
                size="sm"
                @click="handleVolumeAction('edit', volume)"
                class="p-2 text-blue-400 hover:bg-blue-500/20 hover:text-blue-300 transition-all duration-200 rounded-lg"
                title="Edit Volume"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                </svg>
              </FButton>
              <FButton
                variant="ghost"
                size="sm"
                @click="handleVolumeAction('clone', volume)"
                class="p-2 text-purple-400 hover:bg-purple-500/20 hover:text-purple-300 transition-all duration-200 rounded-lg"
                title="Clone Volume"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                  <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                </svg>
              </FButton>
              <FButton
                variant="ghost"
                size="sm"
                @click="handleVolumeAction('delete', volume)"
                class="p-2 text-red-400 hover:bg-red-500/20 hover:text-red-300 transition-all duration-200 rounded-lg"
                title="Delete Volume"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
              </FButton>
            </div>
          </div>

          <!-- Mobile Size Info (shown on smaller screens) -->
          <div class="lg:hidden mt-3 pt-3 border-t border-slate-600/20 flex items-center gap-4 text-sm text-slate-400">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full bg-purple-400"></div>
              <span>{{ formatBytes(volume.capacity) }}</span>
            </div>
            <div class="md:hidden flex items-center gap-2">
              <svg class="w-3 h-3 text-slate-500" fill="currentColor" viewBox="0 0 20 20">
                <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
              </svg>
              <span>{{ volume.pool }}</span>
            </div>
          </div>

          <!-- Usage Information (if in use) -->
          <div v-if="volume.usedBy" class="mt-3 pt-3 border-t border-slate-600/20">
            <div class="flex items-center gap-2 text-sm text-slate-400">
              <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Used by: <span class="text-blue-300">{{ volume.usedBy }}</span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="px-6 pb-6">
        <div class="flex items-center justify-between">
          <div class="text-sm text-slate-400">
            Showing {{ startItem }} to {{ endItem }} of {{ totalVolumes }} volumes
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="currentPage = Math.max(1, currentPage - 1)"
              :disabled="currentPage === 1"
              class="px-3 py-2 text-sm bg-slate-800/50 border border-slate-600/50 rounded-lg text-slate-300 hover:text-white hover:bg-slate-700/50 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
            >
              Previous
            </button>
            <span class="px-3 py-2 text-sm text-slate-300">
              Page {{ currentPage }} of {{ totalPages }}
            </span>
            <button
              @click="currentPage = Math.min(totalPages, currentPage + 1)"
              :disabled="currentPage === totalPages"
              class="px-3 py-2 text-sm bg-slate-800/50 border border-slate-600/50 rounded-lg text-slate-300 hover:text-white hover:bg-slate-700/50 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </FCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'

interface StorageVolume {
  id: string
  name: string
  pool: string
  format: string
  capacity: number
  status: string
  usedBy?: string
  path: string
}

// Reactive data
const storageVolumes = ref<StorageVolume[]>([])
const selectedVolume = ref<StorageVolume | null>(null)

// Search and filter state
const searchQuery = ref('')
const showFilterDropdown = ref(false)
const showFilterHelp = ref(false)
const showItemsDropdown = ref(false)
const poolFilter = ref('')
const formatFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref<number | 'all'>(10)

// Sorting state
const sortField = ref<'name' | 'pool' | 'status' | 'capacity'>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')

// Computed properties
const availablePools = computed(() => {
  const pools = new Set(storageVolumes.value.map(volume => volume.pool))
  return Array.from(pools).sort()
})

const filteredVolumes = computed(() => {
  let filtered = storageVolumes.value

  // Apply search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(volume =>
      volume.name.toLowerCase().includes(query) ||
      volume.pool.toLowerCase().includes(query) ||
      volume.format.toLowerCase().includes(query) ||
      volume.status.toLowerCase().includes(query) ||
      (volume.usedBy && volume.usedBy.toLowerCase().includes(query))
    )
  }

  // Apply pool filter
  if (poolFilter.value) {
    filtered = filtered.filter(volume => volume.pool === poolFilter.value)
  }

  // Apply format filter
  if (formatFilter.value) {
    filtered = filtered.filter(volume => volume.format === formatFilter.value)
  }

  return filtered
})

const sortedVolumes = computed(() => {
  const sorted = [...filteredVolumes.value].sort((a, b) => {
    let aValue: any = a[sortField.value]
    let bValue: any = b[sortField.value]

    if (sortField.value === 'capacity') {
      aValue = a.capacity
      bValue = b.capacity
    }

    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }

    if (aValue < bValue) return sortDirection.value === 'asc' ? -1 : 1
    if (aValue > bValue) return sortDirection.value === 'asc' ? 1 : -1
    return 0
  })

  return sorted
})

const paginatedVolumes = computed(() => {
  if (itemsPerPage.value === 'all') {
    return sortedVolumes.value
  }

  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return sortedVolumes.value.slice(start, end)
})

const sortedPaginatedVolumes = computed(() => paginatedVolumes.value)

const totalVolumes = computed(() => storageVolumes.value.length)
const totalPages = computed(() => {
  if (itemsPerPage.value === 'all') return 1
  return Math.ceil(filteredVolumes.value.length / itemsPerPage.value)
})

const startItem = computed(() => {
  if (itemsPerPage.value === 'all') return 1
  return (currentPage.value - 1) * itemsPerPage.value + 1
})

const endItem = computed(() => {
  if (itemsPerPage.value === 'all') return filteredVolumes.value.length
  return Math.min(currentPage.value * itemsPerPage.value, filteredVolumes.value.length)
})

const availableVolumesCount = computed(() =>
  filteredVolumes.value.filter(volume => volume.status === 'available').length
)

const inUseVolumesCount = computed(() =>
  filteredVolumes.value.filter(volume => volume.status !== 'available').length
)

const totalCapacity = computed(() =>
  filteredVolumes.value.reduce((sum, volume) => sum + volume.capacity, 0)
)

const averageVolumeSize = computed(() => {
  if (filteredVolumes.value.length === 0) return 0
  return totalCapacity.value / filteredVolumes.value.length
})

const largestVolumeSize = computed(() => {
  if (filteredVolumes.value.length === 0) return 0
  return Math.max(...filteredVolumes.value.map(volume => volume.capacity))
})

const formatCounts = computed(() => {
  const counts: Record<string, number> = {}
  filteredVolumes.value.forEach(volume => {
    counts[volume.format] = (counts[volume.format] || 0) + 1
  })
  return counts
})

const hasActiveFilters = computed(() => {
  return !!(searchQuery.value.trim() || poolFilter.value || formatFilter.value)
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (searchQuery.value.trim()) count++
  if (poolFilter.value) count++
  if (formatFilter.value) count++
  return count
})

const itemsPerPageOptions = [
  { value: 10, label: '10' },
  { value: 25, label: '25' },
  { value: 50, label: '50' },
  { value: 'all', label: 'All' }
]

// Methods
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const getFormatColor = (format: string): string => {
  const colors: Record<string, string> = {
    qcow2: 'bg-gradient-to-br from-blue-500 to-blue-600 shadow-lg shadow-blue-500/25',
    raw: 'bg-gradient-to-br from-amber-500 to-amber-600 shadow-lg shadow-amber-500/25',
    default: 'bg-gradient-to-br from-slate-500 to-slate-600 shadow-lg shadow-slate-500/25'
  }
  return colors[format] || 'bg-gradient-to-br from-slate-500 to-slate-600 shadow-lg shadow-slate-500/25'
}

const getSortIcon = (field: string): string => {
  if (sortField.value !== field) return ''
  return sortDirection.value === 'asc' ? '↑' : '↓'
}

const handleSort = (field: 'name' | 'pool' | 'status' | 'capacity') => {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

const toggleFilterDropdown = () => {
  showFilterDropdown.value = !showFilterDropdown.value
}

const toggleItemsDropdown = () => {
  showItemsDropdown.value = !showItemsDropdown.value
}

const setItemsPerPage = (value: number | 'all') => {
  itemsPerPage.value = value
  currentPage.value = 1
  showItemsDropdown.value = false
}

const applyQuickFilter = (filter: string) => {
  searchQuery.value = filter
  showFilterDropdown.value = false
}

const clearAllFilters = () => {
  searchQuery.value = ''
  poolFilter.value = ''
  formatFilter.value = ''
  showFilterDropdown.value = false
}

const selectVolume = (volume: StorageVolume) => {
  selectedVolume.value = selectedVolume.value?.id === volume.id ? null : volume
}

const handleVolumeAction = (action: string, volume: StorageVolume) => {
  console.log(`Volume action: ${action} for volume ${volume.name}`)
  // TODO: Implement volume actions
}

const openCreateVolumeModal = () => {
  console.log('Opening create volume modal')
  // TODO: Implement create volume modal
}

const refreshVolumes = async () => {
  console.log('Refreshing volumes...')
  // TODO: Implement refresh functionality
}

const loadStorageVolumes = async () => {
  try {
    // TODO: Fetch real storage volumes from API
    storageVolumes.value = [
      {
        id: '1',
        name: 'ubuntu-20.04.qcow2',
        pool: 'default',
        format: 'qcow2',
        capacity: 21474836480, // 20GB
        status: 'in-use',
        usedBy: 'ubuntu-vm-01',
        path: '/var/lib/libvirt/images/ubuntu-20.04.qcow2'
      },
      {
        id: '2',
        name: 'centos-8-stream.qcow2',
        pool: 'default',
        format: 'qcow2',
        capacity: 32212254720, // 30GB
        status: 'in-use',
        usedBy: 'centos-vm-01',
        path: '/var/lib/libvirt/images/centos-8-stream.qcow2'
      },
      {
        id: '3',
        name: 'data-disk-01.qcow2',
        pool: 'ssd-pool',
        format: 'qcow2',
        capacity: 107374182400, // 100GB
        status: 'available',
        path: '/dev/vg-ssd/storage/data-disk-01.qcow2'
      },
      {
        id: '4',
        name: 'backup-image.raw',
        pool: 'default',
        format: 'raw',
        capacity: 53687091200, // 50GB
        status: 'available',
        path: '/var/lib/libvirt/images/backup-image.raw'
      },
      {
        id: '5',
        name: 'windows-server-2019.qcow2',
        pool: 'ssd-pool',
        format: 'qcow2',
        capacity: 85899345920, // 80GB
        status: 'in-use',
        usedBy: 'windows-vm-01',
        path: '/dev/vg-ssd/storage/windows-server-2019.qcow2'
      },
      {
        id: '6',
        name: 'database-disk.raw',
        pool: 'ssd-pool',
        format: 'raw',
        capacity: 268435456000, // 250GB
        status: 'available',
        path: '/dev/vg-ssd/storage/database-disk.raw'
      }
    ]
  } catch (error) {
    console.error('Failed to load storage volumes:', error)
  }
}

onMounted(() => {
  loadStorageVolumes()
})
</script>
