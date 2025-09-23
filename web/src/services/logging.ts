import { ref, reactive } from 'vue';
import { useUIStore } from '@/stores/uiStore';

// Types for the logging system
export interface LogEntry {
  id: string;
  timestamp: Date;
  level: 'error' | 'warning' | 'info' | 'debug' | 'success';
  message: string;
  source: 'ui' | 'api' | 'websocket' | 'store' | 'system' | 'user';
  operation?: string;
  component?: string;
  userId?: string;
  duration?: number;
  context?: any;
  error?: any;
  category?: 'network' | 'vm' | 'host' | 'auth' | 'storage' | 'general';
}

export interface LogFilter {
  level?: string;
  source?: string;
  category?: string;
  timeRange?: {
    start: Date;
    end: Date;
  };
  search?: string;
}

export interface LogStats {
  total: number;
  errors: number;
  warnings: number;
  info: number;
  debug: number;
  success: number;
  lastError?: Date;
  lastWarning?: Date;
}

class LoggingService {
  private logs = ref<LogEntry[]>([]);
  private logIdCounter = 1;
  private maxLogs = 1000;
  private uiStore: any = null;

  // Initialize the service
  init() {
    try {
      this.uiStore = useUIStore();
    } catch (error) {
      // UIStore might not be available during initial load
      console.warn('UIStore not available during logging service init');
    }

    // Setup error handlers
    this.setupGlobalErrorHandlers();
    this.setupConsoleInterceptors();
    
    // Log service initialization
    this.addLog({
      level: 'info',
      message: 'Logging service initialized',
      source: 'system',
      operation: 'init',
      component: 'LoggingService'
    });
  }

  // Add a new log entry
  addLog(entry: Omit<LogEntry, 'id' | 'timestamp'>): string {
    const logEntry: LogEntry = {
      id: `log_${this.logIdCounter++}_${Date.now()}`,
      timestamp: new Date(),
      ...entry
    };

    this.logs.value.unshift(logEntry);

    // Keep only the most recent logs to prevent memory issues
    if (this.logs.value.length > this.maxLogs) {
      this.logs.value = this.logs.value.slice(0, this.maxLogs);
    }

    // Also send critical errors to toast notifications
    if (entry.level === 'error' && this.uiStore) {
      try {
        this.uiStore.addToast(entry.message, 'error');
      } catch (e) {
        console.warn('Failed to show error toast:', e);
      }
    }

    return logEntry.id;
  }

  // Get all logs
  getLogs(): LogEntry[] {
    return this.logs.value;
  }

  // Get filtered logs
  getFilteredLogs(filter: LogFilter): LogEntry[] {
    let filtered = this.logs.value;

    if (filter.level) {
      filtered = filtered.filter(log => log.level === filter.level);
    }

    if (filter.source) {
      filtered = filtered.filter(log => log.source === filter.source);
    }

    if (filter.category) {
      filtered = filtered.filter(log => log.category === filter.category);
    }

    if (filter.timeRange) {
      filtered = filtered.filter(log => 
        log.timestamp >= filter.timeRange!.start && 
        log.timestamp <= filter.timeRange!.end
      );
    }

    if (filter.search) {
      const searchTerm = filter.search.toLowerCase();
      filtered = filtered.filter(log =>
        log.message.toLowerCase().includes(searchTerm) ||
        log.operation?.toLowerCase().includes(searchTerm) ||
        log.component?.toLowerCase().includes(searchTerm)
      );
    }

    return filtered;
  }

  // Get log statistics
  getStats(): LogStats {
    const stats: LogStats = {
      total: this.logs.value.length,
      errors: 0,
      warnings: 0,
      info: 0,
      debug: 0,
      success: 0
    };

    for (const log of this.logs.value) {
      switch (log.level) {
        case 'error':
          stats.errors++;
          break;
        case 'warning':
          stats.warnings++;
          break;
        case 'info':
          stats.info++;
          break;
        case 'debug':
          stats.debug++;
          break;
        case 'success':
          stats.success++;
          break;
      }
      
      if (log.level === 'error' && (!stats.lastError || log.timestamp > stats.lastError)) {
        stats.lastError = log.timestamp;
      }
      
      if (log.level === 'warning' && (!stats.lastWarning || log.timestamp > stats.lastWarning)) {
        stats.lastWarning = log.timestamp;
      }
    }

    return stats;
  }

  // Clear all logs
  clearLogs(): void {
    this.logs.value = [];
    this.addLog({
      level: 'info',
      message: 'Log history cleared',
      source: 'system',
      operation: 'clear_logs',
      component: 'LoggingService'
    });
  }

  // Export logs as JSON
  exportLogs(filter?: LogFilter): string {
    const logsToExport = filter ? this.getFilteredLogs(filter) : this.logs.value;
    return JSON.stringify({
      exported: new Date().toISOString(),
      count: logsToExport.length,
      logs: logsToExport
    }, null, 2);
  }

  // Log specific events
  logUserAction(action: string, component?: string, context?: any) {
    this.addLog({
      level: 'info',
      message: `User action: ${action}`,
      source: 'user',
      operation: action,
      component,
      context,
      category: 'general'
    });
  }

  logNetworkEvent(message: string, level: LogEntry['level'] = 'info', context?: any) {
    this.addLog({
      level,
      message,
      source: 'api',
      category: 'network',
      context
    });
  }

  logVMEvent(message: string, vmName?: string, hostId?: string, level: LogEntry['level'] = 'info') {
    this.addLog({
      level,
      message,
      source: 'system',
      category: 'vm',
      context: { vmName, hostId }
    });
  }

  logHostEvent(message: string, hostId?: string, level: LogEntry['level'] = 'info') {
    this.addLog({
      level,
      message,
      source: 'system',
      category: 'host',
      context: { hostId }
    });
  }

  logStoreAction(storeName: string, action: string, context?: any) {
    this.addLog({
      level: 'debug',
      message: `${storeName}: ${action}`,
      source: 'store',
      operation: action,
      component: storeName,
      context,
      category: 'general'
    });
  }

  logError(error: Error | unknown, operation?: string, component?: string, context?: any) {
    let errorMessage = 'Unknown error occurred';
    let errorDetails = error;

    if (error instanceof Error) {
      errorMessage = error.message;
      errorDetails = {
        name: error.name,
        message: error.message,
        stack: error.stack
      };
    } else if (typeof error === 'string') {
      errorMessage = error;
    } else if (error && typeof error === 'object') {
      try {
        errorMessage = JSON.stringify(error);
      } catch {
        errorMessage = 'Error object could not be serialized';
      }
    }

    this.addLog({
      level: 'error',
      message: errorMessage,
      source: 'system',
      operation,
      component,
      context,
      error: errorDetails,
      category: 'general'
    });
  }

  // Setup global error handlers
  private setupGlobalErrorHandlers() {
    // Unhandled errors
    window.addEventListener('error', (event) => {
      this.logError(event.error, 'unhandled_error', 'window', {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno
      });
    });

    // Unhandled promise rejections
    window.addEventListener('unhandledrejection', (event) => {
      this.logError(event.reason, 'unhandled_promise_rejection', 'window', {
        promise: event.promise
      });
    });

    // Network errors (if fetch is available)
    if (typeof window.fetch !== 'undefined') {
      const originalFetch = window.fetch;
      window.fetch = async (...args) => {
        const startTime = Date.now();
        try {
          const response = await originalFetch(...args);
          const duration = Date.now() - startTime;
          
          if (!response.ok) {
            this.addLog({
              level: 'warning',
              message: `HTTP ${response.status}: ${response.statusText}`,
              source: 'api',
              operation: 'fetch',
              duration,
              context: { 
                url: args[0],
                status: response.status,
                statusText: response.statusText
              },
              category: 'network'
            });
          } else {
            this.addLog({
              level: 'debug',
              message: `HTTP ${response.status}: ${args[0]}`,
              source: 'api',
              operation: 'fetch',
              duration,
              context: { 
                url: args[0],
                status: response.status
              },
              category: 'network'
            });
          }
          
          return response;
        } catch (error) {
          const duration = Date.now() - startTime;
          this.addLog({
            level: 'error',
            message: `Network error: ${error instanceof Error ? error.message : 'Unknown error'}`,
            source: 'api',
            operation: 'fetch',
            duration,
            context: { url: args[0] },
            error: error instanceof Error ? {
              name: error.name,
              message: error.message
            } : error,
            category: 'network'
          });
          throw error;
        }
      };
    }
  }

  // Setup console interceptors (optional - can be enabled/disabled)
  private setupConsoleInterceptors() {
    const originalConsole = {
      log: console.log,
      error: console.error,
      warn: console.warn,
      info: console.info,
      debug: console.debug
    };

    // Only intercept in development or when explicitly enabled
    const enableConsoleInterception = import.meta.env.DEV || 
      localStorage.getItem('virtumancer_log_console') === 'true';

    if (!enableConsoleInterception) {
      return;
    }

    console.error = (...args) => {
      originalConsole.error(...args);
      this.addLog({
        level: 'error',
        message: args.join(' '),
        source: 'system',
        operation: 'console_error',
        error: args.length === 1 && typeof args[0] === 'object' ? args[0] : { args }
      });
    };

    console.warn = (...args) => {
      originalConsole.warn(...args);
      this.addLog({
        level: 'warning',
        message: args.join(' '),
        source: 'system',
        operation: 'console_warn'
      });
    };

    console.info = (...args) => {
      originalConsole.info(...args);
      this.addLog({
        level: 'info',
        message: args.join(' '),
        source: 'system',
        operation: 'console_info'
      });
    };

    // Only log debug messages if explicitly enabled
    if (localStorage.getItem('virtumancer_log_debug') === 'true') {
      console.debug = (...args) => {
        originalConsole.debug(...args);
        this.addLog({
          level: 'debug',
          message: args.join(' '),
          source: 'system',
          operation: 'console_debug'
        });
      };
    }
  }
}

// Create and export the singleton instance
export const loggingService = new LoggingService();

// Initialize the service when imported (but handle errors gracefully)
try {
  loggingService.init();
} catch (error) {
  console.warn('Failed to initialize logging service:', error);
}

export default loggingService;