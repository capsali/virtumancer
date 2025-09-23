<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
      <!-- Backdrop -->
      <div 
        class="fixed inset-0 bg-black/50 transition-opacity"
        @click="handleBackdropClick"
      ></div>
      
      <!-- Modal content -->
      <div class="flex min-h-full items-center justify-center p-4">
        <div 
          :class="modalClasses"
          class="relative bg-gray-800 rounded-lg shadow-xl border border-gray-700 max-h-full overflow-hidden"
        >
          <!-- Close button -->
          <button
            v-if="showCloseButton"
            @click="$emit('close')"
            class="absolute top-4 right-4 z-10 text-gray-400 hover:text-white transition-colors"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          
          <!-- Modal body -->
          <div class="p-6 overflow-y-auto max-h-[90vh]">
            <slot></slot>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  show: boolean
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  showCloseButton?: boolean
  closeOnBackdrop?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  showCloseButton: true,
  closeOnBackdrop: true
})

const emit = defineEmits<{
  close: []
}>()

const modalClasses = computed(() => {
  const sizeClasses = {
    sm: 'max-w-sm w-full',
    md: 'max-w-md w-full',
    lg: 'max-w-lg w-full',
    xl: 'max-w-xl w-full',
    full: 'max-w-7xl w-full'
  }
  
  return sizeClasses[props.size]
})

const handleBackdropClick = () => {
  if (props.closeOnBackdrop) {
    emit('close')
  }
}
</script>