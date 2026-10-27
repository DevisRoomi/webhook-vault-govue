<script setup lang="ts">
import { ref, onMounted } from 'vue'
import HeaderBar from './components/molecules/HeaderBar.vue'
import ChannelPicker from './components/molecules/ChannelPicker.vue'
import FilterToolbar from './components/molecules/FilterToolbar.vue'
import EventRow from './components/molecules/EventRow.vue'
import InspectorDrawer from './components/organisms/InspectorDrawer.vue'
import NewEndpointModal from './components/organisms/NewEndpointModal.vue'
import AuthModal from './components/organisms/AuthModal.vue'
import SkeletonBox from './components/atoms/SkeletonBox.vue'

import { useAuth } from './composables/useAuth'
import { useEndpoints } from './composables/useEndpoints'
import { useWebhooks } from './composables/useWebhooks'
import { useDispatchRetry } from './composables/useDispatchRetry'
import { useToast } from './composables/useToast'
import type { FilterState } from './types'

const filter = ref<FilterState>({
  endpointId: 'ALL',
  status: 'ALL',
  search: '',
})

const { checkSession, isAuthenticated } = useAuth()
const { endpoints, refetch: refetchEndpoints } = useEndpoints()
const {
  events,
  isLoading: isEventsLoading,
  isFetching,
  isError: isEventsError,
  error: eventsError,
  refetchEvents,
  selectedEventId,
  selectedEvent,
  isLoadingDetail,
  selectEvent,
  stats,
} = useWebhooks(filter)

const { simulateTest, isSimulating } = useDispatchRetry()
const { toasts, remove: removeToast } = useToast()

const showAuthModal = ref(false)
const showNewEndpointModal = ref(false)

onMounted(async () => {
  await checkSession()
  // Select first event if present
  if (events.value && events.value.length > 0 && !selectedEventId.value) {
    selectEvent(events.value[0].id)
  }
})

function handleChannelSelect(channelId: string) {
  filter.value.endpointId = channelId
}

async function handleSimulate(mode: string) {
  if (!isAuthenticated.value) {
    showAuthModal.value = true
    return
  }
  const eventName = mode === 'SERVER_ERROR_500' ? 'invoice.payment_deadlock' : 'order.fulfilled'
  await simulateTest({ event_name: eventName, simulate: mode })
}
</script>

<template>
  <div class="app-layout">
    <!-- Top System Telemetry Bar -->
    <HeaderBar
      :stats="stats"
      @open-auth="showAuthModal = true"
      @open-new-endpoint="showNewEndpointModal = true"
      @simulate="handleSimulate"
    />

    <!-- Main Asymmetric Workspace Grid -->
    <main class="workspace-grid">
      <!-- Left Column: Channel / Ingress Picker -->
      <ChannelPicker
        :endpoints="endpoints"
        :selected-id="filter.endpointId"
        @select="handleChannelSelect"
      />

      <!-- Center Column: Ingress Stream Ledger -->
      <section class="stream-panel">
        <FilterToolbar
          v-model="filter"
          :is-fetching="isFetching || isSimulating"
          :total-count="events?.length || 0"
          @refresh="() => { refetchEvents(); refetchEndpoints(); }"
        />

        <div class="stream-container">
          <!-- Backend Connection / Query Error State -->
          <div v-if="isEventsError" class="state-banner banner-err">
            <span class="err-title">[BACKEND CONNECTION FAILURE]</span>
            <span class="err-text font-editorial">
              Gagal menghubungi server Go API di port 8086. Pastikan `vault-server.exe` sedang berjalan.
              ({{ (eventsError as any)?.message }})
            </span>
            <button class="btn-retry" @click="() => refetchEvents()">[RETRY CONNECTION]</button>
          </div>

          <!-- Pure Skeleton Loading State -->
          <div v-else-if="isEventsLoading" class="stream-loading">
            <div class="loading-label">[ACQUIRING INGRESS STREAM TELEMETRY...]</div>
            <SkeletonBox height="42px" :count="6" />
          </div>

          <!-- Communicative Empty State -->
          <div v-else-if="!events || events.length === 0" class="stream-empty font-editorial">
            <div class="empty-bracket">[NO INCOMING WEBHOOKS IN BUFFER]</div>
            <p>
              Belum ada permintaan HTTP yang masuk pada saluran ini. Kirim permintaan webhook menggunakan cURL
              atau tekan tombol simulasi untuk memicu pengujian.
            </p>
            <div class="empty-code">
              <code>curl -X POST "http://webhook-vault-govue.test/api/v1/ingest/stripe-checkout" -H "Content-Type: application/json" -d '{"event":"test"}'</code>
            </div>
            <button class="btn-fire-test" @click="handleSimulate('SUCCESS_200')">
              [DISPATCH SYNTHETIC TEST EVENT]
            </button>
          </div>

          <!-- Real Ingress Events Ledger -->
          <div v-else class="event-stream-list">
            <EventRow
              v-for="ev in events"
              :key="ev.id"
              :event="ev"
              :is-selected="selectedEventId === ev.id"
              @select="selectEvent(ev.id)"
            />
          </div>
        </div>
      </section>

      <!-- Right Column: Deep Inspector Drawer -->
      <InspectorDrawer
        :event="selectedEvent || null"
        :is-loading="isLoadingDetail"
        @close="selectEvent(null)"
        @open-auth="showAuthModal = true"
      />
    </main>

    <!-- Modals -->
    <AuthModal
      v-if="showAuthModal"
      @close="showAuthModal = false"
    />

    <NewEndpointModal
      v-if="showNewEndpointModal"
      @close="showNewEndpointModal = false"
    />

    <!-- Toast Stack -->
    <div class="toast-stack">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="toast-item"
        :class="`toast-${t.type.toLowerCase()}`"
        @click="removeToast(t.id)"
      >
        <span class="toast-type">[{{ t.type }}]</span>
        <span class="toast-msg">{{ t.message }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--bg-canvas);
}

.workspace-grid {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

.stream-panel {
  display: flex;
  flex-direction: column;
  flex: 1;
  background: var(--bg-canvas);
  overflow: hidden;
}

.stream-container {
  flex: 1;
  overflow-y: auto;
}

.state-banner {
  margin: 16px;
  padding: 14px;
  border: 1px solid;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.banner-err {
  background: rgba(248, 81, 73, 0.1);
  border-color: var(--accent-red);
}
.err-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--accent-red);
}
.err-text {
  font-size: 12px;
  color: #ff7b72;
}
.btn-retry {
  align-self: flex-start;
  background: transparent;
  border: 1px solid var(--accent-red);
  color: var(--accent-red);
  font-size: 10px;
  padding: 4px 8px;
}

.stream-loading {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.loading-label {
  font-size: 11px;
  color: var(--accent-amber);
  font-weight: 700;
}

.stream-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
  color: var(--text-muted);
  max-width: 580px;
  margin: 40px auto;
  border: 1px dashed var(--border-wire);
  background: var(--bg-panel);
  gap: 14px;
}
.empty-bracket {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 700;
  color: var(--accent-amber);
}
.empty-code {
  font-family: var(--font-mono);
  font-size: 11px;
  background: #090a0d;
  padding: 8px 12px;
  border: 1px solid var(--border-wire);
  color: #c9d1d9;
  word-break: break-all;
}
.btn-fire-test {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  background: var(--bg-panel-sub);
  border: 1px solid var(--accent-cyan);
  color: var(--accent-cyan);
  padding: 8px 14px;
}
.btn-fire-test:hover {
  background: rgba(57, 197, 187, 0.15);
}

.event-stream-list {
  display: flex;
  flex-direction: column;
}

/* Toast Notification Stack */
.toast-stack {
  position: fixed;
  bottom: 16px;
  right: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  z-index: 1000;
}

.toast-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #11141c;
  border: 1px solid var(--border-wire-strong);
  font-size: 11px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.8);
  cursor: pointer;
}

.toast-item.toast-success { border-left: 3px solid var(--accent-green); }
.toast-item.toast-error { border-left: 3px solid var(--accent-red); }
.toast-item.toast-warn { border-left: 3px solid var(--accent-amber); }
.toast-item.toast-info { border-left: 3px solid var(--accent-cyan); }

.toast-type {
  font-weight: 700;
  font-size: 10px;
}
.toast-success .toast-type { color: var(--accent-green); }
.toast-error .toast-type { color: var(--accent-red); }
.toast-warn .toast-type { color: var(--accent-amber); }
.toast-info .toast-type { color: var(--accent-cyan); }

.toast-msg {
  color: var(--text-primary);
}

@media (max-width: 1024px) {
  .workspace-grid {
    flex-direction: column;
  }
}
</style>
