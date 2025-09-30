<template>
  <div
    :class="[
      'glass-card transition-all duration-300',
      sizeClasses,
      paddingClasses,
      roundingClasses,
      interactiveClasses,
      glowClasses
    ]"
    v-bind="$attrs"
  >
    <!-- Header (optional) -->
    <div v-if="hasHeader" :class="headerClasses">
      <slot name="header">
        <h3 v-if="title" class="text-lg font-semibold text-white mb-1">{{ title }}</h3>
        <p v-if="subtitle" class="text-sm text-slate-400">{{ subtitle }}</p>
      </slot>
    </div>

    <!-- Content -->
    <div :class="contentClasses">
      <slot></slot>
    </div>

    <!-- Footer (optional) -->
    <div v-if="hasFooter" :class="footerClasses">
      <slot name="footer"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'subtle' | 'strong'
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
  rounding?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  interactive?: boolean
  glow?: boolean
  title?: string
  subtitle?: string
  hasHeader?: boolean
  hasFooter?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'full',
  padding: 'md',
  rounding: 'lg',
  interactive: false,
  glow: true,
  hasHeader: false,
  hasFooter: false
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'max-w-sm'
    case 'md':
      return 'max-w-md'
    case 'lg':
      return 'max-w-lg'
    case 'xl':
      return 'max-w-xl'
    case 'full':
    default:
      return 'w-full'
  }
})

const paddingClasses = computed(() => {
  switch (props.padding) {
    case 'none':
      return ''
    case 'sm':
      return 'p-3'
    case 'md':
      return 'p-4'
    case 'lg':
      return 'p-6'
    case 'xl':
      return 'p-8'
    default:
      return 'p-4'
  }
})

const roundingClasses = computed(() => {
  switch (props.rounding) {
    case 'sm':
      return 'rounded'
    case 'md':
      return 'rounded-md'
    case 'lg':
      return 'rounded-lg'
    case 'xl':
      return 'rounded-xl'
    case '2xl':
      return 'rounded-2xl'
    default:
      return 'rounded-lg'
  }
})

const interactiveClasses = computed(() => {
  if (!props.interactive) return ''
  return 'cursor-pointer hover:scale-[1.02] active:scale-[0.98]'
})

const glowClasses = computed(() => {
  const baseClasses = []
  
  // Glass effect
  switch (props.variant) {
    case 'subtle':
      baseClasses.push('glass-subtle')
      break
    case 'strong':
      baseClasses.push('glass-strong')
      break
    case 'default':
    default:
      baseClasses.push('glass-medium')
      break
  }
  
  // Glow effect
  if (props.glow) {
    baseClasses.push('card-glow')
  }
  
  return baseClasses.join(' ')
})

const headerClasses = computed(() => {
  const classes = ['border-b border-white/10']
  
  switch (props.padding) {
    case 'sm':
      classes.push('p-3 pb-2')
      break
    case 'md':
      classes.push('p-4 pb-3')
      break
    case 'lg':
      classes.push('p-6 pb-4')
      break
    case 'xl':
      classes.push('p-8 pb-6')
      break
    default:
      classes.push('p-4 pb-3')
      break
  }
  
  return classes.join(' ')
})

const contentClasses = computed(() => {
  if (props.hasHeader || props.hasFooter) {
    return '' // No padding when header/footer handle it
  }
  return '' // Main container handles padding
})

const footerClasses = computed(() => {
  const classes = ['border-t border-white/10']
  
  switch (props.padding) {
    case 'sm':
      classes.push('p-3 pt-2')
      break
    case 'md':
      classes.push('p-4 pt-3')
      break
    case 'lg':
      classes.push('p-6 pt-4')
      break
    case 'xl':
      classes.push('p-8 pt-6')
      break
    default:
      classes.push('p-4 pt-3')
      break
  }
  
  return classes.join(' ')
})
</script>

<style scoped>
.glass-card {
  position: relative;
  overflow: hidden;
}

.glass-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  pointer-events: none;
}

.glass-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  pointer-events: none;
}
</style>