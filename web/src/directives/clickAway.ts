import type { Directive } from 'vue'

interface ClickAwayElement extends HTMLElement {
  clickAwayHandler?: (event: Event) => void
}

export const vClickAway: Directive = {
  mounted(el: ClickAwayElement, binding) {
    el.clickAwayHandler = (event: Event) => {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value(event)
      }
    }
    document.addEventListener('click', el.clickAwayHandler)
  },
  
  beforeUnmount(el: ClickAwayElement) {
    if (el.clickAwayHandler) {
      document.removeEventListener('click', el.clickAwayHandler)
      delete el.clickAwayHandler
    }
  }
}