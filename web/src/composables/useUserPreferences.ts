import { ref, watch } from 'vue'

export type ViewMode = 'cards' | 'compact' | 'list'

export interface VMListPreferences {
  viewMode: ViewMode
  sortBy: string
  sortDirection: 'asc' | 'desc'
}

export interface UserPreferences {
  vmList: VMListPreferences
}

const defaultPreferences: UserPreferences = {
  vmList: {
    viewMode: 'cards',
    sortBy: 'name',
    sortDirection: 'asc'
  }
}

const PREFERENCES_KEY = 'virtumancer_user_preferences'

export function useUserPreferences() {
  // Load preferences from localStorage or use defaults
  const loadPreferences = (): UserPreferences => {
    try {
      const stored = localStorage.getItem(PREFERENCES_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        // Merge with defaults to ensure all keys exist
        return {
          vmList: {
            ...defaultPreferences.vmList,
            ...parsed.vmList
          }
        }
      }
    } catch (error) {
      console.warn('Failed to load user preferences:', error)
    }
    return { ...defaultPreferences }
  }

  // Save preferences to localStorage
  const savePreferences = (preferences: UserPreferences) => {
    try {
      localStorage.setItem(PREFERENCES_KEY, JSON.stringify(preferences))
    } catch (error) {
      console.warn('Failed to save user preferences:', error)
    }
  }

  // Reactive preferences state
  const preferences = ref<UserPreferences>(loadPreferences())

  // Watch for changes and auto-save
  watch(preferences, (newPreferences) => {
    savePreferences(newPreferences)
  }, { deep: true })

  // Helper methods for VM list preferences
  const vmListPreferences = {
    get viewMode() {
      return preferences.value.vmList.viewMode
    },
    set viewMode(value: ViewMode) {
      preferences.value.vmList.viewMode = value
    },
    get sortBy() {
      return preferences.value.vmList.sortBy
    },
    set sortBy(value: string) {
      preferences.value.vmList.sortBy = value
    },
    get sortDirection() {
      return preferences.value.vmList.sortDirection
    },
    set sortDirection(value: 'asc' | 'desc') {
      preferences.value.vmList.sortDirection = value
    }
  }

  // Reset to defaults
  const resetPreferences = () => {
    preferences.value = { ...defaultPreferences }
  }

  return {
    preferences,
    vmListPreferences,
    resetPreferences,
    defaultPreferences
  }
}