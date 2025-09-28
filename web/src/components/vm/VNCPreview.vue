<template>
  <div class="vnc-preview-container">
    <div
      ref="screen"
      class="vnc-preview-screen"
      :class="{ 'connecting': connectionStatus === 'Connecting', 'error': connectionStatus === 'Error' }"
    ></div>

    <!-- Connection status overlay -->
    <div v-if="connectionStatus !== 'Connected'" class="vnc-preview-overlay">
      <div class="vnc-preview-status">
        <div v-if="connectionStatus === 'Connecting'" class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500 mx-auto mb-2"></div>
        <div v-else-if="connectionStatus === 'Error'" class="text-lg mb-2">⚠️</div>
        <div class="text-xs text-gray-400">{{ statusMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
// @ts-ignore
import RFB from '@novnc/novnc/lib/rfb';
import { useSettingsStore } from '@/stores/settingsStore';

interface Props {
  hostId: string;
  vmName: string;
  width?: number;
  height?: number;
}

const props = withDefaults(defineProps<Props>(), {
  width: 320,
  height: 240,
});

const screen = ref<HTMLDivElement | null>(null);
const rfb = ref<RFB | null>(null);
const connectionStatus = ref<'Connecting' | 'Connected' | 'Error'>('Connecting');
const statusMessage = ref<string>('Connecting...');

const settings = useSettingsStore();

// Watch for prop changes to reconnect
watch(() => [props.hostId, props.vmName], () => {
  disconnect();
  connect();
});

const connect = (): void => {
  if (!screen.value) return;

  connectionStatus.value = 'Connecting';
  statusMessage.value = 'Connecting...';

  try {
    // Build the WebSocket URL for the VNC console
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/hosts/${props.hostId}/vms/${props.vmName}/console`;

    // Create RFB instance
    rfb.value = new RFB(screen.value, wsUrl, {
      credentials: {},
    });

    // Configure for preview mode
    rfb.value.viewOnly = true; // Read-only for preview
    rfb.value.scaleViewport = true; // Scale to fit container
    rfb.value.resizeSession = false;
    rfb.value.clipboardPasteSupport = false;

    // Set up event handlers
    rfb.value.addEventListener('connect', onConnect);
    rfb.value.addEventListener('disconnect', onDisconnect);
    rfb.value.addEventListener('securityfailure', onSecurityFailure);

  } catch (error) {
    console.error('Failed to create VNC preview connection:', error);
    connectionStatus.value = 'Error';
    statusMessage.value = 'Connection failed';
  }
};

const disconnect = (): void => {
  if (rfb.value) {
    rfb.value.disconnect();
    rfb.value = null;
  }
  connectionStatus.value = 'Connecting';
  statusMessage.value = 'Disconnected';
};

// Event handlers
const onConnect = (): void => {
  connectionStatus.value = 'Connected';
  statusMessage.value = 'Connected';

  // Update canvas size to fit preview container
  if (rfb.value && screen.value) {
    const canvas = screen.value.querySelector('canvas');
    if (canvas) {
      canvas.style.width = `${props.width}px`;
      canvas.style.height = `${props.height}px`;
      canvas.style.maxWidth = '100%';
      canvas.style.maxHeight = '100%';
      canvas.style.objectFit = 'contain';
        // If previewScale is fill, scale canvas to fill the preview container
        if (settings.previewScale === 'fill') {
          // compute required scale
          const containerRect = screen.value!.getBoundingClientRect();
          const scaleX = containerRect.width / canvas.width;
          const scaleY = containerRect.height / canvas.height;
          const scale = Math.max(scaleX, scaleY);
          canvas.style.transformOrigin = 'center center';
          canvas.style.transform = `scale(${scale})`;
          canvas.style.maxWidth = 'none';
          canvas.style.maxHeight = 'none';
        } else {
          canvas.style.transform = '';
        }
    }
  }
};

const onDisconnect = (e: any): void => {
  connectionStatus.value = 'Error';
  statusMessage.value = e.detail?.clean ? 'Disconnected' : 'Connection lost';
  rfb.value = null;
};

const onSecurityFailure = (e: any): void => {
  connectionStatus.value = 'Error';
  statusMessage.value = `Security error: ${e.detail.reason}`;
};

// Lifecycle
onMounted(() => {
  connect();
});

onUnmounted(() => {
  disconnect();
});
</script>

<style scoped>
.vnc-preview-container {
  position: relative;
  background: black;
  border-radius: 8px;
  overflow: hidden;
  width: v-bind('props.width + "px"');
  height: v-bind('props.height + "px"');
}

.vnc-preview-screen {
  width: 100%;
  height: 100%;
}

.vnc-preview-screen.connecting {
  background: rgb(17 24 39);
}

.vnc-preview-screen.error {
  background: rgba(153, 27, 27, 0.2);
}

.vnc-preview-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
}

.vnc-preview-status {
  text-align: center;
  color: white;
}
</style>