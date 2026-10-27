<script setup lang="ts">
import { computed } from 'vue'
import type { FilterState } from '../../types'

const props = defineProps<{
  modelValue: FilterState
  isFetching?: boolean
  totalCount?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: FilterState): void
  (e: 'refresh'): void
}>()

function updateSearch(val: string) {
  emit('update:modelValue', { ...props.modelValue, search: val })
}

function updateStatus(val: string) {
  emit('update:modelValue', { ...props.modelValue, status: val })
}

const statusOptions = [
  { label: '[ALL]', value: 'ALL' },
  { label: '[FORWARDED 2xx]', value: 'FORWARDED' },
  { label: '[FAILED 5xx/4xx]', value: 'FAILED' },
  { label: '[RECEIVED]', value: 'RECEIVED' },
]
</script>

<template>
  <div class="filter-toolbar">
    <div class="toolbar-left">
      <div class="search-wrap">
        <span class="search-prefix">QUERY &gt;</span>
        <input
          type="text"
          class="search-input"
          placeholder="Filter request_id, payload keywords, header tokens..."
          :value="modelValue.search"
          @input="updateSearch(($event.target as HTMLInputElement).value)"
        />
        <button
          v-if="modelValue.search"
          class="btn-clear"
          @click="updateSearch('')"
        >
          [CLEAR]
        </button>
      </div>

      <div class="status-tabs">
        <button
          v-for="opt in statusOptions"
          :key="opt.value"
          class="tab-btn"
          :class="{ active: (modelValue.status || 'ALL') === opt.value }"
          @click="updateStatus(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <div class="toolbar-right">
      <span class="count-tag">{{ totalCount || 0 }} EVENTS BUFFERED</span>
      <button class="btn-refresh" :disabled="isFetching" @click="emit('refresh')">
        <span :class="{ 'animate-pulse-subtle': isFetching }">
          {{ isFetching ? '[POLLING...]' : '[SYNC NOW]' }}
        </span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.filter-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 14px;
  background: var(--bg-panel-sub);
  border-bottom: 1px solid var(--border-wire);
}

.toolbar-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.search-wrap {
  display: flex;
  align-items: center;
  background: var(--bg-canvas);
  border: 1px solid var(--border-wire);
  padding: 0 8px;
  flex: 1;
  max-width: 440px;
}

.search-prefix {
  font-size: 11px;
  color: var(--accent-amber);
  margin-right: 6px;
  font-weight: 700;
}

.search-input {
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 12px;
  padding: 6px 0;
  width: 100%;
  outline: none;
}

.btn-clear {
  background: transparent;
  border: none;
  color: var(--text-ghost);
  font-size: 10px;
}
.btn-clear:hover { color: var(--accent-red); }

.status-tabs {
  display: flex;
  border: 1px solid var(--border-wire);
}

.tab-btn {
  background: var(--bg-canvas);
  border: none;
  border-right: 1px solid var(--border-wire);
  color: var(--text-muted);
  font-size: 10px;
  font-weight: 700;
  padding: 5px 8px;
}
.tab-btn:last-child {
  border-right: none;
}
.tab-btn:hover {
  color: var(--text-primary);
  background: #1c202a;
}
.tab-btn.active {
  color: var(--accent-cyan);
  background: #1e2633;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.count-tag {
  font-size: 10px;
  color: var(--text-ghost);
}

.btn-refresh {
  background: var(--bg-canvas);
  border: 1px solid var(--border-wire);
  color: var(--text-primary);
  font-size: 10px;
  font-weight: 700;
  padding: 5px 8px;
}
.btn-refresh:hover {
  border-color: var(--accent-cyan);
}
</style>
