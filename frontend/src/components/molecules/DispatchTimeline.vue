<script setup lang="ts">
import StatusBadge from '../atoms/StatusBadge.vue'
import type { DeliveryAttempt } from '../../types'

defineProps<{
  attempts: DeliveryAttempt[]
}>()

function formatTime(dStr: string) {
  const d = new Date(dStr)
  return d.toTimeString().substring(0, 8)
}
</script>

<template>
  <div class="timeline-wrap">
    <div class="timeline-header">
      <span class="timeline-title">[DISPATCH ATTEMPTS &amp; AUDIT TRAIL]</span>
      <span class="attempt-count">{{ attempts.length }} ATTEMPTS</span>
    </div>

    <div v-if="attempts.length === 0" class="empty-timeline">
      <span class="empty-text font-editorial">No dispatch attempts recorded. Target forwarding not configured or still queued.</span>
    </div>

    <div v-else class="attempt-list">
      <div
        v-for="att in attempts"
        :key="att.id"
        class="attempt-card"
        :class="{ 'card-err': att.http_status >= 400 || att.http_status === 0 }"
      >
        <div class="attempt-top">
          <div class="attempt-badge-row">
            <span class="att-num">#{{ att.attempt_num }}</span>
            <StatusBadge :http-status="att.http_status" />
            <span class="att-latency">{{ att.latency_ms }}ms</span>
          </div>
          <span class="att-time">{{ formatTime(att.dispatched_at) }}</span>
        </div>

        <div class="target-line">
          <span class="tgt-label">TARGET:</span>
          <span class="tgt-val">{{ att.target_url }}</span>
        </div>

        <div v-if="att.error_message" class="att-err-box">
          <span class="err-tag">[DOWNSTREAM_FAILURE]</span>
          <span class="err-msg">{{ att.error_message }}</span>
        </div>

        <div v-if="att.response_body" class="resp-body-box">
          <div class="resp-label">TARGET RESPONSE BODY:</div>
          <pre class="resp-pre"><code>{{ att.response_body }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline-wrap {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  background: var(--bg-panel-sub);
  border: 1px solid var(--border-wire);
}

.timeline-title {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--accent-cyan);
}

.attempt-count {
  font-size: 10px;
  color: var(--text-ghost);
}

.empty-timeline {
  padding: 16px;
  text-align: center;
  border: 1px dashed var(--border-wire);
  color: var(--text-muted);
}

.attempt-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attempt-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px;
  background: #0d0f14;
  border: 1px solid var(--border-wire);
}

.attempt-card.card-err {
  border-left: 3px solid var(--accent-red);
}

.attempt-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.attempt-badge-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.att-num {
  font-size: 11px;
  font-weight: 700;
  color: var(--accent-amber);
}

.att-latency {
  font-size: 10px;
  color: var(--text-ghost);
}

.att-time {
  font-size: 10px;
  color: var(--text-muted);
}

.target-line {
  font-size: 10px;
  display: flex;
  gap: 6px;
  background: rgba(0, 0, 0, 0.3);
  padding: 3px 6px;
  border: 1px solid #1f232b;
}

.tgt-label { color: var(--text-ghost); }
.tgt-val { color: var(--text-muted); word-break: break-all; }

.att-err-box {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  padding: 6px 8px;
}

.err-tag {
  font-size: 9px;
  font-weight: 700;
  color: var(--accent-red);
}

.err-msg {
  font-size: 11px;
  color: #ff7b72;
}

.resp-body-box {
  background: #08090b;
  border: 1px solid #1a1e27;
  padding: 6px 8px;
}

.resp-label {
  font-size: 9px;
  font-weight: 700;
  color: var(--text-ghost);
  margin-bottom: 4px;
}

.resp-pre {
  font-size: 10px;
  color: #c9d1d9;
  line-height: 1.4;
  max-height: 100px;
  overflow-y: auto;
}
</style>
