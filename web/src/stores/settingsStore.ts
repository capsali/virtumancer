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
  // Preview scaling: 'fit' keeps original behavior, 'fill' scales framebuffer to fill preview card
  const previewScale = ref<'fit'|'fill'>('fit')

  function setDiskAlpha(v: number) { diskSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setNetAlpha(v: number) { netSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setCpuAlpha(v: number) { cpuSmoothAlpha.value = Math.max(0, Math.min(1, v)) }
  function setCpuDefault(v: 'host'|'guest'|'raw') { cpuDisplayDefault.value = v }
  function setUnits(metric: 'disk'|'network', u: string) { units.value[metric] = u }
  function setPreviewScale(s: 'fit'|'fill') { previewScale.value = s }

  return {
    diskSmoothAlpha,
    netSmoothAlpha,
  cpuSmoothAlpha,
    cpuDisplayDefault,
    units,
  previewScale,
    setDiskAlpha,
    setNetAlpha,
  setCpuAlpha,
    setCpuDefault,
    setUnits,
  setPreviewScale,
  }
})
