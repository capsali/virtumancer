<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">{{ host?.name || 'Host Details' }}</h1>
        <p class="text-slate-400 mt-2">{{ host?.uri }}</p>
      </div>
      <div class="flex items-center gap-4">
        <span :class="[
          'px-3 py-1.5 rounded-full text-sm font-medium',
          host?.state === 'CONNECTED' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
        ]">
          {{ host?.state || 'UNKNOWN' }}
        </span>
        <FButton variant="secondary" @click="router.push('/hosts')">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m0 0h6a2 2 0 012 2v1M3 12h6m6 7l7-7m0 0l-7-7m0 0H3a2 2 0 00-2 2v1"/>
          </svg>
          Back to Hosts
        </FButton>
      </div>
    </div>

    <!-- Host Statistics Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
      <!-- Connection Status Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div :class="[
                'w-12 h-12 rounded-xl flex items-center justify-center shadow-lg',
                host?.state === 'CONNECTED' ? 'bg-gradient-to-br from-green-500 to-green-600 shadow-green-500/25' : 'bg-gradient-to-br from-red-500 to-red-600 shadow-red-500/25'
              ]">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div :class="[
              'text-2xl font-bold mb-2',
              host?.state === 'CONNECTED' ? 'text-green-400' : 'text-red-400'
            ]">{{ host?.state || 'UNKNOWN' }}</div>
            <div class="text-sm text-slate-400">Connection Status</div>
          </div>
        </div>
      </FCard>
      
      <!-- VMs Count Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostVMs.length }}</div>
            <div class="text-sm text-slate-400">Virtual Machines</div>
          </div>
        </div>
      </FCard>
      
      <!-- CPU Usage Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg shadow-purple-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostStats.cpuUsage }}%</div>
            <div class="text-sm text-slate-400">CPU Usage</div>
          </div>
        </div>
      </FCard>
      
      <!-- Memory Usage Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-amber-600 flex items-center justify-center shadow-lg shadow-amber-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
              </div>
            </div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-white mb-2">{{ hostStats.memoryUsage }}%</div>
            <div class="text-sm text-slate-400">Memory Usage</div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Host Details and Capabilities -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Host Information -->
      <FCard class="card-glow">
        <div class="p-6">
          <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
            <svg class="w-5 h-5 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            Host Information
          </h2>
          <div class="space-y-4">
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">Host ID:</span>
              <span class="text-white font-mono text-sm">{{ host?.id }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">Name:</span>
              <span class="text-white">{{ host?.name || 'Not Set' }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">URI:</span>
              <span class="text-white font-mono text-sm break-all">{{ host?.uri }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">State:</span>
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium',
                host?.state === 'CONNECTED' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
              ]">{{ host?.state }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">Hypervisor:</span>
              <span class="text-white">{{ getHypervisorType(host?.uri) }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">Auto Reconnect:</span>
              <span :class="[
                'px-2 py-1 rounded text-xs font-medium',
                !host?.auto_reconnect_disabled ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
              ]">{{ host?.auto_reconnect_disabled ? 'Disabled' : 'Enabled' }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-slate-700/50">
              <span class="text-slate-400">Created:</span>
              <span class="text-white text-sm">{{ formatDate(host?.createdAt) }}</span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-slate-400">Last Updated:</span>
              <span class="text-white text-sm">{{ formatDate(host?.updatedAt) }}</span>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Resource Usage & Statistics -->
      <FCard class="card-glow">
        <div class="p-6">
          <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
            <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
            </svg>
            Resource Usage
          </h2>
          
          <!-- CPU Usage -->
          <div class="space-y-4">
            <div>
              <div class="flex justify-between mb-2">
                <span class="text-slate-400 text-sm">CPU Usage</span>
                <span class="text-white font-medium">{{ hostStats.cpuUsage }}%</span>
              </div>
              <div class="w-full bg-slate-700/50 rounded-full h-2">
                <div 
                  class="bg-gradient-to-r from-purple-500 to-purple-600 h-2 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.cpuUsage}%` }"
                ></div>
              </div>
              <div class="text-xs text-slate-500 mt-1">{{ hostStats.cpuCores || 'N/A' }} cores available</div>
            </div>
            
            <!-- Memory Usage -->
            <div>
              <div class="flex justify-between mb-2">
                <span class="text-slate-400 text-sm">Memory Usage</span>
                <span class="text-white font-medium">{{ hostStats.memoryUsage }}%</span>
              </div>
              <div class="w-full bg-slate-700/50 rounded-full h-2">
                <div 
                  class="bg-gradient-to-r from-amber-500 to-amber-600 h-2 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.memoryUsage}%` }"
                ></div>
              </div>
              <div class="text-xs text-slate-500 mt-1">{{ formatBytes(hostStats.memoryTotal || 0) }} total</div>
            </div>

            <!-- Storage Usage -->
            <div>
              <div class="flex justify-between mb-2">
                <span class="text-slate-400 text-sm">Storage Usage</span>
                <span class="text-white font-medium">{{ hostStats.storageUsage || 0 }}%</span>
              </div>
              <div class="w-full bg-slate-700/50 rounded-full h-2">
                <div 
                  class="bg-gradient-to-r from-emerald-500 to-emerald-600 h-2 rounded-full transition-all duration-500"
                  :style="{ width: `${hostStats.storageUsage || 0}%` }"
                ></div>
              </div>
              <div class="text-xs text-slate-500 mt-1">{{ formatBytes(hostStats.storageTotal || 0) }} total</div>
            </div>

            <!-- Network Interfaces -->
            <div v-if="hostStats.networkInterfaces">
              <div class="flex justify-between mb-2">
                <span class="text-slate-400 text-sm">Network Interfaces</span>
                <span class="text-white font-medium">{{ hostStats.networkInterfaces }} active</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span class="text-xs text-slate-400">All interfaces operational</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Virtual Machines on Host -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
              Virtual Machines
            </h2>
            <div class="flex items-center gap-2">
              <span class="text-sm text-slate-400">{{ hostVMs.length }} total</span>
              <FButton variant="primary" size="sm">
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Create VM
              </FButton>
            </div>
          </div>
          
          <!-- VM Statistics -->
          <div class="grid grid-cols-3 gap-4 mb-6">
            <div class="text-center p-3 bg-green-500/10 rounded-lg border border-green-500/20">
              <div class="text-lg font-bold text-green-400">{{ activeVMsCount }}</div>
              <div class="text-xs text-green-300">Active</div>
            </div>
            <div class="text-center p-3 bg-red-500/10 rounded-lg border border-red-500/20">
              <div class="text-lg font-bold text-red-400">{{ stoppedVMsCount }}</div>
              <div class="text-xs text-red-300">Stopped</div>
            </div>
            <div class="text-center p-3 bg-yellow-500/10 rounded-lg border border-yellow-500/20">
              <div class="text-lg font-bold text-yellow-400">{{ pausedVMsCount }}</div>
              <div class="text-xs text-yellow-300">Other</div>
            </div>
          </div>
          
          <!-- VM List -->
          <div class="space-y-2 max-h-60 overflow-y-auto custom-scrollbar">
            <div
              v-for="vm in hostVMs"
              :key="vm.uuid"
              class="flex items-center justify-between p-3 glass-subtle rounded-lg hover:glass-medium transition-all duration-200 cursor-pointer"
              @click="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
            >
              <div class="flex items-center gap-3">
                <div :class="[
                  'w-3 h-3 rounded-full',
                  vm.state === 'ACTIVE' ? 'bg-green-400' : 
                  vm.state === 'STOPPED' ? 'bg-slate-400' : 
                  vm.state === 'PAUSED' ? 'bg-yellow-400' : 'bg-red-400'
                ]"></div>
                <div>
                  <div class="text-white font-medium text-sm">{{ vm.name }}</div>
                  <div class="text-slate-400 text-xs flex items-center gap-2">
                    <span>{{ vm.state }}</span>
                    <span>•</span>
                    <span>{{ vm.vcpu_count || 0 }}C</span>
                    <span>•</span>
                    <span>{{ formatBytes(vm.memory_bytes || 0) }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-1">
                <button
                  @click.stop="openVMConsole(vm)"
                  :disabled="vm.state !== 'ACTIVE'"
                  class="p-1 text-blue-400 hover:bg-blue-500/20 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Open Console"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm8 2a1 1 0 100 2h2a1 1 0 100-2h-2z" clip-rule="evenodd" />
                  </svg>
                </button>
                <button
                  @click.stop="$router.push(`/hosts/${vm.hostId}/vms/${vm.name}`)"
                  class="p-1 text-primary-400 hover:text-primary-300 hover:bg-primary-500/20 rounded transition-colors"
                  title="View Details"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                  </svg>
                </button>
              </div>
            </div>
            
            <div v-if="hostVMs.length === 0" class="text-center py-8 text-slate-400">
              <svg class="w-12 h-12 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd"/>
              </svg>
              <p class="text-sm">No virtual machines found on this host</p>
              <p class="text-xs text-slate-500 mt-1">Create your first VM to get started</p>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Host Capabilities & Actions -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Host Capabilities -->
      <FCard class="card-glow">
        <div class="p-6">
          <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
            <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            Host Capabilities
          </h2>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">KVM Support</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">Hardware Virtualization</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">IOMMU Support</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">Nested Virtualization</span>
              </div>
            </div>
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">SPICE Protocol</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">VNC Protocol</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">Live Migration</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-green-400 rounded-full"></div>
                <span class="text-sm text-slate-300">Snapshots</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>

    <!-- Quick Actions -->
    <FCard class="card-glow">
      <div class="p-6">
        <h2 class="text-xl font-semibold text-white mb-6">Quick Actions</h2>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <FButton variant="secondary" class="w-full justify-center">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            Refresh Host
          </FButton>
          <FButton variant="secondary" class="w-full justify-center">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
            </svg>
            Discover VMs
          </FButton>
          <FButton variant="secondary" class="w-full justify-center">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
            Host Settings
          </FButton>
          <FButton variant="danger" class="w-full justify-center">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            Disconnect
          </FButton>
        </div>
      </div>
    </FCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import FCard from '@/components/ui/FCard.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import FButton from '@/components/ui/FButton.vue'

const router = useRouter()
const route = useRoute()
const hostStore = useHostStore()
const vmStore = useVMStore()

const hostId = computed(() => route.params.id as string)

const host = computed(() => {
  return hostStore.hosts.find(h => h.id === hostId.value)
})

const hostVMs = computed(() => {
  return vmStore.vms.filter(vm => vm.hostId === hostId.value)
})

const hostStats = ref({
  cpuUsage: 45,
  memoryUsage: 62,
  cpuCores: 8,
  memoryTotal: 34359738368, // 32GB in bytes
  storageUsage: 35,
  storageTotal: 1099511627776, // 1TB in bytes
  networkInterfaces: 4
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

const openVMConsole = (vm: any) => {
  // Open console in new window/tab
  const consoleUrl = `/spice/${vm.hostId}/${vm.name}`
  window.open(consoleUrl, '_blank')
}

onMounted(async () => {
  await hostStore.fetchHosts()
  if (hostId.value) {
    await vmStore.fetchVMs(hostId.value)
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