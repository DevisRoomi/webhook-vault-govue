<script setup lang="ts">
import { useToast } from '../../composables/useToast'
import type { Endpoint } from '../../types'

const props = defineProps<{
  endpoints?: Endpoint[]
  selectedId: string
}>()

const emit = defineEmits<{
  (e: 'select', id: string): void
}>()

const toast = useToast()

function copyIngressUrl(slug: string, ev: MouseEvent) {
  ev.stopPropagation()
  const origin = window.location.origin
  const url = `${origin}/api/v1/ingest/${slug}`
  navigator.clipboard.writeText(url)
  toast.show(`INGRESS URL COPIED // ${url}`, 'INFO')
}

function copyCurlSnippet(slug: string, ev: MouseEvent) {
  ev.stopPropagation()
  const origin = window.location.origin
  const snippet = `curl -X POST "${origin}/api/v1/ingest/${slug}" -H "Content-Type: application/json" -d '{"event":"ping","timestamp":"${new Date().toISOString()}"}'`
  navigator.clipboard.writeText(snippet)
  toast.show(`cURL DISPATCH COMMAND COPIED`, 'INFO')
}
</script>

<template>
  <aside class="channel-picker">
    <div class="picker-header">
      <span class="picker-title">[INGRESS CHANNELS]</span>
      <span class="picker-count">{{ endpoints?.length || 0 }} CH</span>
    </div>

    <div class="channel-list">
      <!-- All Channels option -->
      <button
        class="channel-item"
        :class="{ active: selectedId === 'ALL' || selectedId === '' }"
        @click="emit('select', 'ALL')"
      >
        <div class="item-top">
          <span class="channel-slug">[ALL_CHANNELS]</span>
        </div>
        <span class="channel-desc font-editorial">Aggregated live ingress stream from all active gateways.</span>
      </button>

      <!-- Channel entries -->
      <div
        v-for="ep in endpoints"
        :key="ep.id"
        class="channel-item"
        :class="{ active: selectedId === String(ep.id) }"
        @click="emit('select', String(ep.id))"
      >
        <div class="item-top">
          <span class="channel-slug">/{{ ep.slug }}</span>
          <span class="channel-badge" :class="ep.is_active ? 'badge-live' : 'badge-idle'">
            {{ ep.is_active ? '[LIVE]' : '[OFFLINE]' }}
          </span>
        </div>

        <div class="channel-name">{{ ep.name }}</div>
        <div class="channel-desc font-editorial">{{ ep.description }}</div>

        <div class="target-row">
          <span class="target-label">DEST:</span>
          <span class="target-url" :title="ep.target_url">{{ ep.target_url || '[UNFORWARDED / ARCHIVE ONLY]' }}</span>
        </div>

        <div class="stats-counter-row">
          <span>RCV: <b>{{ ep.total_received }}</b></span>
          <span>FWD: <b class="text-green">{{ ep.total_forwarded }}</b></span>
        </div>

        <div class="action-row">
          <button class="btn-micro" @click="copyIngressUrl(ep.slug, $event)">[COPY URL]</button>
          <button class="btn-micro" @click="copyCurlSnippet(ep.slug, $event)">[CURL]</button>
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.channel-picker {
  display: flex;
  flex-direction: column;
  width: 310px;
  min-width: 300px;
  background: var(--bg-panel);
  border-right: 1px solid var(--border-wire);
  height: 100%;
  overflow-y: auto;
}

.picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-panel-sub);
  border-bottom: 1px solid var(--border-wire);
}

.picker-title {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--accent-cyan);
}

.picker-count {
  font-size: 10px;
  color: var(--text-ghost);
}

.channel-list {
  display: flex;
  flex-direction: column;
}

.channel-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 12px 14px;
  background: transparent;
  border: none;
  border-bottom: 1px dashed var(--border-wire);
  text-align: left;
  color: var(--text-primary);
  width: 100%;
}

.channel-item:hover {
  background: #181a20;
}

.channel-item.active {
  background: #1b2029;
  border-left: 3px solid var(--accent-cyan);
}

.item-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.channel-slug {
  font-size: 12px;
  font-weight: 700;
  color: var(--accent-amber);
}

.channel-badge {
  font-size: 9px;
  font-weight: 700;
}
.badge-live { color: var(--accent-green); }
.badge-idle { color: var(--text-ghost); }

.channel-name {
  font-size: 12px;
  font-weight: 600;
}

.channel-desc {
  font-size: 11px;
  color: var(--text-muted);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.target-row {
  display: flex;
  gap: 4px;
  font-size: 10px;
  background: rgba(0, 0, 0, 0.3);
  padding: 2px 6px;
  border: 1px solid #232731;
}

.target-label {
  color: var(--text-ghost);
  font-weight: 700;
}

.target-url {
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stats-counter-row {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-ghost);
  padding: 2px 0;
}
.text-green { color: var(--accent-green); }

.action-row {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}

.btn-micro {
  font-size: 9px;
  font-weight: 700;
  background: var(--bg-panel-sub);
  border: 1px solid var(--border-wire);
  color: var(--text-muted);
  padding: 2px 6px;
}
.btn-micro:hover {
  color: var(--text-primary);
  border-color: var(--accent-cyan);
}
</style>
