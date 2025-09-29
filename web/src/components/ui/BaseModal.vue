<template>
  <teleport to="body">
    <transition name="modal-fade" appear>
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click="handleBackdropClick"
      >
        <!-- Backdrop -->
        <div class="fixed inset-0 bg-black/60 backdrop-blur-sm"></div>
        
        <!-- Modal -->
        <FCard
          ref="modalRef"
          :class="[
            'relative w-full glass-medium border border-white/10 modal-glow',
            modalSizeClass
          ]"
          @click.stop
        >
          <div ref="trapRef" class="space-y-6">
            <!-- Header -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <!-- Icon -->
                <div
                  v-if="icon"
                  :class="[
                    'w-10 h-10 rounded-lg flex items-center justify-center shadow-glow',
                    iconClass
                  ]"
                >
                  <component :is="icon" class="w-5 h-5 text-white" />
                </div>
                
                <!-- Title and Description -->
                <div>
                  <h3 class="text-lg font-semibold text-white">{{ title }}</h3>
                  <p v-if="description" class="text-sm text-slate-400">{{ description }}</p>
                </div>
              </div>
              
              <!-- Close Button -->
              <button
                @click="$emit('close')"
                class="text-slate-400 hover:text-white transition-colors p-2 hover:bg-slate-700/50 rounded-lg"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            
            <!-- Content Slot -->
            <div class="space-y-4">
              <slot></slot>
            </div>
            
            <!-- Footer Actions -->
            <div v-if="$slots.actions || showDefaultActions" class="flex justify-end gap-3 pt-4 border-t border-slate-600/30">
              <slot name="actions">
                <!-- Default Actions -->
                <template v-if="showDefaultActions">
                  <FButton
                    variant="ghost"
                    @click="handleCancel"
                    :disabled="cancelDisabled"
                    class="button-glow cancel"
                    tabindex="98"
                  >
                    {{ cancelText }}
                  </FButton>
                  <FButton
                    :variant="confirmVariant"
                    @click="$emit('confirm')"
                    :disabled="confirmDisabled"
                    class="button-glow confirm"
                    tabindex="99"
                  >
                    {{ confirmText }}
                  </FButton>
                </template>
              </slot>
            </div>
          </div>
        </FCard>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import FCard from '@/components/ui/FCard.vue'
import FButton from '@/components/ui/FButton.vue'
import { useFocusTrap } from '@/composables/useFocusTrap'
import { useModalKeyboard } from '@/composables/useModalKeyboard'

interface Props {
  show: boolean
  title: string
  description?: string
  icon?: any
  iconClass?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  showDefaultActions?: boolean
  cancelText?: string
  confirmText?: string
  confirmVariant?: 'primary' | 'danger' | 'outline' | 'accent' | 'secondary' | 'ghost' | 'neon'
  confirmDisabled?: boolean
  cancelDisabled?: boolean
  closeOnBackdrop?: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'cancel'): void
  (e: 'confirm'): void
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  showDefaultActions: true,
  cancelText: 'Cancel',
  confirmText: 'Confirm',
  confirmVariant: 'primary',
  confirmDisabled: false,
  closeOnBackdrop: true,
  iconClass: 'bg-gradient-to-br from-primary-500 to-accent-500'
})

const emit = defineEmits<Emits>()

// Focus trap and keyboard navigation
const { trapRef, startTrap, stopTrap } = useFocusTrap()
const { createKeyboardHandler, getStandardModalHandlers } = useModalKeyboard()

const handleCancel = () => {
  // Emit both cancel and close to ensure parent components that only listen for close still close the modal
  emit('cancel')
  emit('close')
}

// Modal reference
const modalRef = ref<HTMLElement>()

// Keyboard handler
const handleKeydown = createKeyboardHandler(
  getStandardModalHandlers(
    () => emit('close'),
    () => emit('confirm')
  )
)

// Watch for show changes to manage focus trap and keyboard
watch(() => props.show, (isVisible) => {
  if (isVisible) {
    startTrap()
    document.addEventListener('keydown', handleKeydown)
  } else {
    stopTrap()
    document.removeEventListener('keydown', handleKeydown)
  }
})

// Cleanup on unmount
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

// Computed properties
const modalSizeClass = computed(() => {
  switch (props.size) {
    case 'sm': return 'max-w-sm'
    case 'md': return 'max-w-md'
    case 'lg': return 'max-w-lg'
    case 'xl': return 'max-w-xl'
    default: return 'max-w-md'
  }
})

const handleBackdropClick = () => {
  if (props.closeOnBackdrop) {
    emit('close')
  }
}
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 0.2s ease-in-out;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>