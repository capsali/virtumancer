<template>
  <BaseModal 
    :show="props.show"
    @close="$emit('close')"
    @cancel="$emit('close')"
    title="Metrics & Smoothing Settings"
    size="md"
    cancel-text="Cancel"
    confirm-text="Apply"
    confirm-variant="ghost"
    @confirm="apply"
  >
    <div class="space-y-4">
      <div>
        <label for="disk-alpha" class="block text-sm text-slate-400 mb-2">Disk smoothing alpha (0..1)</label>
        <input 
          id="disk-alpha"
          type="range" 
          min="0" 
          max="1" 
          step="0.05" 
          v-model.number="localDiskAlpha"
          tabindex="1"
          class="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer focus:outline-none focus:ring-2 focus:ring-slate-400"
        />
        <div class="text-xs text-slate-500 mt-1">{{ localDiskAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label for="cpu-alpha" class="block text-sm text-slate-400 mb-2">CPU smoothing alpha (0..1)</label>
        <input 
          id="cpu-alpha"
          type="range" 
          min="0" 
          max="1" 
          step="0.05" 
          v-model.number="localCpuAlpha"
          tabindex="2"
          class="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer focus:outline-none focus:ring-2 focus:ring-slate-400"
        />
        <div class="text-xs text-slate-500 mt-1">{{ localCpuAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label for="net-alpha" class="block text-sm text-slate-400 mb-2">Network smoothing alpha (0..1)</label>
        <input 
          id="net-alpha"
          type="range" 
          min="0" 
          max="1" 
          step="0.05" 
          v-model.number="localNetAlpha"
          tabindex="3"
          class="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer focus:outline-none focus:ring-2 focus:ring-slate-400"
        />
        <div class="text-xs text-slate-500 mt-1">{{ localNetAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label for="cpu-display" class="block text-sm text-slate-400 mb-2">Default CPU display</label>
        <select 
          id="cpu-display"
          v-model="localCpuDefault" 
          tabindex="4"
          class="mt-1 w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-slate-400 focus:border-transparent transition-all"
        >
          <option value="host">Host CPU</option>
          <option value="guest">Guest CPU</option>
          <option value="raw">Raw %</option>
        </select>
      </div>

      <div>
        <label for="disk-units" class="block text-sm text-slate-400 mb-2">Disk units</label>
        <select 
          id="disk-units"
          v-model="localDiskUnit" 
          tabindex="5"
          class="mt-1 w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-slate-400 focus:border-transparent transition-all"
        >
          <option value="kib">KiB/s</option>
          <option value="mib">MiB/s</option>
        </select>
      </div>

      <div>
        <label for="net-units" class="block text-sm text-slate-400 mb-2">Network units</label>
        <select 
          id="net-units"
          v-model="localNetUnit" 
          tabindex="6"
          class="mt-1 w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-slate-400 focus:border-transparent transition-all"
        >
          <option value="kb">KB/s</option>
          <option value="mb">MB/s</option>
        </select>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settingsStore'
import { useUIStore } from '@/stores/uiStore'
import { settingsApi } from '@/services/api'
import BaseModal from '@/components/ui/BaseModal.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits(['close','applied'])

const store = useSettingsStore()

const localDiskAlpha = ref(store.diskSmoothAlpha)
const localNetAlpha = ref(store.netSmoothAlpha)
const localCpuAlpha = ref(store.cpuSmoothAlpha ?? 0.3)
const localCpuDefault = ref(store.cpuDisplayDefault)
const localDiskUnit = ref(store.units.disk)
const localNetUnit = ref(store.units.network)

onMounted(async () => {
  try {
    const remote = await settingsApi.getMetrics()
    if (remote) {
      if (typeof remote.diskSmoothAlpha === 'number') localDiskAlpha.value = remote.diskSmoothAlpha
        if (typeof remote.cpuSmoothAlpha === 'number') localCpuAlpha.value = remote.cpuSmoothAlpha
      if (typeof remote.netSmoothAlpha === 'number') localNetAlpha.value = remote.netSmoothAlpha
      if (typeof remote.cpuDisplayDefault === 'string') localCpuDefault.value = remote.cpuDisplayDefault
      if (remote.units) {
        if (remote.units.disk) localDiskUnit.value = remote.units.disk
        if (remote.units.network) localNetUnit.value = remote.units.network
      }
    }
  } catch (e) {
    // ignore - fallback to local store defaults
  }
})

function apply() {
  store.setDiskAlpha(localDiskAlpha.value)
  store.setNetAlpha(localNetAlpha.value)
  store.setCpuAlpha(localCpuAlpha.value)
  store.setCpuDefault(localCpuDefault.value)
  store.setUnits('disk', localDiskUnit.value)
  store.setUnits('network', localNetUnit.value)
  // Persist to backend
  settingsApi.updateMetrics({
    diskSmoothAlpha: localDiskAlpha.value,
    cpuSmoothAlpha: localCpuAlpha.value,
    netSmoothAlpha: localNetAlpha.value,
    cpuDisplayDefault: localCpuDefault.value,
    units: { disk: localDiskUnit.value, network: localNetUnit.value }
  }).then(() => {
    const uiStore = useUIStore()
    uiStore.addToast('Metrics settings saved', 'success', 3000)
    emit('applied')
    emit('close')
  }).catch((err) => {
    const uiStore = useUIStore()
    uiStore.addToast('Failed to save metrics settings', 'error', 5000)
  })
}
</script>
