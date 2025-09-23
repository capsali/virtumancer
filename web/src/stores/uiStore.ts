import { defineStore } from 'pinia';
import { ref, computed, readonly, watch } from 'vue';
import type { UIPreferences, ViewState, ToastMessage } from '@/types';

// Toast management
let nextToastId = 1;
const toastTimers: Record<number, ReturnType<typeof setTimeout>> = {};

export const useUIStore = defineStore('ui', () => {
  // State
  const preferences = ref<UIPreferences>({
    sidebarCollapsed: false,
    theme: 'auto',
    colorScheme: 'blue',
    reducedMotion: false,
    particleEffects: true,
    glowEffects: true
  });

  const viewState = ref<ViewState>({
    currentView: 'dashboard',
    breadcrumbs: ['Dashboard'],
    filters: {},
    sortBy: 'name',
    sortOrder: 'asc'
  });

  const toasts = ref<ToastMessage[]>([]);
  const modals = ref({
    addHost: false,
    confirmDialog: false,
    vmCreate: false,
    vmEdit: false,
    hostEdit: false
  });

  const searchQuery = ref('');
  const isLoading = ref(false);
  const sidebarCollapsed = ref(false);

  // Layout states
  const mobileMenuOpen = ref(false);
  const fullscreen = ref(false);
  const splitViewEnabled = ref(false);

  // Computed properties
  const isDarkMode = computed((): boolean => {
    if (preferences.value.theme === 'auto') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return preferences.value.theme === 'dark';
  });

  const currentThemeClass = computed((): string => {
    const base = isDarkMode.value ? 'dark' : 'light';
    return `${base} theme-${preferences.value.colorScheme}`;
  });

  const toastCount = computed((): number => {
    return toasts.value.length;
  });

  const hasActiveModals = computed((): boolean => {
    return Object.values(modals.value).some(isOpen => isOpen);
  });

  const viewConfig = computed(() => {
    return {
      ...viewState.value,
      isFullscreen: fullscreen.value,
      isSplitView: splitViewEnabled.value,
      isMobile: window.innerWidth < 768
    };
  });

  // Actions
  const updatePreferences = (updates: Partial<UIPreferences>): void => {
    preferences.value = { ...preferences.value, ...updates };
    savePreferences();
  };

  const toggleSidebar = (): void => {
    sidebarCollapsed.value = !sidebarCollapsed.value;
    preferences.value.sidebarCollapsed = sidebarCollapsed.value;
    savePreferences();
  };

  const setSidebarCollapsed = (collapsed: boolean): void => {
    sidebarCollapsed.value = collapsed;
    preferences.value.sidebarCollapsed = collapsed;
    savePreferences();
  };

  const setTheme = (theme: 'light' | 'dark' | 'auto'): void => {
    preferences.value.theme = theme;
    savePreferences();
  };

  const setColorScheme = (scheme: 'blue' | 'purple' | 'cyan' | 'neon'): void => {
    preferences.value.colorScheme = scheme;
    savePreferences();
  };

  const toggleParticleEffects = (): void => {
    preferences.value.particleEffects = !preferences.value.particleEffects;
    savePreferences();
  };

  const toggleGlowEffects = (): void => {
    preferences.value.glowEffects = !preferences.value.glowEffects;
    savePreferences();
  };

  // View state management
  const setCurrentView = (view: string, breadcrumbs?: string[]): void => {
    viewState.value.currentView = view;
    if (breadcrumbs) {
      viewState.value.breadcrumbs = breadcrumbs;
    }
  };

  const updateFilters = (filters: Record<string, any>): void => {
    viewState.value.filters = { ...viewState.value.filters, ...filters };
  };

  const clearFilters = (): void => {
    viewState.value.filters = {};
  };

  const setSorting = (sortBy: string, sortOrder: 'asc' | 'desc' = 'asc'): void => {
    viewState.value.sortBy = sortBy;
    viewState.value.sortOrder = sortOrder;
  };

  const toggleSortOrder = (): void => {
    viewState.value.sortOrder = viewState.value.sortOrder === 'asc' ? 'desc' : 'asc';
  };

  // Toast management
  const addToast = (
    message: string, 
    type: 'success' | 'error' | 'warning' | 'info' = 'success', 
    timeout: number = 8000
  ): number => {
    const id = nextToastId++;
    const toast: ToastMessage = { id, message, type, timeout };
    
    toasts.value.push(toast);
    
    if (timeout > 0) {
      toastTimers[id] = setTimeout(() => {
        removeToast(id);
      }, timeout);
    }
    
    return id;
  };

  const removeToast = (id: number): void => {
    if (toastTimers[id]) {
      clearTimeout(toastTimers[id]);
      delete toastTimers[id];
    }
    toasts.value = toasts.value.filter(toast => toast.id !== id);
  };

  const clearAllToasts = (): void => {
    // Clear all timers
    Object.values(toastTimers).forEach(timer => clearTimeout(timer));
    Object.keys(toastTimers).forEach(key => delete toastTimers[Number(key)]);
    
    toasts.value = [];
  };

  // Modal management
  const openModal = (modalName: keyof typeof modals.value): void => {
    modals.value[modalName] = true;
  };

  const closeModal = (modalName: keyof typeof modals.value): void => {
    modals.value[modalName] = false;
  };

  const closeAllModals = (): void => {
    Object.keys(modals.value).forEach(key => {
      modals.value[key as keyof typeof modals.value] = false;
    });
  };

  // Search functionality
  const setSearchQuery = (query: string): void => {
    searchQuery.value = query;
  };

  const clearSearch = (): void => {
    searchQuery.value = '';
  };

  // Layout actions
  const toggleMobileMenu = (): void => {
    mobileMenuOpen.value = !mobileMenuOpen.value;
  };

  const setMobileMenuOpen = (open: boolean): void => {
    mobileMenuOpen.value = open;
  };

  const toggleFullscreen = (): void => {
    fullscreen.value = !fullscreen.value;
    
    if (fullscreen.value) {
      document.documentElement.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
  };

  const toggleSplitView = (): void => {
    splitViewEnabled.value = !splitViewEnabled.value;
  };

  const setLoading = (loading: boolean): void => {
    isLoading.value = loading;
  };

  // Keyboard shortcuts
  const handleKeyboardShortcut = (event: KeyboardEvent): void => {
    const { key, ctrlKey, metaKey, shiftKey } = event;
    const isCmd = ctrlKey || metaKey;

    // Prevent default behavior for handled shortcuts
    let handled = false;

    if (isCmd) {
      switch (key) {
        case 'k':
          // Focus search
          handled = true;
          const searchInput = document.querySelector('input[placeholder*="Search"]') as HTMLInputElement;
          searchInput?.focus();
          break;
          
        case 'b':
          // Toggle sidebar
          handled = true;
          toggleSidebar();
          break;
          
        case 'Enter':
          if (shiftKey) {
            // Toggle fullscreen
            handled = true;
            toggleFullscreen();
          }
          break;
      }
    }

    if (key === 'Escape') {
      // Close modals or mobile menu
      handled = true;
      if (mobileMenuOpen.value) {
        setMobileMenuOpen(false);
      } else if (hasActiveModals.value) {
        closeAllModals();
      }
    }

    if (handled) {
      event.preventDefault();
    }
  };

  // Persistence
  const PREFERENCES_KEY = 'virtumancer_preferences';
  
  const loadPreferences = (): void => {
    try {
      const saved = localStorage.getItem(PREFERENCES_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        preferences.value = { ...preferences.value, ...parsed };
        sidebarCollapsed.value = preferences.value.sidebarCollapsed;
      }
    } catch (error) {
      console.warn('Failed to load UI preferences:', error);
    }
  };

  const savePreferences = (): void => {
    try {
      localStorage.setItem(PREFERENCES_KEY, JSON.stringify(preferences.value));
    } catch (error) {
      console.warn('Failed to save UI preferences:', error);
    }
  };

  // Responsive handling
  const handleResize = (): void => {
    const isMobile = window.innerWidth < 768;
    if (isMobile && !sidebarCollapsed.value) {
      setSidebarCollapsed(true);
    }
  };

  // Initialize keyboard shortcuts and resize handler
  const initializeEventListeners = (): void => {
    document.addEventListener('keydown', handleKeyboardShortcut);
    window.addEventListener('resize', handleResize);
  };

  const cleanupEventListeners = (): void => {
    document.removeEventListener('keydown', handleKeyboardShortcut);
    window.removeEventListener('resize', handleResize);
  };

  // Watch for system theme changes
  watch(
    () => preferences.value.theme,
    (newTheme) => {
      if (newTheme === 'auto') {
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        const handleThemeChange = () => {
          // Trigger reactivity for computed properties
          preferences.value = { ...preferences.value };
        };
        
        mediaQuery.addEventListener('change', handleThemeChange);
        return () => mediaQuery.removeEventListener('change', handleThemeChange);
      }
    },
    { immediate: true }
  );

  // Initialize on store creation
  loadPreferences();

  return {
    // State
    preferences: readonly(preferences),
    viewState: readonly(viewState),
    toasts: readonly(toasts),
    modals: readonly(modals),
    searchQuery,
    isLoading,
    sidebarCollapsed,
    mobileMenuOpen,
    fullscreen,
    splitViewEnabled,
    
    // Computed
    isDarkMode,
    currentThemeClass,
    toastCount,
    hasActiveModals,
    viewConfig,
    
    // Preferences
    updatePreferences,
    toggleSidebar,
    setSidebarCollapsed,
    setTheme,
    setColorScheme,
    toggleParticleEffects,
    toggleGlowEffects,
    
    // View state
    setCurrentView,
    updateFilters,
    clearFilters,
    setSorting,
    toggleSortOrder,
    
    // Toasts
    addToast,
    removeToast,
    clearAllToasts,
    
    // Modals
    openModal,
    closeModal,
    closeAllModals,
    
    // Search
    setSearchQuery,
    clearSearch,
    
    // Layout
    toggleMobileMenu,
    setMobileMenuOpen,
    toggleFullscreen,
    toggleSplitView,
    setLoading,
    
    // Utilities
    loadPreferences,
    savePreferences,
    initializeEventListeners,
    cleanupEventListeners
  };
});