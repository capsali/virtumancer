<template>
  <FModal :show="show" @close="handleClose" size="full">
    <FCard class="space-y-0 h-full">
      <!-- Header -->
      <div class="border-b border-slate-700/50 p-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-xl font-semibold text-white flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37a1.724 1.724 0 002.572-1.065z"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                </svg>
              </div>
              {{ editMode ? 'Edit Virtual Machine' : 'Create Virtual Machine' }}
            </h3>
            <p class="text-slate-400 mt-1">{{ editMode ? 'Modify VM configuration' : 'Configure VM settings' }}</p>
          </div>
          <FButton
            variant="ghost"
            size="sm"
            @click="handleClose"
            class="text-slate-400 hover:text-white"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </FButton>
        </div>
      </div>

      <!-- Content -->
      <div class="p-6 flex-1 overflow-y-auto">
        <div class="max-w-2xl mx-auto space-y-8">
          <!-- Basic Information -->
          <FCard class="p-6">
            <div class="mb-6">
              <h4 class="text-lg font-semibold text-white mb-2">Basic Information</h4>
              <p class="text-sm text-slate-400">Configure basic VM properties</p>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="block text-sm font-medium text-white mb-2">VM Name</label>
                <FInput
                  v-model="config.name"
                  placeholder="Enter VM name"
                  class="w-full"
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-white mb-2">Operating System</label>
                <select v-model="config.osType" class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all">
                  <option value="linux">Linux</option>
                  <option value="windows">Windows</option>
                  <option value="other">Other</option>
                </select>
              </div>
              
              <div class="md:col-span-2">
                <label class="block text-sm font-medium text-white mb-2">Description</label>
                <textarea
                  v-model="config.description"
                  rows="3"
                  placeholder="Optional description"
                  class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all resize-none"
                ></textarea>
              </div>
            </div>
          </FCard>

          <!-- Hardware Configuration -->
          <FCard class="p-6">
            <div class="mb-6">
              <h4 class="text-lg font-semibold text-white mb-2">Hardware Configuration</h4>
              <p class="text-sm text-slate-400">Configure CPU and memory settings</p>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="block text-sm font-medium text-white mb-2">vCPUs</label>
                <FInput
                  v-model.number="config.vcpus"
                  type="number"
                  min="1"
                  max="32"
                  placeholder="2"
                  class="w-full"
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-white mb-2">Memory (GB)</label>
                <FInput
                  v-model.number="config.memory"
                  type="number"
                  min="1"
                  max="512"
                  placeholder="4"
                  class="w-full"
                />
              </div>
            </div>
          </FCard>

          <!-- Storage Configuration -->
          <FCard class="p-6">
            <div class="mb-6">
              <h4 class="text-lg font-semibold text-white mb-2">Storage Configuration</h4>
              <p class="text-sm text-slate-400">Configure virtual disks</p>
            </div>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-white mb-2">Disk Size (GB)</label>
                <FInput
                  v-model.number="config.diskSize"
                  type="number"
                  min="1"
                  max="1000"
                  placeholder="20"
                  class="w-full"
                />
              </div>
            </div>
          </FCard>
        </div>
      </div>

      <!-- Footer -->
      <div class="border-t border-slate-700/50 p-6">
        <div class="flex justify-between items-center">
          <div class="text-sm text-slate-400">
            {{ editMode ? 'Update the VM configuration' : 'Create a new virtual machine' }}
          </div>
          <div class="flex gap-3">
            <FButton
              variant="ghost"
              @click="handleClose"
            >
              Cancel
            </FButton>
            <FButton
              variant="primary"
              @click="handleSave"
              :disabled="!isConfigValid || loading"
            >
              <span v-if="loading" class="flex items-center gap-2">
                <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                {{ editMode ? 'Updating...' : 'Creating...' }}
              </span>
              <span v-else>
                {{ editMode ? 'Update VM' : 'Create VM' }}
              </span>
            </FButton>
          </div>
        </div>
      </div>
    </FCard>
  </FModal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import FModal from '@/components/ui/FModal.vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FInput from '@/components/ui/FInput.vue';

interface Props {
  show: boolean;
  hostId?: string;
  vmData?: any;
  editMode?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  editMode: false
});

const emit = defineEmits<{
  close: [];
  'vm-created': [vm: any];
  'vm-updated': [vm: any];
}>();

// State
const loading = ref(false);

// Configuration
const config = reactive({
  name: '',
  description: '',
  osType: 'linux',
  vcpus: 2,
  memory: 4,
  diskSize: 20
});

// Validation
const isConfigValid = computed(() => {
  return config.name.trim().length > 0 && 
         config.vcpus > 0 && 
         config.memory > 0 && 
         config.diskSize > 0;
});

// Handlers
const handleClose = () => {
  emit('close');
};

const handleSave = async () => {
  if (!isConfigValid.value || loading.value) return;
  
  loading.value = true;
  
  try {
    if (props.editMode) {
      emit('vm-updated', config);
    } else {
      emit('vm-created', config);
    }
    
    handleClose();
  } catch (error) {
    console.error('Failed to save VM configuration:', error);
  } finally {
    loading.value = false;
  }
};

// Initialize with existing VM data if in edit mode
watch(() => props.vmData, (newData) => {
  if (newData && props.editMode) {
    Object.assign(config, {
      name: newData.name || '',
      description: newData.description || '',
      osType: newData.osType || 'linux',
      vcpus: newData.vcpuCount || 2,
      memory: Math.round((newData.memoryMB || 4096) / 1024),
      diskSize: 20 // Default since we don't have this data
    });
  }
}, { immediate: true });
</script>

<style scoped>
.modal-glow {
  box-shadow: 0 0 50px rgba(59, 130, 246, 0.3);
}

.glass-medium {
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
</style>