<template>
  <nav
    :class="[
      'glass-strong backdrop-blur-xl border-r border-white/10',
      'transition-all duration-300 flex flex-col',
      sidebarClasses
    ]"
  >
    <!-- Logo Section -->
    <div class="p-6 border-b border-white/10">
      <div class="flex items-center gap-3">
        <!-- Enhanced Logo Icon with VM Symbol -->
        <div class="relative w-12 h-12 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-xl flex items-center justify-center shadow-neon-blue group hover:scale-105 transition-transform duration-300">
          <!-- Main VM/Server Icon -->
          <svg class="w-7 h-7 text-white relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
          </svg>
          <!-- Virtualization Layer Indicator -->
          <div class="absolute inset-1 rounded-lg border border-white/20 opacity-60"></div>
          <!-- Power/Activity Indicator -->
          <div class="absolute top-1 right-1 w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
        </div>
        <div v-if="!collapsed" class="flex-1">
          <div class="flex items-center gap-2">
            <h1 class="text-xl font-bold bg-gradient-to-r from-primary-400 via-accent-400 to-secondary-400 bg-clip-text text-transparent">
              Virtumancer
            </h1>
            <!-- Version Badge -->
            <span class="px-2 py-0.5 bg-primary-500/20 text-primary-300 text-xs font-medium rounded-full">
              v1.0
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-1">Hypervisor Management Platform</p>
        </div>
      </div>

      <!-- Optional: Animated Glow Effect on Hover -->
      <div class="absolute inset-0 bg-gradient-to-r from-primary-600/5 via-accent-600/5 to-secondary-600/5 opacity-0 hover:opacity-100 transition-opacity duration-500 rounded-xl pointer-events-none"></div>
    </div>

    <!-- Navigation Items -->
    <div class="flex-1 p-4 space-y-4 overflow-y-auto">
      <!-- Main Navigation -->
      <div class="space-y-2">
        <div
          v-for="item in navigationItems"
          :key="item.id"
          class="group relative overflow-hidden rounded-xl transition-all duration-300"
        >
          <!-- Main Navigation Item -->
            <div
            :class="[
              'relative overflow-hidden transition-all duration-300',
              {
                'bg-gradient-to-r from-primary-600/20 to-accent-600/20 shadow-glow-sm': item.active,
                'hover:bg-white/5': !item.active
              }
            ]"
          >
            <button
              @click="handleNavItemClick(item)"
              :class="[
                'w-full flex items-center gap-3 p-3 text-left transition-all duration-300',
                {
                  'text-white': item.active,
                  'text-slate-300 hover:text-white': !item.active
                }
              ]"
            >
              <!-- Icon -->
              <div :class="[
                'w-8 h-8 flex items-center justify-center rounded-lg transition-all duration-300',
                {
                  'bg-gradient-to-br from-primary-500 to-accent-500 shadow-neon-blue': item.active,
                  'bg-slate-600/50 group-hover:bg-slate-500/50': !item.active
                }
              ]">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon"/>
                </svg>
              </div>

              <!-- Label -->
              <div v-if="!collapsed" class="flex-1">
                <div class="font-medium">{{ item.label }}</div>
                <div v-if="item.description" class="text-xs text-slate-400">{{ item.description }}</div>
              </div>

              <!-- Expand Arrow (always shows as expandable since clicking expands/keeps expanded) -->
              <div
                v-if="item.children && !collapsed"
                class="w-5 h-5 text-slate-400"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                </svg>
              </div>

              <!-- Badge -->
              <div
                v-if="item.badge && !collapsed && !item.children"
                :class="[
                  'px-2 py-1 rounded-full text-xs font-medium',
                  item.badgeColor || 'bg-accent-500/20 text-accent-400'
                ]"
              >
                {{ item.badge }}
              </div>

              <!-- Active Indicator -->
              <div
                v-if="item.active"
                class="absolute right-0 top-1/2 transform -translate-y-1/2 w-1 h-8 bg-gradient-to-b from-primary-400 to-accent-400 rounded-l-full pointer-events-none"
              ></div>
            </button>

            <!-- Hover Glow Effect -->
            <div
              v-if="!item.active && !item.expanded"
              class="absolute inset-0 bg-gradient-to-r from-primary-600/0 via-primary-600/5 to-accent-600/0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 rounded-xl pointer-events-none"
            ></div>
          </div>

          <!-- Submenu Items -->
          <div
            v-if="item.children && item.expanded && !collapsed"
            class="mt-1 ml-4 space-y-1 animate-slideDown"
          >
            <router-link
              v-for="child in item.children"
              :key="child.id"
              :to="child.path"
              :class="[
                'flex items-center gap-3 p-2 rounded-lg transition-all duration-200 group/child',
                {
                  'bg-primary-500/20 text-primary-300 shadow-glow-xs': child.active,
                  'text-slate-400 hover:text-white hover:bg-white/5': !child.active
                }
              ]"
            >
              <div :class="[
                'w-6 h-6 flex items-center justify-center rounded-md transition-all duration-200',
                {
                  'bg-primary-500/30': child.active,
                  'bg-slate-600/30 group-hover/child:bg-slate-500/30': !child.active
                }
              ]">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="child.icon"/>
                </svg>
              </div>
              <span class="text-sm font-medium">{{ child.label }}</span>
              <span
                v-if="child.badge"
                :class="[
                  'ml-auto px-1.5 py-0.5 rounded text-xs font-medium',
                  child.badgeColor || 'bg-accent-500/20 text-accent-400'
                ]"
              >
                {{ child.badge }}
              </span>
            </router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Section -->
    <div class="p-4 border-t border-white/10 space-y-3">
      <!-- Settings Button -->
      <router-link
        to="/settings"
        :class="[
          'flex items-center gap-3 p-3 rounded-xl glass-subtle hover:glass-medium transition-all duration-300 group',
          {
            'bg-gradient-to-r from-primary-600/20 to-accent-600/20 shadow-glow-sm': route.path === '/settings',
            'hover:bg-white/5': route.path !== '/settings'
          }
        ]"
      >
        <div :class="[
          'w-8 h-8 flex items-center justify-center rounded-lg transition-all duration-300',
          {
            'bg-gradient-to-br from-primary-500 to-accent-500 shadow-neon-blue': route.path === '/settings',
            'bg-slate-600/50 group-hover:bg-slate-500/50': route.path !== '/settings'
          }
        ]">
          <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          </svg>
        </div>
        <span v-if="!collapsed" :class="[
          'font-medium transition-colors duration-300',
          {
            'text-white': route.path === '/settings',
            'text-slate-300 group-hover:text-white': route.path !== '/settings'
          }
        ]">Settings</span>
      </router-link>
      
      <!-- User Section -->
      <div
        :class="[
          'flex items-center gap-3 p-3 rounded-xl glass-subtle hover:glass-medium transition-all duration-300 cursor-pointer group'
        ]"
      >
        <div class="w-8 h-8 bg-gradient-to-br from-slate-600 to-slate-700 rounded-lg flex items-center justify-center">
          <svg class="w-5 h-5 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
          </svg>
        </div>
        <div v-if="!collapsed" class="flex-1">
          <div class="text-sm font-medium text-white">Administrator</div>
          <div class="text-xs text-slate-400">admin@virtumancer.dev</div>
        </div>
        <div v-if="!collapsed" class="text-slate-400 group-hover:text-white transition-colors duration-200">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"/>
          </svg>
        </div>
      </div>
    </div>

    <!-- Collapse Toggle -->
    <div class="absolute -right-4 top-8 z-20">
      <button
        :class="[
          'w-8 h-8 bg-gradient-to-br from-primary-500 to-accent-500 rounded-full',
          'flex items-center justify-center text-white shadow-lg hover:shadow-xl',
          'transition-all duration-300 hover:scale-110 group border-2 border-white/20',
          'hover:border-white/40 backdrop-blur-sm'
        ]"
        @click="toggleCollapse"
        :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      >
        <svg
          :class="[
            'w-4 h-4 transition-transform duration-300',
            { 'rotate-180': collapsed }
          ]"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
        </svg>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUpdated, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useHostStore } from '@/stores/hostStore';

interface NavigationItem {
  id: string;
  label: string;
  icon: string;
  active: boolean;
  path: string;
  badge?: number;
  description?: string;
  badgeColor?: string;
  requiresHostId?: boolean;
  expanded?: boolean;
  children?: NavigationChild[];
}

interface NavigationChild {
  id: string;
  label: string;
  icon: string;
  path: string;
  active: boolean;
  badge?: number;
  badgeColor?: string;
}

interface Props {
  collapsed?: boolean;
}

interface Emits {
  (e: 'update:collapsed', value: boolean): void;
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
});

const emit = defineEmits<Emits>();

const router = useRouter();
const route = useRoute();
const hostStore = useHostStore();

const collapsed = ref(props.collapsed);

const navigationItems = ref<NavigationItem[]>([
  // Main Navigation
  {
    id: 'home',
    label: 'Home',
    icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
    active: false,
    path: '/',
    expanded: false
  },
  {
    id: 'vms',
    label: 'Virtual Machines',
    description: 'Browse all virtual machines',
    icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
    active: false,
    path: '/vms',
    expanded: false,
    children: [
      {
        id: 'managed-vms',
        label: 'Managed VMs',
        icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
        path: '/vms/managed',
        active: false
      },
      {
        id: 'discovered-vms',
        label: 'Discovered VMs',
        icon: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z',
        path: '/vms/discovered',
        active: false
      }
    ]
  },
  {
    id: 'hosts',
    label: 'Hosts',
    description: 'Manage virtualization hosts',
    icon: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01',
    active: false,
    path: '/hosts',
    expanded: false,
    children: []
  },
  {
    id: 'network',
    label: 'Networks',
    description: 'Network infrastructure',
    icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
    active: false,
    path: '/network',
    expanded: false,
    children: [
      {
        id: 'networks-list',
        label: 'Networks',
        icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
        path: '/network/networks',
        active: false
      },
      {
        id: 'network-ports',
        label: 'Ports',
        icon: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2z',
        path: '/network/ports',
        active: false
      },
      {
        id: 'network-topology',
        label: 'Topology',
        icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z',
        path: '/network/topology',
        active: false
      }
    ]
  },
  {
    id: 'storage',
    label: 'Storage',
    description: 'Storage pools and volumes',
    icon: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4',
    active: false,
    path: '/storage',
    expanded: false,
    children: [
      {
        id: 'storage-pools',
        label: 'Storage Pools',
        icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
        path: '/storage/pools',
        active: false
      },
      {
        id: 'storage-volumes',
        label: 'Volumes',
        icon: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',
        path: '/storage/volumes',
        active: false
      }
    ]
  }
]);

// Update active state based on current route
const updateActiveState = () => {
  // Update hosts children with current hosts - only if changed
  const hostsItem = navigationItems.value.find(item => item.id === 'hosts');
  if (hostsItem) {
    const newChildren = hostStore.hosts.map(host => ({
      id: `host-${host.id}`,
      label: host.name || host.uri || 'Unknown Host',
      icon: host.state === 'CONNECTED' ? 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' : 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.35 15.5c-.77.833.192 2.5 1.732 2.5z',
      path: `/hosts/${host.id}`,
      active: false
    }));
    
    // Only update if the children actually changed
    const childrenChanged = !hostsItem.children || 
      hostsItem.children.length !== newChildren.length ||
      hostsItem.children.some((child, index) => 
        !newChildren[index] || 
        child.id !== newChildren[index].id || 
        child.label !== newChildren[index].label ||
        child.icon !== newChildren[index].icon
      );
    
    if (childrenChanged) {
      hostsItem.children = newChildren;
    }
  }
  
  navigationItems.value.forEach(item => {
    // Check if main item is active (exact match only)
    item.active = route.path === item.path;

    // Check children for active state
    if (item.children) {
      item.children.forEach(child => {
        child.active = route.path === child.path;
        if (child.active) {
          // Auto-expand parent if a child is active, but do NOT mark parent as active
          item.expanded = true;
        }
      });
    }
  });
};

// Handle navigation item click
const handleNavItemClick = (item: NavigationItem) => {
  // Always navigate to the overview page first
  router.push(item.path);

  // If item has children, expand it (but never collapse - only collapse icon controls collapse)
  if (item.children && item.children.length > 0) {
    item.expanded = true;
  }
};

// Watch for route changes
onMounted(() => {
  updateActiveState();
});

// Remove onUpdated to prevent recursive updates
// onUpdated(() => {
//   updateActiveState();
// });

watch(route, () => {
  updateActiveState();
});

// Remove redundant watcher - route watcher above already handles this
// watch(() => route.path, () => {
//   updateActiveState();
// }, { immediate: true });

const sidebarClasses = computed(() => [
  collapsed.value ? 'w-20' : 'w-72',
  'fixed left-0 top-0 h-full z-50'
]);

const toggleCollapse = () => {
  collapsed.value = !collapsed.value;
  emit('update:collapsed', collapsed.value);
};

// Handle navigation to host dashboard with dynamic host selection
const handleHostNavigation = () => {
  const hosts = hostStore.hosts;
  if (hosts.length > 0 && hosts[0]?.id) {
    // Navigate to the first available host
    router.push(`/hosts/${hosts[0].id}`);
  } else {
    // If no hosts available, stay on current page or redirect to home
    console.warn('No hosts available for navigation');
    // Could also show a toast notification here
  }
};
</script>

<style scoped>
.animate-slideDown {
  animation: slideDown 0.3s ease-out forwards;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
    max-height: 0;
  }
  to {
    opacity: 1;
    transform: translateY(0);
    max-height: 200px;
  }
}
</style>