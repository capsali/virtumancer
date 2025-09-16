<template>
  <div ref="containerRef" class="fixed bottom-4 right-4 z-80 space-y-2">
    <!-- ARIA live region: polite for non-errors; errors have role="alert" which is assertive -->
    <div aria-live="polite" aria-atomic="true">
      <div v-for="t in visibleToasts" :key="t.id" class="max-w-sm w-full px-4 py-2 rounded shadow-lg" :class="t.type === 'error' ? 'bg-red-600 text-white' : 'bg-green-600 text-white'" :role="t.type === 'error' ? 'alert' : 'status'" aria-atomic="true">
        <div class="flex items-center justify-between">
          <div class="text-sm">{{ t.message }}</div>
          <button @click="onDismiss(t.id)" type="button" :aria-label="`Dismiss notification: ${t.message}`" class="ml-3 text-sm opacity-80 hover:opacity-100">✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { watch, ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useMainStore } from '@/stores/mainStore';
const store = useMainStore();
// ToastContainer: minimal runtime diagnostics removed

// Track hidden toast ids locally so dismiss is instant
const hiddenRef = ref(new Set());

// Keep a shallow copy of the store toasts to avoid proxy/closure oddities
const currentToasts = ref([]);
const getToastsArray = () => {
  try {
    if (!store || store.toasts === undefined) return [];
    if (Array.isArray(store.toasts)) return store.toasts;
    if (store.toasts && store.toasts.value !== undefined) return store.toasts.value;
    return [];
  } catch (e) { return []; }
};

watch(() => getToastsArray(), (nv) => {
  currentToasts.value = (nv || []).slice();
}, { immediate: true });

const visibleToasts = computed(() => (currentToasts.value || []).filter(t => !hiddenRef.value.has(t.id)));

const containerRef = ref(null);

let unsubscribe = null;

onMounted(() => {
  try {
    // no-op: mounted
    // Log container presence and bounding rect
    nextTick(() => {
      const el = containerRef.value;
      if (!el) {
        console.warn('[ToastContainer] containerRef is null - component not in DOM');
        return;
      }
      try {
        const rect = el.getBoundingClientRect();
    // no-op in production
      } catch (e) {
        console.warn('[ToastContainer] error querying DOM', e);
      }
    });
  } catch (e) {
    console.warn('[ToastContainer] onMounted error', e);
  }

  // Fallback: subscribe to Pinia store changes to keep currentToasts in sync
  try {
    if (store && store.$subscribe) {
      unsubscribe = store.$subscribe(() => {
        try {
          const arr = getToastsArray();
          currentToasts.value = arr.slice();
        } catch (e) {
          // ignore
        }
      });
    }
  } catch (e) {
    // ignore
  }

  // Listen to DOM events from the store for added/removed toasts (fallback)
  try {
    const onAdded = (ev) => {
      try {
        const d = ev.detail || {};
        const arr = (store.toasts && store.toasts.value) || [];
        currentToasts.value = arr.slice();
        console.log('[ToastContainer] DOM event toast-added received; synced currentToasts length=', currentToasts.value.length);
      } catch (e) { console.warn('[ToastContainer] toast-added handler error', e); }
    };
    const onRemoved = (ev) => {
      try {
        const arr = (store.toasts && store.toasts.value) || [];
        currentToasts.value = arr.slice();
        console.log('[ToastContainer] DOM event toast-removed received; synced currentToasts length=', currentToasts.value.length);
      } catch (e) { console.warn('[ToastContainer] toast-removed handler error', e); }
    };
    window.addEventListener('virtumancer:toast-added', onAdded);
    window.addEventListener('virtumancer:toast-removed', onRemoved);
    unsubscribe = () => {
      window.removeEventListener('virtumancer:toast-added', onAdded);
      window.removeEventListener('virtumancer:toast-removed', onRemoved);
    };
  } catch (e) {
    console.warn('[ToastContainer] DOM event listeners setup failed', e);
  }
});

onUnmounted(() => {
  try { if (unsubscribe) unsubscribe(); } catch (e) {}
});

// Watch visibleToasts specifically for diagnostics
watch(visibleToasts, (nv) => {}, { immediate: true });

const remove = (id) => {
  // Optimistically hide in UI
  try { hiddenRef.value.add(id); } catch (e) { /* ignore */ }
  console.log('[ToastContainer] calling store.removeToast for', id);
  store.removeToast(id);
};

const onDismiss = (id) => {
  console.log('[ToastContainer] dismiss clicked for', id);
  remove(id);
};

watch(() => getToastsArray(), (nv) => {
  // Clean up hidden entries that no longer exist in the store
  const ids = new Set((nv || []).map(t => t.id));
  const toRemove = Array.from(hiddenRef.value).filter(id => !ids.has(id));
  toRemove.forEach(id => hiddenRef.value.delete(id));
  // toasts changed; currentToasts already synced via subscription
  // After DOM updates, log visible toast elements and their rects for debugging
  nextTick(() => {
    try {
  const nodes = Array.from(document.querySelectorAll('.max-w-sm'));
    // minimal DOM probe removed after verification
      try { console.log('[ToastContainer] containerRef children count=', containerRef.value ? containerRef.value.children.length : '<no containerRef>'); } catch (e) {}
  try { console.log('[ToastContainer] containerRef innerHTML length=', containerRef.value ? containerRef.value.innerHTML.length : '<no containerRef>'); } catch (e) {}
      nodes.forEach((n, i) => {
        try {
          const r = n.getBoundingClientRect();
          console.log(`[ToastContainer] toast[${i}] rect`, r, 'outerHTML:', n.outerHTML.slice(0, 200));
        } catch (e) {
          console.warn('[ToastContainer] error reading toast node', e);
        }
      });
    } catch (e) {
      console.warn('[ToastContainer] nextTick DOM probe failed', e);
    }
  });
}, { immediate: true });
</script>

<style scoped>
/* minimal container, styling by Tailwind */
</style>
