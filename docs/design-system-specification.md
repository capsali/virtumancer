# VirtuMancer Design System Specification

## Design Philosophy

VirtuMancer embodies the future of virtualization management - clean, powerful, and inspiring. Our design system creates a seamless blend of advanced technology aesthetics with exceptional usability.

### Core Principles
- **Futuristic Elegance**: Clean lines, subtle animations, and modern typography
- **Glassmorphism**: Transparent layers with backdrop blur for depth
- **Neon Accents**: Strategic use of glowing elements for emphasis
- **Responsive Flow**: Smooth transitions and micro-interactions
- **Accessibility First**: WCAG 2.1 AA compliance with beautiful design

## Color Palette

### Light Mode
```css
--primary-50: #eff6ff
--primary-100: #dbeafe
--primary-200: #bfdbfe
--primary-300: #93c5fd
--primary-400: #60a5fa
--primary-500: #3b82f6
--primary-600: #2563eb
--primary-700: #1d4ed8
--primary-800: #1e40af
--primary-900: #1e3a8a

--accent-50: #f0f9ff
--accent-100: #e0f2fe
--accent-200: #bae6fd
--accent-300: #7dd3fc
--accent-400: #38bdf8
--accent-500: #0ea5e9
--accent-600: #0284c7
--accent-700: #0369a1
--accent-800: #075985
--accent-900: #0c4a6e

--neutral-50: #f8fafc
--neutral-100: #f1f5f9
--neutral-200: #e2e8f0
--neutral-300: #cbd5e1
--neutral-400: #94a3b8
--neutral-500: #64748b
--neutral-600: #475569
--neutral-700: #334155
--neutral-800: #1e293b
--neutral-900: #0f172a
```

### Dark Mode
```css
--surface-primary: #0f172a
--surface-secondary: #1e293b
--surface-tertiary: #334155
--glass-primary: rgba(255, 255, 255, 0.05)
--glass-secondary: rgba(255, 255, 255, 0.1)
--glass-tertiary: rgba(255, 255, 255, 0.15)
```

### Neon Colors
```css
--neon-blue: #00d4ff
--neon-purple: #8b5cf6
--neon-green: #00ff88
--neon-pink: #ff0080
--neon-orange: #ff8c00
```

## Typography

### Font Stack
- **Primary**: Inter Variable (headings, UI elements)
- **Monospace**: JetBrains Mono (code, data display)
- **Fallback**: system-ui, -apple-system, sans-serif

### Scale
```css
--text-xs: 0.75rem      /* 12px */
--text-sm: 0.875rem     /* 14px */
--text-base: 1rem       /* 16px */
--text-lg: 1.125rem     /* 18px */
--text-xl: 1.25rem      /* 20px */
--text-2xl: 1.5rem      /* 24px */
--text-3xl: 1.875rem    /* 30px */
--text-4xl: 2.25rem     /* 36px */
--text-5xl: 3rem        /* 48px */
```

## Effects System

### Glassmorphism
```css
.glass-subtle {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.glass-medium {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.glass-strong {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}
```

### Neon Glows
```css
.neon-glow-blue {
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.5);
}

.neon-glow-purple {
  box-shadow: 0 0 20px rgba(139, 92, 246, 0.5);
}

.neon-border-animated {
  position: relative;
  border: 2px solid transparent;
  background: linear-gradient(45deg, var(--neon-blue), var(--neon-purple)) border-box;
  animation: border-flow 3s linear infinite;
}
```

### Floating Effects
```css
.float-gentle {
  animation: float-gentle 6s ease-in-out infinite;
}

.float-medium {
  animation: float-medium 4s ease-in-out infinite;
}

.float-active {
  animation: float-active 2s ease-in-out infinite;
}

@keyframes float-gentle {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
}

@keyframes float-medium {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  25% { transform: translateY(-5px) rotate(1deg); }
  75% { transform: translateY(-15px) rotate(-1deg); }
}
```

## Component Architecture

### Button Variants
- **Primary**: Gradient with neon glow on hover
- **Secondary**: Glass with subtle border
- **Danger**: Red neon with pulsing animation
- **Ghost**: Transparent with neon border on hover

### Card Types
- **Glass Card**: Primary content containers
- **Neon Card**: Critical alerts and status
- **Floating Card**: Interactive elements with hover effects
- **Data Card**: Monospace content with subtle animations

### Navigation Elements
- **Sidebar**: Collapsible glass panel with floating nav items
- **Breadcrumbs**: Neon-accented path indicators
- **Tabs**: Smooth sliding indicator with glass background

## Animation Guidelines

### Timing Functions
- **Ease-Out**: For entrances and reveals
- **Ease-In**: For exits and dismissals  
- **Elastic**: For playful interactions
- **Bounce**: For error states and alerts

### Duration Standards
- **Micro**: 150ms (hover states)
- **Short**: 300ms (transitions)
- **Medium**: 500ms (page changes)
- **Long**: 800ms (complex animations)

### Performance Considerations
- Use `transform` and `opacity` for animations
- Leverage `will-change` property sparingly
- Implement `prefers-reduced-motion` support
- Optimize for 60fps on all interactions

## Responsive Breakpoints

```css
--breakpoint-sm: 640px
--breakpoint-md: 768px
--breakpoint-lg: 1024px
--breakpoint-xl: 1280px
--breakpoint-2xl: 1536px
```

## Accessibility Features

- High contrast ratios (4.5:1 minimum)
- Focus indicators with neon styling
- Keyboard navigation with visual feedback
- Screen reader optimized markup
- Reduced motion preferences respected
- Color-blind friendly palette combinations

This design system will serve as the foundation for creating the most beautiful and functional virtualization management interface ever built.