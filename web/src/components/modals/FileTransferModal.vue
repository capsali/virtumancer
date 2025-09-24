<template>
  <teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center z-50 p-4"
      @click="handleBackdropClick"
    >
      <div
        ref="modalRef"
        class="bg-slate-900/95 border border-slate-600/50 rounded-xl shadow-2xl backdrop-blur-xl modal-glow max-w-md w-full mx-4"
        @click.stop
      >
        <!-- Header -->
        <div class="p-6 border-b border-slate-600/30">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center shadow-glow">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
                </svg>
              </div>
              <div>
                <h3 class="text-lg font-semibold text-white">File Transfer</h3>
                <p class="text-sm text-slate-400">Transfer files to {{ vmName }}</p>
              </div>
            </div>
            <button
              @click="$emit('close')"
              class="text-slate-400 hover:text-white transition-colors p-2 hover:bg-slate-700/50 rounded-lg"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Content -->
        <div class="p-6 space-y-4">
          <!-- Drag and Drop Area -->
          <div
            ref="dropZone"
            :class="[
              'border-2 border-dashed rounded-lg p-8 text-center transition-all duration-300',
              isDragging
                ? 'border-blue-400 bg-blue-500/10 text-blue-400'
                : 'border-slate-600 hover:border-slate-500 text-slate-400 hover:text-slate-300'
            ]"
            @dragover.prevent="handleDragOver"
            @dragleave.prevent="handleDragLeave"
            @drop.prevent="handleDrop"
          >
            <div class="flex flex-col items-center gap-3">
              <div :class="[
                'w-12 h-12 rounded-full flex items-center justify-center transition-all duration-300',
                isDragging ? 'bg-blue-500/20 text-blue-400' : 'bg-slate-700/50 text-slate-400'
              ]">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 12l2 2 4-4"/>
                </svg>
              </div>
              <div>
                <p class="font-medium">{{ isDragging ? 'Drop files here' : 'Drag files here to transfer' }}</p>
                <p class="text-sm text-slate-500 mt-1">or</p>
              </div>
            </div>
          </div>

          <!-- Browse Button -->
          <div class="text-center">
            <input
              ref="fileInput"
              type="file"
              multiple
              class="hidden"
              @change="handleFileSelect"
            />
            <FButton
              variant="outline"
              @click="() => fileInput?.click()"
              class="w-full button-glow"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5a2 2 0 012-2h4a2 2 0 012 2v6H8V5z"/>
              </svg>
              Browse Files
            </FButton>
          </div>

          <!-- Selected Files -->
          <div v-if="selectedFiles.length > 0" class="space-y-2">
            <p class="text-sm font-medium text-white">Selected Files:</p>
            <div class="max-h-32 overflow-y-auto space-y-1">
              <div
                v-for="(file, index) in selectedFiles"
                :key="index"
                class="flex items-center justify-between p-2 bg-slate-800/50 rounded-lg text-sm"
              >
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <svg class="w-4 h-4 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                  </svg>
                  <span class="text-white truncate">{{ file.name }}</span>
                  <span class="text-slate-400 text-xs flex-shrink-0">({{ formatFileSize(file.size) }})</span>
                </div>
                <button
                  @click="removeFile(index)"
                  class="text-slate-400 hover:text-red-400 transition-colors p-1"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Transfer Status -->
          <div v-if="transferStatus" class="p-3 rounded-lg" :class="transferStatus.type === 'error' ? 'bg-red-500/10 border border-red-500/20' : 'bg-blue-500/10 border border-blue-500/20'">
            <p class="text-sm" :class="transferStatus.type === 'error' ? 'text-red-400' : 'text-blue-400'">
              {{ transferStatus.message }}
            </p>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-6 border-t border-slate-600/30 flex justify-end gap-3">
          <FButton variant="ghost" @click="$emit('close')">
            Cancel
          </FButton>
          <FButton
            variant="primary"
            @click="handleTransfer"
            :disabled="selectedFiles.length === 0 || isTransferring"
            class="button-glow"
          >
            <svg v-if="isTransferring" class="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span v-if="isTransferring">Transferring...</span>
            <span v-else>Transfer Files ({{ selectedFiles.length }})</span>
          </FButton>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import FButton from '@/components/ui/FButton.vue'

interface Props {
  show: boolean
  vmName: string
}

interface Emits {
  (e: 'close'): void
  (e: 'transfer', files: File[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const modalRef = ref<HTMLElement>()
const dropZone = ref<HTMLElement>()
const fileInput = ref<HTMLInputElement>()
const selectedFiles = ref<File[]>([])
const isDragging = ref(false)
const isTransferring = ref(false)
const transferStatus = ref<{ type: 'success' | 'error', message: string } | null>(null)

// File handling
const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files) {
    selectedFiles.value = [...selectedFiles.value, ...Array.from(target.files)]
    target.value = '' // Reset input
  }
}

const handleDrop = (event: DragEvent) => {
  isDragging.value = false
  if (event.dataTransfer?.files) {
    selectedFiles.value = [...selectedFiles.value, ...Array.from(event.dataTransfer.files)]
  }
}

const handleDragOver = (event: DragEvent) => {
  isDragging.value = true
}

const handleDragLeave = (event: DragEvent) => {
  // Only set isDragging to false if we're leaving the dropZone itself
  if (!dropZone.value?.contains(event.relatedTarget as Node)) {
    isDragging.value = false
  }
}

const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1)
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const handleTransfer = async () => {
  if (selectedFiles.value.length === 0) return
  
  isTransferring.value = true
  transferStatus.value = null
  
  try {
    // Emit transfer event - parent component should handle the actual file transfer
    emit('transfer', selectedFiles.value)
    
    transferStatus.value = {
      type: 'success',
      message: `Successfully initiated transfer of ${selectedFiles.value.length} file(s)`
    }
    
    // Clear files after successful transfer
    setTimeout(() => {
      selectedFiles.value = []
      transferStatus.value = null
      emit('close')
    }, 2000)
    
  } catch (error) {
    console.error('File transfer error:', error)
    transferStatus.value = {
      type: 'error',
      message: 'Failed to transfer files. Please try again.'
    }
  } finally {
    isTransferring.value = false
  }
}

// Modal handling
const handleBackdropClick = () => {
  emit('close')
}

// Cleanup when modal is closed
const cleanup = () => {
  selectedFiles.value = []
  isDragging.value = false
  isTransferring.value = false
  transferStatus.value = null
}

// Watch for show changes to cleanup
const handleShowChange = () => {
  if (!props.show) {
    cleanup()
  }
}

// ESC key handling
const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && props.show) {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>