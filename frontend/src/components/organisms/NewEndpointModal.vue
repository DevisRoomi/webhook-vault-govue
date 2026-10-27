<script setup lang="ts">
import { ref } from 'vue'
import { useEndpoints } from '../../composables/useEndpoints'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { createEndpoint, isCreating } = useEndpoints()

const form = ref({
  slug: '',
  name: '',
  description: '',
  target_url: '',
  secret_token: '',
})

async function submit() {
  if (!form.value.slug || !form.value.name) return
  try {
    await createEndpoint(form.value)
    emit('close')
  } catch {
    // Handled in composable toast
  }
}

function fillDemoPreset() {
  form.value = {
    slug: 'custom-erp-webhook',
    name: 'Enterprise Inventory Ingress',
    description: 'Captures warehouse sync mutations from ERP SAP connector and forwards to ledger worker.',
    target_url: 'http://127.0.0.1:8086/api/mock/target-success',
    secret_token: 'vault_sec_erp9912',
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="modal-box">
      <div class="modal-header">
        <span class="modal-tag">[PROVISION CHANNEL]</span>
        <button class="btn-close" @click="emit('close')">[x]</button>
      </div>

      <div class="modal-body">
        <div class="preset-row">
          <span class="preset-label font-editorial">Cepat isi contoh:</span>
          <button class="btn-preset" @click="fillDemoPreset">[LOAD PRESET]</button>
        </div>

        <div class="field-block">
          <label class="field-label">CHANNEL SLUG (INGRESS PATH)</label>
          <div class="slug-input-wrap">
            <span class="prefix">/api/v1/ingest/</span>
            <input
              type="text"
              v-model="form.slug"
              placeholder="e.g. stripe-checkout"
              required
            />
          </div>
        </div>

        <div class="field-block">
          <label class="field-label">CHANNEL NAME</label>
          <input
            type="text"
            v-model="form.name"
            class="input-raw"
            placeholder="e.g. Stripe Charge Ledger"
            required
          />
        </div>

        <div class="field-block">
          <label class="field-label">FORWARDING TARGET URL (DESTINATION)</label>
          <input
            type="text"
            v-model="form.target_url"
            class="input-raw"
            placeholder="http://127.0.0.1:8086/api/mock/target-success"
          />
          <span class="field-hint font-editorial">Kosongkan jika hanya ingin mengarsipkan tanpa forward otomatis.</span>
        </div>

        <div class="field-block">
          <label class="field-label">SECRET TOKEN / SIGNING KEY (OPTIONAL)</label>
          <input
            type="text"
            v-model="form.secret_token"
            class="input-raw"
            placeholder="whsec_xxxxxxxx"
          />
        </div>

        <div class="field-block">
          <label class="field-label">DESCRIPTION</label>
          <textarea
            v-model="form.description"
            class="input-raw area-raw"
            rows="2"
            placeholder="Technical details of this webhook pipe..."
          ></textarea>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="emit('close')">[CANCEL]</button>
        <button
          class="btn-submit"
          :disabled="isCreating || !form.slug || !form.name"
          @click="submit"
        >
          {{ isCreating ? '[REGISTERING...]' : '[+ PROVISION INGRESS CHANNEL]' }}
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
  max-width: 520px;
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
  color: var(--accent-cyan);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-ghost);
  font-size: 12px;
}
.btn-close:hover { color: var(--accent-red); }

.modal-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preset-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(0, 0, 0, 0.2);
  padding: 6px 10px;
  border: 1px dashed var(--border-wire);
}

.preset-label {
  font-size: 11px;
  color: var(--text-muted);
}

.btn-preset {
  font-size: 10px;
  background: var(--bg-panel-sub);
  border: 1px solid var(--border-wire);
  color: var(--accent-amber);
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
  letter-spacing: 0.04em;
}

.slug-input-wrap {
  display: flex;
  align-items: center;
  background: var(--bg-canvas);
  border: 1px solid var(--border-wire);
  padding: 0 8px;
}

.slug-input-wrap .prefix {
  font-size: 11px;
  color: var(--text-ghost);
}

.slug-input-wrap input {
  background: transparent;
  border: none;
  color: var(--accent-amber);
  font-size: 12px;
  padding: 7px 4px;
  width: 100%;
  outline: none;
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
  border-color: var(--accent-cyan);
}

.area-raw {
  resize: vertical;
}

.field-hint {
  font-size: 10px;
  color: var(--text-muted);
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
  background: #152220;
  border: 1px solid var(--accent-cyan);
  color: var(--accent-cyan);
  font-size: 11px;
  font-weight: 700;
  padding: 6px 14px;
}
.btn-submit:hover:not(:disabled) {
  background: rgba(57, 197, 187, 0.2);
}
</style>
