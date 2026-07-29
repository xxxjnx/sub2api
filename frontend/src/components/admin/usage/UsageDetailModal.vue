<template>
  <BaseDialog
    :show="show"
    :title="t('usage.detail.title')"
    width="extra-wide"
    @close="emit('update:show', false)"
  >
    <div v-if="loading" class="flex justify-center py-12">
      <svg class="h-7 w-7 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    </div>

    <div v-else-if="loadError" class="py-10 text-center text-sm text-red-500">
      {{ t('usage.detail.loadFailed') }}
    </div>

    <div v-else-if="detail" class="space-y-6">
      <section
        v-if="detail.refunded_at"
        class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm dark:border-emerald-900/60 dark:bg-emerald-950/30"
      >
        <div class="font-semibold text-emerald-800 dark:text-emerald-300">
          {{ t('usage.detail.refunded') }} · ${{ (detail.refund_amount ?? 0).toFixed(6) }}
        </div>
        <div class="mt-1 text-emerald-700 dark:text-emerald-400">
          {{ formatDateTime(detail.refunded_at) }} · {{ detail.refund_reason }}
        </div>
      </section>
      <section>
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('usage.detail.requestContext') }}
        </h4>
        <dl class="grid grid-cols-1 gap-x-6 gap-y-4 rounded-xl border border-gray-200 bg-gray-50 p-4 sm:grid-cols-2 lg:grid-cols-3 dark:border-dark-700 dark:bg-dark-900/50">
          <div v-for="item in contextItems" :key="item.label" class="min-w-0">
            <dt class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ item.label }}</dt>
            <dd
              class="mt-1 break-all text-sm text-gray-900 dark:text-dark-100"
              :class="{ 'font-mono text-xs': item.mono }"
            >
              {{ item.value }}
            </dd>
          </div>
        </dl>
      </section>

      <section>
        <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('usage.detail.requestData') }}
          </h4>
          <span
            v-if="detail.request_data_encoding"
            class="rounded bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300"
          >
            {{ detail.request_data_encoding }}
          </span>
        </div>
        <p class="mb-3 text-xs text-gray-500 dark:text-dark-400">
          {{ t('usage.detail.rawHint') }}
        </p>
        <pre
          v-if="detail.request_data != null"
          data-testid="usage-request-data"
          class="max-h-[52vh] overflow-auto whitespace-pre-wrap break-all rounded-xl border border-gray-200 bg-gray-950 p-4 font-mono text-xs leading-5 text-gray-100 dark:border-dark-700"
        >{{ detail.request_data }}</pre>
        <div
          v-else
          class="rounded-xl border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-dark-400"
        >
          {{ t('usage.detail.notRecorded') }}
        </div>
      </section>
    </div>

    <template #footer>
      <div class="flex w-full flex-col gap-3">
        <div v-if="showRefundForm" class="w-full">
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-dark-200">
            {{ t('usage.detail.refundReason') }}
          </label>
          <textarea
            v-model="refundReason"
            data-testid="refund-reason"
            rows="3"
            maxlength="500"
            class="input w-full resize-y"
            :placeholder="t('usage.detail.refundReasonPlaceholder')"
          />
          <p class="mt-1 text-xs text-amber-600 dark:text-amber-400">
            {{ t('usage.detail.refundWarning', { amount: detail?.actual_cost.toFixed(6) ?? '0' }) }}
          </p>
          <p v-if="refundError" class="mt-1 text-xs text-red-600 dark:text-red-400">
            {{ t('usage.detail.refundFailed') }}
          </p>
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="close">
            {{ t('common.close') }}
          </button>
          <button
            v-if="canRefund && !showRefundForm"
            type="button"
            class="btn btn-danger"
            data-testid="refund-button"
            @click="showRefundForm = true"
          >
            {{ t('usage.detail.refund') }}
          </button>
          <button
            v-if="canRefund && showRefundForm"
            type="button"
            class="btn btn-danger"
            data-testid="confirm-refund-button"
            :disabled="refunding || !refundReason.trim()"
            @click="submitRefund"
          >
            {{ refunding ? t('usage.detail.refunding') : t('usage.detail.confirmRefund') }}
          </button>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminUsageAPI } from '@/api/admin/usage'
import { formatDateTime } from '@/utils/format'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import type { AdminUsageLog } from '@/types'

const props = defineProps<{
  show: boolean
  usageId: number | null
}>()

const emit = defineEmits<{
  (event: 'update:show', value: boolean): void
  (event: 'refunded'): void
}>()

const { t } = useI18n()
const loading = ref(false)
const loadError = ref(false)
const detail = ref<AdminUsageLog | null>(null)
const showRefundForm = ref(false)
const refundReason = ref('')
const refunding = ref(false)
const refundError = ref(false)
let requestSequence = 0
const canRefund = computed(() =>
  Boolean(detail.value && detail.value.actual_cost > 0 && !detail.value.refunded_at),
)

function close() {
  emit('update:show', false)
}

const requestTypeLabel = computed(() => {
  if (!detail.value) return '-'
  const requestType = resolveUsageRequestType(detail.value)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'live') return t('usage.live')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
})

interface ContextItem {
  label: string
  value: string
  mono?: boolean
}

const contextItems = computed(() => {
  const record = detail.value
  if (!record) return []
  const items: ContextItem[] = [
    { label: t('usage.time'), value: formatDateTime(record.created_at) },
    { label: t('usage.detail.requestId'), value: record.request_id || '-', mono: true },
    { label: t('usage.model'), value: record.model || '-' },
    { label: t('usage.upstreamModel'), value: record.upstream_model || '-' },
    { label: t('usage.type'), value: requestTypeLabel.value },
    { label: t('usage.inboundEndpoint'), value: record.inbound_endpoint || '-', mono: true },
    { label: t('usage.upstreamEndpoint'), value: record.upstream_endpoint || '-', mono: true },
    {
      label: t('usage.detail.user'),
      value: record.user?.email ? `${record.user.email} (#${record.user_id})` : `#${record.user_id}`,
    },
    {
      label: t('usage.detail.apiKey'),
      value: record.api_key?.name ? `${record.api_key.name} (#${record.api_key_id})` : `#${record.api_key_id}`,
    },
    {
      label: t('usage.detail.account'),
      value: record.account?.name
        ? `${record.account.name} (#${record.account_id})`
        : (record.account_id != null ? `#${record.account_id}` : '-'),
    },
    {
      label: t('usage.detail.group'),
      value: record.group?.name
        ? `${record.group.name}${record.group_id != null ? ` (#${record.group_id})` : ''}`
        : (record.group_id != null ? `#${record.group_id}` : '-'),
    },
    { label: t('usage.detail.subscription'), value: record.subscription_id != null ? `#${record.subscription_id}` : '-' },
    { label: t('usage.detail.ipAddress'), value: record.ip_address || '-', mono: true },
    { label: t('usage.userAgent'), value: record.user_agent || '-' },
    { label: t('usage.detail.sessionId'), value: record.session_id || '-', mono: true },
    { label: t('usage.detail.contentType'), value: record.request_content_type || '-', mono: true },
  ]
  return items
})

watch(
  () => [props.show, props.usageId] as const,
  ([show, id]) => {
    if (show && id != null) {
      void fetchDetail(id)
      return
    }
    requestSequence += 1
    detail.value = null
    showRefundForm.value = false
    refundReason.value = ''
    refundError.value = false
    loadError.value = false
    loading.value = false
  },
  { immediate: true },
)

async function fetchDetail(id: number) {
  const sequence = ++requestSequence
  loading.value = true
  loadError.value = false
  detail.value = null
  try {
    const result = await adminUsageAPI.getById(id)
    if (sequence === requestSequence) {
      detail.value = result
    }
  } catch (error) {
    if (sequence === requestSequence) {
      console.error('[UsageDetailModal] Failed to load usage detail:', error)
      loadError.value = true
    }
  } finally {
    if (sequence === requestSequence) {
      loading.value = false
    }
  }
}

async function submitRefund() {
  if (!detail.value || !canRefund.value || !refundReason.value.trim()) return
  refunding.value = true
  refundError.value = false
  try {
    detail.value = await adminUsageAPI.refund(detail.value.id, refundReason.value.trim())
    showRefundForm.value = false
    refundReason.value = ''
    emit('refunded')
  } catch (error) {
    console.error('[UsageDetailModal] Failed to refund usage:', error)
    refundError.value = true
  } finally {
    refunding.value = false
  }
}
</script>
