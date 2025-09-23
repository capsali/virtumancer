// Store exports for easy importing
export { useHostStore } from './hostStore';
export { useVMStore } from './vmStore';
export { useUIStore } from './uiStore';
export { useAppStore } from './appStore';

// Re-export types for convenience
export type * from '@/types';

// Utility function to initialize all stores
export const initializeStores = async () => {
  const { useAppStore } = await import('./appStore');
  const appStore = useAppStore();
  
  return appStore.initialize();
};