<template>
  <div class="space-y-8">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />

    <!-- Host Overview (when no host selected) -->
    <div v-if="!selectedHost" class="space-y-8">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-white mb-2">Host Management</h1>
        <p class="text-xl text-slate-400">Select a host to view and manage its virtual machines</p>
      </div>

      <!-- Host Overview Cards -->
      <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        <FCard
          v-for="host in hostsWithStats"
          :key="host.id"
          class="cursor-pointer transition-all duration-300 card-glow hover:scale-105"
          interactive
          @click="selectHost(host.id)"
        >
          <div class="p-6 space-y-4">
            <!-- Host Header -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-accent-500 to-accent-600 flex items-center justify-center shadow-lg">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                  </svg>
                </div>
                <div>
                  <h3 class="font-bold text-white">{{ getHostDisplayName(host) }}</h3>
                  <p class="text-sm text-slate-400">{{ extractHostname(host.uri) }}</p>
                </div>
              </div>
              
              <!-- Connection Status -->
              <div :class="[
                'px-3 py-1 rounded-full text-xs font-medium border',
                host.state === 'CONNECTED' ? 'bg-green-500/20 border-green-500/30 text-green-400' :
                host.isConnecting ? 'bg-yellow-500/20 border-yellow-500/30 text-yellow-400' :
                'bg-red-500/20 border-red-500/30 text-red-400'
              ]">
                {{ getHostStatusText(host) }}
              </div>
            </div>

            <!-- Performance Insights -->
            <div v-if="host.stats" class="grid grid-cols-2 gap-4 p-4 bg-white/5 rounded-lg border border-white/10">
              <div class="text-center">
                <div class="text-2xl font-bold text-green-400">{{ host.stats.vm_count || 0 }}</div>
                <div class="text-xs text-slate-400">Virtual Machines</div>
              </div>
              <div class="text-center">
                <div class="text-2xl font-bold text-blue-400">{{ Math.round(host.stats.cpu_percent || 0) }}%</div>
                <div class="text-xs text-slate-400">CPU Usage</div>
              </div>
            </div>

            <!-- Memory Usage Bar -->
            <div v-if="host.stats" class="space-y-2">
              <div class="flex justify-between text-sm">
                <span class="text-slate-400">Memory</span>
                <span class="text-slate-300">{{ formatBytes(host.stats.memory_total - host.stats.memory_available) }} / {{ formatBytes(host.stats.memory_total) }}</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-2">
                <div 
                  class="h-2 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                  :style="{ width: `${((host.stats.memory_total - host.stats.memory_available) / host.stats.memory_total) * 100}%` }"
                ></div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="flex gap-2 pt-2 border-t border-white/10 justify-center">
              <FButton
                v-if="host.state === 'DISCONNECTED'"
                size="sm"
                variant="primary"
                :disabled="loading.connectHost[host.id]"
                @click.stop="connectHost(host.id)"
                class="px-3"
                :title="loading.connectHost[host.id] ? 'Connecting...' : 'Connect to host'"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                </svg>
              </FButton>
              
              <FButton
                v-if="host.state === 'CONNECTED'"
                size="sm"
                variant="ghost"
                @click.stop="disconnectHost(host.id)"
                class="px-3"
                title="Disconnect from host"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </FButton>
              
              <FButton
                size="sm"
                variant="accent"
                @click.stop="selectHost(host.id)"
                class="flex-1"
              >
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                </svg>
                Manage
              </FButton>
            </div>
          </div>
        </FCard>
      </div>

      <!-- Add Host Button -->
      <div class="flex justify-center">
        <FButton
          variant="primary"
          glow
          @click="openAddHostModal"
          :disabled="loading.addHost"
        >
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          {{ loading.addHost ? 'Adding...' : 'Add New Host' }}
        </FButton>
      </div>
    </div>

    <!-- Host Header -->
    <div v-if="selectedHost" class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <FButton
          variant="ghost"
          size="sm"
          @click="selectedHostId = null"
          class="p-2"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
          </svg>
        </FButton>
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-accent-500 to-accent-600 flex items-center justify-center shadow-lg shadow-accent-500/25">
          <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
          </svg>
        </div>
        <div>
          <h1 class="text-3xl font-bold text-white">{{ getHostDisplayName(selectedHost) }}</h1>
          <p class="text-slate-400">{{ selectedHost.uri }}</p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <div :class="[
          'px-4 py-2 rounded-lg border text-center',
          selectedHost.state === 'CONNECTED' ? 'bg-green-500/10 border-green-500/30 text-green-400' :
          'bg-red-500/10 border-red-500/30 text-red-400'
        ]">
          <span class="font-medium">{{ getHostStatusText(selectedHost) }}</span>
        </div>
        <FButton
          variant="ghost"
          @click="openHostModal(selectedHost)"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          </svg>
          Settings
        </FButton>
      </div>
    </div>

    <!-- Host Information Cards -->  
    <div v-if="selectedHost" class="space-y-6">
      <!-- System Information - Wide Card -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-4 mb-6">
            <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-2xl font-bold text-white">System Information</h3>
              <p class="text-slate-400">Host configuration and connection details</p>
            </div>
          </div>
          
          <!-- System Info Grid -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <div class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Host ID</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <span class="text-white font-mono text-sm" :title="selectedHost.id">{{ selectedHost.id.substring(0, 8) }}...{{ selectedHost.id.substring(-8) }}</span>
              </div>
            </div>
            
            <div class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Hostname</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <span class="text-white font-medium" :title="selectedHost.stats?.host_info?.hostname || extractHostname(selectedHost.uri)">{{ selectedHost.stats?.host_info?.hostname || extractHostname(selectedHost.uri) }}</span>
              </div>
            </div>
            
            <div class="space-y-2 sm:col-span-2 lg:col-span-1">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Connection URI</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <span class="text-slate-300 text-sm break-all font-mono" :title="selectedHost.uri">{{ selectedHost.uri }}</span>
              </div>
            </div>
            
            <div v-if="selectedHost.stats?.host_info" class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">CPU Cores</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <div class="flex items-center gap-2">
                  <div class="w-2 h-2 bg-blue-400 rounded-full"></div>
                  <span class="text-blue-400 text-lg font-bold">{{ selectedHost.stats.host_info.cpu || selectedHost.stats.resources?.cpu_count || 'N/A' }}</span>
                  <span class="text-slate-400 text-sm">cores</span>
                </div>
              </div>
            </div>
            
            <div v-if="selectedHost.stats?.host_info" class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Total Memory</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <div class="flex items-center gap-2">
                  <div class="w-2 h-2 bg-purple-400 rounded-full"></div>
                  <span class="text-purple-400 text-lg font-bold">{{ formatBytes(selectedHost.stats.host_info.memory || selectedHost.stats.resources?.memory_bytes || 0) }}</span>
                </div>
              </div>
            </div>
            
            <div v-if="selectedHost.stats && selectedHost.stats.uptime !== undefined" class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Uptime</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <div class="flex items-center gap-2">
                  <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                  <span class="text-green-400 text-lg font-bold">{{ formatUptime(selectedHost.stats.uptime) }}</span>
                </div>
              </div>
            </div>
            
            <div v-if="selectedHost.stats?.host_info?.version" class="space-y-2">
              <span class="text-xs text-slate-500 uppercase tracking-wide font-medium">Hypervisor</span>
              <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/50">
                <div class="flex items-center gap-2">
                  <div class="w-2 h-2 bg-cyan-400 rounded-full"></div>
                  <span class="text-cyan-400 font-medium" :title="selectedHost.stats.host_info.version">{{ selectedHost.stats.host_info.version }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </FCard>
      
      <!-- Performance Metrics - Compact Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">

        <!-- CPU Usage -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 002 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">CPU</h4>
            </div>
            
            <div class="text-center mb-3">
              <div v-if="selectedHost.stats && selectedHost.stats.cpu_percent !== undefined" class="text-3xl font-bold mb-1" :class="getCPUUsageColor(selectedHost.stats.cpu_percent)">{{ Math.round(selectedHost.stats.cpu_percent) }}%</div>
              <div v-else-if="loading.hostStats[selectedHost.id]" class="text-3xl font-bold text-slate-500 mb-1">--</div>
              <div v-else class="text-3xl font-bold text-slate-500 mb-1">--</div>
            </div>
            
            <div v-if="selectedHost.stats" class="w-full bg-slate-700/50 rounded-full h-2 mb-3">
              <div 
                class="h-2 rounded-full transition-all duration-500 ease-out"
                :class="getCPUUsageGradient(selectedHost.stats.cpu_percent || 0)"
                :style="{ width: `${Math.round(selectedHost.stats.cpu_percent || 0)}%` }"
              ></div>
            </div>
            
            <div class="text-xs text-slate-400 text-center">
              {{ selectedHost.stats?.host_info?.cpu || selectedHost.stats?.resources?.cpu_count || 'N/A' }} cores available
            </div>
          </div>
        </FCard>

        <!-- Memory Usage -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Memory</h4>
            </div>
            
            <div v-if="selectedHost.stats && selectedHost.stats.memory_total !== undefined">
              <div class="text-center mb-3">
                <div class="text-3xl font-bold mb-1" :class="getMemoryUsageColor(((selectedHost.stats.memory_total - selectedHost.stats.memory_available) / selectedHost.stats.memory_total) * 100)">
                  {{ Math.round(((selectedHost.stats.memory_total - selectedHost.stats.memory_available) / selectedHost.stats.memory_total) * 100) }}%
                </div>
              </div>
              
              <div class="w-full bg-slate-700/50 rounded-full h-2 mb-3">
                <div 
                  class="h-2 rounded-full transition-all duration-500 ease-out"
                  :class="getMemoryUsageGradient(((selectedHost.stats.memory_total - selectedHost.stats.memory_available) / selectedHost.stats.memory_total) * 100)"
                  :style="{ width: `${((selectedHost.stats.memory_total - selectedHost.stats.memory_available) / selectedHost.stats.memory_total) * 100}%` }"
                ></div>
              </div>
              
              <div class="text-xs text-slate-400 text-center">
                {{ formatBytes(selectedHost.stats.memory_total - selectedHost.stats.memory_available) }} / {{ formatBytes(selectedHost.stats.memory_total) }}
              </div>
            </div>
            
            <div v-else-if="loading.hostStats[selectedHost.id]" class="text-center">
              <div class="text-3xl font-bold text-slate-500 mb-1">--</div>
              <div class="text-xs text-slate-400">Loading...</div>
            </div>
            
            <div v-else class="text-center">
              <div class="text-3xl font-bold text-slate-500 mb-1">--</div>
              <div class="text-xs text-slate-400">Unavailable</div>
            </div>
          </div>
        </FCard>

        <!-- Storage Usage -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7h16l2 2v9a2 2 0 01-2 2H4a2 2 0 01-2-2V9l2-2zM14 2H6a2 2 0 00-2 2v5m16 0V4a2 2 0 00-2-2h-8"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">Storage</h4>
            </div>
            
            <div v-if="selectedHost.stats && selectedHost.stats.disk_total > 0">
              <div class="text-center mb-3">
                <div class="text-3xl font-bold mb-1" :class="getStorageUsageColor(((selectedHost.stats.disk_total - selectedHost.stats.disk_free) / selectedHost.stats.disk_total) * 100)">
                  {{ Math.round(((selectedHost.stats.disk_total - selectedHost.stats.disk_free) / selectedHost.stats.disk_total) * 100) }}%
                </div>
              </div>
              
              <div class="w-full bg-slate-700/50 rounded-full h-2 mb-3">
                <div 
                  class="h-2 rounded-full transition-all duration-500 ease-out"
                  :class="getStorageUsageGradient(((selectedHost.stats.disk_total - selectedHost.stats.disk_free) / selectedHost.stats.disk_total) * 100)"
                  :style="{ width: `${((selectedHost.stats.disk_total - selectedHost.stats.disk_free) / selectedHost.stats.disk_total) * 100}%` }"
                ></div>
              </div>
              
              <div class="text-xs text-slate-400 text-center">
                {{ formatBytes(selectedHost.stats.disk_total - selectedHost.stats.disk_free) }} / {{ formatBytes(selectedHost.stats.disk_total) }}
              </div>
            </div>
            
            <div v-else class="text-center">
              <div class="text-3xl font-bold text-slate-500 mb-1">--</div>
              <div class="text-xs text-slate-400">Unavailable</div>
            </div>
          </div>
        </FCard>

        <!-- VM Summary -->
        <FCard class="card-glow">
          <div class="p-4">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              <h4 class="text-lg font-bold text-white">VMs</h4>
            </div>
            
            <div class="grid grid-cols-2 gap-4 mb-3">
              <div class="text-center">
                <div class="text-2xl font-bold text-green-400">{{ managedVMs.length }}</div>
                <div class="text-xs text-slate-400">Managed</div>
              </div>
              <div class="text-center">
                <div class="text-2xl font-bold text-blue-400">{{ discoveredVMs.length }}</div>
                <div class="text-xs text-slate-400">Discovered</div>
              </div>
            </div>
            
            <div class="text-xs text-slate-400 text-center">
              Total: {{ managedVMs.length + discoveredVMs.length }} virtual machines
            </div>
          </div>
        </FCard>
      </div>
    </div>

    <!-- Virtual Machine Management -->
    <div v-if="selectedHost" class="space-y-6">
      <!-- Host Actions -->
      <div class="flex justify-between items-center">
        <h2 class="text-2xl font-bold text-white">Virtual Machines</h2>
        <div class="flex gap-3">
          <FButton
            v-if="selectedHost.state === 'CONNECTED'"
            variant="accent"
            @click="importAllVMs(selectedHost.id)"
            :disabled="loading.hostImportAll === selectedHost.id"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"/>
            </svg>
            {{ loading.hostImportAll === selectedHost.id ? 'Importing...' : 'Import All VMs' }}
          </FButton>
          <FButton
            variant="ghost"
            @click="refreshDiscoveredVMs(selectedHost.id)"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            Refresh
          </FButton>
        </div>
      </div>

      <!-- Managed VMs Card -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6 cursor-pointer" @click="managedVMsCollapsed = !managedVMsCollapsed">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center shadow-lg">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Managed Virtual Machines</h3>
                <p class="text-slate-400">{{ managedVMs.length }} VMs under management</p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="px-4 py-2 bg-green-500/20 border border-green-500/30 rounded-lg">
                <span class="text-green-400 font-semibold">{{ managedVMs.length }}</span>
              </div>
              <svg 
                class="w-5 h-5 text-slate-400 transition-transform duration-200" 
                :class="{ 'rotate-180': !managedVMsCollapsed }"
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </div>
          </div>

          <div v-show="!managedVMsCollapsed" class="space-y-4">
            <div class="flex justify-between items-center">
              <div class="text-sm text-slate-400">Active virtual machines managed by this host</div>
              <FButton
                v-if="selectedHost && selectedHost.state === 'CONNECTED'"
                variant="primary"
                size="sm"
                @click="openCreateVMModal"
              >
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Create VM
              </FButton>
            </div>
            
            <div v-if="managedVMs.length === 0" class="text-center py-12 border-2 border-dashed border-slate-600 rounded-xl">
              <div class="w-16 h-16 mx-auto mb-4 bg-slate-700 rounded-full flex items-center justify-center">
                <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              <h4 class="text-lg font-semibold text-white mb-2">No managed VMs</h4>
              <p class="text-slate-400 mb-4">Create your first virtual machine to get started</p>
              <FButton
                v-if="selectedHost && selectedHost.state === 'CONNECTED'"
                variant="primary"
                @click="openCreateVMModal"
              >
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Create Virtual Machine
              </FButton>
            </div>
            
            <div v-else class="grid gap-4">
              <VMCard
                v-for="vm in managedVMs"
                :key="vm.uuid"
                :vm="vm"
                :host-id="selectedHost.id"
                @action="handleVMAction"
              />
            </div>
          </div>
        </div>
      </FCard>

      <!-- Discovered VMs Card -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6 cursor-pointer" @click="discoveredVMsCollapsed = !discoveredVMsCollapsed">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Discovered Virtual Machines</h3>
                <p class="text-slate-400">{{ discoveredVMs.length }} VMs found on host</p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="px-4 py-2 bg-blue-500/20 border border-blue-500/30 rounded-lg">
                <span class="text-blue-400 font-semibold">{{ discoveredVMs.length }}</span>
              </div>
              <svg 
                class="w-5 h-5 text-slate-400 transition-transform duration-200" 
                :class="{ 'rotate-180': !discoveredVMsCollapsed }"
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </div>
          </div>

          <div v-show="!discoveredVMsCollapsed" class="space-y-4">
            <div class="text-sm text-slate-400">Virtual machines detected on the host that can be imported</div>
            
            <div v-if="discoveredVMs.length === 0" class="text-center py-12 border-2 border-dashed border-slate-600 rounded-xl">
              <div class="w-16 h-16 mx-auto mb-4 bg-slate-700 rounded-full flex items-center justify-center">
                <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
              </div>
              <h4 class="text-lg font-semibold text-white mb-2">No discovered VMs</h4>
              <p class="text-slate-400">No unmanaged virtual machines found on this host</p>
            </div>
            
            <div v-else>
              <DiscoveredVMBulkManager
                :vms="[...discoveredVMs]"
                :host-id="selectedHost.id"
                :importing="!!loading.hostImportAll"
                :deleting="false"
                @bulk-import="handleBulkImport"
                @bulk-delete="handleBulkDelete"
                @single-import="importVM"
              />
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Add Host Modal -->
    <AddHostModal
      :open="modals.addHost"
      @close="closeAddHostModal"
      @submit="handleAddHost"
    />

    <!-- Create VM Modal -->
    <CreateVMModal
      v-if="selectedHost"
      :open="modals.createVM"
      :host-id="selectedHost.id"
      @close="closeCreateVMModal"
      @vm-created="handleVMCreated"
    />

    <!-- Host Settings Modal -->
    <HostSettingsModal
      :open="modals.hostSettings"
      :host-id="selectedHostForSettings?.id"
      @close="closeHostModal"
      @host-updated="handleHostUpdated"
      @host-deleted="handleHostDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useHostStore, useVMStore, useUIStore } from '@/stores';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue';
import VMCard from '@/components/vm/VMCard.vue';
import DiscoveredVMCard from '@/components/vm/DiscoveredVMCard.vue';
import DiscoveredVMBulkManager from '@/components/vm/DiscoveredVMBulkManager.vue';
import AddHostModal from '@/components/modals/AddHostModal.vue';
import CreateVMModal from '@/components/modals/CreateVMModal.vue';
import HostSettingsModal from '@/components/modals/HostSettingsModal.vue';
import type { Host, VirtualMachine, DiscoveredVM } from '@/types';

// Store instances
const hostStore = useHostStore();
const vmStore = useVMStore();
const uiStore = useUIStore();

// Local state
const selectedHostId = ref<string | null>(null);
const selectedHostForSettings = ref<Host | null>(null);
const managedVMsCollapsed = ref(false);
const discoveredVMsCollapsed = ref(false);

const modals = ref({
  addHost: false,
  createVM: false,
  hostSettings: false
});

// Computed properties
const hostsWithStats = computed(() => hostStore.hostsWithStats);
const loading = computed(() => hostStore.loading);
const selectedHost = computed(() => {
  if (!selectedHostId.value) return null;
  const hosts = hostStore.hostsWithStats;
  const host = hosts.find(host => host.id === selectedHostId.value) || null;
  console.log('Selected host:', host);
  console.log('Selected host stats:', host?.stats);
  return host;
});

const managedVMs = computed(() => 
  selectedHostId.value ? vmStore.vmsByHost(selectedHostId.value) : []
);

const discoveredVMs = computed(() => 
  selectedHostId.value ? hostStore.discoveredVMs[selectedHostId.value] || [] : []
);

// Actions
const selectHost = (hostId: string): void => {
  selectedHostId.value = hostId;
  hostStore.selectHost(hostId);
  
  // Always try to fetch host stats, even if not connected
  hostStore.fetchHostStats(hostId);
  
  // Fetch VMs for this host
  if (hostStore.getHostById(hostId)?.state === 'CONNECTED') {
    vmStore.fetchVMs(hostId);
    hostStore.refreshDiscoveredVMs(hostId);
  }
};

const connectHost = async (hostId: string): Promise<void> => {
  try {
    await hostStore.connectHost(hostId);
    uiStore.addToast('Host connection initiated', 'success');
  } catch (error) {
    uiStore.addToast('Failed to connect to host', 'error');
  }
};

const disconnectHost = async (hostId: string): Promise<void> => {
  try {
    await hostStore.disconnectHost(hostId);
    uiStore.addToast('Host disconnected', 'success');
  } catch (error) {
    uiStore.addToast('Failed to disconnect host', 'error');
  }
};

const refreshHostData = async (hostId: string): Promise<void> => {
  try {
    await Promise.all([
      hostStore.fetchHostStats(hostId),
      vmStore.fetchVMs(hostId),
      hostStore.refreshDiscoveredVMs(hostId)
    ]);
    uiStore.addToast('Host data refreshed', 'success');
  } catch (error) {
    uiStore.addToast('Failed to refresh host data', 'error');
  }
};

const refreshDiscoveredVMs = async (hostId: string): Promise<void> => {
  try {
    await hostStore.refreshDiscoveredVMs(hostId);
    uiStore.addToast('Discovered VMs refreshed', 'success');
  } catch (error) {
    uiStore.addToast('Failed to refresh discovered VMs', 'error');
  }
};

const importAllVMs = async (hostId: string): Promise<void> => {
  try {
    await hostStore.importAllVMs(hostId);
    uiStore.addToast('All VMs imported successfully', 'success');
    // Refresh managed VMs
    await vmStore.fetchVMs(hostId);
  } catch (error) {
    uiStore.addToast('Failed to import VMs', 'error');
  }
};

const importVM = async (hostId: string, vmName: string): Promise<void> => {
  try {
    await vmStore.importVM(hostId, vmName);
    uiStore.addToast(`VM ${vmName} imported successfully`, 'success');
    // Refresh both lists
    await Promise.all([
      vmStore.fetchVMs(hostId),
      hostStore.refreshDiscoveredVMs(hostId)
    ]);
  } catch (error) {
    uiStore.addToast(`Failed to import VM ${vmName}`, 'error');
  }
};

const handleBulkImport = async (domainUUIDs: string[]): Promise<void> => {
  if (!selectedHost.value) return;
  
  try {
    await hostStore.importSelectedVMs(selectedHost.value.id, domainUUIDs);
    uiStore.addToast(`${domainUUIDs.length} VMs imported successfully`, 'success');
    // Refresh both lists
    await Promise.all([
      vmStore.fetchVMs(selectedHost.value.id),
      hostStore.refreshDiscoveredVMs(selectedHost.value.id)
    ]);
  } catch (error) {
    uiStore.addToast(`Failed to import selected VMs`, 'error');
  }
};

const handleBulkDelete = async (domainUUIDs: string[]): Promise<void> => {
  if (!selectedHost.value) return;
  
  try {
    await hostStore.deleteSelectedDiscoveredVMs(selectedHost.value.id, domainUUIDs);
    uiStore.addToast(`${domainUUIDs.length} discovered VMs removed`, 'success');
    // Refresh discovered VMs list
    await hostStore.refreshDiscoveredVMs(selectedHost.value.id);
  } catch (error) {
    uiStore.addToast(`Failed to remove selected VMs`, 'error');
  }
};

const handleVMAction = async (action: string, hostId: string, vmName: string): Promise<void> => {
  try {
    switch (action) {
      case 'start':
        await vmStore.startVM(hostId, vmName);
        break;
      case 'shutdown':
        await vmStore.stopVM(hostId, vmName);
        break;
      case 'reboot':
        await vmStore.restartVM(hostId, vmName);
        break;
      case 'forceOff':
        await vmStore.forceOffVM(hostId, vmName);
        break;
      case 'forceReset':
        await vmStore.forceResetVM(hostId, vmName);
        break;
      case 'sync':
        await vmStore.syncVM(hostId, vmName);
        break;
      case 'rebuild':
        await vmStore.rebuildVM(hostId, vmName);
        break;
      default:
        throw new Error(`Unknown VM action: ${action}`);
    }
    uiStore.addToast(`VM ${action} initiated`, 'success');
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : `Failed to ${action} VM`;
    uiStore.addToast(errorMessage, 'error');
    console.error('VM action error:', error);
  }
};

// Modal handlers
const openAddHostModal = (): void => {
  modals.value.addHost = true;
};

const closeAddHostModal = (): void => {
  modals.value.addHost = false;
};

const openCreateVMModal = (): void => {
  modals.value.createVM = true;
};

const closeCreateVMModal = (): void => {
  modals.value.createVM = false;
};

const openHostModal = (host: Host): void => {
  selectedHostForSettings.value = host;
  modals.value.hostSettings = true;
};

const closeHostModal = (): void => {
  modals.value.hostSettings = false;
  selectedHostForSettings.value = null;
};

const handleAddHost = async (hostData: Omit<Host, 'id'>): Promise<void> => {
  try {
    await hostStore.addHost(hostData);
    uiStore.addToast('Host added successfully', 'success');
    closeAddHostModal();
  } catch (error) {
    uiStore.addToast('Failed to add host', 'error');
  }
};

const handleVMCreated = async (vm: VirtualMachine): Promise<void> => {
  try {
    uiStore.addToast(`VM "${vm.name}" created successfully`, 'success');
    closeCreateVMModal();
    // Refresh the VM list for the current host
    if (selectedHost.value) {
      await vmStore.fetchVMs(selectedHost.value.id);
    }
  } catch (error) {
    uiStore.addToast('Failed to update VM list', 'error');
  }
};

const handleHostUpdated = async (host: Host): Promise<void> => {
  try {
    uiStore.addToast('Host settings updated successfully', 'success');
    closeHostModal();
    // Refresh host data
    await hostStore.fetchHosts();
  } catch (error) {
    uiStore.addToast('Failed to refresh host data', 'error');
  }
};

const handleHostDeleted = async (hostId: string): Promise<void> => {
  try {
    uiStore.addToast('Host removed successfully', 'success');
    closeHostModal();
    // Refresh host data and clear selection if needed
    await hostStore.fetchHosts();
    if (selectedHostId.value === hostId) {
      selectedHostId.value = null;
    }
  } catch (error) {
    uiStore.addToast('Failed to refresh host data', 'error');
  }
};

// Utility functions
const getHostDisplayName = (host: Host): string => {
  return host.id || 'Unknown Host';
};

const getHostStatusColor = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-400';
    case 'DISCONNECTED': return 'bg-red-400';
    case 'ERROR': return 'bg-red-500';
    default: return 'bg-yellow-400';
  }
};

const getHostGlowColor = (host: any): 'primary' | 'accent' | 'neon-blue' | 'neon-cyan' | 'neon-purple' => {
  switch (host.state) {
    case 'CONNECTED': return 'accent';
    case 'ERROR': return 'neon-purple';
    default: return 'primary';
  }
};

const getHostStatusBadgeClass = (host: Host): string => {
  switch (host.state) {
    case 'CONNECTED': return 'bg-green-500/20 text-green-400';
    case 'DISCONNECTED': return 'bg-red-500/20 text-red-400';
    case 'ERROR': return 'bg-red-500/20 text-red-400';
    default: return 'bg-yellow-500/20 text-yellow-400';
  }
};

const getHostStatusText = (host: any): string => {
  if (host.isConnecting) return 'Connecting...';
  return host.state?.toLowerCase() || 'unknown';
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

const formatUptime = (seconds: number): string => {
  // Handle invalid input
  if (typeof seconds !== 'number' || isNaN(seconds) || seconds < 0) {
    return 'N/A';
  }
  
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m`;
  return `${Math.floor(seconds)}s`;
};

const extractHostname = (uri: string): string => {
  try {
    // Handle different URI formats
    if (uri.includes('://')) {
      const url = new URL(uri);
      return url.hostname || url.host || 'localhost';
    }
    // Handle simple host:port format
    if (uri.includes(':')) {
      return uri.split(':')[0] || 'localhost';
    }
    return uri;
  } catch (error) {
    // Fallback for malformed URIs
    return uri.split('@').pop()?.split('/')[0] || 'unknown';
  }
};

const getCPUUsageColor = (percentage: number): string => {
  if (percentage < 50) return 'text-green-400';
  if (percentage < 75) return 'text-yellow-400';
  if (percentage < 90) return 'text-orange-400';
  return 'text-red-400';
};

const getCPUUsageGradient = (percentage: number): string => {
  if (percentage < 50) return 'bg-gradient-to-r from-green-500 to-emerald-500';
  if (percentage < 75) return 'bg-gradient-to-r from-yellow-500 to-orange-500';
  if (percentage < 90) return 'bg-gradient-to-r from-orange-500 to-red-500';
  return 'bg-gradient-to-r from-red-500 to-red-600';
};

const getMemoryUsageColor = (percentage: number): string => {
  if (percentage < 60) return 'text-purple-400';
  if (percentage < 80) return 'text-yellow-400';
  if (percentage < 95) return 'text-orange-400';
  return 'text-red-400';
};

const getMemoryUsageGradient = (percentage: number): string => {
  if (percentage < 60) return 'bg-gradient-to-r from-purple-500 to-pink-500';
  if (percentage < 80) return 'bg-gradient-to-r from-yellow-500 to-orange-500';
  if (percentage < 95) return 'bg-gradient-to-r from-orange-500 to-red-500';
  return 'bg-gradient-to-r from-red-500 to-red-600';
};

const getStorageUsageColor = (percentage: number): string => {
  if (percentage < 70) return 'text-blue-400';
  if (percentage < 85) return 'text-yellow-400';
  if (percentage < 95) return 'text-orange-400';
  return 'text-red-400';
};

const getStorageUsageGradient = (percentage: number): string => {
  if (percentage < 70) return 'bg-gradient-to-r from-blue-500 to-cyan-500';
  if (percentage < 85) return 'bg-gradient-to-r from-yellow-500 to-orange-500';
  if (percentage < 95) return 'bg-gradient-to-r from-orange-500 to-red-500';
  return 'bg-gradient-to-r from-red-500 to-red-600';
};

// Lifecycle
onMounted(async () => {
  // Load initial data
  await hostStore.fetchHosts();
  
  // Select first connected host by default
  const connectedHost = hostStore.connectedHosts[0];
  if (connectedHost) {
    selectHost(connectedHost.id);
  }
});

onUnmounted(() => {
  // Cleanup if needed
});
</script>