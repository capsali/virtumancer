<template>
  <div id="app" class="relative min-h-screen overflow-hidden">
    <!-- Animated Background -->
    <div class="absolute inset-0 bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900">
      <!-- Floating Background Elements -->
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-primary-600/10 rounded-full blur-3xl animate-float-gentle"></div>
      <div class="absolute top-1/3 right-1/4 w-80 h-80 bg-accent-500/10 rounded-full blur-3xl animate-float-medium delay-1000"></div>
      <div class="absolute bottom-1/4 left-1/3 w-72 h-72 bg-neon-purple/10 rounded-full blur-3xl animate-float-active delay-2000"></div>
    </div>

    <!-- Main Content -->
    <div class="relative z-10 min-h-screen flex items-center justify-center p-8">
      <div class="glass rounded-3xl p-12 max-w-4xl w-full text-center animate-fade-in">
        <!-- Header -->
        <div class="mb-12">
          <h1 class="text-7xl font-bold bg-gradient-to-r from-primary-400 via-accent-400 to-primary-400 bg-clip-text text-transparent mb-6 animate-slide-up">
            VirtuMancer
          </h1>
          <p class="text-xl text-slate-300 font-light animate-slide-up delay-200">
            Next-Generation Virtualization Management Platform
          </p>
        </div>

        <!-- Feature Grid -->
        <div class="grid md:grid-cols-3 gap-8 mb-12">
          <div class="glass-subtle rounded-2xl p-6 hover:glass-medium transition-all duration-300 hover:scale-105 animate-slide-up delay-300">
            <div class="w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-neon-blue">
              <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Virtual Machines</h3>
            <p class="text-slate-400 text-sm">Manage and monitor your virtual infrastructure with advanced controls</p>
          </div>

          <div class="glass-subtle rounded-2xl p-6 hover:glass-medium transition-all duration-300 hover:scale-105 animate-slide-up delay-500">
            <div class="w-16 h-16 bg-gradient-to-br from-accent-500 to-accent-600 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-neon-cyan">
              <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Host Management</h3>
            <p class="text-slate-400 text-sm">Centralized control of hypervisor hosts and resource allocation</p>
          </div>

          <div class="glass-subtle rounded-2xl p-6 hover:glass-medium transition-all duration-300 hover:scale-105 animate-slide-up delay-700">
            <div class="w-16 h-16 bg-gradient-to-br from-neon-purple to-neon-pink rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Real-time Monitoring</h3>
            <p class="text-slate-400 text-sm">Live performance metrics and intelligent resource monitoring</p>
          </div>
        </div>

        <!-- CTA Button -->
        <div class="animate-slide-up delay-1000">
          <button class="bg-gradient-to-r from-primary-600 to-accent-600 hover:from-primary-500 hover:to-accent-500 px-8 py-4 rounded-2xl text-white font-semibold shadow-floating hover:shadow-floating-lg transform hover:scale-105 transition-all duration-300">
            Enter Dashboard
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// VirtuMancer - Modern Virtualization Management Platform
</script>
