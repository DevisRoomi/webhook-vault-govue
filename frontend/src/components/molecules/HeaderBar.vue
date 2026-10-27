<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuth } from '../../composables/useAuth'
import type { TelemetryStats } from '../../types'

defineProps<{
  stats?: TelemetryStats
}>()

const emit = defineEmits<{
  (e: 'open-auth'): void
  (e: 'open-new-endpoint'): void
  (e: 'simulate', mode: string): void
}>()

const { currentUser, isAuthenticated, logout } = useAuth()
const timeStr = ref('')

function updateTime() {
  const d = new Date()
  timeStr.value = d.toISOString().replace('T', ' ').substring(0, 19) + ' UTC'
}

let timer: any = null
onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <header class="header-wire">
    <div class="header-left">
      <div class="brand-badge">
        <span class="brand-tag">[RAW INGRESS]</span>
        <h1 class="brand-title">WEBHOOK VAULT</h1>
        <span class="version-tag">v1.0.0</span>
      </div>
      <p class="brand-sub font-editorial">
        Platform pencatat, inspeksi payload JSON, dan retry pengiriman webhook HTTP 500 secara real-time.
      </p>
    </div>

    <!-- Live Telemetry Counters -->
    <div class="telemetry-bar" v-if="stats">
      <div class="telemetry-item">
        <span class="label">INGRESS EVENTS</span>
        <span class="val">{{ stats.total_events }}</span>
      </div>
      <div class="telemetry-item">
        <span class="label">DELIVERED (2xx)</span>
        <span class="val text-green">{{ stats.delivered_count }}</span>
      </div>
      <div class="telemetry-item">
        <span class="label">FAILED (5xx/4xx)</span>
        <span class="val text-red">{{ stats.failed_count }}</span>
      </div>
      <div class="telemetry-item">
        <span class="label">CHANNELS</span>
        <span class="val text-cyan">{{ stats.active_endpoints }}</span>
      </div>
    </div>

    <!-- Actions & Auth -->
    <div class="header-right">
      <div class="clock-display">{{ timeStr }}</div>

      <div class="button-group">
        <button
          v-if="isAuthenticated"
          class="btn-raw btn-cyan"
          @click="emit('open-new-endpoint')"
        >
          [+ INGRESS CHANNEL]
        </button>

        <button
          v-if="isAuthenticated"
          class="btn-raw btn-amber"
          @click="emit('simulate', 'SERVER_ERROR_500')"
          title="Fire synthetic webhook that triggers HTTP 500 downstream error"
        >
          [SIMULATE 500 ERR]
        </button>

        <div v-if="isAuthenticated" class="user-block">
          <span class="user-name">@{{ currentUser?.username }}</span>
          <button class="btn-raw-subtle" @click="logout">[LOGOUT]</button>
        </div>
        <button
          v-else
          class="btn-raw btn-primary"
          @click="emit('open-auth')"
        >
          [OPERATOR SIGN-IN]
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header-wire {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 20px;
  background: var(--bg-panel);
  border-bottom: 1px solid var(--border-wire);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.brand-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-tag {
  font-size: 10px;
  font-weight: 700;
  color: var(--accent-amber);
  background: rgba(210, 153, 34, 0.1);
  padding: 2px 6px;
  border: 1px solid var(--accent-amber);
}

.brand-title {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-primary);
}

.version-tag {
  font-size: 10px;
  color: var(--text-ghost);
}

.brand-sub {
  font-size: 12px;
  font-style: italic;
  color: var(--text-muted);
  max-width: 480px;
}

.telemetry-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  border-left: 1px dashed var(--border-wire);
  border-right: 1px dashed var(--border-wire);
  padding: 0 16px;
}

.telemetry-item {
  display: flex;
  flex-direction: column;
}

.telemetry-item .label {
  font-size: 9px;
  font-weight: 700;
  color: var(--text-ghost);
  letter-spacing: 0.06em;
}

.telemetry-item .val {
  font-size: 15px;
  font-weight: 700;
}

.text-green { color: var(--accent-green); }
.text-red { color: var(--accent-red); }
.text-cyan { color: var(--accent-cyan); }

.header-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.clock-display {
  font-size: 11px;
  color: var(--text-muted);
}

.button-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-raw {
  background: var(--bg-panel-sub);
  border: 1px solid var(--border-wire);
  color: var(--text-primary);
  font-size: 11px;
  font-weight: 700;
  padding: 6px 10px;
}

.btn-raw:hover {
  background: #242933;
  border-color: var(--border-wire-strong);
}

.btn-primary {
  border-color: var(--text-primary);
}

.btn-cyan {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}
.btn-cyan:hover {
  background: rgba(57, 197, 187, 0.15);
}

.btn-amber {
  border-color: var(--accent-amber);
  color: var(--accent-amber);
}
.btn-amber:hover {
  background: rgba(210, 153, 34, 0.15);
}

.user-block {
  display: flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-wire);
  padding: 4px 8px;
  background: var(--bg-panel-sub);
}

.user-name {
  font-size: 11px;
  color: var(--accent-cyan);
  font-weight: 700;
}

.btn-raw-subtle {
  background: transparent;
  border: none;
  color: var(--accent-red);
  font-size: 10px;
}
.btn-raw-subtle:hover {
  text-decoration: underline;
}

@media (max-width: 900px) {
  .telemetry-bar {
    display: none;
  }
}
</style>
