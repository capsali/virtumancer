<template>
  <nav class="flex items-center space-x-1 text-sm">
    <template v-for="(item, index) in breadcrumbs" :key="`${item.label}-${index}`">
      <!-- Separator -->
      <div v-if="index > 0" class="text-slate-400 mx-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
      </div>

      <!-- Breadcrumb Item -->
      <div class="flex items-center">
        <!-- Interactive breadcrumb -->
        <button
          v-if="item.path && !item.isActive"
          @click="navigateTo(item.path)"
          class="flex items-center gap-1.5 px-2 py-1 rounded-md hover:bg-slate-800/30 transition-colors duration-200 text-slate-300 hover:text-white"
        >
          <svg v-if="item.icon" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon"/>
          </svg>
          <span>{{ item.label }}</span>
        </button>

        <!-- Active breadcrumb (current page) -->
        <div
          v-else
          class="flex items-center gap-1.5 px-2 py-1 text-white font-medium"
        >
          <svg v-if="item.icon" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon"/>
          </svg>
          <span>{{ item.label }}</span>
        </div>
      </div>
    </template>
  </nav>
</template>

<script setup lang="ts">
import { useBreadcrumbs } from '@/composables/useBreadcrumbs'

const { breadcrumbs, navigateTo } = useBreadcrumbs()
</script>