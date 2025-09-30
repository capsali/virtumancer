<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
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
          <div// Search parsing for advanced syntax
const searchTerms = computed(() => {
  if (!searchQuery.value) return { query: '', filters: {} }
  
  const filters: Record<string, string> = {}
  let cleanQuery = searchQuery.value
  
  // Parse advanced search syntax: key:value
  const advancedPattern = /(\w+):(\w+)/g
  let match
  
  while ((match = advancedPattern.exec(searchQuery.value)) !== null) {
    if (match[1] && match[2]) {
      const [, key, value] = match
      filters[key.toLowerCase()] = value.toLowerCase()
      cleanQuery = cleanQuery.replace(match[0], '').trim()
    }
  }
  
  return { query: cleanQuery, filters }
})nt-bold text-green-400">{{ activeVMs }}</div>
          <div class="text-sm text-slate-400">Active</div>
        </div>
      </div>
    </div>

    <!-- Filters and Search -->
    <FCard class="p-6 card-glow">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <h2 class="text-lg font-semibold text-white">Search & Filters</h2>
          <!-- Active Filters Badge -->
          <div v-if="hasActiveFilters" class="px-2 py-1 bg-blue-500/20 text-blue-400 text-xs rounded-full border border-blue-500/30">
            {{ activeFiltersCount }} active
          </div>
        </div>
        <div class="text-sm text-slate-400">
          {{ activeTab === 'managed' ? `${filteredVMs.length} of ${totalVMs} VMs` : `${filteredDiscoveredVMs.length} of ${discoveredVMs.length} VMs` }}
        </div>
      </div>
      <div class="flex flex-col gap-4">
        <!-- Unified Search and Filter Bar -->
        <div class="flex gap-3">
          <div class="flex-1 relative">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search virtual machines..."
              class="w-full pl-4 pr-12 py-3 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus-glow transition-all duration-200"
            />
            <!-- Filter Dropdown Button -->
            <div class="absolute right-1 top-1/2 -translate-y-1/2">
              <button
                @click="showFilterDropdown = !showFilterDropdown"
                class="p-2 hover:bg-slate-700/50 rounded transition-colors"
                title="Filter options"
              >
                <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.207A1 1 0 013 6.5V4z" />
                </svg>
              </button>
            </div>
            
            <!-- Advanced Filter Dropdown -->
            <div
              v-if="showFilterDropdown"
              class="absolute top-full left-0 right-0 mt-2 bg-slate-800 border border-slate-600 rounded-lg shadow-xl z-50 overflow-hidden filter-dropdown-container"
            >
              <div class="p-4 space-y-4">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <!-- VM State Filter -->
                  <div>
                    <label class="block text-sm font-medium text-slate-300 mb-2">VM State</label>
                    <select
                      v-model="selectedState"
                      class="w-full p-2 bg-slate-700 border border-slate-600 rounded text-white text-sm"
                    >
                      <option value="all">All States</option>
                      <option value="running">Running</option>
                      <option value="stopped">Stopped</option>
                      <option value="paused">Paused</option>
                      <option value="error">Error</option>
                    </select>
                  </div>
                  
                  <!-- Host Filter -->
                  <div>
                    <label class="block text-sm font-medium text-slate-300 mb-2">Host</label>
                    <select
                      v-model="selectedHostId"
                      class="w-full p-2 bg-slate-700 border border-slate-600 rounded text-white text-sm"
                    >
                      <option value="all">All Hosts</option>
                      <option
                        v-for="host in allHosts"
                        :key="host.id"
                        :value="host.id"
                      >
                        {{ host.name }}
                      </option>
                    </select>
                  </div>
                </div>
                
                <!-- Active Filters Display -->
                <div v-if="activeFilters.length > 0" class="border-t border-slate-600 pt-3">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-slate-300">Active Filters:</span>
                    <button
                      @click="clearFilters"
                      class="text-xs text-slate-400 hover:text-white transition-colors"
                    >
                      Clear All
                    </button>
                  </div>
                  <div class="flex flex-wrap gap-2">
                    <span
                      v-for="filter in activeFilters"
                      :key="filter"
                      class="px-2 py-1 bg-blue-500/20 text-blue-400 text-xs rounded border border-blue-500/30"
                    >
                      {{ filter }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Main VM List Card -->
    <FCard class="card-glow overflow-hidden">
      <!-- Tab Navigation -->
      <div class="border-b border-slate-600">
        <div class="flex">
          <button
            @click="activeTab = 'managed'"
            :class="[
              'px-6 py-4 text-sm font-medium border-b-2 transition-colors',
              activeTab === 'managed'
                ? 'border-blue-500 text-blue-400 bg-blue-500/10'
                : 'border-transparent text-slate-400 hover:text-white hover:bg-slate-800/50'
            ]"
          >
            Managed VMs ({{ filteredVMs.length }})
          </button>
          <button
            @click="activeTab = 'discovered'"
            :class="[
              'px-6 py-4 text-sm font-medium border-b-2 transition-colors',
              activeTab === 'discovered'
                ? 'border-blue-500 text-blue-400 bg-blue-500/10'
                : 'border-transparent text-slate-400 hover:text-white hover:bg-slate-800/50'
            ]"
          >
            Discovered VMs ({{ filteredDiscoveredVMs.length }})
          </button>
        </div>
        
        <!-- Tab Controls -->
        <div v-if="activeTab === 'managed'" class="flex justify-between items-center p-4 bg-slate-800/30">
          <!-- View Mode Controls -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-slate-400 mr-2">View:</span>
            <button
              @click="viewMode = 'grid'"
              :class="[
                'p-2 rounded transition-colors',
                viewMode === 'grid'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                  : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              ]"
              title="Grid view"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>
            <button
              @click="viewMode = 'list'"
              :class="[
                'p-2 rounded transition-colors',
                viewMode === 'list'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                  : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              ]"
              title="List view"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
              </svg>
            </button>
            <button
              @click="viewMode = 'compact'"
              :class="[
                'p-2 rounded transition-colors',
                viewMode === 'compact'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                  : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              ]"
              title="Compact view"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 7a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 11a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 15a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
          
          <!-- Pagination Controls -->
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2">
              <span class="text-sm text-slate-400">Show:</span>
              <select
                v-model="itemsPerPage"
                class="px-2 py-1 bg-slate-700 border border-slate-600 rounded text-white text-sm"
              >
                <option :value="10">10</option>
                <option :value="25">25</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
                <option value="all">All</option>
              </select>
              <span class="text-sm text-slate-400">per page</span>
            </div>
            
            <div v-if="itemsPerPage !== 'all' && Math.ceil(filteredVMs.length / itemsPerPage) > 1" class="flex items-center gap-2">
              <button
                @click="currentPage = Math.max(1, currentPage - 1)"
                :disabled="currentPage === 1"
                class="p-1 rounded hover:bg-slate-700/50 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Previous page"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
              
              <span class="text-sm text-slate-400">
                {{ currentPage }} of {{ Math.ceil(filteredVMs.length / itemsPerPage) }}
              </span>
              
              <button
                @click="currentPage = Math.min(Math.ceil(filteredVMs.length / itemsPerPage), currentPage + 1)"
                :disabled="currentPage === Math.ceil(filteredVMs.length / itemsPerPage)"
                class="p-1 rounded hover:bg-slate-700/50 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Next page"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Tab Content -->
      <div class="p-6 pt-0">
        <!-- Managed VMs Tab Content -->
        <div v-if="activeTab === 'managed'">
          <!-- VM Grid View -->
          <div v-if="filteredVMs.length > 0 && viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
            <FCard
              v-for="vm in paginatedVMs"
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
                      <p class="text-sm text-slate-400">{{ vm.os_type || 'Unknown OS' }}</p>
                    </div>
                  </div>
                  <span class="px-2 py-1 text-xs font-medium rounded bg-slate-700 text-slate-300">
                    {{ formatVMState(vm.state) }}
                  </span>
                </div>

                <!-- VM Stats -->
                <div class="grid grid-cols-2 gap-4 pt-2">
                  <div>
                    <p class="text-sm text-slate-400">CPU</p>
                    <p class="text-white font-medium">{{ vm.vcpu_count || 'N/A' }} vCPUs</p>
                  </div>
                  <div>
                    <p class="text-sm text-slate-400">Memory</p>
                    <p class="text-white font-medium">{{ formatMemory(vm.memoryMB || Math.round(vm.memory_bytes / 1024 / 1024)) }}</p>
                  </div>
                </div>

                <!-- VM Host -->
                <div class="pt-2 border-t border-slate-700">
                  <p class="text-sm text-slate-400">Host</p>
                  <p class="text-white font-medium">{{ getHostName(vm.hostId || '') }}</p>
                </div>
              </div>
            </FCard>
          </div>

          <!-- VM List View -->
          <div v-else-if="filteredVMs.length > 0 && viewMode === 'list'" class="mt-6 overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-slate-600">
                  <th class="text-left py-3 px-4 font-medium text-slate-300">Name</th>
                  <th class="text-left py-3 px-4 font-medium text-slate-300">State</th>
                  <th class="text-left py-3 px-4 font-medium text-slate-300">Host</th>
                  <th class="text-left py-3 px-4 font-medium text-slate-300">CPU</th>
                  <th class="text-left py-3 px-4 font-medium text-slate-300">Memory</th>
                  <th class="text-right py-3 px-4 font-medium text-slate-300">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="vm in paginatedVMs"
                  :key="vm.uuid"
                  class="border-b border-slate-700 hover:bg-slate-800/50 transition-colors cursor-pointer"
                  @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
                >
                  <td class="py-3 px-4">
                    <div class="flex items-center gap-3">
                      <div :class="[
                        'w-3 h-3 rounded-full',
                        getVMStatusColor(vm.state)
                      ]"></div>
                      <div>
                        <div class="text-white font-medium">{{ vm.name }}</div>
                        <div class="text-sm text-slate-400">{{ vm.os_type || 'Unknown OS' }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="py-3 px-4">
                    <span class="px-2 py-1 text-xs font-medium rounded bg-slate-700 text-slate-300">
                      {{ formatVMState(vm.state) }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-slate-300">{{ getHostName(vm.hostId || '') }}</td>
                  <td class="py-3 px-4 text-slate-300">{{ vm.vcpu_count || 'N/A' }} vCPUs</td>
                  <td class="py-3 px-4 text-slate-300">{{ formatMemory(vm.memoryMB || Math.round(vm.memory_bytes / 1024 / 1024)) }}</td>
                  <td class="py-3 px-4 text-right">
                    <div class="flex items-center justify-end gap-2">
                      <FButton
                        @click.stop="openVMConsole(vm)"
                        variant="ghost"
                        size="sm"
                        title="Open console"
                      >
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M3 5a2 2 0 012-2h10a2 2 0 012 2v8a2 2 0 01-2 2h-2.22l.123.489.804.804A1 1 0 0113 18H7a1 1 0 01-.707-1.707l.804-.804L7.22 15H5a2 2 0 01-2-2V5zm5.771 7H5V5h10v7H8.771z" clip-rule="evenodd" />
                        </svg>
                      </FButton>
                      <FButton
                        @click.stop="toggleVMState(vm)"
                        :variant="vm.state === 'ACTIVE' ? 'ghost' : 'primary'"
                        size="sm"
                        :title="vm.state === 'ACTIVE' ? 'Stop VM' : 'Start VM'"
                      >
                        <svg v-if="vm.state === 'ACTIVE'" class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
                        </svg>
                        <svg v-else class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                        </svg>
                      </FButton>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- VM Compact View -->
          <div v-else-if="filteredVMs.length > 0 && viewMode === 'compact'" class="space-y-1 mt-6">
            <div
              v-for="vm in paginatedVMs"
              :key="vm.uuid"
              class="group hover:bg-slate-800/30 rounded-lg transition-colors duration-200"
            >
              <FCard class="p-3 card-glow cursor-pointer" @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3 min-w-0 flex-1">
                    <div :class="[
                      'w-2 h-2 rounded-full flex-shrink-0',
                      getVMStatusColor(vm.state)
                    ]"></div>
                    <div class="min-w-0 flex-1">
                      <h3 class="text-sm font-medium text-white truncate">{{ vm.name }}</h3>
                      <p class="text-xs text-slate-400 truncate">{{ getHostName(vm.hostId || '') }} • {{ formatVMState(vm.state) }}</p>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 text-xs text-slate-400">
                    <span>{{ vm.vcpu_count || 'N/A' }} vCPUs</span>
                    <span>•</span>
                    <span>{{ formatMemory(vm.memoryMB || Math.round(vm.memory_bytes / 1024 / 1024)) }}</span>
                  </div>
                </div>
              </FCard>
            </div>
          </div>

          <!-- Empty State for Managed VMs -->
          <div v-else class="text-center py-12">
            <div class="flex justify-center mb-4">
              <svg class="w-16 h-16 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-2">No Virtual Machines Found</h3>
            <p class="text-slate-400 mb-4">No virtual machines match your current filters.</p>
            <div class="space-y-2 text-sm text-slate-500">
              <p>Try:</p>
              <ul class="list-disc list-inside space-y-1">
                <li>Clearing your search filters</li>
                <li>Checking different host connections</li>
                <li>Switching to the "Discovered VMs" tab to import new VMs</li>
              </ul>
            </div>
          </div>
        </div>
        
        <!-- Discovered VMs Tab Content -->
        <div v-else-if="activeTab === 'discovered'">
          <div class="flex items-center justify-between mb-6 mt-6">
            <div>
              <h2 class="text-xl font-semibold text-white mb-1">Discovered Virtual Machines</h2>
              <p class="text-slate-400 text-sm">Import VMs that were discovered on your hosts</p>
            </div>
            <div class="flex items-center gap-3">
              <FButton
                @click="handleRefreshDiscoveredVMs"
                variant="ghost"
                :disabled="!!hostStore.loading.refreshDiscoveredVMs"
                class="flex items-center gap-2"
                title="Refresh discovered VMs from all hosts"
              >
                <svg
                  :class="['w-4 h-4', !!hostStore.loading.refreshDiscoveredVMs ? 'animate-spin' : '']"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                {{ !!hostStore.loading.refreshDiscoveredVMs ? 'Refreshing...' : 'Refresh' }}
              </FButton>
              <FButton
                v-if="discoveredVMs.length > 0"
                @click="handleImportAllVMs"
                variant="primary"
                :disabled="!!hostStore.loading.refreshDiscoveredVMs"
                class="flex items-center gap-2"
              >
                <svg v-if="!hostStore.loading.refreshDiscoveredVMs" class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 4.414V13a1 1 0 11-2 0V4.414L7.707 5.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <svg v-else class="animate-spin w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H8a1 1 0 110 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H12a1 1 0 110-2h4a1 1 0 011 1v4a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
                </svg>
                {{ hostStore.loading.refreshDiscoveredVMs ? 'Importing...' : 'Import All VMs' }}
              </FButton>
            </div>
          </div>

          <!-- Discovered VMs Table -->
          <div class="space-y-4">
            <div v-if="filteredDiscoveredVMs.length === 0" class="text-center py-8 text-slate-400">
              <svg class="w-12 h-12 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zm9 1a1 1 0 010-2h4a1 1 0 011 1v4a1 1 0 01-2 0V6.414l-2.293 2.293a1 1 0 11-1.414-1.414L13.586 5H12z" clip-rule="evenodd" />
              </svg>
              <p class="text-lg font-medium mb-2">No discovered VMs found</p>
              <p class="text-sm">Try refreshing or check your host connections</p>
            </div>
            
            <div v-else class="overflow-x-auto">
              <table class="w-full">
                <thead>
                  <tr class="border-b border-slate-600">
                    <th class="text-left py-3 px-4 font-medium text-slate-300">
                      <input
                        type="checkbox"
                        :checked="selectedDiscoveredVMs.length === filteredDiscoveredVMs.length && filteredDiscoveredVMs.length > 0"
                        :indeterminate="selectedDiscoveredVMs.length > 0 && selectedDiscoveredVMs.length < filteredDiscoveredVMs.length"
                        @change="toggleSelectAllDiscovered"
                        class="rounded border-slate-500 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                      />
                    </th>
                    <th class="text-left py-3 px-4 font-medium text-slate-300">Name</th>
                    <th class="text-left py-3 px-4 font-medium text-slate-300">Host</th>
                    <th class="text-left py-3 px-4 font-medium text-slate-300">UUID</th>
                    <th class="text-right py-3 px-4 font-medium text-slate-300">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="vm in filteredDiscoveredVMs"
                    :key="`${vm.hostId}-${vm.path}`"
                    class="border-b border-slate-700 hover:bg-slate-800/50 transition-colors"
                  >
                    <td class="py-3 px-4">
                      <input
                        type="checkbox"
                        :value="`${vm.hostId}-${vm.path}`"
                        v-model="selectedDiscoveredVMs"
                        class="rounded border-slate-500 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                      />
                    </td>
                    <td class="py-3 px-4 text-white font-medium">{{ vm.name }}</td>
                    <td class="py-3 px-4 text-slate-300">{{ getHostName(vm.hostId) }}</td>
                    <td class="py-3 px-4 text-slate-400 font-mono text-sm">{{ vm.path }}</td>
                    <td class="py-3 px-4 text-right">
                      <FButton
                        @click="importDiscoveredVM(vm)"
                        variant="primary"
                        size="sm"
                        :disabled="!!importingVMs[`${vm.hostId}-${vm.path}`]"
                        class="flex items-center gap-2"
                      >
                        <svg v-if="!importingVMs[`${vm.hostId}-${vm.path}`]" class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 4.414V13a1 1 0 11-2 0V4.414L7.707 5.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                        </svg>
                        <svg v-else class="animate-spin w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                        </svg>
                        {{ importingVMs[`${vm.hostId}-${vm.path}`] ? 'Importing...' : 'Import' }}
                      </FButton>
                    </td>
                  </tr>
                </tbody>
              </table>
              
              <!-- Bulk Actions -->
              <div v-if="selectedDiscoveredVMs.length > 0" class="mt-4 p-4 bg-slate-800/50 rounded-lg border border-slate-600">
                <div class="flex items-center justify-between">
                  <div class="text-sm text-slate-300">
                    {{ selectedDiscoveredVMs.length }} VM{{ selectedDiscoveredVMs.length === 1 ? '' : 's' }} selected
                  </div>
                  <div class="flex gap-2">
                    <FButton
                      @click="clearSelectedDiscovered"
                      variant="ghost"
                      size="sm"
                    >
                      Clear Selection
                    </FButton>
                    <FButton
                      @click="importSelectedDiscoveredVMs"
                      variant="primary"
                      size="sm"
                      :disabled="bulkImporting"
                    >
                      <svg v-if="!bulkImporting" class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 4.414V13a1 1 0 11-2 0V4.414L7.707 5.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                      </svg>
                      <svg v-else class="w-4 h-4 mr-2 animate-spin" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      {{ bulkImporting ? 'Importing...' : `Import Selected (${selectedDiscoveredVMs.length})` }}
                    </FButton>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </FCard>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import { useUserPreferences } from '@/composables/useUserPreferences'

// UI Components
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import { getConsoleRoute } from '@/utils/console'

const router = useRouter()
const hostStore = useHostStore()
const vmStore = useVMStore()
const { vmListPreferences } = useUserPreferences()

// Search state
const searchQuery = ref('')
const showFilterHelp = ref(false)
const showFilterDropdown = ref(false)
const showViewSettings = ref(false)
const itemsPerPage = ref<number | 'all'>(10)
const currentPage = ref(1)

// Discovered VMs state
const selectedDiscoveredVMs = ref<string[]>([])
const importingVMs = ref<Record<string, boolean>>({})
const bulkImporting = ref(false)

// Use preferences for persistent state
const viewMode = computed({
  get: () => vmListPreferences.viewMode,
  set: (value) => { vmListPreferences.viewMode = value }
})

// Active tab
const activeTab = ref<'managed' | 'discovered'>('managed')

// Host filter state
const selectedHostId = ref<string>('all')

// VM state filter
const selectedState = ref<string>('all')

// Computed properties
const allHosts = computed(() => hostStore.hosts)
const discoveredVMs = computed(() => hostStore.allDiscoveredVMs || [])
const totalVMs = computed(() => vmStore.vms.length)
const activeVMs = computed(() => vmStore.vms.filter(vm => vm.state === 'ACTIVE').length)

// Filter helpers
const hasActiveFilters = computed(() => {
  return selectedState.value !== 'all' || selectedHostId.value !== 'all' || searchQuery.value.length > 0
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (selectedState.value !== 'all') count++
  if (selectedHostId.value !== 'all') count++
  if (searchQuery.value) count++
  return count
})

// Search parsing for advanced syntax
const searchTerms = computed(() => {
  if (!searchQuery.value) return { query: '', filters: {} }
  
  const filters: Record<string, string> = {}
  let cleanQuery = searchQuery.value
  
  // Parse advanced search syntax: key:value
  const advancedPattern = /(\\w+):(\\w+)/g
  let match
  
  while ((match = advancedPattern.exec(searchQuery.value)) !== null) {
    const [fullMatch, key, value] = match
    filters[key.toLowerCase()] = value.toLowerCase()
    cleanQuery = cleanQuery.replace(fullMatch, '').trim()
  }
  
  return { query: cleanQuery, filters }
})

// Filtered VMs with enhanced search
const filteredVMs = computed(() => {
  let vms = vmStore.vms
  
  // Apply state filter
  if (selectedState.value !== 'all') {
    vms = vms.filter(vm => {
      if (selectedState.value === 'running') return vm.state === 'ACTIVE'
      if (selectedState.value === 'stopped') return vm.state === 'STOPPED'
      if (selectedState.value === 'paused') return vm.state === 'PAUSED' || vm.state === 'SUSPENDED'
      if (selectedState.value === 'error') return vm.state === 'UNKNOWN'
      return true
    })
  }
  
      // Apply host filter
  if (selectedHostId.value !== 'all') {
    vms = vms.filter(vm => vm.hostId === selectedHostId.value)
  }
  
  // Apply search query
  if (searchQuery.value) {
    const { query, filters } = searchTerms.value
    
    vms = vms.filter(vm => {
      // Check advanced filters first
      for (const [key, value] of Object.entries(filters)) {
        if (key === 'state') {
          if (value === 'running' && vm.state !== 'ACTIVE') return false
          if (value === 'stopped' && vm.state !== 'STOPPED') return false
          if (value === 'paused' && !['PAUSED', 'SUSPENDED'].includes(vm.state)) return false
          if (value === 'error' && vm.state !== 'UNKNOWN') return false
        }
        if (key === 'host') {
          const hostName = getHostName(vm.hostId || '').toLowerCase()
          if (!hostName.includes(value)) return false
        }
        if (key === 'cpu' || key === 'vcpu' || key === 'vcpus') {
          if (!vm.vcpu_count?.toString().includes(value)) return false
        }
        if (key === 'memory' || key === 'mem' || key === 'ram') {
          const memoryMB = vm.memoryMB || Math.round(vm.memory_bytes / 1024 / 1024) || 0
          const memoryGB = Math.round(memoryMB / 1024)
          if (!memoryGB.toString().includes(value) && !memoryMB.toString().includes(value)) return false
        }
      }
      
      // Then check regular text search
      if (query) {
        const searchText = query.toLowerCase()
        return (
          vm.name.toLowerCase().includes(searchText) ||
          (vm.os_type && vm.os_type.toLowerCase().includes(searchText)) ||
          getHostName(vm.hostId || '').toLowerCase().includes(searchText) ||
          formatVMState(vm.state).toLowerCase().includes(searchText)
        )
      }
      
      return true
    })
  }
  
  return vms
})

// Paginated VMs
const paginatedVMs = computed(() => {
  if (itemsPerPage.value === 'all') return filteredVMs.value
  
  const start = (currentPage.value - 1) * (itemsPerPage.value as number)
  const end = start + (itemsPerPage.value as number)
  return filteredVMs.value.slice(start, end)
})

// Filtered discovered VMs
const filteredDiscoveredVMs = computed(() => {
  const allDiscovered = hostStore.allDiscoveredVMs || []
  if (!searchQuery.value) return allDiscovered
  
  const query = searchQuery.value.toLowerCase()
  return allDiscovered.filter((vm: any) => 
    vm.name.toLowerCase().includes(query) ||
    getHostName(vm.host_id).toLowerCase().includes(query)
  )
})

// Active filters for display
const activeFilters = computed(() => {
  const filters: string[] = []
  
  if (selectedState.value !== 'all') {
    filters.push(`State: ${selectedState.value}`)
  }
  
  if (selectedHostId.value !== 'all') {
    const hostName = getHostName(selectedHostId.value)
    filters.push(`Host: ${hostName}`)
  }
  
  if (searchQuery.value) {
    const { query, filters: advancedFilters } = searchTerms.value
    if (query) filters.push(`Search: "${query}"`)
    
    for (const [key, value] of Object.entries(advancedFilters)) {
      filters.push(`${key}: ${value}`)
    }
  }
  
  return filters
})

// Utility functions
const getHostName = (hostId: string): string => {
  const host = allHosts.value.find(h => h.id === hostId)
  return host?.name || 'Unknown Host'
}

const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500'
    case 'INACTIVE':
    case 'SHUTOFF': return 'bg-slate-500'
    case 'PAUSED':
    case 'SUSPENDED': return 'bg-yellow-500'
    case 'ERROR':
    case 'CRASHED': return 'bg-red-500'
    default: return 'bg-slate-500'
  }
}

const formatVMState = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'Running'
    case 'INACTIVE': return 'Stopped'
    case 'SHUTOFF': return 'Shut Off'
    case 'PAUSED': return 'Paused'
    case 'SUSPENDED': return 'Suspended'
    case 'ERROR': return 'Error'
    case 'CRASHED': return 'Crashed'
    default: return state
  }
}

const formatMemory = (memory?: number): string => {
  if (!memory) return 'N/A'
  if (memory >= 1024) {
    return `${(memory / 1024).toFixed(1)} GB`
  }
  return `${memory} MB`
}

// Filter management
const clearFilters = () => {
  searchQuery.value = ''
  selectedState.value = 'all'
  selectedHostId.value = 'all'
  showFilterDropdown.value = false
}

// VM actions
const toggleVMState = async (vm: any) => {
  try {
    if (vm.state === 'ACTIVE') {
      await vmStore.stopVM(vm.hostId, vm.name)
    } else {
      await vmStore.startVM(vm.hostId, vm.name)
    }
  } catch (error) {
    console.error('Failed to toggle VM state:', error)
  }
}

const openVMConsole = (vm: any) => {
  const consoleRoute = getConsoleRoute(vm.hostId || '', vm.name, vm)
  if (consoleRoute) {
    window.open(consoleRoute, '_blank')
  }
}

// Discovered VMs actions
const handleRefreshDiscoveredVMs = async () => {
  try {
    await hostStore.refreshAllDiscoveredVMs()
  } catch (error) {
    console.error('Failed to refresh discovered VMs:', error)
  }
}

const handleImportAllVMs = async () => {
  const allDiscovered = hostStore.allDiscoveredVMs || []
  if (allDiscovered.length === 0) return
  
  bulkImporting.value = true
  try {
    for (const vm of allDiscovered) {
      // Note: Need to check if import method exists in hostStore
      console.log('Would import VM:', vm.name, 'from host:', vm.host_id)
    }
    await vmStore.fetchVMs()
  } catch (error) {
    console.error('Failed to import all VMs:', error)
  } finally {
    bulkImporting.value = false
  }
}

const importDiscoveredVM = async (vm: any) => {
  const key = `${vm.host_id}-${vm.domain_uuid}`
  importingVMs.value[key] = true
  
  try {
    // Note: Need to check if import method exists in hostStore
    console.log('Would import VM:', vm.name, 'from host:', vm.host_id)
    await vmStore.fetchVMs()
  } catch (error) {
    console.error('Failed to import VM:', error)
  } finally {
    delete importingVMs.value[key]
  }
}

const toggleSelectAllDiscovered = () => {
  const allDiscovered = hostStore.allDiscoveredVMs || []
  if (selectedDiscoveredVMs.value.length === allDiscovered.length) {
    selectedDiscoveredVMs.value = []
  } else {
    selectedDiscoveredVMs.value = allDiscovered.map((vm: any) => `${vm.host_id}-${vm.domain_uuid}`)
  }
}

const clearSelectedDiscovered = () => {
  selectedDiscoveredVMs.value = []
}

const importSelectedDiscoveredVMs = async () => {
  if (selectedDiscoveredVMs.value.length === 0) return
  
  bulkImporting.value = true
  try {
    const allDiscovered = hostStore.allDiscoveredVMs || []
    for (const vmKey of selectedDiscoveredVMs.value) {
      const [hostId, domainUuid] = vmKey.split('-')
      const vm = allDiscovered.find((v: any) => v.host_id === hostId && v.domain_uuid === domainUuid)
      if (vm) {
        console.log('Would import VM:', vm.name, 'from host:', vm.host_id)
      }
    }
    await vmStore.fetchVMs()
    selectedDiscoveredVMs.value = []
  } catch (error) {
    console.error('Failed to import selected VMs:', error)
  } finally {
    bulkImporting.value = false
  }
}

// Lifecycle
onMounted(async () => {
  await Promise.all([
    hostStore.fetchHosts(),
    vmStore.fetchVMs(),
    hostStore.fetchGlobalDiscoveredVMs()
  ])
})

// Watch for search changes to reset pagination
watch(searchQuery, () => {
  currentPage.value = 1
})

watch([selectedState, selectedHostId], () => {
  currentPage.value = 1
})

// Close dropdowns when clicking outside
const handleClickOutside = (event: Event) => {
  const target = event.target as Element
  if (!target.closest('.filter-dropdown-container')) {
    showFilterDropdown.value = false
  }
  if (!target.closest('.view-settings-container')) {
    showViewSettings.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})
</script>