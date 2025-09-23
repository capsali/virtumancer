<template>
  <div id="app" :class="['relative min-h-screen overflow-hidden bg-slate-900', themeClasses]">
    <!-- Animated Background -->
    <div class="fixed inset-0 bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900">
      <!-- Floating Background Elements -->
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-primary-600/10 rounded-full blur-3xl animate-float-gentle"></div>
      <div class="absolute top-1/3 right-1/4 w-80 h-80 bg-accent-500/10 rounded-full blur-3xl animate-float-medium delay-1000"></div>
      <div class="absolute bottom-1/4 left-1/3 w-72 h-72 bg-neon-purple/10 rounded-full blur-3xl animate-float-active delay-2000"></div>
    </div>

    <!-- Main Layout -->
    <div class="relative z-10 flex h-screen">
      <!-- Sidebar -->
      <FSidebar
        v-model:collapsed="sidebarCollapsed"
        @navigate="handleNavigation"
        @update:collapsed="handleSidebarToggle"
      />

      <!-- Main Content Area -->
      <div
        :class="[
          'flex-1 flex flex-col transition-all duration-300',
          sidebarCollapsed ? 'ml-20' : 'ml-72'
        ]"
      >
        <!-- Header -->
        <header class="glass-subtle border-b border-white/10 p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-4">
              <h2 class="text-2xl font-bold text-white">{{ currentView.title }}</h2>
              <div class="px-3 py-1 bg-primary-500/20 text-primary-400 rounded-full text-sm font-medium">
                {{ currentView.status }}
              </div>
            </div>
            
            <div class="flex items-center gap-4">
              <!-- Search -->
              <div class="relative">
                <FInput
                  placeholder="Search..."
                  size="sm"
                  class="w-64"
                />
              </div>
              
              <!-- Notifications -->
              <FButton variant="ghost" size="sm">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-5-5 5-5h-5m-6 0H4l5 5-5 5h5m2-5a3 3 0 11-6 0 3 3 0 016 0z"/>
                </svg>
              </FButton>

              <!-- Theme Toggle -->
              <FThemeToggle />
            </div>
          </div>
        </header>

        <!-- Page Content -->
        <main class="flex-1 p-6 overflow-auto">
          <div v-if="currentView.id === 'dashboard'" class="space-y-6">
            <!-- Welcome Card -->
            <FCard
              :floating-orbs="true"
              :border-glow="true"
              glow-color="primary"
              class="animate-fade-in"
            >
              <div class="text-center">
                <h1 class="text-4xl font-bold bg-gradient-to-r from-primary-400 via-accent-400 to-primary-400 bg-clip-text text-transparent mb-4">
                  Welcome to VirtuMancer
                </h1>
                <p class="text-lg text-slate-300 mb-6">
                  Next-Generation Virtualization Management Platform
                </p>
                <div class="flex justify-center gap-4">
                  <FButton variant="primary" glow>
                    Create New VM
                  </FButton>
                  <FButton variant="accent" glow>
                    Add Host
                  </FButton>
                </div>
              </div>
            </FCard>

            <!-- Stats Grid -->
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
                    <div class="text-2xl font-bold text-white">{{ stat.value }}</div>
                    <div class="text-sm text-slate-400">{{ stat.label }}</div>
                    <div :class="[
                      'text-xs font-medium',
                      stat.trend === 'up' ? 'text-green-400' : 'text-red-400'
                    ]">
                      {{ stat.change }}
                    </div>
                  </div>
                </div>
              </FCard>
            </div>

            <!-- Feature Grid -->
            <div class="grid md:grid-cols-3 gap-6">
              <FCard
                v-for="(feature, index) in features"
                :key="feature.id"
                :class="`animate-slide-up delay-${(index + 1) * 200}`"
                interactive
                :border-glow="true"
                :glow-color="feature.glowColor"
              >
                <div class="text-center">
                  <div :class="[
                    'w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4',
                    feature.iconBg,
                    feature.shadow
                  ]">
                    <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="feature.iconPath"/>
                    </svg>
                  </div>
                  <h3 class="text-lg font-semibold text-white mb-2">{{ feature.title }}</h3>
                  <p class="text-slate-400 text-sm mb-4">{{ feature.description }}</p>
                  <FButton :variant="feature.buttonVariant" size="sm">
                    {{ feature.buttonText }}
                  </FButton>
                </div>
              </FCard>
            </div>
          </div>

          <!-- Other views would go here -->
          <div v-else class="text-center py-12">
            <h3 class="text-2xl font-bold text-white mb-4">{{ currentView.title }}</h3>
            <p class="text-slate-400">This view is under construction...</p>
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import FSidebar from './components/layout/FSidebar.vue';
import FCard from './components/ui/FCard.vue';
import FButton from './components/ui/FButton.vue';
import FInput from './components/ui/FInput.vue';
import FThemeToggle from './components/ui/FThemeToggle.vue';
import { useTheme, initializeTheme } from './composables/useTheme';
import { useAppStore, useUIStore, useHostStore, useVMStore } from './stores';

// Initialize theme system
onMounted(async () => {
  initializeTheme();
  
  // Initialize application stores
  try {
    const appStore = useAppStore();
    await appStore.initialize();
  } catch (error) {
    console.error('Failed to initialize application:', error);
  }
});

// Cleanup on unmount
onUnmounted(() => {
  const appStore = useAppStore();
  appStore.cleanup();
});

const { themeClasses } = useTheme();

// Store instances
const uiStore = useUIStore();
const hostStore = useHostStore();
const vmStore = useVMStore();
const appStore = useAppStore();

// Use real data from stores
const sidebarCollapsed = ref(uiStore.sidebarCollapsed);

const currentView = ref({
  id: 'dashboard',
  title: 'Dashboard',
  status: 'All Systems Operational'
});

// Use computed stats from stores
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
    iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',
    glowColor: 'neon-purple' as const
  },
  {
    id: 'sync',
    label: 'Last Sync',
    value: appStore.lastSyncTime ? new Date(appStore.lastSyncTime).toLocaleTimeString() : 'Never',
    change: appStore.isSyncing ? 'Syncing...' : 'Up to date',
    trend: 'stable' as const,
    iconBg: 'bg-gradient-to-br from-green-500 to-green-600',
    iconPath: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4',
    glowColor: 'primary' as const
  }
]);

const features = ref([
  {
    id: 'vms',
    title: 'Virtual Machines',
    description: 'Manage and monitor your virtual infrastructure with advanced controls',
    iconBg: 'bg-gradient-to-br from-primary-500 to-primary-600',
    iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',
    shadow: 'shadow-neon-blue',
    buttonVariant: 'primary' as const,
    buttonText: 'Manage VMs',
    glowColor: 'primary' as const
  },
  {
    id: 'hosts',
    title: 'Host Management',
    description: 'Centralized control of hypervisor hosts and resource allocation',
    iconBg: 'bg-gradient-to-br from-accent-500 to-accent-600',
    iconPath: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
    shadow: 'shadow-neon-cyan',
    buttonVariant: 'accent' as const,
    buttonText: 'View Hosts',
    glowColor: 'accent' as const
  },
  {
    id: 'monitoring',
    title: 'Real-time Monitoring',
    description: 'Live performance metrics and intelligent resource monitoring',
    iconBg: 'bg-gradient-to-br from-neon-purple to-neon-pink',
    iconPath: 'M13 10V3L4 14h7v7l9-11h-7z',
    shadow: '',
    buttonVariant: 'neon' as const,
    buttonText: 'View Metrics',
    glowColor: 'neon-purple' as const
  }
]);

// Navigation and UI handlers
const handleNavigation = (item: any) => {
  currentView.value = {
    id: item.id,
    title: item.label,
    status: 'Active'
  };
  uiStore.setCurrentView(item.id, [item.label]);
};

const handleSidebarToggle = (collapsed: boolean) => {
  sidebarCollapsed.value = collapsed;
  uiStore.setSidebarCollapsed(collapsed);
};
</script>
