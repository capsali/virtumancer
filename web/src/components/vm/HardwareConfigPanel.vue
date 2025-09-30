<template>
  <div class="space-y-4">
    <!-- USB Controller Config -->
    <div v-if="device.type === 'usb-controller'" class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">USB Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="usb3">USB 3.0 (XHCI)</option>
          <option value="usb2">USB 2.0 (EHCI)</option>
          <option value="usb1">USB 1.1 (UHCI)</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Port Count</label>
        <input
          v-model.number="config.ports"
          type="number"
          min="1"
          max="8"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        />
      </div>
    </div>

    <!-- Sound Card Config -->
    <div v-else-if="device.type === 'sound-card'" class="grid grid-cols-1 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Audio Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="hda">Intel HDA (Recommended)</option>
          <option value="ac97">AC97</option>
          <option value="es1370">ES1370</option>
          <option value="sb16">Sound Blaster 16</option>
        </select>
      </div>
    </div>

    <!-- Graphics Card Config -->
    <div v-else-if="device.type === 'graphics-card'" class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Graphics Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="qxl">QXL (Recommended)</option>
          <option value="virtio">VirtIO GPU</option>
          <option value="vga">Standard VGA</option>
          <option value="cirrus">Cirrus Logic</option>
          <option value="vmvga">VMware SVGA</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">VRAM (MB)</label>
        <input
          v-model.number="config.vram"
          type="number"
          min="16"
          max="512"
          step="16"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        />
      </div>
    </div>

    <!-- Serial Port Config -->
    <div v-else-if="device.type === 'serial-port'" class="grid grid-cols-1 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Serial Type</label>
        <select
          v-model="config.type"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="pty">Pseudo Terminal</option>
          <option value="tcp">TCP Socket</option>
          <option value="file">Log to File</option>
          <option value="null">Null Device</option>
        </select>
      </div>
    </div>

    <!-- TPM Config -->
    <div v-else-if="device.type === 'tpm'" class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">TPM Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="tpm-tis">TPM-TIS</option>
          <option value="tpm-crb">TPM-CRB</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">TPM Version</label>
        <select
          v-model="config.version"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="2.0">TPM 2.0 (Recommended)</option>
          <option value="1.2">TPM 1.2</option>
        </select>
      </div>
    </div>

    <!-- Watchdog Config -->
    <div v-else-if="device.type === 'watchdog'" class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Watchdog Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="i6300esb">Intel 6300ESB</option>
          <option value="ib700">IB700</option>
          <option value="diag288">DIAG288</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Action on Timeout</label>
        <select
          v-model="config.action"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="reset">Reset VM</option>
          <option value="shutdown">Shutdown VM</option>
          <option value="poweroff">Power Off VM</option>
          <option value="pause">Pause VM</option>
          <option value="none">No Action</option>
        </select>
      </div>
    </div>

    <!-- RNG Config -->
    <div v-else-if="device.type === 'rng'" class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">RNG Model</label>
        <select
          v-model="config.model"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="virtio">VirtIO RNG</option>
          <option value="default">Default</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-300 mb-1">Backend Source</label>
        <select
          v-model="config.backend"
          class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
        >
          <option value="/dev/random">/dev/random</option>
          <option value="/dev/urandom">/dev/urandom</option>
          <option value="egd">EGD Socket</option>
        </select>
      </div>
    </div>

    <!-- Generic Config for other types -->
    <div v-else class="text-sm text-slate-400">
      <p>Configuration options for this device type will be available in a future update.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, computed } from 'vue'

interface Props {
  device: {
    id: string
    type: string
    config: any
  }
  modelValue: any
}

interface Emits {
  (e: 'update:modelValue', value: any): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// Create a reactive config object that syncs with modelValue
const config = computed({
  get: () => props.modelValue,
  set: (value: any) => emit('update:modelValue', value)
})
</script>