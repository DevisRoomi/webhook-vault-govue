<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status?: string
  httpStatus?: number
}>()

const badgeConfig = computed(() => {
  if (props.httpStatus) {
    if (props.httpStatus >= 200 && props.httpStatus < 300) {
      return { text: `[${props.httpStatus} OK]`, color: '#3fb950', border: '#238636' }
    }
    if (props.httpStatus >= 500) {
      return { text: `[${props.httpStatus} ERR]`, color: '#f85149', border: '#da3633' }
    }
    if (props.httpStatus >= 400) {
      return { text: `[${props.httpStatus} WARN]`, color: '#d29922', border: '#9e6a03' }
    }
    return { text: `[${props.httpStatus}]`, color: '#7d8590', border: '#30363d' }
  }

  switch (props.status) {
    case 'FORWARDED':
      return { text: '[FORWARDED]', color: '#3fb950', border: '#238636' }
    case 'FAILED':
      return { text: '[FAILED]', color: '#f85149', border: '#da3633' }
    case 'RETRIED':
      return { text: '[RETRIED]', color: '#58a6ff', border: '#1f6feb' }
    case 'RECEIVED':
    default:
      return { text: '[RECEIVED]', color: '#d29922', border: '#9e6a03' }
  }
})
</script>

<template>
  <span
    class="status-badge"
    :style="{ color: badgeConfig.color, borderColor: badgeConfig.border }"
  >
    {{ badgeConfig.text }}
  </span>
</template>

<style scoped>
.status-badge {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  padding: 1px 6px;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid;
  white-space: nowrap;
}
</style>
