import type { VirtualMachine } from '@/types';

/**
 * Determines the appropriate console type for a VM based on its graphics capabilities
 * @param vm - The virtual machine object
 * @returns The console type ('vnc', 'spice', or null if no console available)
 */
export function getConsoleType(vm: VirtualMachine): 'vnc' | 'spice' | null {
  if (!vm.graphics) {
    return null;
  }
  
  // Prefer SPICE over VNC if both are available
  if (vm.graphics.spice) {
    return 'spice';
  }
  
  if (vm.graphics.vnc) {
    return 'vnc';
  }
  
  return null;
}

/**
 * Gets the console route for a VM based on its console type
 * @param hostId - The host ID
 * @param vmName - The VM name  
 * @param vm - The virtual machine object
 * @returns The route string for the console, or null if no console available
 */
export function getConsoleRoute(hostId: string, vmName: string, vm: VirtualMachine): string | null {
  const consoleType = getConsoleType(vm);
  
  if (!consoleType) {
    return null;
  }
  
  return `/${consoleType}/${hostId}/${vmName}`;
}

/**
 * Gets the console display name
 * @param vm - The virtual machine object  
 * @returns Human readable console type name
 */
export function getConsoleDisplayName(vm: VirtualMachine): string | null {
  const consoleType = getConsoleType(vm);
  
  switch (consoleType) {
    case 'vnc':
      return 'VNC Console';
    case 'spice':
      return 'SPICE Console'; 
    default:
      return null;
  }
}