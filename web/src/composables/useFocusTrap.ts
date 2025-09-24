import { ref, onMounted, onUnmounted, nextTick } from 'vue'

/**
 * Composable for managing focus trap within a modal or dialog
 * Ensures keyboard navigation stays within the modal boundaries
 */
export function useFocusTrap() {
  const trapRef = ref<HTMLElement>()
  const isTrapping = ref(false)
  const previousActiveElement = ref<Element | null>(null)

  // Get all focusable elements within the trap container
  const getFocusableElements = (): HTMLElement[] => {
    if (!trapRef.value) return []
    
    const focusableSelectors = [
      'button:not([disabled]):not([hidden])',
      'input:not([disabled]):not([hidden])',
      'select:not([disabled]):not([hidden])',
      'textarea:not([disabled]):not([hidden])',
      'a[href]:not([disabled]):not([hidden])',
      '[tabindex]:not([tabindex="-1"]):not([disabled]):not([hidden])',
      '[contenteditable="true"]'
    ].join(', ')
    
    const elements = Array.from(trapRef.value.querySelectorAll(focusableSelectors))
      .filter((el) => {
        const element = el as HTMLElement
        const style = window.getComputedStyle(element)
        return element.offsetWidth > 0 && 
               element.offsetHeight > 0 && 
               !element.hidden &&
               style.visibility !== 'hidden' &&
               style.display !== 'none' &&
               !element.hasAttribute('inert')
      }) as HTMLElement[]
    
    // Sort by tab order (tabindex)
    return elements.sort((a, b) => {
      const aIndex = parseInt(a.getAttribute('tabindex') || '0')
      const bIndex = parseInt(b.getAttribute('tabindex') || '0')
      
      if (aIndex === bIndex) return 0
      if (aIndex === 0) return 1  // 0 comes last
      if (bIndex === 0) return -1
      return aIndex - bIndex
    })
  }

  // Handle tab key to cycle through focusable elements
  const handleTabKey = (event: KeyboardEvent, focusableElements: HTMLElement[]) => {
    if (focusableElements.length === 0) return

    const currentIndex = focusableElements.indexOf(document.activeElement as HTMLElement)
    
    if (event.shiftKey) {
      // Shift + Tab (backwards)
      const nextIndex = currentIndex <= 0 ? focusableElements.length - 1 : currentIndex - 1
      const nextElement = focusableElements[nextIndex]
      if (nextElement) {
        event.preventDefault()
        nextElement.focus()
      }
    } else {
      // Tab (forwards)
      const nextIndex = currentIndex >= focusableElements.length - 1 ? 0 : currentIndex + 1
      const nextElement = focusableElements[nextIndex]
      if (nextElement) {
        event.preventDefault()
        nextElement.focus()
      }
    }
  }

  // Main keyboard event handler
  const handleKeydown = (event: KeyboardEvent) => {
    if (!isTrapping.value || !trapRef.value) return

    // Only handle Tab key in focus trap - let other keys bubble up
    if (event.key === 'Tab') {
      console.log('Tab key pressed in focus trap')
      const focusableElements = getFocusableElements()
      console.log('Current focusable elements:', focusableElements.length)
      
      // Only prevent default and handle if we have focusable elements
      if (focusableElements.length > 0) {
        handleTabKey(event, focusableElements)
      }
    }
  }

  // Start focus trapping
  const startTrap = async () => {
    if (isTrapping.value) return

    // Store the previously active element to restore later
    previousActiveElement.value = document.activeElement

    // Wait for next tick to ensure DOM is updated
    await nextTick()

    isTrapping.value = true
    document.addEventListener('keydown', handleKeydown, true) // Use capture phase for better control

    // Focus the first focusable element
    const focusableElements = getFocusableElements()
    console.log('Focus trap started, found elements:', focusableElements.length, focusableElements)
    const firstElement = focusableElements[0]
    if (firstElement) {
      firstElement.focus()
    }
  }

  // Stop focus trapping
  const stopTrap = () => {
    if (!isTrapping.value) return

    console.log('Focus trap stopped')
    isTrapping.value = false
    document.removeEventListener('keydown', handleKeydown, true) // Match the capture phase

    // Restore focus to the previously active element
    if (previousActiveElement.value && 'focus' in previousActiveElement.value) {
      (previousActiveElement.value as HTMLElement).focus()
    }
    previousActiveElement.value = null
  }

  // Cleanup on unmount
  onUnmounted(() => {
    stopTrap()
  })

  return {
    trapRef,
    isTrapping,
    startTrap,
    stopTrap,
    getFocusableElements
  }
}