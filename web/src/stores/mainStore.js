import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useMainStore = defineStore('main', () => {
  // --- State ---
  const hosts = ref([]);
  const selectedHostId = ref(null);
  const vms = ref([]);
  const isLoading = ref({
    hosts: false,
    vms: false,
    addHost: false,
    vmAction: null, // Tracks which VM is performing an action, e.g., 'vm-name-start'
  });
  const errorMessage = ref('');

  // --- Host Actions ---

  async function fetchHosts() {
    isLoading.value.hosts = true;
    errorMessage.value = '';
    try {
      const response = await fetch('/api/v1/hosts');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      hosts.value = data || [];
    } catch (error) {
      console.error("Error fetching hosts:", error);
      errorMessage.value = "Failed to fetch hosts. Is the backend running?";
    } finally {
      isLoading.value.hosts = false;
    }
  }

  async function addHost(newHostId, newHostUri) {
    if (!newHostId || !newHostUri) {
      errorMessage.value = "Both Host ID and URI are required.";
      return;
    }
    isLoading.value.addHost = true;
    errorMessage.value = '';
    try {
      const response = await fetch('/api/v1/hosts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: newHostId, uri: newHostUri }),
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP error! status: ${response.status}`);
      }
      await fetchHosts(); // Refresh host list
    } catch (error) {
      console.error("Error adding host:", error);
      errorMessage.value = `Failed to add host: ${error.message}`;
    } finally {
      isLoading.value.addHost = false;
    }
  }

  async function deleteHost(hostId) {
    if (!confirm(`Are you sure you want to delete host "${hostId}"?`)) {
      return;
    }
    errorMessage.value = '';
    try {
      const response = await fetch(`/api/v1/hosts/${hostId}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP error! status: ${response.status}`);
      }
      if (selectedHostId.value === hostId) {
        selectedHostId.value = null;
        vms.value = [];
      }
      await fetchHosts();
    } catch (error) {
      console.error("Error deleting host:", error);
      errorMessage.value = `Failed to delete host: ${error.message}`;
    }
  }

  async function fetchVmsForHost(hostId) {
    isLoading.value.vms = true;
    vms.value = [];
    errorMessage.value = '';
    try {
      const response = await fetch(`/api/v1/hosts/${hostId}/vms`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      vms.value = data || [];
    } catch (error) {
      console.error(`Error fetching VMs for host ${hostId}:`, error);
      errorMessage.value = `Failed to fetch VMs for host ${hostId}.`;
    } finally {
      isLoading.value.vms = false;
    }
  }

  function selectHost(hostId) {
    if (selectedHostId.value === hostId) {
      // Deselect if clicking the same host
      selectedHostId.value = null;
      vms.value = [];
    } else {
      selectedHostId.value = hostId;
      fetchVmsForHost(hostId);
    }
  }

  // --- VM Actions ---
  async function _performVmAction(hostId, vmName, action) {
    const actionKey = `${vmName}-${action}`;
    isLoading.value.vmAction = actionKey;
    errorMessage.value = '';
    try {
      const response = await fetch(`/api/v1/hosts/${hostId}/vms/${encodeURIComponent(vmName)}/${action}`, {
        method: 'POST',
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP error! status: ${response.status}`);
      }
      // Refresh VM list after a short delay to allow state to update
      setTimeout(() => {
        if (selectedHostId.value === hostId) {
          fetchVmsForHost(hostId);
        }
      }, 2000); // 2 second delay
    } catch (error) {
      console.error(`Error performing action ${action} on ${vmName}:`, error);
      errorMessage.value = `Failed to ${action.replace('-', ' ')} VM: ${error.message}`;
    } finally {
      // Clear loading state after a delay as well
      setTimeout(() => {
        if (isLoading.value.vmAction === actionKey) {
            isLoading.value.vmAction = null;
        }
      }, 2500);
    }
  }

  function startVm(hostId, vmName) {
    _performVmAction(hostId, vmName, 'start');
  }

  function gracefulShutdownVm(hostId, vmName) {
    _performVmAction(hostId, vmName, 'graceful-shutdown');
  }

  function gracefulRebootVm(hostId, vmName) {
    _performVmAction(hostId, vmName, 'graceful-reboot');
  }

  function forceOffVm(hostId, vmName) {
    if (confirm(`Are you sure you want to FORCE POWER OFF ${vmName}? This is like unplugging the power cord and may cause data loss.`)) {
        _performVmAction(hostId, vmName, 'force-off');
    }
  }

  function forceResetVm(hostId, vmName) {
     if (confirm(`Are you sure you want to FORCE RESET ${vmName}? This is like pressing the physical reset button and may cause data loss.`)) {
        _performVmAction(hostId, vmName, 'force-reset');
    }
  }

  return {
    hosts,
    selectedHostId,
    vms,
    isLoading,
    errorMessage,
    fetchHosts,
    addHost,
    deleteHost,
    selectHost,
    startVm,
    gracefulShutdownVm,
    gracefulRebootVm,
    forceOffVm,
    forceResetVm,
  };
});


