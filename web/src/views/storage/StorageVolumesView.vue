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

      <!-- Volume Content -->
      <div class="p-6">
        <!-- Cards View -->
        <div v-if="viewMode === 'cards'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          <div
            v-for="volume in sortedPaginatedVolumes"
            :key="volume.id"
            class="group relative bg-slate-800/30 hover:bg-slate-700/40 rounded-xl p-6 border border-slate-600/20 hover:border-slate-500/40 transition-all duration-300 cursor-pointer backdrop-blur-sm card-glow"
            @click="selectVolume(volume)"
          >
            <!-- Card Header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-3">
                <div :class="[
                  'w-4 h-4 rounded-full transition-all duration-300',
                  getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-400 shadow-lg shadow-emerald-400/50' :
                  getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-400 shadow-lg shadow-blue-400/50' :
                  'bg-slate-400'
                ]"></div>
                <div class="min-w-0">
                  <h3 class="text-lg font-semibold text-white truncate group-hover:text-blue-300 transition-colors">
                    {{ volume.name || 'Unnamed Volume' }}
                  </h3>
                  <p class="text-sm text-slate-400 truncate">{{ volume.format.toUpperCase() }}</p>
                </div>
              </div>
              
              <!-- Status Badge -->
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium backdrop-blur-sm border',
                getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' :
                getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-500/20 text-blue-300 border-blue-500/30' :
                'bg-slate-500/20 text-slate-300 border-slate-500/30'
              ]">
                {{ getVolumeStatus(volume.id) }}
              </span>
            </div>

            <!-- Card Content -->
            <div class="space-y-3">
              <!-- Pool Info -->
              <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-3">
                <svg class="w-4 h-4 text-slate-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <div>
                  <div class="text-xs text-slate-400">Pool</div>
                  <div class="text-sm font-medium truncate">{{ getPoolName(volume.storage_pool_id) }}</div>
                </div>
              </div>

              <!-- Volume Info -->
              <div class="grid grid-cols-2 gap-3">
                <!-- Size -->
                <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-3">
                  <svg class="w-4 h-4 text-purple-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                  </svg>
                  <div>
                    <div class="text-xs text-slate-400">Size</div>
                    <div class="text-sm font-medium">{{ formatBytes(volume.capacity_bytes || 0) }}</div>
                  </div>
                </div>

                <!-- Type -->
                <div class="flex items-center gap-2 text-slate-300 bg-slate-700/30 rounded-lg p-3">
                  <svg class="w-4 h-4 text-amber-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M2 5a2 2 0 012-2h8a2 2 0 012 2v10a2 2 0 002 2H4a2 2 0 01-2-2V5zm3 1h6v4H5V6zm6 6H5v2h6v-2z" clip-rule="evenodd" />
                  </svg>
                  <div>
                    <div class="text-xs text-slate-400">Type</div>
                    <div class="text-sm font-medium">{{ volume.type }}</div>
                  </div>
                </div>
              </div>

              <!-- Path Info -->
              <div v-if="volume.path" class="text-xs text-slate-400 truncate mt-2 p-2 bg-slate-800/30 rounded border border-slate-600/20">
                <span class="font-medium">Path:</span> {{ volume.path }}
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="flex justify-end gap-2 mt-4 opacity-0 group-hover:opacity-100 transition-all duration-300">
              <FButton
                variant="ghost"
                size="sm"
                @click.stop="handleVolumeAction('edit', volume)"
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
                @click.stop="handleVolumeAction('clone', volume)"
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
                @click.stop="handleVolumeAction('delete', volume)"
                class="p-2 text-red-400 hover:bg-red-500/20 hover:text-red-300 transition-all duration-200 rounded-lg"
                title="Delete Volume"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
              </FButton>
            </div>
          </div>
        </div>

        <!-- Compact View -->
        <div v-else-if="viewMode === 'compact'" class="space-y-2">
          <div
            v-for="volume in sortedPaginatedVolumes"
            :key="volume.id"
            class="group relative bg-slate-800/30 hover:bg-slate-700/40 rounded-lg p-4 border border-slate-600/20 hover:border-slate-500/40 transition-all duration-300 cursor-pointer backdrop-blur-sm"
            @click="selectVolume(volume)"
          >
            <div class="flex items-center justify-between">
              <!-- Left Side: Status, Name, and Quick Info -->
              <div class="flex items-center gap-4 min-w-0 flex-1">
                <!-- Status Orb -->
                <div :class="[
                  'w-3 h-3 rounded-full transition-all duration-300 flex-shrink-0',
                  getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-400 animate-pulse shadow-lg shadow-emerald-400/50' :
                  getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-400 animate-pulse shadow-lg shadow-blue-400/50' :
                  'bg-slate-400'
                ]"></div>
                
                <!-- Volume Info -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-4">
                    <h3 class="text-base font-semibold text-white truncate group-hover:text-blue-300 transition-colors">
                      {{ volume.name || 'Unnamed Volume' }}
                    </h3>
                    <span class="text-xs text-slate-400 bg-slate-700/50 px-2 py-1 rounded">
                      {{ volume.format.toUpperCase() }}
                    </span>
                    <span class="text-xs text-slate-400">
                      {{ getPoolName(volume.storage_pool_id) }}
                    </span>
                    <span class="text-xs text-slate-300 font-medium">
                      {{ formatBytes(volume.capacity_bytes || 0) }}
                    </span>
                  </div>
                  <div v-if="volume.path" class="text-xs text-slate-500 truncate mt-1">
                    {{ volume.path }}
                  </div>
                </div>
              </div>

              <!-- Right Side: Status and Actions -->
              <div class="flex items-center gap-3">
                <!-- Status Badge -->
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium backdrop-blur-sm border',
                  getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' :
                  getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-500/20 text-blue-300 border-blue-500/30' :
                  'bg-slate-500/20 text-slate-300 border-slate-500/30'
                ]">
                  {{ getVolumeStatus(volume.id) }}
                </span>

                <!-- Actions -->
                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-all duration-300">
                  <FButton
                    variant="ghost"
                    size="sm"
                    @click.stop="handleVolumeAction('edit', volume)"
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
                    @click.stop="toggleVolumeDropdown(volume.id)"
                    class="p-2 text-slate-400 hover:text-white hover:bg-slate-600/30 transition-all duration-200 rounded-lg"
                    title="More Actions"
                  >
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
                    </svg>
                  </FButton>
                </div>
              </div>
            </div>

            <!-- Dropdown Menu -->
            <div 
              v-if="activeVolumeDropdown === volume.id"
              v-click-away="() => activeVolumeDropdown = null"
              class="absolute right-4 top-full mt-1 w-48 bg-slate-800/95 border border-slate-600/50 rounded-lg shadow-lg z-50 backdrop-blur-sm"
            >
              <div class="py-1">
                <button
                  @click.stop="handleVolumeAction('clone', volume)"
                  class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                    <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                  </svg>
                  Clone Volume
                </button>
                <button
                  @click.stop="handleVolumeAction('delete', volume)"
                  class="w-full px-4 py-2 text-left text-sm text-red-400 hover:bg-red-500/20 hover:text-red-300 transition-colors flex items-center gap-3"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                  Delete Volume
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- List View -->
        <div v-else class="space-y-2">

          <!-- Sorting Header -->
          <div class="grid grid-cols-12 gap-4 px-5 py-2 bg-slate-800/20 rounded-lg border border-slate-600/20 backdrop-blur-sm text-sm font-medium text-slate-300">
            <div class="col-span-1"></div> <!-- Status column -->
            <button
              @click="handleSort('name')"
              class="col-span-2 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none text-left"
            >
              Name {{ getSortIcon('name') }}
            </button>
            <button
              @click="handleSort('format')"
              class="col-span-1 hidden md:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Format {{ getSortIcon('format') }}
            </button>
            <button
              @click="handleSort('pool')"
              class="col-span-2 hidden md:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Pool {{ getSortIcon('pool') }}
            </button>
            <button
              @click="handleSort('type')"
              class="col-span-1 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Type {{ getSortIcon('type') }}
            </button>
            <button
              @click="handleSort('capacity_bytes')"
              class="col-span-2 hidden lg:flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Size {{ getSortIcon('capacity_bytes') }}
            </button>
            <button
              @click="handleSort('status')"
              class="col-span-1 flex items-center gap-2 hover:text-white transition-colors cursor-pointer select-none"
            >
              Status {{ getSortIcon('status') }}
            </button>
            <div class="col-span-2 text-center">
              Actions
            </div>
          </div>

          <!-- Volume List Items -->
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
                  getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-400 animate-pulse shadow-lg shadow-emerald-400/50' :
                  getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-400 animate-pulse shadow-lg shadow-blue-400/50' :
                  'bg-slate-400'
                ]"></div>
              </div>

              <!-- Volume Name -->
              <div class="col-span-2 min-w-0">
                <h3 class="text-base font-semibold text-white truncate group-hover:text-blue-300 transition-colors">
                  {{ volume.name || 'Unnamed Volume' }}
                </h3>
                <p v-if="volume.path" class="text-xs text-slate-500 truncate">
                  {{ volume.path }}
                </p>
              </div>

              <!-- Format -->
              <div class="col-span-1 hidden md:flex items-center text-slate-300 min-w-0">
                <span class="text-sm font-medium bg-slate-700/50 px-2 py-1 rounded">
                  {{ volume.format.toUpperCase() }}
                </span>
              </div>

              <!-- Pool Info -->
              <div class="col-span-2 hidden md:flex items-center gap-2 text-slate-300 min-w-0">
                <svg class="w-3 h-3 text-slate-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <span class="text-sm truncate">{{ getPoolName(volume.storage_pool_id) }}</span>
              </div>

              <!-- Type -->
              <div class="col-span-1 flex items-center text-slate-300 min-w-0">
                <span class="text-sm">{{ volume.type }}</span>
              </div>

              <!-- Size Information -->
              <div class="col-span-2 hidden lg:flex items-center gap-1 text-slate-400">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span class="text-sm font-medium">{{ formatBytes(volume.capacity_bytes || 0) }}</span>
              </div>

              <!-- Status Badge -->
              <div class="col-span-1 flex items-center">
                <span :class="[
                  'px-2 py-1 rounded-full text-xs font-medium backdrop-blur-sm border',
                  getVolumeStatus(volume.id) === 'available' ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' :
                  getVolumeStatus(volume.id) === 'in-use' ? 'bg-blue-500/20 text-blue-300 border-blue-500/30' :
                  'bg-slate-500/20 text-slate-300 border-slate-500/30'
                ]">
                  {{ getVolumeStatus(volume.id) }}
                </span>
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

            <!-- Mobile Info (shown on smaller screens) -->
            <div class="lg:hidden mt-3 pt-3 border-t border-slate-600/20 flex items-center gap-4 text-sm text-slate-400">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span>{{ formatBytes(volume.capacity_bytes || 0) }}</span>
              </div>
              <div class="md:hidden flex items-center gap-2">
                <span class="text-slate-300 bg-slate-700/50 px-2 py-1 rounded text-xs">
                  {{ volume.format.toUpperCase() }}
                </span>
              </div>
              <div class="md:hidden flex items-center gap-2">
                <svg class="w-3 h-3 text-slate-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
                <span>{{ getPoolName(volume.storage_pool_id) }}</span>
              </div>
            </div>

            <!-- Usage Information (if in use) -->
            <div v-if="getVolumeUsedBy(volume.id)" class="mt-3 pt-3 border-t border-slate-600/20">
              <div class="flex items-center gap-2 text-sm text-slate-400">
                <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>Used by: <span class="text-blue-300">{{ getVolumeUsedBy(volume.id) }}</span></span>
              </div>
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
import { ref, computed, onMounted, watch } from 'vue'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import FButton from '@/components/ui/FButton.vue'
import { useStorageStore } from '@/stores/storageStore'
import { useUserPreferences, type ViewMode } from '@/composables/useUserPreferences'
import { vClickAway } from '@/directives/clickAway'
import type { StorageVolume, DiskAttachment, StoragePool } from '@/types'

const storageStore = useStorageStore()
const { preferences } = useUserPreferences()

// Reactive data
const storageVolumes = computed(() => storageStore.storageVolumes)
const diskAttachments = computed(() => storageStore.diskAttachments)
const storagePools = computed(() => storageStore.storagePools as StoragePool[])
const selectedVolume = ref<StorageVolume | null>(null)

// View mode settings
const viewMode = ref<ViewMode>('list')
const activeVolumeDropdown = ref<string | null>(null)

// Watch viewMode changes and save to preferences
watch(viewMode, (newMode) => {
  if (preferences.value.storageVolumes) {
    preferences.value.storageVolumes.viewMode = newMode
  }
})

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
const sortField = ref<'name' | 'type' | 'format' | 'capacity_bytes' | 'pool' | 'status'>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')

// Computed properties
const availablePools = computed(() => {
  const pools = new Set(storageVolumes.value.map(volume => volume.storage_pool_id))
  return Array.from(pools).sort()
})

const filteredVolumes = computed(() => {
  let filtered = storageVolumes.value

  // Apply search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(volume =>
      volume.name.toLowerCase().includes(query) ||
      volume.storage_pool_id.toLowerCase().includes(query) ||
      volume.format.toLowerCase().includes(query) ||
      volume.type.toLowerCase().includes(query)
    )
  }

  // Apply pool filter
  if (poolFilter.value) {
    filtered = filtered.filter(volume => volume.storage_pool_id === poolFilter.value)
  }

  // Apply format filter
  if (formatFilter.value) {
    filtered = filtered.filter(volume => volume.format === formatFilter.value)
  }

  return filtered
})

const sortedVolumes = computed(() => {
  const sorted = [...filteredVolumes.value].sort((a, b) => {
    let aValue: any, bValue: any

    switch (sortField.value) {
      case 'pool':
        aValue = getPoolName(a.storage_pool_id)
        bValue = getPoolName(b.storage_pool_id)
        break
      case 'status':
        aValue = getVolumeStatus(a.id)
        bValue = getVolumeStatus(b.id)
        break
      case 'capacity_bytes':
        aValue = a.capacity_bytes || 0
        bValue = b.capacity_bytes || 0
        break
      default:
        aValue = a[sortField.value]
        bValue = b[sortField.value]
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
  filteredVolumes.value.filter(volume => !isVolumeAttached(volume.id)).length
)

const inUseVolumesCount = computed(() =>
  filteredVolumes.value.filter(volume => isVolumeAttached(volume.id)).length
)

const totalCapacity = computed(() =>
  filteredVolumes.value.reduce((sum, volume) => sum + volume.capacity_bytes, 0)
)

const averageVolumeSize = computed(() => {
  if (filteredVolumes.value.length === 0) return 0
  return totalCapacity.value / filteredVolumes.value.length
})

const largestVolumeSize = computed(() => {
  if (filteredVolumes.value.length === 0) return 0
  return Math.max(...filteredVolumes.value.map(volume => volume.capacity_bytes))
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

// Helper functions
const isVolumeAttached = (volumeId: string): boolean => {
  return diskAttachments.value.some(attachment => attachment.disk_id === volumeId)
}

const getPoolName = (poolId?: string) => {
  if (!poolId) return ''
  const pool = storagePools.value.find(p => p.id === poolId)
  return pool ? pool.name : poolId
}

const cycleViewMode = () => {
  const modes: ViewMode[] = ['cards', 'compact', 'list']
  const currentIndex = modes.indexOf(viewMode.value)
  viewMode.value = modes[(currentIndex + 1) % modes.length]
}

const toggleVolumeDropdown = (volumeId: string) => {
  activeVolumeDropdown.value = activeVolumeDropdown.value === volumeId ? null : volumeId
}

const getVolumeStatus = (volumeId: string): string => {
  return isVolumeAttached(volumeId) ? 'in-use' : 'available'
}

const getVolumeUsedBy = (volumeId: string): string | undefined => {
  const attachment = diskAttachments.value.find(att => att.disk_id === volumeId)
  if (attachment) {
    // This would need VM data to get the VM name, for now return the VM UUID
    return attachment.vm_uuid
  }
  return undefined
}

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

const handleSort = (field: 'name' | 'type' | 'format' | 'capacity_bytes' | 'pool' | 'status') => {
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

const setItemsPerPage = (value: any) => {
  // coerce numeric strings to numbers and otherwise accept 'all'
  if (value === 'all') {
    itemsPerPage.value = 'all'
  } else if (typeof value === 'string' && !isNaN(Number(value))) {
    itemsPerPage.value = Number(value)
  } else if (typeof value === 'number') {
    itemsPerPage.value = value
  } else {
    // fallback to 10
    itemsPerPage.value = 10
  }
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
  await Promise.all([
    storageStore.fetchStorageVolumes(),
    storageStore.fetchDiskAttachments()
  ])
}

onMounted(() => {
  loadStorageVolumes()
})
</script>
