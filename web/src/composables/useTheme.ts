import { ref, computed, watch, onMounted, readonly } from 'vue';

export type ThemeMode = 'light' | 'dark' | 'auto';
export type ColorScheme = 'blue' | 'purple' | 'cyan' | 'neon';

interface ThemeConfig {
  mode: ThemeMode;
  colorScheme: ColorScheme;
  animations: boolean;
  particles: boolean;
  glowEffects: boolean;
  reducedMotion: boolean;
}

// Global theme state
const themeConfig = ref<ThemeConfig>({
  mode: 'dark',
  colorScheme: 'blue',
  animations: true,
  particles: true,
  glowEffects: true,
  reducedMotion: false
});

const isDark = ref(true);
const isTransitioning = ref(false);

export function useTheme() {
  // Computed properties for theme classes
  const themeClasses = computed(() => ({
    'dark': isDark.value,
    'light': !isDark.value,
    [`theme-${themeConfig.value.colorScheme}`]: true,
    'animations-disabled': !themeConfig.value.animations || themeConfig.value.reducedMotion,
    'particles-disabled': !themeConfig.value.particles,
    'glow-disabled': !themeConfig.value.glowEffects,
    'transitioning': isTransitioning.value
  }));

  // Color scheme variables
  const colorSchemes = {
    blue: {
      primary: {
        50: '#eff6ff',
        100: '#dbeafe',
        200: '#bfdbfe',
        300: '#93c5fd',
        400: '#60a5fa',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        800: '#1e40af',
        900: '#1e3a8a',
        950: '#172554'
      },
      accent: {
        50: '#ecfeff',
        100: '#cffafe',
        200: '#a5f3fc',
        300: '#67e8f9',
        400: '#22d3ee',
        500: '#06b6d4',
        600: '#0891b2',
        700: '#0e7490',
        800: '#155e75',
        900: '#164e63',
        950: '#083344'
      }
    },
    purple: {
      primary: {
        50: '#faf5ff',
        100: '#f3e8ff',
        200: '#e9d5ff',
        300: '#d8b4fe',
        400: '#c084fc',
        500: '#a855f7',
        600: '#9333ea',
        700: '#7c3aed',
        800: '#6b21a8',
        900: '#581c87',
        950: '#3b0764'
      },
      accent: {
        50: '#fdf4ff',
        100: '#fae8ff',
        200: '#f5d0fe',
        300: '#f0abfc',
        400: '#e879f9',
        500: '#d946ef',
        600: '#c026d3',
        700: '#a21caf',
        800: '#86198f',
        900: '#701a75',
        950: '#4a044e'
      }
    },
    cyan: {
      primary: {
        50: '#ecfeff',
        100: '#cffafe',
        200: '#a5f3fc',
        300: '#67e8f9',
        400: '#22d3ee',
        500: '#06b6d4',
        600: '#0891b2',
        700: '#0e7490',
        800: '#155e75',
        900: '#164e63',
        950: '#083344'
      },
      accent: {
        50: '#f0fdfa',
        100: '#ccfbf1',
        200: '#99f6e4',
        300: '#5eead4',
        400: '#2dd4bf',
        500: '#14b8a6',
        600: '#0d9488',
        700: '#0f766e',
        800: '#115e59',
        900: '#134e4a',
        950: '#042f2e'
      }
    },
    neon: {
      primary: {
        50: '#fefffe',
        100: '#fdfefc',
        200: '#fbfef9',
        300: '#f7fdf5',
        400: '#f0fbee',
        500: '#e6f9e3',
        600: '#c8f0c0',
        700: '#9ee190',
        800: '#6acc57',
        900: '#4ade80',
        950: '#16a34a'
      },
      accent: {
        50: '#fffeeb',
        100: '#fffcc6',
        200: '#fff888',
        300: '#ffee3a',
        400: '#ffdd00',
        500: '#efc319',
        600: '#d19c0d',
        700: '#a6700e',
        800: '#8a5614',
        900: '#754617',
        950: '#452408'
      }
    }
  };

  // Get current theme colors
  const currentColors = computed(() => colorSchemes[themeConfig.value.colorScheme]);

  // Theme persistence
  const STORAGE_KEY = 'virtumancer-theme';

  const saveTheme = () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(themeConfig.value));
  };

  const loadTheme = () => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        Object.assign(themeConfig.value, parsed);
      }
    } catch (error) {
      console.warn('Failed to load theme from localStorage:', error);
    }
  };

  // System preferences detection
  const detectSystemPreference = () => {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return true; // Default to dark
  };

  const detectReducedMotion = () => {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    }
    return false;
  };

  // Theme switching with smooth transitions
  const setTheme = async (mode: ThemeMode) => {
    if (isTransitioning.value) return;

    isTransitioning.value = true;

    // Add transition class to body
    document.body.classList.add('theme-transitioning');

    themeConfig.value.mode = mode;

    // Determine if we should be in dark mode
    if (mode === 'auto') {
      isDark.value = detectSystemPreference();
    } else {
      isDark.value = mode === 'dark';
    }

    // Apply theme class to document
    document.documentElement.classList.toggle('dark', isDark.value);
    document.documentElement.classList.toggle('light', !isDark.value);

    // Wait for transition to complete
    await new Promise(resolve => setTimeout(resolve, 300));

    // Remove transition class
    document.body.classList.remove('theme-transitioning');
    isTransitioning.value = false;

    saveTheme();
  };

  const setColorScheme = (scheme: ColorScheme) => {
    themeConfig.value.colorScheme = scheme;
    
    // Update CSS custom properties for the new color scheme
    updateCSSProperties();
    saveTheme();
  };

  const updateCSSProperties = () => {
    const colors = currentColors.value;
    const root = document.documentElement;

    // Update CSS custom properties
    Object.entries(colors.primary).forEach(([key, value]) => {
      root.style.setProperty(`--color-primary-${key}`, value);
    });

    Object.entries(colors.accent).forEach(([key, value]) => {
      root.style.setProperty(`--color-accent-${key}`, value);
    });
  };

  const toggleTheme = () => {
    const currentMode = themeConfig.value.mode;
    if (currentMode === 'auto') {
      setTheme(isDark.value ? 'light' : 'dark');
    } else {
      setTheme(currentMode === 'dark' ? 'light' : 'dark');
    }
  };

  const setAnimations = (enabled: boolean) => {
    themeConfig.value.animations = enabled;
    document.documentElement.classList.toggle('animations-disabled', !enabled);
    saveTheme();
  };

  const setParticles = (enabled: boolean) => {
    themeConfig.value.particles = enabled;
    document.documentElement.classList.toggle('particles-disabled', !enabled);
    saveTheme();
  };

  const setGlowEffects = (enabled: boolean) => {
    themeConfig.value.glowEffects = enabled;
    document.documentElement.classList.toggle('glow-disabled', !enabled);
    saveTheme();
  };

  // Initialize theme system
  const initTheme = () => {
    // Load saved preferences
    loadTheme();

    // Detect system preferences
    themeConfig.value.reducedMotion = detectReducedMotion();

    // Set initial theme
    if (themeConfig.value.mode === 'auto') {
      isDark.value = detectSystemPreference();
    } else {
      isDark.value = themeConfig.value.mode === 'dark';
    }

    // Apply initial classes
    document.documentElement.classList.toggle('dark', isDark.value);
    document.documentElement.classList.toggle('light', !isDark.value);
    document.documentElement.classList.toggle('animations-disabled', !themeConfig.value.animations || themeConfig.value.reducedMotion);
    document.documentElement.classList.toggle('particles-disabled', !themeConfig.value.particles);
    document.documentElement.classList.toggle('glow-disabled', !themeConfig.value.glowEffects);

    // Update CSS properties
    updateCSSProperties();

    // Listen for system preference changes
    if (typeof window !== 'undefined') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      mediaQuery.addEventListener('change', (e) => {
        if (themeConfig.value.mode === 'auto') {
          setTheme('auto');
        }
      });

      const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
      motionQuery.addEventListener('change', (e) => {
        themeConfig.value.reducedMotion = e.matches;
        setAnimations(themeConfig.value.animations);
      });
    }
  };

  return {
    // State
    themeConfig: readonly(themeConfig),
    isDark: readonly(isDark),
    isTransitioning: readonly(isTransitioning),
    currentColors: readonly(currentColors),
    themeClasses,

    // Methods
    setTheme,
    setColorScheme,
    toggleTheme,
    setAnimations,
    setParticles,
    setGlowEffects,
    initTheme,

    // Utilities
    detectSystemPreference,
    detectReducedMotion
  };
}

// Auto-initialize on import
let initialized = false;
export function initializeTheme() {
  if (!initialized && typeof window !== 'undefined') {
    const { initTheme } = useTheme();
    initTheme();
    initialized = true;
  }
}