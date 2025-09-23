<template>
  <div class="bg-black w-screen h-screen flex flex-col text-white font-sans">
    <header class="bg-gray-800 p-2 flex items-center justify-between shadow-md z-10 flex-shrink-0">
      <div class="flex items-center">
        <FButton
          variant="ghost"
          size="sm"
          @click="goBack"
          class="mr-4 text-indigo-400 hover:text-indigo-300"
        >
          ← Back
        </FButton>
        <div>
          <h1 class="font-bold text-lg">SPICE Console: {{ vmName }}</h1>
          <p class="text-xs text-gray-400">Host: {{ hostId }}</p>
        </div>
      </div>
      <div class="text-right">
        <p class="font-semibold text-sm" :class="statusColor">
          {{ connectionStatus }}
        </p>
      </div>
    </header>
    
    <main class="flex-grow w-full h-full relative bg-black">
      <iframe
        v-if="spiceIframeSrc"
        :src="spiceIframeSrc"
        @load="onIframeLoad"
        class="w-full h-full border-0"
        title="SPICE Console"
      />
      <div v-else class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="w-8 h-8 border-2 border-white/30 border-t-white rounded-full animate-spin mx-auto mb-4"></div>
          <p>Generating connection URL...</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import FButton from '@/components/ui/FButton.vue';

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const router = useRouter();

const connectionStatus = ref('Loading...');

// Dynamically construct the source URL for the iframe
const spiceIframeSrc = computed(() => {
  if (!props.hostId || !props.vmName) {
    return '';
  }

  // Use hostname to avoid including the port in the host parameter
  const host = window.location.hostname;
  // The port is needed separately by the spice-html5 client
  const port = window.location.port || (window.location.protocol === 'https:' ? '443' : '80');
  
  // The backend proxy path for the SPICE connection
  const path = `api/v1/hosts/${props.hostId}/vms/${props.vmName}/spice`;

  // Assemble the query parameters for spice_auto.html
  const params = new URLSearchParams({
    host: host,
    port: port,
    password: '', // Assuming no password for now
    path: path,
    encrypt: window.location.protocol === 'https:' ? '1' : '0'
  });

  return `/spice/spice_auto.html?${params.toString()}`;
});

const statusColor = computed(() => {
  switch (connectionStatus.value) {
    case 'Connected':
      return 'text-green-400';
    case 'Client Loaded':
      return 'text-yellow-400';
    case 'Loading...':
      return 'text-blue-400';
    case 'Error':
      return 'text-red-400';
    default:
      return 'text-gray-400';
  }
});

// Update status when the iframe has loaded the page
const onIframeLoad = (): void => {
  connectionStatus.value = 'Client Loaded';
};

const goBack = (): void => {
  router.back();
};

// Try to detect connection status (limited by CORS)
onMounted(() => {
  // Listen for potential postMessage events from the SPICE client
  const handleMessage = (event: MessageEvent) => {
    if (event.origin !== window.location.origin) return;
    
    if (event.data.type === 'spice-status') {
      connectionStatus.value = event.data.status;
    }
  };

  window.addEventListener('message', handleMessage);
  
  // Cleanup listener when component unmounts
  return () => {
    window.removeEventListener('message', handleMessage);
  };
});
</script>

<style scoped>
iframe {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
</style>