const apiMocks = vi.hoisted(() => ({
  getAdminById: vi.fn(),
  refund: vi.fn(),
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    getById: apiMocks.getAdminById,
    refund: apiMocks.refund,
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UsageDetailModal from '../UsageDetailModal.vue'

const rawRequest = '{\n  "authorization": "Bearer raw-secret",\n  "api_key": "sk-visible"\n}'

const usageDetail = {
  id: 42,
  user_id: 7,
  api_key_id: 9,
  account_id: 11,
  request_id: 'req-detail-42',
  model: 'gpt-5',
  group_id: null,
  subscription_id: null,
  input_tokens: 1,
  output_tokens: 2,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  input_cost: 0,
  output_cost: 0,
  cache_creation_cost: 0,
  cache_read_cost: 0,
  total_cost: 0,
  actual_cost: 0,
  rate_multiplier: 1,
  long_context_billing_applied: false,
  billing_type: 1,
  request_type: 'sync',
  stream: false,
  duration_ms: 10,
  first_token_ms: null,
  image_count: 0,
  image_size: null,
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
  image_input_tokens: 0,
  image_input_cost: 0,
  image_output_tokens: 0,
  image_output_cost: 0,
  user_agent: 'test-agent',
  ip_address: '203.0.113.7',
  cache_ttl_overridden: false,
  created_at: '2026-07-28T10:00:00Z',
  request_data: rawRequest,
  request_data_encoding: 'utf-8',
  request_content_type: 'application/json',
}

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
}

describe('UsageDetailModal', () => {
  beforeEach(() => {
    apiMocks.getAdminById.mockReset()
    apiMocks.refund.mockReset()
  })

  it('refunds a paid record once with a required reason', async () => {
    apiMocks.getAdminById.mockResolvedValue({ ...usageDetail, actual_cost: 1.25 })
    apiMocks.refund.mockResolvedValue({
      ...usageDetail,
      actual_cost: 0,
      refund_amount: 1.25,
      refund_reason: 'billing error',
      refunded_at: '2026-07-28T11:00:00Z',
    })
    const wrapper = mount(UsageDetailModal, {
      props: { show: true, usageId: 42 },
      global: { stubs: { BaseDialog: BaseDialogStub } },
    })
    await flushPromises()

    await wrapper.get('[data-testid="refund-button"]').trigger('click')
    await wrapper.get('[data-testid="refund-reason"]').setValue('billing error')
    await wrapper.get('[data-testid="confirm-refund-button"]').trigger('click')
    await flushPromises()

    expect(apiMocks.refund).toHaveBeenCalledWith(42, 'billing error')
    expect(wrapper.emitted('refunded')).toHaveLength(1)
    expect(wrapper.find('[data-testid="refund-button"]').exists()).toBe(false)
  })

  it('loads from the admin endpoint and shows the original request without redaction or JSON reformatting', async () => {
    apiMocks.getAdminById.mockResolvedValue({ ...usageDetail })

    const wrapper = mount(UsageDetailModal, {
      props: {
        show: true,
        usageId: 42,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })
    await flushPromises()

    expect(apiMocks.getAdminById).toHaveBeenCalledWith(42)
    expect(wrapper.get('[data-testid="usage-request-data"]').element.textContent).toBe(rawRequest)
    expect(wrapper.text()).toContain('raw-secret')
    expect(wrapper.text()).toContain('sk-visible')
  })

  it('shows administrator-only upstream and account context', async () => {
    apiMocks.getAdminById.mockResolvedValue({
      ...usageDetail,
      upstream_model: 'gpt-5-upstream',
      account: { id: 11, name: 'Account 11' },
    })

    const wrapper = mount(UsageDetailModal, {
      props: {
        show: true,
        usageId: 42,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })
    await flushPromises()

    expect(apiMocks.getAdminById).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('gpt-5-upstream')
    expect(wrapper.text()).toContain('Account 11')
  })
})
