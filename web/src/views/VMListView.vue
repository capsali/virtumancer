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
    <div v-if="filteredVMs.length > 0 && viewMode === 'list'" class="overflow-x-auto">
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
                <div class="flex items-center gap-1">
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="handleVMAction(vm, 'start')"
                  :disabled="vm.state === 'ACTIVE' || !!vm.taskState"
                  class="text-xs p-1 xl:px-2"
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
                  class="text-xs p-1 xl:px-2"
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
                  class="text-xs p-1 xl:px-2"
                  title="Open Console"
                >
                  <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                </FButton>
                <!-- Dropdown Menu -->
                <div class="relative">
                  <FButton
                    variant="ghost"
                    size="sm"
                    @click.stop="toggleDropdown(vm.uuid)"
                    class="text-xs p-1 xl:px-2"
                    title="More Actions"
                  >
                    <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
                    </svg>
                  </FButton>
                  <div 
                    v-if="activeDropdown === vm.uuid"
                    class="absolute right-0 top-full mt-1 w-48 bg-slate-800 border border-slate-600/50 rounded-lg shadow-lg z-50 card-glow"
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
                        @click.stop="editVM(vm)"
                        class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                      >
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                        </svg>
                        Edit Settings
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
    <div v-if="filteredVMs.length > 0 && viewMode === 'compact'" class="space-y-1">
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



    <!-- Empty State -->
    <div v-else-if="filteredVMs.length === 0" class="text-center py-12">
      <div class="flex justify-center mb-4">
        <svg class="w-16 h-16 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-white mb-2">No Virtual Machines Found</h3>
      <p class="text-slate-400">No virtual machines match your current filters.</p>
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

const router = useRouter()
const hostStore = useHostStore()
const vmStore = useVMStore()
const { vmListPreferences } = useUserPreferences()

// Reactive data
const searchQuery = ref('')
const statusFilter = ref('all')
const hostFilter = ref('all')
const activeDropdown = ref<string | null>(null)

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
    const matchesSearch = !searchQuery.value ||
      vm.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (vm.osType && vm.osType.toLowerCase().includes(searchQuery.value.toLowerCase()))

    const matchesStatus = statusFilter.value === 'all' || vm.state === statusFilter.value
    const matchesHost = hostFilter.value === 'all' || vm.hostId === hostFilter.value

    return matchesSearch && matchesStatus && matchesHost
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

const editVM = (vm: any) => {
  activeDropdown.value = null
  // TODO: Implement edit VM functionality
  console.log('Edit VM:', vm.name)
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