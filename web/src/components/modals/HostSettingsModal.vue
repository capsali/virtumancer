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
        class="relative w-full max-w-lg glass-medium border border-white/10 modal-glow"
      >
        <div class="space-y-6">
          <!-- Header -->
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-semibold text-white">Host Settings</h2>
            <FButton
              size="sm"
              variant="ghost"
              @click="close"
            >
              ✕
            </FButton>
          </div>

          <!-- Host Info -->
          <div v-if="host" class="p-4 bg-white/5 rounded-lg border border-white/10">
            <div class="flex items-center gap-3 mb-3">
              <div :class="[
                'w-3 h-3 rounded-full',
                host.state === 'CONNECTED' ? 'bg-green-400' : 'bg-red-400'
              ]"></div>
              <div>
                <h3 class="font-medium text-white">{{ host.name || host.uri }}</h3>
                <p class="text-sm text-slate-400">{{ host.uri }}</p>
                <p class="text-xs text-slate-500">{{ host.state }}</p>
              </div>
            </div>
          </div>

          <!-- Settings Form -->
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div>
              <label for="host-name-edit" class="block text-sm font-medium text-white mb-2">
                Host Name
              </label>
              <input
                id="host-name-edit"
                v-model="formData.name"
                type="text"
                placeholder="e.g., Production Server, Development VM Host"
                class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:border-transparent transition-all"
                tabindex="1"
              />
              <p class="text-xs text-slate-400 mt-1">
                Friendly name to identify this host
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-white mb-2">
                Auto Reconnect
              </label>
              <div class="flex items-center gap-2">
                <input
                  v-model="formData.auto_reconnect_enabled"
                  type="checkbox"
                  id="auto-reconnect-enabled"
                  class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                />
                <label for="auto-reconnect-enabled" class="text-sm text-white">
                  Enable automatic reconnection on connection loss
                </label>
              </div>
            </div>

            <!-- Connection Actions -->
            <div class="space-y-3">
              <h4 class="text-sm font-medium text-white">Connection Actions</h4>
              
              <div class="flex gap-2">
                <FButton
                  v-if="host?.state === 'DISCONNECTED'"
                  type="button"
                  variant="primary"
                  size="sm"
                  @click="connectHost"
                  :disabled="isLoading"
                >
                  <span v-if="!isConnecting">Connect</span>
                  <span v-else class="flex items-center gap-2">
                    <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    Connecting...
                  </span>
                </FButton>
                
                <FButton
                  v-if="host?.state === 'CONNECTED'"
                  type="button"
                  variant="ghost"
                  size="sm"
                  @click="disconnectHost"
                  :disabled="isLoading"
                >
                  Disconnect
                </FButton>
                
                <FButton
                  type="button"
                  variant="ghost"
                  size="sm"
                  @click="testConnection"
                  :disabled="isLoading"
                >
                  Test Connection
                </FButton>
              </div>
            </div>

            <!-- Danger Zone -->
            <div class="space-y-3 pt-4 border-t border-white/10">
              <h4 class="text-sm font-medium text-red-400">Danger Zone</h4>
              
              <FButton
                type="button"
                variant="ghost"
                size="sm"
                @click="confirmDelete = true"
                :disabled="isLoading"
                class="text-red-400 hover:bg-red-500/10 border-red-400/20"
              >
                Remove Host
              </FButton>
            </div>

            <!-- Form Actions -->
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
                :disabled="isLoading"
                class="flex-1"
              >
                <span v-if="!isLoading">Save Settings</span>
                <span v-else class="flex items-center gap-2">
                  <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  Saving...
                </span>
              </FButton>
            </div>
          </form>

          <!-- Error Display -->
          <div v-if="error" class="p-3 bg-red-500/10 border border-red-400/20 rounded-lg">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>

          <!-- Confirm Delete Modal -->
          <div v-if="confirmDelete" class="absolute inset-0 bg-black/50 flex items-center justify-center rounded-lg">
            <div class="bg-slate-900 p-6 rounded-lg border border-white/20 space-y-4">
              <h3 class="text-lg font-semibold text-white">Confirm Removal</h3>
              <p class="text-sm text-slate-300">
                Are you sure you want to remove this host? This action cannot be undone.
              </p>
              <div class="flex gap-3">
                <FButton
                  variant="ghost"
                  size="sm"
                  @click="confirmDelete = false"
                  class="flex-1"
                >
                  Cancel
                </FButton>
                <FButton
                  variant="primary"
                  size="sm"
                  @click="deleteHost"
                  class="flex-1 bg-red-600 hover:bg-red-700"
                >
                  Remove Host
                </FButton>
              </div>
            </div>
          </div>
        </div>
      </FCard>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import { useHostStore } from '@/stores/hostStore';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import type { Host } from '@/types';

interface Props {
  open: boolean;
  hostId?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  hostUpdated: [host: Host];
  hostDeleted: [hostId: string];
}>();

const hostStore = useHostStore();

// Form state
const formData = reactive({
  name: '',
  auto_reconnect_enabled: true
});

const isLoading = ref(false);
const isConnecting = ref(false);
const error = ref<string | null>(null);
const confirmDelete = ref(false);

// Get host data
const host = computed(() => {
  return props.hostId ? hostStore.getHostById(props.hostId) : null;
});

// Reset form when modal opens/closes
watch(() => props.open, (newValue) => {
  if (newValue && host.value) {
    resetForm();
  }
});

const resetForm = (): void => {
  if (host.value) {
    formData.name = host.value.name || '';
    formData.auto_reconnect_enabled = !host.value.auto_reconnect_disabled;
  }
  error.value = null;
  isLoading.value = false;
  confirmDelete.value = false;
};

const close = (): void => {
  if (!isLoading.value) {
    emit('close');
  }
};

const handleSubmit = async (): Promise<void> => {
  if (isLoading.value || !host.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    const updates = {
      name: formData.name.trim() || undefined,
      auto_reconnect_disabled: !formData.auto_reconnect_enabled
    };
    
    await hostStore.updateHost(host.value.id, updates);
    emit('hostUpdated', host.value);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update host settings';
  } finally {
    isLoading.value = false;
  }
};

const connectHost = async (): Promise<void> => {
  if (!host.value || isConnecting.value) return;
  
  error.value = null;
  isConnecting.value = true;
  
  try {
    await hostStore.connectHost(host.value.id);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to connect to host';
  } finally {
    isConnecting.value = false;
  }
};

const disconnectHost = async (): Promise<void> => {
  if (!host.value || isLoading.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    await hostStore.disconnectHost(host.value.id);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to disconnect from host';
  } finally {
    isLoading.value = false;
  }
};

const testConnection = async (): Promise<void> => {
  if (!host.value || isLoading.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    // Test connection by attempting to fetch host info
    await hostStore.fetchHosts();
    error.value = null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Connection test failed';
  } finally {
    isLoading.value = false;
  }
};

const deleteHost = async (): Promise<void> => {
  if (!host.value || isLoading.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    await hostStore.deleteHost(host.value.id);
    emit('hostDeleted', host.value.id);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete host';
    isLoading.value = false;
  }
};

// Close on Escape key
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape' && props.open) {
    if (confirmDelete.value) {
      confirmDelete.value = false;
    } else {
      close();
    }
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
