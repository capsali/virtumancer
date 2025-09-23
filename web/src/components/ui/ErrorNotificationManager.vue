<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-50 space-y-3 max-w-md">
      <TransitionGroup
        name="error-notification"
        tag="div"
        class="space-y-3"
      >
        <div
          v-for="error in visibleErrors"
          :key="error.id"
          :class="[
            'bg-gray-900/95 backdrop-blur-sm border rounded-lg shadow-xl p-4',
            'transition-all duration-300 ease-in-out',
            getSeverityBorderClass(error.severity)
          ]"
        >
          <!-- Error Header -->
          <div class="flex items-start justify-between mb-2">
            <div class="flex items-center gap-2">
              <div :class="[
                'w-3 h-3 rounded-full flex-shrink-0',
                getSeverityColorClass(error.severity)
              ]"></div>
              <h4 class="text-white font-medium text-sm">
                {{ getSeverityLabel(error.severity) }} Error
              </h4>
            </div>
            
            <button
              @click="dismissError(error.id)"
              class="text-gray-400 hover:text-white transition-colors p-1"
            >
              ✕
            </button>
          </div>

          <!-- Error Message -->
          <p class="text-gray-300 text-sm mb-3 leading-relaxed">
            {{ error.message }}
          </p>

          <!-- Error Details (collapsible) -->
          <div v-if="error.code !== 'UNKNOWN_ERROR'" class="mb-3">
            <button
              @click="toggleDetails(error.id)"
              class="text-xs text-gray-400 hover:text-gray-300 transition-colors flex items-center gap-1"
            >
              <span :class="[
                'transition-transform',
                expandedDetails.has(error.id) ? 'rotate-90' : ''
              ]">▶</span>
              Details
            </button>
            
            <div v-if="expandedDetails.has(error.id)" class="mt-2 p-2 bg-black/30 rounded text-xs">
              <div class="space-y-1 text-gray-400 font-mono">
                <div><span class="text-gray-500">Code:</span> {{ error.code }}</div>
                <div><span class="text-gray-500">Time:</span> {{ formatTime(error.timestamp) }}</div>
                <div v-if="error.operation"><span class="text-gray-500">Operation:</span> {{ error.operation }}</div>
                <div v-if="error.retryable && error.retryCount > 0">
                  <span class="text-gray-500">Retries:</span> {{ error.retryCount }}/{{ error.maxRetries }}
                </div>
              </div>
            </div>
          </div>

          <!-- Recovery Actions -->
          <div v-if="error.recoveryActions && error.recoveryActions.length > 0" class="flex gap-2">
            <button
              v-for="action in error.recoveryActions"
              :key="action.label"
              @click="executeAction(action, error.id)"
              :class="[
                'px-3 py-1.5 rounded text-xs font-medium transition-colors',
                action.primary
                  ? 'bg-primary-500 hover:bg-primary-600 text-white'
                  : 'bg-gray-700 hover:bg-gray-600 text-gray-300'
              ]"
            >
              {{ action.label }}
            </button>
          </div>

          <!-- Auto-retry Progress -->
          <div v-if="error.retryable && error.retryCount < error.maxRetries" class="mt-3">
            <div class="flex items-center gap-2 text-xs text-gray-400">
              <div class="w-3 h-3 border-2 border-gray-500 border-t-primary-500 rounded-full animate-spin"></div>
              <span>Auto-retrying in {{ getRetryCountdown(error.id) }}s...</span>
            </div>
            <div class="w-full bg-gray-700 rounded-full h-1 mt-2">
              <div 
                class="bg-primary-500 h-1 rounded-full transition-all duration-1000"
                :style="{ width: `${getRetryProgress(error.id)}%` }"
              ></div>
            </div>
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!-- Connection Status Banner -->
    <div
      v-if="!connectionState.isOnline || !connectionState.apiReachable"
      class="fixed top-0 left-0 right-0 z-40 bg-red-600/90 backdrop-blur-sm border-b border-red-500/50 p-3"
    >
      <div class="max-w-6xl mx-auto flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 bg-red-400 rounded-full animate-pulse"></div>
          <span class="text-white font-medium">
            {{ !connectionState.isOnline ? 'No internet connection' : 'Server unreachable' }}
          </span>
        </div>
        
        <button
          @click="checkConnection"
          class="px-3 py-1 bg-red-500/50 hover:bg-red-500/70 rounded text-white text-sm transition-colors"
        >
          Retry Connection
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useErrorRecovery } from '@/services/errorRecovery';
import type { EnhancedError, RecoveryAction, ErrorSeverity } from '@/services/errorRecovery';

const { getErrors, dismissError, connectionState } = useErrorRecovery();

// Component state
const expandedDetails = ref<Set<string>>(new Set());
const retryCountdowns = ref<Map<string, number>>(new Map());

// Computed properties
const visibleErrors = computed(() => {
  return getErrors()
    .filter(error => !error.dismissed)
    .sort((a, b) => {
      // Sort by severity first, then by timestamp
      const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
      const severityDiff = severityOrder[b.severity] - severityOrder[a.severity];
      if (severityDiff !== 0) return severityDiff;
      
      return b.timestamp.getTime() - a.timestamp.getTime();
    })
    .slice(0, 5); // Limit to 5 visible errors
});

// Methods
const getSeverityBorderClass = (severity: ErrorSeverity): string => {
  const classes = {
    low: 'border-blue-400/30',
    medium: 'border-yellow-400/30',
    high: 'border-orange-400/30',
    critical: 'border-red-400/30'
  };
  return classes[severity];
};

const getSeverityColorClass = (severity: ErrorSeverity): string => {
  const classes = {
    low: 'bg-blue-400',
    medium: 'bg-yellow-400',
    high: 'bg-orange-400',
    critical: 'bg-red-400'
  };
  return classes[severity];
};

const getSeverityLabel = (severity: ErrorSeverity): string => {
  const labels = {
    low: 'Info',
    medium: 'Warning',
    high: 'Error',
    critical: 'Critical'
  };
  return labels[severity];
};

const toggleDetails = (errorId: string): void => {
  if (expandedDetails.value.has(errorId)) {
    expandedDetails.value.delete(errorId);
  } else {
    expandedDetails.value.add(errorId);
  }
};

const executeAction = async (action: RecoveryAction, errorId: string): Promise<void> => {
  try {
    await action.action();
    dismissError(errorId);
  } catch (error) {
    console.error('Recovery action failed:', error);
  }
};

const formatTime = (timestamp: Date): string => {
  return timestamp.toLocaleTimeString();
};

const getRetryCountdown = (errorId: string): number => {
  return retryCountdowns.value.get(errorId) || 0;
};

const getRetryProgress = (errorId: string): number => {
  const countdown = getRetryCountdown(errorId);
  const maxCountdown = 10; // Assume 10 second max countdown
  return ((maxCountdown - countdown) / maxCountdown) * 100;
};

const checkConnection = async (): Promise<void> => {
  // This would trigger a connection check in the error recovery service
  try {
    await fetch('/api/v1/health');
  } catch {
    // Connection check handled by error recovery service
  }
};

// Countdown timers for auto-retry visualization
const updateCountdowns = (): void => {
  visibleErrors.value.forEach(error => {
    if (error.retryable && error.retryCount < error.maxRetries) {
      // Simulate countdown (this would be more accurate with actual retry timing)
      const currentCountdown = retryCountdowns.value.get(error.id) || 10;
      if (currentCountdown > 0) {
        retryCountdowns.value.set(error.id, currentCountdown - 1);
      }
    }
  });
};

// Lifecycle
let countdownInterval: number;

onMounted(() => {
  countdownInterval = setInterval(updateCountdowns, 1000);
});

onUnmounted(() => {
  clearInterval(countdownInterval);
});
</script>

<style scoped>
.error-notification-enter-active,
.error-notification-leave-active {
  transition: all 0.3s ease;
}

.error-notification-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.error-notification-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

.error-notification-move {
  transition: transform 0.3s ease;
}
</style>