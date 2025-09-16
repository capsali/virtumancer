<script setup>
import { RouterView } from 'vue-router';
import ToastContainer from '@/components/ui/ToastContainer.vue';
</script>

<template>
  <div class="h-screen w-screen bg-gray-900 text-gray-100">
    <RouterView />
    <ToastContainer />
  </div>
</template>


