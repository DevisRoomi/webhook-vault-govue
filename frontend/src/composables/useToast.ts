import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'INFO' | 'SUCCESS' | 'WARN' | 'ERROR'
  message: string
}

const toasts = ref<Toast[]>([])

export function useToast() {
  function show(message: string, type: Toast['type'] = 'INFO', timeout = 4000) {
    const id = Math.random().toString(36).substring(2, 9)
    toasts.value.push({ id, type, message })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, timeout)
  }

  function remove(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return { toasts, show, remove }
}
