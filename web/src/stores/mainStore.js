import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

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
        vmHardware: false,
        vmReconcile: null,
    });

    const activeVmStats = ref(null);
    const activeVmHardware = ref(null);
    const hostStats = ref({}); // New state for host stats
    const hostConnecting = ref({}); // map hostId -> bool when we're connecting / fetching VMs

    // New state for tracking active subscriptions
    const currentlySubscribedHostId = ref(null);
    const currentlySubscribedVmName = ref(null);

    const totalVms = computed(() => {
        return hosts.value.reduce((total, host) => total + (host.vms ? host.vms.length : 0), 0);
    });

    let ws = null;
    let isWsConnected = false;
    const messageQueue = [];

    // --- WebSocket Logic ---

    function sendMessage(type, payload) {
        const message = { type, payload };
        if (isWsConnected) {
            ws.send(JSON.stringify(message));
        } else {
            messageQueue.push(message);
        }
    }

    const connectWebSocket = () => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}/ws`;

        ws = new WebSocket(wsURL);
        ws.onopen = () => {
            console.log('WebSocket for UI updates connected');
            isWsConnected = true;
            // Send any queued messages
            while (messageQueue.length > 0) {
                const message = messageQueue.shift();
                ws.send(JSON.stringify(message));
            }
            // Re-subscribe to active monitors if any
            if (currentlySubscribedHostId.value) {
                subscribeHostStats(currentlySubscribedHostId.value);
            }
            if (currentlySubscribedVmName.value && selectedHostId.value) { // Need hostId for VM
                subscribeToVmStats(selectedHostId.value, currentlySubscribedVmName.value);
            }
        };
        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                switch (message.type) {
                    case 'hosts-changed':
                        console.log('WebSocket received hosts-changed, refetching all hosts.');
                        fetchHosts();
                        break;
                    case 'host-connection-changed':
                        console.log(`WebSocket received host-connection-changed for ${message.payload.hostId}, connected=${message.payload.connected}`);
                        // Update local host entry if present
                        const idx = hosts.value.findIndex(h => h.id === message.payload.hostId);
                        if (idx !== -1) {
                            const updatedHost = { ...hosts.value[idx], connected: message.payload.connected };
                            hosts.value.splice(idx, 1, updatedHost);
                            // If it just connected, proactively refresh its VMs and info
                            if (message.payload.connected) {
                                (async () => {
                                    const [vms, info] = await Promise.all([fetchVmsForHost(message.payload.hostId), fetchHostInfo(message.payload.hostId)]);
                                    const refreshed = { ...updatedHost, vms, info, connected: true };
                                    const curIdx = hosts.value.findIndex(h => h.id === message.payload.hostId);
                                        if (curIdx !== -1) hosts.value.splice(curIdx, 1, refreshed);
                                        // Clear any connecting indicator
                                        if (hostConnecting.value[message.payload.hostId]) {
                                            delete hostConnecting.value[message.payload.hostId];
                                        }
                                })();
                            }
                        } else if (message.payload.connected) {
                            // If host connected but not present, refetch all hosts
                            fetchHosts();
                        }
                        break;
                    case 'vms-changed':
                        console.log(`WebSocket received vms-changed for host ${message.payload.hostId}, refreshing host data.`);
                        refreshHostData(message.payload.hostId);
                        break;
                    case 'vm-stats-updated':
                        // Directly update the stats ref. The component will check if it's for the current VM.
                        activeVmStats.value = message.payload;
                        break;
                    case 'host-stats-updated':
                        hostStats.value[message.payload.hostId] = message.payload.stats;
                        break;
                    default:
                        console.log('Received unhandled WebSocket message type:', message.type);
                }
            } catch (e) {
                console.error("Failed to parse websocket message", e);
            }
        };
        ws.onclose = () => {
            console.log('WebSocket disconnected. Reconnecting in 5s...');
            isWsConnected = false;
            // Clear current subscriptions as they are no longer active
            currentlySubscribedHostId.value = null;
            currentlySubscribedVmName.value = null;
            setTimeout(connectWebSocket, 5000);
        };
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            isWsConnected = false;
            // Clear current subscriptions on error
            currentlySubscribedHostId.value = null;
            currentlySubscribedVmName.value = null;
            ws.close();
        };
    };

    const initializeRealtime = () => {
        connectWebSocket();
    };

    // --- Host Actions ---

    const refreshHostData = async (hostId) => {
        const hostIndex = hosts.value.findIndex(h => h.id === hostId);
        if (hostIndex === -1) {
            console.warn(`Host ${hostId} not found in state during refresh, performing full fetch.`);
            fetchHosts();
            return;
        }

        // Fetch new data for the specific host
        const [vms, info] = await Promise.all([
            fetchVmsForHost(hostId),
            fetchHostInfo(hostId)
        ]);
        
        // Create a new host object to ensure reactivity
        const updatedHost = {
            ...hosts.value[hostIndex],
            vms,
            info,
        };

        // Replace the old host object with the new one
        hosts.value.splice(hostIndex, 1, updatedHost);
    };


    const fetchHosts = async () => {
        isLoading.value.hosts = true;
        errorMessage.value = '';
        try {
            const controller = new AbortController();
            const to = setTimeout(() => controller.abort(), 7000);
            const response = await fetch('/api/v1/hosts', { signal: controller.signal });
            clearTimeout(to);
            if (!response.ok) {
                const text = await response.text().catch(() => '<no body>');
                throw new Error(`HTTP error! status: ${response.status} body: ${text}`);
            }
            const data = await response.json();
            console.log('[mainStore] fetchHosts: got response', data);
            // API may return either an array of Host objects or objects with
            // an embedded Host + connected flag. Normalize both shapes.
            // Per-host fetch with timeouts so a slow host doesn't block everything.
            const hostFetchTasks = (data || []).map(item => {
                return (async () => {
                    try {
                        let hostObj = item || {};
                        let connected = false;

                        if (hostObj && typeof hostObj === 'object') {
                            if (hostObj.Host && hostObj.Host.id) hostObj = hostObj.Host;
                            if (hostObj.host && hostObj.host.id) hostObj = hostObj.host;
                            if (item.connected !== undefined) connected = !!item.connected;
                            else if (hostObj.connected !== undefined) connected = !!hostObj.connected;
                        }

                        if (!hostObj || !hostObj.id) {
                            console.warn('[mainStore] fetchHosts: skipping invalid host entry', item);
                            return null;
                        }

                        const host = { ...hostObj, connected };

                        // Per-host VM/info fetch with AbortController timeout
                        const vmPromise = fetchVmsForHost(host.id);
                        const infoPromise = fetchHostInfo(host.id);

                        const [vms, info] = await Promise.all([vmPromise, infoPromise]);
                        host.vms = vms || [];
                        host.info = info || null;
                        return host;
                    } catch (err) {
                        console.warn('[mainStore] fetchHosts: host fetch failed', item, err && err.message ? err.message : err);
                        return null;
                    }
                })();
            });

            const settled = await Promise.allSettled(hostFetchTasks);
            const resolved = settled.map(s => (s.status === 'fulfilled' ? s.value : null)).filter(h => h !== null);
            console.log(`[mainStore] fetchHosts: resolved ${resolved.length} hosts`);
            hosts.value = resolved;
            // Clear any stale connecting flags for hosts we now know about
            hosts.value.forEach(h => { if (hostConnecting.value[h.id]) delete hostConnecting.value[h.id]; });

        } catch (error) {
                console.error("Error fetching hosts:", error, error && error.stack ? error.stack : error);
            // Try a quick health check to provide more useful diagnostics
            try {
                const hres = await fetch('/api/v1/health');
                if (hres.ok) {
                    const body = await hres.json();
                    errorMessage.value = `Failed to fetch hosts: ${error.message} (health ok)`;
                } else {
                    const text = await hres.text();
                    errorMessage.value = `Failed to fetch hosts: ${error.message}; health endpoint returned ${hres.status}: ${text}`;
                }
            } catch (herr) {
                console.error('Health check failed:', herr);
                errorMessage.value = `Failed to fetch hosts: ${error.message}; health check failed: ${herr.message}`;
            }
        } finally {
            isLoading.value.hosts = false;
        }
    };

    const fetchHostInfo = async (hostId) => {
        if (!hostId) return null;
        try {
            const controller = new AbortController();
            const id = setTimeout(() => controller.abort(), 5000);
            const response = await fetch(`/api/v1/hosts/${hostId}/info`, { signal: controller.signal });
            clearTimeout(id);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();
            // API returns { connected: bool, info: { ... } } so normalize to return the inner info object.
            if (data && typeof data === 'object' && data.info !== undefined) return data.info;
            return data;
        } catch (error) {
            console.error(`Error fetching info for host ${hostId}:`, error);
            return null;
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
            const newHost = await response.json();

            // Optimistically add the new host to the local state.
            // The websocket will later trigger a full refresh to sync VMs and other info.
            hosts.value.push({
                ...newHost,
                connected: false,
                vms: [], // Initialize with empty VMs
                info: null, // Info will be fetched by the full sync
            });
            // Mark as connecting while backend attempts to connect and sync
            hostConnecting.value[newHost.id] = true;

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
             // The websocket will trigger a full refresh
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
        try {
            const controller = new AbortController();
            const id = setTimeout(() => controller.abort(), 5000);
            const response = await fetch(`/api/v1/hosts/${hostId}/vms`, { signal: controller.signal });
            clearTimeout(id);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json() || [];
        } catch (error) {
            console.error(`Failed to fetch VMs for ${hostId}:`, error);
            return [];
        }
    };

    const fetchHostPorts = async (hostId) => {
        if (!hostId) return [];
        try {
            const controller = new AbortController();
            const id = setTimeout(() => controller.abort(), 5000);
            const response = await fetch(`/api/v1/hosts/${hostId}/ports`, { signal: controller.signal });
            clearTimeout(id);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json() || [];
        } catch (error) {
            console.error(`Failed to fetch ports for ${hostId}:`, error);
            return [];
        }
    };

        const fetchVideoModels = async () => {
            try {
                const res = await fetch('/api/v1/video/models');
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return await res.json();
            } catch (e) {
                console.error('Failed to fetch video models', e);
                return [];
            }
        };

        const fetchHostVideoDevices = async (hostId) => {
            if (!hostId) return [];
            try {
                const res = await fetch(`/api/v1/hosts/${hostId}/video/devices`);
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return await res.json();
            } catch (e) {
                console.error('Failed to fetch host video devices', e);
                return [];
            }
        };

        const fetchVmVideoAttachments = async (hostId, vmName) => {
            if (!hostId || !vmName) return [];
            try {
                const res = await fetch(`/api/v1/hosts/${hostId}/vms/${vmName}/video-attachments`);
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return await res.json();
            } catch (e) {
                console.error('Failed to fetch VM video attachments', e);
                return [];
            }
        };

    const fetchVmPortAttachments = async (hostId, vmName) => {
        if (!hostId || !vmName) return [];
        try {
            const controller = new AbortController();
            const id = setTimeout(() => controller.abort(), 5000);
            const response = await fetch(`/api/v1/hosts/${hostId}/vms/${vmName}/port-attachments`, { signal: controller.signal });
            clearTimeout(id);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json() || [];
        } catch (error) {
            console.error(`Failed to fetch port attachments for ${hostId}/${vmName}:`, error);
            return [];
        }
    };

    const subscribeToVmStats = (hostId, vmName) => {
        console.log(`[mainStore] subscribeToVmStats: Attempting to subscribe to ${hostId}/${vmName}`);
        console.log(`[mainStore] Current VM subscription: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);

        // If there's an active VM subscription and it's different from the new one, unsubscribe it.
        if (currentlySubscribedVmName.value && 
            (currentlySubscribedVmName.value !== vmName || currentlySubscribedHostId.value !== hostId)) {
            console.log(`[mainStore] Unsubscribing from previous VM: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);
            sendMessage('unsubscribe-vm-stats', { hostId: currentlySubscribedHostId.value, vmName: currentlySubscribedVmName.value });
            activeVmStats.value = null; // Clear old stats immediately
        }
        // If there's an active host subscription, unsubscribe it.
        if (currentlySubscribedHostId.value && currentlySubscribedHostId.value !== hostId) {
            console.log(`[mainStore] Unsubscribing from previous Host: ${currentlySubscribedHostId.value}`);
            sendMessage('unsubscribe-host-stats', { hostId: currentlySubscribedHostId.value });
            hostStats.value[currentlySubscribedHostId.value] = null; // Clear old host stats
        }

        currentlySubscribedVmName.value = vmName;
        currentlySubscribedHostId.value = hostId;
        sendMessage('subscribe-vm-stats', { hostId, vmName });
        console.log(`[mainStore] Subscribed to ${hostId}/${vmName}. New state: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);
    };

    const unsubscribeFromVmStats = (hostId, vmName) => {
        console.log(`[mainStore] unsubscribeFromVmStats: Attempting to unsubscribe from ${hostId}/${vmName}`);
        console.log(`[mainStore] Current VM subscription: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);

        // Only unsubscribe if the VM being unsubscribed is the one currently subscribed
        if (currentlySubscribedVmName.value === vmName && currentlySubscribedHostId.value === hostId) {
            console.log(`[mainStore] Confirmed match. Sending unsubscribe for ${hostId}/${vmName}`);
            sendMessage('unsubscribe-vm-stats', { hostId, vmName });
            currentlySubscribedVmName.value = null;
            currentlySubscribedHostId.value = null; // Clear host ID as well if no VM is subscribed
            activeVmStats.value = null; // Clear last known stats
        } else {
            console.log(`[mainStore] No match. Not unsubscribing from ${hostId}/${vmName}. Current: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);
        }
    };

    const subscribeHostStats = (hostId) => {
        console.log(`[mainStore] subscribeHostStats: Attempting to subscribe to host ${hostId}`);
        console.log(`[mainStore] Current Host subscription: ${currentlySubscribedHostId.value}`);
        console.log(`[mainStore] Current VM subscription: ${currentlySubscribedVmName.value}`);

        // If there's an active VM subscription, unsubscribe it.
        if (currentlySubscribedVmName.value) {
            console.log(`[mainStore] Unsubscribing from previous VM: ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);
            sendMessage('unsubscribe-vm-stats', { hostId: currentlySubscribedHostId.value, vmName: currentlySubscribedVmName.value });
            currentlySubscribedVmName.value = null;
            activeVmStats.value = null; // Clear old stats immediately
        }
        // If a different host is currently subscribed, unsubscribe it.
        if (currentlySubscribedHostId.value && currentlySubscribedHostId.value !== hostId) {
            console.log(`[mainStore] Unsubscribing from previous Host: ${currentlySubscribedHostId.value}`);
            sendMessage('unsubscribe-host-stats', { hostId: currentlySubscribedHostId.value });
            hostStats.value[currentlySubscribedHostId.value] = null; // Clear old host stats
        }

        currentlySubscribedHostId.value = hostId;
        currentlySubscribedVmName.value = null; // Not viewing a VM
        sendMessage('subscribe-host-stats', { hostId });
        console.log(`[mainStore] Subscribed to host ${hostId}. New state: ${currentlySubscribedHostId.value}`);
    };

    const unsubscribeHostStats = (hostId) => {
        console.log(`[mainStore] unsubscribeHostStats: Attempting to unsubscribe from host ${hostId}`);
        console.log(`[mainStore] Current Host subscription: ${currentlySubscribedHostId.value}`);

        // Only unsubscribe if the host being unsubscribed is the one currently subscribed
        if (currentlySubscribedHostId.value === hostId) {
            console.log(`[mainStore] Confirmed match. Sending unsubscribe for host ${hostId}`);
            sendMessage('unsubscribe-host-stats', { hostId });
            currentlySubscribedHostId.value = null;
            hostStats.value[hostId] = null; // Clear last known stats for this host
        } else {
            console.log(`[mainStore] No match. Not unsubscribing from host ${hostId}. Current: ${currentlySubscribedHostId.value}`);
        }
    };

    // The clearAllSubscriptions can remain mostly the same, but ensure it clears both.
    const clearAllSubscriptions = () => {
        console.log("[mainStore] clearAllSubscriptions: Clearing all active subscriptions.");
        if (currentlySubscribedHostId.value) {
            console.log(`[mainStore] Clearing host subscription for ${currentlySubscribedHostId.value}`);
            sendMessage('unsubscribe-host-stats', { hostId: currentlySubscribedHostId.value });
            hostStats.value[currentlySubscribedHostId.value] = null;
            currentlySubscribedHostId.value = null;
        }
        if (currentlySubscribedVmName.value) { // No need for selectedHostId.value here, use currentlySubscribedHostId.value
            console.log(`[mainStore] Clearing VM subscription for ${currentlySubscribedHostId.value}/${currentlySubscribedVmName.value}`);
            sendMessage('unsubscribe-vm-stats', { hostId: currentlySubscribedHostId.value, vmName: currentlySubscribedVmName.value });
            activeVmStats.value = null;
            currentlySubscribedVmName.value = null;
        }
        console.log("[mainStore] All subscriptions cleared.");
    };


    const fetchVmHardware = async (hostId, vmName) => {
        if (!hostId || !vmName) {
            activeVmHardware.value = null;
            return;
        }
        isLoading.value.vmHardware = true;
        try {
            const response = await fetch(`/api/v1/hosts/${hostId}/vms/${vmName}/hardware`);
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            activeVmHardware.value = await response.json();
        } catch (error) {
            console.error(`Failed to fetch hardware for VM ${vmName}:`, error);
            activeVmHardware.value = null;
        } finally {
            isLoading.value.vmHardware = false;
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
            // The websocket will handle the UI update
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

    const performVmReconciliationAction = async (hostId, vmName, action) => {
        isLoading.value.vmReconcile = `${vmName}:${action}`;
        errorMessage.value = '';
        try {
            const response = await fetch(`/api/v1/hosts/${hostId}/vms/${vmName}/${action}`, { method: 'POST' });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            // The websocket will handle the UI update
        } catch (error) {
            errorMessage.value = `Action '${action}' on VM '${vmName}' failed: ${error.message}`;
            console.error(error);
        } finally {
            isLoading.value.vmReconcile = null;
        }
    };

    const syncVmFromLibvirt = (hostId, vmName) => performVmReconciliationAction(hostId, vmName, 'sync-from-libvirt');
    const rebuildVmFromDb = (hostId, vmName) => performVmReconciliationAction(hostId, vmName, 'rebuild-from-db');

    return {
        hosts,
        selectedHostId,
        errorMessage,
        isLoading,
        activeVmStats,
        activeVmHardware,
        hostStats, // Export hostStats
        hostConnecting,
        initializeRealtime,
        fetchHosts,
        addHost,
        deleteHost,
        selectHost,
        fetchVmHardware,
    fetchHostPorts,
    fetchVmPortAttachments,
    fetchVideoModels,
    fetchHostVideoDevices,
    fetchVmVideoAttachments,
        startVm,
        gracefulShutdownVm,
        gracefulRebootVm,
        forceOffVm,
        forceResetVm,
        subscribeToVmStats,
        unsubscribeFromVmStats,
        subscribeHostStats, // Export subscribeHostStats
        unsubscribeHostStats, // Export unsubscribeHostStats
        syncVmFromLibvirt,
        rebuildVmFromDb,
        totalVms,
        clearAllSubscriptions,
    };
});