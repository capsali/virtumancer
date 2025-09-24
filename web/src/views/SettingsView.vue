<template>
  <div class="space-y-6">
    <!-- Breadcrumbs -->
    <FBreadcrumbs />
    
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Settings</h1>
        <p class="text-slate-400 mt-2">Configure application preferences and system settings</p>
      </div>
    </div>

    <!-- Settings Sections -->
    <div class="space-y-6">
      <!-- Metrics Settings -->
      <FCard class="p-6 card-glow">
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">Metrics Settings</h3>
              <p class="text-sm text-slate-400">Configure how performance metrics are displayed</p>
            </div>
            <FButton
              variant="outline"
              size="sm"
              @click="showMetricsModal = true"
            >
              ⚙️ Configure
            </FButton>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">CPU Display Mode</span>
              <span class="text-sm font-medium text-white">{{ settingsStore.cpuDisplayDefault === 'guest' ? 'Guest Usage' : 'Host Usage' }}</span>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">Memory Display</span>
              <span class="text-sm font-medium text-white">Percentage</span>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Theme Settings -->
      <FCard class="p-6 card-glow">
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">Appearance</h3>
            <p class="text-sm text-slate-400">Customize the application theme and interface</p>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Theme</span>
              <select
                v-model="currentTheme"
                @change="handleThemeChange"
                class="px-3 py-1 bg-slate-800/50 border border-slate-600/50 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/50"
              >
                <option value="dark">Dark</option>
                <option value="light">Light</option>
              </select>
            </div>
          </div>
        </div>
      </FCard>

      <!-- View Preferences -->
      <FCard class="p-6 card-glow">
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">View Preferences</h3>
            <p class="text-sm text-slate-400">Default view and sorting settings for virtual machines</p>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Default View Mode</span>
              <select
                v-model="userPrefs.vmListPreferences.viewMode"
                class="px-3 py-1 bg-slate-800/50 border border-slate-600/50 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/50"
              >
                <option value="grid">Grid (Cards)</option>
                <option value="list">List (Table)</option>
                <option value="compact">Compact</option>
              </select>
            </div>
            
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Default Sort Column</span>
              <select
                v-model="userPrefs.vmListPreferences.sortBy"
                class="px-3 py-1 bg-slate-800/50 border border-slate-600/50 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/50"
              >
                <option value="name">Name</option>
                <option value="host">Host</option>
                <option value="status">Status</option>
                <option value="vcpus">vCPUs</option>
                <option value="memory">Memory</option>
              </select>
            </div>
            
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-300">Default Sort Direction</span>
              <select
                v-model="userPrefs.vmListPreferences.sortDirection"
                class="px-3 py-1 bg-slate-800/50 border border-slate-600/50 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/50"
              >
                <option value="asc">Ascending</option>
                <option value="desc">Descending</option>
              </select>
            </div>
            
            <div class="flex justify-end">
              <FButton
                variant="ghost"
                size="sm"
                @click="resetViewPreferences"
                class="text-slate-400 hover:text-white"
              >
                Reset to Defaults
              </FButton>
            </div>
          </div>
        </div>
      </FCard>

      <!-- System Information -->
      <FCard class="p-6 card-glow">
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">System Information</h3>
            <p class="text-sm text-slate-400">Application and system details</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">Version</span>
              <span class="text-sm font-medium text-white">1.0.0</span>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">Build Date</span>
              <span class="text-sm font-medium text-white">{{ new Date().toLocaleDateString() }}</span>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">Hosts Connected</span>
              <span class="text-sm font-medium text-white">{{ connectedHosts }}</span>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-800/30 rounded-lg">
              <span class="text-sm text-slate-300">Active VMs</span>
              <span class="text-sm font-medium text-white">{{ activeVMs }}</span>
            </div>
          </div>
        </div>
      </FCard>

      <!-- Danger Zone -->
      <FCard class="p-6 border-red-500/20 card-glow">
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-red-400">Danger Zone</h3>
            <p class="text-sm text-slate-400">Irreversible and destructive actions</p>
          </div>

          <div class="flex gap-3">
            <FButton
              variant="ghost"
              class="text-red-400 hover:text-red-300 hover:bg-red-500/10"
              @click="handleResetSettings"
            >
              🔄 Reset All Settings
            </FButton>
            <FButton
              variant="ghost"
              class="text-red-400 hover:text-red-300 hover:bg-red-500/10"
              @click="handleClearCache"
            >
              🗑️ Clear Cache
            </FButton>
          </div>
        </div>
      </FCard>
    </div>

    <!-- Metrics Settings Modal -->
    <MetricSettingsModal
      :show="showMetricsModal"
      @close="showMetricsModal = false"
      @applied="showMetricsModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useHostStore } from '@/stores/hostStore'
import { useVMStore } from '@/stores/vmStore'
import { useSettingsStore } from '@/stores/settingsStore'
import { useTheme } from '@/composables/useTheme'
import { useUserPreferences } from '@/composables/useUserPreferences'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import FBreadcrumbs from '@/components/ui/FBreadcrumbs.vue'
import MetricSettingsModal from '@/components/modals/MetricSettingsModal.vue'

const hostStore = useHostStore()
const vmStore = useVMStore()
const settingsStore = useSettingsStore()
const { themeConfig, setTheme } = useTheme()
const userPrefs = useUserPreferences()

// Reactive data
const showMetricsModal = ref(false)
const currentTheme = ref(themeConfig.value.mode)

// Computed properties
const connectedHosts = computed(() => hostStore.connectedHosts.length)
const activeVMs = computed(() => vmStore.activeVMs.length)

// Methods
const handleThemeChange = () => {
  setTheme(currentTheme.value)
}

const handleResetSettings = () => {
  if (confirm('Are you sure you want to reset all settings to defaults? This action cannot be undone.')) {
    // Reset settings logic would go here
    console.log('Resetting settings...')
  }
}

const handleClearCache = () => {
  if (confirm('Are you sure you want to clear the application cache?')) {
    // Clear cache logic would go here
    localStorage.clear()
    console.log('Cache cleared')
  }
}

const resetViewPreferences = () => {
  if (confirm('Reset view preferences to defaults?')) {
    userPrefs.resetPreferences()
  }
}

// Lifecycle
onMounted(async () => {
  await hostStore.fetchHosts()
})
</script>