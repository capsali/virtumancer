<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Discovered Virtual Machines</h1>
        <p class="text-slate-400 mt-2">Import VMs discovered from connected hosts</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="text-right">
          <div class="text-2xl font-bold text-white">{{ filteredVMs.length }}</div>
          <div class="text-sm text-slate-400">Discovered</div>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-amber-400">{{ filteredVMs.length }}</div>
          <div class="text-sm text-slate-400">Available</div>
        </div>
        <FButton
          @click="refreshDiscoveredVMs"
          :disabled="loading"
          variant="secondary"
          size="md"
          class="ml-6 px-3 py-2"
          title="Refresh discovered VMs"
        >
          <svg 
            class="w-4 h-4" 
            :class="{ 'animate-spin': loading }" 
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
          </svg>
        </FButton>
        <div class="relative">
          <FButton
            @click="importSelectedVMs"
            :disabled="selectedVMs.length === 0"
            variant="primary"
            size="lg"
            class="px-4 py-3 bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700"
            title="Import selected VMs"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"></path>
            </svg>
          </FButton>
          <div v-if="selectedVMs.length > 0" class="absolute -top-2 -right-2 bg-amber-500 text-white text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center">
            {{ selectedVMs.length }}
          </div>
        </div>
      </div>
    </div>

    <!-- Search & Filters Card -->
    <FCard class="card-glow">
      <div class="p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <h2 class="text-lg font-semibold text-white">Search & Filters</h2>
            <div v-if="hasActiveFilters" class="px-2 py-1 bg-amber-500/20 text-amber-400 text-xs rounded-full border border-amber-500/30">
              {{ activeFiltersCount }} active
            </div>
          </div>
          <div class="text-sm text-slate-400">
            {{ filteredVMs.length }} of {{ totalVMs }} discovered VMs
          </div>
        </div>
        
        <!-- Search Bar with integrated Filter Button -->
        <div class="relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search discovered VMs... (try 'host:myhost', 'name:web')"
            class="w-full px-4 py-3 pr-20 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus-glow transition-all duration-200"
          />
          <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
            <button
              @click.stop="toggleFilterDropdown"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                showFilterDropdown ? 'text-amber-400 bg-amber-500/20' : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
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
            <div><code class="text-amber-400">host:myhost</code> - Filter by host</div>
            <div><code class="text-amber-400">name:web</code> - Filter by VM name</div>
            <div><code class="text-amber-400">uuid:abc123</code> - Filter by UUID</div>
            <div><code class="text-amber-400">state:active</code> - Filter by state</div>
          </div>
        </div>

        <!-- Expandable Filter Section -->
        <div v-if="showFilterDropdown" class="mt-4 space-y-4 p-4 bg-slate-800/30 rounded-lg border border-slate-600/30 animate-slideDown">
          <!-- Quick Filters -->
          <div>
            <h4 class="text-sm font-medium text-slate-300 mb-3">Quick Filters</h4>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="filter in quickFilters"
                :key="filter.value"
                @click="quickFilter(filter.value)"
                :class="[
                  'p-2 text-xs rounded-lg transition-all hover:scale-105 flex items-center justify-center gap-1',
                  'bg-slate-700 text-slate-300 hover:bg-slate-600'
                ]"
              >
                <span>{{ filter.icon }}</span>
                <span class="hidden sm:inline">{{ filter.label }}</span>
              </button>
            </div>
          </div>

          <!-- Detailed Filters -->
          <div class="grid grid-cols-1 md:grid-cols-1 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Host</label>
              <select
                v-model="hostFilter"
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-amber-500"
              >
                <option value="all">All Hosts</option>
                <option v-for="host in hosts" :key="host.id" :value="host.id">{{ host.name || host.uri }}</option>
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
              class="flex-1 px-4 py-2 text-sm bg-amber-600 hover:bg-amber-700 text-white rounded-lg transition-all"
            >
              Apply Filters
            </button>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Discovered VMs List Card -->
    <FCard class="card-glow">
      <div class="border-b border-slate-700/50">
        <div class="flex items-center justify-between px-6 pt-6 pb-4">
          <div class="flex items-center gap-4">
            <h2 class="text-lg font-semibold text-white">Discovered VMs ({{ filteredVMs.length }})</h2>
            
            <!-- Bulk Selection -->
            <div v-if="filteredVMs.length > 0" class="flex items-center gap-2">
              <input
                type="checkbox"
                :checked="allSelected"
                :indeterminate="someSelected && !allSelected"
                @change="toggleSelectAll"
                class="w-4 h-4 text-amber-600 bg-slate-700 border-slate-600 rounded focus:ring-amber-500 focus:ring-2"
              />
              <span class="text-sm text-slate-400">
                {{ selectedVMs.length > 0 ? `${selectedVMs.length} selected` : 'Select all' }}
              </span>
            </div>
          </div>
          
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
                      itemsPerPage === option.value ? 'text-amber-400 bg-amber-500/10' : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
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

      <!-- Discovered VMs Content -->
      <div class="p-6">
        <!-- Cards View -->
        <div v-if="viewMode === 'cards'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <FCard
            v-for="vm in paginatedVMs"
            :key="vm.uuid"
            class="card-glow border transition-all duration-300 hover:scale-105 border-amber-400/30"
          >
            <div class="p-4 space-y-4">
              <!-- VM Header -->
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <input
                    type="checkbox"
                    :checked="selectedVMs.includes(vm.uuid)"
                    @change="toggleVMSelection(vm.uuid)"
                    class="w-4 h-4 text-amber-600 bg-slate-700 border-slate-600 rounded focus:ring-amber-500 focus:ring-2"
                  />
                  <div class="w-10 h-10 bg-amber-500/20 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                    </svg>
                  </div>
                  <div>
                    <h3 class="font-semibold text-white">{{ vm.name || 'Unnamed VM' }}</h3>
                    <p class="text-xs text-slate-400">{{ vm.uuid.slice(0, 16) }}...</p>
                  </div>
                </div>
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400">
                  Available
                </span>
              </div>

              <!-- VM Details -->
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span class="text-slate-400">Host:</span>
                  <span class="text-white">{{ getHostName(vm.host_id) }}</span>
                </div>
                <div class="flex justify-between text-sm">
                  <span class="text-slate-400">Uptime:</span>
                  <span class="text-white">{{ vm.uptime ? formatUptime(vm.uptime) : 'Unknown' }}</span>
                </div>
                <div v-if="vm.vcpu" class="flex justify-between text-sm">
                  <span class="text-slate-400">vCPUs:</span>
                  <span class="text-white">{{ vm.vcpu }}</span>
                </div>
                <div v-if="vm.memory" class="flex justify-between text-sm">
                  <span class="text-slate-400">Memory:</span>
                  <span class="text-white">{{ Math.round(vm.memory / 1024 / 1024) }} MB</span>
                </div>
              </div>

              <!-- Actions -->
              <div class="flex gap-2 pt-2">
                <FButton
                  @click="importSingleVM(vm)"
                  variant="primary"
                  size="sm"
                  class="flex-1 bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700"
                >
                  Import
                </FButton>
              </div>
            </div>
          </FCard>
        </div>

        <!-- Compact View -->
        <div v-else-if="viewMode === 'compact'" class="space-y-2">
          <div
            v-for="vm in paginatedVMs"
            :key="vm.uuid"
            class="group bg-slate-800/30 rounded-lg p-4 border border-amber-400/20 hover:bg-slate-700/40 hover:border-amber-500/40 transition-all duration-200"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-4 min-w-0 flex-1">
                <!-- Selection Checkbox -->
                <input
                  type="checkbox"
                  :checked="selectedVMs.includes(vm.uuid)"
                  @change="toggleVMSelection(vm.uuid)"
                  class="w-4 h-4 text-amber-600 bg-slate-700 border-slate-600 rounded focus:ring-amber-500 focus:ring-2"
                />
                
                <!-- VM Icon with Status Indicator -->
                <div class="relative">
                  <div class="w-10 h-10 bg-gradient-to-br from-amber-500/20 to-orange-600/20 rounded-lg flex items-center justify-center border border-amber-500/30">
                    <svg class="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                    </svg>
                  </div>
                  <div class="absolute -top-1 -right-1 w-3 h-3 rounded-full border-2 border-slate-800 bg-amber-400"></div>
                </div>
                
                <!-- VM Info -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="font-semibold text-white truncate">{{ vm.name || 'Unnamed VM' }}</span>
                    <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400 border border-amber-500/30">
                      Available
                    </span>
                  </div>
                  <div class="flex items-center gap-4 text-xs text-slate-400 mt-1">
                    <span class="flex items-center gap-1">
                      <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                      </svg>
                      {{ getHostName(vm.host_id) }}
                    </span>
                    <span class="flex items-center gap-1" v-if="vm.vcpu">
                      <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M3 5a2 2 0 012-2h10a2 2 0 012 2v8a2 2 0 01-2 2h-2.22l.123.489.804.804A1 1 0 0113 18H7a1 1 0 01-.707-1.707l.804-.804L7.22 15H5a2 2 0 01-2-2V5zm5.771 7H5V5h10v7H8.771z" clip-rule="evenodd" />
                      </svg>
                      {{ vm.vcpu }} vCPUs
                    </span>
                    <span class="flex items-center gap-1" v-if="vm.memory">
                      <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
                        <path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
                        <path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
                      </svg>
                      {{ Math.round(vm.memory / 1024 / 1024) }} MB
                    </span>
                    <span class="flex items-center gap-1">
                      UUID: {{ vm.uuid.slice(0, 16) }}...
                    </span>
                  </div>
                </div>
              </div>
              
              <!-- Action Button - Hover to Show -->
              <div class="flex items-center gap-2 ml-4 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                <FButton
                  @click="importSingleVM(vm)"
                  variant="primary"
                  size="sm"
                  class="bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 text-white"
                >
                  Import
                </FButton>
              </div>
            </div>
          </div>
        </div>

        <!-- List View -->
        <div v-else class="space-y-3">
          <!-- Sorting Header -->
          <div class="flex items-center gap-4 px-4 py-2 bg-slate-800/20 rounded-lg border border-slate-600/20 backdrop-blur-sm">
            <span class="text-sm font-medium text-slate-300">Name</span>
            <span class="text-sm font-medium text-slate-300">Host</span>
            <span class="text-sm font-medium text-slate-300">Status</span>
            <span class="text-sm font-medium text-slate-300">Resources</span>
          </div>

          <!-- VM List Items -->
          <div
            v-for="vm in paginatedVMs"
            :key="vm.uuid"
            class="group relative bg-slate-800/30 hover:bg-slate-700/40 rounded-xl p-5 border border-amber-400/20 hover:border-amber-500/40 transition-all duration-300 backdrop-blur-sm"
          >
            <!-- Status Orb with Pulse Animation -->
            <div class="absolute top-4 left-4">
              <div class="w-3 h-3 rounded-full bg-amber-400 animate-pulse shadow-lg shadow-amber-400/50"></div>
            </div>

            <!-- Main Content -->
            <div class="flex items-center justify-between pl-8">
              <!-- VM Information -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-4">
                  <!-- Selection Checkbox -->
                  <input
                    type="checkbox"
                    :checked="selectedVMs.includes(vm.uuid)"
                    @change="toggleVMSelection(vm.uuid)"
                    class="w-4 h-4 text-amber-600 bg-slate-700 border-slate-600 rounded focus:ring-amber-500 focus:ring-2"
                  />

                  <!-- VM Name & UUID -->
                  <div class="min-w-0 flex-1">
                    <h3 class="text-lg font-semibold text-white truncate group-hover:text-amber-300 transition-colors">
                      {{ vm.name || 'Unnamed VM' }}
                    </h3>
                    <p class="text-sm text-slate-400 truncate">
                      UUID: {{ vm.uuid.slice(0, 32) }}...
                    </p>
                  </div>

                  <!-- Host Info -->
                  <div class="hidden md:flex items-center gap-2 text-slate-300">
                    <svg class="w-4 h-4 text-slate-500" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                    </svg>
                    <span class="text-sm">{{ getHostName(vm.host_id) }}</span>
                  </div>

                  <!-- Status Badge -->
                  <div class="flex items-center">
                    <span class="px-3 py-1 rounded-full text-xs font-medium backdrop-blur-sm border bg-amber-500/20 text-amber-300 border-amber-500/30">
                      Available
                    </span>
                  </div>

                  <!-- Resource Information -->
                  <div class="hidden lg:flex items-center gap-4 text-sm text-slate-400">
                    <div class="flex items-center gap-2" v-if="vm.vcpu">
                      <div class="w-2 h-2 rounded-full bg-blue-400"></div>
                      <span>{{ vm.vcpu }} vCPUs</span>
                    </div>
                    <div class="flex items-center gap-2" v-if="vm.memory">
                      <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                      <span>{{ Math.round(vm.memory / 1024 / 1024) }} MB</span>
                    </div>
                    <div class="flex items-center gap-2" v-if="vm.uptime">
                      <div class="w-2 h-2 rounded-full bg-green-400"></div>
                      <span>{{ formatUptime(vm.uptime) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Action Button (Hidden by default, shown on hover) -->
              <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300 transform translate-x-4 group-hover:translate-x-0">
                <FButton
                  @click="importSingleVM(vm)"
                  variant="primary"
                  size="sm"
                  class="bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 text-white transition-all duration-200"
                >
                  Import VM
                </FButton>
              </div>
            </div>

            <!-- Mobile Resource Info (shown on smaller screens) -->
            <div class="lg:hidden mt-3 pt-3 border-t border-slate-600/20 flex items-center gap-4 text-sm text-slate-400">
              <div class="flex items-center gap-2" v-if="vm.vcpu">
                <div class="w-2 h-2 rounded-full bg-blue-400"></div>
                <span>{{ vm.vcpu }} vCPUs</span>
              </div>
              <div class="flex items-center gap-2" v-if="vm.memory">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span>{{ Math.round(vm.memory / 1024 / 1024) }} MB</span>
              </div>
              <div class="md:hidden flex items-center gap-2">
                <svg class="w-3 h-3 text-slate-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <span>{{ getHostName(vm.host_id) }}</span>
              </div>
              <div class="flex items-center gap-2" v-if="vm.uptime">
                <div class="w-2 h-2 rounded-full bg-green-400"></div>
                <span>{{ formatUptime(vm.uptime) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="filteredVMs.length === 0" class="text-center py-12">
          <div class="w-16 h-16 bg-slate-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-white mb-2">No Discovered VMs</h3>
          <p class="text-slate-400 mb-4">
            {{ totalVMs === 0 ? 'No VMs have been discovered yet.' : 'No VMs match your current filters.' }}
          </p>
          <div class="flex gap-3 justify-center">
            <button
              v-if="totalVMs > 0"
              @click="clearAllFilters"
              class="px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded-lg transition-all"
            >
              Clear Filters
            </button>
            <button
              @click="refreshDiscoveredVMs"
              class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-all"
            >
              Refresh Discovery
            </button>
          </div>
        </div>
      </div>
    </FCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useUserPreferences } from '@/composables/useUserPreferences'
import type { DiscoveredVMWithHost } from '@/types'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import FButton from '@/components/ui/FButton.vue'
import { vClickAway } from '@/directives/clickAway'

// Store instances
const route = useRoute()
const hostStore = useHostStore()
const { vmListPreferences } = useUserPreferences()

// Loading states
const loading = ref(false)

// Search and filters
const searchQuery = ref('')
const hostFilter = ref('all')
const showFilterDropdown = ref(false)
const showFilterHelp = ref(false)

// VM selection
const selectedVMs = ref<string[]>([])

// View settings
const viewMode = ref(vmListPreferences.viewMode)
const itemsPerPage = ref<number | 'all'>(12)
const showItemsDropdown = ref(false)
const currentPage = ref(1)

// Quick filters data
const quickFilters = [
  { value: 'all', label: 'All VMs', icon: '📋', color: 'blue' }
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
const discoveredVMs = computed(() => hostStore.allDiscoveredVMs)
const hosts = computed(() => hostStore.hosts)
const totalVMs = computed(() => discoveredVMs.value.length)


const filteredVMs = computed(() => {
  let filtered = discoveredVMs.value

  // Filter out VMs from disconnected hosts first
  filtered = filtered.filter((vm: DiscoveredVMWithHost) => {
    const host = hosts.value.find(h => h.id === vm.host_id)
    return host && host.state === 'CONNECTED'
  })

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter((vm: DiscoveredVMWithHost) => 
      vm.name?.toLowerCase().includes(query) ||
      vm.uuid.toLowerCase().includes(query) ||
      hosts.value.find(h => h.id === vm.host_id)?.name?.toLowerCase().includes(query)
    )
  }

  // No import status filtering needed for DiscoveredVMWithHost type

  // Host filter
  if (hostFilter.value !== 'all') {
    filtered = filtered.filter((vm: DiscoveredVMWithHost) => vm.host_id === hostFilter.value)
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

const hasActiveFilters = computed(() => {
  return searchQuery.value !== '' ||
         hostFilter.value !== 'all'
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (searchQuery.value) count++
  if (hostFilter.value !== 'all') count++
  return count
})

const allSelected = computed(() => {
  const availableVMs = filteredVMs.value
  return availableVMs.length > 0 && availableVMs.every((vm: DiscoveredVMWithHost) => selectedVMs.value.includes(vm.uuid))
})

const someSelected = computed(() => {
  return selectedVMs.value.length > 0
})

// Methods
const toggleFilterDropdown = () => {
  showFilterDropdown.value = !showFilterDropdown.value
}

const toggleItemsDropdown = () => {
  showItemsDropdown.value = !showItemsDropdown.value
}

const quickFilter = (filterValue: string) => {
  // Quick filters are simplified for discovered VMs
  console.log('Quick filter:', filterValue)
}

const clearAllFilters = () => {
  searchQuery.value = ''
  hostFilter.value = 'all'
  showFilterDropdown.value = false
}

const cycleViewMode = () => {
  const modes: Array<'cards' | 'compact' | 'list'> = ['cards', 'compact', 'list']
  const currentIndex = modes.indexOf(viewMode.value)
  const nextIndex = (currentIndex + 1) % modes.length
  viewMode.value = modes[nextIndex]!
  vmListPreferences.viewMode = modes[nextIndex]!
}

const setItemsPerPage = (value: number | 'all') => {
  itemsPerPage.value = value
  currentPage.value = 1
  showItemsDropdown.value = false
}

const toggleVMSelection = (vmUuid: string) => {
  const index = selectedVMs.value.indexOf(vmUuid)
  if (index > -1) {
    selectedVMs.value.splice(index, 1)
  } else {
    selectedVMs.value.push(vmUuid)
  }
}

const toggleSelectAll = () => {
  const availableVMs = filteredVMs.value
  if (allSelected.value) {
    // Deselect all
    selectedVMs.value = selectedVMs.value.filter(uuid => 
      !availableVMs.some((vm: DiscoveredVMWithHost) => vm.uuid === uuid)
    )
  } else {
    // Select all available VMs
    const newSelections = availableVMs
      .filter((vm: DiscoveredVMWithHost) => !selectedVMs.value.includes(vm.uuid))
      .map((vm: DiscoveredVMWithHost) => vm.uuid)
    selectedVMs.value.push(...newSelections)
  }
}

const refreshDiscoveredVMs = async () => {
  loading.value = true
  try {
    await hostStore.refreshAllDiscoveredVMs()
  } finally {
    loading.value = false
  }
}

const importSingleVM = async (vm: DiscoveredVMWithHost) => {
  try {
    await hostStore.importSelectedVMs(vm.host_id, [vm.uuid])
    // Remove from selection if it was selected
    const index = selectedVMs.value.indexOf(vm.uuid)
    if (index > -1) {
      selectedVMs.value.splice(index, 1)
    }
  } catch (error) {
    console.error('Failed to import VM:', error)
  }
}

const importSelectedVMs = async () => {
  if (selectedVMs.value.length === 0) return

  try {
    // Group by host_id for batch import
    const vmsByHost = selectedVMs.value.reduce((acc, vmUuid) => {
      const vm = discoveredVMs.value.find((v: DiscoveredVMWithHost) => v.uuid === vmUuid)
      if (vm) {
        if (!acc[vm.host_id]) {
          acc[vm.host_id] = []
        }
        acc[vm.host_id]!.push(vmUuid)
      }
      return acc
    }, {} as Record<string, string[]>)

    // Import VMs by host
    for (const [hostId, vmUuids] of Object.entries(vmsByHost)) {
      await hostStore.importSelectedVMs(hostId, vmUuids)
    }
    
    selectedVMs.value = []
  } catch (error) {
    console.error('Failed to import selected VMs:', error)
  }
}

const getHostName = (hostId: string): string => {
  const host = hosts.value.find(h => h.id === hostId)
  return host?.name || host?.uri || 'Unknown Host'
}

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

// Lifecycle
onMounted(() => {
  hostStore.fetchGlobalDiscoveredVMs()
  hostStore.fetchHosts()
  
  // Check for hostId query parameter and set filter
  if (route.query.hostId && typeof route.query.hostId === 'string') {
    hostFilter.value = route.query.hostId
  }
})

// Watch for route changes to handle navigation from host detail
watch(() => route.query.hostId, (newHostId) => {
  if (newHostId && typeof newHostId === 'string') {
    hostFilter.value = newHostId
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
