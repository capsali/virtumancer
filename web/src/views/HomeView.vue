<script setup>
import { ref, onMounted } from 'vue';

// Reactive state
const hosts = ref([]);
const newHostId = ref('');
const newHostUri = ref('');
const isLoading = ref(false);
const errorMessage = ref('');

// Fetch hosts from the API
const fetchHosts = async () => {
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
  }
};

// Add a new host
const addHost = async () => {
  if (!newHostId.value || !newHostUri.value) {
    errorMessage.value = "Both Host ID and URI are required.";
    return;
  }
  isLoading.value = true;
  errorMessage.value = '';
  try {
    const response = await fetch('/api/v1/hosts', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ id: newHostId.value, uri: newHostUri.value }),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP error! status: ${response.status}`);
    }
    // Clear form and refresh host list
    newHostId.value = '';
    newHostUri.value = '';
    await fetchHosts();
  } catch (error) {
    console.error("Error adding host:", error);
    errorMessage.value = `Failed to add host: ${error.message}`;
  } finally {
    isLoading.value = false;
  }
};

// Delete a host
const deleteHost = async (hostId) => {
  if (!confirm(`Are you sure you want to delete host "${hostId}"?`)) {
    return;
  }
  try {
    const response = await fetch(`/api/v1/hosts/${hostId}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP error! status: ${response.status}`);
    }
    await fetchHosts();
  } catch (error) {
    console.error("Error deleting host:", error);
    errorMessage.value = `Failed to delete host: ${error.message}`;
  }
};

// Fetch hosts when the component is first mounted
onMounted(fetchHosts);
</script>

<template>
  <div class="bg-gray-900 text-gray-100 min-h-screen font-sans">
    <div class="container mx-auto p-4 md:p-8">

      <header class="mb-10">
        <h1 class="text-4xl md:text-5xl font-bold text-white tracking-wider">Virtu<span class="text-indigo-400">Mancer</span></h1>
        <p class="text-gray-400 mt-2">The Modern Web UI for Libvirt</p>
      </header>

      <main class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        <!-- Left Column: Add Host Form -->
        <div class="lg:col-span-1">
          <div class="bg-gray-800 p-6 rounded-lg shadow-lg">
            <h2 class="text-2xl font-semibold mb-4 border-b border-gray-700 pb-3">Add New Host</h2>
            <form @submit.prevent="addHost" class="space-y-4">
              <div>
                <label for="hostId" class="block text-sm font-medium text-gray-300">Host ID (a short name)</label>
                <input 
                  id="hostId"
                  v-model="newHostId" 
                  type="text" 
                  placeholder="e.g., proxmox-1"
                  class="mt-1 block w-full bg-gray-700 border-gray-600 rounded-md shadow-sm text-white placeholder-gray-400 focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>
              <div>
                <label for="hostUri" class="block text-sm font-medium text-gray-300">Connection URI</label>
                <input 
                  id="hostUri"
                  v-model="newHostUri" 
                  type="text" 
                  placeholder="qemu+ssh://user@hostname/system"
                  class="mt-1 block w-full bg-gray-700 border-gray-600 rounded-md shadow-sm text-white placeholder-gray-400 focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>
              <button 
                type="submit"
                :disabled="isLoading"
                class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 focus:ring-offset-gray-800 disabled:bg-indigo-800 disabled:cursor-not-allowed"
              >
                <svg v-if="isLoading" class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                {{ isLoading ? 'Connecting...' : 'Add Host' }}
              </button>
            </form>
            <p v-if="errorMessage" class="mt-4 text-sm text-red-400 bg-red-900/50 p-3 rounded-md">{{ errorMessage }}</p>
          </div>
        </div>

        <!-- Right Column: Managed Hosts List -->
        <div class="lg:col-span-2">
          <div class="bg-gray-800 p-6 rounded-lg shadow-lg">
            <h2 class="text-2xl font-semibold mb-4 border-b border-gray-700 pb-3">Managed Hosts</h2>
            <div v-if="hosts.length === 0" class="text-gray-400">
              No hosts configured yet. Add one to get started.
            </div>
            <ul v-else class="space-y-3">
              <li v-for="host in hosts" :key="host.id" class="flex items-center justify-between bg-gray-700 p-4 rounded-md hover:bg-gray-600 transition-colors duration-200">
                <div>
                  <p class="font-semibold text-white">{{ host.id }}</p>
                  <p class="text-sm text-gray-400 font-mono">{{ host.uri }}</p>
                </div>
                <button 
                  @click="deleteHost(host.id)"
                  class="text-gray-400 hover:text-red-400 focus:outline-none focus:text-red-400 transition-colors"
                  aria-label="Delete host"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </li>
            </ul>
          </div>
        </div>
      </main>

    </div>
  </div>
</template>

