<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @click.self="close"
    >
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>
      
      <!-- Modal -->
      <FCard
        class="relative w-full max-w-2xl glass-medium border border-white/20"
        border-glow
        glow-color="primary"
      >
        <div class="space-y-6 max-h-[80vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex items-center justify-between p-6 border-b border-white/10">
            <h2 class="text-xl font-semibold text-white">Create Virtual Machine</h2>
            <FButton
              size="sm"
              variant="ghost"
              @click="close"
            >
              ✕
            </FButton>
          </div>

          <!-- Form -->
          <form @submit.prevent="handleSubmit" class="p-6 space-y-6">
            <!-- Basic Configuration -->
            <div class="space-y-4">
              <h3 class="text-lg font-medium text-white">Basic Configuration</h3>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    VM Name *
                  </label>
                  <input
                    v-model="formData.name"
                    type="text"
                    placeholder="my-virtual-machine"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    OS Type
                  </label>
                  <select
                    v-model="formData.osType"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="linux">Linux</option>
                    <option value="windows">Windows</option>
                    <option value="other">Other</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-white mb-2">
                  Description
                </label>
                <textarea
                  v-model="formData.description"
                  placeholder="Optional description for this VM"
                  rows="2"
                  class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all resize-none"
                ></textarea>
              </div>
            </div>

            <!-- Hardware Configuration -->
            <div class="space-y-4">
              <h3 class="text-lg font-medium text-white">Hardware Configuration</h3>
              
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    CPU Cores *
                  </label>
                  <input
                    v-model.number="formData.vcpuCount"
                    type="number"
                    min="1"
                    max="32"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Memory (MB) *
                  </label>
                  <input
                    v-model.number="formData.memoryMB"
                    type="number"
                    min="512"
                    max="65536"
                    step="512"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                  <p class="text-xs text-slate-400 mt-1">
                    {{ formatBytes(formData.memoryMB * 1024 * 1024) }}
                  </p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Disk Size (GB) *
                  </label>
                  <input
                    v-model.number="formData.diskSizeGB"
                    type="number"
                    min="10"
                    max="1000"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                </div>
              </div>
            </div>

            <!-- Network Configuration -->
            <div class="space-y-4">
              <h3 class="text-lg font-medium text-white">Network Configuration</h3>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Network Interface
                  </label>
                  <select
                    v-model="formData.networkInterface"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="default">Default</option>
                    <option value="bridge">Bridge</option>
                    <option value="nat">NAT</option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Boot Device
                  </label>
                  <select
                    v-model="formData.bootDevice"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="hd">Hard Disk</option>
                    <option value="cdrom">CD-ROM</option>
                    <option value="network">Network (PXE)</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Advanced Options -->
            <div class="space-y-4">
              <h3 class="text-lg font-medium text-white">Advanced Options</h3>
              
              <div class="space-y-3">
                <div class="flex items-center gap-2">
                  <input
                    v-model="formData.autoStart"
                    type="checkbox"
                    id="auto-start"
                    class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                  />
                  <label for="auto-start" class="text-sm text-white">
                    Start VM automatically
                  </label>
                </div>

                <div class="flex items-center gap-2">
                  <input
                    v-model="formData.enableSpice"
                    type="checkbox"
                    id="enable-spice"
                    class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                  />
                  <label for="enable-spice" class="text-sm text-white">
                    Enable SPICE console
                  </label>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex gap-3 pt-4 border-t border-white/10">
              <FButton
                type="button"
                variant="ghost"
                @click="close"
                class="flex-1"
                :disabled="isLoading"
              >
                Cancel
              </FButton>
              <FButton
                type="submit"
                variant="primary"
                :disabled="isLoading || !isFormValid"
                class="flex-1"
              >
                <span v-if="!isLoading">Create VM</span>
                <span v-else class="flex items-center gap-2">
                  <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  Creating...
                </span>
              </FButton>
            </div>
          </form>

          <!-- Error Display -->
          <div v-if="error" class="p-3 mx-6 mb-6 bg-red-500/10 border border-red-400/20 rounded-lg">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>
        </div>
      </FCard>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import { useVMStore } from '@/stores/vmStore';
import type { VirtualMachine } from '@/types';

interface Props {
  open: boolean;
  hostId: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:open': [value: boolean];
  vmCreated: [vm: VirtualMachine];
  close: [];
}>();

// Store
const vmStore = useVMStore();

// Form state
const formData = reactive({
  name: '',
  description: '',
  osType: 'linux',
  vcpuCount: 2,
  memoryMB: 2048,
  diskSizeGB: 20,
  networkInterface: 'default',
  bootDevice: 'hd',
  autoStart: false,
  enableSpice: true
});

const isLoading = ref(false);
const error = ref<string | null>(null);

// Computed properties
const isFormValid = computed(() => {
  return formData.name.trim().length > 0 &&
         formData.vcpuCount > 0 &&
         formData.memoryMB >= 512 &&
         formData.diskSizeGB >= 10;
});

// Reset form when modal opens/closes
watch(() => props.open, (newValue) => {
  if (newValue) {
    resetForm();
  }
});

const resetForm = (): void => {
  formData.name = '';
  formData.description = '';
  formData.osType = 'linux';
  formData.vcpuCount = 2;
  formData.memoryMB = 2048;
  formData.diskSizeGB = 20;
  formData.networkInterface = 'default';
  formData.bootDevice = 'hd';
  formData.autoStart = false;
  formData.enableSpice = true;
  error.value = null;
  isLoading.value = false;
};

const close = (): void => {
  if (!isLoading.value) {
    emit('update:open', false);
    emit('close');
  }
};

const handleSubmit = async (): Promise<void> => {
  if (isLoading.value || !isFormValid.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    // Validate VM name
    if (!formData.name.match(/^[a-zA-Z0-9_-]+$/)) {
      throw new Error('VM name can only contain letters, numbers, underscores, and hyphens');
    }
    
    const vmData: Omit<VirtualMachine, 'uuid' | 'createdAt' | 'updatedAt'> = {
      name: formData.name.trim(),
      description: formData.description.trim() || 'No description',
      osType: formData.osType,
      vcpuCount: formData.vcpuCount,
      memoryMB: formData.memoryMB,
      diskSizeGB: formData.diskSizeGB,
      networkInterface: formData.networkInterface,
      bootDevice: formData.bootDevice,
      cpuModel: 'host',
      source: 'managed',
      syncStatus: 'SYNCED',
      libvirtState: 'STOPPED',
      hostId: props.hostId,
      domainUuid: '',
      title: formData.name.trim(),
      state: 'STOPPED'
    };
    
    // Create the VM using the store
    const newVM = await vmStore.createVM(vmData);
    
    // If auto-start is enabled, start the VM
    if (formData.autoStart) {
      await vmStore.startVM(props.hostId, newVM.name);
    }
    
    // Emit the vmCreated event with the new VM
    emit('vmCreated', newVM);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create VM';
  } finally {
    isLoading.value = false;
  }
};

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Close on Escape key
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape' && props.open) {
    close();
  }
};

// Add/remove event listener
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    document.addEventListener('keydown', handleKeydown);
  } else {
    document.removeEventListener('keydown', handleKeydown);
  }
});
</script>