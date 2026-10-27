<script setup lang="ts">
import { computed } from 'vue'
import StatusBadge from '../atoms/StatusBadge.vue'
import MethodTag from '../atoms/MethodTag.vue'
import type { WebhookEventSummary } from '../../types'

const props = defineProps<{
  event: WebhookEventSummary
  isSelected?: boolean
}>()

const emit = defineEmits<{
  (e: 'select', id: number): void
}>()

const formattedTime = computed(() => {
  const d = new Date(props.event.created_at)
  return d.toTimeString().substring(0, 8)
})

const previewBody = computed(() => {
  try {
    const compact = props.event.body.replace(/\s+/g, ' ').trim()
    return compact.length > 70 ? compact.substring(0, 70) + '...' : compact
  } catch {
    return props.event.body
  }
})
</script>

<template>
  <div
    class="event-row"
    :class="{ selected: isSelected }"
    @click="emit('select', event.id)"
  >
    <div class="col-status">
      <StatusBadge :status="event.status" :http-status="event.latest_http_status" />
    </div>

    <div class="col-method">
      <MethodTag :method="event.method" />
    </div>

    <div class="col-id">
      <span class="req-id">{{ event.request_id }}</span>
      <span class="endpoint-name">/{{ event.endpoint_slug }}</span>
    </div>

    <div class="col-payload font-editorial">
      {{ previewBody }}
    </div>

    <div class="col-meta">
      <span class="byte-badge">{{ event.byte_size }}B</span>
      <span class="latency-badge" v-if="event.latest_latency_ms">
        {{ event.latest_latency_ms }}ms
      </span>
      <span class="time-stamp">{{ formattedTime }}</span>
    </div>

    <div class="col-action">
      <button class="btn-inspect">
        {{ isSelected ? '[SELECTED]' : '[INSPECT]' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.event-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--bg-panel);
  border-bottom: 1px solid var(--border-wire);
  cursor: pointer;
  transition: background 0.1s ease;
}

.event-row:hover {
  background: #171b23;
}

.event-row.selected {
  background: #1d232f;
  border-left: 3px solid var(--accent-amber);
}

.col-status {
  width: 90px;
  flex-shrink: 0;
}

.col-method {
  width: 45px;
  flex-shrink: 0;
}

.col-id {
  display: flex;
  flex-direction: column;
  width: 170px;
  flex-shrink: 0;
}

.req-id {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-primary);
}

.endpoint-name {
  font-size: 10px;
  color: var(--accent-cyan);
}

.col-payload {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.byte-badge {
  font-size: 10px;
  color: var(--text-ghost);
  border: 1px solid #232731;
  padding: 1px 4px;
}

.latency-badge {
  font-size: 10px;
  color: var(--accent-amber);
  border: 1px solid rgba(210, 153, 34, 0.3);
  padding: 1px 4px;
}

.time-stamp {
  font-size: 11px;
  color: var(--text-muted);
}

.col-action {
  flex-shrink: 0;
}

.btn-inspect {
  font-size: 10px;
  font-weight: 700;
  background: var(--bg-panel-sub);
  border: 1px solid var(--border-wire);
  color: var(--text-primary);
  padding: 4px 8px;
}

.btn-inspect:hover {
  border-color: var(--accent-cyan);
}
</style>
