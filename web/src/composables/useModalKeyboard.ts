import { ref } from 'vue'

export interface KeyboardHandlers {
  onEscape?: () => void
  onEnter?: () => void
  onCtrlEnter?: () => void
  onCustomKey?: (event: KeyboardEvent) => void
}

/**
 * Composable for consistent keyboard navigation patterns in modals
 * Handles common shortcuts like Escape, Enter, Ctrl+Enter
 */
export function useModalKeyboard() {
  const keyboardEnabled = ref(true)

  // Create a keyboard event handler with common modal shortcuts
  const createKeyboardHandler = (handlers: KeyboardHandlers = {}) => {
    return (event: KeyboardEvent) => {
      if (!keyboardEnabled.value) return

      // Skip Tab key - let focus trap handle it
      if (event.key === 'Tab') return

      const {
        onEscape,
        onEnter,
        onCtrlEnter,
        onCustomKey
      } = handlers

      // Handle custom keys first
      if (onCustomKey) {
        onCustomKey(event)
        if (event.defaultPrevented) return
      }

      switch (event.key) {
        case 'Escape':
          if (onEscape) {
            event.preventDefault()
            onEscape()
          }
          break

        case 'Enter':
          // Check for Ctrl+Enter or Cmd+Enter first
          if ((event.ctrlKey || event.metaKey) && onCtrlEnter) {
            event.preventDefault()
            onCtrlEnter()
          } else if (onEnter && !event.ctrlKey && !event.metaKey) {
            // Only trigger regular enter if not combined with modifier keys
            const target = event.target as HTMLElement
            // Don't interfere with button clicks or form submissions
            if (target && (
              target.tagName === 'BUTTON' ||
              target.tagName === 'TEXTAREA' ||
              (target as HTMLInputElement).type === 'submit'
            )) {
              return
            }
            event.preventDefault()
            onEnter()
          }
          break
      }
    }
  }

  // Common modal keyboard shortcuts
  const getStandardModalHandlers = (
    onClose: () => void,
    onConfirm?: () => void
  ): KeyboardHandlers => ({
    onEscape: onClose,
    onCtrlEnter: onConfirm || onClose,
    onEnter: onConfirm
  })

  // Disable/enable keyboard handling (useful for nested modals)
  const disableKeyboard = () => {
    keyboardEnabled.value = false
  }

  const enableKeyboard = () => {
    keyboardEnabled.value = true
  }

  return {
    keyboardEnabled,
    createKeyboardHandler,
    getStandardModalHandlers,
    disableKeyboard,
    enableKeyboard
  }
}