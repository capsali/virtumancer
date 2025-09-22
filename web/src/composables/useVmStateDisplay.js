import { computed } from 'vue';

// VM state display logic based on host connection and state/libvirtState
export function useVmStateDisplay() {

  const getVmDisplayState = (vm, host) => {
    const isHostConnected = host?.state === 'CONNECTED';

    if (isHostConnected) {
      // When host is connected, check for drift between intended and observed state
      if (vm.state !== vm.libvirtState) {
        return {
          status: vm.libvirtState,
          intendedState: vm.state,
          observedState: vm.libvirtState,
          hasDrift: true,
          color: 'red',
          message: `Intended: ${vm.state}, Observed: ${vm.libvirtState}`,
          driftType: 'state-mismatch'
        };
      } else {
        // States match
        return {
          status: vm.state,
          hasDrift: false,
          color: getStateColor(vm.state, vm.task_state),
          message: null
        };
      }
    } else {
      // Host is disconnected
      return {
        status: 'UNKNOWN',
        intendedState: vm.state,
        lastKnownState: vm.state,
        hasDrift: false,
        color: 'yellow',
        message: `Last known: ${vm.state}`,
        driftType: null
      };
    }
  };

  const getStateColor = (state, taskState) => {
    if (taskState) return 'blue'; // Busy/transient state

    switch (state) {
      case 'ACTIVE':
      case 'RUNNING':
        return 'green';
      case 'PAUSED':
      case 'SUSPENDED':
        return 'yellow';
      case 'STOPPED':
      case 'SHUTOFF':
        return 'gray';
      case 'ERROR':
        return 'red';
      default:
        return 'gray';
    }
  };

  const isVmRunning = (vm, host) => {
    const displayState = getVmDisplayState(vm, host);
    return displayState.status === 'ACTIVE' || displayState.status === 'RUNNING';
  };

  return {
    getVmDisplayState,
    isVmRunning
  };
}