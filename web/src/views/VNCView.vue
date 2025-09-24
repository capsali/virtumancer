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
        <div class="flex items-center gap-2">
          <FButton
            variant="ghost"
            size="sm"
            @click="toggleFullscreen"
            :title="isFullscreen ? 'Exit Fullscreen' : 'Enter Fullscreen'"
            class="text-gray-300 hover:text-white"
          >
            {{ isFullscreen ? '⛶' : '⛶' }}
          </FButton>
          <FButton
            variant="ghost"
            size="sm"
            @click="reconnect"
            :disabled="connectionStatus === 'Connecting'"
            title="Reconnect"
            class="text-gray-300 hover:text-white"
          >
            ↻
          </FButton>
        </div>
        <div class="text-right">
          <p class="font-semibold text-sm" :class="statusColor">
            {{ connectionStatus }}
          </p>
          <p v-if="connectionStatus === 'Connected'" class="text-xs text-gray-500">
            {{ rfb?._fbName || 'Unknown' }}
          </p>
        </div>
      </div>
    </header>

    <main class="flex-grow relative bg-black overflow-hidden" :class="{ 'fullscreen-main': isFullscreen }">
      <div
        ref="screenContainer"
        class="screen-container"
        :class="{ 'fullscreen': isFullscreen }"
      >
        <div
          ref="screen"
          class="screen-canvas"
        ></div>

        <!-- Connection overlay -->
        <div v-if="connectionStatus !== 'Connected'" class="connection-overlay">
          <div class="connection-status">
            <div v-if="connectionStatus === 'Connecting'" class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
            <div v-else-if="connectionStatus === 'Disconnected'" class="text-4xl mb-4">🔌</div>
            <div v-else-if="connectionStatus === 'Error'" class="text-4xl mb-4">❌</div>
            <h3 class="text-lg font-semibold mb-2">{{ connectionStatus }}</h3>
            <p class="text-gray-300 text-sm mb-4">{{ statusMessage }}</p>
            <FButton
              v-if="connectionStatus === 'Disconnected' || connectionStatus === 'Error'"
              variant="primary"
              @click="connect"
            >
              Reconnect
            </FButton>
          </div>
        </div>

        <!-- Control bar overlay -->
        <div v-if="showControls && connectionStatus === 'Connected'" class="control-overlay">
          <div class="control-bar">
            <div class="control-group">
              <FButton
                variant="ghost"
                size="sm"
                @click="sendCtrlAltDel"
                title="Send Ctrl+Alt+Del"
                class="text-white hover:bg-gray-700"
              >
                Ctrl+Alt+Del
              </FButton>
              <FButton
                variant="ghost"
                size="sm"
                @click="toggleClipboard"
                :title="clipboardEnabled ? 'Disable Clipboard' : 'Enable Clipboard'"
                class="text-white hover:bg-gray-700"
              >
                📋 {{ clipboardEnabled ? 'On' : 'Off' }}
              </FButton>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue';
import { useRouter } from 'vue-router';
import FButton from '@/components/ui/FButton.vue';
// @ts-ignore
import RFB from '@novnc/novnc/lib/rfb';

const router = useRouter();

interface Props {
  hostId: string;
  vmName: string;
}

const props = defineProps<Props>();

// DOM refs
const screen = ref<HTMLDivElement | null>(null);
const screenContainer = ref<HTMLDivElement | null>(null);

// Reactive state
const rfb = ref<RFB | null>(null);
const connectionStatus = ref<'Connecting' | 'Connected' | 'Disconnected' | 'Error'>('Connecting');
const statusMessage = ref<string>('');
const isFullscreen = ref(false);
const showControls = ref(false);
const clipboardEnabled = ref(true);
const canvasWidth = ref(1024);
const canvasHeight = ref(768);

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

const toggleFullscreen = (): void => {
  isFullscreen.value = !isFullscreen.value;
  updateCanvasSize();
};

const reconnect = (): void => {
  disconnect();
  setTimeout(() => connect(), 100);
};

const sendCtrlAltDel = (): void => {
  if (rfb.value) {
    rfb.value.sendCtrlAltDel();
  }
};

const toggleClipboard = (): void => {
  clipboardEnabled.value = !clipboardEnabled.value;
  if (rfb.value) {
    rfb.value.clipboardPasteSupport = clipboardEnabled.value;
  }
};

const updateCanvasSize = (): void => {
  if (!screen.value || !screenContainer.value) return;

  const container = screenContainer.value;
  const containerRect = container.getBoundingClientRect();
  const canvas = screen.value.querySelector('canvas');

  // Calculate the best fit while maintaining aspect ratio
  const containerWidth = containerRect.width;
  const containerHeight = containerRect.height;

  if (rfb.value && canvas) {
    const fbWidth = rfb.value._fbWidth;
    const fbHeight = rfb.value._fbHeight;

    console.log('Canvas size update:', { fbWidth, fbHeight, containerWidth, containerHeight });

    if (fbWidth && fbHeight) {
      const scaleX = containerWidth / fbWidth;
      const scaleY = containerHeight / fbHeight;
      const scale = Math.min(scaleX, scaleY);

      const newWidth = Math.floor(fbWidth * scale);
      const newHeight = Math.floor(fbHeight * scale);

      console.log('Setting canvas size:', { width: newWidth, height: newHeight });

      // Update the actual canvas created by noVNC
      canvas.style.width = `${newWidth}px`;
      canvas.style.height = `${newHeight}px`;
      
      canvasWidth.value = newWidth;
      canvasHeight.value = newHeight;
    }
  }
};

const connect = (): void => {
  if (!screen.value) return;

  connectionStatus.value = 'Connecting';
  statusMessage.value = 'Establishing connection...';

  try {
    // Build the WebSocket URL for the VNC console
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/hosts/${props.hostId}/vms/${props.vmName}/console`;

    // Create RFB instance
    rfb.value = new RFB(screen.value, wsUrl, {
      credentials: {}, // Add credentials if needed
    });

    // Configure RFB
    rfb.value.clipboardPasteSupport = clipboardEnabled.value;
    rfb.value.viewOnly = false;
    rfb.value.scaleViewport = false; // We'll handle scaling manually
    rfb.value.resizeSession = false;
    
    console.log('RFB created with target canvas:', screen.value);
    console.log('WebSocket URL:', wsUrl);

    // Set up event handlers
    rfb.value.addEventListener('connect', onConnect);
    rfb.value.addEventListener('disconnect', onDisconnect);
    rfb.value.addEventListener('credentialsrequired', onCredentialsRequired);
    rfb.value.addEventListener('securityfailure', onSecurityFailure);
    rfb.value.addEventListener('clipboard', onClipboard);
    rfb.value.addEventListener('bell', onBell);
    rfb.value.addEventListener('desktopname', onDesktopName);
    rfb.value.addEventListener('fb-update', onFbUpdate);

  } catch (error) {
    console.error('Failed to create RFB connection:', error);
    connectionStatus.value = 'Error';
    statusMessage.value = 'Failed to initialize VNC connection';
  }
};

const disconnect = (): void => {
  if (rfb.value) {
    rfb.value.disconnect();
    rfb.value = null;
  }
  connectionStatus.value = 'Disconnected';
  statusMessage.value = 'Disconnected from VNC console';
};

// Event handlers
const onConnect = (): void => {
  connectionStatus.value = 'Connected';
  statusMessage.value = 'Connected successfully';

  console.log('VNC connected, container element:', screen.value);
  console.log('Container dimensions:', {
    clientWidth: screen.value?.clientWidth,
    clientHeight: screen.value?.clientHeight
  });
  
  console.log('RFB after connect:', {
    fbWidth: rfb.value?._fbWidth,
    fbHeight: rfb.value?._fbHeight,
    connected: rfb.value?.connected,
    canvas: rfb.value?._target
  });

  // Small delay to ensure framebuffer dimensions are available
  setTimeout(() => {
    updateCanvasSize();
  }, 100);
};

const onDisconnect = (e: any): void => {
  connectionStatus.value = 'Disconnected';
  statusMessage.value = e.detail?.clean ? 'Disconnected' : 'Connection lost';
  rfb.value = null;
};

const onCredentialsRequired = (e: any): void => {
  // Handle credential requests if needed
  console.log('Credentials required:', e.detail.types);
  statusMessage.value = 'Authentication required';
};

const onSecurityFailure = (e: any): void => {
  connectionStatus.value = 'Error';
  statusMessage.value = `Security failure: ${e.detail.reason}`;
};

const onClipboard = (e: any): void => {
  if (clipboardEnabled.value && e.detail.text) {
    navigator.clipboard.writeText(e.detail.text).catch(console.error);
  }
};

const onBell = (): void => {
  // Visual bell effect could be implemented here
  console.log('Bell received');
};

const onDesktopName = (e: any): void => {
  console.log('Desktop name:', e.detail.name);
  console.log('RFB state after desktop name:', {
    fbWidth: rfb.value?._fbWidth,
    fbHeight: rfb.value?._fbHeight,
    fbName: rfb.value?._fbName,
    connected: rfb.value?.connected
  });
};

const onFbUpdate = (e: any): void => {
  console.log('Framebuffer update received:', {
    x: e.detail.x,
    y: e.detail.y,
    width: e.detail.width,
    height: e.detail.height
  });
  
  // Force canvas repaint
  if (screen.value) {
    const canvas = screen.value.querySelector('canvas');
    console.log('Canvas after FB update:', {
      width: canvas?.width,
      height: canvas?.height,
      style: canvas?.style.cssText
    });
  }
};

// Lifecycle
onMounted(() => {
  nextTick(() => {
    connect();
  });

  // Handle window resize
  const handleResize = () => {
    updateCanvasSize();
  };

  window.addEventListener('resize', handleResize);

  // Handle fullscreen changes
  const handleFullscreenChange = () => {
    isFullscreen.value = !!document.fullscreenElement;
    updateCanvasSize();
  };

  document.addEventListener('fullscreenchange', handleFullscreenChange);

  // Show/hide controls on mouse movement
  let controlsTimeout: number;
  const handleMouseMove = () => {
    showControls.value = true;
    clearTimeout(controlsTimeout);
    controlsTimeout = window.setTimeout(() => {
      showControls.value = false;
    }, 3000);
  };

  if (screenContainer.value) {
    screenContainer.value.addEventListener('mousemove', handleMouseMove);
  }

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize);
    document.removeEventListener('fullscreenchange', handleFullscreenChange);
    if (screenContainer.value) {
      screenContainer.value.removeEventListener('mousemove', handleMouseMove);
    }
    clearTimeout(controlsTimeout);
    disconnect();
  });
});

// Watch for fullscreen changes
watch(isFullscreen, (newVal) => {
  if (newVal) {
    if (screenContainer.value) {
      screenContainer.value.requestFullscreen().catch(console.error);
    }
  } else {
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(console.error);
    }
  }
});
</script>

<style scoped>
.screen-container {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #000;
}

.screen-container.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
}

.screen-canvas {
  max-width: 100%;
  max-height: 100%;
  image-rendering: -moz-crisp-edges;
  image-rendering: -webkit-crisp-edges;
  image-rendering: pixelated;
  image-rendering: crisp-edges;
  cursor: none;
}

.connection-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.connection-status {
  text-align: center;
  color: white;
  max-width: 300px;
}

.control-overlay {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 5;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.screen-container:hover .control-overlay {
  opacity: 1;
}

.control-bar {
  background-color: rgba(0, 0, 0, 0.8);
  border-radius: 8px;
  padding: 8px;
  backdrop-filter: blur(10px);
}

.control-group {
  display: flex;
  gap: 8px;
}

.fullscreen-main {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
}
</style>