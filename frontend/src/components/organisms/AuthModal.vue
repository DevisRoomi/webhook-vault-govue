<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '../../composables/useAuth'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { login, register, isAuthenticating } = useAuth()

const mode = ref<'LOGIN' | 'REGISTER'>('LOGIN')
const username = ref('')
const email = ref('')
const password = ref('')

function fillDemo() {
  username.value = 'dispatcher'
  password.value = 'vault2026'
}

async function submit() {
  if (mode.value === 'LOGIN') {
    const ok = await login(username.value, password.value)
    if (ok) emit('close')
  } else {
    const ok = await register(username.value, email.value, password.value)
    if (ok) emit('close')
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="modal-box">
      <div class="modal-header">
        <span class="modal-tag">
          {{ mode === 'LOGIN' ? '[OPERATOR AUTHENTICATION]' : '[NEW OPERATOR REGISTRATION]' }}
        </span>
        <button class="btn-close" @click="emit('close')">[x]</button>
      </div>

      <div class="modal-tabs">
        <button
          class="tab-btn"
          :class="{ active: mode === 'LOGIN' }"
          @click="mode = 'LOGIN'"
        >
          [SIGN-IN]
        </button>
        <button
          class="tab-btn"
          :class="{ active: mode === 'REGISTER' }"
          @click="mode = 'REGISTER'"
        >
          [INITIALIZE ACCOUNT]
        </button>
      </div>

      <div class="modal-body">
        <div v-if="mode === 'LOGIN'" class="demo-box">
          <span class="demo-label font-editorial">Akun operator bawaan demo:</span>
          <button class="btn-demo" @click="fillDemo">[AUTO-FILL: dispatcher]</button>
        </div>

        <div class="field-block">
          <label class="field-label">OPERATOR IDENTIFIER (USERNAME)</label>
          <input
            type="text"
            v-model="username"
            class="input-raw"
            placeholder="dispatcher"
            required
          />
        </div>

        <div v-if="mode === 'REGISTER'" class="field-block">
          <label class="field-label">OFFICIAL EMAIL</label>
          <input
            type="email"
            v-model="email"
            class="input-raw"
            placeholder="operator@vault.wire"
            required
          />
        </div>

        <div class="field-block">
          <label class="field-label">ACCESS SECRET (PASSWORD)</label>
          <input
            type="password"
            v-model="password"
            class="input-raw"
            placeholder="••••••••"
            required
          />
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="emit('close')">[CANCEL]</button>
        <button
          class="btn-submit"
          :disabled="isAuthenticating || !username || !password"
          @click="submit"
        >
          {{ isAuthenticating ? '[VERIFYING...]' : (mode === 'LOGIN' ? '[VERIFY & ENTER]' : '[INITIALIZE OPERATOR]') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 16px;
}

.modal-box {
  width: 100%;
  max-width: 440px;
  background: var(--bg-panel);
  border: 1px solid var(--border-wire-strong);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.9);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: var(--bg-panel-sub);
  border-bottom: 1px solid var(--border-wire);
}

.modal-tag {
  font-size: 11px;
  font-weight: 700;
  color: var(--accent-amber);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-ghost);
  font-size: 12px;
}
.btn-close:hover { color: var(--accent-red); }

.modal-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-wire);
}

.tab-btn {
  flex: 1;
  background: var(--bg-canvas);
  border: none;
  border-right: 1px solid var(--border-wire);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 700;
  padding: 8px 10px;
}
.tab-btn:last-child {
  border-right: none;
}
.tab-btn.active {
  color: var(--text-primary);
  background: var(--bg-panel);
  border-bottom: 2px solid var(--accent-amber);
}

.modal-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.demo-box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  background: rgba(210, 153, 34, 0.08);
  border: 1px dashed var(--accent-amber);
}

.demo-label {
  font-size: 11px;
  color: var(--accent-amber);
}

.btn-demo {
  background: var(--bg-panel);
  border: 1px solid var(--accent-amber);
  color: var(--accent-amber);
  font-size: 10px;
  padding: 2px 6px;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-ghost);
}

.input-raw {
  background: var(--bg-canvas);
  border: 1px solid var(--border-wire);
  color: var(--text-primary);
  font-size: 12px;
  padding: 7px 10px;
  outline: none;
}
.input-raw:focus {
  border-color: var(--accent-amber);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px;
  background: var(--bg-panel-sub);
  border-top: 1px solid var(--border-wire);
}

.btn-cancel {
  background: transparent;
  border: 1px solid var(--border-wire);
  color: var(--text-muted);
  font-size: 11px;
  padding: 6px 12px;
}

.btn-submit {
  background: #231d14;
  border: 1px solid var(--accent-amber);
  color: var(--accent-amber);
  font-size: 11px;
  font-weight: 700;
  padding: 6px 14px;
}
.btn-submit:hover:not(:disabled) {
  background: rgba(210, 153, 34, 0.2);
}
</style>
