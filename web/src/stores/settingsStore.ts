import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  const diskSmoothAlpha = ref(0.3)
  const netSmoothAlpha = ref(0.6)
  const cpuSmoothAlpha = ref(0.3)
  const cpuDisplayDefault = ref<'host'|'guest'|'raw'>('host')
  const units = ref({
    disk: 'kib', // options: 'kib', 'mib'
    network: 'mb', // options: 'kb','mb'
  })

  function setDiskAlpha(v: number) { diskSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setNetAlpha(v: number) { netSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setCpuAlpha(v: number) { cpuSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setCpuDefault(v: 'host'|'guest'|'raw') { cpuDisplayDefault.value = v }
  function setUnits(metric: 'disk'|'network', u: string) { units.value[metric] = u }

  return {
    diskSmoothAlpha,
    netSmoothAlpha,
  cpuSmoothAlpha,
    cpuDisplayDefault,
    units,
    setDiskAlpha,
    setNetAlpha,
  setCpuAlpha,
    setCpuDefault,
    setUnits,
  }
})
