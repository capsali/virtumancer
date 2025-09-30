<template>
  <Teleport to="body">
    <div
      v-if="isVisible"
      ref="dropdownRef"
      :class="[
        'fixed z-[99999]',
        'glass-strong backdrop-blur-xl rounded-2xl border border-white/10',
        'shadow-floating-lg transform origin-top-right transition-all duration-200'
      ]"
      :style="dropdownStyle"
      v-click-away="onClickAway"
    >
      <!-- Header (optional) -->
      <div v-if="hasHeader" class="p-4 border-b border-white/10">
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
      <div v-if="hasFooter" class="p-4 border-t border-white/10">
        <slot name="footer"></slot>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, nextTick, watch } from 'vue'

interface Props {
  isVisible: boolean
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right'
  size?: 'sm' | 'md' | 'lg' | 'xl'
  title?: string
  subtitle?: string
  hasHeader?: boolean
  hasFooter?: boolean
  contentPadding?: boolean
  triggerElement?: HTMLElement
}

interface Emits {
  close: []
}

const props = withDefaults(defineProps<Props>(), {
  position: 'bottom-right',
  size: 'md',
  hasHeader: false,
  hasFooter: false,
  contentPadding: true
})

const emit = defineEmits<Emits>()

const dropdownRef = ref<HTMLElement>()
const dropdownStyle = ref({})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'w-48'
    case 'md':
      return 'w-64'
    case 'lg':
      return 'w-80'
    case 'xl':
      return 'w-96'
    default:
      return 'w-64'
  }
})

const contentClasses = computed(() => {
  return props.contentPadding ? 'p-4' : ''
})

const updatePosition = () => {
  if (!props.triggerElement || !dropdownRef.value) return

  const trigger = props.triggerElement
  const rect = trigger.getBoundingClientRect()
  const dropdown = dropdownRef.value
  
  let top = 0
  let left = 0

  switch (props.position) {
    case 'top-left':
      top = rect.top - dropdown.offsetHeight - 8
      left = rect.left
      break
    case 'top-right':
      top = rect.top - dropdown.offsetHeight - 8
      left = rect.right - dropdown.offsetWidth
      break
    case 'bottom-left':
      top = rect.bottom + 8
      left = rect.left
      break
    case 'bottom-right':
    default:
      top = rect.bottom + 8
      left = rect.right - dropdown.offsetWidth
      break
  }

  // Ensure dropdown stays within viewport
  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  }

  // Adjust horizontal position if needed
  if (left + dropdown.offsetWidth > viewport.width - 16) {
    // Try positioning to the left of the trigger first
    const leftAligned = rect.left - dropdown.offsetWidth + rect.width
    if (leftAligned >= 16) {
      left = leftAligned
    } else {
      // If that doesn't work, constrain to viewport
      left = viewport.width - dropdown.offsetWidth - 16
    }
  }
  if (left < 16) {
    left = 16
  }

  // Adjust vertical position if needed
  if (top + dropdown.offsetHeight > viewport.height - 16) {
    top = rect.top - dropdown.offsetHeight - 8
  }
  if (top < 16) {
    top = 16
  }

  dropdownStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    width: sizeClasses.value.replace('w-', '').replace('48', '12rem').replace('64', '16rem').replace('80', '20rem').replace('96', '24rem')
  }
}

watch(() => props.isVisible, async (visible) => {
  if (visible) {
    await nextTick()
    updatePosition()
  }
})

const onClickAway = () => {
  emit('close')
}

onMounted(() => {
  window.addEventListener('resize', updatePosition)
  window.addEventListener('scroll', updatePosition)
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePosition)
  window.removeEventListener('scroll', updatePosition)
})
</script>