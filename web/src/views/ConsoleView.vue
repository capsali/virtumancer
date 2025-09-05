<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import RFB from '@novnc/novnc/lib/rfb';

const route = useRoute();
const vncCanvas = ref(null);
const connectionStatus = ref('Connecting...');
const rfb = ref(null);

const hostId = computed(() => route.params.hostId);
const vmName = computed(() => route.params.vmName);

const connect = () => {
  if (!vncCanvas.value) {
    console.error("VNC canvas ref is not available.");
    connectionStatus.value = 'Error: Canvas not ready.';
    return;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${protocol}//${window.location.host}/api/v1/hosts/${hostId.value}/vms/${vmName.value}/console`;
  
  console.log(`Connecting to VNC via ${url}`);

  const options = {
    wsProtocols: ['binary'], // Important for our Go proxy
  };

  const newRfb = new RFB(vncCanvas.value, url, options);

  newRfb.addEventListener('connect', () => {
    connectionStatus.value = 'Connected';
    console.log('noVNC connected');
  });

  newRfb.addEventListener('disconnect', () => {
    connectionStatus.value = 'Disconnected';
    console.log('noVNC disconnected');
  });

  newRfb.addEventListener('credentials', () => {
    // If VNC requires a password, libvirt can provide it.
    // For now, we assume no password is set on the VNC server.
    // newRfb.sendCredentials({ password: 'your-password' });
    console.log('Credentials requested');
  });
  
  rfb.value = newRfb;
};

const disconnect = () => {
  if (rfb.value) {
    rfb.value.disconnect();
    rfb.value = null;
  }
};

onMounted(() => {
  connect();
});

onUnmounted(() => {
  disconnect();
});
</script>

<template>
  <div class="bg-black w-screen h-screen flex flex-col text-white font-sans">
    <header class="bg-gray-800 p-2 flex items-center justify-between shadow-md z-10">
      <div class="flex items-center">
         <router-link to="/" class="text-indigo-400 hover:text-indigo-300 mr-4">
          &larr; Back
        </router-link>
        <div>
          <h1 class="font-bold text-lg">Console: {{ vmName }}</h1>
          <p class="text-xs text-gray-400">Host: {{ hostId }}</p>
        </div>
      </div>
      <div class="text-right">
         <p class="font-semibold text-sm" :class="{
          'text-green-400': connectionStatus === 'Connected',
          'text-red-400': connectionStatus === 'Disconnected',
          'text-yellow-400': connectionStatus === 'Connecting...'
         }">
          {{ connectionStatus }}
        </p>
      </div>
    </header>
    <main class="flex-grow w-full h-full relative">
      <!-- This is the target element for noVNC -->
      <div ref="vncCanvas" class="w-full h-full bg-black"></div>
    </main>
  </div>
</template>

<style>
/* Ensure the canvas fills the container */
#vnc-canvas canvas {
  width: 100%;
  height: 100%;
}
</style>

