<template>
  <div class="space-y-8">
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent mb-4">
        Welcome to Virtumancer
      </h2>
      <p class="text-slate-400 text-lg">Your virtualization command center</p>
    </div>

    <!-- Main Infrastructure Cards -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Virtual Machines Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="router.push('/vms')">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-lg shadow-primary-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Virtual Machines</h3>
                <p class="text-slate-400 text-sm">Infrastructure VMs</p>
              </div>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
          
          <div class="grid grid-cols-3 gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-white">{{ dashboardStats.infrastructure.totalVMs }}</div>
              <div class="text-xs text-slate-400">Total</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-green-400">{{ dashboardStats.infrastructure.runningVMs }}</div>
              <div class="text-xs text-slate-400">Running</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-slate-400">{{ dashboardStats.infrastructure.stoppedVMs }}</div>
              <div class="text-xs text-slate-400">Stopped</div>
            </div>
          </div>
          
          <div class="mt-4 h-2 bg-slate-700 rounded-full overflow-hidden">
            <div 
              class="h-full bg-gradient-to-r from-green-500 to-primary-500 transition-all duration-300"
              :style="{ width: `${dashboardStats.infrastructure.totalVMs > 0 ? (dashboardStats.infrastructure.runningVMs / dashboardStats.infrastructure.totalVMs) * 100 : 0}%` }"
            ></div>
          </div>
        </div>
      </FCard>

      <!-- Hosts Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="router.push('/hosts')">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-accent-500 to-accent-600 flex items-center justify-center shadow-lg shadow-accent-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Hosts</h3>
                <p class="text-slate-400 text-sm">Hypervisor nodes</p>
              </div>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div class="text-center">
              <div class="text-3xl font-bold text-green-400">{{ dashboardStats.infrastructure.connectedHosts }}</div>
              <div class="text-xs text-slate-400">Connected</div>
            </div>
            <div class="text-center">
              <div class="text-3xl font-bold text-white">{{ dashboardStats.infrastructure.totalHosts }}</div>
              <div class="text-xs text-slate-400">Total</div>
            </div>
          </div>
          
          <div class="mt-6 flex items-center justify-center">
            <div :class="[
              'w-3 h-3 rounded-full mr-2',
              dashboardStats.infrastructure.connectedHosts > 0 ? 'bg-green-400 animate-pulse' : 'bg-red-400'
            ]"></div>
            <span class="text-sm text-slate-400">
              {{ dashboardStats.infrastructure.connectedHosts > 0 ? 'All systems operational' : 'No hosts connected' }}
            </span>
          </div>
        </div>
      </FCard>

      <!-- Networks Card -->
      <FCard class="card-glow hover:scale-105 transition-all duration-300" interactive @click="router.push('/network')">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-lg shadow-purple-500/25">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                </svg>
              </div>
              <div>
                <h3 class="text-xl font-bold text-white">Networks</h3>
                <p class="text-slate-400 text-sm">Network topology</p>
              </div>
            </div>
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
          
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Active Connections</span>
              <span class="text-lg font-bold text-green-400">{{ dashboardStats.infrastructure.connectedHosts * 2 }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-400">Network Interfaces</span>
              <span class="text-lg font-bold text-white">{{ dashboardStats.infrastructure.totalVMs * 1.5 | 0 }}</span>
            </div>
          </div>
          
          <div class="mt-6">
            <div class="flex items-center justify-center space-x-1">
              <div v-for="i in 8" :key="i" class="w-2 h-2 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full animate-pulse" :style="{ animationDelay: `${i * 0.1}s` }"></div>
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Resource Overview Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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
          
          <div class="space-y-4">
            <div>
              <div class="flex justify-between items-center mb-2">
                <span class="text-sm text-slate-400">Memory Usage</span>
                <span class="text-sm font-medium text-white">
                  {{ Math.round(dashboardStats.resources.memoryUtilization) }}%
                </span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-2">
                <div 
                  class="h-2 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full transition-all duration-300"
                  :style="{ width: `${dashboardStats.resources.memoryUtilization}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-500 mt-1">
                <span>{{ Math.round(dashboardStats.resources.usedMemoryGB) }} GB used</span>
                <span>{{ Math.round(dashboardStats.resources.totalMemoryGB) }} GB total</span>
              </div>
            </div>
            
            <div>
              <div class="flex justify-between items-center mb-2">
                <span class="text-sm text-slate-400">CPU Allocation</span>
                <span class="text-sm font-medium text-white">
                  {{ Math.round(dashboardStats.resources.cpuUtilization) }}%
                </span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-2">
                <div 
                  class="h-2 bg-gradient-to-r from-green-500 to-emerald-500 rounded-full transition-all duration-300"
                  :style="{ width: `${dashboardStats.resources.cpuUtilization}%` }"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-slate-500 mt-1">
                <span>{{ dashboardStats.resources.allocatedCPUs }} allocated</span>
                <span>{{ dashboardStats.resources.totalCPUs }} total cores</span>
              </div>
            </div>
          </div>
        </div>
      </FCard>

      <!-- System Health -->
      <FCard class="card-glow">
        <div class="p-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/>
              </svg>
            </div>
            <h3 class="text-xl font-bold text-white">System Health</h3>
          </div>
          
          <div class="space-y-4">
            <div class="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
              <div class="flex items-center gap-3">
                <div class="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
                <span class="text-white font-medium">System Status</span>
              </div>
              <span class="text-green-400 text-sm font-medium">{{ dashboardStats.health.systemStatus }}</span>
            </div>
            
            <div class="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
              <div class="flex items-center gap-3">
                <div class="w-3 h-3 bg-blue-400 rounded-full"></div>
                <span class="text-white font-medium">Errors</span>
              </div>
              <span class="text-white text-sm font-medium">{{ dashboardStats.health.errors }}</span>
            </div>
            
            <div class="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
              <div class="flex items-center gap-3">
                <div class="w-3 h-3 bg-yellow-400 rounded-full"></div>
                <span class="text-white font-medium">Warnings</span>
              </div>
              <span class="text-white text-sm font-medium">{{ dashboardStats.health.warnings }}</span>
            </div>
            
            <div class="text-xs text-slate-500 text-center pt-2">
              Last sync: {{ new Date(dashboardStats.health.lastSync).toLocaleTimeString() }}
            </div>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Quick Actions -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <FCard
        v-for="(action, index) in quickActions"
        :key="action.id"
        :class="`animate-fade-in delay-${(index + 1) * 150} card-glow`"
        interactive
      >
        <button 
          class="w-full text-center cursor-pointer focus:outline-none"
          @click="handleActionClick(action)"
        >
          <div :class="[
            'w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4',
            action.iconBg,
            action.shadow
          ]">
            <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="action.iconPath"/>
            </svg>
          </div>
          <h3 class="text-xl font-bold text-white mb-2">{{ action.title }}</h3>
          <p class="text-slate-400 mb-6">{{ action.description }}</p>
          <div class="w-full">
            <FButton :variant="action.buttonVariant" class="w-full" @click.stop>
              {{ action.buttonText }}
            </FButton>
          </div>
        </button>
      </FCard>
    </div>

    <!-- Recent Activity -->
    <div v-if="recentActivities.length > 0">
      <h3 class="text-2xl font-bold text-white mb-6">Recent Activity</h3>
      <div class="space-y-4">
        <FCard
          v-for="(activity, index) in recentActivities"
          :key="index"
          class="animate-fade-in card-glow"
          :style="{ animationDelay: `${index * 100}ms` }"
        >
          <div class="flex items-center gap-4">
            <div :class="[
              'w-10 h-10 rounded-full flex items-center justify-center',
              activity.type === 'vm' ? 'bg-primary-500/20 text-primary-400' :
              activity.type === 'host' ? 'bg-accent-500/20 text-accent-400' :
              'bg-slate-500/20 text-slate-400'
            ]">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="activity.iconPath"/>
              </svg>
            </div>
            <div class="flex-1">
              <h4 class="text-white font-medium">{{ activity.title }}</h4>
              <p class="text-slate-400 text-sm">{{ activity.description }}</p>
            </div>
            <div class="text-xs text-slate-500">
              {{ activity.timestamp }}
            </div>
          </div>
        </FCard>
      </div>
    </div>
  </div>

  <!-- Add Host Modal -->
  <AddHostModal
    v-model:open="showAddHostModal"
    @host-added="handleHostAdded"
  />
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import AddHostModal from '@/components/modals/AddHostModal.vue';
import { useAppStore, useHostStore, useVMStore, useUIStore } from '@/stores';
import { dashboardApi } from '@/services/api';

const router = useRouter();
const hostStore = useHostStore();
const vmStore = useVMStore();
const appStore = useAppStore();
const uiStore = useUIStore();

// Modal state
const showAddHostModal = ref(false);

// Dashboard data
const dashboardStats = ref({
  infrastructure: {
    totalHosts: 0,
    connectedHosts: 0,
    totalVMs: 0,
    runningVMs: 0,
    stoppedVMs: 0
  },
  resources: {
    totalMemoryGB: 0,
    usedMemoryGB: 0,
    memoryUtilization: 0,
    totalCPUs: 0,
    allocatedCPUs: 0,
    cpuUtilization: 0
  },
  health: {
    systemStatus: 'healthy',
    lastSync: new Date().toISOString(),
    errors: 0,
    warnings: 0
  }
});

const recentActivitiesData = ref<Array<{
  id: string;
  type: 'vm_state_change' | 'host_connect' | 'host_disconnect' | 'system';
  message: string;
  hostId: string;
  vmUuid?: string;
  vmName?: string;
  timestamp: string;
  severity: 'info' | 'warning' | 'error';
  details?: string;
}>>([]);

const isLoading = ref(true);

// Load dashboard data using the new backend endpoints
const loadDashboardData = async () => {
  isLoading.value = true;
  try {
    const [statsResponse, activityResponse] = await Promise.all([
      dashboardApi.getStats(),
      dashboardApi.getActivity()
    ]);
    
    dashboardStats.value = statsResponse;
    recentActivitiesData.value = activityResponse.activities;
  } catch (error) {
    console.error('Failed to load dashboard data:', error);
    // Fallback to store data if API fails
    dashboardStats.value.infrastructure = {
      totalHosts: hostStore.hosts.length,
      connectedHosts: hostStore.connectedHosts.length,
      totalVMs: vmStore.vms.length,
      runningVMs: vmStore.activeVMs.length,
      stoppedVMs: vmStore.vms.length - vmStore.activeVMs.length
    };
    dashboardStats.value.health.lastSync = new Date().toISOString();
    
    // Ensure recentActivitiesData has a fallback value
    if (!recentActivitiesData.value) {
      recentActivitiesData.value = [];
    }
  } finally {
    isLoading.value = false;
  }
};

// Dashboard statistics computed from real data
const stats = computed(() => {
  const data = dashboardStats.value;
  
  return [
    {
      id: 'vms',
      label: 'Virtual Machines',
      value: data.infrastructure.totalVMs.toString(),
      change: `${data.infrastructure.runningVMs} running`,
      trend: data.infrastructure.runningVMs > 0 ? 'up' : 'stable' as const,
      iconBg: 'bg-gradient-to-br from-primary-500 to-primary-600',
      iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',
      glowColor: 'primary' as const
    },
    {
      id: 'hosts',
      label: 'Active Hosts',
      value: data.infrastructure.connectedHosts.toString(),
      change: `${data.infrastructure.totalHosts} total`,
      trend: data.infrastructure.connectedHosts > 0 ? 'up' : 'down' as const,
      iconBg: 'bg-gradient-to-br from-accent-500 to-accent-600',
      iconPath: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
      glowColor: 'accent' as const
    },
    {
      id: 'memory',
      label: 'Memory Usage',
      value: data.resources.memoryUtilization > 0 ? `${Math.round(data.resources.memoryUtilization)}%` : 'N/A',
      change: data.resources.totalMemoryGB > 0 ? `${Math.round(data.resources.usedMemoryGB)} GB used` : 'No data',
      trend: data.resources.memoryUtilization > 80 ? 'down' : 'stable' as const,
      iconBg: 'bg-gradient-to-br from-neon-purple to-neon-pink',
      iconPath: 'M8 16l2.879-2.879m0 0a3 3 0 104.243-4.242 3 3 0 00-4.243 4.242zM21 12a9 9 0 11-18 0 9 9 0 0118 0z',
      glowColor: 'neon-purple' as const
    },
    {
      id: 'cpu',
      label: 'CPU Allocation',
      value: data.resources.allocatedCPUs.toString(),
      change: `${data.resources.totalCPUs} total CPUs`,
      trend: 'stable' as const,
      iconBg: 'bg-gradient-to-br from-green-500 to-green-600',
      iconPath: 'M13 10V3L4 14h7v7l9-11h-7z',
      glowColor: 'primary' as const
    }
  ];
});

// Format recent activities for display with icon mapping
const recentActivities = computed(() => {
  // Add null check to prevent map error
  if (!recentActivitiesData.value || !Array.isArray(recentActivitiesData.value)) {
    return [];
  }
  
  return recentActivitiesData.value.map(activity => {
    let iconPath = '';
    let type = '';
    
    // Map activity types to icons and display types
    switch (activity.type) {
      case 'vm_state_change':
        iconPath = 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z';
        type = 'vm';
        break;
      case 'host_connect':
      case 'host_disconnect':
        iconPath = 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10';
        type = 'host';
        break;
      case 'system':
      default:
        iconPath = 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z';
        type = 'system';
        break;
    }
    
    return {
      ...activity,
      iconPath,
      type,
      title: activity.message, // Use message as title for backward compatibility
      description: activity.details || `Activity on host ${activity.hostId}`,
      timestamp: formatTimestamp(activity.timestamp)
    };
  });
});

// Utility function to format timestamps
const formatTimestamp = (timestamp: string) => {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} min ago`;
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)} hr ago`;
  return `${Math.floor(diffMins / 1440)} day ago`;
};

// Quick action buttons - now with real functionality
const quickActions = ref([
  {
    id: 'hosts',
    title: 'Manage Hosts',
    description: 'Add, configure, and monitor hypervisor hosts',
    iconBg: 'bg-gradient-to-br from-accent-500 to-accent-600',
    iconPath: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
    shadow: 'shadow-neon-cyan',
    buttonVariant: 'accent' as const,
    buttonText: 'View Hosts',
    glowColor: 'accent' as const,
    action: () => {
      // Navigate to first available host or show host management
      if (hostStore.connectedHosts.length > 0) {
        const firstHost = hostStore.connectedHosts[0];
        if (firstHost?.id) {
          router.push(`/hosts/${firstHost.id}`);
        }
      } else {
        // Open add host modal
        showAddHostModal.value = true;
      }
    }
  },
  {
    id: 'network',
    title: 'Network Topology',
    description: 'Visualize infrastructure and network relationships',
    iconBg: 'bg-gradient-to-br from-primary-500 to-primary-600',
    iconPath: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
    shadow: 'shadow-neon-blue',
    buttonVariant: 'primary' as const,
    buttonText: 'View Topology',
    glowColor: 'primary' as const,
    action: () => router.push('/network')
  },
  {
    id: 'monitoring',
    title: 'System Monitor',
    description: 'Real-time performance metrics and alerts',
    iconBg: 'bg-gradient-to-br from-neon-purple to-neon-pink',
    iconPath: 'M13 10V3L4 14h7v7l9-11h-7z',
    shadow: '',
    buttonVariant: 'neon' as const,
    buttonText: 'View Metrics',
    glowColor: 'neon-purple' as const,
    action: () => {
      // Refresh dashboard data to show latest metrics
      loadDashboardData();
      uiStore.addToast('Dashboard data refreshed', 'success', 3000);
    }
  }
]);

const handleActionClick = (action: any) => {
  if (action.action) {
    action.action();
  }
};

const handleHostAdded = (host: any) => {
  console.log('Host added successfully:', host);
  // Refresh dashboard data
  loadDashboardData();
  // Optionally navigate to the new host
  if (host?.id) {
    router.push(`/hosts/${host.id}`);
  }
};

// Load data on component mount
onMounted(() => {
  loadDashboardData();
});
</script>