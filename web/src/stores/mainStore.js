import { defineStore } from 'pinia';
import { ref, watch } from 'vue';

export const useMainStore = defineStore('main', () => {
    // State
    const hosts = ref([]);
    const selectedHostId = ref(null);
    const errorMessage = ref('');
    const isLoading = ref({
        hosts: false,
        vms: false,
        addHost: false,
        vmAction: null,
    });
    
    let ws = null;

    // --- WebSocket and Polling Logic ---

    const connectWebSocket = () => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}/ws`;

        ws = new WebSocket(wsURL);
        ws.onopen = () => console.log('WebSocket for UI updates connected');
        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                if (message.type === 'refresh') {
                    console.log('WebSocket received refresh message');
                    fetchHosts();
                }
            } catch (e) {
                console.error("Failed to parse websocket message", e);
            }
        };
        ws.onclose = () => {
            console.log('WebSocket disconnected. Reconnecting in 5s...');
            setTimeout(connectWebSocket, 5000);
        };
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            ws.close();
        };
    };
    
    const initializeRealtime = () => {
        connectWebSocket();
    };

    // --- Host Actions ---

    const fetchHosts = async () => {
        isLoading.value.hosts = true;
        errorMessage.value = '';
        try {
            const response = await fetch('/api/v1/hosts');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();
            
            const hostPromises = (data || []).map(async host => {
                host.vms = await fetchVmsForHost(host.id);
                return host;
            });

            hosts.value = await Promise.all(hostPromises);

        } catch (error) {
            console.error("Error fetching hosts:", error);
            errorMessage.value = "Failed to fetch hosts.";
        } finally {
            isLoading.value.hosts = false;
        }
    };

    const addHost = async (hostData) => {
        isLoading.value.addHost = true;
        errorMessage.value = '';
        try {
            const response = await fetch('/api/v1/hosts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(hostData),
            });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            await fetchHosts();
        } catch (error) {
            errorMessage.value = `Failed to add host: ${error.message}`;
            console.error(error);
        } finally {
            isLoading.value.addHost = false;
        }
    };
    
    const deleteHost = async (hostId) => {
        errorMessage.value = '';
        try {
            const response = await fetch(`/api/v1/hosts/${hostId}`, { method: 'DELETE' });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            if (selectedHostId.value === hostId) {
                selectedHostId.value = null;
            }
            await fetchHosts();
        } catch (error) {
            errorMessage.value = `Failed to delete host: ${error.message}`;
            console.error(error);
        }
    };

    const selectHost = (hostId) => {
        if (selectedHostId.value !== hostId) {
            selectedHostId.value = hostId;
        }
    };

    // --- VM Actions ---

    const fetchVmsForHost = async (hostId) => {
        if (!hostId) return [];
        isLoading.value.vms = true;
        try {
            const response = await fetch(`/api/v1/hosts/${hostId}/vms`);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json() || [];
        } catch (error) {
            errorMessage.value = `Failed to fetch VMs for ${hostId}.`;
            console.error(error);
            return [];
        } finally {
            isLoading.value.vms = false;
        }
    };

    const performVmAction = async (hostId, vmName, action) => {
        isLoading.value.vmAction = `${vmName}:${action}`;
        errorMessage.value = '';
        try {
            const response = await fetch(`/api/v1/hosts/${hostId}/vms/${vmName}/${action}`, { method: 'POST' });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
        } catch (error) {
            errorMessage.value = `Action '${action}' on VM '${vmName}' failed: ${error.message}`;
            console.error(error);
        } finally {
            isLoading.value.vmAction = null;
        }
    };

    const startVm = (hostId, vmName) => performVmAction(hostId, vmName, 'start');
    const gracefulShutdownVm = (hostId, vmName) => performVmAction(hostId, vmName, 'shutdown');
    const gracefulRebootVm = (hostId, vmName) => performVmAction(hostId, vmName, 'reboot');
    const forceOffVm = (hostId, vmName) => performVmAction(hostId, vmName, 'forceoff');
    const forceResetVm = (hostId, vmName) => performVmAction(hostId, vmName, 'forcereset');

    return {
        hosts,
        selectedHostId,
        errorMessage,
        isLoading,
        initializeRealtime,
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


