<script setup lang="ts">
import { computed } from 'vue'
import { useToast } from '../../composables/useToast'

const props = defineProps<{
  title: string
  rawText: string
  isJson?: boolean
}>()

const toast = useToast()

const formattedContent = computed(() => {
  if (!props.rawText) return '// Empty body'
  if (props.isJson) {
    try {
      const parsed = JSON.parse(props.rawText)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return props.rawText
    }
  }
  return props.rawText
})

function copyContent() {
  navigator.clipboard.writeText(formattedContent.value)
  toast.show(`PAYLOAD COPIED TO CLIPBOARD`, 'INFO')
}
</script>

<template>
  <div class="payload-box">
    <div class="box-bar">
      <span class="box-title">{{ title }}</span>
      <button class="btn-copy" @click="copyContent">[COPY RAW]</button>
    </div>
    <pre class="box-pre"><code>{{ formattedContent }}</code></pre>
  </div>
</template>

<style scoped>
.payload-box {
  background: #090a0d;
  border: 1px solid var(--border-wire);
  margin-bottom: 12px;
}

.box-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  background: var(--bg-panel-sub);
  border-bottom: 1px solid var(--border-wire);
}

.box-title {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.btn-copy {
  background: transparent;
  border: none;
  font-size: 10px;
  color: var(--accent-cyan);
}
.btn-copy:hover {
  text-decoration: underline;
}

.box-pre {
  padding: 10px;
  font-size: 11px;
  color: #c9d1d9;
  line-height: 1.45;
  overflow-x: auto;
  max-height: 240px;
}
</style>
