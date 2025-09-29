<template>
  <FModal :show="show" @close="handleClose" size="full">
    <FCard class="space-y-0 h-full">
      <!-- Header -->
      <div class="border-b border-slate-700/50 p-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-xl font-semibold text-white flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                </svg>
              </div>
              {{ editMode ? 'Edit Virtual Machine' : 'Create Virtual Machine' }}
            </h3>
            <p class="text-slate-400 mt-1">{{ editMode ? 'Modify VM configuration' : 'Configure all VM settings with full libvirt support' }}</p>
          </div>
          <FButton
            variant="ghost"
            size="sm"
            @click="handleClose"
            class="text-slate-400 hover:text-white"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </FButton>
        </div>
      </div>

      <div class="flex h-full">
        <!-- Sidebar Navigation -->
        <div class="w-64 bg-slate-900/50 border-r border-slate-700/50 p-4">
          <div class="space-y-2">
            <button
              v-for="section in configSections"
              :key="section.id"
              @click="activeSection = section.id"
              :class="[
                'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-all',
                activeSection === section.id
                  ? 'bg-primary-500/20 text-primary-400 ring-1 ring-primary-500/50'
                  : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              ]"
            >
              <span class="text-lg">{{ section.icon }}</span>
              <span class="font-medium">{{ section.label }}</span>
              <div v-if="!validateSection(section.id)" class="ml-auto w-2 h-2 bg-red-500 rounded-full"></div>
            </button>
          </div>
        </div>

        <!-- Main Content -->
        <div class="flex-1 flex flex-col">
          <div class="flex-1 overflow-y-auto p-6">
            
            <!-- General Configuration -->
            <div v-if="activeSection === 'general'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">General Configuration</h4>
                <p class="text-sm text-slate-400">Basic VM identity and overview settings</p>
              </div>
              
              <FCard class="p-6">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Virtual Machine Name *</label>
                    <FInput
                      v-model="config.general.name"
                      placeholder="my-virtual-machine"
                      class="w-full"
                      :class="{ 'border-red-500': !validateField('name') }"
                    />
                    <p class="text-xs text-slate-400 mt-1">Must be unique and follow naming conventions</p>
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Title</label>
                    <FInput
                      v-model="config.general.title"
                      placeholder="Display title"
                      class="w-full"
                    />
                  </div>
                  
                  <div class="lg:col-span-2">
                    <label class="block text-sm font-medium text-white mb-2">Description</label>
                    <textarea
                      v-model="config.general.description"
                      rows="3"
                      placeholder="Describe the purpose and configuration of this VM"
                      class="w-full px-3 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all resize-none"
                    ></textarea>
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Operating System Type</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.general.osType"
                      :options="osTypeOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Machine Type</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.general.machineType"
                      :options="machineTypeOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Architecture</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.general.architecture"
                      :options="architectureOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Firmware</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.general.firmware"
                      :options="firmwareOptions"
                      class="w-full"
                    />
                  </div>
                </div>
                
                <div class="mt-6 space-y-3">
                  <label class="flex items-center space-x-2">
                    <input v-model="config.general.autoStart" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Start VM automatically when host boots</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.general.enableSpice" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable SPICE display</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.general.enableVNC" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable VNC access</span>
                  </label>
                </div>
              </FCard>
            </div>

            <!-- CPU Configuration -->
            <div v-if="activeSection === 'cpu'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">CPU Configuration</h4>
                <p class="text-sm text-slate-400">Virtual CPU settings and performance tuning</p>
              </div>
              
              <FCard class="p-6">
                <h4 class="text-lg font-semibold text-white mb-4">Basic CPU Settings</h4>
                <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">vCPU Count *</label>
                    <FInput
                      v-model.number="config.cpu.vcpuCount"
                      type="number"
                      :min="1"
                      :max="128"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">CPU Sockets</label>
                    <FInput
                      v-model.number="config.cpu.sockets"
                      type="number"
                      :min="1"
                      :max="4"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Cores per Socket</label>
                    <FInput
                      v-model.number="config.cpu.coresPerSocket"
                      type="number"
                      :min="1"
                      :max="32"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Threads per Core</label>
                    <FInput
                      v-model.number="config.cpu.threadsPerCore"
                      type="number"
                      :min="1"
                      :max="2"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">CPU Model</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.cpu.model"
                      :options="cpuModelOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">CPU Mode</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.cpu.mode"
                      :options="cpuModeOptions"
                      class="w-full"
                    />
                  </div>
                </div>
              </FCard>

              <FCard class="p-6">
                <h4 class="text-lg font-semibold text-white mb-4">CPU Features & Performance</h4>
                <div class="space-y-4">
                  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
                    <label class="flex items-center space-x-2">
                      <input v-model="config.cpu.features.hypervisor" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-white">Hypervisor</span>
                    </label>
                    <label class="flex items-center space-x-2">
                      <input v-model="config.cpu.features.kvm" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-white">KVM</span>
                    </label>
                    <label class="flex items-center space-x-2">
                      <input v-model="config.cpu.features.vmx" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-white">VMX (Intel VT)</span>
                    </label>
                    <label class="flex items-center space-x-2">
                      <input v-model="config.cpu.features.svm" type="checkbox" class="form-checkbox">
                      <span class="text-sm text-white">SVM (AMD-V)</span>
                    </label>
                  </div>
                  
                  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div>
                      <label class="block text-sm font-medium text-white mb-2">CPU Pinning</label>
                      <FInput
                        v-model="config.cpu.pinning"
                        placeholder="0-3,6,8-11"
                        class="w-full"
                      />
                      <p class="text-xs text-slate-400 mt-1">Host CPU cores to pin VM CPUs to</p>
                    </div>
                    
                    <div>
                      <label class="block text-sm font-medium text-white mb-2">NUMA Node</label>
                      <FInput
                        v-model.number="config.cpu.numaNode"
                        type="number"
                        :min="0"
                        class="w-full"
                      />
                    </div>
                  </div>
                </div>
              </FCard>
            </div>

            <!-- Memory Configuration -->
            <div v-if="activeSection === 'memory'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Memory Configuration</h4>
                <p class="text-sm text-slate-400">RAM allocation and memory-related settings</p>
              </div>
              
              <FCard class="p-6">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Maximum Memory (MB) *</label>
                    <FInput
                      v-model.number="config.memory.maxMemoryMB"
                      type="number"
                      :min="128"
                      :max="1048576"
                      class="w-full"
                    />
                    <p class="text-xs text-slate-400 mt-1">{{ formatBytes(config.memory.maxMemoryMB * 1024 * 1024) }}</p>
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Current Memory (MB)</label>
                    <FInput
                      v-model.number="config.memory.currentMemoryMB"
                      type="number"
                      :min="128"
                      :max="config.memory.maxMemoryMB"
                      class="w-full"
                    />
                    <p class="text-xs text-slate-400 mt-1">{{ formatBytes(config.memory.currentMemoryMB * 1024 * 1024) }}</p>
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Memory Backing</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.memory.backing"
                      :options="memoryBackingOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Memory Balloon</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.memory.balloon"
                      :options="memoryBalloonOptions"
                      class="w-full"
                    />
                  </div>
                </div>
                
                <div class="mt-6 space-y-3">
                  <label class="flex items-center space-x-2">
                    <input v-model="config.memory.hugepages" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Use hugepages for better performance</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.memory.locked" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Lock memory (prevent swapping)</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.memory.discard" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable memory discard</span>
                  </label>
                </div>
              </FCard>
            </div>

            <!-- Storage Configuration -->
            <div v-if="activeSection === 'storage'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Storage Configuration</h4>
                <p class="text-sm text-slate-400">Virtual disks and storage devices</p>
              </div>
              
              <FCard class="p-6">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="text-lg font-semibold text-white">Virtual Disks</h4>
                  <FButton @click="addDisk" variant="primary" size="sm">
                    <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                    Add Disk
                  </FButton>
                </div>
                
                <div class="space-y-4">
                  <div v-for="(disk, index) in config.storage.disks" :key="index" class="border border-slate-700/50 rounded-lg p-4">
                    <div class="grid grid-cols-1 lg:grid-cols-4 gap-4 mb-4">
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Device</label>
                        <FInput v-model="disk.device" placeholder="vda" class="w-full" />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Bus Type</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="disk.bus"
                          :options="diskBusOptions"
                          class="w-full"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Size (GB)</label>
                        <FInput v-model.number="disk.sizeGB" type="number" :min="1" class="w-full" />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Format</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="disk.format"
                          :options="diskFormatOptions"
                          class="w-full"
                        />
                      </div>
                    </div>
                    
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Source Type</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="disk.sourceType"
                          :options="diskSourceOptions"
                          class="w-full"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Cache Mode</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="disk.cache"
                          :options="diskCacheOptions"
                          class="w-full"
                        />
                      </div>
                    </div>
                    
                    <div class="flex justify-between items-center">
                      <div class="space-x-4">
                        <label class="flex items-center space-x-2">
                          <input v-model="disk.readOnly" type="checkbox" class="form-checkbox">
                          <span class="text-sm text-white">Read Only</span>
                        </label>
                        <label class="flex items-center space-x-2">
                          <input v-model="disk.shareable" type="checkbox" class="form-checkbox">
                          <span class="text-sm text-white">Shareable</span>
                        </label>
                      </div>
                      
                      <FButton @click="removeDisk(index)" variant="danger" size="sm">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                        </svg>
                      </FButton>
                    </div>
                  </div>
                </div>
              </FCard>
            </div>

            <!-- Network Configuration -->
            <div v-if="activeSection === 'network'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Network Configuration</h4>
                <p class="text-sm text-slate-400">Virtual network interfaces and connectivity</p>
              </div>
              
              <FCard class="p-6">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="text-lg font-semibold text-white">Network Interfaces</h4>
                  <FButton @click="addNetworkInterface" variant="primary" size="sm">
                    <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                    Add Interface
                  </FButton>
                </div>
                
                <div class="space-y-4">
                  <div v-for="(nic, index) in config.network.interfaces" :key="index" class="border border-slate-700/50 rounded-lg p-4">
                    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Source Type</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="nic.sourceType"
                          :options="networkSourceOptions"
                          class="w-full"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Source</label>
                        <FInput v-model="nic.source" placeholder="default" class="w-full" />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Model</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="nic.model"
                          :options="networkModelOptions"
                          class="w-full"
                        />
                      </div>
                    </div>
                    
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">MAC Address</label>
                        <FInput
                          v-model="nic.macAddress"
                          placeholder="Auto-generated"
                          class="w-full font-mono"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white mb-2">Link State</label>
                        <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                          v-model="nic.linkState"
                          :options="networkLinkOptions"
                          class="w-full"
                        />
                      </div>
                    </div>
                    
                    <div class="flex justify-end">
                      <FButton @click="removeNetworkInterface(index)" variant="danger" size="sm">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                        </svg>
                      </FButton>
                    </div>
                  </div>
                </div>
              </FCard>
            </div>

            <!-- Graphics & Display -->
            <div v-if="activeSection === 'graphics'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Graphics & Display</h4>
                <p class="text-sm text-slate-400">Video adapters and display configuration</p>
              </div>
              
              <FCard class="p-6">
                <h4 class="text-lg font-semibold text-white mb-4">Display Settings</h4>
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Graphics Type</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.graphics.type"
                      :options="graphicsTypeOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Video Model</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.graphics.videoModel"
                      :options="videoModelOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Video RAM (MB)</label>
                    <FInput
                      v-model.number="config.graphics.videoRamMB"
                      type="number"
                      :min="1"
                      :max="512"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Listen Address</label>
                    <FInput
                      v-model="config.graphics.listenAddress"
                      placeholder="0.0.0.0"
                      class="w-full"
                    />
                  </div>
                </div>
              </FCard>
            </div>

            <!-- Boot Configuration -->
            <div v-if="activeSection === 'boot'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Boot Configuration</h4>
                <p class="text-sm text-slate-400">Boot order and startup options</p>
              </div>
              
              <FCard class="p-6">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Primary Boot Device</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.boot.primaryDevice"
                      :options="bootDeviceOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">Boot Menu</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.boot.bootMenu"
                      :options="bootMenuOptions"
                      class="w-full"
                    />
                  </div>
                </div>
                
                <div class="mt-6">
                  <h5 class="text-sm font-medium text-white mb-3">Boot Order</h5>
                  <div class="space-y-2">
                    <div v-for="(device, index) in config.boot.order" :key="index" class="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
                      <div class="flex items-center space-x-3">
                        <span class="w-6 h-6 bg-blue-500/20 text-blue-400 rounded-full flex items-center justify-center text-xs font-medium">
                          {{ index + 1 }}
                        </span>
                        <span class="text-white">{{ device.label }}</span>
                      </div>
                      <div class="flex items-center space-x-2">
                        <FButton @click="moveBootDevice(index, -1)" :disabled="index === 0" variant="ghost" size="xs">
                          ↑
                        </FButton>
                        <FButton @click="moveBootDevice(index, 1)" :disabled="index === config.boot.order.length - 1" variant="ghost" size="xs">
                          ↓
                        </FButton>
                        <FButton @click="removeBootDevice(index)" variant="danger" size="xs">
                          ×
                        </FButton>
                      </div>
                    </div>
                  </div>
                  
                  <FButton @click="addBootDevice" variant="outline" size="sm" class="mt-3">
                    Add Boot Device
                  </FButton>
                </div>
              </FCard>
            </div>

            <!-- Advanced Configuration -->
            <div v-if="activeSection === 'advanced'" class="space-y-6">
              <div class="mb-6">
                <h4 class="text-lg font-semibold text-white mb-2">Advanced Configuration</h4>
                <p class="text-sm text-slate-400">Security, performance, and special features</p>
              </div>
              
              <FCard class="p-6">
                <h4 class="text-lg font-semibold text-white mb-4">Security Features</h4>
                <div class="space-y-3">
                  <label class="flex items-center space-x-2">
                    <input v-model="config.advanced.enableSecureBoot" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable Secure Boot</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.advanced.enableTPM" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable TPM (Trusted Platform Module)</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.advanced.enableVTd" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable VT-d (Intel IOMMU)</span>
                  </label>
                </div>
              </FCard>
              
              <FCard class="p-6">
                <h4 class="text-lg font-semibold text-white mb-4">Performance Options</h4>
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">CPU Governor</label>
                    <select class="w-full px-3 py-2 bg-slate-800/50 border border-slate-600/50 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-400 focus:border-transparent transition-all"
                      v-model="config.advanced.cpuGovernor"
                      :options="cpuGovernorOptions"
                      class="w-full"
                    />
                  </div>
                  
                  <div>
                    <label class="block text-sm font-medium text-white mb-2">I/O Thread Count</label>
                    <FInput
                      v-model.number="config.advanced.ioThreads"
                      type="number"
                      :min="0"
                      :max="8"
                      class="w-full"
                    />
                  </div>
                </div>
                
                <div class="mt-6 space-y-3">
                  <label class="flex items-center space-x-2">
                    <input v-model="config.advanced.enableNestedVirt" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable nested virtualization</span>
                  </label>
                  <label class="flex items-center space-x-2">
                    <input v-model="config.advanced.enableKSM" type="checkbox" class="form-checkbox">
                    <span class="text-sm text-white">Enable Kernel Samepage Merging (KSM)</span>
                  </label>
                </div>
              </FCard>
            </div>

          </div>

          <!-- Footer -->
          <div class="border-t border-slate-700/50 p-6 bg-slate-900/30">
            <div class="flex justify-between items-center">
              <div class="flex items-center space-x-4">
                <FButton @click="resetToDefaults" variant="ghost" size="sm">
                  Reset to Defaults
                </FButton>
                <FButton @click="loadTemplate" variant="outline" size="sm">
                  Load Template
                </FButton>
                <FButton @click="saveAsTemplate" variant="outline" size="sm">
                  Save as Template
                </FButton>
              </div>
              
              <div class="flex items-center space-x-3">
                <FButton @click="handleClose" variant="ghost">
                  Cancel
                </FButton>
                <FButton
                  @click="handleSave"
                  variant="primary"
                  :disabled="!isConfigValid || loading"
                >
                  <span v-if="loading" class="flex items-center gap-2">
                    <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    {{ editMode ? 'Updating...' : 'Creating...' }}
                  </span>
                  <span v-else>
                    {{ editMode ? 'Update VM' : 'Create VM' }}
                  </span>
                </FButton>
              </div>
            </div>
          </div>
        </div>
      </div>
    </FCard>
  </FModal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import FModal from '@/components/ui/FModal.vue';
import FCard from '@/components/ui/FCard.vue';
import FButton from '@/components/ui/FButton.vue';
import FInput from '@/components/ui/FInput.vue';

interface Props {
  show: boolean;
  hostId?: string;
  vmData?: any; // Existing VM data for edit mode
  editMode?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  editMode: false
});

const emit = defineEmits<{
  close: [];
  'vm-created': [vm: any];
  'vm-updated': [vm: any];
}>();

// State
const activeSection = ref('general');
const loading = ref(false);

// Configuration sections
const configSections = [
  { id: 'general', label: 'General', icon: '⚙️' },
  { id: 'cpu', label: 'CPU', icon: '🖥️' },
  { id: 'memory', label: 'Memory', icon: '💾' },
  { id: 'storage', label: 'Storage', icon: '💿' },
  { id: 'network', label: 'Network', icon: '🌐' },
  { id: 'graphics', label: 'Graphics', icon: '🖼️' },
  { id: 'boot', label: 'Boot', icon: '🚀' },
  { id: 'advanced', label: 'Advanced', icon: '🔧' }
];

// VM Configuration
const config = reactive({
  general: {
    name: '',
    title: '',
    description: '',
    osType: 'linux',
    machineType: 'pc-q35-6.2',
    architecture: 'x86_64',
    firmware: 'bios',
    autoStart: false,
    enableSpice: true,
    enableVNC: false
  },
  cpu: {
    vcpuCount: 2,
    sockets: 1,
    coresPerSocket: 2,
    threadsPerCore: 1,
    model: 'host-passthrough',
    mode: 'host-passthrough',
    features: {
      hypervisor: true,
      kvm: true,
      vmx: false,
      svm: false
    },
    pinning: '',
    numaNode: 0
  },
  memory: {
    maxMemoryMB: 2048,
    currentMemoryMB: 2048,
    backing: 'anonymous',
    balloon: 'virtio',
    hugepages: false,
    locked: false,
    discard: false
  },
  storage: {
    disks: [
      {
        device: 'vda',
        bus: 'virtio',
        sizeGB: 20,
        format: 'qcow2',
        sourceType: 'file',
        cache: 'writeback',
        readOnly: false,
        shareable: false
      }
    ]
  },
  network: {
    interfaces: [
      {
        sourceType: 'network',
        source: 'default',
        model: 'virtio',
        macAddress: '',
        linkState: 'up'
      }
    ]
  },
  graphics: {
    type: 'spice',
    videoModel: 'qxl',
    videoRamMB: 16,
    listenAddress: '127.0.0.1'
  },
  boot: {
    primaryDevice: 'hd',
    bootMenu: 'enabled',
    order: [
      { device: 'hd', label: 'Hard Disk' },
      { device: 'cdrom', label: 'CD-ROM' },
      { device: 'network', label: 'Network (PXE)' }
    ]
  },
  advanced: {
    enableSecureBoot: false,
    enableTPM: false,
    enableVTd: false,
    cpuGovernor: 'performance',
    ioThreads: 1,
    enableNestedVirt: false,
    enableKSM: false
  }
});

// Options for various dropdowns
const osTypeOptions = [
  { value: 'linux', label: 'Linux' },
  { value: 'windows', label: 'Windows' },
  { value: 'freebsd', label: 'FreeBSD' },
  { value: 'openbsd', label: 'OpenBSD' },
  { value: 'netbsd', label: 'NetBSD' },
  { value: 'solaris', label: 'Solaris' },
  { value: 'other', label: 'Other' }
];

const machineTypeOptions = [
  { value: 'pc-q35-6.2', label: 'Q35 (Recommended)' },
  { value: 'pc-i440fx-6.2', label: 'i440FX (Legacy)' },
  { value: 'pc-q35-5.2', label: 'Q35 v5.2' },
  { value: 'pc-i440fx-5.2', label: 'i440FX v5.2' }
];

const architectureOptions = [
  { value: 'x86_64', label: 'x86_64' },
  { value: 'i686', label: 'i686' },
  { value: 'aarch64', label: 'ARM64' },
  { value: 'armv7l', label: 'ARM32' }
];

const firmwareOptions = [
  { value: 'bios', label: 'BIOS (Legacy)' },
  { value: 'uefi', label: 'UEFI' },
  { value: 'uefi-secure', label: 'UEFI with Secure Boot' }
];

const cpuModelOptions = [
  { value: 'host-passthrough', label: 'Host CPU (Passthrough)' },
  { value: 'host-model', label: 'Host CPU (Model)' },
  { value: 'qemu64', label: 'QEMU64 (Generic)' },
  { value: 'Haswell', label: 'Intel Haswell' },
  { value: 'Broadwell', label: 'Intel Broadwell' },
  { value: 'Skylake-Client', label: 'Intel Skylake' },
  { value: 'EPYC', label: 'AMD EPYC' },
  { value: 'EPYC-Rome', label: 'AMD EPYC Rome' }
];

const cpuModeOptions = [
  { value: 'host-passthrough', label: 'Host Passthrough' },
  { value: 'host-model', label: 'Host Model' },
  { value: 'custom', label: 'Custom' }
];

const memoryBackingOptions = [
  { value: 'anonymous', label: 'Anonymous' },
  { value: 'file', label: 'File' },
  { value: 'memfd', label: 'MemFD' }
];

const memoryBalloonOptions = [
  { value: 'virtio', label: 'VirtIO' },
  { value: 'xen', label: 'Xen' },
  { value: 'none', label: 'Disabled' }
];

const diskBusOptions = [
  { value: 'virtio', label: 'VirtIO (Recommended)' },
  { value: 'sata', label: 'SATA' },
  { value: 'ide', label: 'IDE' },
  { value: 'scsi', label: 'SCSI' },
  { value: 'usb', label: 'USB' }
];

const diskFormatOptions = [
  { value: 'qcow2', label: 'QCOW2 (Recommended)' },
  { value: 'raw', label: 'RAW' },
  { value: 'vmdk', label: 'VMDK' },
  { value: 'vdi', label: 'VDI' },
  { value: 'qed', label: 'QED' }
];

const diskSourceOptions = [
  { value: 'file', label: 'File' },
  { value: 'block', label: 'Block Device' },
  { value: 'network', label: 'Network (iSCSI/NFS)' }
];

const diskCacheOptions = [
  { value: 'writeback', label: 'Writeback (Default)' },
  { value: 'writethrough', label: 'Writethrough' },
  { value: 'none', label: 'None' },
  { value: 'unsafe', label: 'Unsafe (Fast)' }
];

const networkSourceOptions = [
  { value: 'network', label: 'Virtual Network' },
  { value: 'bridge', label: 'Bridge' },
  { value: 'hostdev', label: 'Host Device' },
  { value: 'user', label: 'User Mode' }
];

const networkModelOptions = [
  { value: 'virtio', label: 'VirtIO (Recommended)' },
  { value: 'e1000', label: 'Intel E1000' },
  { value: 'e1000e', label: 'Intel E1000e' },
  { value: 'rtl8139', label: 'Realtek RTL8139' },
  { value: 'ne2k_pci', label: 'NE2000 PCI' }
];

const networkLinkOptions = [
  { value: 'up', label: 'Up' },
  { value: 'down', label: 'Down' }
];

const graphicsTypeOptions = [
  { value: 'spice', label: 'SPICE (Recommended)' },
  { value: 'vnc', label: 'VNC' },
  { value: 'none', label: 'None (Headless)' }
];

const videoModelOptions = [
  { value: 'qxl', label: 'QXL (SPICE)' },
  { value: 'virtio', label: 'VirtIO GPU' },
  { value: 'cirrus', label: 'Cirrus' },
  { value: 'vga', label: 'VGA' },
  { value: 'vmvga', label: 'VMware VGA' }
];

const bootDeviceOptions = [
  { value: 'hd', label: 'Hard Disk' },
  { value: 'cdrom', label: 'CD-ROM' },
  { value: 'network', label: 'Network (PXE)' },
  { value: 'fd', label: 'Floppy Disk' }
];

const bootMenuOptions = [
  { value: 'enabled', label: 'Enabled' },
  { value: 'disabled', label: 'Disabled' }
];

const cpuGovernorOptions = [
  { value: 'performance', label: 'Performance' },
  { value: 'powersave', label: 'Power Save' },
  { value: 'ondemand', label: 'On Demand' },
  { value: 'conservative', label: 'Conservative' }
];

// Methods
const addDisk = () => {
  config.storage.disks.push({
    device: `vd${String.fromCharCode(97 + config.storage.disks.length)}`,
    bus: 'virtio',
    sizeGB: 20,
    format: 'qcow2',
    sourceType: 'file',
    cache: 'writeback',
    readOnly: false,
    shareable: false
  });
};

const removeDisk = (index: number) => {
  config.storage.disks.splice(index, 1);
};

const addNetworkInterface = () => {
  config.network.interfaces.push({
    sourceType: 'network',
    source: 'default',
    model: 'virtio',
    macAddress: '',
    linkState: 'up'
  });
};

const removeNetworkInterface = (index: number) => {
  config.network.interfaces.splice(index, 1);
};

const addBootDevice = () => {
  // Implementation for adding boot device
};

const removeBootDevice = (index: number) => {
  config.boot.order.splice(index, 1);
};

const moveBootDevice = (index: number, direction: number) => {
  const newIndex = index + direction;
  if (newIndex >= 0 && newIndex < config.boot.order.length) {
    const item = config.boot.order.splice(index, 1)[0];
    if (item) {
      config.boot.order.splice(newIndex, 0, item);
    }
  }
};

const validateSection = (sectionId: string): boolean => {
  switch (sectionId) {
    case 'general':
      return config.general.name.trim().length > 0;
    case 'cpu':
      return config.cpu.vcpuCount > 0;
    case 'memory':
      return config.memory.maxMemoryMB >= 128;
    case 'storage':
      return config.storage.disks.length > 0;
    case 'network':
      return config.network.interfaces.length > 0;
    default:
      return true;
  }
};

const validateField = (field: string): boolean => {
  switch (field) {
    case 'name':
      return config.general.name.trim().length > 0 && /^[a-zA-Z0-9_-]+$/.test(config.general.name);
    default:
      return true;
  }
};

const isConfigValid = computed(() => {
  return configSections.every(section => validateSection(section.id));
});

const resetToDefaults = () => {
  // Reset configuration to defaults
  Object.assign(config, {
    // ... default configuration
  });
};

const loadTemplate = () => {
  // Implementation for loading template
};

const saveAsTemplate = () => {
  // Implementation for saving as template
};

const handleClose = () => {
  emit('close');
};

const handleSave = async () => {
  if (!isConfigValid.value || loading.value) return;
  
  loading.value = true;
  
  try {
    if (props.editMode) {
      // Update existing VM
      emit('vm-updated', config);
    } else {
      // Create new VM
      emit('vm-created', config);
    }
    
    handleClose();
  } catch (error) {
    console.error('Failed to save VM configuration:', error);
  } finally {
    loading.value = false;
  }
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Initialize with existing VM data if in edit mode
watch(() => props.vmData, (newData) => {
  if (newData && props.editMode) {
    // Populate config with existing VM data
    Object.assign(config, newData);
  }
}, { immediate: true });
</script>

<style scoped>
.modal-glow {
  box-shadow: 0 0 50px rgba(59, 130, 246, 0.3);
}

.glass-medium {
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
</style>