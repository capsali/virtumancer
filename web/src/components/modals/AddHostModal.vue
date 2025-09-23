<template>
  <div>Add Host Placeholder</div>
</template>

<script setup lang="ts">
interface Props {
  open: boolean;
}

defineProps<Props>();

defineEmits<{
  close: [];
  submit: [hostData: any];
}>();
</script>
