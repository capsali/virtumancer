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
          <h1 class="font-bold text-lg">VNC Console: {{ vmName }}</h1>
          <p class="text-xs text-gray-400">Host: {{ hostId }}</p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <div class="text-right">
          <p class="font-semibold text-sm" :class="statusColor">
            {{ connectionStatus }}
          </p>
        </div>
      </div>
    </header>
    
    <main class="flex-grow relative bg-black overflow-hidden" style="width: 100%; height: calc(100vh - 4rem);">
      <iframe
        v-if="vncIframeSrc"
        ref="consoleIframe"
        :src="vncIframeSrc"
        @load="onIframeLoad"
        class="console-iframe"
        title="VNC Console"
        scrolling="no"
        frameborder="0"
      />
      
      <div v-else class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p class="text-gray-300">Connecting to VNC console...</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import FButton from '@/components/ui/FButton.vue';

const router = useRouter();

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();

// Reactive state
const vncIframeSrc = ref<string>('');
const connectionStatus = ref<'Connecting' | 'Connected' | 'Disconnected' | 'Error'>('Connecting');
const consoleIframe = ref<HTMLIFrameElement | null>(null);

// Computed properties
const statusColor = computed(() => {
  switch (connectionStatus.value) {
    case 'Connected':
      return 'text-green-400';
    case 'Connecting':
      return 'text-yellow-400';
    case 'Error':
      return 'text-red-400';
    case 'Disconnected':
      return 'text-gray-400';
    default:
      return 'text-gray-400';
  }
});

// Methods
const goBack = (): void => {
  router.push(`/hosts/${props.hostId}/vms/${props.vmName}`);
};

const onIframeLoad = (): void => {
  connectionStatus.value = 'Connected';
};

onMounted(async () => {
  // Build the VNC console URL using noVNC
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  const consoleWsUrl = `${protocol}//${host}/api/v1/hosts/${props.hostId}/vms/${props.vmName}/console`;
  
  // Use noVNC with the websocket URL
  // For now, we'll use a simple approach - you may want to implement a full noVNC integration
  vncIframeSrc.value = `/vnc/vnc.html?host=${encodeURIComponent(host)}&port=80&path=${encodeURIComponent(`/api/v1/hosts/${props.hostId}/vms/${props.vmName}/console`)}&encrypt=${window.location.protocol === 'https:' ? '1' : '0'}`;
  
  // Handle window resize to ensure console scales properly
  const handleResize = () => {
    if (consoleIframe.value) {
      // Trigger resize in the iframe content
      consoleIframe.value.style.width = '99%';
      consoleIframe.value.offsetHeight; // Force reflow
      consoleIframe.value.style.width = '100%';
    }
  };

  window.addEventListener('resize', handleResize);
  
  // Initial resize check after a short delay
  setTimeout(handleResize, 500);
  
  // Cleanup listeners when component unmounts
  return () => {
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
</style>