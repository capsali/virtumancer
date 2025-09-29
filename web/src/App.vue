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
              <h2 class="text-2xl font-bold text-white">{{ currentView?.title || 'VirtuMancer' }}</h2>
              <div class="px-3 py-1 bg-primary-500/20 text-primary-400 rounded-full text-sm font-medium">
                {{ currentView?.status || 'Active' }}
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
                <!-- Main Content -->
        <main class="flex-1 p-6 overflow-y-auto">
          <!-- Router View for all pages -->
          <router-view />
        </main>
      </div>
    </div>

    <!-- Error Notifications -->
    <ErrorNotificationManager />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import FSidebar from './components/layout/FSidebar.vue';
import FCard from './components/ui/FCard.vue';
import FButton from './components/ui/FButton.vue';
import FInput from './components/ui/FInput.vue';
import FThemeToggle from './components/ui/FThemeToggle.vue';
import ErrorNotificationManager from './components/ui/ErrorNotificationManager.vue';
import { useTheme, initializeTheme } from './composables/useTheme';
import { useAppStore, useUIStore, useHostStore, useVMStore } from './stores';
import { errorRecoveryService } from './services/errorRecovery';

// Initialize theme system
onMounted(async () => {
  initializeTheme();
  
  // Initialize application stores
  try {
    const appStore = useAppStore();
    await appStore.initialize();
  } catch (error) {
    console.error('Failed to initialize application:', error);
    // Use error recovery service for initialization errors
    errorRecoveryService.addError(
      error instanceof Error ? error : new Error(String(error)), 
      'application_initialization',
      { stage: 'startup' }
    );
  }
});

// Cleanup on unmount
onUnmounted(() => {
  const appStore = useAppStore();
  appStore.cleanup();
  
  // Cleanup error recovery service
  errorRecoveryService.destroy();
});

const { themeClasses } = useTheme();
const route = useRoute();

// Store instances
const uiStore = useUIStore();
const hostStore = useHostStore();
const vmStore = useVMStore();
const appStore = useAppStore();

// Current view information based on route
const currentView = computed(() => {
  const path = route.path;
  
  // Dynamic route matching
  if (path === '/') {
    return { title: 'Home', status: 'All Systems Operational' };
  } else if (path === '/network') {
    return { title: 'Network Topology', status: 'Active' };
  } else if (path === '/settings') {
    return { title: 'Settings', status: 'Configuration' };
  } else if (path === '/vms') {
    return { title: 'Virtual Machines', status: 'Managing VMs' };
  } else if (path === '/logs') {
    return { title: 'Logs', status: 'Monitoring' };
  } else if (path.startsWith('/hosts/')) {
    if (path.includes('/vms/')) {
      const vmName = route.params.vmName as string;
      return { title: `VM: ${vmName || 'Details'}`, status: 'VM Management' };
    } else {
      return { title: 'Host Dashboard', status: 'Host Management' };
    }
  } else if (path.startsWith('/vnc/') || path.startsWith('/spice/')) {
    const vmName = route.params.vmName as string;
    return { title: `Console: ${vmName || 'VM'}`, status: 'Remote Access' };
  }
  
  // Default fallback for unknown routes
  return { title: 'Virtumancer', status: 'Active' };
});

// Use real data from stores
const sidebarCollapsed = ref(uiStore.sidebarCollapsed);

// UI handlers
const handleSidebarToggle = (collapsed: boolean) => {
  sidebarCollapsed.value = collapsed;
  uiStore.setSidebarCollapsed(collapsed);
};
</script>
