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
          <div class="text-2xl font-bold text-green-400">{{ activeVMs }}</div>
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
        <!-- Search and Filters Row -->
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1">
            <div class="relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search VMs... (try 'host:myhost', 'cpu:>2', 'memory:>4GB')"
                class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus-glow transition-all duration-200"
              />
              <div class="absolute right-3 top-1/2 -translate-y-1/2">
                <button
                  @click="showFilterHelp = !showFilterHelp"
                  class="text-slate-400 hover:text-white transition-colors"
                  title="Show filter syntax help"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
            <!-- Filter Help -->
            <div v-if="showFilterHelp" class="mt-2 p-3 bg-slate-700 rounded-lg text-sm text-slate-300">
              <div class="font-medium mb-2">Filter Syntax:</div>
              <div class="space-y-1">
                <div><code class="text-blue-400">host:myhost</code> - Filter by host name</div>
                <div><code class="text-blue-400">cpu:>2</code> - VMs with more than 2 CPUs</div>
                <div><code class="text-blue-400">memory:>4GB</code> - VMs with more than 4GB RAM</div>
                <div><code class="text-blue-400">state:active</code> - Filter by VM state</div>
                <div><code class="text-blue-400">name:web</code> - VMs with 'web' in name</div>
              </div>
            </div>
          </div>
          <div class="flex gap-2">
            <select
              v-model="statusFilter"
              class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
            >
              <option value="all">All States</option>
              <option value="ACTIVE">Active</option>
              <option value="STOPPED">Stopped</option>
              <option value="PAUSED">Paused</option>
              <option value="ERROR">Error</option>
            </select>
            <select
              v-model="hostFilter"
              class="px-4 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus-glow transition-all duration-200"
            >
              <option value="all">All Hosts</option>
              <option v-for="host in hosts" :key="host.id" :value="host.id">{{ host.name || host.uri }}</option>
            </select>
            <!-- Clear Filters Button -->
            <button
              v-if="searchQuery || statusFilter !== 'all' || hostFilter !== 'all'"
              @click="clearAllFilters"
              class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-slate-300 hover:text-white rounded-lg transition-colors text-sm"
              title="Clear all filters"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Tab Navigation -->
        <div class="flex items-center gap-1 border-b border-slate-700/50">
          <button
            @click="activeTab = 'managed'"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-t-lg transition-all duration-200',
              activeTab === 'managed'
                ? 'bg-primary-500/20 text-primary-400 border-b-2 border-primary-500'
                : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
            ]"
          >
            Managed VMs ({{ totalVMs }})
          </button>
          <button
            @click="activeTab = 'discovered'"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-t-lg transition-all duration-200',
              activeTab === 'discovered'
                ? 'bg-primary-500/20 text-primary-400 border-b-2 border-primary-500'
                : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
            ]"
          >
            Discovered VMs ({{ discoveredVMs.length }})
          </button>
        </div>

        <!-- View Toggle Row -->
        <div v-if="activeTab === 'managed'" class="flex justify-between items-center">
          <div class="text-sm text-slate-400">
            Showing {{ filteredVMs.length }} of {{ totalVMs }} virtual machines
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm text-slate-400 mr-2">View:</span>
            <div class="flex bg-slate-800/50 rounded-lg p-1 border border-slate-600/50 overflow-x-auto">
              <button
                @click="viewMode = 'grid'"
                :class="[
                  'px-2 xl:px-3 py-1.5 rounded-md text-xs xl:text-sm transition-all duration-200 flex items-center gap-1 xl:gap-2 whitespace-nowrap',
                  viewMode === 'grid'
                    ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                ]"
                title="Card View"
              >
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
                <span class="hidden sm:inline">Cards</span>
              </button>
              <button
                @click="viewMode = 'list'"
                :class="[
                  'px-2 xl:px-3 py-1.5 rounded-md text-xs xl:text-sm transition-all duration-200 flex items-center gap-1 xl:gap-2 whitespace-nowrap',
                  viewMode === 'list'
                    ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                ]"
                title="List View"
              >
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 6.707 6.293a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="hidden sm:inline">List</span>
              </button>
              <button
                @click="viewMode = 'compact'"
                :class="[
                  'px-2 xl:px-3 py-1.5 rounded-md text-xs xl:text-sm transition-all duration-200 flex items-center gap-1 xl:gap-2 whitespace-nowrap',
                  viewMode === 'compact'
                    ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                ]"
                title="Compact View"
              >
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M2 4a1 1 0 011-1h14a1 1 0 110 2H3a1 1 0 01-1-1zM2 8a1 1 0 011-1h14a1 1 0 110 2H3a1 1 0 01-1-1zM2 12a1 1 0 011-1h14a1 1 0 110 2H3a1 1 0 01-1-1zM2 16a1 1 0 011-1h14a1 1 0 110 2H3a1 1 0 01-1-1z" />
                </svg>
                <span class="hidden sm:inline">Compact</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Discovered VMs Tab Content -->
    <div v-if="activeTab === 'discovered'">
      <FCard class="p-6 card-glow">
        <div class="flex items-center justify-between mb-6">
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
                  <th class="text-left py-3 px-4 font-medium text-slate-300">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="vm in filteredDiscoveredVMs"
                  :key="vm.uuid"
                  class="border-b border-slate-700 hover:bg-slate-800/50 transition-colors"
                >
                  <td class="py-3 px-4">
                    <input
                      type="checkbox"
                      :value="vm.uuid"
                      v-model="selectedDiscoveredVMs"
                      class="rounded border-slate-500 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                    />
                  </td>
                  <td class="py-3 px-4">
                    <div class="font-medium text-white">{{ vm.name }}</div>
                  </td>
                  <td class="py-3 px-4">
                    <div class="text-slate-300">
                      <span class="inline-flex items-center gap-2">
                        <div class="w-2 h-2 rounded-full bg-green-400"></div>
                        {{ vm.host_name || vm.host_id }}
                      </span>
                    </div>
                  </td>
                  <td class="py-3 px-4 text-slate-400 font-mono text-sm">{{ vm.uuid.slice(0, 8) }}...</td>
                  <td class="py-3 px-4">
                    <FButton
                      @click="importSingleDiscoveredVM(vm)"
                      variant="primary"
                      size="sm"
                      :disabled="!!importingVMs[vm.uuid]"
                      class="text-xs"
                    >
                      <svg v-if="!importingVMs[vm.uuid]" class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 4.414V13a1 1 0 11-2 0V4.414L7.707 5.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                      </svg>
                      <svg v-else class="w-3 h-3 mr-1 animate-spin" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      {{ importingVMs[vm.uuid] ? 'Importing...' : 'Import' }}
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
      </FCard>
    </div>

    <!-- VM Grid View -->
    <div v-if="activeTab === 'managed' && filteredVMs.length > 0 && viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
              class="flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
              </svg>
              Start
            </FButton>
            <FButton
              variant="ghost"
              size="sm"
              @click.stop="handleVMAction(vm, 'stop')"
              :disabled="vm.state === 'STOPPED' || !!vm.taskState"
              class="flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
              </svg>
              Stop
            </FButton>
            <FButton
              variant="outline"
              size="sm"
              @click.stop="openVMConsole(vm)"
              :disabled="vm.state !== 'ACTIVE'"
              class="flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
              </svg>
              Console
            </FButton>
          </div>
        </div>
      </FCard>
    </div>

    <!-- VM List View -->
    <div v-if="activeTab === 'managed' && filteredVMs.length > 0 && viewMode === 'list'" class="overflow-x-auto">
      <FCard class="card-glow overflow-hidden">
        <table class="w-full">
          <thead class="border-b border-slate-700/50">
            <tr class="text-left">
              <th 
                class="px-6 py-3 text-sm font-medium text-slate-300 cursor-pointer hover:text-white transition-colors select-none"
                @click="handleSort('name')"
              >
                <div class="flex items-center gap-2">
                  Name
                  <span class="text-xs opacity-70">{{ getSortIcon('name') }}</span>
                </div>
              </th>
              <th 
                class="px-6 py-3 text-sm font-medium text-slate-300 cursor-pointer hover:text-white transition-colors select-none"
                @click="handleSort('host')"
              >
                <div class="flex items-center gap-2">
                  Host
                  <span class="text-xs opacity-70">{{ getSortIcon('host') }}</span>
                </div>
              </th>
              <th 
                class="px-6 py-3 text-sm font-medium text-slate-300 cursor-pointer hover:text-white transition-colors select-none"
                @click="handleSort('status')"
              >
                <div class="flex items-center gap-2">
                  Status
                  <span class="text-xs opacity-70">{{ getSortIcon('status') }}</span>
                </div>
              </th>
              <th 
                class="px-6 py-3 text-sm font-medium text-slate-300 text-center cursor-pointer hover:text-white transition-colors select-none"
                @click="handleSort('vcpus')"
              >
                <div class="flex items-center justify-center gap-2">
                  vCPUs
                  <span class="text-xs opacity-70">{{ getSortIcon('vcpus') }}</span>
                </div>
              </th>
              <th 
                class="px-6 py-3 text-sm font-medium text-slate-300 cursor-pointer hover:text-white transition-colors select-none"
                @click="handleSort('memory')"
              >
                <div class="flex items-center gap-2">
                  Memory
                  <span class="text-xs opacity-70">{{ getSortIcon('memory') }}</span>
                </div>
              </th>
              <th class="px-6 py-3 text-sm font-medium text-slate-300">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700/30">
            <tr
              v-for="vm in filteredVMs"
              :key="vm.uuid"
              class="group hover:bg-slate-800/30 transition-colors duration-200 cursor-pointer"
              @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
            >
              <!-- Name Column -->
              <td class="px-6 py-3">
                <div class="flex items-center gap-3">
                  <div :class="[
                    'w-3 h-3 rounded-full flex-shrink-0',
                    getVMStatusColor(vm.state)
                  ]"></div>
                  <div class="min-w-0">
                    <div class="text-white font-medium truncate">{{ vm.name }}</div>
                    <div class="text-xs text-slate-400 truncate">{{ vm.osType || 'Unknown OS' }}</div>
                  </div>
                </div>
              </td>
              
              <!-- Host Column -->
              <td class="px-6 py-3 text-slate-300">{{ vm.hostName || 'Unknown Host' }}</td>
              
              <!-- Status Column -->
              <td class="px-6 py-3">
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium whitespace-nowrap',
                  vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
                  vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-400' :
                  vm.state === 'ERROR' ? 'bg-red-600/20 text-red-300' :
                  'bg-yellow-500/20 text-yellow-400'
                ]">
                  {{ vm.state }}
                </span>
              </td>
              
              <!-- vCPUs Column -->
              <td class="px-6 py-3 text-slate-300 text-center">{{ vm.vcpuCount || 'N/A' }}</td>
              
              <!-- Memory Column -->
              <td class="px-6 py-3 text-slate-300">{{ vm.memoryMB ? `${Math.round(vm.memoryMB / 1024)}GB` : 'N/A' }}</td>
              
              <!-- Actions Column -->
              <td class="px-6 py-3">
                <div class="flex items-center gap-2">
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="handleVMAction(vm, 'start')"
                  :disabled="vm.state === 'ACTIVE' || !!vm.taskState"
                  class="p-2"
                  title="Start VM"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="handleVMAction(vm, 'stop')"
                  :disabled="vm.state === 'STOPPED' || !!vm.taskState"
                  class="p-2"
                  title="Stop VM"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <FButton
                  variant="outline"
                  size="sm"
                  @click.stop="openVMConsole(vm)"
                  :disabled="vm.state !== 'ACTIVE'"
                  class="p-2"
                  title="Open Console"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <!-- View Details button - moved from dropdown -->
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="viewVMDetails(vm)"
                  class="p-2 hidden lg:flex"
                  title="View Details"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M2 10a8 8 0 018-8v8h8a8 8 0 11-16 0z" />
                    <path d="M12 2.252A8.014 8.014 0 0117.748 8H12V2.252z" />
                  </svg>
                </FButton>
                <!-- Dropdown Menu -->
                <div class="relative">
                  <FButton
                    variant="ghost"
                    size="sm"
                    @click.stop="toggleDropdown(vm.uuid)"
                    class="p-2"
                    title="More Actions"
                  >
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
                    </svg>
                  </FButton>
                  <div 
                    v-if="activeDropdown === vm.uuid"
                    class="absolute right-0 top-full mt-1 w-48 bg-slate-800 border border-slate-600/50 rounded-lg shadow-lg z-50 card-glow"
                  >
                    <div class="py-1">
                      <!-- Show View Details only on smaller screens -->
                      <button
                        @click.stop="viewVMDetails(vm)"
                        class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3 lg:hidden"
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
              </td>
            </tr>
          </tbody>
        </table>
      </FCard>
    </div>

    <!-- VM Compact View -->
    <div v-if="activeTab === 'managed' && filteredVMs.length > 0 && viewMode === 'compact'" class="space-y-1">
      <div
        v-for="vm in filteredVMs"
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
                <div class="flex items-center gap-2">
                  <span class="text-white font-medium truncate">{{ vm.name }}</span>
                  <span :class="[
                    'px-1.5 py-0.5 rounded text-xs font-medium flex-shrink-0',
                    vm.state === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
                    vm.state === 'STOPPED' ? 'bg-red-500/20 text-red-400' :
                    vm.state === 'ERROR' ? 'bg-red-600/20 text-red-300' :
                    'bg-yellow-500/20 text-yellow-400'
                  ]">
                    {{ vm.state }}
                  </span>
                </div>
                <div class="text-xs text-slate-400 truncate">
                  {{ vm.hostName }} • {{ vm.vcpuCount || 'N/A' }} vCPUs • {{ vm.memoryMB ? `${Math.round(vm.memoryMB / 1024)}GB` : 'N/A' }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-1 ml-2">
              <FButton
                variant="ghost"
                size="sm"
                @click.stop="handleVMAction(vm, 'start')"
                :disabled="vm.state === 'ACTIVE' || !!vm.taskState"
                class="text-xs p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                title="Start VM"
              >
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                </svg>
              </FButton>
              <FButton
                variant="ghost"
                size="sm"
                @click.stop="handleVMAction(vm, 'stop')"
                :disabled="vm.state === 'STOPPED' || !!vm.taskState"
                class="text-xs p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                title="Stop VM"
              >
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
                </svg>
              </FButton>
              <FButton
                variant="outline"
                size="sm"
                @click.stop="openVMConsole(vm)"
                :disabled="vm.state !== 'ACTIVE'"
                class="text-xs p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                title="Console"
              >
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                </svg>
              </FButton>
            </div>
          </div>
        </FCard>
      </div>
    </div>



    <!-- Empty State for Managed VMs -->
    <div v-else-if="activeTab === 'managed' && filteredVMs.length === 0" class="text-center py-12">
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import { useUserPreferences } from '@/composables/useUserPreferences'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import { getConsoleRoute } from '@/utils/console'
import DiscoveredVMBulkManager from '@/components/vm/DiscoveredVMBulkManager.vue'

const router = useRouter()
const hostStore = useHostStore()
const vmStore = useVMStore()
const { vmListPreferences } = useUserPreferences()

// Reactive data
const searchQuery = ref('')
const statusFilter = ref('all')
const hostFilter = ref('all')
const activeDropdown = ref<string | null>(null)
const activeTab = ref<'managed' | 'discovered'>('managed')

// Filter UI state
const showFilterHelp = ref(false)

// Discovered VMs state
const selectedDiscoveredVMs = ref<string[]>([])
const importingVMs = ref<Record<string, boolean>>({})
const bulkImporting = ref(false)

// Use preferences for persistent state
const viewMode = computed({
  get: () => vmListPreferences.viewMode,
  set: (value: 'grid' | 'list' | 'compact') => { vmListPreferences.viewMode = value }
})

const sortBy = computed({
  get: () => vmListPreferences.sortBy,
  set: (value: string) => { vmListPreferences.sortBy = value }
})

const sortDirection = computed({
  get: () => vmListPreferences.sortDirection,
  set: (value: 'asc' | 'desc') => { vmListPreferences.sortDirection = value }
})

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
        hostName: host.name || host.uri
      })
    })
  })
  return vms
})

const filteredVMs = computed(() => {
  let filtered = allVMs.value.filter(vm => {
    // If no search query, apply basic filters
    if (!searchQuery.value) {
      const matchesStatus = statusFilter.value === 'all' || vm.state === statusFilter.value
      const matchesHost = hostFilter.value === 'all' || vm.hostId === hostFilter.value
      return matchesStatus && matchesHost
    }
    
    // Parse advanced search filters
    const filters = parseFilterQuery(searchQuery.value)
    
    // Host filter
    if (filters.host && !vm.hostName?.toLowerCase().includes(filters.host.toLowerCase()) && 
        !vm.hostId?.toLowerCase().includes(filters.host.toLowerCase())) {
      return false
    }
    
    // Name filter
    if (filters.name && !vm.name?.toLowerCase().includes(filters.name.toLowerCase())) {
      return false
    }
    
    // State filter
    if (filters.state && vm.state?.toLowerCase() !== filters.state.toLowerCase()) {
      return false
    }
    
    // CPU filter (supports >2, <4, =2, etc.)
    if (filters.cpu) {
      const cpuValue = vm.vcpuCount || 0
      if (!evaluateNumericFilter(cpuValue, filters.cpu)) {
        return false
      }
    }
    
    // Memory filter (supports >4GB, <2GB, etc.)
    if (filters.memory) {
      const memoryValue = (vm.memoryMB || 0) / 1024 // Convert MB to GB
      const filterValue = parseMemoryFilter(filters.memory)
      if (filterValue !== null && !evaluateNumericFilter(memoryValue, filterValue)) {
        return false
      }
    }
    
    // General text search
    if (filters.text) {
      const searchText = filters.text
      const matchesText = vm.name?.toLowerCase().includes(searchText) ||
                         vm.hostName?.toLowerCase().includes(searchText) ||
                         (vm.osType && vm.osType.toLowerCase().includes(searchText))
      if (!matchesText) {
        return false
      }
    }
    
    // Apply basic filters if no specific filters found
    const matchesStatus = statusFilter.value === 'all' || vm.state === statusFilter.value
    const matchesHost = hostFilter.value === 'all' || vm.hostId === hostFilter.value
    
    return matchesStatus && matchesHost
  })

  // Apply sorting
  filtered.sort((a, b) => {
    let aValue, bValue

    switch (sortBy.value) {
      case 'name':
        aValue = a.name?.toLowerCase() || ''
        bValue = b.name?.toLowerCase() || ''
        break
      case 'host':
        aValue = a.hostName?.toLowerCase() || ''
        bValue = b.hostName?.toLowerCase() || ''
        break
      case 'status':
        aValue = a.state || ''
        bValue = b.state || ''
        break
      case 'vcpus':
        aValue = a.vcpuCount || 0
        bValue = b.vcpuCount || 0
        return sortDirection.value === 'asc' ? aValue - bValue : bValue - aValue
      case 'memory':
        aValue = a.memoryMB || 0
        bValue = b.memoryMB || 0
        return sortDirection.value === 'asc' ? aValue - bValue : bValue - aValue
      default:
        aValue = a.name?.toLowerCase() || ''
        bValue = b.name?.toLowerCase() || ''
    }

    if (aValue < bValue) return sortDirection.value === 'asc' ? -1 : 1
    if (aValue > bValue) return sortDirection.value === 'asc' ? 1 : -1
    return 0
  })

  return filtered
})

const totalVMs = computed(() => allVMs.value.length)
const activeVMs = computed(() => allVMs.value.filter(vm => vm.state === 'ACTIVE').length)

// Filter status computed properties
const hasActiveFilters = computed(() => {
  return searchQuery.value !== '' || statusFilter.value !== 'all' || hostFilter.value !== 'all'
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (searchQuery.value) count++
  if (statusFilter.value !== 'all') count++
  if (hostFilter.value !== 'all') count++
  return count
})

const discoveredVMs = computed(() => {
  return hostStore.allDiscoveredVMs
})

// Advanced filtering for discovered VMs and managed VMs
const parseFilterQuery = (query: string) => {
  const filters: any = { text: '' }
  const filterPattern = /(\w+):([^\s]+)/g
  let match
  
  // Extract structured filters
  let remainingQuery = query
  while ((match = filterPattern.exec(query)) !== null) {
    const [fullMatch, key, value] = match
    if (key && value) {
      filters[key.toLowerCase()] = value
      remainingQuery = remainingQuery.replace(fullMatch, '').trim()
    }
  }
  
  // Remaining text is general search
  filters.text = remainingQuery.toLowerCase()
  
  return filters
}

const filteredDiscoveredVMs = computed(() => {
  if (!searchQuery.value) {
    return discoveredVMs.value
  }
  
  const filters = parseFilterQuery(searchQuery.value)
  
  return discoveredVMs.value.filter((vm: any) => {
    // Host filter
    if (filters.host && !vm.host_name?.toLowerCase().includes(filters.host.toLowerCase()) && 
        !vm.host_id?.toLowerCase().includes(filters.host.toLowerCase())) {
      return false
    }
    
    // Name filter
    if (filters.name && !vm.name?.toLowerCase().includes(filters.name.toLowerCase())) {
      return false
    }
    
    // CPU filter (supports >2, <4, =2, etc.)
    if (filters.cpu) {
      const cpuValue = vm.vcpu || 0
      if (!evaluateNumericFilter(cpuValue, filters.cpu)) {
        return false
      }
    }
    
    // Memory filter (supports >4GB, <2GB, etc.)
    if (filters.memory) {
      const memoryValue = convertMemoryToGB(vm.memory || vm.max_mem || 0)
      const filterValue = parseMemoryFilter(filters.memory)
      if (filterValue !== null && !evaluateNumericFilter(memoryValue, filterValue)) {
        return false
      }
    }
    
    // General text search
    if (filters.text) {
      const searchText = filters.text
      return vm.name?.toLowerCase().includes(searchText) ||
             vm.host_name?.toLowerCase().includes(searchText) ||
             vm.uuid?.toLowerCase().includes(searchText)
    }
    
    return true
  })
})

// Helper functions for advanced filtering
const evaluateNumericFilter = (value: number, filter: string): boolean => {
  if (filter.startsWith('>')) {
    return value > parseFloat(filter.slice(1))
  } else if (filter.startsWith('<')) {
    return value < parseFloat(filter.slice(1))
  } else if (filter.startsWith('=')) {
    return value === parseFloat(filter.slice(1))
  } else {
    return value === parseFloat(filter)
  }
}

const parseMemoryFilter = (filter: string): string | null => {
  const match = filter.match(/^([><=]?)(\d+(?:\.\d+)?)(gb|mb|kb)?$/i)
  if (!match) return null
  
  const [, operator, valueStr, unit] = match
  if (!valueStr) return null
  
  let value = parseFloat(valueStr)
  
  // Convert to GB
  switch (unit?.toLowerCase()) {
    case 'mb':
      value = value / 1024
      break
    case 'kb':
      value = value / (1024 * 1024)
      break
    case 'gb':
    default:
      // Already in GB or no unit specified (assume GB)
      break
  }
  
  return `${operator}${value}`
}

const convertMemoryToGB = (bytes: number): number => {
  return bytes / (1024 * 1024 * 1024)
}

// Discovered VMs selection and import methods
const toggleSelectAllDiscovered = () => {
  if (selectedDiscoveredVMs.value.length === filteredDiscoveredVMs.value.length) {
    selectedDiscoveredVMs.value = []
  } else {
    selectedDiscoveredVMs.value = filteredDiscoveredVMs.value.map(vm => vm.uuid)
  }
}

const clearSelectedDiscovered = () => {
  selectedDiscoveredVMs.value = []
}

const importSingleDiscoveredVM = async (vm: any) => {
  importingVMs.value[vm.uuid] = true
  try {
    // Use the VM store's import method
    await vmStore.importVM(vm.host_id, vm.name)
    // Remove from selected if it was selected
    selectedDiscoveredVMs.value = selectedDiscoveredVMs.value.filter(id => id !== vm.uuid)
    // Refresh discovered VMs to reflect the import
    await hostStore.fetchGlobalDiscoveredVMs()
  } catch (error) {
    console.error('Failed to import VM:', error)
  } finally {
    importingVMs.value[vm.uuid] = false
  }
}

const importSelectedDiscoveredVMs = async () => {
  if (selectedDiscoveredVMs.value.length === 0) return
  
  bulkImporting.value = true
  try {
    // Group by host for efficient import
    const vmsByHost = new Map<string, any[]>()
    
    filteredDiscoveredVMs.value.forEach(vm => {
      if (selectedDiscoveredVMs.value.includes(vm.uuid)) {
        if (!vmsByHost.has(vm.host_id)) {
          vmsByHost.set(vm.host_id, [])
        }
        vmsByHost.get(vm.host_id)!.push(vm)
      }
    })
    
    // Import VMs for each host
    const importPromises: Promise<void>[] = []
    vmsByHost.forEach((vms, hostId) => {
      vms.forEach(vm => {
        importPromises.push(vmStore.importVM(hostId, vm.name))
      })
    })
    
    await Promise.allSettled(importPromises)
    
    // Clear selection and refresh
    selectedDiscoveredVMs.value = []
    await hostStore.fetchGlobalDiscoveredVMs()
  } catch (error) {
    console.error('Failed to import selected VMs:', error)
  } finally {
    bulkImporting.value = false
  }
}

const handleImportAllVMs = async () => {
  try {
    await hostStore.refreshAllDiscoveredVMs()
    // The refresh will handle the import and update the cache
    await hostStore.fetchHosts()
    await fetchVMsForAllHosts()
  } catch (error) {
    console.error('Failed to import VMs:', error)
  }
}

const handleRefreshDiscoveredVMs = async () => {
  try {
    await hostStore.refreshAllDiscoveredVMs()
  } catch (error) {
    console.error('Failed to refresh discovered VMs:', error)
  }
}

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

// Filter methods
const clearAllFilters = () => {
  searchQuery.value = ''
  statusFilter.value = 'all'
  hostFilter.value = 'all'
  showFilterHelp.value = false
}

// Sorting methods
const handleSort = (column: string) => {
  if (sortBy.value === column) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortDirection.value = 'asc'
  }
}

const getSortIcon = (column: string) => {
  if (sortBy.value !== column) return '↕️'
  return sortDirection.value === 'asc' ? '↑' : '↓'
}

// Dropdown methods
const toggleDropdown = (vmUuid: string) => {
  activeDropdown.value = activeDropdown.value === vmUuid ? null : vmUuid
}

const openVMConsole = (vm: any) => {
  const consoleRoute = getConsoleRoute(vm.hostId, vm.name, vm)
  if (consoleRoute) {
    router.push(consoleRoute)
  }
}

const viewVMDetails = (vm: any) => {
  activeDropdown.value = null
  router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)
}

const cloneVM = (vm: any) => {
  activeDropdown.value = null
  // TODO: Implement clone VM functionality
  console.log('Clone VM:', vm.name)
}

const exportVM = (vm: any) => {
  activeDropdown.value = null
  // TODO: Implement export VM functionality
  console.log('Export VM:', vm.name)
}

const deleteVM = async (vm: any) => {
  activeDropdown.value = null
  // TODO: Add confirmation dialog
  if (confirm(`Are you sure you want to delete VM "${vm.name}"?`)) {
    try {
      // TODO: Implement delete VM functionality
      console.log('Delete VM:', vm.name)
    } catch (error) {
      console.error('Failed to delete VM:', error)
    }
  }
}

// Close dropdown when clicking outside
const handleClickOutside = () => {
  activeDropdown.value = null
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
  await fetchVMsForAllHosts()
  
  // Add click outside listener
  document.addEventListener('click', handleClickOutside)
})

// Cleanup
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>