<template>
  <BaseModal 
    :show="open"
    @close="close"
    title="Add New Host"
    size="md"
    cancel-text="Cancel"
    :confirm-text="isLoading ? 'Adding...' : 'Add Host'"
    confirm-variant="ghost"
    :confirm-disabled="isLoading || !formData.uri.trim()"
    :cancel-disabled="isLoading"
    @confirm="handleSubmit"
  >
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label for="host-uri" class="block text-sm font-medium text-white mb-2">
          Host URI *
        </label>
        <input
          id="host-uri"
          v-model="formData.uri"
          type="text"
          placeholder="qemu+ssh://user@hostname/system"
          class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:border-transparent transition-all"
          tabindex="1"
          required
        />
        <p class="text-xs text-slate-400 mt-1">
          Examples: qemu:///system, qemu+ssh://user@host/system
        </p>
      </div>

      <div class="flex items-center gap-2">
        <input
          v-model="formData.auto_reconnect_disabled"
          type="checkbox"
          id="auto-reconnect"
          tabindex="2"
          class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-slate-400"
        />
        <label for="auto-reconnect" class="text-sm text-white">
          Disable auto-reconnect
        </label>
      </div>

      <!-- Error Display -->
      <div v-if="error" class="p-3 bg-red-500/10 border border-red-400/20 rounded-lg">
        <p class="text-sm text-red-400">{{ error }}</p>
      </div>
    </form>

    <template #confirm-content v-if="isLoading">
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
        Adding...
      </div>
    </template>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { useHostStore } from '@/stores/hostStore';
import type { Host } from '@/types';

interface Props {
  open: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:open': [value: boolean];
  hostAdded: [host: Host];
  close: [];
  submit: [hostData: Omit<Host, 'id' | 'state' | 'createdAt' | 'updatedAt'>];
}>();

// Store
const hostStore = useHostStore();

// Form state
const formData = reactive({
  uri: '',
  auto_reconnect_disabled: false
});

const isLoading = ref(false);
const error = ref<string | null>(null);

// Reset form when modal opens/closes
watch(() => props.open, (newValue) => {
  if (newValue) {
    resetForm();
  }
});

const resetForm = (): void => {
  formData.uri = '';
  formData.auto_reconnect_disabled = false;
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
  if (isLoading.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    // Validate URI format
    if (!formData.uri.includes('://')) {
      throw new Error('Invalid URI format. Must include protocol (e.g., qemu://)');
    }
    
    const hostData = {
      uri: formData.uri.trim(),
      auto_reconnect_disabled: formData.auto_reconnect_disabled,
      state: 'DISCONNECTED' as const,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    // Add the host using the store
    const newHost = await hostStore.addHost(hostData);
    
    // Emit the hostAdded event with the new host
    emit('hostAdded', newHost);
    emit('submit', hostData);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to add host';
  } finally {
    isLoading.value = false;
  }
};

// Modal keyboard navigation is now handled by BaseModal
</script>
