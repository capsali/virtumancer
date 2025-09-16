
<template>
  <div class="fixed inset-0 flex items-center justify-center z-50">
    <div class="absolute inset-0 bg-black opacity-50"></div>
    <div class="bg-gray-800 text-white rounded-lg shadow-lg max-w-lg w-full mx-4 p-6 z-10">
      <h3 class="text-lg font-semibold mb-2">{{ title }}</h3>
      <p class="text-sm text-gray-300 mb-4">{{ message }}</p>

      <div class="flex justify-end space-x-3">
        <button @click="$emit('cancel')" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded" :disabled="loading">{{ cancelText }}</button>
        <button @click="$emit('confirm')" :disabled="loading" :aria-busy="loading ? 'true' : 'false'" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 rounded text-white flex items-center">
          <svg v-if="loading" class="animate-spin mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, default: 'Confirm' },
  message: { type: String, default: '' },
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' },
  loading: { type: Boolean, default: false }
});
</script>

<style scoped>
/* Minimal styling provided by Tailwind in app; keep minimal fallback */
</style>
