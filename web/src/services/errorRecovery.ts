import { ref, reactive } from 'vue';
import type { ApiError } from '@/services/api';
import { useHostStore } from '@/stores/hostStore';

// Error severity levels
export type ErrorSeverity = 'low' | 'medium' | 'high' | 'critical';

// Recovery action types
export type RecoveryAction = {
  label: string;
  action: () => Promise<void> | void;
  primary?: boolean;
};

// Enhanced error information
export interface EnhancedError {
  id: string;
  message: string;
  code: string;
  severity: ErrorSeverity;
  timestamp: Date;
  operation?: string;
  retryable: boolean;
  retryCount: number;
  maxRetries: number;
  context?: Record<string, any>;
  recoveryActions?: RecoveryAction[];
  dismissed: boolean;
}

// Connection state
export interface ConnectionState {
  isOnline: boolean;
  apiReachable: boolean;
  websocketConnected: boolean;
  lastError?: Date;
  retryAttempts: number;
}

// Error recovery strategies
class ErrorRecoveryService {
  private errors = reactive<Map<string, EnhancedError>>(new Map());
  private connectionState = reactive<ConnectionState>({
    isOnline: navigator.onLine,
    apiReachable: true,
    websocketConnected: false,
    retryAttempts: 0
  });

  private retryTimeouts = new Map<string, number>();
  private reconnectInterval: number | null = null;

  constructor() {
    this.setupNetworkListeners();
    this.startConnectionMonitoring();
  }

  // Add an error with automatic classification and recovery options
  addError(
    error: unknown,
    operation?: string,
    context?: Record<string, any>
  ): EnhancedError {
    const enhancedError = this.classifyError(error, operation, context);
    
    // Don't add errors for manually disconnected hosts
    if (enhancedError.code === 'HOST_DISCONNECTED' && enhancedError.severity === 'low' && !enhancedError.retryable) {
      // This is a manually disconnected host error, don't show it
      return enhancedError;
    }
    
    this.errors.set(enhancedError.id, enhancedError);

    // Automatically attempt recovery for retryable errors
    if (enhancedError.retryable && enhancedError.retryCount < enhancedError.maxRetries) {
      this.scheduleRetry(enhancedError);
    }

    return enhancedError;
  }

  // Classify error and determine recovery strategy
  private classifyError(
    error: unknown,
    operation?: string,
    context?: Record<string, any>
  ): EnhancedError {
    const id = this.generateErrorId();
    const timestamp = new Date();

    let message = 'An unexpected error occurred';
    let code = 'UNKNOWN_ERROR';
    let severity: ErrorSeverity = 'medium';
    let retryable = false;
    let maxRetries = 0;
    let recoveryActions: RecoveryAction[] = [];

    if (error instanceof Error) {
      message = error.message;
      
      // Check if it's an API error
      const apiError = error as any;
      if (apiError.code) {
        code = apiError.code;
        
        // Classify based on error code
        switch (apiError.code) {
          case 'NETWORK_ERROR':
          case 'SERVICE_UNAVAILABLE':
          case 'TIMEOUT':
            severity = 'high';
            retryable = true;
            maxRetries = 3;
            recoveryActions = this.getNetworkRecoveryActions(operation);
            break;

          case 'HOST_DISCONNECTED': {
            // Check if auto-reconnection is disabled for this host
            const hostStore = useHostStore();
            const host = context?.hostId ? hostStore.hosts.find(h => h.id === context.hostId) : null;
            const autoReconnectDisabled = host?.auto_reconnect_disabled;
            
            if (autoReconnectDisabled) {
              // Host was manually disconnected, don't show reconnection messages
              severity = 'low';
              retryable = false;
              maxRetries = 0;
              recoveryActions = [];
              message = 'The host is disconnected. Connect manually when ready.';
            } else {
              // Host disconnected unexpectedly, show reconnection options but don't auto-retry
              severity = 'medium';
              retryable = false; // Don't auto-retry host operations
              maxRetries = 0;
              recoveryActions = this.getHostRecoveryActions(context?.hostId);
            }
            break;
          }

          case 'VM_BUSY':
          case 'VM_STATE_ERROR':
            severity = 'low';
            retryable = true;
            maxRetries = 2;
            recoveryActions = this.getVMRecoveryActions(context?.vmName, context?.hostId);
            break;

          case 'VALIDATION_ERROR':
          case 'BAD_REQUEST':
            severity = 'low';
            retryable = false;
            break;

          case 'UNAUTHORIZED':
          case 'FORBIDDEN':
            severity = 'high';
            retryable = false;
            recoveryActions = this.getAuthRecoveryActions();
            break;

          case 'LIBVIRT_ERROR':
          case 'DATABASE_ERROR':
            severity = 'high';
            retryable = true;
            maxRetries = 1;
            break;

          default:
            severity = 'medium';
            retryable = true;
            maxRetries = 1;
        }
      }
    }

    return {
      id,
      message: this.humanizeErrorMessage(message, code, context),
      code,
      severity,
      timestamp,
      operation,
      retryable,
      retryCount: 0,
      maxRetries,
      context,
      recoveryActions,
      dismissed: false
    };
  }

  // Generate unique error ID
  private generateErrorId(): string {
    return `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // Convert technical error messages to user-friendly ones
  private humanizeErrorMessage(message: string, code: string, context?: Record<string, any>): string {
    // Check if this is a HOST_DISCONNECTED error with auto_reconnect_disabled
    if (code === 'HOST_DISCONNECTED' && context?.hostId) {
      const hostStore = useHostStore();
      const host = hostStore.hosts.find(h => h.id === context.hostId);
      if (host?.auto_reconnect_disabled) {
        return 'The host is disconnected. Connect manually when ready.';
      }
    }

    const errorMessages: Record<string, string> = {
      'NETWORK_ERROR': 'Unable to connect to the server. Please check your internet connection.',
      'SERVICE_UNAVAILABLE': 'The service is temporarily unavailable. Please try again later.',
      'TIMEOUT': 'The request timed out. Please try again.',
      'HOST_DISCONNECTED': 'The host is disconnected. Attempting to reconnect...',
      'HOST_NOT_FOUND': 'The requested host could not be found.',
      'VM_NOT_FOUND': 'The virtual machine could not be found.',
      'VM_BUSY': 'The virtual machine is busy. Please wait and try again.',
      'VM_STATE_ERROR': 'Invalid operation for the current VM state.',
      'VALIDATION_ERROR': 'Please check your input and try again.',
      'UNAUTHORIZED': 'You are not authorized to perform this action.',
      'FORBIDDEN': 'Access to this resource is forbidden.',
      'LIBVIRT_ERROR': 'A virtualization system error occurred.',
      'DATABASE_ERROR': 'A database error occurred. Please try again.',
      'RATE_LIMIT': 'Too many requests. Please wait before trying again.'
    };

    return errorMessages[code] || message;
  }

  // Get network-related recovery actions
  private getNetworkRecoveryActions(operation?: string): RecoveryAction[] {
    const actions: RecoveryAction[] = [
      {
        label: 'Retry',
        action: () => this.retryOperation(operation),
        primary: true
      },
      {
        label: 'Check Connection',
        action: () => this.checkConnection()
      }
    ];

    return actions;
  }

  // Get host-related recovery actions
  private getHostRecoveryActions(hostId?: string): RecoveryAction[] {
    const actions: RecoveryAction[] = [];

    if (hostId) {
      actions.push({
        label: 'Reconnect Host',
        action: () => this.reconnectHost(hostId),
        primary: true
      });
    }

    actions.push({
      label: 'Refresh Hosts',
      action: () => this.refreshHosts()
    });

    return actions;
  }

  // Get VM-related recovery actions
  private getVMRecoveryActions(vmName?: string, hostId?: string): RecoveryAction[] {
    const actions: RecoveryAction[] = [
      {
        label: 'Retry',
        action: () => this.retryOperation(),
        primary: true
      }
    ];

    if (vmName && hostId) {
      actions.push({
        label: 'Refresh VM State',
        action: () => this.refreshVMState(hostId, vmName)
      });
    }

    return actions;
  }

  // Get authentication-related recovery actions
  private getAuthRecoveryActions(): RecoveryAction[] {
    return [
      {
        label: 'Refresh Page',
        action: () => window.location.reload(),
        primary: true
      }
    ];
  }

  // Schedule automatic retry for an error
  private scheduleRetry(error: EnhancedError): void {
    const delay = this.calculateRetryDelay(error.retryCount);
    
    const timeoutId = setTimeout(async () => {
      try {
        error.retryCount++;
        await this.retryOperation(error.operation);
        this.errors.delete(error.id);
      } catch (retryError) {
        if (error.retryCount < error.maxRetries) {
          this.scheduleRetry(error);
        } else {
          // Max retries reached, convert to manual retry
          error.retryable = false;
          error.recoveryActions = [
            {
              label: 'Retry Manually',
              action: () => this.retryOperation(error.operation),
              primary: true
            }
          ];
        }
      }
    }, delay);

    this.retryTimeouts.set(error.id, timeoutId);
  }

  // Calculate exponential backoff delay
  private calculateRetryDelay(retryCount: number): number {
    const baseDelay = 1000; // 1 second
    const maxDelay = 30000; // 30 seconds
    const delay = Math.min(baseDelay * Math.pow(2, retryCount), maxDelay);
    
    // Add jitter to prevent thundering herd
    return delay + Math.random() * 1000;
  }

  // Retry operation (placeholder - to be implemented by specific services)
  private async retryOperation(operation?: string): Promise<void> {
    // This would be implemented by the specific service that registered the error
    console.log(`Retrying operation: ${operation}`);
  }

  // Connection monitoring
  private startConnectionMonitoring(): void {
    this.reconnectInterval = setInterval(() => {
      this.checkConnection();
    }, 30000); // Check every 30 seconds
  }

  private async checkConnection(): Promise<void> {
    try {
      const response = await fetch('/api/v1/health', {
        method: 'GET',
        cache: 'no-cache'
      });
      
      this.connectionState.apiReachable = response.ok;
      this.connectionState.retryAttempts = 0;
    } catch {
      this.connectionState.apiReachable = false;
      this.connectionState.retryAttempts++;
    }
  }

  // Network event listeners
  private setupNetworkListeners(): void {
    window.addEventListener('online', () => {
      this.connectionState.isOnline = true;
      this.checkConnection();
    });

    window.addEventListener('offline', () => {
      this.connectionState.isOnline = false;
      this.connectionState.apiReachable = false;
    });
  }

  // Recovery action implementations (placeholders)
  private async reconnectHost(hostId: string): Promise<void> {
    // To be implemented by host service
    console.log(`Reconnecting host: ${hostId}`);
  }

  private async refreshHosts(): Promise<void> {
    // To be implemented by host service
    console.log('Refreshing hosts');
  }

  private async refreshVMState(hostId: string, vmName: string): Promise<void> {
    // To be implemented by VM service
    console.log(`Refreshing VM state: ${hostId}/${vmName}`);
  }

  // Public API
  getErrors(): EnhancedError[] {
    return Array.from(this.errors.values()).filter(e => !e.dismissed);
  }

  getErrorById(id: string): EnhancedError | undefined {
    return this.errors.get(id);
  }

  dismissError(id: string): void {
    const error = this.errors.get(id);
    if (error) {
      error.dismissed = true;
      
      // Clear retry timeout if exists
      const timeoutId = this.retryTimeouts.get(id);
      if (timeoutId) {
        clearTimeout(timeoutId);
        this.retryTimeouts.delete(id);
      }
      
      // Remove after a delay to allow for animations
      setTimeout(() => {
        this.errors.delete(id);
      }, 1000);
    }
  }

  clearAllErrors(): void {
    this.errors.clear();
    this.retryTimeouts.forEach(timeoutId => clearTimeout(timeoutId));
    this.retryTimeouts.clear();
  }

  getConnectionState(): ConnectionState {
    return this.connectionState;
  }

  destroy(): void {
    if (this.reconnectInterval) {
      clearInterval(this.reconnectInterval);
    }
    this.retryTimeouts.forEach(timeoutId => clearTimeout(timeoutId));
    this.retryTimeouts.clear();
  }
}

// Global error recovery service instance
export const errorRecoveryService = new ErrorRecoveryService();

// Composable for using error recovery in components
export function useErrorRecovery() {
  return {
    addError: errorRecoveryService.addError.bind(errorRecoveryService),
    getErrors: errorRecoveryService.getErrors.bind(errorRecoveryService),
    dismissError: errorRecoveryService.dismissError.bind(errorRecoveryService),
    clearAllErrors: errorRecoveryService.clearAllErrors.bind(errorRecoveryService),
    connectionState: errorRecoveryService.getConnectionState()
  };
}