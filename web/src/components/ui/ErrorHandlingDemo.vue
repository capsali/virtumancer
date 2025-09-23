<template>
  <div class="p-6 space-y-4">
    <h2 class="text-2xl font-bold text-white mb-4">Error Handling Demo</h2>
    
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- Network Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibold text-white mb-2">Network Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate a network connection error</p>
        <FButton @click="triggerNetworkError" variant="secondary" size="sm">
          Trigger Network Error
        </FButton>
      </FCard>

      <!-- API Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibold text-white mb-2">API Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate a server API error</p>
        <FButton @click="triggerApiError" variant="secondary" size="sm">
          Trigger API Error
        </FButton>
      </FCard>

      <!-- Validation Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibold text-white mb-2">Validation Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate invalid input data</p>
        <FButton @click="triggerValidationError" variant="secondary" size="sm">
          Trigger Validation Error
        </FButton>
      </FCard>

      <!-- Connection Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibold text-white mb-2">Connection Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate host connection failure</p>
        <FButton @click="triggerConnectionError" variant="secondary" size="sm">
          Trigger Connection Error
        </FButton>
      </FCard>

      <!-- Authorization Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibent text-white mb-2">Authorization Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate unauthorized access</p>
        <FButton @click="triggerAuthError" variant="secondary" size="sm">
          Trigger Auth Error
        </FButton>
      </FCard>

      <!-- Critical Error -->
      <FCard class="p-4">
        <h3 class="text-lg font-semibold text-white mb-2">Critical Error</h3>
        <p class="text-gray-400 text-sm mb-3">Simulate critical system error</p>
        <FButton @click="triggerCriticalError" variant="danger" size="sm">
          Trigger Critical Error
        </FButton>
      </FCard>
    </div>

    <!-- Current Errors Display -->
    <div class="mt-8">
      <h3 class="text-xl font-semibold text-white mb-4">Current Errors</h3>
      <div class="space-y-2">
        <div v-if="currentErrors.length === 0" class="text-gray-400">
          No active errors
        </div>
        <div v-else>
          <div v-for="error in currentErrors" :key="error.id" 
               class="p-3 bg-gray-800 rounded border-l-4"
               :class="{
                 'border-yellow-500': error.severity === 'low',
                 'border-orange-500': error.severity === 'medium', 
                 'border-red-500': error.severity === 'high',
                 'border-purple-500': error.severity === 'critical'
               }">
            <div class="flex justify-between items-start">
              <div>
                <div class="text-white font-medium">{{ error.message }}</div>
                <div class="text-gray-400 text-sm">{{ error.code }} • {{ error.operation }}</div>
                <div class="text-gray-500 text-xs">{{ formatTime(error.timestamp) }}</div>
              </div>
              <div class="flex items-center space-x-2">
                <span class="px-2 py-1 text-xs rounded"
                      :class="{
                        'bg-yellow-500/20 text-yellow-400': error.severity === 'low',
                        'bg-orange-500/20 text-orange-400': error.severity === 'medium',
                        'bg-red-500/20 text-red-400': error.severity === 'high',
                        'bg-purple-500/20 text-purple-400': error.severity === 'critical'
                      }">
                  {{ error.severity }}
                </span>
                <FButton @click="dismissError(error.id)" variant="ghost" size="sm">
                  ✕
                </FButton>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="mt-6 flex space-x-4">
      <FButton @click="clearAllErrors" variant="outline" size="sm">
        Clear All Errors
      </FButton>
      <FButton @click="refreshErrors" variant="outline" size="sm">
        Refresh Error List
      </FButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import FCard from './FCard.vue';
import FButton from './FButton.vue';
import { errorRecoveryService } from '@/services/errorRecovery';
import type { EnhancedError } from '@/services/errorRecovery';
import { ApiError } from '@/services/api';

const currentErrors = ref<EnhancedError[]>([]);
let refreshInterval: number | null = null;

// Mock error triggers
const triggerNetworkError = () => {
  const error = new Error('Network connection failed');
  errorRecoveryService.addError(error, 'demo_network_operation', {
    demo: true,
    type: 'network',
    endpoint: '/api/demo/network'
  });
};

const triggerApiError = () => {
  const error = new ApiError('Internal server error occurred', 500, 'INTERNAL_SERVER_ERROR', {
    details: 'Database connection timeout'
  });
  errorRecoveryService.addError(error, 'demo_api_operation', {
    demo: true,
    type: 'api',
    endpoint: '/api/demo/server'
  });
};

const triggerValidationError = () => {
  const error = new ApiError('Invalid input parameters', 400, 'VALIDATION_ERROR', {
    field: 'hostname',
    reason: 'Invalid hostname format'
  });
  errorRecoveryService.addError(error, 'demo_validation_operation', {
    demo: true,
    type: 'validation',
    field: 'hostname'
  });
};

const triggerConnectionError = () => {
  const error = new Error('Failed to connect to libvirt host');
  errorRecoveryService.addError(error, 'demo_host_connection', {
    demo: true,
    type: 'connection',
    hostId: 'demo-host-1',
    uri: 'qemu+ssh://demo@192.168.1.100/system'
  });
};

const triggerAuthError = () => {
  const error = new ApiError('Authentication required', 401, 'UNAUTHORIZED', {
    reason: 'JWT token expired'
  });
  errorRecoveryService.addError(error, 'demo_auth_operation', {
    demo: true,
    type: 'auth'
  });
};

const triggerCriticalError = () => {
  const error = new Error('System integrity check failed');
  errorRecoveryService.addError(error, 'demo_critical_operation', {
    demo: true,
    type: 'critical',
    severity: 'critical'
  });
};

// Error management
const refreshErrors = () => {
  currentErrors.value = errorRecoveryService.getErrors();
};

const dismissError = (id: string) => {
  errorRecoveryService.dismissError(id);
  refreshErrors();
};

const clearAllErrors = () => {
  errorRecoveryService.clearAllErrors();
  refreshErrors();
};

const formatTime = (timestamp: Date) => {
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(timestamp);
};

onMounted(() => {
  refreshErrors();
  // Refresh errors every 2 seconds to show real-time updates
  refreshInterval = setInterval(refreshErrors, 2000);
});

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>