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
    </div>

    <!-- VM Control Panel -->
    <FCard v-if="vm" class="p-6 card-glow">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-white">VM Controls</h2>
        <div v-if="vm.taskState" class="animate-pulse">
          <span class="px-2 py-1 rounded-full text-xs font-medium bg-yellow-500/20 text-yellow-400">
            {{ vm.taskState }}
          </span>
        </div>
      </div>
      
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <!-- Power Controls -->
        <FButton
          v-if="vm.state === 'STOPPED'"
          variant="primary"
          @click="handleVMAction('start')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
          </svg>
          Start
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('shutdown')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clip-rule="evenodd" />
          </svg>
          Shutdown
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('reboot')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          Reboot
        </FButton>
        
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="openConsole"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
          </svg>
          Console
        </FButton>
      </div>

      <!-- Hardware Configuration Button -->
      <div class="mt-4 pt-4 border-t border-gray-700">
        <FButton
          variant="outline"
          @click="showExtendedHardwareModal = true"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          </svg>
          Extended Hardware Configuration
        </FButton>
      </div>
    </FCard>

    <!-- VM Information - Wide Card -->
    <FCard v-if="vm" class="card-glow">
      <div class="p-6">
        <div class="flex items-center gap-4 mb-6">
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
      </div>
    </FCard>

    <!-- Performance Metrics - Compact Cards -->
    <div v-if="vm && vm.state === 'ACTIVE'" class="space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-xl font-bold text-white">Performance Metrics</h3>
        <div class="flex items-center gap-3">
          <FButton variant="outline" size="sm" @click="showMetricSettings = true">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            </svg>
            Metrics
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
            <div class="text-center">
              <div class="text-lg font-bold text-green-400">{{ formatDisk((vmStats.disk_read_kib_per_sec || 0)) }}</div>
              <div class="text-xs text-slate-400 mt-1">R: {{ (vmStats.disk_read_kib_per_sec || 0).toFixed(1) }} KiB/s</div>
              <div class="text-xs text-slate-400">W: {{ (vmStats.disk_write_kib_per_sec || 0).toFixed(1) }} KiB/s</div>
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
            <div class="text-center">
              <div class="text-lg font-bold text-cyan-400">{{ formatNetwork((vmStats.network_rx_mbps || 0)) }}</div>
              <div class="text-xs text-slate-400 mt-1">RX: {{ (vmStats.network_rx_mbps || vmStats.network_rx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</div>
              <div class="text-xs text-slate-400">TX: {{ (vmStats.network_tx_mbps || vmStats.network_tx_mb || 0).toFixed(2) }} {{ settings.units.network === 'kb' ? 'KB/s' : 'MB/s' }}</div>
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

    <!-- Advanced Actions -->
    <FCard v-if="vm" class="p-6 card-glow">
      <h3 class="text-lg font-semibold text-white mb-4">Advanced Actions</h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <FButton
          variant="ghost"
          @click="handleVMAction('sync')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          Sync from Libvirt
        </FButton>
        <FButton
          variant="ghost"
          @click="handleVMAction('rebuild')"
          :disabled="!!vm.taskState"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          Rebuild from DB
        </FButton>
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('forceOff')"
          :disabled="!!vm.taskState"
          class="text-orange-400 hover:bg-orange-500/10 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
          Force Off
        </FButton>
        <FButton
          v-if="vm.state === 'ACTIVE'"
          variant="ghost"
          @click="handleVMAction('forceReset')"
          :disabled="!!vm.taskState"
          class="text-red-400 hover:bg-red-500/10 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
          Force Reset
        </FButton>
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
import { ref, computed, onMounted, onUnmounted } from 'vue';
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
import { getConsoleRoute } from '@/utils/console';

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
    console.log(`Subscribing to VM stats: ${props.hostId}/${vm.value.name}`);
    
    // Connect WebSocket if not connected
    wsManager.connect().then(() => {
      // Subscribe to stats updates
      wsManager.subscribeToVMStats(props.hostId, vm.value!.name);
      isSubscribed = true;
      
      // Listen for stats updates
      wsManager.on('vm-stats-updated', handleStatsUpdate);
      
      // Also do an initial fetch
      refreshStats();
    }).catch(error => {
      console.error('Failed to connect WebSocket:', error);
      // Fallback to initial fetch only
      refreshStats();
    });
  }
};

const stopStatsMonitoring = (): void => {
  if (isSubscribed && vm.value) {
    console.log(`Unsubscribing from VM stats: ${props.hostId}/${vm.value.name}`);
    wsManager.unsubscribeFromVMStats(props.hostId, vm.value.name);
    wsManager.off('vm-stats-updated', handleStatsUpdate);
    isSubscribed = false;
  }
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
  });
});

onUnmounted(() => {
  stopStatsMonitoring();
});
</script>