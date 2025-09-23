<template>
  <div class="space-y-4">
    <!-- Selection Header -->
    <div class="flex items-center justify-between p-4 bg-white/5 rounded-lg border border-white/10">
      <div class="flex items-center gap-3">
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            :checked="isAllSelected"
            :indeterminate="isPartiallySelected"
            @change="toggleSelectAll"
            class="w-4 h-4 rounded border-white/20 bg-white/10 text-primary-500 focus:ring-primary-500 focus:ring-offset-0"
          />
          <span class="text-white text-sm">
            {{ selectedCount === 0 ? 'Select All' : 
               selectedCount === totalCount ? `All ${totalCount} VMs selected` :
               `${selectedCount} of ${totalCount} VMs selected` }}
          </span>
        </label>
      </div>

      <!-- Bulk Actions -->
      <div v-if="selectedCount > 0" class="flex items-center gap-2">
        <FButton
          variant="primary"
          size="sm"
          @click="$emit('bulk-import', Array.from(selectedVMUUIDs))"
          :disabled="importing"
        >
          <span v-if="!importing">📥 Import Selected ({{ selectedCount }})</span>
          <span v-else class="flex items-center gap-2">
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Importing...
          </span>
        </FButton>
        
        <FButton
          variant="ghost"
          size="sm"
          @click="$emit('bulk-delete', Array.from(selectedVMUUIDs))"
          :disabled="deleting"
        >
          <span v-if="!deleting">🗑️ Remove Selected</span>
          <span v-else class="flex items-center gap-2">
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Removing...
          </span>
        </FButton>

        <FButton
          variant="ghost"
          size="sm"
          @click="clearSelection"
        >
          Clear Selection
        </FButton>
      </div>
    </div>

    <!-- Search and Filter -->
    <div class="flex items-center gap-4">
      <div class="flex-1 relative">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search discovered VMs..."
          class="w-full px-4 py-2 pl-10 bg-white/5 border border-white/10 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
        <div class="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400">
          🔍
        </div>
      </div>

      <select
        v-model="sortOrder"
        class="px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
      >
        <option value="name-asc">Name A-Z</option>
        <option value="name-desc">Name Z-A</option>
        <option value="date-desc">Latest First</option>
        <option value="date-asc">Oldest First</option>
      </select>
    </div>

    <!-- VM List -->
    <div class="space-y-3">
      <div
        v-for="vm in filteredAndSortedVMs"
        :key="vm.domain_uuid"
        class="flex items-center gap-3"
      >
        <label class="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            :checked="isVMSelected(vm.domain_uuid)"
            @change="toggleVMSelection(vm.domain_uuid)"
            class="w-4 h-4 rounded border-white/20 bg-white/10 text-primary-500 focus:ring-primary-500 focus:ring-offset-0"
          />
        </label>
        
        <div class="flex-1">
          <DiscoveredVMCard
            :vm="vm"
            :host-id="hostId"
            :importing="importing && isVMSelected(vm.uuid)"
            @import="handleSingleImport"
          />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredAndSortedVMs.length === 0 && vms.length > 0" class="text-center py-8">
      <p class="text-slate-400">No VMs match your search criteria.</p>
      <FButton variant="ghost" size="sm" @click="clearSearch" class="mt-2">
        Clear Search
      </FButton>
    </div>

    <div v-if="vms.length === 0" class="text-center py-8 text-slate-400">
      No discovered VMs found.
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import FButton from '@/components/ui/FButton.vue';
import DiscoveredVMCard from '@/components/vm/DiscoveredVMCard.vue';
import type { DiscoveredVM } from '@/types';

interface Props {
  vms: DiscoveredVM[];
  hostId: string;
  importing?: boolean;
  deleting?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'bulk-import': [vmUUIDs: string[]];
  'bulk-delete': [vmUUIDs: string[]];
  'single-import': [hostId: string, vmName: string];
}>();

// Selection state
const selectedVMUUIDs = ref<Set<string>>(new Set());
const searchQuery = ref('');
const sortOrder = ref<'name-asc' | 'name-desc' | 'date-desc' | 'date-asc'>('name-asc');

// Computed properties
const totalCount = computed(() => props.vms?.length || 0);
const selectedCount = computed(() => selectedVMUUIDs.value.size);

const isAllSelected = computed(() => {
  return totalCount.value > 0 && selectedCount.value === totalCount.value;
});

const isPartiallySelected = computed(() => {
  return selectedCount.value > 0 && selectedCount.value < totalCount.value;
});

const filteredAndSortedVMs = computed(() => {
  let filtered = props.vms || [];

  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(vm => 
      vm && vm.name && vm.domain_uuid &&
      (vm.name.toLowerCase().includes(query) ||
      vm.domain_uuid.toLowerCase().includes(query))
    );
  }

  // Apply sorting
  return filtered.sort((a, b) => {
    if (!a || !b) return 0;
    switch (sortOrder.value) {
      case 'name-asc':
        return (a.name || '').localeCompare(b.name || '');
      case 'name-desc':
        return (b.name || '').localeCompare(a.name || '');
      case 'date-desc':
        return new Date(b.last_seen_at || 0).getTime() - new Date(a.last_seen_at || 0).getTime();
      case 'date-asc':
        return new Date(a.last_seen_at || 0).getTime() - new Date(b.last_seen_at || 0).getTime();
      default:
        return 0;
    }
  });
});

// Selection methods
const isVMSelected = (uuid: string): boolean => {
  return selectedVMUUIDs.value.has(uuid);
};

const toggleVMSelection = (uuid: string): void => {
  if (selectedVMUUIDs.value.has(uuid)) {
    selectedVMUUIDs.value.delete(uuid);
  } else {
    selectedVMUUIDs.value.add(uuid);
  }
};

const toggleSelectAll = (): void => {
  if (isAllSelected.value) {
    selectedVMUUIDs.value.clear();
  } else {
    selectedVMUUIDs.value = new Set(filteredAndSortedVMs.value
      .filter(vm => vm && vm.domain_uuid)
      .map(vm => vm.domain_uuid));
  }
};

const clearSelection = (): void => {
  selectedVMUUIDs.value.clear();
};

const clearSearch = (): void => {
  searchQuery.value = '';
};

const handleSingleImport = (hostId: string, vmName: string): void => {
  emit('single-import', hostId, vmName);
};
</script>