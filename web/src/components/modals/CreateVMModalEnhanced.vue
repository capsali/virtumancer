<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @click.self="close"
    >
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>
      
      <!-- Modal -->
      <FCard
        class="relative w-full max-w-6xl glass-medium border border-white/10 modal-glow"
      >
        <div class="space-y-6 max-h-[90vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex items-center justify-between p-6 border-b border-white/10 sticky top-0 bg-slate-900/80 backdrop-blur-sm z-10">
            <h2 class="text-xl font-semibold text-white">Create Virtual Machine</h2>
            <div class="flex items-center gap-3">
              <span class="text-sm text-slate-400">Step {{ currentStep }} of {{ totalSteps }}</span>
              <FButton
                size="sm"
                variant="ghost"
                @click="close"
              >
                ✕
              </FButton>
            </div>
          </div>

          <!-- Progress Bar -->
          <div class="px-6">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-slate-300">Configuration Progress</span>
              <span class="text-sm text-slate-400">{{ Math.round((currentStep / totalSteps) * 100) }}%</span>
            </div>
            <div class="w-full bg-slate-700 rounded-full h-2">
              <div 
                class="bg-gradient-to-r from-blue-500 to-purple-500 h-2 rounded-full transition-all duration-300"
                :style="{ width: `${(currentStep / totalSteps) * 100}%` }"
              ></div>
            </div>
          </div>

          <!-- Form -->
          <form @submit.prevent="handleSubmit" class="p-6 space-y-8">
            <!-- Step Navigation -->
            <div class="flex flex-wrap gap-2 border-b border-slate-700 pb-4">
              <button
                v-for="(step, index) in steps"
                :key="step.id"
                type="button"
                @click="currentStep = index + 1"
                :class="[
                  'px-4 py-2 rounded-lg text-sm font-medium transition-all',
                  currentStep === index + 1
                    ? 'bg-blue-600 text-white shadow-lg'
                    : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                ]"
              >
                <span class="flex items-center gap-2">
                  <span class="text-lg">{{ step.icon }}</span>
                  {{ step.name }}
                </span>
              </button>
            </div>

            <!-- Step 1: Basic Configuration -->
            <div v-if="currentStep === 1" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">Basic Configuration</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    VM Name *
                  </label>
                  <input
                    v-model="formData.name"
                    type="text"
                    placeholder="my-virtual-machine"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Host Selection *
                  </label>
                  <select
                    v-model="formData.hostId"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  >
                    <option value="">Select Host</option>
                    <option v-for="host in availableHosts" :key="host.id" :value="host.id">
                      {{ host.name || host.uri }}
                    </option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    OS Type
                  </label>
                  <select
                    v-model="formData.osType"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="linux">Linux</option>
                    <option value="windows">Windows</option>
                    <option value="macos">macOS</option>
                    <option value="freebsd">FreeBSD</option>
                    <option value="other">Other</option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Architecture
                  </label>
                  <select
                    v-model="formData.architecture"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="x86_64">x86_64 (64-bit)</option>
                    <option value="i686">i686 (32-bit)</option>
                    <option value="aarch64">ARM64</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-white mb-2">
                  Description
                </label>
                <textarea
                  v-model="formData.description"
                  placeholder="Optional description for this VM"
                  rows="3"
                  class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all resize-none"
                ></textarea>
              </div>
            </div>

            <!-- Step 2: CPU Configuration -->
            <div v-if="currentStep === 2" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-purple-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">CPU Configuration</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    CPU Cores *
                  </label>
                  <input
                    v-model.number="formData.vcpuCount"
                    type="number"
                    min="1"
                    max="64"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    CPU Model
                  </label>
                  <select
                    v-model="formData.cpuModel"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="host-passthrough">Host Passthrough (Recommended)</option>
                    <option value="host-model">Host Model</option>
                    <option value="qemu64">QEMU64 (Generic)</option>
                    <option value="Skylake-Client">Intel Skylake</option>
                    <option value="Haswell">Intel Haswell</option>
                    <option value="EPYC">AMD EPYC</option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    CPU Topology
                  </label>
                  <select
                    v-model="formData.cpuTopology"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="auto">Auto (Recommended)</option>
                    <option value="manual">Manual Configuration</option>
                  </select>
                </div>
              </div>

              <div v-if="formData.cpuTopology === 'manual'" class="grid grid-cols-3 gap-4 p-4 bg-slate-800/50 rounded-lg">
                <div>
                  <label class="block text-xs font-medium text-slate-300 mb-1">Sockets</label>
                  <input
                    v-model.number="formData.cpuSockets"
                    type="number"
                    min="1"
                    max="4"
                    class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                  />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-300 mb-1">Cores per Socket</label>
                  <input
                    v-model.number="formData.cpuCores"
                    type="number"
                    min="1"
                    max="32"
                    class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                  />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-300 mb-1">Threads per Core</label>
                  <input
                    v-model.number="formData.cpuThreads"
                    type="number"
                    min="1"
                    max="2"
                    class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                  />
                </div>
              </div>

              <!-- Advanced CPU Features -->
              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">CPU Features & Cache</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">
                      CPU Cache Mode
                    </label>
                    <select
                      v-model="formData.cpuCacheMode"
                      class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    >
                      <option value="passthrough">Passthrough (Default)</option>
                      <option value="emulate">Emulate</option>
                      <option value="disable">Disable</option>
                    </select>
                    <p class="text-xs text-slate-400 mt-1">CPU cache configuration for performance tuning</p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-white mb-2">
                      NUMA Configuration
                    </label>
                    <select
                      v-model="formData.numaMode"
                      class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    >
                      <option value="auto">Auto</option>
                      <option value="strict">Strict</option>
                      <option value="preferred">Preferred</option>
                      <option value="interleave">Interleave</option>
                    </select>
                    <p class="text-xs text-slate-400 mt-1">Non-Uniform Memory Access topology</p>
                  </div>
                </div>

                <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.cpuFeatures.pae"
                      type="checkbox"
                      id="cpu-pae"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="cpu-pae" class="text-xs text-white">PAE</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.cpuFeatures.acpi"
                      type="checkbox"
                      id="cpu-acpi"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="cpu-acpi" class="text-xs text-white">ACPI</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.cpuFeatures.apic"
                      type="checkbox"
                      id="cpu-apic"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="cpu-apic" class="text-xs text-white">APIC</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.cpuFeatures.hyperv"
                      type="checkbox"
                      id="cpu-hyperv"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="cpu-hyperv" class="text-xs text-white">Hyper-V</label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 3: Memory Configuration -->
            <div v-if="currentStep === 3" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-green-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">Memory Configuration</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Memory (MB) *
                  </label>
                  <input
                    v-model.number="formData.memoryMB"
                    type="number"
                    min="512"
                    max="131072"
                    step="512"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    required
                  />
                  <p class="text-xs text-slate-400 mt-1">
                    {{ formatBytes(formData.memoryMB * 1024 * 1024) }}
                  </p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Max Memory (MB)
                  </label>
                  <input
                    v-model.number="formData.maxMemoryMB"
                    type="number"
                    :min="formData.memoryMB"
                    max="131072"
                    step="512"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  />
                  <p class="text-xs text-slate-400 mt-1">
                    {{ formatBytes((formData.maxMemoryMB || formData.memoryMB) * 1024 * 1024) }}
                  </p>
                </div>
              </div>

              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">Memory Features</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">
                      Memory Backing
                    </label>
                    <select
                      v-model="formData.memoryBacking"
                      class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    >
                      <option value="standard">Standard Pages</option>
                      <option value="hugepages">Hugepages (2MB)</option>
                      <option value="hugepages-1g">Hugepages (1GB)</option>
                      <option value="file">File-backed</option>
                    </select>
                    <p class="text-xs text-slate-400 mt-1">Memory allocation strategy</p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-white mb-2">
                      Memory Access
                    </label>
                    <select
                      v-model="formData.memoryAccess"
                      class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                    >
                      <option value="private">Private</option>
                      <option value="shared">Shared</option>
                    </select>
                    <p class="text-xs text-slate-400 mt-1">Memory sharing mode</p>
                  </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.memoryBalloon"
                      type="checkbox"
                      id="memory-balloon"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="memory-balloon" class="text-sm text-white">
                      Enable Memory Balloon
                    </label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.memoryLocking"
                      type="checkbox"
                      id="memory-locking"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="memory-locking" class="text-sm text-white">
                      Lock Memory in RAM
                    </label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.memoryHotplug"
                      type="checkbox"
                      id="memory-hotplug"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="memory-hotplug" class="text-sm text-white">
                      Enable Memory Hotplug
                    </label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.memoryDiscard"
                      type="checkbox"
                      id="memory-discard"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="memory-discard" class="text-sm text-white">
                      Enable Memory Discard
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 4: Storage Configuration -->
            <div v-if="currentStep === 4" class="space-y-6">
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-orange-600 rounded-full flex items-center justify-center">
                    <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                    </svg>
                  </div>
                  <h3 class="text-lg font-medium text-white">Storage Configuration</h3>
                </div>
                <FButton
                  type="button"
                  variant="outline"
                  size="sm"
                  @click="addStorageDevice"
                  class="flex items-center gap-2"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
                  </svg>
                  Add Disk
                </FButton>
              </div>
              
              <div class="space-y-4">
                <div
                  v-for="(disk, index) in formData.storageDevices"
                  :key="disk.id"
                  class="p-4 bg-slate-800/50 rounded-lg border border-slate-700"
                >
                  <div class="flex items-center justify-between mb-4">
                    <h4 class="text-sm font-medium text-white">Disk {{ index + 1 }}</h4>
                    <FButton
                      v-if="formData.storageDevices.length > 1"
                      type="button"
                      variant="ghost"
                      size="sm"
                      @click="removeStorageDevice(index)"
                      class="text-red-400 hover:text-red-300"
                    >
                      Remove
                    </FButton>
                  </div>
                  
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Size (GB)</label>
                      <input
                        v-model.number="disk.sizeGB"
                        type="number"
                        min="1"
                        max="2048"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Format</label>
                      <select
                        v-model="disk.format"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="qcow2">QCOW2 (Recommended)</option>
                        <option value="raw">RAW</option>
                        <option value="vmdk">VMDK</option>
                        <option value="vdi">VDI</option>
                        <option value="vhd">VHD</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Bus Type</label>
                      <select
                        v-model="disk.bus"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="virtio">VirtIO (Recommended)</option>
                        <option value="sata">SATA</option>
                        <option value="ide">IDE</option>
                        <option value="scsi">SCSI</option>
                        <option value="nvme">NVMe</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Cache Mode</label>
                      <select
                        v-model="disk.cache"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="none">None</option>
                        <option value="writethrough">Write Through</option>
                        <option value="writeback">Write Back</option>
                        <option value="unsafe">Unsafe (Fastest)</option>
                        <option value="directsync">Direct Sync</option>
                      </select>
                    </div>
                  </div>

                  <!-- Advanced Storage Options -->
                  <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">I/O Mode</label>
                      <select
                        v-model="disk.ioMode"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="threads">Threads</option>
                        <option value="native">Native (Linux AIO)</option>
                        <option value="io_uring">io_uring</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Discard</label>
                      <select
                        v-model="disk.discard"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="ignore">Ignore</option>
                        <option value="unmap">Unmap (TRIM/UNMAP)</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Detect Zeroes</label>
                      <select
                        v-model="disk.detectZeroes"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="off">Off</option>
                        <option value="on">On</option>
                        <option value="unmap">Unmap</option>
                      </select>
                    </div>
                  </div>

                  <!-- QoS Settings -->
                  <div class="mt-4">
                    <h5 class="text-xs font-medium text-slate-300 mb-2">QoS Limits (Optional)</h5>
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">Read IOPS</label>
                        <input
                          v-model.number="disk.qos.readIops"
                          type="number"
                          min="0"
                          placeholder="Unlimited"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">Write IOPS</label>
                        <input
                          v-model.number="disk.qos.writeIops"
                          type="number"
                          min="0"
                          placeholder="Unlimited"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">Read BW (MB/s)</label>
                        <input
                          v-model.number="disk.qos.readBandwidth"
                          type="number"
                          min="0"
                          placeholder="Unlimited"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">Write BW (MB/s)</label>
                        <input
                          v-model.number="disk.qos.writeBandwidth"
                          type="number"
                          min="0"
                          placeholder="Unlimited"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 5: Network Configuration -->
            <div v-if="currentStep === 5" class="space-y-6">
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-cyan-600 rounded-full flex items-center justify-center">
                    <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/>
                    </svg>
                  </div>
                  <h3 class="text-lg font-medium text-white">Network Configuration</h3>
                </div>
                <FButton
                  type="button"
                  variant="outline"
                  size="sm"
                  @click="addNetworkInterface"
                  class="flex items-center gap-2"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
                  </svg>
                  Add Interface
                </FButton>
              </div>
              
              <div class="space-y-4">
                <div
                  v-for="(nic, index) in formData.networkInterfaces"
                  :key="nic.id"
                  class="p-4 bg-slate-800/50 rounded-lg border border-slate-700"
                >
                  <div class="flex items-center justify-between mb-4">
                    <h4 class="text-sm font-medium text-white">Network Interface {{ index + 1 }}</h4>
                    <FButton
                      v-if="formData.networkInterfaces.length > 1"
                      type="button"
                      variant="ghost"
                      size="sm"
                      @click="removeNetworkInterface(index)"
                      class="text-red-400 hover:text-red-300"
                    >
                      Remove
                    </FButton>
                  </div>
                  
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Network</label>
                      <select
                        v-model="nic.network"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="default">Default</option>
                        <option value="bridge">Bridge</option>
                        <option value="nat">NAT</option>
                        <option value="host-only">Host Only</option>
                        <option value="macvtap">MacVTap</option>
                        <option value="sr-iov">SR-IOV</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Model</label>
                      <select
                        v-model="nic.model"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="virtio">VirtIO (Recommended)</option>
                        <option value="e1000">Intel e1000</option>
                        <option value="e1000e">Intel e1000e</option>
                        <option value="vmxnet3">VMware vmxnet3</option>
                        <option value="rtl8139">Realtek RTL8139</option>
                        <option value="ne2k_pci">NE2000 PCI</option>
                        <option value="pcnet">AMD PCnet</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">MAC Address</label>
                      <input
                        v-model="nic.mac"
                        type="text"
                        placeholder="Auto-generate"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400"
                      />
                    </div>
                  </div>

                  <!-- Advanced Network Options -->
                  <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Link State</label>
                      <select
                        v-model="nic.linkState"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      >
                        <option value="up">Up</option>
                        <option value="down">Down</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">Queue Count</label>
                      <input
                        v-model.number="nic.queues"
                        type="number"
                        min="1"
                        max="16"
                        placeholder="Auto"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-slate-300 mb-1">MTU Size</label>
                      <input
                        v-model.number="nic.mtu"
                        type="number"
                        min="68"
                        max="9000"
                        placeholder="1500"
                        class="w-full px-2 py-1 text-sm bg-slate-700 border border-slate-600 rounded text-white"
                      />
                    </div>
                  </div>

                  <!-- VLAN Configuration -->
                  <div v-if="nic.network === 'bridge'" class="mt-4">
                    <h5 class="text-xs font-medium text-slate-300 mb-2">VLAN Configuration</h5>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">VLAN ID</label>
                        <input
                          v-model.number="nic.vlanId"
                          type="number"
                          min="1"
                          max="4094"
                          placeholder="Untagged"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">Native VLAN</label>
                        <select
                          v-model="nic.nativeVlan"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        >
                          <option value="tagged">Tagged</option>
                          <option value="untagged">Untagged</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  <!-- SR-IOV Configuration -->
                  <div v-if="nic.network === 'sr-iov'" class="mt-4">
                    <h5 class="text-xs font-medium text-slate-300 mb-2">SR-IOV Configuration</h5>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">PF Device</label>
                        <input
                          v-model="nic.pfDevice"
                          type="text"
                          placeholder="eth0"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                      <div>
                        <label class="block text-xs text-slate-400 mb-1">VF Number</label>
                        <input
                          v-model.number="nic.vfNumber"
                          type="number"
                          min="0"
                          max="63"
                          placeholder="Auto"
                          class="w-full px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 6: Additional Hardware -->
            <div v-if="currentStep === 6" class="space-y-6">
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center">
                    <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                    </svg>
                  </div>
                  <h3 class="text-lg font-medium text-white">Additional Hardware</h3>
                </div>
                <div class="relative">
                  <FButton
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="showAddHardwareDropdown = !showAddHardwareDropdown"
                    class="flex items-center gap-2"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
                    </svg>
                    Add Hardware
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                  </FButton>
                  
                  <!-- Add Hardware Dropdown -->
                  <div v-if="showAddHardwareDropdown" class="absolute right-0 top-full mt-2 w-56 bg-slate-800 border border-slate-600/50 rounded-lg shadow-lg z-50">
                    <div class="py-2">
                      <button
                        v-for="hardwareType in availableHardwareTypes"
                        :key="hardwareType.id"
                        type="button"
                        @click="addHardwareDevice(hardwareType)"
                        class="w-full px-4 py-2 text-left text-sm text-slate-300 hover:bg-slate-700/50 hover:text-white transition-colors flex items-center gap-3"
                      >
                        <span class="text-lg">{{ hardwareType.icon }}</span>
                        {{ hardwareType.name }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- Existing Hardware Devices -->
              <div class="space-y-4">
                <div
                  v-for="(device, index) in formData.additionalHardware"
                  :key="device.id"
                  class="p-4 bg-slate-800/50 rounded-lg border border-slate-700"
                >
                  <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center gap-3">
                      <span class="text-lg">{{ getHardwareIcon(device.type) }}</span>
                      <h4 class="text-sm font-medium text-white">{{ getHardwareName(device.type) }}</h4>
                    </div>
                    <FButton
                      type="button"
                      variant="ghost"
                      size="sm"
                      @click="removeHardwareDevice(index)"
                      class="text-red-400 hover:text-red-300"
                    >
                      Remove
                    </FButton>
                  </div>
                  
                  <!-- Hardware-specific configuration -->
                  <HardwareConfigPanel
                    v-model="device.config"
                    :device="device"
                  />
                </div>
                
                <!-- Empty state -->
                <div v-if="formData.additionalHardware.length === 0" class="text-center py-8 text-slate-400">
                  <svg class="w-12 h-12 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                  </svg>
                  <p class="text-lg font-medium mb-2">No additional hardware configured</p>
                  <p class="text-sm">Click "Add Hardware" to add devices like USB controllers, audio, graphics cards, etc.</p>
                </div>
              </div>
            </div>

            <!-- Step 7: Security Configuration -->
            <div v-if="currentStep === 7" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-yellow-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">Security Configuration</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Secure Boot
                  </label>
                  <select
                    v-model="formData.secureBoot"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="disabled">Disabled</option>
                    <option value="required">Required</option>
                    <option value="optional">Optional</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Secure Boot helps ensure bootloader integrity</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    TPM Module
                  </label>
                  <select
                    v-model="formData.tpmVersion"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="none">Disabled</option>
                    <option value="1.2">TPM 1.2</option>
                    <option value="2.0">TPM 2.0 (Recommended)</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Virtual Trusted Platform Module for enhanced security</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Memory Encryption
                  </label>
                  <select
                    v-model="formData.memoryEncryption"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="none">Disabled</option>
                    <option value="sev">AMD SEV</option>
                    <option value="sev-es">AMD SEV-ES</option>
                    <option value="sev-snp">AMD SEV-SNP</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Memory encryption (requires AMD EPYC processor)</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Random Number Generator
                  </label>
                  <select
                    v-model="formData.rngDevice"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="none">Disabled</option>
                    <option value="virtio">VirtIO RNG</option>
                    <option value="egd">EGD Socket</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Hardware random number generator for cryptographic operations</p>
                </div>
              </div>

              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">Security Features</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableSMM"
                      type="checkbox"
                      id="enable-smm"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="enable-smm" class="text-sm text-white">
                      Enable SMM (System Management Mode)
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableIOMMU"
                      type="checkbox"
                      id="enable-iommu"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="enable-iommu" class="text-sm text-white">
                      Enable IOMMU Protection
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 8: Advanced Configuration -->
            <div v-if="currentStep === 8" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-purple-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z"/>
                    <path d="M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">Advanced Configuration</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Resource Class
                  </label>
                  <select
                    v-model="formData.resourceClass"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="standard">Standard</option>
                    <option value="compute-optimized">Compute Optimized</option>
                    <option value="memory-optimized">Memory Optimized</option>
                    <option value="storage-optimized">Storage Optimized</option>
                    <option value="gpu-accelerated">GPU Accelerated</option>
                    <option value="custom">Custom</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Predefined resource allocation templates</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    QEMU Guest Agent
                  </label>
                  <select
                    v-model="formData.qemuGuestAgent"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="enabled">Enabled</option>
                    <option value="disabled">Disabled</option>
                    <option value="auto">Auto-detect</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Enable guest-host communication channel</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Placement Policy
                  </label>
                  <select
                    v-model="formData.placementPolicy"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="automatic">Automatic</option>
                    <option value="performance">Performance</option>
                    <option value="balanced">Balanced</option>
                    <option value="power-saving">Power Saving</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">VM placement and scheduling policy</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Watchdog Timer
                  </label>
                  <select
                    v-model="formData.watchdogTimer"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="none">Disabled</option>
                    <option value="i6300esb">i6300ESB</option>
                    <option value="ib700">IB700</option>
                    <option value="diag288">DIAG288 (s390)</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Hardware watchdog for system monitoring</p>
                </div>
              </div>

              <!-- Required Traits Section -->
              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">Required Hardware Traits</h4>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.requiredTraits.avx2"
                      type="checkbox"
                      id="trait-avx2"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="trait-avx2" class="text-xs text-white">AVX2</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.requiredTraits.ssd"
                      type="checkbox"
                      id="trait-ssd"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="trait-ssd" class="text-xs text-white">SSD Storage</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.requiredTraits.sriovNic"
                      type="checkbox"
                      id="trait-sriov"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="trait-sriov" class="text-xs text-white">SR-IOV NIC</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.requiredTraits.gpuAccel"
                      type="checkbox"
                      id="trait-gpu"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="trait-gpu" class="text-xs text-white">GPU Accel</label>
                  </div>
                </div>
              </div>

              <!-- Performance Tuning -->
              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">Performance Tuning</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableNestedVirt"
                      type="checkbox"
                      id="nested-virt"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="nested-virt" class="text-sm text-white">
                      Enable Nested Virtualization
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableKvmClock"
                      type="checkbox"
                      id="kvm-clock"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="kvm-clock" class="text-sm text-white">
                      Enable KVM Paravirtualized Clock
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableVhostNet"
                      type="checkbox"
                      id="vhost-net"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="vhost-net" class="text-sm text-white">
                      Enable vhost-net (Network Performance)
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableMultiqueue"
                      type="checkbox"
                      id="multiqueue"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="multiqueue" class="text-sm text-white">
                      Enable Multiqueue I/O
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 9: Boot & Final Configuration -->
            <div v-if="currentStep === 9" class="space-y-6">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-8 h-8 bg-red-600 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
                  </svg>
                </div>
                <h3 class="text-lg font-medium text-white">Boot & Advanced Options</h3>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Boot Device
                  </label>
                  <select
                    v-model="formData.bootDevice"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="hd">Hard Disk</option>
                    <option value="cdrom">CD-ROM</option>
                    <option value="network">Network (PXE)</option>
                    <option value="floppy">Floppy</option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Firmware Type
                  </label>
                  <select
                    v-model="formData.firmware"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="bios">BIOS</option>
                    <option value="uefi">UEFI</option>
                    <option value="uefi-secure">UEFI with Secure Boot</option>
                  </select>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Boot Order Priority
                  </label>
                  <input
                    v-model="formData.bootOrder"
                    type="text"
                    placeholder="e.g., hd,cdrom,network"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  />
                  <p class="text-xs text-slate-400 mt-1">Comma-separated boot device priority</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Boot Delay (seconds)
                  </label>
                  <input
                    v-model.number="formData.bootDelay"
                    type="number"
                    min="0"
                    max="60"
                    placeholder="0"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  />
                  <p class="text-xs text-slate-400 mt-1">Delay before automatic boot</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    QEMU Guest Agent
                  </label>
                  <select
                    v-model="formData.qemuGuestAgent"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="enabled">Enabled (Recommended)</option>
                    <option value="disabled">Disabled</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Enhanced VM management capabilities</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-white mb-2">
                    Power Management
                  </label>
                  <select
                    v-model="formData.powerManagement"
                    class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                  >
                    <option value="acpi">ACPI (Standard)</option>
                    <option value="guest">Guest-controlled</option>
                    <option value="host">Host-controlled</option>
                  </select>
                  <p class="text-xs text-slate-400 mt-1">Power state management method</p>
                </div>
              </div>

              <div class="space-y-4">
                <h4 class="text-sm font-medium text-slate-300">Advanced Options</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.autoStart"
                      type="checkbox"
                      id="auto-start"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="auto-start" class="text-sm text-white">
                      Start VM automatically
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableSpice"
                      type="checkbox"
                      id="enable-spice"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="enable-spice" class="text-sm text-white">
                      Enable SPICE console
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableVNC"
                      type="checkbox"
                      id="enable-vnc"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="enable-vnc" class="text-sm text-white">
                      Enable VNC console
                    </label>
                  </div>

                  <div class="flex items-center gap-2">
                    <input
                      v-model="formData.enableSerial"
                      type="checkbox"
                      id="enable-serial"
                      class="w-4 h-4 text-primary-600 bg-white/10 border-white/20 rounded focus:ring-primary-400"
                    />
                    <label for="enable-serial" class="text-sm text-white">
                      Enable Serial Console
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Navigation Actions -->
            <div class="flex gap-3 pt-6 border-t border-white/10">
              <FButton
                v-if="currentStep > 1"
                type="button"
                variant="ghost"
                @click="currentStep--"
                :disabled="isLoading"
                class="px-6"
              >
                ← Previous
              </FButton>
              
              <div class="flex-1"></div>
              
              <FButton
                type="button"
                variant="ghost"
                @click="close"
                :disabled="isLoading"
                class="px-6"
              >
                Cancel
              </FButton>
              
              <FButton
                v-if="currentStep < totalSteps"
                type="button"
                variant="primary"
                @click="currentStep++"
                :disabled="!isStepValid(currentStep)"
                class="px-6"
              >
                Next →
              </FButton>
              
              <FButton
                v-else
                type="submit"
                variant="primary"
                :disabled="isLoading || !isFormValid"
                class="px-6"
              >
                <span v-if="!isLoading">Create VM</span>
                <span v-else class="flex items-center gap-2">
                  <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  Creating...
                </span>
              </FButton>
            </div>
          </form>

          <!-- Error Display -->
          <div v-if="error" class="p-3 mx-6 mb-6 bg-red-500/10 border border-red-400/20 rounded-lg">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>
        </div>
      </FCard>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import HardwareConfigPanel from '@/components/vm/HardwareConfigPanel.vue';
import { useVMStore } from '@/stores/vmStore';
import { useHostStore } from '@/stores/hostStore';
import type { VirtualMachine, CreateVMData } from '@/types';

interface Props {
  open: boolean;
  hostId?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:open': [value: boolean];
  vmCreated: [vm: VirtualMachine];
  close: [];
}>();

// Store
const vmStore = useVMStore();
const hostStore = useHostStore();

// Step management
const currentStep = ref(1);
const totalSteps = 9;

const steps = [
  { id: 1, name: 'Basic', icon: '⚙️' },
  { id: 2, name: 'CPU', icon: '🔧' },
  { id: 3, name: 'Memory', icon: '💾' },
  { id: 4, name: 'Storage', icon: '💿' },
  { id: 5, name: 'Network', icon: '🌐' },
  { id: 6, name: 'Hardware', icon: '�' },
  { id: 7, name: 'Security', icon: '🔒' },
  { id: 8, name: 'Advanced', icon: '⚡' },
  { id: 9, name: 'Boot', icon: '🚀' }
];

// Hardware types for add hardware dropdown
const availableHardwareTypes = [
  { id: 'usb-controller', name: 'USB Controller', icon: '🔌' },
  { id: 'sound-card', name: 'Sound Card', icon: '🔊' },
  { id: 'graphics-card', name: 'Graphics Card', icon: '🖥️' },
  { id: 'serial-port', name: 'Serial Port', icon: '📡' },
  { id: 'parallel-port', name: 'Parallel Port', icon: '🖨️' },
  { id: 'tpm', name: 'TPM Module', icon: '🔒' },
  { id: 'watchdog', name: 'Watchdog Timer', icon: '⏰' },
  { id: 'rng', name: 'Random Number Generator', icon: '🎲' },
  { id: 'memory-balloon', name: 'Memory Balloon', icon: '🎈' },
  { id: 'pci-passthrough', name: 'PCI Passthrough', icon: '⚡' },
  { id: 'hostdev', name: 'Host Device', icon: '📱' }
];

// State
const showAddHardwareDropdown = ref(false);

// Generate unique ID for devices
const generateId = () => Math.random().toString(36).substr(2, 9);

// Form state
const formData = reactive({
  name: '',
  description: '',
  hostId: props.hostId || '',
  osType: 'linux',
  architecture: 'x86_64',
  vcpuCount: 2,
  cpuModel: 'host-passthrough',
  cpuTopology: 'auto',
  cpuSockets: 1,
  cpuCores: 2,
  cpuThreads: 1,
  cpuCacheMode: 'passthrough',
  numaMode: 'auto',
  cpuFeatures: {
    pae: true,
    acpi: true,
    apic: true,
    hyperv: false
  },
  memoryMB: 2048,
  maxMemoryMB: null as number | null,
  memoryBalloon: true,
  memoryBacking: 'standard',
  memoryAccess: 'private',
  memoryLocking: false,
  memoryHotplug: false,
  memoryDiscard: false,
  hugepages: false,
  storageDevices: [
    {
      id: generateId(),
      sizeGB: 20,
      format: 'qcow2',
      bus: 'virtio',
      cache: 'none',
      ioMode: 'threads',
      discard: 'ignore',
      detectZeroes: 'off',
      qos: {
        readIops: undefined,
        writeIops: undefined,
        readBandwidth: undefined,
        writeBandwidth: undefined
      }
    }
  ],
  networkInterfaces: [
    {
      id: generateId(),
      network: 'default',
      model: 'virtio',
      mac: '',
      linkState: 'up',
      queues: undefined,
      mtu: undefined,
      vlanId: undefined,
      nativeVlan: 'tagged',
      pfDevice: '',
      vfNumber: undefined
    }
  ],
  additionalHardware: [] as any[],
  bootDevice: 'hd',
  bootOrder: '',
  bootDelay: 0,
  firmware: 'bios',
  powerManagement: 'acpi',
  autoStart: false,
  enableSpice: true,
  enableVNC: false,
  enableSerial: false,
  // Security Configuration
  secureBoot: 'disabled',
  tpmVersion: 'none',
  memoryEncryption: 'none',
  rngDevice: 'none',
  enableSMM: false,
  enableIOMMU: false,
  // Advanced Configuration
  resourceClass: 'standard',
  qemuGuestAgent: 'enabled',
  placementPolicy: 'automatic',
  watchdogTimer: 'none',
  requiredTraits: {
    avx2: false,
    ssd: false,
    sriovNic: false,
    gpuAccel: false
  },
  // Performance Tuning
  enableNestedVirt: false,
  enableKvmClock: true,
  enableVhostNet: true,
  enableMultiqueue: false
});

const isLoading = ref(false);
const error = ref<string | null>(null);

// Computed properties
const availableHosts = computed(() => hostStore.hosts);

const isStepValid = (step: number): boolean => {
  switch (step) {
    case 1:
      return formData.name.trim().length > 0 && formData.hostId.length > 0;
    case 2:
      return formData.vcpuCount > 0;
    case 3:
      return formData.memoryMB >= 512;
    case 4:
      return formData.storageDevices.length > 0 && formData.storageDevices.every(d => d.sizeGB > 0);
    case 5:
      return formData.networkInterfaces.length > 0;
    case 6:
      return true; // Hardware step is optional
    case 7:
      return true; // Security step is optional
    case 8:
      return true; // Advanced step is optional
    case 9:
      return true; // Boot step is optional
    default:
      return true;
  }
};

const isFormValid = computed(() => {
  return isStepValid(1) && isStepValid(2) && isStepValid(3) && isStepValid(4) && isStepValid(5);
});

// Storage device management
const addStorageDevice = () => {
  formData.storageDevices.push({
    id: generateId(),
    sizeGB: 20,
    format: 'qcow2',
    bus: 'virtio',
    cache: 'none',
    ioMode: 'threads',
    discard: 'ignore',
    detectZeroes: 'off',
    qos: {
      readIops: undefined,
      writeIops: undefined,
      readBandwidth: undefined,
      writeBandwidth: undefined
    }
  });
};

const removeStorageDevice = (index: number) => {
  formData.storageDevices.splice(index, 1);
};

// Network interface management
const addNetworkInterface = () => {
  formData.networkInterfaces.push({
    id: generateId(),
    network: 'default',
    model: 'virtio',
    mac: '',
    linkState: 'up',
    queues: undefined,
    mtu: undefined,
    vlanId: undefined,
    nativeVlan: 'tagged',
    pfDevice: '',
    vfNumber: undefined
  });
};

const removeNetworkInterface = (index: number) => {
  formData.networkInterfaces.splice(index, 1);
};

// Hardware device management
const addHardwareDevice = (hardwareType: any) => {
  const device = {
    id: generateId(),
    type: hardwareType.id,
    config: getDefaultHardwareConfig(hardwareType.id)
  };
  
  formData.additionalHardware.push(device);
  showAddHardwareDropdown.value = false;
};

const removeHardwareDevice = (index: number) => {
  formData.additionalHardware.splice(index, 1);
};

const getDefaultHardwareConfig = (type: string) => {
  switch (type) {
    case 'usb-controller':
      return { model: 'usb3', ports: 4 };
    case 'sound-card':
      return { model: 'hda' };
    case 'graphics-card':
      return { model: 'qxl', vram: 64 };
    case 'serial-port':
      return { type: 'pty' };
    case 'parallel-port':
      return { type: 'pty' };
    case 'tpm':
      return { model: 'tpm-tis', version: '2.0' };
    case 'watchdog':
      return { model: 'i6300esb', action: 'reset' };
    case 'rng':
      return { model: 'virtio', backend: '/dev/random' };
    case 'memory-balloon':
      return { model: 'virtio' };
    default:
      return {};
  }
};

const getHardwareName = (type: string): string => {
  const hardware = availableHardwareTypes.find(h => h.id === type);
  return hardware?.name || type;
};

const getHardwareIcon = (type: string): string => {
  const hardware = availableHardwareTypes.find(h => h.id === type);
  return hardware?.icon || '🔧';
};

// Reset form when modal opens/closes
watch(() => props.open, (newValue) => {
  if (newValue) {
    resetForm();
  }
});

const resetForm = (): void => {
  currentStep.value = 1;
  Object.assign(formData, {
    name: '',
    description: '',
    hostId: props.hostId || '',
    osType: 'linux',
    architecture: 'x86_64',
    vcpuCount: 2,
    cpuModel: 'host-passthrough',
    cpuTopology: 'auto',
    cpuSockets: 1,
    cpuCores: 2,
    cpuThreads: 1,
    cpuCacheMode: 'passthrough',
    numaMode: 'auto',
    cpuFeatures: {
      pae: true,
      acpi: true,
      apic: true,
      hyperv: false
    },
    memoryMB: 2048,
    maxMemoryMB: null,
    memoryBalloon: true,
    memoryBacking: 'standard',
    memoryAccess: 'private',
    memoryLocking: false,
    memoryHotplug: false,
    memoryDiscard: false,
    hugepages: false,
    storageDevices: [
      {
        id: generateId(),
        sizeGB: 20,
        format: 'qcow2',
        bus: 'virtio',
        cache: 'none',
        ioMode: 'threads',
        discard: 'ignore',
        detectZeroes: 'off',
        qos: {
          readIops: undefined,
          writeIops: undefined,
          readBandwidth: undefined,
          writeBandwidth: undefined
        }
      }
    ],
    networkInterfaces: [
      {
        id: generateId(),
        network: 'default',
        model: 'virtio',
        mac: '',
        linkState: 'up',
        queues: undefined,
        mtu: undefined,
        vlanId: undefined,
        nativeVlan: 'tagged',
        pfDevice: '',
        vfNumber: undefined
      }
    ],
    additionalHardware: [],
    bootDevice: 'hd',
    bootOrder: '',
    bootDelay: 0,
    firmware: 'bios',
    powerManagement: 'acpi',
    autoStart: false,
    enableSpice: true,
    enableVNC: false,
    enableSerial: false,
    // Security Configuration
    secureBoot: 'disabled',
    tpmVersion: 'none',
    memoryEncryption: 'none',
    rngDevice: 'none',
    enableSMM: false,
    enableIOMMU: false,
    // Advanced Configuration
    resourceClass: 'standard',
    qemuGuestAgent: 'enabled',
    placementPolicy: 'automatic',
    watchdogTimer: 'none',
    requiredTraits: {
      avx2: false,
      ssd: false,
      sriovNic: false,
      gpuAccel: false
    },
    // Performance Tuning
    enableNestedVirt: false,
    enableKvmClock: true,
    enableVhostNet: true,
    enableMultiqueue: false
  });
  error.value = null;
  isLoading.value = false;
  showAddHardwareDropdown.value = false;
};

const close = (): void => {
  if (!isLoading.value) {
    emit('update:open', false);
    emit('close');
  }
};

const handleSubmit = async (): Promise<void> => {
  if (isLoading.value || !isFormValid.value) return;
  
  error.value = null;
  isLoading.value = true;
  
  try {
    // Validate VM name
    if (!formData.name.match(/^[a-zA-Z0-9_-]+$/)) {
      throw new Error('VM name can only contain letters, numbers, underscores, and hyphens');
    }
    
    // Create VM data with all configured options
    const vmData: CreateVMData = {
      name: formData.name.trim(),
      description: formData.description.trim() || 'No description',
      os_type: formData.osType,
      vcpu_count: formData.vcpuCount,
      memory_bytes: formData.memoryMB * 1024 * 1024,
      disk_size_gb: formData.storageDevices[0]?.sizeGB || 20,
      network_interface: formData.networkInterfaces[0]?.network || 'default',
      boot_device: formData.bootDevice,
      cpu_model: formData.cpuModel,
      source: 'managed',
      sync_status: 'SYNCED',
      libvirtState: 'STOPPED',
      hostId: formData.hostId,
      domain_uuid: '',
      title: formData.name.trim(),
      state: 'STOPPED'
    };
    
    // Create the VM using the store
    const newVM = await vmStore.createVM(vmData);
    
    // If auto-start is enabled, start the VM
    if (formData.autoStart) {
      await vmStore.startVM(formData.hostId, newVM.name);
    }
    
    // Emit the vmCreated event with the new VM
    emit('vmCreated', newVM);
    close();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create VM';
  } finally {
    isLoading.value = false;
  }
};

// Utility functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Close on Escape key
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape' && props.open) {
    close();
  }
};

// Add/remove event listener
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    document.addEventListener('keydown', handleKeydown);
  } else {
    document.removeEventListener('keydown', handleKeydown);
  }
});
</script>

<style scoped>
.modal-glow {
  box-shadow: 0 0 50px rgba(59, 130, 246, 0.15);
}

.card-glow {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

@keyframes slide-in-from-top-1 {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-in {
  animation-duration: 200ms;
  animation-fill-mode: both;
}

.slide-in-from-top-1 {
  animation-name: slide-in-from-top-1;
}
</style>