<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  method: string
}>()

const color = computed(() => {
  switch (props.method.toUpperCase()) {
    case 'POST': return '#39c5bb'
    case 'PUT': return '#d29922'
    case 'DELETE': return '#f85149'
    case 'GET': return '#58a6ff'
    default: return '#7d8590'
  }
})
</script>

<template>
  <span class="method-tag" :style="{ borderColor: color, color }">
    {{ method.toUpperCase() }}
  </span>
</template>

<style scoped>
.method-tag {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 700;
  padding: 1px 5px;
  background: rgba(0, 0, 0, 0.5);
  border: 1px solid;
}
</style>
