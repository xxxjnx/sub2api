import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserIPRiskModal from '../UserIPRiskModal.vue'

const { listRegistrationIPRisks, blockRegistrationIP, unblockRegistrationIP, showError, showSuccess } = vi.hoisted(() => ({
  listRegistrationIPRisks: vi.fn(),
  blockRegistrationIP: vi.fn(),
  unblockRegistrationIP: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      listRegistrationIPRisks,
      blockRegistrationIP,
      unblockRegistrationIP
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
}

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div v-for="row in data" :key="row.ip_address" data-test="risk-row">
        <slot name="cell-actions" :row="row" :value="null" />
      </div>
      <slot v-if="!loading && data.length === 0" name="empty" />
    </div>
  `
}

const ConfirmDialogStub = {
  props: ['show', 'title', 'message'],
  emits: ['confirm', 'cancel'],
  template: '<button v-if="show" data-test="confirm-unblock" @click="$emit(\'confirm\')">confirm</button>'
}

const mountModal = (initialIp = '') => mount(UserIPRiskModal, {
  props: { show: true, initialIp },
  global: {
    stubs: {
      BaseDialog: BaseDialogStub,
      DataTable: DataTableStub,
      Pagination: true,
      ConfirmDialog: ConfirmDialogStub,
      Icon: true
    }
  }
})

describe('UserIPRiskModal', () => {
  beforeEach(() => {
    listRegistrationIPRisks.mockReset()
    blockRegistrationIP.mockReset()
    unblockRegistrationIP.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    listRegistrationIPRisks.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 1
    })
    blockRegistrationIP.mockResolvedValue({
      id: 1,
      ip_address: '203.0.113.10',
      reason: 'duplicate registrations',
      created_at: '2026-08-01T00:00:00Z',
      updated_at: '2026-08-01T00:00:00Z'
    })
    unblockRegistrationIP.mockResolvedValue({ ip_address: '203.0.113.10' })
  })

  it('prefills a user registration IP and blocks it with an optional reason', async () => {
    const wrapper = mountModal('203.0.113.10')
    await flushPromises()

    expect(listRegistrationIPRisks).toHaveBeenCalledWith(1, 20)
    expect(wrapper.get('[data-test="registration-ip-input"]').element).toHaveProperty('value', '203.0.113.10')

    await wrapper.get('[data-test="prepare-registration-ip-block"]').trigger('click')
    await wrapper.get('[data-test="registration-ip-block-reason"]').setValue('duplicate registrations')
    await wrapper.get('[data-test="confirm-registration-ip-block"]').trigger('click')
    await flushPromises()

    expect(blockRegistrationIP).toHaveBeenCalledWith('203.0.113.10', 'duplicate registrations')
    expect(wrapper.emitted('changed')).toHaveLength(1)
    expect(showSuccess).toHaveBeenCalled()
  })

  it('unblocks an IP listed in the risk table after confirmation', async () => {
    listRegistrationIPRisks.mockResolvedValue({
      items: [{
        ip_address: '203.0.113.10',
        user_count: 2,
        users: [],
        blocked: true,
        block_reason: 'duplicate registrations'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    const wrapper = mountModal()
    await flushPromises()

    const unblockButton = wrapper.findAll('button').find((button) => button.text() === 'admin.users.ipRisk.unblock')
    expect(unblockButton).toBeTruthy()
    await unblockButton!.trigger('click')
    await wrapper.get('[data-test="confirm-unblock"]').trigger('click')
    await flushPromises()

    expect(unblockRegistrationIP).toHaveBeenCalledWith('203.0.113.10')
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })
})
