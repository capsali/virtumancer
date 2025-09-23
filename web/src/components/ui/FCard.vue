<template>
  <div
    :class="[
      'glass transition-all duration-500 group',
      {
        'hover:glass-medium': interactive,
        'animate-fade-in': animate,
        'animate-scale-in': scaleIn,
        'animate-slide-up': slideUp
      },
      roundedClass,
      paddingClass,
      shadowClass
    ]"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <!-- Floating Orbs -->
    <div v-if="floatingOrbs" class="absolute inset-0 overflow-hidden rounded-inherit">
      <div
        v-for="orb in orbs"
        :key="orb.id"
        :class="[
          'absolute rounded-full blur-xl opacity-20 transition-all duration-1000',
          orb.color,
          orb.animation
        ]"
        :style="{
          width: orb.size,
          height: orb.size,
          left: orb.x,
          top: orb.y
        }"
      ></div>
    </div>

    <!-- Border Glow -->
    <div
      v-if="borderGlow"
      :class="[
        'absolute inset-0 rounded-inherit opacity-0 transition-opacity duration-300',
        'group-hover:opacity-100',
        borderGlowClass
      ]"
    ></div>

    <!-- Content -->
    <div class="relative z-10">
      <slot />
    </div>

    <!-- Bottom Gradient -->
    <div
      v-if="bottomGradient"
      class="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';

interface Props {
  interactive?: boolean;
  animate?: boolean;
  scaleIn?: boolean;
  slideUp?: boolean;
  floatingOrbs?: boolean;
  borderGlow?: boolean;
  bottomGradient?: boolean;
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
  rounded?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl';
  shadow?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | 'glass';
  glowColor?: 'primary' | 'accent' | 'neon-blue' | 'neon-cyan' | 'neon-purple';
}

const props = withDefaults(defineProps<Props>(), {
  interactive: true,
  animate: false,
  scaleIn: false,
  slideUp: false,
  floatingOrbs: false,
  borderGlow: false,
  bottomGradient: false,
  padding: 'md',
  rounded: 'xl',
  shadow: 'glass',
  glowColor: 'primary'
});

const emit = defineEmits<{
  mouseenter: [event: MouseEvent];
  mouseleave: [event: MouseEvent];
}>();

const orbs = ref<Array<{
  id: number;
  x: string;
  y: string;
  size: string;
  color: string;
  animation: string;
}>>([]);

const paddingClass = computed(() => {
  const paddings = {
    none: 'p-0',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
    xl: 'p-12'
  };
  return paddings[props.padding];
});

const roundedClass = computed(() => {
  const rounded = {
    none: 'rounded-none',
    sm: 'rounded-sm',
    md: 'rounded-md',
    lg: 'rounded-lg',
    xl: 'rounded-xl',
    '2xl': 'rounded-2xl',
    '3xl': 'rounded-3xl'
  };
  return rounded[props.rounded];
});

const shadowClass = computed(() => {
  const shadows = {
    none: '',
    sm: 'shadow-sm',
    md: 'shadow-md',
    lg: 'shadow-lg',
    xl: 'shadow-xl',
    glass: 'shadow-glass'
  };
  return shadows[props.shadow];
});

const borderGlowClass = computed(() => {
  const glows = {
    primary: 'border border-primary-400/50 shadow-glow-md',
    accent: 'border border-accent-400/50 shadow-neon-cyan',
    'neon-blue': 'border border-neon-blue/50 shadow-neon-blue',
    'neon-cyan': 'border border-neon-cyan/50 shadow-neon-cyan',
    'neon-purple': 'border border-neon-purple/50 shadow-[0_0_20px_rgba(191,0,255,0.3)]'
  };
  return glows[props.glowColor];
});

const generateOrbs = () => {
  const colors = ['bg-primary-500/20', 'bg-accent-500/20', 'bg-neon-purple/20', 'bg-neon-cyan/20'];
  const animations = ['animate-float-gentle', 'animate-float-medium', 'animate-float-active'];
  
  orbs.value = Array.from({ length: 3 }, (_, i) => ({
    id: i,
    x: `${Math.random() * 80 + 10}%`,
    y: `${Math.random() * 80 + 10}%`,
    size: `${Math.random() * 60 + 40}px`,
    color: colors[Math.floor(Math.random() * colors.length)] as string,
    animation: animations[Math.floor(Math.random() * animations.length)] as string
  }));
};

const handleMouseEnter = (event: MouseEvent) => {
  emit('mouseenter', event);
};

const handleMouseLeave = (event: MouseEvent) => {
  emit('mouseleave', event);
};

onMounted(() => {
  if (props.floatingOrbs) {
    generateOrbs();
  }
});
</script>