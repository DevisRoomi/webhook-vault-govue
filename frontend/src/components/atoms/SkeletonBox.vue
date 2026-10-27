<script setup lang="ts">
defineProps<{
  height?: string
  width?: string
  count?: number
}>()
</script>

<template>
  <div class="skeleton-stack">
    <div
      v-for="i in (count || 1)"
      :key="i"
      class="skeleton-item animate-pulse-subtle"
      :style="{ height: height || '18px', width: width || '100%' }"
    />
  </div>
</template>

<style scoped>
.skeleton-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.skeleton-item {
  background: #1c202a;
  border: 1px dashed #2d3342;
}
</style>
