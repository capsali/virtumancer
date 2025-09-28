<template>
  <div class="space-y-6">
    <!-- Navigation -->
    <div class="flex items-center justify-between mb-4">
      <FBreadcrumbs />
    </div>

    <!-- VM Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <FBackButton :context-actions="vmContextActions" />
        <div>
          <h1 class="text-2xl font-bold text-white">{{ vm?.name || 'Loading...' }}</h1>
          <p class="text-slate-400">{{ vm?.description || 'VM Details' }}</p>
        </div>
      </div>
      
      <div v-if="vm" class="flex items-center gap-3">
        <!-- Status Badge -->
        <div class="flex items-center gap-2">
          <div :class="[
            'w-3 h-3 rounded-full',
            getVMStatusColor(vm.state)
          ]"></div>
          <span :class="[
            'px-3 py-1 rounded-full text-sm font-medium',
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
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
            </svg>
          </FButton>
          
          <!-- Active VM Controls -->
          <div v-if="vm.state === 'ACTIVE'" class="flex items-center gap-1">
            <FButton
              variant="ghost"
              size="sm"
              @click="handleVMAction('shutdown')"
              :disabled="!!vm.taskState"
              class="px-3 py-2"
              title="Shutdown VM"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="handleVMAction('reboot')"
              :disabled="!!vm.taskState"
              class="px-3 py-2"
              title="Reboot VM"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="handleVMAction('forceOff')"
              :disabled="!!vm.taskState"
              class="px-3 py-2 text-orange-400 hover:bg-orange-500/10"
              title="Force Off VM"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </FButton>
            
            <FButton
              variant="ghost"
              size="sm"
              @click="handleVMAction('forceReset')"
              :disabled="!!vm.taskState"
              class="px-3 py-2 text-red-400 hover:bg-red-500/10"
              title="Force Reset VM"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
              </svg>
            </FButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Performance Metrics - Compact Cards -->
    <div v-if="vm && vm.state === 'ACTIVE'" class="space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-xl font-bold text-white">Performance Metrics</h3>
        <div class="flex items-center gap-3">
          <FButton variant="outline" size="sm" @click="showMetricSettings = true" title="Metrics Settings">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            </svg>
          </FButton>
          <FButton
            variant="ghost"
            size="sm"
            @click="refreshStats"
            :disabled="loadingStats"
          >
            <span v-if="!loadingStats" class="flex items-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
              Refresh
            </span>
            <span v-else class="flex items-center gap-2">
              <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              Loading...
            </span>
          </FButton>
        </div>
      </div>
      
      <div v-if="vmStats" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <!-- CPU Usage -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 00-2 2zM9 9h6v6H9V9z"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">CPU</h4>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-blue-400">{{ cpuValue.toFixed(1) }}%</div>
              <div class="text-xs text-slate-400 mt-1">{{ cpuLabel }}</div>
            </div>
          </div>
        </FCard>
        
        <!-- Memory Usage -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Memory</h4>
            </div>
            <div class="text-center">
              <div class="text-xl font-bold text-purple-400">{{ formatBytes((vmStats.memory_mb || 0) * 1024 * 1024) }}</div>
              <div class="text-xs text-slate-400 mt-1">Usage</div>
            </div>
          </div>
        </FCard>
        
        <!-- Disk I/O -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500 to-teal-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3-3m0 0l-3 3m3-3v12"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Disk I/O</h4>
            </div>
            <div class="text-center space-y-2">
              <div class="text-sm font-medium text-green-400">
                Read: {{ (vmStats.disk_read_kib_per_sec || 0).toFixed(1) }} KiB/s
              </div>
              <div class="text-sm font-medium text-green-300">
                Write: {{ (vmStats.disk_write_kib_per_sec || 0).toFixed(1) }} KiB/s
              </div>
            </div>
          </div>
        </FCard>
        
        <!-- Network -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Network</h4>
            </div>
            <div class="text-center space-y-2">
              <div class="text-sm font-medium text-cyan-400">
                RX: {{ (vmStats.network_rx_mbps || vmStats.network_rx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}
              </div>
              <div class="text-sm font-medium text-cyan-300">
                TX: {{ (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}
              </div>
            </div>
          </div>
        </FCard>
        
        <!-- Uptime -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Uptime</h4>
            </div>
            <div class="text-center">
              <div class="text-xl font-bold text-orange-400">{{ formatUptime(vmStats.uptime || 0) }}</div>
              <div class="text-xs text-slate-400 mt-1">Running</div>
            </div>
          </div>
        </FCard>
      </div>
      
      <div v-else class="text-center py-8">
        <FCard class="card-glow p-8">
          <div class="flex justify-center mb-4">
            <svg class="w-12 h-12 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 6.707 6.293a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
          </div>
          <h4 class="text-lg font-semibold text-white mb-2">No Performance Data Available</h4>
          <p class="text-slate-400">Performance metrics are only available when the VM is running.</p>
        </FCard>
      </div>
    </div>

    <!-- Console Preview Card (Compact) -->
    <div v-if="vm && vm.state === 'ACTIVE' && getConsoleType(vm)" class="mb-6">
      <FCard class="card-glow">
        <div class="p-4">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center">
                <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="text-lg font-bold text-white">Console Preview</h4>
                <p class="text-xs text-slate-400">{{ getConsoleStatusText() }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <FButton
                variant="ghost"
                size="sm"
                @click="refreshConsolePreview"
                class="p-1"
                title="Refresh Preview"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                </svg>
              </FButton>
              <FButton
                variant="accent"
                size="sm"
                @click="openConsole"
                class="px-3 py-1 text-xs"
              >
                Open Full
              </FButton>
            </div>
          </div>
          
          <!-- Console Preview Content -->
          <div class="bg-slate-900/50 rounded-lg border border-slate-700/50 overflow-hidden">
            <div class="aspect-[16/10] bg-black rounded-lg overflow-hidden relative">
              <!-- Direct console connection attempt -->
              <div class="absolute inset-0">
                <iframe
                  ref="consolePreviewIframe"
                  v-if="consolePreviewSrc"
                  :src="consolePreviewSrc"
                  class="w-full h-full border-0 pointer-events-none"
                  :title="`${vm.name} Console Preview`"
                  scrolling="no"
                  frameborder="0"
                  @load="onConsolePreviewLoad"
                />
                <div v-else class="w-full h-full bg-slate-800 flex items-center justify-center">
                  <div class="text-center text-slate-400">
                    <div class="w-4 h-4 border-2 border-slate-500 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
                    <p class="text-xs">Connecting to console...</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Console Preview Card (Compact) -->
    <div v-if="vm && vm.state === 'ACTIVE' && getConsoleType(vm)" class="mb-6">
      <FCard class="card-glow">
        <div class="p-4">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center">
                <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="text-lg font-bold text-white">Console Preview</h4>
                <p class="text-xs text-slate-400">{{ getConsoleStatusText() }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <FButton
                variant="ghost"
                size="sm"
                @click="refreshConsolePreview"
                class="p-1"
                title="Refresh Preview"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                </svg>
              </FButton>
              <FButton
                variant="accent"
                size="sm"
                @click="openConsole"
                class="px-3 py-1 text-xs"
              >
                Open Full
              </FButton>
            </div>
          </div>
          
          <!-- Console Preview Content -->
          <div class="bg-slate-900/50 rounded-lg border border-slate-700/50 overflow-hidden">
            <div class="aspect-[16/10] bg-black rounded-lg overflow-hidden relative">
              <!-- Console Preview Display -->
              <div v-if="vm.state === 'ACTIVE'" class="w-full h-full bg-slate-900 flex items-center justify-center">
                <div class="text-center text-slate-400">
                  <svg class="w-12 h-12 mx-auto mb-3 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                  <h5 class="text-sm font-semibold text-white mb-1">Console Ready</h5>
                  <p class="text-xs">Click "Open Full" to access console</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- VM Information - Wide Card -->
    <FCard v-if="vm" class="card-glow cursor-pointer hover:shadow-lg transition-all duration-300" @click="vmDetailsExpanded = !vmDetailsExpanded">
      <div class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-4">
            <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500 to-violet-500 flex items-center justify-center shadow-lg">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-2xl font-bold text-white">Virtual Machine Details</h3>
              <p class="text-slate-400">Hardware configuration and system information</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <!-- Settings Cog Wheel -->
            <FButton
              variant="ghost"
              size="sm"
              @click.stop="showExtendedHardwareModal = true"
              class="p-2"
              title="Hardware Settings"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
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
                <span class="text-purple-400 text-lg font-bold">{{ formatBytes((vm.memoryMB || 0) * 1024 * 1024) }}</span>
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
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Basic Information -->
            <div class="space-y-4">
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
                  <span class="text-slate-400 font-medium">Description:</span>
                  <span class="col-span-2 text-white">{{ vm.description || 'No description' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Source:</span>
                  <span class="col-span-2 text-white capitalize">{{ vm.source || 'Unknown' }}</span>
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
            </div>

            <!-- Hardware Configuration -->
            <div class="space-y-4">
              <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
                Hardware Configuration
              </h4>
              <div class="space-y-3">
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">CPU Cores:</span>
                  <span class="col-span-2 text-white">{{ vm.vcpuCount || 0 }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Memory:</span>
                  <span class="col-span-2 text-white">{{ formatBytes((vm.memoryMB || 0) * 1024 * 1024) }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Disk Size:</span>
                  <span class="col-span-2 text-white">{{ vm.diskSizeGB || 0 }} GB</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">CPU Model:</span>
                  <span class="col-span-2 text-white">{{ vm.cpuModel || 'Default' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Boot Device:</span>
                  <span class="col-span-2 text-white">{{ vm.bootDevice || 'hd' }}</span>
                </div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <span class="text-slate-400 font-medium">Network:</span>
                  <span class="col-span-2 text-white">{{ vm.networkInterface || 'default' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- System Status -->
          <div class="mt-6 pt-6 border-t border-slate-700/50">
            <h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
              System Status
            </h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
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
                  <span class="text-slate-400 font-medium">Sync Status:</span>
                  <span :class="[
                    'px-2 py-1 rounded-full text-xs font-medium',
                    vm.syncStatus === 'SYNCED' ? 'bg-green-500/20 text-green-400' : 
                    vm.syncStatus === 'DRIFTED' ? 'bg-yellow-500/20 text-yellow-400' : 
                    'bg-gray-500/20 text-gray-400'
                  ]">
                    {{ vm.syncStatus }}
                  </span>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Task State:</span>
                  <span class="text-white">{{ vm.taskState || 'None' }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">OS Type:</span>
                  <span class="text-white">{{ vm.osType || 'Unknown' }}</span>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Host ID:</span>
                  <span class="text-white">{{ vm.hostId }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-400 font-medium">Graphics:</span>
                  <span class="text-white">{{ vm.graphics || 'None' }}</span>
                </div>
              </div>
            </div>
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

    <!-- Extended Hardware Configuration Modal -->
    <VMHardwareConfigModalExtended
      v-if="vm"
      :show="showExtendedHardwareModal"
      :host-id="props.hostId"
      :vm-name="vm.name"
      @close="showExtendedHardwareModal = false"
      @hardware-updated="loadVM"
    />

    <!-- Metrics Settings Modal (overlay) -->
    <MetricSettingsModal
      v-if="showMetricSettings"
      :show="showMetricSettings"
      @close="showMetricSettings = false"
      @applied="refreshStats"
    />


  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useVMStore } from '@/stores/vmStore';
import { useUIStore } from '@/stores/uiStore';
import { useSettingsStore } from '@/stores/settingsStore';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';
import FBackButton from '@/components/ui/FBackButton.vue';
import VMHardwareConfigModalExtended from '@/components/modals/VMHardwareConfigModalExtended.vue';
import MetricSettingsModal from '@/components/modals/MetricSettingsModal.vue';
import type { VirtualMachine, VMStats } from '@/types';
import { wsManager } from '@/services/api';
import { getConsoleRoute, getConsoleType, getConsoleDisplayName } from '@/utils/console';

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const route = useRoute();
const router = useRouter();

const vmStore = useVMStore();
const uiStore = useUIStore();

// Component state
const vm = ref<VirtualMachine | null>(null);
const vmStats = ref<VMStats | null>(null);
const error = ref<string | null>(null);
const loadingStats = ref(false);
const showExtendedHardwareModal = ref(false);
// simplified CPU display: show smoothed host-normalized `cpu_percent`
const showMetricSettings = ref(false);
const vmDetailsExpanded = ref(false);

// Console preview state
const consolePreviewIframe = ref<HTMLIFrameElement | null>(null);
const consoleConnected = ref(false);
const consoleRefreshKey = ref(0);

// Console preview source
const consolePreviewSrc = computed(() => {
  if (!vm.value || vm.value.state !== 'ACTIVE' || !getConsoleType(vm.value)) {
    return null;
  }

  const consoleType = getConsoleType(vm.value);
  const host = window.location.hostname;
  const port = window.location.port || (window.location.protocol === 'https:' ? '443' : '80');
  
  if (consoleType === 'spice') {
    const path = `api/v1/hosts/${props.hostId}/vms/${props.vmName}/spice`;
    const params = new URLSearchParams({
      host,
      port,
      path,
      autoconnect: '1',
      resize: 'scale',
      show_control: '0',
      _refresh: consoleRefreshKey.value.toString()
    });
    return `/spice/spice_responsive.html?${params.toString()}`;
  } else if (consoleType === 'vnc') {
    const path = `api/v1/hosts/${props.hostId}/vms/${props.vmName}/vnc`;
    const params = new URLSearchParams({
      host,
      port,
      path,
      autoconnect: 'true',
      resize: 'scale',
      _refresh: consoleRefreshKey.value.toString()
    });
    return `/vnc/vnc.html?${params.toString()}`;
  }
  
  return null;
});

// Console preview handlers
const onConsolePreviewLoad = () => {
  consoleConnected.value = true;
};

const refreshConsolePreview = () => {
  consoleConnected.value = false;
  consoleRefreshKey.value++;
};

// Auto-refresh console preview every 30 seconds when active
let consoleRefreshInterval: number | null = null;

const startConsoleRefresh = () => {
  if (consoleRefreshInterval) {
    clearInterval(consoleRefreshInterval);
  }
  consoleRefreshInterval = setInterval(() => {
    if (vm.value?.state === 'ACTIVE' && getConsoleType(vm.value)) {
      refreshConsolePreview();
    }
  }, 30000); // Refresh every 30 seconds
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
};

// Context actions for the back button
const vmContextActions = computed(() => [
  {
    label: 'Clone',
    action: () => {
      // TODO: Implement VM cloning
      console.log('Clone VM:', vm.value?.name);
    },
    icon: 'copy'
  },
  {
    label: 'Export',
    action: () => {
      // TODO: Implement VM export
      console.log('Export VM:', vm.value?.name);
    },
    icon: 'download'
  }
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

// Get VM data
const loadVM = async (): Promise<void> => {
  try {
    error.value = null;
    
    // Stop current monitoring if active
    stopStatsMonitoring();
    
    await vmStore.fetchVMs(props.hostId);
    
    // Find the VM by name
    const foundVM = vmStore.vmsByHost(props.hostId).find((v: VirtualMachine) => v.name === props.vmName);
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
      case 'sync':
        await vmStore.syncVM(props.hostId, vm.value.name);
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

// WebSocket-based stats monitoring
let isSubscribed = false;

const startStatsMonitoring = (): void => {
  if (vm.value?.state === 'ACTIVE' && !isSubscribed) {
    console.log(`Starting stats monitoring for VM: ${props.hostId}/${vm.value.name}`);
    
    // Always do an initial fetch first
    refreshStats();
    
    // Connect WebSocket for real-time updates
    wsManager.connect().then(() => {
      if (vm.value?.state === 'ACTIVE' && !isSubscribed) {
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
  } else if (vm.value?.state === 'ACTIVE') {
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
    startStatsMonitoring();
    startConsoleRefresh();
  });
});

onUnmounted(() => {
  stopStatsMonitoring();
  stopConsoleRefresh();
});
</script>