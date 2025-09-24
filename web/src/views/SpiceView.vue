<template>
  <div class="bg-black w-full h-full flex flex-col text-white font-sans overflow-hidden">
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
      <div class="flex items-center gap-3">
        <FButton
          variant="outline"
          size="sm"
          @click="showFileTransfer = true"
          class="text-blue-400 hover:text-blue-300 border-blue-500/30 hover:border-blue-400/50"
          :disabled="connectionStatus !== 'Connected'"
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
          </svg>
          Transfer Files
        </FButton>
        <div class="text-right">
          <p class="font-semibold text-sm" :class="statusColor">
            {{ connectionStatus }}
          </p>
        </div>
      </div>
    </header>
    
    <main class="flex-grow relative bg-black overflow-hidden" style="width: 100%; height: calc(100vh - 4rem);">
      <iframe
        v-if="spiceIframeSrc"
        ref="consoleIframe"
        :src="spiceIframeSrc"
        @load="onIframeLoad"
        class="console-iframe"
        title="SPICE Console"
        allow="clipboard-read; clipboard-write"
        scrolling="no"
        frameborder="0"
      />
      <div v-else class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="w-8 h-8 border-2 border-white/30 border-t-white rounded-full animate-spin mx-auto mb-4"></div>
          <p>Generating connection URL...</p>
        </div>
      </div>
    </main>

    <!-- File Transfer Modal -->
    <FileTransferModal
      :show="showFileTransfer"
      :vm-name="vmName"
      @close="showFileTransfer = false"
      @transfer="handleFileTransfer"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import FButton from '@/components/ui/FButton.vue';
import FileTransferModal from '@/components/modals/FileTransferModal.vue';

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();
const router = useRouter();

const connectionStatus = ref('Loading...');
const consoleIframe = ref<HTMLIFrameElement | null>(null);
const showFileTransfer = ref(false);

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

  return `/spice/spice_responsive.html?${params.toString()}`;
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
  
  // Ensure iframe is properly sized after load
  if (consoleIframe.value) {
    // Force a resize to ensure proper scaling
    setTimeout(() => {
      if (consoleIframe.value) {
        consoleIframe.value.style.width = '100%';
        consoleIframe.value.style.height = '100%';
      }
    }, 100);
  }
};

const goBack = (): void => {
  router.back();
};

// File transfer handling
const handleFileTransfer = (files: File[]): void => {
  // Send message to SPICE client to handle file transfer
  if (consoleIframe.value && consoleIframe.value.contentWindow) {
    consoleIframe.value.contentWindow.postMessage({
      type: 'file-transfer',
      files: Array.from(files).map(file => ({
        name: file.name,
        size: file.size,
        type: file.type,
        file: file
      }))
    }, '*');
  }
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

  // Handle window resize to ensure console scales properly
  const handleResize = () => {
    if (consoleIframe.value) {
      // Trigger resize in the iframe content
      consoleIframe.value.style.width = '99%';
      consoleIframe.value.offsetHeight; // Force reflow
      consoleIframe.value.style.width = '100%';
    }
  };

  window.addEventListener('message', handleMessage);
  window.addEventListener('resize', handleResize);
  
  // Initial resize check after a short delay
  setTimeout(handleResize, 500);
  
  // Cleanup listeners when component unmounts
  return () => {
    window.removeEventListener('message', handleMessage);
    window.removeEventListener('resize', handleResize);
  };
});
</script>

<style scoped>
.console-iframe {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border: 0;
  background-color: #1a1a1a;
  overflow: hidden;
  box-sizing: border-box;
  max-width: 100%;
  max-height: 100%;
}

/* Ensure proper scaling on different screen sizes */
@media (max-width: 768px) {
  .console-iframe {
    width: 100%;
    height: 100%;
  }
}

/* Fix for some browsers that might add scrollbars */
main {
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE 10+ */
}

main::-webkit-scrollbar {
  display: none; /* Chrome/Safari */
}
</style>