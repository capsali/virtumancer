<template>
  <FModal
    v-model="isOpen"
    :title="`Configure Hardware - ${vmName}`"
    size="full"
    @close="handleClose"
  >
    <div class="flex h-full">
      <!-- Hardware Categories Sidebar -->
      <div class="w-64 bg-gray-900/50 border-r border-white/10 p-4 overflow-y-auto">
        <div class="space-y-2">
          <div
            v-for="category in hardwareCategories"
            :key="category.id"
            :class="[
              'flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-all duration-200',
              {
                'bg-primary-500/20 text-primary-400 border border-primary-500/30': selectedCategory === category.id,
                'hover:bg-white/5 text-gray-300': selectedCategory !== category.id
              }
            ]"
            @click="selectedCategory = category.id"
          >
            <component :is="category.icon" class="w-5 h-5" />
            <div>
              <div class="font-medium">{{ category.label }}</div>
              <div class="text-xs text-gray-500">{{ category.count }} items</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Hardware Configuration Content -->
      <div class="flex-1 p-6 overflow-y-auto">
        <!-- Loading State -->
        <div v-if="loading" class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full mx-auto mb-4"></div>
            <p class="text-gray-400">Loading hardware configuration...</p>
          </div>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-8">
          <div class="text-red-400 mb-4">{{ error }}</div>
          <FButton @click="loadHardwareData" variant="outline" size="sm">
            Retry
          </FButton>
        </div>

        <!-- Hardware Configuration Panels -->
        <div v-else class="space-y-6">
          <!-- CPU Configuration -->
          <div v-if="selectedCategory === 'cpu'" class="space-y-6">
            <FCard class="p-6">
              <h3 class="text-lg font-semibold text-white mb-4">CPU Configuration</h3>
              
              <!-- Basic CPU Settings -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-2">vCPU Count</label>
                  <FInput
                    v-model.number="hardwareConfig.cpu.vcpuCount"
                    type="number"
                    :min="1"
                    :max="64"
                    class="w-full"
                  />
                </div>
                
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-2">CPU Model</label>
                  <FSelect
                    v-model="hardwareConfig.cpu.model"
                    :options="cpuModels"
                    class="w-full"
                  />
                </div>
              </div>

              <!-- CPU Topology -->
              <div class="mb-6">
                <h4 class="text-md font-medium text-white mb-3">CPU Topology</h4>
                <div class="grid grid-cols-3 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Sockets</label>
                    <FInput
                      v-model.number="hardwareConfig.cpu.topology.sockets"
                      type="number"
                      :min="1"
                      class="w-full"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Cores</label>
                    <FInput
                      v-model.number="hardwareConfig.cpu.topology.cores"
                      type="number"
                      :min="1"
                      class="w-full"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Threads</label>
                    <FInput
                      v-model.number="hardwareConfig.cpu.topology.threads"
                      type="number"
                      :min="1"
                      class="w-full"
                    />
                  </div>
                </div>
                <p class="text-xs text-gray-500 mt-2">
                  Total vCPUs: {{ hardwareConfig.cpu.topology.sockets * hardwareConfig.cpu.topology.cores * hardwareConfig.cpu.topology.threads }}
                </p>
              </div>

              <!-- CPU Features -->
              <div>
                <h4 class="text-md font-medium text-white mb-3">CPU Features</h4>
                <div class="space-y-2 max-h-48 overflow-y-auto">
                  <div
                    v-for="feature in hardwareConfig.cpu.features"
                    :key="feature.name"
                    class="flex items-center justify-between p-3 bg-gray-800 rounded"
                  >
                    <div>
                      <span class="text-white">{{ feature.name }}</span>
                      <span class="text-xs text-gray-400 ml-2">({{ feature.policy }})</span>
                    </div>
                    <div class="flex items-center space-x-2">
                      <FSelect
                        v-model="feature.policy"
                        :options="[
                          { value: 'require', label: 'Require' },
                          { value: 'optional', label: 'Optional' },
                          { value: 'disable', label: 'Disable' },
                          { value: 'forbid', label: 'Forbid' }
                        ]"
                        size="sm"
                        class="w-24"
                      />
                      <FButton @click="removeCPUFeature(feature.name)" variant="ghost" size="sm">
                        ×
                      </FButton>
                    </div>
                  </div>
                </div>
                <div class="mt-3">
                  <FButton @click="showAddCPUFeatureModal = true" variant="outline" size="sm">
                    Add CPU Feature
                  </FButton>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Memory Configuration -->
          <div v-if="selectedCategory === 'memory'" class="space-y-6">
            <FCard class="p-6">
              <h3 class="text-lg font-semibold text-white mb-4">Memory Configuration</h3>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-2">Maximum Memory (MB)</label>
                  <FInput
                    v-model.number="hardwareConfig.memory.maxMemoryMB"
                    type="number"
                    :min="128"
                    :step="128"
                    class="w-full"
                  />
                </div>
                
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-2">Current Memory (MB)</label>
                  <FInput
                    v-model.number="hardwareConfig.memory.currentMemoryMB"
                    type="number"
                    :min="128"
                    :max="hardwareConfig.memory.maxMemoryMB"
                    :step="128"
                    class="w-full"
                  />
                </div>
              </div>

              <!-- Memory Backing -->
              <div class="mb-6">
                <h4 class="text-md font-medium text-white mb-3">Memory Backing</h4>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Mode</label>
                    <FSelect
                      v-model="hardwareConfig.memory.backing.mode"
                      :options="[
                        { value: 'shared', label: 'Shared' },
                        { value: 'private', label: 'Private' }
                      ]"
                      class="w-full"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Source Type</label>
                    <FSelect
                      v-model="hardwareConfig.memory.backing.sourceType"
                      :options="[
                        { value: 'anonymous', label: 'Anonymous' },
                        { value: 'file', label: 'File' },
                        { value: 'memfd', label: 'MemFD' }
                      ]"
                      class="w-full"
                    />
                  </div>
                </div>
                
                <div class="mt-4 space-y-2">
                  <label class="flex items-center space-x-2">
                    <input
                      v-model="hardwareConfig.memory.backing.locked"
                      type="checkbox"
                      class="form-checkbox"
                    >
                    <span class="text-sm text-gray-300">Lock pages in memory</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input
                      v-model="hardwareConfig.memory.backing.nosharepages"
                      type="checkbox"
                      class="form-checkbox"
                    >
                    <span class="text-sm text-gray-300">Disable page sharing</span>
                  </label>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Storage Configuration -->
          <div v-if="selectedCategory === 'storage'" class="space-y-6">
            <FCard class="p-6">
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-white">Storage Devices</h3>
                <FButton @click="showAddDiskModal = true" variant="primary" size="sm">
                  Add Disk
                </FButton>
              </div>
              
              <div class="space-y-4">
                <div
                  v-for="(disk, index) in hardwareConfig.storage.disks"
                  :key="disk.id || index"
                  class="p-4 border border-gray-700 rounded-lg"
                >
                  <div class="flex items-center justify-between mb-3">
                    <h4 class="font-medium text-white">{{ disk.deviceName || `Disk ${index + 1}` }}</h4>
                    <FButton @click="removeDisk(index)" variant="ghost" size="sm">
                      Remove
                    </FButton>
                  </div>
                  
                  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Device Name</label>
                      <FInput v-model="disk.deviceName" placeholder="vda" class="w-full" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Bus Type</label>
                      <FSelect
                        v-model="disk.busType"
                        :options="[
                          { value: 'virtio', label: 'VirtIO' },
                          { value: 'sata', label: 'SATA' },
                          { value: 'ide', label: 'IDE' },
                          { value: 'scsi', label: 'SCSI' }
                        ]"
                        class="w-full"
                      />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Size (GB)</label>
                      <FInput v-model.number="disk.capacityGB" type="number" class="w-full" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Format</label>
                      <FSelect
                        v-model="disk.format"
                        :options="[
                          { value: 'qcow2', label: 'QCOW2' },
                          { value: 'raw', label: 'RAW' },
                          { value: 'vmdk', label: 'VMDK' }
                        ]"
                        class="w-full"
                      />
                    </div>
                  </div>
                  
                  <div class="mt-3 flex space-x-4">
                    <label class="flex items-center space-x-2">
                      <input v-model="disk.readOnly" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-gray-300">Read Only</span>
                    </label>
                    <label class="flex items-center space-x-2">
                      <input v-model="disk.shareable" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-gray-300">Shareable</span>
                    </label>
                  </div>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Network Configuration -->
          <div v-if="selectedCategory === 'network'" class="space-y-6">
            <FCard class="p-6">
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-white">Network Interfaces</h3>
                <FButton @click="showAddNetworkModal = true" variant="primary" size="sm">
                  Add Network Interface
                </FButton>
              </div>
              
              <div class="space-y-4">
                <div
                  v-for="(port, index) in hardwareConfig.network.ports"
                  :key="port.id || index"
                  class="p-4 border border-gray-700 rounded-lg"
                >
                  <div class="flex items-center justify-between mb-3">
                    <h4 class="font-medium text-white">{{ port.macAddress || `Interface ${index + 1}` }}</h4>
                    <FButton @click="removePort(index)" variant="ghost" size="sm">
                      Remove
                    </FButton>
                  </div>
                  
                  <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">MAC Address</label>
                      <FInput v-model="port.macAddress" placeholder="Auto-generate" class="w-full" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Model</label>
                      <FSelect
                        v-model="port.modelName"
                        :options="[
                          { value: 'virtio', label: 'VirtIO' },
                          { value: 'e1000', label: 'E1000' },
                          { value: 'rtl8139', label: 'RTL8139' }
                        ]"
                        class="w-full"
                      />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Source Type</label>
                      <FSelect
                        v-model="port.sourceType"
                        :options="[
                          { value: 'network', label: 'Network' },
                          { value: 'bridge', label: 'Bridge' },
                          { value: 'hostdev', label: 'Host Device' }
                        ]"
                        class="w-full"
                      />
                    </div>
                  </div>
                  
                  <div class="mt-3">
                    <label class="block text-sm font-medium text-gray-300 mb-1">Source Reference</label>
                    <FInput v-model="port.sourceRef" placeholder="Network or bridge name" class="w-full" />
                  </div>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Video/Graphics Configuration -->
          <div v-if="selectedCategory === 'video'" class="space-y-6">
            <FCard class="p-6">
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-white">Video Devices</h3>
                <FButton @click="showAddVideoModal = true" variant="primary" size="sm">
                  Add Video Device
                </FButton>
              </div>
              
              <div class="space-y-4">
                <div
                  v-for="(video, index) in hardwareConfig.video.devices"
                  :key="video.id || index"
                  class="p-4 border border-gray-700 rounded-lg"
                >
                  <div class="flex items-center justify-between mb-3">
                    <h4 class="font-medium text-white">Video Device {{ index + 1 }}</h4>
                    <FButton @click="removeVideoDevice(index)" variant="ghost" size="sm">
                      Remove
                    </FButton>
                  </div>
                  
                  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Model</label>
                      <FSelect
                        v-model="video.model"
                        :options="[
                          { value: 'qxl', label: 'QXL' },
                          { value: 'virtio', label: 'VirtIO-GPU' },
                          { value: 'vga', label: 'VGA' },
                          { value: 'cirrus', label: 'Cirrus' }
                        ]"
                        class="w-full"
                      />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">VRAM (MB)</label>
                      <FInput v-model.number="video.vram" type="number" class="w-full" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-300 mb-1">Heads</label>
                      <FInput v-model.number="video.heads" type="number" :min="1" :max="4" class="w-full" />
                    </div>
                    <div class="flex items-center">
                      <label class="flex items-center space-x-2">
                        <input v-model="video.accel3d" type="checkbox" class="form-checkbox">
                        <span class="text-sm text-gray-300">3D Acceleration</span>
                      </label>
                    </div>
                  </div>
                </div>
              </div>
            </FCard>
          </div>

          <!-- Advanced Hardware -->
          <div v-if="selectedCategory === 'advanced'" class="space-y-6">
            <!-- Controllers -->
            <FCard class="p-6">
              <h3 class="text-lg font-semibold text-white mb-4">Controllers</h3>
              <div class="space-y-3">
                <div
                  v-for="controller in hardwareConfig.advanced.controllers"
                  :key="controller.id"
                  class="flex items-center justify-between p-3 bg-gray-800 rounded"
                >
                  <div>
                    <span class="text-white">{{ controller.type }}</span>
                    <span class="text-xs text-gray-400 ml-2">{{ controller.model }}</span>
                  </div>
                  <FButton @click="removeController(controller.id)" variant="ghost" size="sm">
                    Remove
                  </FButton>
                </div>
              </div>
            </FCard>

            <!-- Host Devices -->
            <FCard class="p-6">
              <h3 class="text-lg font-semibold text-white mb-4">Host Device Passthrough</h3>
              <div class="space-y-3">
                <div
                  v-for="device in hardwareConfig.advanced.hostDevices"
                  :key="device.id"
                  class="flex items-center justify-between p-3 bg-gray-800 rounded"
                >
                  <div>
                    <span class="text-white">{{ device.name }}</span>
                    <span class="text-xs text-gray-400 ml-2">{{ device.address }}</span>
                  </div>
                  <FButton @click="removeHostDevice(device.id)" variant="ghost" size="sm">
                    Remove
                  </FButton>
                </div>
              </div>
            </FCard>

            <!-- TPM -->
            <FCard class="p-6">
              <h3 class="text-lg font-semibold text-white mb-4">TPM (Trusted Platform Module)</h3>
              <div class="space-y-3">
                <label class="flex items-center space-x-2">
                  <input
                    v-model="hardwareConfig.advanced.tpm.enabled"
                    type="checkbox"
                    class="form-checkbox"
                  >
                  <span class="text-sm text-gray-300">Enable TPM</span>
                </label>
                
                <div v-if="hardwareConfig.advanced.tpm.enabled" class="grid grid-cols-2 gap-4 mt-3">
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-1">Version</label>
                    <FSelect
                      v-model="hardwareConfig.advanced.tpm.version"
                      :options="[
                        { value: '1.2', label: 'TPM 1.2' },
                        { value: '2.0', label: 'TPM 2.0' }
                      ]"
                      class="w-full"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-300 mb-1">Backend</label>
                    <FSelect
                      v-model="hardwareConfig.advanced.tpm.backend"
                      :options="[
                        { value: 'emulator', label: 'Emulator' },
                        { value: 'passthrough', label: 'Passthrough' }
                      ]"
                      class="w-full"
                    />
                  </div>
                </div>
              </div>
            </FCard>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Footer -->
    <div class="flex justify-between p-6 border-t border-white/10 bg-gray-900/30">
      <div class="flex space-x-2">
        <FButton @click="exportConfiguration" variant="outline" size="sm">
          Export Config
        </FButton>
        <FButton @click="importConfiguration" variant="outline" size="sm">
          Import Config
        </FButton>
      </div>
      <div class="flex space-x-2">
        <FButton @click="handleClose" variant="ghost">
          Cancel
        </FButton>
        <FButton @click="resetToDefaults" variant="outline">
          Reset to Defaults
        </FButton>
        <FButton
          @click="saveConfiguration"
          variant="primary"
          :loading="saving"
          :disabled="!hasChanges"
        >
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </FButton>
      </div>
    </div>
  </FModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import FModal from './FModal.vue';
import FCard from './FCard.vue';
import FButton from './FButton.vue';
import FInput from './FInput.vue';
import FSelect from './FSelect.vue';
import { vmApi } from '@/services/api';
import { errorRecoveryService } from '@/services/errorRecovery';

// Icons (simplified for demo - in real app these would be proper icon components)
const CpuIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v4a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };
const MemoryIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };
const StorageIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };
const NetworkIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };
const VideoIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };
const AdvancedIcon = { template: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"></path></svg>' };

interface Props {
  hostId: string;
  vmName: string;
  modelValue: boolean;
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void;
  (e: 'configurationSaved'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// Reactive state
const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
});

const selectedCategory = ref<string>('cpu');
const loading = ref(true);
const saving = ref(false);
const error = ref<string | null>(null);

// Hardware configuration state
const hardwareConfig = ref({
  cpu: {
    vcpuCount: 2,
    model: 'host-model',
    topology: {
      sockets: 1,
      cores: 2,
      threads: 1
    },
    features: [] as Array<{ name: string; policy: string }>
  },
  memory: {
    maxMemoryMB: 2048,
    currentMemoryMB: 2048,
    backing: {
      mode: 'shared',
      sourceType: 'anonymous',
      locked: false,
      nosharepages: false
    }
  },
  storage: {
    disks: [] as Array<{
      id?: string;
      deviceName: string;
      busType: string;
      capacityGB: number;
      format: string;
      readOnly: boolean;
      shareable: boolean;
    }>
  },
  network: {
    ports: [] as Array<{
      id?: string;
      macAddress: string;
      modelName: string;
      sourceType: string;
      sourceRef: string;
    }>
  },
  video: {
    devices: [] as Array<{
      id?: string;
      model: string;
      vram: number;
      heads: number;
      accel3d: boolean;
    }>
  },
  advanced: {
    controllers: [] as Array<{
      id: string;
      type: string;
      model: string;
    }>,
    hostDevices: [] as Array<{
      id: string;
      name: string;
      address: string;
    }>,
    tpm: {
      enabled: false,
      version: '2.0',
      backend: 'emulator'
    }
  }
});

const originalConfig = ref<string>('');

// Hardware categories for sidebar
const hardwareCategories = computed(() => [
  {
    id: 'cpu',
    label: 'CPU',
    icon: CpuIcon,
    count: 1 + hardwareConfig.value.cpu.features.length
  },
  {
    id: 'memory',
    label: 'Memory',
    icon: MemoryIcon,
    count: 1
  },
  {
    id: 'storage',
    label: 'Storage',
    icon: StorageIcon,
    count: hardwareConfig.value.storage.disks.length
  },
  {
    id: 'network',
    label: 'Network',
    icon: NetworkIcon,
    count: hardwareConfig.value.network.ports.length
  },
  {
    id: 'video',
    label: 'Video/Graphics',
    icon: VideoIcon,
    count: hardwareConfig.value.video.devices.length
  },
  {
    id: 'advanced',
    label: 'Advanced',
    icon: AdvancedIcon,
    count: hardwareConfig.value.advanced.controllers.length + hardwareConfig.value.advanced.hostDevices.length + (hardwareConfig.value.advanced.tpm.enabled ? 1 : 0)
  }
]);

// Available options
const cpuModels = [
  { value: 'host-model', label: 'Host Model' },
  { value: 'host-passthrough', label: 'Host Passthrough' },
  { value: 'qemu64', label: 'QEMU64' },
  { value: 'Haswell', label: 'Intel Haswell' },
  { value: 'Skylake-Server', label: 'Intel Skylake' },
  { value: 'EPYC', label: 'AMD EPYC' }
];

// Computed properties
const hasChanges = computed(() => {
  return JSON.stringify(hardwareConfig.value) !== originalConfig.value;
});

// Add modal states
const showAddCPUFeatureModal = ref(false);
const showAddDiskModal = ref(false);
const showAddNetworkModal = ref(false);
const showAddVideoModal = ref(false);

// Methods
const loadHardwareData = async () => {
  try {
    loading.value = true;
    error.value = null;
    
    const hardware = await vmApi.getHardware(props.hostId, props.vmName);
    
    // Map the API response to our internal structure
    hardwareConfig.value = {
      cpu: {
        vcpuCount: hardware.vcpus || 2,
        model: hardware.cpu_model || 'host-model',
        topology: {
          sockets: hardware.cpu_topology?.sockets || 1,
          cores: hardware.cpu_topology?.cores || 2,
          threads: hardware.cpu_topology?.threads || 1
        },
        features: hardware.cpu_features || []
      },
      memory: {
        maxMemoryMB: hardware.memory_bytes ? Math.round(hardware.memory_bytes / 1048576) : hardware.memory_mb || 2048,
        currentMemoryMB: hardware.current_memory ? Math.round(hardware.current_memory / 1048576) : hardware.memory_mb || 2048,
        backing: hardware.memory_backing || {
          mode: 'shared',
          sourceType: 'anonymous',
          locked: false,
          nosharepages: false
        }
      },
      storage: {
        disks: (hardware.disks || []).map(disk => ({
          id: disk.id,
          deviceName: disk.deviceName || disk.device || disk.target,
          busType: disk.busType || disk.type || 'virtio',
          capacityGB: disk.capacityGB || disk.size_gb || 20,
          format: disk.format || 'qcow2',
          readOnly: disk.readOnly || false,
          shareable: disk.shareable || false
        }))
      },
      network: {
        ports: (hardware.networks || []).map(network => ({
          id: network.id,
          macAddress: network.macAddress || network.mac || '',
          modelName: network.modelName || network.model || 'virtio',
          sourceType: network.sourceType || network.type || 'network',
          sourceRef: network.sourceRef || network.source || ''
        }))
      },
      video: {
        devices: hardware.video_devices || []
      },
      advanced: {
        controllers: hardware.controllers || [],
        hostDevices: hardware.host_devices || [],
        tpm: hardware.tpm || {
          enabled: false,
          version: '2.0',
          backend: 'emulator'
        }
      }
    };
    
    originalConfig.value = JSON.stringify(hardwareConfig.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load hardware configuration';
    errorRecoveryService.addError(
      err as Error,
      `load_vm_hardware_${props.vmName}`,
      { hostId: props.hostId, vmName: props.vmName }
    );
  } finally {
    loading.value = false;
  }
};

const saveConfiguration = async () => {
  try {
    saving.value = true;
    
    await vmApi.updateHardware(props.hostId, props.vmName, hardwareConfig.value);
    
    originalConfig.value = JSON.stringify(hardwareConfig.value);
    emit('configurationSaved');
    isOpen.value = false;
  } catch (err) {
    errorRecoveryService.addError(
      err as Error,
      `save_vm_hardware_${props.vmName}`,
      { hostId: props.hostId, vmName: props.vmName }
    );
  } finally {
    saving.value = false;
  }
};

const resetToDefaults = () => {
  hardwareConfig.value = JSON.parse(originalConfig.value);
};

const handleClose = () => {
  if (hasChanges.value) {
    if (confirm('You have unsaved changes. Are you sure you want to close?')) {
      resetToDefaults();
      isOpen.value = false;
    }
  } else {
    isOpen.value = false;
  }
};

// Hardware management methods
const removeCPUFeature = (featureName: string) => {
  const index = hardwareConfig.value.cpu.features.findIndex(f => f.name === featureName);
  if (index >= 0) {
    hardwareConfig.value.cpu.features.splice(index, 1);
  }
};

const removeDisk = (index: number) => {
  hardwareConfig.value.storage.disks.splice(index, 1);
};

const removePort = (index: number) => {
  hardwareConfig.value.network.ports.splice(index, 1);
};

const removeVideoDevice = (index: number) => {
  hardwareConfig.value.video.devices.splice(index, 1);
};

const removeController = (id: string) => {
  const index = hardwareConfig.value.advanced.controllers.findIndex(c => c.id === id);
  if (index >= 0) {
    hardwareConfig.value.advanced.controllers.splice(index, 1);
  }
};

const removeHostDevice = (id: string) => {
  const index = hardwareConfig.value.advanced.hostDevices.findIndex(d => d.id === id);
  if (index >= 0) {
    hardwareConfig.value.advanced.hostDevices.splice(index, 1);
  }
};

const exportConfiguration = () => {
  const config = JSON.stringify(hardwareConfig.value, null, 2);
  const blob = new Blob([config], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${props.vmName}-hardware-config.json`;
  a.click();
  URL.revokeObjectURL(url);
};

const importConfiguration = () => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'application/json';
  input.onchange = (event) => {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const config = JSON.parse(e.target?.result as string);
          hardwareConfig.value = config;
        } catch (err) {
          alert('Invalid configuration file');
        }
      };
      reader.readAsText(file);
    }
  };
  input.click();
};

// Watch for modal open state and load data
watch(isOpen, (open) => {
  if (open) {
    loadHardwareData();
  }
});

onMounted(() => {
  if (isOpen.value) {
    loadHardwareData();
  }
});
</script>