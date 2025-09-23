<template>
  <button
    :class="[
      'relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500',
      {
        'bg-primary-600': modelValue && !disabled,
        'bg-slate-600': !modelValue && !disabled,
        'bg-slate-700 opacity-50 cursor-not-allowed': disabled
      }
    ]"
    :disabled="disabled"
    @click="toggle"
  >
    <!-- Toggle Dot -->
    <span
      :class="[
        'inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200 shadow-lg',
        {
          'translate-x-6': modelValue,
          'translate-x-1': !modelValue
        }
      ]"
    >
      <!-- Glow Effect -->
      <span
        v-if="glow && modelValue && !disabled"
        class="absolute inset-0 rounded-full bg-primary-400/40 blur-sm"
      ></span>
    </span>

    <!-- Background Glow -->
    <div
      v-if="glow && modelValue && !disabled"
      class="absolute inset-0 rounded-full bg-primary-600/30 blur-md opacity-75"
    ></div>
  </button>
</template>

<script setup lang="ts">
interface Props {
  modelValue: boolean;
  disabled?: boolean;
  glow?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  glow: true
});

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const toggle = () => {
  if (!props.disabled) {
    emit('update:modelValue', !props.modelValue);
  }
};
</script>