import type { Directive } from 'vue'

const LAZY_PLACEHOLDER = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"%3E%3C/svg%3E'

const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement
        const src = img.dataset.src
        if (src) {
          img.src = src
          img.removeAttribute('data-src')
          observer.unobserve(img)
        }
      }
    })
  },
  {
    rootMargin: '100px',
    threshold: 0.1
  }
)

export const vLazyload: Directive<HTMLImageElement> = {
  mounted(el, binding) {
    const src = binding.value || el.src
    if (src) {
      el.dataset.src = src
      el.src = LAZY_PLACEHOLDER
      el.loading = 'lazy'
      observer.observe(el)
    }
  },
  updated(el, binding) {
    const src = binding.value || el.src
    if (src && src !== el.dataset.src) {
      el.dataset.src = src
      el.src = LAZY_PLACEHOLDER
      observer.observe(el)
    }
  },
  unmounted(el) {
    observer.unobserve(el)
  }
}
