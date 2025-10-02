<template>
  <div class="space-y-8">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />

    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-4">
        <FButton variant="ghost" @click="router.push('/hosts')" class="p-2">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
          </svg>
        </FButton>
        <div>
          <h1 class="text-3xl font-bold text-white">
            {{ host?.name || 'Unknown Host' }}
          </h1>
          <p class="text-slate-400 text-sm font-mono">UUID: {{ host?.id }}</p>
        </div>
      </div>
      
      <div class="flex items-center gap-3">
        <!-- Connection Status -->
        <div class="flex items-center gap-2 px-3 py-1 rounded-full" :class="[
          host?.state === 'CONNECTED' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
        ]">
          <div class="w-2 h-2 rounded-full" :class="[
            host?.state === 'CONNECTED' ? 'bg-green-400 animate-pulse' : 'bg-red-400'
          ]"></div>
          <span class="text-xs font-medium">{{ host?.state || 'UNKNOWN' }}</span>
        </div>
        
        <!-- Refresh Button -->
        <FButton variant="ghost" size="sm" @click="refreshHost" :disabled="isRefreshing" class="p-3">
          <svg class="w-6 h-6" :class="{ 'animate-spin': isRefreshing }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
        </FButton>
        
        <!-- Connect/Disconnect Button -->
        <FButton 
          v-if="host?.state !== 'CONNECTED'"
          variant="primary" 
          size="sm" 
          @click="showConnectConfirmation = true"
          class="p-3"
          title="Connect Host"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
          </svg>
        </FButton>
        <FButton 
          v-else
          variant="ghost" 
          size="sm" 
          @click="showDisconnectConfirmation = true"
          class="p-3"
          title="Disconnect Host"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636L5.636 18.364m0-12.728L18.364 18.364"/>
          </svg>
        </FButton>
        
        <!-- Settings Button -->
        <FButton variant="ghost" size="sm" @click="showHostSettings = true" class="p-3">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
          </svg>
        </FButton>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- vCPU Usage -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg shadow-purple-500/25">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-1">{{ Math.round(hostStats.vcpuUsage || 0) }}%</div>
            <div class="text-sm text-slate-400">vCPU Usage</div>
            <div class="text-xs text-slate-500 mt-1">{{ hostStats.allocatedVcpus || 0 }}/{{ hostStats.cpuCores }} vCPUs</div>
          </div>
        </div>
      </FCard>

      <!-- Memory Usage -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg shadow-blue-500/25">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
              </svg>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-1">{{ Math.round(hostStats.memoryUsage) }}%</div>
            <div class="text-sm text-slate-400">Memory Usage</div>
            <div class="text-xs text-slate-500 mt-1">{{ formatBytes(hostStats.memoryTotal) }}</div>
          </div>
        </div>
      </FCard>

      <!-- Storage Usage -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-green-500 flex items-center justify-center shadow-lg shadow-emerald-500/25">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"/>
              </svg>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-1">{{ Math.round(hostStats.storageUsage) }}%</div>
            <div class="text-sm text-slate-400">Storage Usage</div>
            <div class="text-xs text-slate-500 mt-1">{{ formatBytes(hostStats.storageTotal) }}</div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Detailed Resource Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- System Resources -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
            <h3 class="text-xl font-bold text-white">System Resources</h3>
          </div>

          <!-- Show message if host not connected or no stats available -->
          <div v-if="!host || host.state !== 'CONNECTED' || !hasStatsData" class="text-center py-8">
            <svg class="w-16 h-16 mx-auto mb-4 text-slate-600" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
            </svg>
            <h4 class="text-white font-semibold mb-2">
              {{ !host ? 'Host Not Found' : host.state !== 'CONNECTED' ? 'Host Not Connected' : 'No Resource Data Available' }}
            </h4>
            <p class="text-slate-400 text-sm">
              {{ !host 
                ? 'The requested host could not be found.' 
                : host.state !== 'CONNECTED' 
                  ? `Connect the host to view system resource information. (Current state: ${host.state})` 
                  : `Resource statistics are not available for this host. (Stats: ${JSON.stringify(hostStore.hostStats[hostId])})` }}
            </p>
            <FButton 
              v-if="host && host.state !== 'CONNECTED'" 
              variant="primary" 
              class="mt-4" 
                            @click="showConnectConfirmation = true"
            >
              Connect Host
            </FButton>
          </div>

          <!-- Show resource data if available -->
          <div v-else class="space-y-6">
            <!-- CPU Details -->
            <div>
              <div class="flex justify-between items-center mb-3">
                <span class="text-sm font-medium text-white">CPU Usage</span>
                <span class="text-sm font-medium text-white">{{ Math.round(hostStats.cpuUsage || 0) }}%</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-3 bg-gradient-to-r from-purple-500 to-purple-600 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.cpuUsage}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-500 mt-2">
                <span>{{ hostStats.cpuCores }} CPU cores available</span>
                <span>{{ hostCapabilities?.host_info?.mhz || 'N/A' }} MHz</span>
              </div>
            </div>

            <!-- Memory Details -->
            <div>
              <div class="flex justify-between items-center mb-3">
                <span class="text-sm font-medium text-white">Memory Usage</span>
                <span class="text-sm font-medium text-white">{{ Math.round(hostStats.memoryUsage || 0) }}%</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-3 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.memoryUsage}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-500 mt-2">
                <span>{{ formatBytes(hostStats.memoryTotal - (hostStats.memoryTotal * hostStats.memoryUsage / 100)) }} used</span>
                <span>{{ formatBytes(hostStats.memoryTotal) }} total</span>
              </div>
            </div>

            <!-- Storage Details -->
            <div>
              <div class="flex justify-between items-center mb-3">
                <span class="text-sm font-medium text-white">Storage Usage</span>
                <span class="text-sm font-medium text-white">{{ Math.round(hostStats.storageUsage || 0) }}%</span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-3 bg-gradient-to-r from-emerald-500 to-green-500 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.storageUsage}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-500 mt-2">
                <span>{{ formatBytes(hostStats.storageTotal - (hostStats.storageTotal * hostStats.storageUsage / 100)) }} used</span>
                <span>{{ formatBytes(hostStats.storageTotal) }} total</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Host Information -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-slate-500 to-slate-600 flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <h3 class="text-xl font-bold text-white">Host Information</h3>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">Host ID</span>
              <span class="text-white font-mono text-sm">{{ host?.id }}</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">Name</span>
              <span class="text-white">{{ host?.name || 'Not Set' }}</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">URI</span>
              <span class="text-white font-mono text-sm break-all">{{ host?.uri }}</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">Hypervisor</span>
              <span class="text-white">{{ getHypervisorType(host?.uri) }}</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">Architecture</span>
              <span class="text-white">{{ hostCapabilities?.host_info?.architecture || 'Unknown' }}</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-slate-700/50">
              <span class="text-slate-400">CPU Cores</span>
              <span class="text-white">{{ hostCapabilities?.host_info?.cpus || hostStats.cpuCores }}</span>
            </div>
            <div class="flex justify-between items-center py-3">
              <span class="text-slate-400">Created</span>
              <span class="text-white text-sm">{{ formatDate(host?.createdAt) }}</span>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Virtualization Capabilities -->
    <FCard class="card-glow">
      <div class="p-6">
        <div class="flex items-center gap-3 mb-6">
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center">
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <h3 class="text-xl font-bold text-white">Virtualization Capabilities</h3>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <!-- KVM Support -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div :class="[
              'w-3 h-3 rounded-full',
              hostCapabilities?.virt_info?.hypervisor?.toLowerCase().includes('kvm') || hostCapabilities?.virt_info?.nested_virt ? 'bg-green-400' : 'bg-gray-400'
            ]"></div>
            <div>
              <div class="text-white font-medium text-sm">KVM Support</div>
              <div class="text-slate-400 text-xs">Hardware virtualization</div>
            </div>
          </div>

          <!-- Nested Virtualization -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div :class="[
              'w-3 h-3 rounded-full',
              hostCapabilities?.virt_info?.nested_virt ? 'bg-green-400' : 'bg-gray-400'
            ]"></div>
            <div>
              <div class="text-white font-medium text-sm">Nested Virt</div>
              <div class="text-slate-400 text-xs">VMs in VMs</div>
            </div>
          </div>

          <!-- IOMMU Support -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div :class="[
              'w-3 h-3 rounded-full',
              hostCapabilities?.security_info?.iommu ? 'bg-green-400' : 'bg-gray-400'
            ]"></div>
            <div>
              <div class="text-white font-medium text-sm">IOMMU</div>
              <div class="text-slate-400 text-xs">Device passthrough</div>
            </div>
          </div>

          <!-- Live Migration -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div :class="[
              'w-3 h-3 rounded-full',
              hostCapabilities?.virt_info?.max_vcpus > 0 ? 'bg-green-400' : 'bg-gray-400'
            ]"></div>
            <div>
              <div class="text-white font-medium text-sm">Live Migration</div>
              <div class="text-slate-400 text-xs">VM relocation</div>
            </div>
          </div>

          <!-- SPICE Protocol -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div class="w-3 h-3 bg-green-400 rounded-full"></div>
            <div>
              <div class="text-white font-medium text-sm">SPICE</div>
              <div class="text-slate-400 text-xs">Remote display</div>
            </div>
          </div>

          <!-- VNC Protocol -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div class="w-3 h-3 bg-green-400 rounded-full"></div>
            <div>
              <div class="text-white font-medium text-sm">VNC</div>
              <div class="text-slate-400 text-xs">Console access</div>
            </div>
          </div>

          <!-- Snapshots -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div class="w-3 h-3 bg-green-400 rounded-full"></div>
            <div>
              <div class="text-white font-medium text-sm">Snapshots</div>
              <div class="text-slate-400 text-xs">VM backups</div>
            </div>
          </div>

          <!-- Max VMs -->
          <div class="flex items-center gap-3 p-4 bg-slate-800/50 rounded-lg">
            <div class="w-3 h-3 bg-blue-400 rounded-full"></div>
            <div>
              <div class="text-white font-medium text-sm">Max VMs</div>
              <div class="text-slate-400 text-xs">{{ hostCapabilities?.virt_info?.max_vcpus || 'Unknown' }} vCPUs</div>
            </div>
          </div>
        </div>
      </div>
    </FCard>

    <!-- Virtual Machines Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Managed Virtual Machines -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300 cursor-pointer" interactive @click="goToManagedVMs">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-lg shadow-primary-500/25">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="text-xl font-bold text-white">Managed Virtual Machines</h3>
              <p class="text-slate-400 text-sm">Virtumancer managed VMs</p>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
          
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Total VMs</span>
              <span class="text-2xl font-bold text-white">{{ hostVMs.length }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Running</span>
              <span class="text-2xl font-bold text-green-400">{{ activeVMsCount }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Stopped</span>
              <span class="text-2xl font-bold text-slate-400">{{ stoppedVMsCount }}</span>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Discovered Virtual Machines -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300 cursor-pointer" interactive @click="goToDiscoveredVMs">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center shadow-lg shadow-orange-500/25">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="text-xl font-bold text-white">Discovered Virtual Machines</h3>
              <p class="text-slate-400 text-sm">Unmanaged VMs on host</p>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
          
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Discovered</span>
              <span class="text-2xl font-bold text-white">{{ discoveredVMsCount }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Available to Import</span>
              <span class="text-2xl font-bold text-orange-400">{{ discoveredVMsCount }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Last Scan</span>
              <span class="text-sm text-slate-500">{{ formatDate(new Date().toISOString()) }}</span>
            </div>
          </div>
        </div>
      </FCard>
    </div>
    
    <!-- Host Settings Modal -->
    <HostSettingsModal 
      :open="showHostSettings" 
      :hostId="hostId" 
      @close="showHostSettings = false"
    />
    
    <!-- Connect Confirmation Modal -->
    <BaseModal
      :show="showConnectConfirmation"
      title="Connect Host"
      description="Are you sure you want to connect to this host?"
      confirm-text="Connect"
      confirm-variant="primary"
      :show-default-actions="true"
      @close="showConnectConfirmation = false"
      @confirm="handleConnectConfirm"
    >
      <div class="space-y-4">
        <p class="text-slate-300">This will establish a connection to:</p>
        <div class="p-3 bg-slate-800/50 rounded-lg">
          <div class="font-medium text-white">{{ host?.name || 'Unknown Host' }}</div>
          <div class="text-sm text-slate-400">{{ host?.uri }}</div>
        </div>
      </div>
    </BaseModal>
    
    <!-- Disconnect Confirmation Modal -->
    <BaseModal
      :show="showDisconnectConfirmation"
      title="Disconnect Host"
      description="Are you sure you want to disconnect from this host?"
      confirm-text="Disconnect"
      confirm-variant="danger"
      :show-default-actions="true"
      @close="showDisconnectConfirmation = false"
      @confirm="handleDisconnectConfirm"
    >
      <div class="space-y-4">
        <p class="text-slate-300">This will disconnect from:</p>
        <div class="p-3 bg-slate-800/50 rounded-lg">
          <div class="font-medium text-white">{{ host?.name || 'Unknown Host' }}</div>
          <div class="text-sm text-slate-400">{{ host?.uri }}</div>
        </div>
        <div class="p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.664-.833-2.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"/>
            </svg>
            <span class="text-sm text-yellow-400 font-medium">Warning</span>
          </div>
          <p class="text-sm text-yellow-300 mt-1">
            The host will not automatically reconnect until you manually connect again.
          </p>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import HostSettingsModal from '@/components/modals/HostSettingsModal.vue'

// Modal state
const showHostSettings = ref(false)
const showConnectConfirmation = ref(false)
const showDisconnectConfirmation = ref(false)

// Loading states
const isRefreshing = ref(false)

const router = useRouter()
const route = useRoute()
const hostStore = useHostStore()
const vmStore = useVMStore()

// router uses the param name `hostId` (see router/index.ts).
// Accept either `hostId` or legacy `id` for robustness.
const hostId = computed(() => (route.params.hostId ?? route.params.id) as string)

const host = computed(() => {
  return hostStore.hosts.find(h => h.id === hostId.value)
})

const hostVMs = computed(() => {
  return vmStore.vms.filter(vm => vm.hostId === hostId.value)
})

const hostStats = computed(() => {
  const stats = hostStore.hostStats[hostId.value]
  const capabilities = hostStore.hostCapabilities[hostId.value]
  if (!stats) {
    return {
      cpuUsage: 0,
      memoryUsage: 0,
      cpuCores: 0,
      memoryTotal: 0,
      storageUsage: 0,
      storageTotal: 0,
      networkInterfaces: 0
    }
  }

  // Calculate memory usage percentage
  const memoryUsage = stats.memory_total && stats.memory_available 
    ? Math.round(((stats.memory_total - stats.memory_available) / stats.memory_total) * 100)
    : 0

  // Calculate storage usage percentage
  const storageUsage = stats.disk_total && stats.disk_free
    ? Math.round(((stats.disk_total - stats.disk_free) / stats.disk_total) * 100)
    : 0

  return {
    cpuUsage: stats.cpu_percent || 0,
    memoryUsage: memoryUsage,
    cpuCores: stats.host_info?.cpu || stats.resources?.cpu_count || 0,
    memoryTotal: stats.memory_total || 0,
    storageUsage: storageUsage,
    storageTotal: stats.disk_total || 0,
    networkInterfaces: capabilities?.network_info?.networks?.length || 0,
    // Calculate vCPU usage based on allocated vCPUs vs total cores
    allocatedVcpus: hostVMs.value.reduce((total, vm) => total + (vm.vcpu_count || 0), 0),
    vcpuUsage: hostVMs.value.length > 0 ? 
      (hostVMs.value.reduce((total, vm) => total + (vm.vcpu_count || 0), 0) / 
       (stats.host_info?.cpu || stats.resources?.cpu_count || 1)) * 100 : 0
  }
})

const hostCapabilities = computed(() => {
  return hostStore.hostCapabilities[hostId.value] || null
})

// Check if stats data is available and meaningful
const hasStatsData = computed(() => {
  const stats = hostStore.hostStats[hostId.value]
  // Check if stats object exists and has the required fields
  return stats && 
         typeof stats.cpu_percent === 'number' && 
         typeof stats.memory_total === 'number' && 
         typeof stats.disk_total === 'number'
})

// VM statistics computed properties
const activeVMsCount = computed(() => {
  return hostVMs.value.filter(vm => vm.state === 'ACTIVE').length
})

const stoppedVMsCount = computed(() => {
  return hostVMs.value.filter(vm => vm.state === 'STOPPED').length
})

const pausedVMsCount = computed(() => {
  return hostVMs.value.filter(vm => vm.state !== 'ACTIVE' && vm.state !== 'STOPPED').length
})

// Discovered VMs count - connect to actual store data
const discoveredVMsCount = computed(() => {
  const hostDiscoveredVMs = hostStore.allDiscoveredVMs.filter(vm => vm.host_id === hostId.value)
  return hostDiscoveredVMs.length
})

// Navigation methods
const goToManagedVMs = () => {
  router.push({ path: '/vms/managed', query: { hostId: hostId.value } })
}

const goToDiscoveredVMs = () => {
  router.push({ path: '/vms/discovered', query: { hostId: hostId.value } })
}

// Helper methods
const formatDate = (dateString?: string): string => {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const getHypervisorType = (uri?: string): string => {
  if (!uri) return 'Unknown'
  if (uri.includes('qemu')) return 'QEMU/KVM'
  if (uri.includes('xen')) return 'Xen'
  if (uri.includes('vmware')) return 'VMware'
  if (uri.includes('vbox')) return 'VirtualBox'
  return 'Libvirt'
}

const getTaskStatusText = (taskState?: string): string => {
  switch (taskState) {
    case 'CONNECTING':
      return 'Connecting...'
    case 'DISCONNECTING':
      return 'Disconnecting...'
    default:
      return taskState || ''
  }
}

const openVMConsole = (vm: any) => {
  // Open console in new window/tab
  const consoleUrl = `/spice/${vm.hostId}/${vm.name}`
  window.open(consoleUrl, '_blank')
}

const refreshHost = async () => {
  if (hostId.value) {
    await hostStore.fetchHostStats(hostId.value)
    await hostStore.fetchHostCapabilities(hostId.value)
    await vmStore.fetchVMs(hostId.value)
  }
}

const discoverVMs = async () => {
  if (hostId.value) {
    await hostStore.refreshDiscoveredVMs(hostId.value)
  }
}

const openHostSettings = () => {
  showHostSettings.value = true
}

const handleDisconnectConfirm = async () => {
  if (hostId.value) {
    try {
      // Immediately dismiss the modal
      showDisconnectConfirmation.value = false
      await hostStore.disconnectHost(hostId.value)
      // Set auto reconnect disabled to prevent auto-reconnect
      await hostStore.updateHost(hostId.value, { auto_reconnect_disabled: true })
    } catch (error) {
      console.error('Failed to disconnect host:', error)
    }
  }
}

const handleConnectConfirm = async () => {
  if (hostId.value) {
    try {
      // Immediately dismiss the modal
      showConnectConfirmation.value = false
      await hostStore.connectHost(hostId.value)
      // Re-enable auto reconnect
      await hostStore.updateHost(hostId.value, { auto_reconnect_disabled: false })
      // Refresh data after connecting
      await refreshHost()
    } catch (error) {
      console.error('Failed to connect host:', error)
    }
  }
}

const openCreateVMModal = () => {
  // TODO: Open create VM modal for this host
  console.log('Create VM modal not implemented yet')
}

onMounted(async () => {
  await hostStore.fetchHosts()
  await hostStore.fetchGlobalDiscoveredVMs() // Fetch discovered VMs on mount
  if (hostId.value) {
    await vmStore.fetchVMs(hostId.value)
    await hostStore.fetchHostStats(hostId.value)
    await hostStore.fetchHostCapabilities(hostId.value)
  }
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(30, 41, 59, 0.3);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(99, 102, 241, 0.5);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(99, 102, 241, 0.7);
}
</style>