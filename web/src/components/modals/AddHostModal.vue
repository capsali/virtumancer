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
        class="relative w-full max-w-md glass-medium border border-white/20"
        border-glow
        glow-color="primary"
      >
        <div class="space-y-6">
          <!-- Header -->
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-semibold text-white">Add New Host</h2>
            <FButton
              size="sm"
              variant="ghost"
              @click="close"
            >
              ✕
            </FButton>
          </div>

          <!-- Form -->
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-white mb-2">
                Host URI *
              </label>
              <input
                v-model="formData.uri"
                type="text"
                placeholder="qemu+ssh://user@hostname/system"
                class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
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
                class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
              />
              <label for="auto-reconnect" class="text-sm text-white">
                Disable auto-reconnect
              </label>
            </div>

            <!-- Actions -->
            <div class="flex gap-3 pt-4">
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
                :disabled="isLoading || !formData.uri.trim()"
                class="flex-1"
              >
                <span v-if="!isLoading">Add Host</span>
                <span v-else class="flex items-center gap-2">
                  <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  Adding...
                </span>
              </FButton>
            </div>
          </form>

          <!-- Error Display -->
          <div v-if="error" class="p-3 bg-red-500/10 border border-red-400/20 rounded-lg">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>
        </div>
      </FCard>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import type { Host } from '@/types';

interface Props {
  open: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  submit: [hostData: Omit<Host, 'id' | 'state' | 'createdAt' | 'updatedAt'>];
}>();

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
      auto_reconnect_disabled: formData.auto_reconnect_disabled
    };
    
    emit('submit', hostData);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to add host';
    isLoading.value = false;
  }
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
