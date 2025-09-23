import { ref, onMounted, onUnmounted } from 'vue';

/**
 * Composable for scroll-triggered animations
 */
export function useScrollAnimation() {
  const isVisible = ref(false);
  const elementRef = ref<HTMLElement | null>(null);

  let observer: IntersectionObserver | null = null;

  onMounted(() => {
    if (elementRef.value) {
      observer = new IntersectionObserver(
        (entries) => {
          const entry = entries[0];
          if (entry) {
            isVisible.value = entry.isIntersecting;
          }
        },
        {
          threshold: 0.1,
          rootMargin: '50px'
        }
      );
      observer.observe(elementRef.value);
    }
  });

  onUnmounted(() => {
    if (observer) {
      observer.disconnect();
    }
  });

  return {
    isVisible,
    elementRef
  };
}

/**
 * Composable for mouse-following animations
 */
export function useMouseFollower() {
  const mouseX = ref(0);
  const mouseY = ref(0);
  const isHovering = ref(false);

  const handleMouseMove = (event: MouseEvent) => {
    mouseX.value = event.clientX;
    mouseY.value = event.clientY;
  };

  const handleMouseEnter = () => {
    isHovering.value = true;
  };

  const handleMouseLeave = () => {
    isHovering.value = false;
  };

  onMounted(() => {
    document.addEventListener('mousemove', handleMouseMove);
  });

  onUnmounted(() => {
    document.removeEventListener('mousemove', handleMouseMove);
  });

  return {
    mouseX,
    mouseY,
    isHovering,
    handleMouseEnter,
    handleMouseLeave
  };
}

/**
 * Composable for particle animations
 */
export function useParticleAnimation() {
  const particles = ref<Array<{
    id: number;
    x: number;
    y: number;
    vx: number;
    vy: number;
    size: number;
    opacity: number;
    color: string;
  }>>([]);

  const generateParticles = (count: number = 50) => {
    const colors = ['#60a5fa', '#22d3ee', '#bf00ff', '#00d9ff'];
    particles.value = Array.from({ length: count }, (_, i) => ({
      id: i,
      x: Math.random() * window.innerWidth,
      y: Math.random() * window.innerHeight,
      vx: (Math.random() - 0.5) * 0.5,
      vy: (Math.random() - 0.5) * 0.5,
      size: Math.random() * 3 + 1,
      opacity: Math.random() * 0.5 + 0.1,
      color: colors[Math.floor(Math.random() * colors.length)] as string
    }));
  };

  const animateParticles = () => {
    particles.value.forEach(particle => {
      particle.x += particle.vx;
      particle.y += particle.vy;

      // Wrap around screen
      if (particle.x < 0) particle.x = window.innerWidth;
      if (particle.x > window.innerWidth) particle.x = 0;
      if (particle.y < 0) particle.y = window.innerHeight;
      if (particle.y > window.innerHeight) particle.y = 0;
    });

    requestAnimationFrame(animateParticles);
  };

  onMounted(() => {
    generateParticles();
    animateParticles();
  });

  return {
    particles
  };
}

/**
 * Composable for typing animation effects
 */
export function useTypingAnimation(text: string, speed: number = 100) {
  const displayText = ref('');
  const isComplete = ref(false);

  const startAnimation = () => {
    let index = 0;
    const interval = setInterval(() => {
      if (index < text.length) {
        displayText.value += text[index];
        index++;
      } else {
        isComplete.value = true;
        clearInterval(interval);
      }
    }, speed);
  };

  return {
    displayText,
    isComplete,
    startAnimation
  };
}