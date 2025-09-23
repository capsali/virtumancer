<template>
  <button
    :class="[
      'relative overflow-hidden transition-all duration-300 transform',
      'focus:outline-none focus:ring-2 focus:ring-offset-2',
      baseClasses,
      sizeClasses,
      variantClasses,
      {
        'hover:scale-105 active:scale-95': !disabled,
        'opacity-50 cursor-not-allowed': disabled,
        'animate-pulse': loading
      }
    ]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <!-- Shimmer Effect -->
    <div 
      v-if="shimmer"
      class="absolute inset-0 bg-shimmer-gradient bg-[length:200%_100%] animate-shimmer"
    ></div>
    
    <!-- Content -->
    <div class="relative z-10 flex items-center justify-center gap-2">
      <!-- Loading Spinner -->
      <div
        v-if="loading"
        class="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin"
      ></div>
      
      <!-- Icon -->
      <component
        v-if="icon && !loading"
        :is="icon"
        class="w-5 h-5"
      />
      
      <!-- Text -->
      <span v-if="$slots.default || text">
        <slot>{{ text }}</slot>
      </span>
    </div>
    
    <!-- Glow Effect -->
    <div
      v-if="glow"
      :class="[
        'absolute inset-0 opacity-0 transition-opacity duration-300',
        'group-hover:opacity-100',
        glowClass
      ]"
    ></div>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  variant?: 'primary' | 'secondary' | 'accent' | 'ghost' | 'outline' | 'danger' | 'neon';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  text?: string;
  icon?: any;
  loading?: boolean;
  disabled?: boolean;
  shimmer?: boolean;
  glow?: boolean;
  rounded?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  loading: false,
  disabled: false,
  shimmer: false,
  glow: false,
  rounded: true
});

const emit = defineEmits<{
  click: [event: MouseEvent];
}>();

const baseClasses = computed(() => [
  'group font-semibold transition-all duration-300',
  props.rounded ? 'rounded-xl' : 'rounded-none'
]);

const sizeClasses = computed(() => {
  const sizes = {
    xs: 'px-3 py-1.5 text-xs',
    sm: 'px-4 py-2 text-sm',
    md: 'px-6 py-3 text-base',
    lg: 'px-8 py-4 text-lg',
    xl: 'px-10 py-5 text-xl'
  };
  return sizes[props.size];
});

const variantClasses = computed(() => {
  const variants = {
    primary: 'bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white shadow-lg hover:shadow-xl focus:ring-primary-500',
    secondary: 'bg-gradient-to-r from-slate-600 to-slate-700 hover:from-slate-500 hover:to-slate-600 text-white shadow-lg hover:shadow-xl focus:ring-slate-500',
    accent: 'bg-gradient-to-r from-accent-600 to-accent-700 hover:from-accent-500 hover:to-accent-600 text-white shadow-lg hover:shadow-xl focus:ring-accent-500',
    ghost: 'bg-transparent hover:bg-white/5 text-slate-300 hover:text-white focus:ring-slate-400',
    outline: 'border-2 border-slate-600 hover:border-slate-500 text-slate-300 hover:text-white hover:bg-slate-600/10 focus:ring-slate-500',
    danger: 'bg-gradient-to-r from-red-600 to-red-700 hover:from-red-500 hover:to-red-600 text-white shadow-lg hover:shadow-xl focus:ring-red-500',
    neon: 'bg-gradient-to-r from-neon-blue to-neon-cyan hover:from-neon-cyan hover:to-neon-blue text-slate-900 shadow-neon-blue hover:shadow-neon-cyan focus:ring-neon-blue'
  };
  return variants[props.variant];
});

const glowClass = computed(() => {
  const glows = {
    primary: 'shadow-glow-md',
    secondary: 'shadow-glow-md',
    accent: 'shadow-neon-cyan',
    ghost: '',
    outline: '',
    danger: 'shadow-glow-md',
    neon: 'shadow-neon-blue'
  };
  return glows[props.variant];
});

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event);
  }
};
</script>