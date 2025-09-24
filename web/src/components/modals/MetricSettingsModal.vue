<template>
  <teleport to="body">
    <transition name="modal-fade" appear>
      <div v-if="props.show" class="fixed inset-0 z-50 flex items-center justify-center p-4" @click.self="$emit('close')">
      <!-- Backdrop -->
      <div class="fixed inset-0 bg-black/60 backdrop-blur-sm" @click="$emit('close')"></div>        <!-- Modal -->
        <FCard class="relative w-full max-w-md glass-medium border border-white/10 modal-glow" @click.stop>
          <div class="space-y-6">
            <!-- Header -->
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold text-white">Metrics & Smoothing Settings</h2>
              <FButton
                size="sm"
                variant="ghost"
                @click="$emit('close')"
              >
                ✕
              </FButton>
            </div>
            
            <!-- Form -->
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

              <!-- Actions -->
              <div class="flex gap-3 pt-4">
                <FButton
                  variant="ghost"
                  @click="$emit('close')"
                  class="flex-1 button-glow cancel"
                >
                  Cancel
                </FButton>
                <FButton
                  variant="primary"
                  @click="apply"
                  class="flex-1 button-glow apply"
                >
                  Apply
                </FButton>
              </div>
            </div>
          </div>
        </FCard>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.modal-fade-enter-active, .modal-fade-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}
.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.9) translateY(-20px);
}
.modal-fade-enter-to, .modal-fade-leave-from {
  opacity: 1;
  transform: scale(1) translateY(0);
}



/* Local styles removed - now using shared glow classes from style.css */
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settingsStore'
import { useUIStore } from '@/stores/uiStore'
import { settingsApi } from '@/services/api'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'

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
