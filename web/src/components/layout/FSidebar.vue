<template>
  <nav
    :class="[
      'glass-strong backdrop-blur-xl border-r border-white/10',
      'transition-all duration-300 flex flex-col',
      sidebarClasses
    ]"
  >
    <!  {
    id: 'network',
    label: 'Network Topology',
    icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
    active: false,
    path: '/network'
  },
  {
    id: 'logs',
    label: 'System Logs',
    description: 'View application logs',
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    active: false,
    path: '/logs'
  },
  {
    id: 'error-demo',ection -->
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
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon"/>
            </svg>
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
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';

interface NavigationItem {
  id: string;
  label: string;
  icon: string;
  active: boolean;
  path: string;
  badge?: number;
  description?: string;
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

const collapsed = ref(props.collapsed);

const navigationItems = ref<NavigationItem[]>([
  {
    id: 'home',
    label: 'Home',
    icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
    active: false,
    path: '/'
  },
  {
    id: 'network',
    label: 'Network Topology',
    icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
    active: false,
    path: '/network'
  },
  {
    id: 'error-demo',
    label: 'Error Demo',
    icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.268 16.5c-.77.833.192 2.5 1.732 2.5z',
    active: false,
    path: '/error-demo'
  }
]);

// Update active state based on current route
const updateActiveState = () => {
  navigationItems.value.forEach(item => {
    item.active = route.path === item.path || 
                  (item.path !== '/' && route.path.startsWith(item.path));
  });
};

// Watch for route changes
onMounted(() => {
  updateActiveState();
});

// Watch route changes to update active state
watch(() => route.path, () => {
  updateActiveState();
}, { immediate: true });

const sidebarClasses = computed(() => [
  collapsed.value ? 'w-20' : 'w-72',
  'fixed left-0 top-0 h-full z-50'
]);

const toggleCollapse = () => {
  collapsed.value = !collapsed.value;
  emit('update:collapsed', collapsed.value);
};

const handleNavigation = (item: NavigationItem) => {
  // Navigate using Vue Router
  router.push(item.path).catch(err => {
    // Handle navigation errors gracefully
    if (err.name !== 'NavigationDuplicated') {
      console.error('Navigation error:', err);
    }
  });
};
</script>