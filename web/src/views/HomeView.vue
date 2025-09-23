<template>
  <div class="space-y-8">
    <!-- Welcome Section -->
    <div class="text-center">
      <h2 class="text-4xl font-bold text-white mb-4">Welcome to VirtuMancer</h2>
      <p class="text-slate-400 text-lg">Manage your virtualization infrastructure with style</p>
    </div>
    
    <!-- Dashboard Stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <FCard
        v-for="(stat, index) in stats"
        :key="stat.id"
        :class="`animate-fade-in delay-${(index + 1) * 100}`"
        :border-glow="true"
        :glow-color="stat.glowColor"
        interactive
      >
        <div class="flex items-center gap-4">
          <div :class="[
            'w-12 h-12 rounded-xl flex items-center justify-center',
            stat.iconBg
          ]">
            <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="stat.iconPath"/>
            </svg>
          </div>
          <div class="flex-1">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold text-white">{{ stat.value }}</h3>
              <span :class="[
                'px-2 py-1 rounded-full text-xs font-medium',
                stat.trend === 'up' ? 'bg-green-500/20 text-green-400' :
                stat.trend === 'down' ? 'bg-red-500/20 text-red-400' :
                'bg-slate-500/20 text-slate-400'
              ]">
                {{ stat.change }}
              </span>
            </div>
            <p class="text-sm text-slate-400">{{ stat.label }}</p>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Quick Actions -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <FCard
        v-for="(action, index) in quickActions"
        :key="action.id"
        :class="`animate-fade-in delay-${(index + 1) * 150}`"
        :border-glow="true"
        :glow-color="action.glowColor"
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
          class="animate-fade-in"
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
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import AddHostModal from '@/components/modals/AddHostModal.vue';
import { useAppStore, useHostStore, useVMStore } from '@/stores';

const router = useRouter();
const hostStore = useHostStore();
const vmStore = useVMStore();
const appStore = useAppStore();

// Modal state
const showAddHostModal = ref(false);

// Dashboard statistics
const stats = computed(() => [
  {
    id: 'vms',
    label: 'Virtual Machines',
    value: vmStore.vms.length.toString(),
    change: `${vmStore.activeVMs.length} running`,
    trend: 'up' as const,
    iconBg: 'bg-gradient-to-br from-primary-500 to-primary-600',
    iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',
    glowColor: 'primary' as const
  },
  {
    id: 'hosts',
    label: 'Active Hosts',
    value: hostStore.connectedHosts.length.toString(),
    change: `${hostStore.hosts.length} total`,
    trend: hostStore.connectedHosts.length > 0 ? 'up' : 'down' as const,
    iconBg: 'bg-gradient-to-br from-accent-500 to-accent-600',
    iconPath: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
    glowColor: 'accent' as const
  },
  {
    id: 'system',
    label: 'System Status',
    value: appStore.connectionStatus === 'connected' ? 'Online' : 'Offline',
    change: appStore.healthStatus,
    trend: appStore.healthStatus === 'healthy' ? 'up' : 'down' as const,
    iconBg: 'bg-gradient-to-br from-neon-purple to-neon-pink',
    iconPath: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    glowColor: 'neon-purple' as const
  },
  {
    id: 'sync',
    label: 'Last Sync',
    value: appStore.lastSyncTime ? new Date(appStore.lastSyncTime).toLocaleTimeString() : 'Never',
    change: appStore.isSyncing ? 'Syncing...' : 'Up to date',
    trend: 'stable' as const,
    iconBg: 'bg-gradient-to-br from-green-500 to-green-600',
    iconPath: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15',
    glowColor: 'primary' as const
  }
]);

// Quick action buttons
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
      // TODO: Navigate to monitoring dashboard
      console.log('Monitoring dashboard would open here');
    }
  }
]);

// Recent activity (mock data for now)
const recentActivities = computed(() => {
  interface Activity {
    type: string;
    title: string;
    description: string;
    timestamp: string;
    iconPath: string;
  }
  
  const activities: Activity[] = [];
  
  // Add VM activities
  vmStore.vms.slice(0, 3).forEach(vm => {
    const host = hostStore.hosts.find(h => h.id === vm.hostId);
    activities.push({
      type: 'vm',
      title: `${vm.name} is ${vm.state}`,
      description: `Virtual machine on ${host?.uri || 'unknown host'}`,
      timestamp: 'Just now',
      iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z'
    });
  });
  
  // Add host activities
  hostStore.connectedHosts.slice(0, 2).forEach(host => {
    activities.push({
      type: 'host',
      title: `${host.uri} connected`,
      description: `Hypervisor host is online and ready`,
      timestamp: '5 min ago',
      iconPath: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10'
    });
  });
  
  return activities.slice(0, 5); // Show only latest 5 activities
});

const handleActionClick = (action: any) => {
  if (action.action) {
    action.action();
  }
};

const handleHostAdded = (host: any) => {
  console.log('Host added successfully:', host);
  // Optionally navigate to the new host
  if (host?.id) {
    router.push(`/hosts/${host.id}`);
  }
};
</script>