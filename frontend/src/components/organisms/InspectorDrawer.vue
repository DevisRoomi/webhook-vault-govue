<script setup lang="ts">
import { ref, watch } from 'vue'
import StatusBadge from '../atoms/StatusBadge.vue'
import MethodTag from '../atoms/MethodTag.vue'
import PayloadViewer from '../molecules/PayloadViewer.vue'
import DispatchTimeline from '../molecules/DispatchTimeline.vue'
import SkeletonBox from '../atoms/SkeletonBox.vue'
import { useDispatchRetry } from '../../composables/useDispatchRetry'
import { useAuth } from '../../composables/useAuth'
import type { EventDetail } from '../../types'

const props = defineProps<{
  event: EventDetail | null
  isLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open-auth'): void
}>()

const { isAuthenticated } = useAuth()
const { retryDispatch, isRetrying } = useDispatchRetry()

const overrideUrl = ref('')

watch(
  () => props.event,
  (ev) => {
    if (ev) {
      overrideUrl.value = ev.target_url || ''
    }
  },
  { immediate: true }
)

async function triggerRetry() {
  if (!props.event) return
  if (!isAuthenticated.value) {
    emit('open-auth')
    return
  }
  await retryDispatch({
    eventId: props.event.id,
    overrideTargetUrl: overrideUrl.value,
  })
}
</script>

<template>
  <aside class="inspector-drawer">
    <div class="drawer-header">
      <div class="drawer-header-left">
        <span class="drawer-tag">[INSPECTOR PANE]</span>
        <span class="drawer-req" v-if="event">{{ event.request_id }}</span>
      </div>
      <button class="btn-close" @click="emit('close')">[CLOSE x]</button>
    </div>

    <!-- Skeleton loader state -->
    <div v-if="isLoading" class="drawer-body">
      <SkeletonBox height="30px" :count="4" />
      <SkeletonBox height="120px" :count="2" />
    </div>

    <!-- Empty state -->
    <div v-else-if="!event" class="drawer-empty font-editorial">
      <div class="empty-icon">[NO SELECTION]</div>
      <p>Pilih salah satu event webhook dari list untuk menginspeksi payload JSON, header HTTP, dan riwayat delivery attempt.</p>
    </div>

    <!-- Loaded Event Inspection -->
    <div v-else class="drawer-body">
      <!-- Quick Info Bar -->
      <div class="quick-meta-box">
        <div class="meta-row">
          <span class="meta-k">STATUS:</span>
          <StatusBadge :status="event.status" :http-status="event.latest_http_status" />
        </div>
        <div class="meta-row">
          <span class="meta-k">METHOD:</span>
          <MethodTag :method="event.method" />
        </div>
        <div class="meta-row">
          <span class="meta-k">CHANNEL:</span>
          <span class="meta-v">/{{ event.endpoint_slug }} ({{ event.endpoint_name }})</span>
        </div>
        <div class="meta-row">
          <span class="meta-k">INGRESS PATH:</span>
          <span class="meta-v">{{ event.path }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-k">SOURCE IP:</span>
          <span class="meta-v">{{ event.source_ip }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-k">SIZE:</span>
          <span class="meta-v">{{ event.byte_size }} Bytes</span>
        </div>
      </div>

      <!-- Action Box: Retry Dispatch (Key Feature) -->
      <div class="dispatch-action-box">
        <div class="action-box-header">
          <span class="action-title">[HTTP DISPATCH &amp; RETRY CONTROLLER]</span>
        </div>

        <div class="action-box-body">
          <label class="input-label">DESTINATION TARGET URL (EDITABLE FOR RETRY):</label>
          <input
            type="text"
            v-model="overrideUrl"
            class="target-input"
            placeholder="http://127.0.0.1:8086/api/mock/target-success"
          />

          <div class="action-btn-row">
            <button
              class="btn-retry-dispatch"
              :disabled="isRetrying"
              @click="triggerRetry"
            >
              <span v-if="isRetrying">[EXECUTING HTTP RETRY...]</span>
              <span v-else>[⚡ EXECUTE HTTP RETRY DISPATCH]</span>
            </button>
          </div>
          <span class="action-hint font-editorial">
            Mengeksekusi HTTP POST riil ke target di atas dan mencatat status HTTP, body respons, serta latensi.
          </span>
        </div>
      </div>

      <!-- Payload Viewers -->
      <PayloadViewer
        title="INCOMING PAYLOAD (JSON BODY)"
        :raw-text="event.body"
        :is-json="true"
      />

      <PayloadViewer
        title="HTTP HEADERS (INGRESS CAPTURE)"
        :raw-text="event.headers"
        :is-json="true"
      />

      <PayloadViewer
        v-if="event.query_params && event.query_params !== '{}'"
        title="URL QUERY PARAMETERS"
        :raw-text="event.query_params"
        :is-json="true"
      />

      <!-- Dispatch Timeline & Retries -->
      <DispatchTimeline :attempts="event.attempts || []" />
    </div>
  </aside>
</template>

<style scoped>
.inspector-drawer {
  display: flex;
  flex-direction: column;
  width: 480px;
  min-width: 380px;
  background: var(--bg-panel);
  border-left: 1px solid var(--border-wire);
  height: 100%;
  overflow-y: auto;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: var(--bg-panel-sub);
  border-bottom: 1px solid var(--border-wire);
  position: sticky;
  top: 0;
  z-index: 10;
}

.drawer-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.drawer-tag {
  font-size: 10px;
  font-weight: 700;
  color: var(--accent-amber);
}

.drawer-req {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-primary);
}

.btn-close {
  background: transparent;
  border: 1px solid var(--border-wire);
  color: var(--text-muted);
  font-size: 10px;
  padding: 2px 6px;
}
.btn-close:hover {
  color: var(--accent-red);
  border-color: var(--accent-red);
}

.drawer-body {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.drawer-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60%;
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
  gap: 12px;
}

.empty-icon {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-ghost);
  border: 1px dashed var(--border-wire);
  padding: 4px 10px;
}

.quick-meta-box {
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: #0b0c10;
  border: 1px solid var(--border-wire);
  padding: 10px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.meta-k {
  width: 95px;
  color: var(--text-ghost);
  font-weight: 700;
  font-size: 10px;
}

.meta-v {
  color: var(--text-primary);
  word-break: break-all;
}

.dispatch-action-box {
  background: #141720;
  border: 1px solid var(--border-wire-strong);
}

.action-box-header {
  padding: 6px 10px;
  background: #1b202c;
  border-bottom: 1px solid var(--border-wire);
}

.action-title {
  font-size: 10px;
  font-weight: 700;
  color: var(--accent-green);
  letter-spacing: 0.05em;
}

.action-box-body {
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-label {
  font-size: 10px;
  color: var(--text-ghost);
  font-weight: 700;
}

.target-input {
  background: #090a0d;
  border: 1px solid var(--border-wire);
  color: var(--text-primary);
  font-size: 11px;
  padding: 6px 8px;
  outline: none;
}
.target-input:focus {
  border-color: var(--accent-cyan);
}

.action-btn-row {
  display: flex;
  gap: 8px;
}

.btn-retry-dispatch {
  width: 100%;
  background: #18231c;
  border: 1px solid var(--accent-green);
  color: var(--accent-green);
  font-size: 11px;
  font-weight: 700;
  padding: 8px 12px;
}
.btn-retry-dispatch:hover:not(:disabled) {
  background: rgba(63, 185, 80, 0.2);
}

.action-hint {
  font-size: 11px;
  color: var(--text-muted);
}

@media (max-width: 1100px) {
  .inspector-drawer {
    width: 100%;
    min-width: 100%;
    border-left: none;
    border-top: 1px solid var(--border-wire);
    height: 50vh;
  }
}
</style>
