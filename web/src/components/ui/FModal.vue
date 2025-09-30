<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[50000] overflow-y-auto">
      <!-- Enhanced Backdrop with Glass Effect -->
      <div 
        class="fixed inset-0 bg-black/60 backdrop-blur-sm transition-all duration-300"
        @click="handleBackdropClick"
      ></div>
      
      <!-- Modal content -->
      <div class="flex min-h-full items-center justify-center p-4">
        <div 
          :class="modalClasses"
          class="relative glass-strong modal-glow rounded-xl shadow-2xl border border-white/10 max-h-full overflow-hidden animate-modal-in"
        >
          <!-- Close button -->
          <button
            v-if="showCloseButton"
            @click="$emit('close')"
            class="absolute top-4 right-4 z-10 p-2 rounded-lg glass-subtle hover:glass-medium text-slate-400 hover:text-white transition-all duration-200 hover:scale-110"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          
          <!-- Modal body -->
          <div :class="paddingClasses">
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
  padding?: 'sm' | 'md' | 'lg'
  showCloseButton?: boolean
  closeOnBackdrop?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  padding: 'lg',
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

const paddingClasses = computed(() => {
  const basePadding = 'overflow-y-auto max-h-[90vh]'
  const paddingMap = {
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8'
  }
  
  return `${basePadding} ${paddingMap[props.padding]}`
})

const handleBackdropClick = () => {
  if (props.closeOnBackdrop) {
    emit('close')
  }
}
</script>