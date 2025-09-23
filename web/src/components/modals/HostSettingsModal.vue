<template>
  <div>Host Settings Placeholder</div>
</template>

<script setup lang="ts">
interface Props {
  open: boolean;
  hostId?: string;
}

defineProps<Props>();

defineEmits<{
  close: [];
}>();
</script>
