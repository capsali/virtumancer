<template>
  <div class="fixed bottom-4 right-4 z-50 space-y-2">
    <!-- ARIA live region: polite for non-errors; errors have role="alert" which is assertive -->
    <div aria-live="polite" aria-atomic="true">
      <div v-for="t in toasts" :key="t.id" class="max-w-sm w-full px-4 py-2 rounded shadow-lg" :class="t.type === 'error' ? 'bg-red-600 text-white' : 'bg-green-600 text-white'" :role="t.type === 'error' ? 'alert' : 'status'" aria-atomic="true">
        <div class="flex items-center justify-between">
          <div class="text-sm">{{ t.message }}</div>
          <button @click="remove(t.id)" type="button" :aria-label="`Dismiss notification: ${t.message}`" class="ml-3 text-sm opacity-80 hover:opacity-100">✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue';
import { useMainStore } from '@/stores/mainStore';
const store = useMainStore();
const toasts = store.toasts;
const remove = (id) => store.removeToast(id);
</script>

<style scoped>
/* minimal container, styling by Tailwind */
</style>
