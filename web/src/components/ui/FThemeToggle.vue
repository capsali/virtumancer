<template>
  <div class="relative">
    <!-- Theme Toggle Button -->
    <button
      :class="[
        'relative overflow-hidden rounded-xl p-3 transition-all duration-300',
        'glass-subtle hover:glass-medium group focus:outline-none focus:ring-2 focus:ring-primary-400/50'
      ]"
      @click="toggleDropdown"
    >
      <!-- Icon Container -->
      <div class="relative w-6 h-6">
        <!-- Sun Icon -->
        <Transition
          name="theme-icon"
          mode="out-in"
        >
          <svg
            v-if="!isDark"
            key="sun"
            class="absolute inset-0 w-6 h-6 text-amber-400 transform transition-transform duration-300"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>

          <!-- Moon Icon -->
          <svg
            v-else
            key="moon"
            class="absolute inset-0 w-6 h-6 text-slate-300 transform transition-transform duration-300"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
            />
          </svg>
        </Transition>

        <!-- Glow Effect -->
        <div
          :class="[
            'absolute inset-0 rounded-full transition-all duration-500',
            isDark 
              ? 'bg-slate-400/20 blur-md opacity-0 group-hover:opacity-100' 
              : 'bg-amber-400/20 blur-md opacity-0 group-hover:opacity-100'
          ]"
        ></div>
      </div>

      <!-- Transition Indicator -->
      <div
        v-if="isTransitioning"
        class="absolute inset-0 bg-gradient-to-r from-primary-600/20 to-accent-600/20 animate-pulse rounded-xl"
      ></div>
    </button>

    <!-- Dropdown Panel -->
    <Transition
      name="dropdown"
      @enter="onEnter"
      @leave="onLeave"
    >
      <div
        v-if="showDropdown"
        :class="[
          'absolute top-full right-0 mt-2 w-80 z-50',
          'glass-strong backdrop-blur-xl rounded-2xl border border-white/10',
          'shadow-floating-lg transform origin-top-right'
        ]"
      >
        <!-- Header -->
        <div class="p-4 border-b border-white/10">
          <h3 class="text-lg font-semibold text-white mb-1">Theme Settings</h3>
          <p class="text-sm text-slate-400">Customize your experience</p>
        </div>

        <!-- Theme Mode Selection -->
        <div class="p-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-3">Theme Mode</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="mode in themeModes"
                :key="mode.value"
                :class="[
                  'flex flex-col items-center p-3 rounded-xl transition-all duration-200',
                  themeConfig.mode === mode.value
                    ? 'bg-primary-600/20 border border-primary-400/50 text-primary-400'
                    : 'glass-subtle hover:glass-medium text-slate-400 hover:text-white'
                ]"
                @click="setTheme(mode.value)"
              >
                <div class="w-5 h-5 mb-2 flex items-center justify-center">
                  {{ mode.value === 'light' ? '☀️' : mode.value === 'dark' ? '🌙' : '🖥️' }}
                </div>
                <span class="text-xs font-medium">{{ mode.label }}</span>
              </button>
            </div>
          </div>

          <!-- Color Scheme Selection -->
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-3">Color Scheme</label>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="scheme in colorSchemes"
                :key="scheme.value"
                :class="[
                  'flex flex-col items-center p-2 rounded-lg transition-all duration-200',
                  themeConfig.colorScheme === scheme.value
                    ? 'ring-2 ring-offset-2 ring-offset-slate-900 scale-105'
                    : 'hover:scale-105'
                ]"
                :style="{ '--tw-ring-color': scheme.preview }"
                @click="setColorScheme(scheme.value)"
              >
                <div
                  :class="[
                    'w-6 h-6 rounded-full mb-1',
                    scheme.gradient
                  ]"
                ></div>
                <span class="text-xs text-slate-400">{{ scheme.label }}</span>
              </button>
            </div>
          </div>

          <!-- Effect Toggles -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Animations</span>
              <FToggle
                :model-value="themeConfig.animations && !themeConfig.reducedMotion"
                :disabled="themeConfig.reducedMotion"
                @update:model-value="setAnimations"
              />
            </div>

            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Particle Effects</span>
              <FToggle
                :model-value="themeConfig.particles"
                @update:model-value="setParticles"
              />
            </div>

            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Glow Effects</span>
              <FToggle
                :model-value="themeConfig.glowEffects"
                @update:model-value="setGlowEffects"
              />
            </div>
          </div>

          <!-- System Info -->
          <div v-if="themeConfig.reducedMotion" class="p-3 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <span class="text-xs text-amber-400">Reduced motion detected</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Backdrop -->
    <div
      v-if="showDropdown"
      class="fixed inset-0 z-40"
      @click="closeDropdown"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useTheme } from '../../composables/useTheme';
import FToggle from './FToggle.vue';

const {
  themeConfig,
  isDark,
  isTransitioning,
  setTheme,
  setColorScheme,
  setAnimations,
  setParticles,
  setGlowEffects,
  toggleTheme
} = useTheme();

const showDropdown = ref(false);

// Theme mode options
const themeModes = [
  {
    value: 'light' as const,
    label: 'Light',
    icon: 'div' // Will be replaced with proper icon
  },
  {
    value: 'dark' as const,
    label: 'Dark', 
    icon: 'div' // Will be replaced with proper icon
  },
  {
    value: 'auto' as const,
    label: 'Auto',
    icon: 'div' // Will be replaced with proper icon
  }
];

// Color scheme options
const colorSchemes = [
  {
    value: 'blue' as const,
    label: 'Blue',
    gradient: 'bg-gradient-to-br from-blue-500 to-blue-600',
    preview: '#3b82f6'
  },
  {
    value: 'purple' as const,
    label: 'Purple',
    gradient: 'bg-gradient-to-br from-purple-500 to-purple-600',
    preview: '#a855f7'
  },
  {
    value: 'cyan' as const,
    label: 'Cyan',
    gradient: 'bg-gradient-to-br from-cyan-500 to-cyan-600',
    preview: '#06b6d4'
  },
  {
    value: 'neon' as const,
    label: 'Neon',
    gradient: 'bg-gradient-to-br from-green-400 to-yellow-400',
    preview: '#4ade80'
  }
];

const toggleDropdown = () => {
  if (showDropdown.value) {
    closeDropdown();
  } else {
    showDropdown.value = true;
  }
};

const closeDropdown = () => {
  showDropdown.value = false;
};

// Animation hooks
const onEnter = (el: Element) => {
  const element = el as HTMLElement;
  element.style.opacity = '0';
  element.style.transform = 'scale(0.95) translateY(-10px)';
  
  requestAnimationFrame(() => {
    element.style.transition = 'opacity 0.2s ease, transform 0.2s ease';
    element.style.opacity = '1';
    element.style.transform = 'scale(1) translateY(0)';
  });
};

const onLeave = (el: Element) => {
  const element = el as HTMLElement;
  element.style.transition = 'opacity 0.15s ease, transform 0.15s ease';
  element.style.opacity = '0';
  element.style.transform = 'scale(0.95) translateY(-10px)';
};

// Close dropdown on escape key
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && showDropdown.value) {
    closeDropdown();
  }
};

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<style scoped>
.theme-icon-enter-active,
.theme-icon-leave-active {
  transition: all 0.3s ease;
}

.theme-icon-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.8);
}

.theme-icon-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.8);
}

.dropdown-enter-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-10px);
}
</style>