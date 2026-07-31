<template>
  <div class="docs-code-wrap">
    <div class="docs-code-head">
      <span>{{ label }}</span>
      <button
        type="button"
        :aria-label="`复制${label}`"
        @click="emit('copy', code)"
      >
        <Icon name="copy" size="sm" />
        复制
      </button>
    </div>
    <pre class="docs-code"><code>{{ code }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  code: string
  label: string
}>()

const emit = defineEmits<{
  copy: [code: string]
}>()
</script>

<style scoped>
.docs-code-wrap {
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  border: 1px solid #2b333d;
  border-radius: 0.8125rem;
  background: #181d23;
}

.docs-code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px solid #303842;
  padding: 0.5625rem 0.8125rem;
  color: #9ba4ad;
  font-size: 0.6875rem;
}

.docs-code-head button {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border: 0;
  background: transparent;
  color: #5eead4;
  font-size: 0.6875rem;
  font-weight: 700;
  transition: color 150ms ease;
}

.docs-code-head button:hover {
  color: #99f6e4;
}

.docs-code-head button:focus-visible {
  border-radius: 0.375rem;
  outline: 2px solid rgba(45, 212, 191, 0.55);
  outline-offset: 2px;
}

.docs-code {
  max-width: 100%;
  margin: 0;
  padding: 1rem;
  overflow: auto;
  color: #f1f5f9;
  white-space: pre;
}

.docs-code code {
  display: block;
  width: max-content;
  min-width: 100%;
  font: 0.75rem/1.8 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (max-width: 640px) {
  .docs-code code {
    font-size: 0.6875rem;
  }
}
</style>
