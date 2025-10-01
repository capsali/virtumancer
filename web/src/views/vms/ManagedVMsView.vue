<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Managed Virtual Machines</h1>
        <p class="text-slate-400 mt-2">Monitor and control your managed VMs</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="text-right">
          <div class="text-2xl font-bold text-white">{{ filteredVMs.length }}</div>
          <div class="text-sm text-slate-400">Total VMs</div>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-green-400">{{ activeVMs }}</div>
          <div class="text-sm text-slate-400">Active</div>
        </div>
        <FButton
          @click="openCreateVMModal"
          variant="primary"
          size="lg"
          class="ml-6 px-4 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
          title="Create new virtual machine"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
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
            {{ filteredVMs.length }} of {{ totalVMs }} VMs
          </div>
        </div>
        
        <!-- Search Bar with integrated Filter Button -->
        <div class="relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search VMs... (try 'host:myhost', 'cpu:>2', 'memory:>4GB')"
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
            <div><code class="text-blue-400">host:myhost</code> - Filter by host</div>
            <div><code class="text-blue-400">cpu:>2</code> - CPUs greater than 2</div>
            <div><code class="text-blue-400">memory:>4GB</code> - Memory over 4GB</div>
            <div><code class="text-blue-400">state:active</code> - Filter by state</div>
          </div>
        </div>

        <!-- Expandable Filter Section -->
        <div v-if="showFilterDropdown" class="mt-4 space-y-4 p-4 bg-slate-800/30 rounded-lg border border-slate-600/30 animate-slideDown">
          <!-- Quick Filters -->
          <div>
            <h4 class="text-sm font-medium text-slate-300 mb-3">Quick Filters</h4>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="filter in quickFilters"
                :key="filter.value"
                @click="quickFilter(filter.value)"
                :class="[
                  'p-2 text-xs rounded-lg transition-all hover:scale-105 flex items-center justify-center gap-1',
                  statusFilter === filter.value ? 'bg-green-500/20 text-green-400 border border-green-500/30' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                ]"
              >
                <span>{{ filter.icon }}</span>
                <span class="hidden sm:inline">{{ filter.label }}</span>
              </button>
            </div>
          </div>

          <!-- Detailed Filters -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">State</label>
              <select
                v-model="statusFilter"
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All States</option>
                <option value="ACTIVE">Active</option>
                <option value="STOPPED">Stopped</option>
                <option value="PAUSED">Paused</option>
                <option value="ERROR">Error</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Host</label>
              <select
                v-model="hostFilter"
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All Hosts</option>
                <option v-for="host in hosts" :key="host.id" :value="host.id">{{ host.name || host.uri }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">OS Type</label>
              <select
                v-model="osTypeFilter"
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All OS Types</option>
                <option value="linux">Linux</option>
                <option value="windows">Windows</option>
                <option value="other">Other</option>
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

    <!-- VM List Card -->
    <FCard class="card-glow">
      <div class="border-b border-slate-700/50">
        <div class="flex items-center justify-between px-6 pt-6 pb-4">
          <h2 class="text-lg font-semibold text-white">Virtual Machines ({{ filteredVMs.length }})</h2>
          
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
                      itemsPerPage === option.value ? 'text-blue-400 bg-blue-500/10' : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
                    ]"
                  >
                    {{ option.label }}
                  </button>
                </div>
              </div>
            </div>

            <!-- View Mode Toggle -->
            <button
              @click="cycleViewMode"
              class="p-2 text-slate-400 hover:text-white transition-all duration-200 rounded hover:bg-slate-700/50"
              :title="`Current view: ${viewMode === 'cards' ? 'Cards' : viewMode === 'compact' ? 'Compact' : 'List'}. Click to change.`"
            >
              <!-- Cards View Icon -->
              <svg v-if="viewMode === 'cards'" class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
              <!-- Compact View Icon -->
              <svg v-else-if="viewMode === 'compact'" class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M3 7a1.5 1.5 0 011.5-1.5h11a1.5 1.5 0 110 3h-11A1.5 1.5 0 013 7zM3 13a1.5 1.5 0 011.5-1.5h11a1.5 1.5 0 110 3h-11A1.5 1.5 0 013 13z" />
              </svg>
              <!-- List View Icon -->
              <svg v-else class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- VM Content -->
      <div class="p-6">
        <!-- Cards View -->
        <div v-if="viewMode === 'cards'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <VMCardSimple
              v-for="vm in paginatedVMs"
              :key="vm.uuid"
              :vm="vm"
              :host-id="vm.hostId || ''"
              @action="handleVMAction"
              @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
              class="cursor-pointer hover:scale-[1.02] transition-transform duration-200"
            />
        </div>

        <!-- List View - Modern Card Style -->
        <div v-else-if="viewMode === 'list'" class="space-y-2">
          <!-- Sorting Header -->
          <div class="grid grid-cols-12 gap-4 px-5 py-2 bg-slate-800/20 rounded-lg border border-slate-600/20 backdrop-blur-sm text-sm font-medium text-slate-300">
            <div class="col-span-1"></div> <!-- Status column -->
            <button
              @click="handleSort('name')"
              class="col-span-3 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none text-left"
            >
              Name {{ getSortIcon('name') }}
            </button>
            <button
              @click="handleSort('host')"
              class="col-span-2 hidden md:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Host {{ getSortIcon('host') }}
            </button>
            <button
              @click="handleSort('status')"
              class="col-span-2 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Status {{ getSortIcon('status') }}
            </button>
            <button
              @click="handleSort('vcpus')"
              class="col-span-2 hidden lg:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Resources {{ getSortIcon('vcpus') }}
            </button>
            <div class="col-span-2 text-center">
              Actions
            </div>
          </div>

          <!-- VM List Items -->
          <div
            v-for="vm in sortedPaginatedVMs"
            :key="vm.uuid"
            class="group relative bg-slate-800/30 hover:bg-slate-700/40 rounded-lg p-4 border border-slate-600/20 hover:border-slate-500/40 transition-all duration-300 cursor-pointer backdrop-blur-sm"
            @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
          >
            <!-- Grid Layout to Match Header -->
            <div class="grid grid-cols-12 gap-4 items-center">
              <!-- Status Orb -->
              <div class="col-span-1">
                <div :class="[
                  'w-3 h-3 rounded-full transition-all duration-300',
                  vm.state === 'ACTIVE' ? 'bg-green-400 animate-pulse shadow-lg shadow-green-400/50' :
                  vm.state === 'STOPPED' ? 'bg-red-400' :
                  vm.state === 'ERROR' ? 'bg-red-600' :
                  'bg-yellow-400'
                ]"></div>
              </div>

              <!-- VM Name & OS -->
              <div class="col-span-3 min-w-0">
                <h3 class="text-base font-semibold text-white truncate group-hover:text-blue-300 transition-colors">
                  {{ vm.name || 'Unnamed VM' }}
                </h3>
                <p class="text-xs text-slate-400 truncate">
                  {{ vm.os_type || 'Unknown OS' }}
                </p>
              </div>

              <!-- Host Info -->
              <div class="col-span-2 hidden md:flex items-center gap-2 text-slate-300 min-w-0">
                <svg class="w-3 h-3 text-slate-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <span class="text-sm truncate">{{ getHostName(vm.hostId || '') }}</span>
              </div>

              <!-- Status Badge -->
              <div class="col-span-2 flex items-center">
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium backdrop-blur-sm border',
                  vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-300 border-green-500/30' :
                  vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-300 border-red-500/30' :
                  vm.state === 'ERROR' ? 'bg-red-600/20 text-red-200 border-red-600/30' :
                  'bg-yellow-500/20 text-yellow-300 border-yellow-500/30'
                ]">
                  {{ vm.state }}
                </span>
              </div>

              <!-- Resource Information -->
              <div class="col-span-2 hidden lg:flex items-center gap-3 text-xs text-slate-400">
                <div class="flex items-center gap-1">
                  <div class="w-2 h-2 rounded-full bg-blue-400"></div>
                  <span>{{ vm.vcpu_count || 'N/A' }}</span>
                </div>
                <div class="flex items-center gap-1">
                  <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                  <span>{{ formatMemory(vm.memory_bytes || 0) }}</span>
                </div>
              </div>

              <!-- Action Buttons Column -->
              <div class="col-span-2 flex items-center justify-center gap-1" @click.stop>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click="handleVMAction('start', vm)"
                  :disabled="vm.state === 'ACTIVE'"
                  class="p-2 text-green-400 hover:bg-green-500/20 hover:text-green-300 transition-all duration-200 rounded-lg"
                  title="Start VM"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click="handleVMAction('stop', vm)"
                  :disabled="vm.state === 'STOPPED'"
                  class="p-2 text-red-400 hover:bg-red-500/20 hover:text-red-300 transition-all duration-200 rounded-lg"
                  title="Stop VM"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click="openVMConsole(vm)"
                  :disabled="vm.state !== 'ACTIVE'"
                  class="p-2 text-blue-400 hover:bg-blue-500/20 hover:text-blue-300 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-200 rounded-lg"
                  title="Console"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
              </div>
            </div>

            <!-- Mobile Resource Info (shown on smaller screens) -->
            <div class="lg:hidden mt-3 pt-3 border-t border-slate-600/20 flex items-center gap-4 text-sm text-slate-400">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-blue-400"></div>
                <span>{{ vm.vcpu_count || 'N/A' }} vCPUs</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span>{{ formatMemory(vm.memory_bytes || 0) }}</span>
              </div>
              <div class="md:hidden flex items-center gap-2">
                <svg class="w-3 h-3 text-slate-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <span>{{ getHostName(vm.hostId || '') }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Compact View - More Expanded -->
        <div v-else class="space-y-4">
          <div
            v-for="vm in paginatedVMs"
            :key="vm.uuid"
            class="group bg-gradient-to-r from-slate-800/40 via-slate-800/30 to-slate-800/40 rounded-xl p-6 border border-slate-600/30 hover:border-slate-500/50 hover:shadow-xl hover:shadow-slate-900/20 transition-all duration-300 cursor-pointer backdrop-blur-sm"
            @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
          >
            <div class="flex items-start gap-6">
              <!-- Enhanced VM Icon with Status -->
              <div class="relative flex-shrink-0">
                <div :class="[
                  'w-16 h-16 rounded-2xl flex items-center justify-center shadow-lg border-2 transition-all duration-300 group-hover:scale-105',
                  vm.state === 'ACTIVE' ? 'bg-gradient-to-br from-green-500/20 to-emerald-600/20 border-green-500/40 shadow-green-500/20' :
                  vm.state === 'STOPPED' ? 'bg-gradient-to-br from-slate-500/20 to-slate-600/20 border-slate-500/40 shadow-slate-500/20' :
                  vm.state === 'ERROR' ? 'bg-gradient-to-br from-red-500/20 to-red-600/20 border-red-500/40 shadow-red-500/20' :
                  'bg-gradient-to-br from-yellow-500/20 to-amber-600/20 border-yellow-500/40 shadow-yellow-500/20'
                ]">
                  <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                  </svg>
                </div>
                
                <!-- Status Indicator with Pulse -->
                <div :class="[
                  'absolute -top-1 -right-1 w-5 h-5 rounded-full border-3 border-slate-800 flex items-center justify-center',
                  vm.state === 'ACTIVE' ? 'bg-green-400 animate-pulse shadow-lg shadow-green-400/50' :
                  vm.state === 'STOPPED' ? 'bg-red-400' :
                  vm.state === 'ERROR' ? 'bg-red-600' :
                  'bg-yellow-400'
                ]">
                  <div v-if="vm.state === 'ACTIVE'" class="w-2 h-2 bg-white rounded-full"></div>
                </div>
              </div>
              
              <!-- VM Information Section -->
              <div class="flex-1 min-w-0 space-y-3">
                <!-- VM Title and Status -->
                <div class="flex items-center justify-between">
                  <div class="min-w-0 flex-1">
                    <h3 class="text-xl font-bold text-white truncate group-hover:text-blue-300 transition-colors">
                      {{ vm.name || 'Unnamed VM' }}
                    </h3>
                    <p class="text-sm text-slate-400 truncate mt-1">
                      {{ vm.description || vm.os_type || 'Virtual Machine' }}
                    </p>
                  </div>
                  
                  <!-- Enhanced Status Badge -->
                  <span :class="[
                    'inline-flex items-center px-3 py-1.5 rounded-full text-sm font-semibold border backdrop-blur-sm ml-4',
                    vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-300 border-green-500/40 shadow-lg shadow-green-500/20' :
                    vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-300 border-red-500/40' :
                    vm.state === 'ERROR' ? 'bg-red-600/20 text-red-200 border-red-600/40' :
                    'bg-yellow-500/20 text-yellow-300 border-yellow-500/40'
                  ]">
                    {{ vm.state }}
                  </span>
                </div>
                
                <!-- Resource Grid -->
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <!-- Host Info -->
                  <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-2">
                    <svg class="w-4 h-4 text-blue-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                    </svg>
                    <div class="min-w-0">
                      <div class="text-xs text-slate-400">Host</div>
                      <div class="text-sm font-medium truncate">{{ getHostName(vm.hostId || '') }}</div>
                    </div>
                  </div>
                  
                  <!-- CPU Info -->
                  <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-2">
                    <svg class="w-4 h-4 text-purple-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M3 5a2 2 0 012-2h10a2 2 0 012 2v8a2 2 0 01-2 2h-2.22l.123.489.804.804A1 1 0 0113 18H7a1 1 0 01-.707-1.707l.804-.804L7.22 15H5a2 2 0 01-2-2V5zm5.771 7H5V5h10v7H8.771z" clip-rule="evenodd" />
                    </svg>
                    <div>
                      <div class="text-xs text-slate-400">vCPUs</div>
                      <div class="text-sm font-medium">{{ vm.vcpu_count || 'N/A' }}</div>
                    </div>
                  </div>
                  
                  <!-- Memory Info -->
                  <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-2">
                    <svg class="w-4 h-4 text-emerald-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
                      <path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
                      <path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
                    </svg>
                    <div>
                      <div class="text-xs text-slate-400">Memory</div>
                      <div class="text-sm font-medium">{{ formatMemory(vm.memory_bytes || 0) }}</div>
                    </div>
                  </div>
                  
                  <!-- OS Type -->
                  <div v-if="vm.os_type" class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-2">
                    <svg class="w-4 h-4 text-amber-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M2 5a2 2 0 012-2h8a2 2 0 012 2v10a2 2 0 002 2H4a2 2 0 01-2-2V5zm3 1h6v4H5V6zm6 6H5v2h6v-2z" clip-rule="evenodd" />
                      <path d="M15 7h1a2 2 0 012 2v5.5a1.5 1.5 0 01-3 0V9a1 1 0 00-1-1h-1v4.5a1.5 1.5 0 01-3 0V7z" />
                    </svg>
                    <div>
                      <div class="text-xs text-slate-400">OS</div>
                      <div class="text-sm font-medium">{{ vm.os_type }}</div>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- Enhanced Action Buttons -->
              <div class="flex flex-col gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300 transform translate-x-4 group-hover:translate-x-0">
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="handleVMAction('start', vm)"
                  :disabled="vm.state === 'ACTIVE'"
                  class="p-3 text-green-400 hover:bg-green-500/20 hover:text-green-300 transition-all duration-200 rounded-xl"
                  title="Start VM"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="handleVMAction('stop', vm)"
                  :disabled="vm.state === 'STOPPED'"
                  class="p-3 text-red-400 hover:bg-red-500/20 hover:text-red-300 transition-all duration-200 rounded-xl"
                  title="Stop VM"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="outline"
                  size="sm"
                  @click.stop="openVMConsole(vm)"
                  :disabled="vm.state !== 'ACTIVE'"
                  class="p-3 text-blue-400 hover:bg-blue-500/20 border-blue-500/30 rounded-xl"
                  title="Console"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="toggleVMDropdown(vm.uuid)"
                  class="p-3 text-slate-400 hover:text-white hover:bg-slate-600/30 transition-all duration-200 rounded-xl"
                  title="More Actions"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
                  </svg>
                </FButton>
              </div>
            </div>
            
            <!-- Dropdown Menu for Compact View -->
            <div 
              v-if="activeVMDropdown === vm.uuid"
              v-click-away="() => activeVMDropdown = null"
              class="absolute right-4 top-full mt-1 w-48 bg-slate-800/95 border border-slate-600/50 rounded-lg shadow-lg z-50 backdrop-blur-sm"
            >
              <div class="py-1">
                <button
                  @click.stop="viewVMDetails(vm)"
                  class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M2 10a8 8 0 018-8v8h8a8 8 0 11-16 0z" />
                    <path d="M12 2.252A8.014 8.014 0 0117.748 8H12V2.252z" />
                  </svg>
                  View Details
                </button>
                <button
                  @click.stop="cloneVM(vm)"
                  class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M8 2a1 1 0 000 2h2a1 1 0 100-2H8z" />
                    <path d="M3 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v6h-4.586l1.293-1.293a1 1 0 00-1.414-1.414l-3 3a1 1 0 000 1.414l3 3a1 1 0 001.414-1.414L10.414 13H15v3a2 2 0 01-2 2H5a2 2 0 01-2-2V5zM15 11.586l-3-3a1 1 0 00-1.414 1.414L11.586 11H9a1 1 0 100 2h2.586l-1 1a1 1 0 001.414 1.414l3-3z" />
                  </svg>
                  Clone VM
                </button>
                <button
                  @click.stop="exportVM(vm)"
                  class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd" />
                  </svg>
                  Export
                </button>
                <hr class="border-slate-600/50 my-1">
                <button
                  @click.stop="deleteVM(vm)"
                  class="w-full px-4 py-2 text-left text-sm text-red-400 hover:bg-red-500/10 hover:text-red-300 transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" clip-rule="evenodd" />
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                  </svg>
                  Delete VM
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="filteredVMs.length === 0" class="text-center py-12">
          <div class="w-16 h-16 bg-slate-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-white mb-2">No VMs Found</h3>
          <p class="text-slate-400 mb-4">No virtual machines match your current filters.</p>
          <button
            @click="clearAllFilters"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-all"
          >
            Clear Filters
          </button>
        </div>
      </div>
    </FCard>

    <!-- VM Action Modals -->
    <CreateVMModalEnhanced
      :open="showCreateVMModal"
      @close="showCreateVMModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useVMStore } from '@/stores/vmStore'
import { useHostStore } from '@/stores/hostStore'
import { useUserPreferences, type ViewMode } from '@/composables/useUserPreferences'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import FButton from '@/components/ui/FButton.vue'
import VMCardSimple from '@/components/vm/VMCardSimple.vue'
import CreateVMModalEnhanced from '@/components/modals/CreateVMModalEnhanced.vue'
import { vClickAway } from '@/directives/clickAway'

// Router and stores
const router = useRouter()
const vmStore = useVMStore()
const hostStore = useHostStore()
const { preferences } = useUserPreferences()

// Modal states
const showCreateVMModal = ref(false)

// Search and filters
const searchQuery = ref('')
const statusFilter = ref('all')
const hostFilter = ref('all')
const osTypeFilter = ref('all')
const showFilterDropdown = ref(false)
const showFilterHelp = ref(false)

// View settings - use persistent preferences
const viewMode = ref<ViewMode>(preferences.value.vmList.viewMode)
const itemsPerPage = ref<number | 'all'>(12)
const showItemsDropdown = ref(false)
const currentPage = ref(1)

// Watch viewMode changes and save to preferences
watch(viewMode, (newMode) => {
  preferences.value.vmList.viewMode = newMode
})

// Sorting and dropdowns
const sortField = ref<string>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')
const activeVMDropdown = ref<string | null>(null)

// Quick filters data
const quickFilters = [
  { value: 'ACTIVE', label: 'Running', icon: '🟢', color: 'green' },
  { value: 'STOPPED', label: 'Stopped', icon: '🔴', color: 'red' },
  { value: 'PAUSED', label: 'Paused', icon: '⏸️', color: 'yellow' },
  { value: 'ERROR', label: 'Error', icon: '❌', color: 'red' }
]

const itemsPerPageOptions: Array<{value: number | 'all', label: string}> = [
  { value: 5, label: '5' },
  { value: 10, label: '10' },
  { value: 15, label: '15' },
  { value: 50, label: '50' },
  { value: 100, label: '100' },
  { value: 'all', label: 'All' }
]

// Computed properties
const vms = computed(() => vmStore.vms)
const hosts = computed(() => hostStore.hosts)
const totalVMs = computed(() => vms.value.length)
const activeVMs = computed(() => vms.value.filter(vm => vm.state === 'ACTIVE').length)

const filteredVMs = computed(() => {
  let filtered = vms.value

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(vm => 
      vm.name?.toLowerCase().includes(query) ||
      vm.uuid.toLowerCase().includes(query) ||
      hosts.value.find(h => h.id === vm.hostId)?.name?.toLowerCase().includes(query)
    )
  }

  // Status filter
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(vm => vm.state === statusFilter.value)
  }

  // Host filter
  if (hostFilter.value !== 'all') {
    filtered = filtered.filter(vm => vm.hostId === hostFilter.value)
  }

  // OS Type filter
  if (osTypeFilter.value !== 'all') {
    filtered = filtered.filter(vm => vm.os_type === osTypeFilter.value)
  }

  return filtered
})

const paginatedVMs = computed(() => {
  if (itemsPerPage.value === 'all') {
    return filteredVMs.value
  }
  
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredVMs.value.slice(start, end)
})

const sortedPaginatedVMs = computed(() => {
  const sorted = [...paginatedVMs.value]
  
  sorted.sort((a, b) => {
    let aValue: any, bValue: any
    
    switch (sortField.value) {
      case 'name':
        aValue = a.name || ''
        bValue = b.name || ''
        break
      case 'host':
        aValue = getHostName(a.hostId || '')
        bValue = getHostName(b.hostId || '')
        break
      case 'status':
        aValue = a.state || ''
        bValue = b.state || ''
        break
      case 'vcpus':
        aValue = a.vcpu_count || 0
        bValue = b.vcpu_count || 0
        break
      case 'memory':
        aValue = a.memory_bytes || 0
        bValue = b.memory_bytes || 0
        break
      default:
        return 0
    }
    
    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }
    
    if (sortDirection.value === 'asc') {
      return aValue < bValue ? -1 : aValue > bValue ? 1 : 0
    } else {
      return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
    }
  })
  
  return sorted
})

const hasActiveFilters = computed(() => {
  return searchQuery.value !== '' ||
         statusFilter.value !== 'all' ||
         hostFilter.value !== 'all' ||
         osTypeFilter.value !== 'all'
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (searchQuery.value) count++
  if (statusFilter.value !== 'all') count++
  if (hostFilter.value !== 'all') count++
  if (osTypeFilter.value !== 'all') count++
  return count
})

// Methods
const toggleFilterDropdown = () => {
  showFilterDropdown.value = !showFilterDropdown.value
}

const toggleItemsDropdown = () => {
  showItemsDropdown.value = !showItemsDropdown.value
}

const quickFilter = (filterValue: string) => {
  if (statusFilter.value === filterValue) {
    statusFilter.value = 'all'
  } else {
    statusFilter.value = filterValue
  }
}

const clearAllFilters = () => {
  searchQuery.value = ''
  statusFilter.value = 'all'
  hostFilter.value = 'all'
  osTypeFilter.value = 'all'
  showFilterDropdown.value = false
}

const cycleViewMode = () => {
  const modes: ViewMode[] = ['cards', 'compact', 'list']
  const currentIndex = modes.indexOf(viewMode.value)
  const nextIndex = (currentIndex + 1) % modes.length
  const nextMode = modes[nextIndex]
  if (nextMode) {
    viewMode.value = nextMode
  }
}

const setItemsPerPage = (value: number | 'all') => {
  itemsPerPage.value = value
  currentPage.value = 1
  showItemsDropdown.value = false
}

const openCreateVMModal = () => {
  showCreateVMModal.value = true
}

const handleVMAction = (action: string, vm: any) => {
  console.log('VM Action:', action, vm)
}

const getHostName = (hostId: string): string => {
  const host = hosts.value.find(h => h.id === hostId)
  return host?.name || host?.uri || 'Unknown Host'
}

const formatMemory = (bytes: number): string => {
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(1)} GB`
}

// Sorting methods
const handleSort = (field: string) => {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

const getSortIcon = (field: string): string => {
  if (sortField.value !== field) return '↕'
  return sortDirection.value === 'asc' ? '↑' : '↓'
}

// VM status color helper
const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE':
      return 'bg-green-500'
    case 'STOPPED':
      return 'bg-red-500'
    case 'PAUSED':
      return 'bg-yellow-500'
    case 'ERROR':
      return 'bg-red-600'
    default:
      return 'bg-slate-500'
  }
}

// VM action methods
const toggleVMDropdown = (vmUuid: string) => {
  activeVMDropdown.value = activeVMDropdown.value === vmUuid ? null : vmUuid
}

const viewVMDetails = (vm: any) => {
  router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)
  activeVMDropdown.value = null
}

const openVMConsole = (vm: any) => {
  // Open console in new window/tab
  const consoleUrl = `/spice/${vm.hostId}/${vm.name}`
  window.open(consoleUrl, '_blank')
  activeVMDropdown.value = null
}

const cloneVM = (vm: any) => {
  console.log('Clone VM:', vm)
  // TODO: Implement VM cloning
  activeVMDropdown.value = null
}

const exportVM = (vm: any) => {
  console.log('Export VM:', vm)
  // TODO: Implement VM export
  activeVMDropdown.value = null
}

const deleteVM = (vm: any) => {
  console.log('Delete VM:', vm)
  // TODO: Implement VM deletion with confirmation
  activeVMDropdown.value = null
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
  // Fetch VMs for all hosts
  for (const host of hostStore.hosts) {
    await vmStore.fetchVMs(host.id)
  }
})
</script>

<style scoped>
.animate-slideDown {
  animation: slideDown 0.3s ease-out forwards;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
    max-height: 0;
  }
  to {
    opacity: 1;
    transform: translateY(0);
    max-height: 400px;
  }
}
</style>
