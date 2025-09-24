<template>
  <FCard class="p-6 max-w-lg mx-auto">
    <h3 class="text-lg font-semibold mb-4">Metrics & Smoothing Settings</h3>
    <div class="space-y-4">
      <div>
        <label class="block text-sm text-slate-400">Disk smoothing alpha (0..1)</label>
        <input type="range" min="0" max="1" step="0.05" v-model.number="localDiskAlpha" />
        <div class="text-xs text-slate-500">{{ localDiskAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label class="block text-sm text-slate-400">CPU smoothing alpha (0..1)</label>
        <input type="range" min="0" max="1" step="0.05" v-model.number="localCpuAlpha" />
        <div class="text-xs text-slate-500">{{ localCpuAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label class="block text-sm text-slate-400">Network smoothing alpha (0..1)</label>
        <input type="range" min="0" max="1" step="0.05" v-model.number="localNetAlpha" />
        <div class="text-xs text-slate-500">{{ localNetAlpha.toFixed(2) }}</div>
      </div>

      <div>
        <label class="block text-sm text-slate-400">Default CPU display</label>
        <select v-model="localCpuDefault" class="mt-1 w-full">
          <option value="host">Host CPU</option>
          <option value="guest">Guest CPU</option>
          <option value="raw">Raw %</option>
        </select>
      </div>

      <div>
        <label class="block text-sm text-slate-400">Disk units</label>
        <select v-model="localDiskUnit" class="mt-1 w-full">
          <option value="kib">KiB/s</option>
          <option value="mib">MiB/s</option>
        </select>
      </div>

      <div>
        <label class="block text-sm text-slate-400">Network units</label>
        <select v-model="localNetUnit" class="mt-1 w-full">
          <option value="kb">KB/s</option>
          <option value="mb">MB/s</option>
        </select>
      </div>

      <div class="flex justify-end gap-2">
        <FButton variant="ghost" @click="$emit('close')">Cancel</FButton>
        <FButton variant="primary" @click="apply">Apply</FButton>
      </div>
    </div>
  </FCard>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settingsStore'
import { useUIStore } from '@/stores/uiStore'
import { settingsApi } from '@/services/api'

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
