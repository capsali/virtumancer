import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { StoragePool, StorageVolume, DiskAttachment, StorageStats } from '@/types'
import { storageApi } from '@/services/api'

export const useStorageStore = defineStore('storage', () => {
  // State
  const storagePools = ref<StoragePool[]>([])
  const storageVolumes = ref<StorageVolume[]>([])
  const diskAttachments = ref<DiskAttachment[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

// Getters
const totalPools = computed(() => storagePools.value.length)
const activePools = computed(() =>
  storagePools.value.length // For now, assume all pools are "active" since we don't have state info
)
const totalVolumes = computed(() => storageVolumes.value.length)
const attachedDisks = computed(() => diskAttachments.value.length)

const totalCapacity = computed(() => {
  return storagePools.value.reduce((total, pool) => total + pool.capacity_bytes, 0)
})

const usedSpace = computed(() => {
  return storagePools.value.reduce((total, pool) => total + pool.allocation_bytes, 0)
})

const availableSpace = computed(() => {
  return storagePools.value.reduce((total, pool) => total + (pool.capacity_bytes - pool.allocation_bytes), 0)
})

const utilization = computed(() => {
  const total = totalCapacity.value
  const used = usedSpace.value
  return total > 0 ? Math.round((used / total) * 100) : 0
})

const storageStats = computed<StorageStats>(() => ({
  totalPools: totalPools.value,
  activePools: activePools.value,
  totalVolumes: totalVolumes.value,
  attachedDisks: attachedDisks.value,
  totalCapacity: totalCapacity.value,
  usedSpace: usedSpace.value,
  availableSpace: availableSpace.value,
  utilization: utilization.value
}))

  // Actions
  const fetchStoragePools = async () => {
    try {
      isLoading.value = true
      error.value = null
      storagePools.value = await storageApi.getStoragePools()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch storage pools'
      console.error('Failed to fetch storage pools:', err)
    } finally {
      isLoading.value = false
    }
  }

  const fetchStorageVolumes = async () => {
    try {
      isLoading.value = true
      error.value = null
      storageVolumes.value = await storageApi.getStorageVolumes()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch storage volumes'
      console.error('Failed to fetch storage volumes:', err)
    } finally {
      isLoading.value = false
    }
  }

  const fetchDiskAttachments = async () => {
    try {
      isLoading.value = true
      error.value = null
      diskAttachments.value = await storageApi.getDiskAttachments()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch disk attachments'
      console.error('Failed to fetch disk attachments:', err)
    } finally {
      isLoading.value = false
    }
  }

  const fetchAllStorageData = async () => {
    await Promise.all([
      fetchStoragePools(),
      fetchStorageVolumes(),
      fetchDiskAttachments()
    ])
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // State
    storagePools,
    storageVolumes,
    diskAttachments,
    isLoading,
    error,

    // Getters
    storageStats,

    // Actions
    fetchStoragePools,
    fetchStorageVolumes,
    fetchDiskAttachments,
    fetchAllStorageData,
    clearError
  }
})