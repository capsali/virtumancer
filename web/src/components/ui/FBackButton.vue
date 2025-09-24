<template>
  <div class="flex items-center gap-3">
    <!-- Back Button -->
    <button
      @click="handleBack"
      class="flex items-center gap-2 px-3 py-1.5 text-sm text-slate-300 hover:text-white hover:bg-slate-800/30 rounded-md transition-colors duration-200"
      :title="backTooltip"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
      </svg>
      <span v-if="!compact">{{ backLabel }}</span>
    </button>

    <!-- Context Actions (if any) -->
    <div v-if="contextActions?.length" class="flex items-center gap-1">
      <div class="w-px h-4 bg-slate-600 mx-1"></div>
      <button
        v-for="action in contextActions"
        :key="action.label"
        @click="action.action"
        :disabled="action.disabled"
        class="flex items-center gap-1 px-2 py-1 text-xs text-slate-400 hover:text-white hover:bg-slate-800/30 rounded transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
        :title="action.tooltip"
      >
        <span v-if="action.icon" v-html="action.icon"></span>
        <span>{{ action.label }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useBreadcrumbs } from '@/composables/useBreadcrumbs'

export interface ContextAction {
  label: string
  action: () => void
  icon?: string
  disabled?: boolean
  tooltip?: string
}

interface Props {
  compact?: boolean
  contextActions?: ContextAction[]
}

const props = withDefaults(defineProps<Props>(), {
  compact: false,
  contextActions: () => []
})

const { breadcrumbs, navigateBack } = useBreadcrumbs()

const backLabel = computed(() => {
  const parentItem = breadcrumbs.value[breadcrumbs.value.length - 2]
  if (parentItem) {
    return `Back to ${parentItem.label}`
  }
  return 'Back'
})

const backTooltip = computed(() => {
  const parentItem = breadcrumbs.value[breadcrumbs.value.length - 2]
  if (parentItem) {
    return `Return to ${parentItem.label}`
  }
  return 'Go back to previous page'
})

const handleBack = () => {
  navigateBack()
}
</script>