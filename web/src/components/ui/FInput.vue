<template>
  <div class="relative group">
    <!-- Label -->
    <label
      v-if="label"
      :for="inputId"
      :class="[
        'block text-sm font-medium mb-2 transition-colors duration-200',
        labelClass
      ]"
    >
      {{ label }}
      <span v-if="required" class="text-red-400 ml-1">*</span>
    </label>

    <!-- Input Container -->
    <div class="relative">
      <!-- Leading Icon -->
      <div
        v-if="leadingIcon"
        class="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 group-focus-within:text-primary-400 transition-colors duration-200"
      >
        <component :is="leadingIcon" class="w-5 h-5" />
      </div>

      <!-- Input Field -->
      <input
        :id="inputId"
        ref="inputRef"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :required="required"
        :class="[
          'w-full transition-all duration-300 glass border-0 focus:outline-none',
          'placeholder:text-slate-500 text-white',
          inputClasses,
          sizeClasses,
          {
            'pl-10': leadingIcon,
            'pr-10': trailingIcon || type === 'password',
            'ring-2 ring-red-400/50': error,
            'ring-2 ring-primary-400/50': focused && !error,
            'opacity-50 cursor-not-allowed': disabled
          }
        ]"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
        @keydown.enter="handleEnter"
      />

      <!-- Trailing Icon -->
      <div
        v-if="trailingIcon || type === 'password'"
        class="absolute right-3 top-1/2 transform -translate-y-1/2"
      >
        <button
          v-if="type === 'password'"
          type="button"
          class="text-slate-400 hover:text-primary-400 transition-colors duration-200"
          @click="togglePasswordVisibility"
        >
          <component :is="showPassword ? EyeSlashIcon : EyeIcon" class="w-5 h-5" />
        </button>
        <component
          v-else-if="trailingIcon"
          :is="trailingIcon"
          class="w-5 h-5 text-slate-400"
        />
      </div>

      <!-- Floating Particles (for special effects) -->
      <div
        v-if="particles && focused"
        class="absolute inset-0 pointer-events-none overflow-hidden rounded-inherit"
      >
        <div
          v-for="particle in particleArray"
          :key="particle.id"
          :class="[
            'absolute w-1 h-1 bg-primary-400 rounded-full opacity-60',
            'animate-ping'
          ]"
          :style="{
            left: particle.x + '%',
            top: particle.y + '%',
            animationDelay: particle.delay + 'ms'
          }"
        ></div>
      </div>
    </div>

    <!-- Helper Text / Error Message -->
    <div
      v-if="helperText || error"
      :class="[
        'mt-2 text-sm transition-colors duration-200',
        error ? 'text-red-400' : 'text-slate-400'
      ]"
    >
      {{ error || helperText }}
    </div>

    <!-- Glow Effect -->
    <div
      v-if="glow && focused"
      class="absolute inset-0 rounded-xl bg-gradient-to-r from-primary-600/20 to-accent-600/20 blur-lg -z-10 opacity-50"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue';

// Icons (you'd import these from your icon library)
const EyeIcon = 'div'; // Replace with actual icon
const EyeSlashIcon = 'div'; // Replace with actual icon

interface Props {
  modelValue?: string | number;
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search';
  label?: string;
  placeholder?: string;
  helperText?: string;
  error?: string;
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  required?: boolean;
  leadingIcon?: any;
  trailingIcon?: any;
  glow?: boolean;
  particles?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  size: 'md',
  disabled: false,
  required: false,
  glow: false,
  particles: false
});

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
  focus: [event: FocusEvent];
  blur: [event: FocusEvent];
  enter: [event: KeyboardEvent];
}>();

const inputRef = ref<HTMLInputElement>();
const focused = ref(false);
const showPassword = ref(false);
const particleArray = ref<Array<{ id: number; x: number; y: number; delay: number }>>([]);

const inputId = computed(() => `input-${Math.random().toString(36).substr(2, 9)}`);

const sizeClasses = computed(() => {
  const sizes = {
    sm: 'px-3 py-2 text-sm rounded-lg',
    md: 'px-4 py-3 text-base rounded-xl',
    lg: 'px-6 py-4 text-lg rounded-2xl'
  };
  return sizes[props.size];
});

const inputClasses = computed(() => [
  'backdrop-blur-md bg-white/5 border border-white/10',
  'hover:border-white/20 focus:border-primary-400/50',
  'transition-all duration-300'
]);

const labelClass = computed(() => [
  focused.value || props.modelValue ? 'text-primary-400' : 'text-slate-300'
]);

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', props.type === 'number' ? Number(target.value) : target.value);
};

const handleFocus = (event: FocusEvent) => {
  focused.value = true;
  if (props.particles) {
    generateParticles();
  }
  emit('focus', event);
};

const handleBlur = (event: FocusEvent) => {
  focused.value = false;
  particleArray.value = [];
  emit('blur', event);
};

const handleEnter = (event: KeyboardEvent) => {
  emit('enter', event);
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
  nextTick(() => {
    if (inputRef.value) {
      inputRef.value.type = showPassword.value ? 'text' : 'password';
    }
  });
};

const generateParticles = () => {
  particleArray.value = Array.from({ length: 8 }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    y: Math.random() * 100,
    delay: Math.random() * 1000
  }));
};

// Focus method for parent components
const focus = () => {
  inputRef.value?.focus();
};

defineExpose({ focus });
</script>