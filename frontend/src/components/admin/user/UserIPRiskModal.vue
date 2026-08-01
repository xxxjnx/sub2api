<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.ipRisk.title')"
    width="extra-wide"
    @close="emit('close')"
  >
    <div class="space-y-5">
      <div
        class="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200"
      >
        <Icon name="exclamationTriangle" size="md" class="mt-0.5 shrink-0" />
        <p>{{ t('admin.users.ipRisk.description') }}</p>
      </div>

      <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
        <div class="flex flex-col gap-3 md:flex-row md:items-end">
          <label class="block min-w-0 flex-1">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.users.ipRisk.manualIp') }}
            </span>
            <input
              v-model.trim="manualIP"
              type="text"
              class="input font-mono"
              :placeholder="t('admin.users.ipRisk.ipPlaceholder')"
              data-test="registration-ip-input"
              @keyup.enter="beginBlock(manualIP)"
            />
          </label>
          <button
            type="button"
            class="btn btn-secondary md:shrink-0"
            :disabled="!manualIP"
            data-test="prepare-registration-ip-block"
            @click="beginBlock(manualIP)"
          >
            <Icon name="ban" size="sm" class="mr-2" />
            {{ t('admin.users.ipRisk.prepareBlock') }}
          </button>
        </div>

        <div
          v-if="pendingBlockIP"
          class="mt-4 rounded-lg border border-red-200 bg-white p-4 dark:border-red-900/50 dark:bg-dark-900"
        >
          <p class="text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.users.ipRisk.blockingIp', { ip: pendingBlockIP }) }}
          </p>
          <textarea
            v-model="blockReason"
            rows="2"
            maxlength="500"
            class="input mt-3 resize-none"
            :placeholder="t('admin.users.ipRisk.reasonPlaceholder')"
            data-test="registration-ip-block-reason"
          ></textarea>
          <div class="mt-3 flex justify-end gap-2">
            <button type="button" class="btn btn-secondary" @click="cancelBlock">
              {{ t('common.cancel') }}
            </button>
            <button
              type="button"
              class="btn btn-danger"
              :disabled="submittingIP === pendingBlockIP"
              data-test="confirm-registration-ip-block"
              @click="confirmBlock"
            >
              {{ submittingIP === pendingBlockIP ? t('common.processing') : t('admin.users.ipRisk.confirmBlock') }}
            </button>
          </div>
        </div>
      </div>

      <DataTable
        :columns="columns"
        :data="risks"
        :loading="loading"
        row-key="ip_address"
      >
        <template #cell-ip_address="{ value }">
          <code class="whitespace-nowrap rounded bg-gray-100 px-2 py-1 text-xs text-gray-800 dark:bg-dark-700 dark:text-gray-200">
            {{ value }}
          </code>
        </template>

        <template #cell-users="{ row }">
          <div class="min-w-56 space-y-1.5 py-1">
            <div
              v-for="user in row.users"
              :key="user.id"
              class="flex items-center justify-between gap-3 text-xs"
            >
              <span class="min-w-0 truncate text-gray-700 dark:text-gray-300" :title="user.email">
                {{ user.email }}
              </span>
              <span
                :class="[
                  'badge shrink-0 text-[10px]',
                  user.status === 'active' ? 'badge-success' : 'badge-gray'
                ]"
              >
                {{ user.status === 'active' ? t('common.active') : t('admin.users.disabled') }}
              </span>
            </div>
            <span v-if="row.users.length === 0" class="text-xs text-gray-400">-</span>
          </div>
        </template>

        <template #cell-user_count="{ value }">
          <span class="badge badge-warning">{{ value }}</span>
        </template>

        <template #cell-first_registered_at="{ value }">
          <span class="whitespace-nowrap text-xs text-gray-500 dark:text-dark-400">
            {{ value ? formatDateTime(value) : '-' }}
          </span>
        </template>

        <template #cell-last_registered_at="{ value }">
          <span class="whitespace-nowrap text-xs text-gray-500 dark:text-dark-400">
            {{ value ? formatDateTime(value) : '-' }}
          </span>
        </template>

        <template #cell-blocked="{ row }">
          <div class="space-y-1">
            <span :class="['badge', row.blocked ? 'badge-danger' : 'badge-success']">
              {{ row.blocked ? t('admin.users.ipRisk.blocked') : t('admin.users.ipRisk.unblocked') }}
            </span>
            <p
              v-if="row.blocked && row.block_reason"
              class="max-w-48 truncate text-xs text-gray-500 dark:text-dark-400"
              :title="row.block_reason"
            >
              {{ row.block_reason }}
            </p>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <button
            v-if="row.blocked"
            type="button"
            class="btn btn-secondary btn-sm whitespace-nowrap"
            :disabled="submittingIP === row.ip_address"
            @click="pendingUnblockIP = row.ip_address"
          >
            {{ t('admin.users.ipRisk.unblock') }}
          </button>
          <button
            v-else
            type="button"
            class="btn btn-danger btn-sm whitespace-nowrap"
            :disabled="submittingIP === row.ip_address"
            @click="beginBlock(row.ip_address)"
          >
            {{ t('admin.users.ipRisk.block') }}
          </button>
        </template>

        <template #empty>
          <div class="py-5 text-center">
            <Icon name="shield" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
            <p class="font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.users.ipRisk.empty') }}
            </p>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.users.ipRisk.emptyHint') }}
            </p>
          </div>
        </template>
      </DataTable>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.pageSize"
        :show-page-size-selector="false"
        @update:page="handlePageChange"
      />
    </div>
  </BaseDialog>

  <ConfirmDialog
    :show="Boolean(pendingUnblockIP)"
    :title="t('admin.users.ipRisk.unblockTitle')"
    :message="t('admin.users.ipRisk.unblockConfirm', { ip: pendingUnblockIP })"
    @confirm="confirmUnblock"
    @cancel="pendingUnblockIP = ''"
  />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { adminAPI } from '@/api/admin'
import type { RegistrationIPRisk } from '@/api/admin/users'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  show: boolean
  initialIp?: string
}>(), {
  initialIp: ''
})

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'changed'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const risks = ref<RegistrationIPRisk[]>([])
const loading = ref(false)
const manualIP = ref('')
const pendingBlockIP = ref('')
const pendingUnblockIP = ref('')
const blockReason = ref('')
const submittingIP = ref('')
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const columns = computed<Column[]>(() => [
  { key: 'ip_address', label: t('admin.users.ipRisk.columns.ip') },
  { key: 'user_count', label: t('admin.users.ipRisk.columns.count') },
  { key: 'users', label: t('admin.users.ipRisk.columns.users') },
  { key: 'first_registered_at', label: t('admin.users.ipRisk.columns.firstRegistered') },
  { key: 'last_registered_at', label: t('admin.users.ipRisk.columns.lastRegistered') },
  { key: 'blocked', label: t('admin.users.ipRisk.columns.status') },
  { key: 'actions', label: t('admin.users.ipRisk.columns.actions') }
])

const loadRisks = async () => {
  loading.value = true
  try {
    const response = await adminAPI.users.listRegistrationIPRisks(
      pagination.page,
      pagination.pageSize
    )
    risks.value = response.items || []
    pagination.total = response.total || 0
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.users.ipRisk.loadFailed'))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.initialIp] as const,
  ([show, initialIP]) => {
    if (!show) return
    manualIP.value = initialIP || ''
    pendingBlockIP.value = ''
    pendingUnblockIP.value = ''
    blockReason.value = ''
    pagination.page = 1
    void loadRisks()
  },
  { immediate: true }
)

const beginBlock = (ipAddress: string) => {
  const normalized = ipAddress.trim()
  if (!normalized) return
  manualIP.value = normalized
  pendingBlockIP.value = normalized
  blockReason.value = ''
}

const cancelBlock = () => {
  pendingBlockIP.value = ''
  blockReason.value = ''
}

const confirmBlock = async () => {
  if (!pendingBlockIP.value || submittingIP.value) return
  submittingIP.value = pendingBlockIP.value
  try {
    await adminAPI.users.blockRegistrationIP(pendingBlockIP.value, blockReason.value.trim())
    appStore.showSuccess(t('admin.users.ipRisk.blockSuccess', { ip: pendingBlockIP.value }))
    cancelBlock()
    pagination.page = 1
    await loadRisks()
    emit('changed')
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.users.ipRisk.blockFailed'))
  } finally {
    submittingIP.value = ''
  }
}

const confirmUnblock = async () => {
  if (!pendingUnblockIP.value || submittingIP.value) return
  const ipAddress = pendingUnblockIP.value
  pendingUnblockIP.value = ''
  submittingIP.value = ipAddress
  try {
    await adminAPI.users.unblockRegistrationIP(ipAddress)
    appStore.showSuccess(t('admin.users.ipRisk.unblockSuccess', { ip: ipAddress }))
    await loadRisks()
    emit('changed')
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.users.ipRisk.unblockFailed'))
  } finally {
    submittingIP.value = ''
  }
}

const handlePageChange = (page: number) => {
  pagination.page = page
  void loadRisks()
}
</script>
