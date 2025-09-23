<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <!-- Header -->
    <div class="bg-gray-800 border-b border-gray-700 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-white">System Logs</h1>
          <p class="text-gray-400 mt-1">View all application logs, errors, and system messages</p>
        </div>
        <div class="flex items-center gap-4">
          <!-- Auto-refresh toggle -->
          <label class="flex items-center gap-2 text-sm">
            <input
              v-model="autoRefresh"
              type="checkbox"
              class="rounded bg-gray-700 border-gray-600 text-blue-500 focus:ring-blue-500"
            />
            Auto-refresh
          </label>
          
          <!-- Refresh button -->
          <button
            @click="refreshLogs"
            class="px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            :disabled="isLoading"
          >
            <div v-if="isLoading" class="animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full"></div>
            <span v-else>🔄 Refresh</span>
          </button>
          
          <!-- Clear logs button -->
          <button
            @click="clearLogs"
            class="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
          >
            🗑️ Clear
          </button>
        </div>
      </div>
    </div>

    <!-- Filters and Controls -->
    <div class="bg-gray-800/50 border-b border-gray-700 px-6 py-4">
      <div class="flex flex-wrap items-center gap-4">
        <!-- Log level filter -->
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium text-gray-300">Level:</label>
          <select
            v-model="selectedLevel"
            class="bg-gray-700 border-gray-600 text-white rounded px-3 py-1 text-sm"
          >
            <option value="">All Levels</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
            <option value="debug">Debug</option>
            <option value="success">Success</option>
          </select>
        </div>

        <!-- Source filter -->
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium text-gray-300">Source:</label>
          <select
            v-model="selectedSource"
            class="bg-gray-700 border-gray-600 text-white rounded px-3 py-1 text-sm"
          >
            <option value="">All Sources</option>
            <option value="ui">UI Events</option>
            <option value="api">API Calls</option>
            <option value="websocket">WebSocket</option>
            <option value="store">Store Actions</option>
            <option value="system">System</option>
          </select>
        </div>

        <!-- Search -->
        <div class="flex items-center gap-2 flex-1 max-w-md">
          <label class="text-sm font-medium text-gray-300">Search:</label>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search logs..."
            class="flex-1 bg-gray-700 border-gray-600 text-white rounded px-3 py-1 text-sm"
          />
        </div>

        <!-- Export button -->
        <button
          @click="exportLogs"
          class="px-3 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors text-sm"
        >
          📥 Export
        </button>

        <!-- Stats -->
        <div class="text-sm text-gray-400">
          Total: {{ filteredLogs.length }} / {{ logs.length }} entries
        </div>
      </div>
    </div>

    <!-- Log Content -->
    <div class="flex-1 overflow-hidden">
      <div class="h-full overflow-y-auto p-4">
        <!-- Loading state -->
        <div v-if="isLoading && logs.length === 0" class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
            <p class="text-gray-400">Loading logs...</p>
          </div>
        </div>

        <!-- Empty state -->
        <div v-else-if="filteredLogs.length === 0" class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="text-6xl mb-4">📝</div>
            <p class="text-gray-400 text-lg">No logs found</p>
            <p class="text-gray-500 text-sm mt-2">
              {{ logs.length === 0 ? 'No logs available' : 'Try adjusting your filters' }}
            </p>
          </div>
        </div>

        <!-- Log entries -->
        <div v-else class="space-y-1">
          <div
            v-for="(log, index) in paginatedLogs"
            :key="log.id"
            :class="[
              'p-3 rounded-lg border-l-4 transition-colors hover:bg-gray-800/50',
              getLogLevelClass(log.level),
              { 'bg-gray-800/30': index % 2 === 0 }
            ]"
          >
            <!-- Log header -->
            <div class="flex items-start justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0 flex-1">
                <!-- Level badge -->
                <span :class="[
                  'px-2 py-1 rounded text-xs font-medium uppercase tracking-wide',
                  getLogLevelBadgeClass(log.level)
                ]">
                  {{ log.level }}
                </span>

                <!-- Timestamp -->
                <span class="text-gray-400 text-sm font-mono">
                  {{ formatTimestamp(log.timestamp) }}
                </span>

                <!-- Source -->
                <span v-if="log.source" class="text-gray-500 text-xs px-2 py-1 bg-gray-700 rounded">
                  {{ log.source }}
                </span>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-2">
                <button
                  @click="toggleLogDetails(log.id)"
                  class="text-gray-400 hover:text-white transition-colors text-sm"
                >
                  {{ expandedLogs.has(log.id) ? '▼' : '▶' }} Details
                </button>
                <button
                  @click="copyLogEntry(log)"
                  class="text-gray-400 hover:text-white transition-colors text-sm"
                  title="Copy to clipboard"
                >
                  📋
                </button>
              </div>
            </div>

            <!-- Log message -->
            <div class="mt-2">
              <p class="text-white leading-relaxed" v-html="highlightSearchTerms(log.message)"></p>
            </div>

            <!-- Expanded details -->
            <div v-if="expandedLogs.has(log.id)" class="mt-3 border-t border-gray-700 pt-3">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <!-- Basic info -->
                <div class="space-y-2">
                  <div v-if="log.operation">
                    <span class="text-gray-400">Operation:</span>
                    <span class="text-white ml-2">{{ log.operation }}</span>
                  </div>
                  <div v-if="log.component">
                    <span class="text-gray-400">Component:</span>
                    <span class="text-white ml-2">{{ log.component }}</span>
                  </div>
                  <div v-if="log.userId">
                    <span class="text-gray-400">User:</span>
                    <span class="text-white ml-2">{{ log.userId }}</span>
                  </div>
                  <div v-if="log.duration">
                    <span class="text-gray-400">Duration:</span>
                    <span class="text-white ml-2">{{ log.duration }}ms</span>
                  </div>
                </div>

                <!-- Context/Details -->
                <div v-if="log.context || log.error">
                  <div v-if="log.error" class="mb-3">
                    <span class="text-gray-400 block mb-1">Error Details:</span>
                    <div class="bg-gray-900 p-2 rounded text-red-400 font-mono text-xs overflow-x-auto">
                      <pre>{{ JSON.stringify(log.error, null, 2) }}</pre>
                    </div>
                  </div>
                  <div v-if="log.context">
                    <span class="text-gray-400 block mb-1">Context:</span>
                    <div class="bg-gray-900 p-2 rounded text-gray-300 font-mono text-xs overflow-x-auto">
                      <pre>{{ JSON.stringify(log.context, null, 2) }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Load more button -->
        <div v-if="hasMoreLogs" class="text-center py-4">
          <button
            @click="loadMoreLogs"
            class="px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
          >
            Load More Logs
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { loggingService, type LogEntry } from '@/services/logging';

// State
const router = useRouter();
const logs = computed(() => loggingService.getLogs());
const isLoading = ref(false);
const autoRefresh = ref(true);
const selectedLevel = ref('');
const selectedSource = ref('');
const searchQuery = ref('');
const expandedLogs = ref(new Set<string>());
const currentPage = ref(1);
const logsPerPage = 50;

let refreshInterval: number | null = null;

// Computed
const filteredLogs = computed(() => {
  let filtered = logs.value;

  // Filter by level
  if (selectedLevel.value) {
    filtered = filtered.filter(log => log.level === selectedLevel.value);
  }

  // Filter by source
  if (selectedSource.value) {
    filtered = filtered.filter(log => log.source === selectedSource.value);
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(log =>
      log.message.toLowerCase().includes(query) ||
      log.operation?.toLowerCase().includes(query) ||
      log.component?.toLowerCase().includes(query)
    );
  }

  return filtered.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
});

const paginatedLogs = computed(() => {
  return filteredLogs.value.slice(0, currentPage.value * logsPerPage);
});

const hasMoreLogs = computed(() => {
  return filteredLogs.value.length > currentPage.value * logsPerPage;
});

// Methods
const refreshLogs = async () => {
  isLoading.value = true;
  try {
    // Simulate loading time
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // Add a log entry to indicate refresh
    loggingService.addLog({
      level: 'info',
      message: 'Logs refreshed manually',
      source: 'ui',
      operation: 'refresh_logs'
    });
  } finally {
    isLoading.value = false;
  }
};

const clearLogs = () => {
  loggingService.clearLogs();
  expandedLogs.value.clear();
  currentPage.value = 1;
};

const exportLogs = () => {
  const filter = {
    level: selectedLevel.value || undefined,
    source: selectedSource.value || undefined,
    search: searchQuery.value || undefined
  };
  
  const dataStr = loggingService.exportLogs(filter);
  const dataBlob = new Blob([dataStr], { type: 'application/json' });
  const url = URL.createObjectURL(dataBlob);
  
  const link = document.createElement('a');
  link.href = url;
  link.download = `virtumancer_logs_${new Date().toISOString().split('T')[0]}.json`;
  link.click();
  
  URL.revokeObjectURL(url);
  
  loggingService.addLog({
    level: 'success',
    message: `Exported ${filteredLogs.value.length} log entries`,
    source: 'ui',
    operation: 'export_logs'
  });
};

const loadMoreLogs = () => {
  currentPage.value++;
};

const toggleLogDetails = (logId: string) => {
  if (expandedLogs.value.has(logId)) {
    expandedLogs.value.delete(logId);
  } else {
    expandedLogs.value.add(logId);
  }
};

const copyLogEntry = async (log: LogEntry) => {
  const logText = `[${formatTimestamp(log.timestamp)}] ${log.level.toUpperCase()}: ${log.message}`;
  
  try {
    await navigator.clipboard.writeText(logText);
    loggingService.addLog({
      level: 'success',
      message: 'Log entry copied to clipboard',
      source: 'ui',
      operation: 'copy_log'
    });
  } catch (error) {
    loggingService.addLog({
      level: 'error',
      message: 'Failed to copy log entry to clipboard',
      source: 'ui',
      operation: 'copy_log',
      error
    });
  }
};

const formatTimestamp = (timestamp: Date): string => {
  return timestamp.toLocaleString('en-US', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }) + '.' + timestamp.getMilliseconds().toString().padStart(3, '0');
};

const highlightSearchTerms = (text: string): string => {
  if (!searchQuery.value) return text;
  
  const regex = new RegExp(`(${searchQuery.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
  return text.replace(regex, '<mark class="bg-yellow-400 text-black px-1 rounded">$1</mark>');
};

const getLogLevelClass = (level: string): string => {
  const classes = {
    error: 'border-red-500',
    warning: 'border-yellow-500',
    info: 'border-blue-500',
    debug: 'border-purple-500',
    success: 'border-green-500'
  };
  return classes[level as keyof typeof classes] || 'border-gray-500';
};

const getLogLevelBadgeClass = (level: string): string => {
  const classes = {
    error: 'bg-red-600 text-white',
    warning: 'bg-yellow-600 text-white',
    info: 'bg-blue-600 text-white',
    debug: 'bg-purple-600 text-white',
    success: 'bg-green-600 text-white'
  };
  return classes[level as keyof typeof classes] || 'bg-gray-600 text-white';
};

// Setup console log interceptor
const setupConsoleInterceptor = () => {
  // This is now handled by the logging service
  return null;
};

let originalConsole: any = null;

// Auto-refresh setup
watch(autoRefresh, (enabled) => {
  if (enabled && !refreshInterval) {
    refreshInterval = setInterval(() => {
      // In a real implementation, this would fetch new logs from the backend
    }, 5000);
  } else if (!enabled && refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
});

// Lifecycle
onMounted(() => {
  // Setup console interceptor (now handled by logging service)
  originalConsole = setupConsoleInterceptor();
  
  // Add initial log entry
  loggingService.addLog({
    level: 'info',
    message: 'Log viewer initialized',
    source: 'ui',
    operation: 'init_logs_view'
  });

  // Add some sample logs for demonstration if there aren't many logs yet
  if (logs.value.length < 5) {
    setTimeout(() => {
      loggingService.addLog({
        level: 'success',
        message: 'Application started successfully',
        source: 'system',
        operation: 'app_start',
        duration: 1250
      });
      
      loggingService.addLog({
        level: 'info',
        message: 'WebSocket connection established',
        source: 'websocket',
        operation: 'ws_connect',
        context: { url: 'wss://localhost:8888/ws' }
      });
      
      loggingService.addLog({
        level: 'warning',
        message: 'Host connection timeout, retrying...',
        source: 'api',
        operation: 'host_connect',
        context: { hostId: '7ca4453d-6c65-4919-b039-ca07d94d5a58', attempt: 2 }
      });
    }, 1000);
  }

  // Start auto-refresh if enabled
  if (autoRefresh.value) {
    refreshInterval = setInterval(() => {
      // In a real implementation, this would fetch new logs from backend
    }, 5000);
  }
});

onUnmounted(() => {
  // Console interceptors are managed by the logging service
  
  // Clear auto-refresh interval
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>

<style scoped>
/* Custom scrollbar for the log area */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background-color: #1f2937;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background-color: #4b5563;
  border-radius: 0.25rem;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background-color: #6b7280;
}

/* Ensure proper highlighting */
:deep(mark) {
  background-color: #fbbf24;
  color: #000000;
  padding: 0 0.25rem;
  border-radius: 0.25rem;
}
</style>