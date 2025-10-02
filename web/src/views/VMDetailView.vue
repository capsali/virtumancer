<template>
  <div class="space-y-6">
    <!-- Navigation -->
    <div class="flex items-center justify-between mb-4">
      <FBreadcrumbs />
    </div>

    <!-- VM Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-6">
        <FBackButton :context-actions="vmContextActions" :compact="true" />
        <div>
          <div class="flex items-center gap-3 mb-1">
            <h1 class="text-3xl font-bold text-white">{{ vm?.name || 'Loading...' }}</h1>
            <span v-if="vm?.uuid" class="text-sm text-slate-400 font-mono bg-slate-800/50 px-3 py-1.5 rounded-lg border border-slate-700/30">
              {{ vm.uuid.substring(0, 8) }}
            </span>
          </div>
          <p class="text-slate-400 text-lg">{{ vm?.description || 'VM Details' }}</p>
        </div>
      </div>
      
      <div v-if="vm" class="flex items-center gap-4">
        <!-- Status Badge -->
        <div class="flex items-center gap-3 bg-slate-800/30 px-4 py-2 rounded-xl border border-slate-700/30">
          <div :class="[
            'w-3 h-3 rounded-full shadow-lg',
            getVMStatusColor(vm.state)
          ]"></div>
          <span :class="[
            'px-3 py-1 rounded-full text-sm font-semibold',
            getVMStateBadgeClass(vm.state)
          ]">
            {{ (vm.state || 'UNKNOWN').toLowerCase() }}
          </span>
        </div>
        
        <!-- Task State Indicator -->
        <div v-if="vm.taskState" class="animate-pulse">
          <span class="px-2 py-1 rounded-full text-xs font-medium bg-yellow-500/20 text-yellow-400">
            {{ vm.taskState }}
          </span>
        </div>
        
        <!-- Power Controls -->
        <div class="flex items-center gap-2">
          <!-- Start Button -->
          <FButton
            v-if="vm.state === 'STOPPED'"
            variant="primary"
            size="sm"
            @click="handleVMAction('start')"
            :disabled="!!vm.taskState"
            class="px-3 py-2"
            title="Start VM"
          >
            <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
            </svg>
          </FButton>
          
          <!-- Active VM Controls -->
          <div v-if="vm.state === 'ACTIVE'" class="flex items-center gap-1">
            <FButton
              variant="ghost"
              size="sm"
              @click="showPowerConfirmation('shutdown')"
              :disabled="!!vm.taskState"
              class="px-3 py-2"
              title="Shutdown VM"
            >
              <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="showPowerConfirmation('reboot')"
              :disabled="!!vm.taskState"
              class="px-3 py-2"
              title="Reboot VM"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="showPowerConfirmation('forceOff')"
              :disabled="!!vm.taskState"
              class="px-3 py-2 text-orange-400 hover:bg-orange-500/10"
              title="Force Power Off VM"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v8m4-4a8 8 0 11-8 0"/>
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="showPowerConfirmation('forceReset')"
              :disabled="!!vm.taskState"
              class="px-3 py-2 text-red-400 hover:bg-red-500/10"
              title="Force Reset VM"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </FButton>
            
            <div class="w-px h-6 bg-slate-600 mx-1"></div>
            
            <FButton
              variant="accent"
              size="sm"
              @click="openConsole"
              class="px-3 py-2"
              title="Open Console"
            >
              <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
              </svg>
            </FButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content Layout - Two Columns for Active VMs -->
    <div v-if="vm && vm.state === 'ACTIVE'" class="grid grid-cols-1 xl:grid-cols-3 gap-6">
      <!-- Left Column: Performance Metrics -->
      <div class="xl:col-span-2 space-y-6">
        <FCard class="card-glow">
          <div class="p-6">
            <div class="flex items-center gap-3 mb-6">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                </svg>
              </div>
              <h3 class="text-xl font-bold text-white">Performance Metrics</h3>
            </div>

            <!-- Host disconnected message -->
            <div v-if="!isHostConnected" class="text-center py-8">
              <svg class="w-16 h-16 mx-auto mb-4 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
              </svg>
              <h4 class="text-white font-semibold mb-2">Host Not Connected</h4>
              <p class="text-slate-400 text-sm">Performance metrics are only available when the host is connected.</p>
            </div>

            <!-- Show metrics data if available -->
            <div v-else-if="vmStats" class="space-y-6">
              <!-- CPU and Memory on the left, Disk I/O and Network I/O on the right -->
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <!-- Left Column: CPU and Memory -->
                <div class="space-y-6">
                  <!-- CPU Details -->
                  <div>
                    <div class="flex justify-between items-center mb-3">
                      <span class="text-sm font-medium text-white">CPU Usage</span>
                      <span class="text-sm font-medium text-white">{{ cpuValue.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                      <div
                        class="h-3 bg-gradient-to-r from-purple-500 to-purple-600 rounded-full transition-all duration-500"
                        :style="{ width: `${cpuValue}%` }"
                      ></div>
                    </div>
                    <div class="flex justify-between text-xs text-slate-500 mt-2">
                      <span>{{ vm?.vcpu_count || 0 }} vCPUs allocated</span>
                      <span>{{ cpuLabel }}</span>
                    </div>
                  </div>

                  <!-- Memory Details -->
                  <div>
                    <div class="flex justify-between items-center mb-3">
                      <span class="text-sm font-medium text-white">Memory Usage</span>
                      <span class="text-sm font-medium text-white">{{ vmStats ? formatBytes((vmStats.memory_mb || 0) * 1024 * 1024) : 'N/A' }}</span>
                    </div>
                    <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                      <div
                        class="h-3 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full transition-all duration-500"
                        :style="{ width: vm && vmStats ? `${Math.min(100, ((vmStats.memory_mb || 0) / (vm.memory_bytes || 1)) * 100)}` : '0%' }"
                      ></div>
                    </div>
                    <div class="flex justify-between text-xs text-slate-500 mt-2">
                      <span>{{ vmStats ? formatBytes((vmStats.memory_mb || 0) * 1024 * 1024) : '0 B' }} used</span>
                      <span>{{ vm ? formatBytes(vm.memory_bytes) : 'N/A' }} total</span>
                    </div>
                  </div>
                </div>

                <!-- Right Column: Disk I/O and Network I/O -->
                <div class="space-y-6">
                  <!-- Disk I/O Details -->
                  <div>
                    <div class="flex justify-between items-center mb-3">
                      <span class="text-sm font-medium text-white">Disk I/O</span>
                      <span class="text-sm font-medium text-white">{{ vmStats ? (vmStats.disk_read_kib_per_sec || 0).toFixed(1) : '0' }} / {{ vmStats ? (vmStats.disk_write_kib_per_sec || 0).toFixed(1) : '0' }} KiB/s</span>
                    </div>
                    <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                      <div
                        class="h-3 bg-gradient-to-r from-green-500 via-green-400 to-orange-500 rounded-full transition-all duration-500"
                        :style="{ width: vmStats && (vmStats.disk_read_kib_per_sec || 0) + (vmStats.disk_write_kib_per_sec || 0) > 0 ? '100%' : '0%' }"
                      ></div>
                    </div>
                    <div class="flex justify-between text-xs text-slate-500 mt-2">
                      <span class="text-green-400">Read: {{ vmStats ? (vmStats.disk_read_kib_per_sec || 0).toFixed(1) : '0' }} KiB/s</span>
                      <span class="text-orange-400">Write: {{ vmStats ? (vmStats.disk_write_kib_per_sec || 0).toFixed(1) : '0' }} KiB/s</span>
                    </div>
                  </div>

                  <!-- Network I/O Details -->
                  <div>
                    <div class="flex justify-between items-center mb-3">
                      <span class="text-sm font-medium text-white">Network I/O</span>
                      <span class="text-sm font-medium text-white">{{ vmStats ? (vmStats.network_rx_mbps || vmStats.network_rx_mb || 0).toFixed(2) : '0' }} / {{ vmStats ? (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0).toFixed(2) : '0' }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</span>
                    </div>
                    <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                      <div
                        class="h-3 bg-gradient-to-r from-blue-500 via-blue-400 to-purple-500 rounded-full transition-all duration-500"
                        :style="{ width: vmStats && ((vmStats.network_rx_mbps || vmStats.network_rx_mb || 0) + (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0)) > 0 ? '100%' : '0%' }"
                      ></div>
                    </div>
                    <div class="flex justify-between text-xs text-slate-500 mt-2">
                      <span class="text-blue-400">RX: {{ vmStats ? (vmStats.network_rx_mbps || vmStats.network_rx_mb || 0).toFixed(2) : '0' }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</span>
                      <span class="text-purple-400">TX: {{ vmStats ? (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0).toFixed(2) : '0' }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- No data available -->
            <div v-else class="text-center py-8">
              <svg class="w-16 h-16 mx-auto mb-4 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
              </svg>
              <h4 class="text-white font-semibold mb-2">No Performance Data</h4>
              <p class="text-slate-400 text-sm">Performance metrics are not available for this virtual machine.</p>
            </div>
          </div>
        </FCard>

      </div>

      <!-- Right Column: Console Preview -->
      <div class="xl:col-span-1">
        <div v-if="getConsoleType(vm)" class="space-y-6">
          <FCard class="card-glow cursor-pointer hover:shadow-xl hover:bg-slate-800/20 transition-all duration-300 hover:scale-[1.02]" @click="openConsole">
            <div class="p-4">
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center shadow-xl ring-2 ring-emerald-500/20">
                    <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div>
                    <h4 class="text-base font-bold text-white">Console Preview</h4>
                    <p class="text-sm text-slate-400">{{ getConsoleStatusText() }}</p>
                  </div>
                </div>
                <FButton
                  variant="ghost"
                  size="sm"
                  @click.stop="refreshConsolePreview"
                  title="Refresh Preview"
                  class="hover:bg-emerald-500/10 hover:text-emerald-400 transition-colors"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                  </svg>
                </FButton>
              </div>

              <!-- Console Preview Content -->
              <div class="bg-slate-900/50 rounded-lg border border-slate-700/50 overflow-hidden">
                <div class="bg-black rounded-lg overflow-hidden relative h-60">
                  <!-- Console Screenshot Preview -->
                  <div class="absolute inset-0">
                    <!-- Screenshot Display -->
                    <div v-if="getConsoleType(vm) && vm" class="w-full h-full relative">
                      <!-- Current screenshot -->
                      <img
                        v-if="consoleSnapshot"
                        :src="consoleSnapshot"
                        class="w-full h-full object-contain rounded transition-opacity duration-300"
                        :class="{ 'opacity-70': isCapturingSnapshot }"
                        alt="Console Preview"
                      />
                      <!-- Loading state -->
                      <div v-else class="w-full h-full bg-slate-800 flex items-center justify-center">
                        <div class="text-center text-slate-400">
                          <div class="w-6 h-6 border-2 border-slate-500 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
                          <p class="text-xs">Capturing console screenshot...</p>
                        </div>
                      </div>
                      <!-- Refresh indicator -->
                      <div v-if="isCapturingSnapshot" class="absolute inset-0 flex items-center justify-center bg-black/20 rounded">
                        <div class="bg-black/60 rounded-full p-3">
                          <div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                        </div>
                      </div>
                    </div>
                    <!-- No console available -->
                    <div v-else class="w-full h-full bg-slate-800 flex items-center justify-center">
                      <div class="text-center text-slate-400">
                        <svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                        </svg>
                        <p class="text-xs">No console available</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </FCard>
        </div>
      </div>
    </div>

    <!-- Fallback for inactive VMs -->
    <div v-else-if="vm && vm.state !== 'ACTIVE'" class="text-center py-8">
      <FCard class="card-glow p-8">
        <div class="flex justify-center mb-4">
          <svg class="w-12 h-12 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
          </svg>
        </div>
        <h4 class="text-lg font-semibold text-white mb-2">Virtual Machine Stopped</h4>
        <p class="text-slate-400">Performance metrics and console preview are only available when the VM is running.</p>
      </FCard>
    </div>

    <!-- VM Information - Wide Card -->
    <FCard v-if="vm" class="card-glow cursor-pointer hover:shadow-lg transition-all duration-300" @click="vmDetailsExpanded = !vmDetailsExpanded">
      <div class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-violet-500 flex items-center justify-center shadow-xl ring-2 ring-purple-500/20">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white">Virtual Machine Details</h3>
              <p class="text-slate-400">Hardware configuration and system information</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <!-- Sync Button -->
            <FButton
              variant="ghost"
              size="sm"
              @click.stop="showSyncConfirmModal = true"
              :disabled="!!vm.taskState"
              class="p-2"
              title="Sync VM from libvirt"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
            </FButton>
            <!-- Expand/Collapse Indicator -->
            <div class="flex items-center text-slate-400">
              <svg v-if="!vmDetailsExpanded" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7"/>
              </svg>
            </div>
          </div>
        </div>
        
        <!-- VM Info Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">VM Name</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-white font-medium" :title="vm.name">{{ vm.name }}</span>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Operating System</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-white font-medium">{{ vm.osType || 'Unknown' }}</span>
            </div>
          </div>
          
          <div class="space-y-2 sm:col-span-2 lg:col-span-1">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">UUID</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-slate-300 text-sm break-all font-mono" :title="vm.uuid">{{ vm.uuid }}</span>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">CPU Cores</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-blue-400 rounded-full"></div>
                <span class="text-blue-400 text-lg font-bold">{{ vm.vcpuCount || 0 }}</span>
                <span class="text-slate-400 text-sm">cores</span>
              </div>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Memory</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-purple-400 rounded-full"></div>
                <span class="text-purple-400 text-lg font-bold">{{ formatBytes(vm.memoryBytes || 0) }}</span>
              </div>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Disk Size</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-green-400 text-lg font-bold">{{ vm.diskSizeGB || 0 }}</span>
                <span class="text-slate-400 text-sm">GB</span>
              </div>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Boot Device</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-white font-medium">{{ vm.bootDevice || 'hd' }}</span>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Network</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-white font-medium">{{ vm.networkInterface || 'default' }}</span>
            </div>
          </div>

          <!-- Uptime -->
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Uptime</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span
                class="text-white font-medium"
                :title="uptimeAvailable ? `${uptimeSeconds}s` : 'Unknown'"
              >
                {{ uptimeAvailable ? formatUptime(uptimeSeconds) : '—' }}
              </span>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">CPU Model</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <span class="text-white font-medium">{{ vm.cpuModel || 'Default' }}</span>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Current State</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <div class="flex items-center gap-2">
                <div :class="[
                  'w-2 h-2 rounded-full',
                  vm.state === 'ACTIVE' ? 'bg-green-400 animate-pulse' : 
                  vm.state === 'STOPPED' ? 'bg-red-400' : 'bg-yellow-400'
                ]"></div>
                <span :class="[
                  'font-medium',
                  vm.state === 'ACTIVE' ? 'text-green-400' : 
                  vm.state === 'STOPPED' ? 'text-red-400' : 'text-yellow-400'
                ]">
                  {{ vm.state }}
                </span>
              </div>
            </div>
          </div>
          
          <div class="space-y-2">
            <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Sync Status</span>
            <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
              <div class="flex items-center gap-2">
                <div :class="[
                  'w-2 h-2 rounded-full',
                  vm.syncStatus === 'SYNCED' ? 'bg-green-400' : 
                  vm.syncStatus === 'DRIFTED' ? 'bg-yellow-400' : 'bg-gray-400'
                ]"></div>
                <span :class="[
                  'font-medium',
                  vm.syncStatus === 'SYNCED' ? 'text-green-400' : 
                  vm.syncStatus === 'DRIFTED' ? 'text-yellow-400' : 'text-gray-400'
                ]">
                  {{ vm.syncStatus }}
                </span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Expanded Details Section -->
        <div v-show="vmDetailsExpanded" class="mt-8 pt-6 border-t border-slate-700/50">
          <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
            <!-- Basic Information -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                Basic Information
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Name:</span>
                  <span class="col-span-2 text-white">{{ vm.name }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">UUID:</span>
                  <span class="col-span-2 text-white font-mono text-xs break-all">{{ vm.uuid }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Domain UUID:</span>
                  <span class="col-span-2 text-white font-mono text-xs break-all">{{ vm.domainUuid || 'Unknown' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Title:</span>
                  <span class="col-span-2 text-white">{{ vm.title || vm.name }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Description:</span>
                  <span class="col-span-2 text-white">{{ vm.description || 'No description' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Source:</span>
                  <span class="col-span-2 text-white capitalize">{{ vm.source || 'Unknown' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Template:</span>
                  <span class="col-span-2 text-white">{{ vm.isTemplate ? 'Yes' : 'No' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Created:</span>
                  <span class="col-span-2 text-white text-xs">{{ vm.createdAt ? new Date(vm.createdAt).toLocaleString() : 'Unknown' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Updated:</span>
                  <span class="col-span-2 text-white text-xs">{{ vm.updatedAt ? new Date(vm.updatedAt).toLocaleString() : 'Unknown' }}</span>
                </div>
              </div>
            </FCard>

            <!-- CPU & Compute Configuration -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
                CPU & Compute
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">vCPUs:</span>
                  <span class="col-span-2 text-white">{{ vm.vcpuCount || 0 }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">CPU Model:</span>
                  <span class="col-span-2 text-white">{{ vm.cpuModel || 'host-passthrough' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm" v-if="vm.cpuTopologyJson">
                  <span class="text-slate-400 font-medium">Topology:</span>
                  <span class="col-span-2 text-white text-xs">{{ formatCPUTopology(vm.cpuTopologyJson) }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Hypervisor:</span>
                  <span class="col-span-2 text-white">KVM/QEMU</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Architecture:</span>
                  <span class="col-span-2 text-white">x86_64</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Machine Type:</span>
                  <span class="col-span-2 text-white">pc-q35</span>
                </div>
              </div>
            </FCard>

            <!-- Memory Configuration -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                </svg>
                Memory
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Max Memory:</span>
                  <span class="col-span-2 text-white">{{ formatBytes(vm.memoryBytes || 0) }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Current:</span>
                  <span class="col-span-2 text-white">{{ formatBytes(vm.currentMemory || vm.memoryBytes || 0) }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Unit:</span>
                  <span class="col-span-2 text-white">KiB</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Balloon:</span>
                  <span class="col-span-2 text-white">Enabled</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Hugepages:</span>
                  <span class="col-span-2 text-white">Auto</span>
                </div>
              </div>
            </FCard>

            <!-- Storage Configuration -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
                Storage
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Disk Size:</span>
                  <span class="col-span-2 text-white">{{ vm.diskSizeGB || 0 }} GB</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Boot Device:</span>
                  <span class="col-span-2 text-white">{{ vm.bootDevice || 'hd' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Bus Type:</span>
                  <span class="col-span-2 text-white">VirtIO</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Cache Mode:</span>
                  <span class="col-span-2 text-white">None</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">IO Mode:</span>
                  <span class="col-span-2 text-white">Native</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Discard:</span>
                  <span class="col-span-2 text-white">Unmap</span>
                </div>
              </div>
            </FCard>

            <!-- Network Configuration -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                </svg>
                Network
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Interface:</span>
                  <span class="col-span-2 text-white">{{ vm.networkInterface || 'default' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm" v-if="vm.portGroup">
                  <span class="text-slate-400 font-medium">Port Group:</span>
                  <span class="col-span-2 text-white">{{ vm.portGroup }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Model:</span>
                  <span class="col-span-2 text-white">VirtIO</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Link State:</span>
                  <span class="col-span-2 text-white">Up</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">MAC Address:</span>
                  <span class="col-span-2 text-white font-mono text-xs">Auto-generated</span>
                </div>
              </div>
            </FCard>

            <!-- System & Status -->
            <FCard class="p-6 card-glow">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                </svg>
                System & Status
              </h4>
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Current State:</span>
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
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Libvirt State:</span>
                  <span :class="[
                    'px-2 py-1 rounded-full text-xs font-medium',
                    vm.libvirtState === 'ACTIVE' ? 'bg-green-500/20 text-green-400' :
                    vm.libvirtState === 'STOPPED' ? 'bg-red-500/20 text-red-400' :
                    'bg-yellow-500/20 text-yellow-400'
                  ]">
                    {{ vm.libvirtState || 'Unknown' }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Sync Status:</span>
                  <div class="flex items-center gap-2">
                    <span :class="[
                      'px-2 py-1 rounded-full text-xs font-medium',
                      vm.syncStatus === 'SYNCED' ? 'bg-green-500/20 text-green-400' : 
                      vm.syncStatus === 'DRIFTED' ? 'bg-yellow-500/20 text-yellow-400' : 
                      'bg-gray-500/20 text-gray-400'
                    ]">
                      {{ vm.syncStatus }}
                    </span>
                    <FButton
                      variant="ghost"
                      size="xs"
                      @click="showSyncConfirmModal = true"
                      :disabled="!!vm.taskState"
                      class="px-2 py-1 text-xs"
                      title="Sync VM from libvirt"
                    >
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                      </svg>
                    </FButton>
                  </div>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Task State:</span>
                  <span class="col-span-2 text-white">{{ vm.taskState || 'None' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">OS Type:</span>
                  <span class="col-span-2 text-white">{{ vm.osType || 'Unknown' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Host ID:</span>
                  <span class="col-span-2 text-white font-mono text-xs">{{ vm.hostId }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Graphics:</span>
                  <span class="col-span-2 text-white">{{ vm.graphics || 'SPICE' }}</span>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Advanced XML Metadata -->
          <div v-if="vm.metadata" class="mt-6 pt-6 border-t border-slate-700/50">
            <FCard class="p-6 card-glow">
              <div class="flex items-center justify-between mb-4">
                <h4 class="text-lg font-semibold text-white flex items-center gap-2">
                  <svg class="w-5 h-5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
                  </svg>
                  XML Metadata
                </h4>
              </div>
              <div class="bg-slate-900/50 rounded-lg p-4 border border-slate-700/50">
                <pre class="text-xs text-slate-300 whitespace-pre-wrap font-mono overflow-x-auto">{{ vm.metadata }}</pre>
              </div>
            </FCard>
          </div>
        </div>
      </div>
    </FCard>




    <!-- Loading State -->
    <div v-if="!vm" class="flex items-center justify-center py-12">
      <div class="flex items-center gap-3">
        <div class="w-6 h-6 border-2 border-primary-400 border-t-transparent rounded-full animate-spin"></div>
        <span class="text-white">Loading VM details...</span>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 bg-red-500/10 border border-red-400/20 rounded-lg">
      <p class="text-red-400">{{ error }}</p>
    </div>

    <!-- Metrics Settings Modal (overlay) -->
    <MetricSettingsModal
      v-if="showMetricSettings"
      :show="showMetricSettings"
      @close="showMetricSettings = false"
      @applied="refreshStats"
    />

    <!-- Sync Confirmation Modal -->
    <FModal
      :show="showSyncConfirmModal"
      @close="showSyncConfirmModal = false"
      size="md"
    >
      <FCard class="space-y-6">
        <!-- Header -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center shadow-lg">
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
          </div>
          <div>
            <h3 class="text-xl font-semibold text-white">Sync VM from Libvirt</h3>
            <p class="text-sm text-slate-400">{{ vm?.name }}</p>
          </div>
        </div>
        
        <!-- Content -->
        <div class="space-y-4">
          <p class="text-slate-300">
            This will synchronize the VM configuration from libvirt, updating the database with the current running state.
          </p>
        </div>
        
        <!-- Actions -->
        <div class="flex justify-end gap-3 pt-4 border-t border-slate-700/50">
          <FButton
            variant="ghost"
            @click="showSyncConfirmModal = false"
          >
            Cancel
          </FButton>
          <FButton
            variant="primary"
            @click="confirmSync"
            :disabled="!!vm?.taskState"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            Sync from Libvirt
          </FButton>
        </div>
      </FCard>
    </FModal>

    <!-- Power Action Confirmation Modal -->
    <FModal 
      :show="showPowerConfirmationModal" 
      @close="showPowerConfirmationModal = false"
      size="sm"
    >
      <FCard class="bg-slate-800 border-slate-700">
        <!-- Header -->
        <div class="p-6 border-b border-slate-700/50">
          <div class="flex items-center gap-4">
            <div :class="[
              'w-12 h-12 rounded-xl flex items-center justify-center shadow-xl',
              pendingPowerAction === 'shutdown' || pendingPowerAction === 'reboot' ? 
                'bg-gradient-to-br from-blue-500 to-cyan-500' :
                'bg-gradient-to-br from-red-500 to-orange-500'
            ]">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-xl font-bold text-white">
                Confirm {{ getActionDisplayName(pendingPowerAction) }}
              </h3>
              <p class="text-sm text-slate-400">{{ getActionDescription(pendingPowerAction) }}</p>
            </div>
          </div>
        </div>
        
        <!-- Content -->
        <div class="p-6">
          <div class="space-y-4">
            <div class="bg-slate-900/50 rounded-lg p-4 border border-slate-700/50">
              <div class="flex items-center gap-3 mb-3">
                <div class="w-8 h-8 rounded-lg bg-blue-500/20 flex items-center justify-center">
                  <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                </div>
                <span class="text-blue-400 font-medium">VM Information</span>
              </div>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-slate-400">Name:</span>
                  <span class="text-white">{{ vm?.name }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-slate-400">Current State:</span>
                  <span :class="getVMStateBadgeClass(vm?.state || '')">{{ vm?.state }}</span>
                </div>
              </div>
            </div>
            
            <div v-if="pendingPowerAction === 'forceOff' || pendingPowerAction === 'forceReset'" 
                 class="bg-amber-500/10 border border-amber-400/20 rounded-lg p-4">
              <div class="flex items-center gap-3">
                <svg class="w-5 h-5 text-amber-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
                <div>
                  <p class="text-amber-400 font-medium">Forced Action Warning</p>
                  <p class="text-amber-300 text-sm mt-1">This action bypasses the guest OS and may cause data loss.</p>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- Actions -->
        <div class="flex justify-end gap-3 p-6 border-t border-slate-700/50">
          <FButton
            variant="ghost"
            @click="showPowerConfirmationModal = false"
          >
            Cancel
          </FButton>
          <FButton
            :variant="pendingPowerAction === 'forceOff' || pendingPowerAction === 'forceReset' ? 'danger' : 'primary'"
            @click="confirmPowerAction"
            :disabled="!!vm?.taskState"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
            </svg>
            {{ getActionDisplayName(pendingPowerAction) }}
          </FButton>
        </div>
      </FCard>
    </FModal>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useVMStore } from '@/stores/vmStore';
import { useUIStore } from '@/stores/uiStore';
import { useSettingsStore } from '@/stores/settingsStore';
import { useHostStore } from '@/stores/hostStore';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';
import FBackButton from '@/components/ui/FBackButton.vue';
import FModal from '@/components/ui/FModal.vue';
import VMHardwareConfigModalExtended from '@/components/modals/VMHardwareConfigModalExtended.vue';
import MetricSettingsModal from '@/components/modals/MetricSettingsModal.vue';
// Types will be inferred from store usage
import { wsManager } from '@/services/api';
import { getConsoleRoute, getConsoleType, getConsoleDisplayName } from '@/utils/console';
// @ts-ignore
import RFB from '@novnc/novnc/lib/rfb';

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const route = useRoute();
const router = useRouter();

const vmStore = useVMStore();
const uiStore = useUIStore();
const hostStore = useHostStore();

// Component state
const vm = ref<any>(null);
const vmStats = ref<any>(null);
const error = ref<string | null>(null);
const loadingStats = ref(false);
// simplified CPU display: show smoothed host-normalized `cpu_percent`
const showMetricSettings = ref(false);
const vmDetailsExpanded = ref(false);
const showSyncConfirmModal = ref(false);
const showPowerConfirmationModal = ref(false);
const pendingPowerAction = ref<string>('');

// Host connection state
const isHostConnected = computed(() => {
  const host = hostStore.hosts.find((h: any) => h.id === props.hostId);
  return host ? host.state === 'CONNECTED' : false;
});

// Console preview state
const consoleConnected = ref(false);
const consoleRefreshKey = ref(0);
const consoleSnapshot = ref<string | null>(null);
const lastSnapshotTime = ref<number>(0);
const isCapturingSnapshot = ref<boolean>(false);

// ...existing code...

const refreshConsolePreview = () => {
  consoleConnected.value = false;
  consoleRefreshKey.value++;
  
  // Capture screenshot for any console type
  if (vm.value && getConsoleType(vm.value)) {
    captureConsoleScreenshot();
  }
};

// Capture console screenshot for preview (works for both VNC and SPICE)
const captureConsoleScreenshot = async () => {
  if (!vm.value) return;
  
  isCapturingSnapshot.value = true;
  
  try {
    const consoleType = getConsoleType(vm.value);
    
    // Create a temporary canvas for screenshot
    const canvas = document.createElement('canvas');
    canvas.width = 320;
    canvas.height = 240;
    const ctx = canvas.getContext('2d');
    
    if (!ctx) {
      isCapturingSnapshot.value = false;
      return;
    }
    
    // Create temporary container for console connection
    const tempDiv = document.createElement('div');
    tempDiv.style.display = 'none';
    tempDiv.style.position = 'absolute';
    tempDiv.style.top = '-9999px';
    document.body.appendChild(tempDiv);
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    
    let connection: any = null;
    let cleanupTimeout: number | null = null;
    
    const cleanup = () => {
      if (cleanupTimeout) {
        clearTimeout(cleanupTimeout);
        cleanupTimeout = null;
      }
      
      if (connection) {
        try {
          connection.disconnect();
        } catch (e) {
          console.debug('Error disconnecting console:', e);
        }
        connection = null;
      }
      
      if (document.body.contains(tempDiv)) {
        document.body.removeChild(tempDiv);
      }
      
      isCapturingSnapshot.value = false;
    };
    
    const captureFrame = () => {
      try {
        // Find the canvas element created by the console connection
        const consoleCanvas = tempDiv.querySelector('canvas') as HTMLCanvasElement;
        if (consoleCanvas && consoleCanvas.width > 0 && consoleCanvas.height > 0) {
          // Draw console canvas to our screenshot canvas
          ctx.drawImage(consoleCanvas, 0, 0, 320, 240);
          
          // Convert to data URL
          consoleSnapshot.value = canvas.toDataURL('image/png');
          lastSnapshotTime.value = Date.now();
          
          console.log(`Console screenshot captured for ${consoleType}`);
        } else {
          console.warn('No valid canvas found for screenshot');
        }
      } catch (error) {
        console.warn('Failed to capture console screenshot:', error);
      } finally {
        // Clean up after capturing
        setTimeout(cleanup, 100);
      }
    };
    
    // Set up automatic cleanup after 10 seconds max
    cleanupTimeout = setTimeout(() => {
      console.warn('Console screenshot capture timed out');
      cleanup();
    }, 10000);
    
    if (consoleType === 'vnc') {
      // VNC connection using noVNC RFB
      const wsUrl = `${protocol}//${host}/api/v1/hosts/${props.hostId}/vms/${props.vmName}/console`;
      
      // @ts-ignore
      connection = new RFB(tempDiv, wsUrl, { credentials: {} });
      connection.viewOnly = true;
      connection.scaleViewport = false;
      
      connection.addEventListener('connect', () => {
        console.log('VNC connected for screenshot');
        // Wait a bit for frame to render, then capture
        setTimeout(captureFrame, 1500);
      });
      
      connection.addEventListener('disconnect', () => {
        console.log('VNC disconnected after screenshot');
      });
      
      connection.addEventListener('securityfailure', () => {
        console.warn('VNC security failure during screenshot');
        cleanup();
      });
      
    } else if (consoleType === 'spice') {
      // For SPICE, create a temporary iframe to capture the console
      const iframe = document.createElement('iframe');
      iframe.style.position = 'absolute';
      iframe.style.top = '-9999px';
      iframe.style.width = '320px';
      iframe.style.height = '240px';
      iframe.style.border = 'none';
      iframe.scrolling = 'no';
      
      // Build SPICE iframe URL
      const spiceHost = window.location.hostname;
      const spicePort = window.location.port || (window.location.protocol === 'https:' ? '443' : '80');
      const path = `api/v1/hosts/${props.hostId}/vms/${props.vmName}/spice`;
      const params = new URLSearchParams({
        host: spiceHost,
        port: spicePort,
        path,
        autoconnect: '1',
        resize: 'scale',
        show_control: '0'
      });
      iframe.src = `/spice/spice_responsive.html?${params.toString()}`;
      
      tempDiv.appendChild(iframe);
      connection = iframe;
      
      // Wait for iframe to load and SPICE to connect
      iframe.onload = () => {
        console.log('SPICE iframe loaded for screenshot');
        
        // Wait for SPICE connection to establish and render
        setTimeout(() => {
          try {
            // Try to capture the iframe content
            const iframeDoc = iframe.contentDocument || iframe.contentWindow?.document;
            if (iframeDoc) {
              const spiceCanvas = iframeDoc.querySelector('canvas') as HTMLCanvasElement;
              if (spiceCanvas && spiceCanvas.width > 0 && spiceCanvas.height > 0) {
                // Draw SPICE canvas to our screenshot canvas
                ctx.drawImage(spiceCanvas, 0, 0, 320, 240);
                
                consoleSnapshot.value = canvas.toDataURL('image/png');
                lastSnapshotTime.value = Date.now();
                
                console.log('SPICE screenshot captured successfully');
              } else {
                console.warn('No SPICE canvas found, creating placeholder');
                // Create a better placeholder that indicates SPICE is available
                ctx.fillStyle = '#0f0f0f';
                ctx.fillRect(0, 0, 320, 240);
                
                // Draw a subtle SPICE logo/indicator
                ctx.fillStyle = '#1f2937';
                ctx.fillRect(20, 20, 280, 200);
                
                ctx.fillStyle = '#6366f1';
                ctx.font = 'bold 16px Arial';
                ctx.textAlign = 'center';
                ctx.fillText('SPICE Console', 160, 110);
                
                ctx.fillStyle = '#9ca3af';
                ctx.font = '12px Arial';
                ctx.fillText('Click to connect', 160, 135);
                
                consoleSnapshot.value = canvas.toDataURL('image/png');
                lastSnapshotTime.value = Date.now();
                
                console.log('SPICE placeholder screenshot created');
              }
            } else {
              console.warn('Cannot access SPICE iframe content (cross-origin)');
              // Fallback for cross-origin iframe
              ctx.fillStyle = '#0f0f0f';
              ctx.fillRect(0, 0, 320, 240);
              ctx.fillStyle = '#6366f1';
              ctx.font = 'bold 14px Arial';
              ctx.textAlign = 'center';
              ctx.fillText('SPICE Console', 160, 115);
              ctx.fillStyle = '#9ca3af';
              ctx.font = '11px Arial';
              ctx.fillText('Preview unavailable', 160, 135);
              
              consoleSnapshot.value = canvas.toDataURL('image/png');
              lastSnapshotTime.value = Date.now();
            }
          } catch (error) {
            console.warn('Error capturing SPICE screenshot:', error);
            // Create error placeholder
            ctx.fillStyle = '#1f1f1f';
            ctx.fillRect(0, 0, 320, 240);
            ctx.fillStyle = '#ef4444';
            ctx.font = '12px Arial';
            ctx.textAlign = 'center';
            ctx.fillText('Preview Error', 160, 120);
            
            consoleSnapshot.value = canvas.toDataURL('image/png');
            lastSnapshotTime.value = Date.now();
          } finally {
            cleanup();
          }
        }, 3000); // Wait 3 seconds for SPICE to load and render
      };
      
      iframe.onerror = () => {
        console.warn('SPICE iframe failed to load');
        cleanup();
      };
    }
    
  } catch (error) {
    console.warn('Failed to setup console screenshot:', error);
    isCapturingSnapshot.value = false;
  }
};

// Auto-refresh console preview every 15 seconds when active
let consoleRefreshInterval: number | null = null;

const startConsoleRefresh = () => {
  if (!isHostConnected.value) {
    return; // Don't start console refresh if host is disconnected
  }

  if (consoleRefreshInterval) {
    clearInterval(consoleRefreshInterval);
  }
  
  // Initial screenshot for any console type
  if (vm.value && getConsoleType(vm.value)) {
    setTimeout(() => captureConsoleScreenshot(), 2000); // Initial delay to let VM load
  }
  
  consoleRefreshInterval = setInterval(() => {
    if (vm.value?.state === 'ACTIVE' && getConsoleType(vm.value) && isHostConnected.value) {
      refreshConsolePreview();
    }
  }, 30000); // Refresh every 30 seconds to avoid too frequent connections
};

const stopConsoleRefresh = () => {
  if (consoleRefreshInterval) {
    clearInterval(consoleRefreshInterval);
    consoleRefreshInterval = null;
  }
};

// Console status text
const getConsoleStatusText = (): string => {
  if (!vm.value) return 'Loading...';

  if (vm.value.state !== 'ACTIVE') {
    return 'Start VM to access console';
  }

  const consoleType = getConsoleType(vm.value);
  if (!consoleType) {
    return 'No graphics console available';
  }

  return `${getConsoleDisplayName(vm.value)} console ready`;
};// Context actions for the back button
const vmContextActions = computed(() => [
  // Clone and Export actions removed as requested
]);

const settings = useSettingsStore();

function formatDisk(valueKiB: number) {
  if (settings.units.disk === 'mib') return (valueKiB/1024).toFixed(2) + ' MiB/s'
  return valueKiB.toFixed(1) + ' KiB/s'
}

function formatNetwork(valueMBps: number) {
  if (settings.units.network === 'kb') return (valueMBps*1024).toFixed(1) + ' KB/s'
  return valueMBps.toFixed(2) + ' MB/s'
}

const cpuValue = computed(() => {
  if (!vmStats.value) return 0
  const s = settings.cpuDisplayDefault
  if (s === 'guest') return (vmStats.value.cpu_percent_guest ?? vmStats.value.cpu_percent ?? 0)
  if (s === 'raw') return (vmStats.value.cpu_percent_raw ?? vmStats.value.cpu_percent_core ?? vmStats.value.cpu_percent ?? 0)
  // default host
  return (vmStats.value.cpu_percent_host ?? vmStats.value.cpu_percent ?? 0)
})

const cpuLabel = computed(() => {
  const s = settings.cpuDisplayDefault
  if (s === 'guest') return 'Guest'
  if (s === 'raw') return 'Raw %'
  return 'Host'
})

// Uptime helpers: prefer vmStats.uptime (seconds), fallback to vm.uptime if available
const uptimeSeconds = computed(() => {
  if (vmStats.value && typeof vmStats.value.uptime === 'number') return vmStats.value.uptime;
  // vm may not have an uptime field; guard access
  // @ts-ignore - vm shape may not include uptime in types
  if (vm.value && typeof (vm.value as any).uptime === 'number') return (vm.value as any).uptime;
  return 0;
});

const uptimeAvailable = computed(() => {
  return (vmStats.value && typeof vmStats.value.uptime === 'number' && vmStats.value.uptime > 0) ||
    (vm.value && typeof (vm.value as any).uptime === 'number' && (vm.value as any).uptime > 0);
});

// Get VM data
const loadVM = async (): Promise<void> => {
  try {
    error.value = null;
    
    // Stop current monitoring if active
    stopStatsMonitoring();
    
    await vmStore.fetchVMs(props.hostId);
    
    // Find the VM by name
    const foundVM = vmStore.vmsByHost(props.hostId).find((v: any) => v.name === props.vmName);
    if (foundVM) {
      vm.value = foundVM;
      
      // Start monitoring if VM is active
      if (foundVM.state === 'ACTIVE') {
        startStatsMonitoring();
      }
    } else {
      error.value = `VM "${props.vmName}" not found`;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load VM details';
  }
};

// Watch for route parameter changes to reload VM data
watch(() => [props.hostId, props.vmName], () => {
  loadVM().then(() => {
    startStatsMonitoring();
    startConsoleRefresh();
  });
}, { immediate: false });

// Refresh performance stats
const refreshStats = async (): Promise<void> => {
  if (!vm.value || vm.value.state !== 'ACTIVE') return;
  
  loadingStats.value = true;
  try {
    await vmStore.fetchVMStats(props.hostId, vm.value.name);
    // Get stats from store after fetching
    const statsKey = `${props.hostId}:${vm.value.name}`;
    vmStats.value = vmStore.vmStats[statsKey] || null;
  } catch (err) {
    console.error('Failed to load VM stats:', err);
  } finally {
    loadingStats.value = false;
  }
};

// Handle VM actions
const handleVMAction = async (action: string): Promise<void> => {
  if (!vm.value) return;
  
  try {
    error.value = null;
    
    switch (action) {
      case 'start':
        await vmStore.startVM(props.hostId, vm.value.name);
        break;
      case 'shutdown':
        await vmStore.stopVM(props.hostId, vm.value.name);
        break;
      case 'reboot':
        await vmStore.restartVM(props.hostId, vm.value.name);
        break;
      case 'forceOff':
        await vmStore.forceOffVM(props.hostId, vm.value.name);
        break;
      case 'forceReset':
        await vmStore.forceResetVM(props.hostId, vm.value.name);
        break;

      case 'rebuild':
        await vmStore.rebuildVM(props.hostId, vm.value.name);
        break;
    }
    
    // Refresh VM data after action
    await loadVM();
  } catch (err) {
    error.value = err instanceof Error ? err.message : `Failed to ${action} VM`;
  }
};

// Show power confirmation for destructive actions
const showPowerConfirmation = (action: string): void => {
  pendingPowerAction.value = action;
  showPowerConfirmationModal.value = true;
};

// Confirm power action
const confirmPowerAction = async (): Promise<void> => {
  if (!pendingPowerAction.value) return;
  
  showPowerConfirmationModal.value = false;
  await handleVMAction(pendingPowerAction.value);
  pendingPowerAction.value = '';
};

// Open console
const openConsole = (): void => {
  if (!vm.value) return;
  
  const consoleRoute = getConsoleRoute(props.hostId, vm.value.name, vm.value);
  if (consoleRoute) {
    router.push(consoleRoute);
  } else {
    uiStore.addToast('No console available for this VM', 'warning');
  }
};

// Confirm sync from libvirt
const confirmSync = async (): Promise<void> => {
  if (!vm.value) return;
  
  try {
    showSyncConfirmModal.value = false;
    await vmStore.syncVM(props.hostId, vm.value.name);
    await loadVM();
    uiStore.addToast('VM synchronized from libvirt successfully', 'success');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to sync VM from libvirt';
    uiStore.addToast('Failed to sync VM from libvirt', 'error');
  }
};

// Utility functions
const getVMStatusColor = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-400';
    case 'STOPPED': return 'bg-red-400';
    case 'PAUSED': return 'bg-yellow-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-gray-400';
  }
};

const getVMStateBadgeClass = (state: string): string => {
  switch (state) {
    case 'ACTIVE': return 'bg-green-500/20 text-green-400';
    case 'STOPPED': return 'bg-red-500/20 text-red-400';
    case 'PAUSED': return 'bg-yellow-500/20 text-yellow-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-gray-500/20 text-gray-400';
  }
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) {
    return `${days}d ${hours}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
};

const formatCPUTopology = (topologyJson: string): string => {
  try {
    const topology = JSON.parse(topologyJson);
    if (topology.sockets && topology.cores && topology.threads) {
      return `${topology.sockets}s × ${topology.cores}c × ${topology.threads}t`;
    }
    return 'Unknown';
  } catch {
    return 'Parse error';
  }
};

const getActionDisplayName = (action: string): string => {
  switch (action) {
    case 'shutdown': return 'Shutdown';
    case 'reboot': return 'Reboot';
    case 'forceOff': return 'Force Power Off';
    case 'forceReset': return 'Force Reset';
    default: return action;
  }
};

const getActionDescription = (action: string): string => {
  switch (action) {
    case 'shutdown': return 'Graceful shutdown through guest OS';
    case 'reboot': return 'Graceful restart through guest OS';
    case 'forceOff': return 'This action bypasses the guest OS';
    case 'forceReset': return 'This action bypasses the guest OS';
    default: return 'This action cannot be undone';
  }
};

// WebSocket-based stats monitoring
let isSubscribed = false;

const startStatsMonitoring = (): void => {
  if (vm.value?.state === 'ACTIVE' && !isSubscribed && isHostConnected.value) {
    console.log(`Starting stats monitoring for VM: ${props.hostId}/${vm.value.name}`);
    
    // Always do an initial fetch first
    refreshStats();
    
    // Connect WebSocket for real-time updates
    wsManager.connect().then(() => {
      if (vm.value?.state === 'ACTIVE' && !isSubscribed && isHostConnected.value) {
        console.log(`Subscribing to VM stats WebSocket: ${props.hostId}/${vm.value.name}`);
        wsManager.subscribeToVMStats(props.hostId, vm.value.name);
        isSubscribed = true;
        
        // Listen for stats updates
        wsManager.on('vm-stats-updated', handleStatsUpdate);
      }
    }).catch(error => {
      console.error('Failed to connect WebSocket for stats:', error);
      // Continue with periodic fetch fallback
      startStatsFallback();
    });
  } else if (vm.value?.state === 'ACTIVE' && isHostConnected.value) {
    // If already subscribed but VM might have changed state, refresh stats
    refreshStats();
  }
};

// Fallback to periodic polling if WebSocket fails
let statsFallbackInterval: number | null = null;

const startStatsFallback = (): void => {
  if (statsFallbackInterval) {
    clearInterval(statsFallbackInterval);
  }
  statsFallbackInterval = setInterval(() => {
    if (vm.value?.state === 'ACTIVE') {
      refreshStats();
    }
  }, 5000); // Poll every 5 seconds
};

const stopStatsFallback = (): void => {
  if (statsFallbackInterval) {
    clearInterval(statsFallbackInterval);
    statsFallbackInterval = null;
  }
};

const stopStatsMonitoring = (): void => {
  if (isSubscribed && vm.value) {
    console.log(`Unsubscribing from VM stats: ${props.hostId}/${vm.value.name}`);
    wsManager.unsubscribeFromVMStats(props.hostId, vm.value.name);
    wsManager.off('vm-stats-updated', handleStatsUpdate);
    isSubscribed = false;
  }
  stopStatsFallback();
};

// Handle incoming WebSocket stats updates
const handleStatsUpdate = (data: any): void => {
  if (data.hostId === props.hostId && data.vmName === vm.value?.name) {
    // Update vmStats with the received data
    if (data.stats) {
      vmStats.value = data.stats;
      console.log('Received VM stats update via WebSocket:', data.stats);
    }
  }
};

// Lifecycle
onMounted(() => {
  loadVM().then(() => {
    // Only start monitoring if host is connected
    if (isHostConnected.value) {
      startStatsMonitoring();
      startConsoleRefresh();
    }
  });
});

onUnmounted(() => {
  stopStatsMonitoring();
  stopConsoleRefresh();
});
</script>