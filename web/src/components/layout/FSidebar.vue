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
        <div class="w-10 h-10 bg-gradient-to-br from-primary-500 to-accent-500 rounded-xl flex items-center justify-center shadow-neon-blue">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
        </div>
        <div v-if="!collapsed">
          <h1 class="text-xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
            VirtuMancer
          </h1>
          <p class="text-xs text-slate-400">Virtualization Platform</p>
        </div>
      </div>
    </div>

    <!-- Navigation Items -->
    <div class="flex-1 p-4 space-y-2">
      <div
        v-for="item in navigationItems"
        :key="item.id"
        :class="[
          'group relative overflow-hidden rounded-xl transition-all duration-300',
          {
            'bg-gradient-to-r from-primary-600/20 to-accent-600/20 shadow-glow-sm': item.active,
            'hover:bg-white/5': !item.active
          }
        ]"
      >
        <button
          :class="[
            'w-full flex items-center gap-3 p-3 text-left transition-all duration-300',
            {
              'text-white': item.active,
              'text-slate-300 hover:text-white': !item.active
            }
          ]"
          @click="handleNavigation(item)"
        >
          <!-- Icon -->
          <div :class="[
            'w-8 h-8 flex items-center justify-center rounded-lg transition-all duration-300',
            {
              'bg-gradient-to-br from-primary-500 to-accent-500 shadow-neon-blue': item.active,
              'bg-slate-600/50 group-hover:bg-slate-500/50': !item.active
            }
          ]">
            <component :is="item.icon" class="w-5 h-5" />
          </div>

          <!-- Label -->
          <div v-if="!collapsed" class="flex-1">
            <div class="font-medium">{{ item.label }}</div>
            <div v-if="item.description" class="text-xs text-slate-400">{{ item.description }}</div>
          </div>

          <!-- Badge -->
          <div
            v-if="item.badge && !collapsed"
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
            class="absolute right-0 top-1/2 transform -translate-y-1/2 w-1 h-8 bg-gradient-to-b from-primary-400 to-accent-400 rounded-l-full"
          ></div>
        </button>

        <!-- Hover Glow Effect -->
        <div
          v-if="!item.active"
          class="absolute inset-0 bg-gradient-to-r from-primary-600/0 via-primary-600/5 to-accent-600/0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 rounded-xl"
        ></div>
      </div>
    </div>

    <!-- User Section -->
    <div class="p-4 border-t border-white/10">
      <div
        :class="[
          'flex items-center gap-3 p-3 rounded-xl glass-subtle hover:glass-medium transition-all duration-300 cursor-pointer group'
        ]"
      >
        <div class="w-10 h-10 bg-gradient-to-br from-slate-600 to-slate-700 rounded-full flex items-center justify-center">
          <svg class="w-6 h-6 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
    <div class="absolute -right-3 top-6 z-20">
      <button
        :class="[
          'w-6 h-6 bg-gradient-to-br from-primary-600 to-accent-600 rounded-full',
          'flex items-center justify-center text-white shadow-lg hover:shadow-xl',
          'transition-all duration-300 hover:scale-110 group'
        ]"
        @click="toggleCollapse"
      >
        <svg
          :class="[
            'w-3 h-3 transition-transform duration-300',
            { 'rotate-180': collapsed }
          ]"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
        </svg>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

interface NavigationItem {
  id: string;
  label: string;
  description?: string;
  icon: any;
  active: boolean;
  badge?: string;
  badgeColor?: string;
  route?: string;
}

interface Props {
  collapsed?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
});

const emit = defineEmits<{
  'update:collapsed': [value: boolean];
  navigate: [item: NavigationItem];
}>();

const collapsed = ref(props.collapsed);

// Mock navigation items - replace with your actual navigation
const navigationItems = ref<NavigationItem[]>([
  {
    id: 'dashboard',
    label: 'Dashboard',
    description: 'Overview & metrics',
    icon: 'div', // Replace with actual icon component
    active: true,
    route: '/dashboard'
  },
  {
    id: 'vms',
    label: 'Virtual Machines',
    description: 'Manage VMs',
    icon: 'div', // Replace with actual icon component
    active: false,
    badge: '12',
    badgeColor: 'bg-primary-500/20 text-primary-400',
    route: '/vms'
  },
  {
    id: 'hosts',
    label: 'Hosts',
    description: 'Hypervisor hosts',
    icon: 'div', // Replace with actual icon component
    active: false,
    badge: '3',
    badgeColor: 'bg-accent-500/20 text-accent-400',
    route: '/hosts'
  },
  {
    id: 'network',
    label: 'Networks',
    description: 'Network topology',
    icon: 'div', // Replace with actual icon component
    active: false,
    route: '/network'
  },
  {
    id: 'storage',
    label: 'Storage',
    description: 'Storage pools',
    icon: 'div', // Replace with actual icon component
    active: false,
    route: '/storage'
  },
  {
    id: 'settings',
    label: 'Settings',
    description: 'Configuration',
    icon: 'div', // Replace with actual icon component
    active: false,
    route: '/settings'
  }
]);

const sidebarClasses = computed(() => [
  collapsed.value ? 'w-20' : 'w-72',
  'h-screen fixed left-0 top-0 z-30'
]);

const toggleCollapse = () => {
  collapsed.value = !collapsed.value;
  emit('update:collapsed', collapsed.value);
};

const handleNavigation = (item: NavigationItem) => {
  // Update active state
  navigationItems.value.forEach(nav => nav.active = false);
  item.active = true;
  
  emit('navigate', item);
};
</script>