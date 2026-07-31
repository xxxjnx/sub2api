import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mount, type VueWrapper } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import DocsView from '../DocsView.vue'

const { copyToClipboard } = vi.hoisted(() => ({
  copyToClipboard: vi.fn(),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copied: { value: false },
    copyToClipboard,
  }),
}))

const here = dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(resolve(here, '../DocsView.vue'), 'utf8')
const routerSource = readFileSync(resolve(here, '../../../router/index.ts'), 'utf8')
const sidebarSource = readFileSync(resolve(here, '../../../components/layout/AppSidebar.vue'), 'utf8')

function mountView(): VueWrapper {
  return mount(DocsView, {
    global: {
      stubs: {
        AppLayout: {
          template: '<div><slot /></div>',
        },
      },
    },
  })
}

function buttonByText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((candidate) => candidate.text() === text)
  if (!button) {
    throw new Error(`Button not found: ${text}`)
  }
  return button
}

describe('DocsView branding and navigation', () => {
  it('uses only the SuyAI gateway URLs and omits the removed action buttons', () => {
    expect(componentSource).toContain("const SITE_ROOT = 'https://suyai.fun'")
    expect(componentSource).not.toMatch(/jiji\.cc|吉吉/)
    expect(componentSource).not.toContain('网页对话')
    expect(componentSource).not.toContain('查看密钥')
    expect(componentSource).not.toContain('to="/chat"')
    expect(componentSource).not.toContain('to="/keys"')

    const absoluteUrls = componentSource.match(/https?:\/\/[^'"`\s<]+/g) ?? []
    expect(absoluteUrls).toEqual(['https://suyai.fun'])
  })

  it('registers an authenticated route and a sidebar entry', () => {
    expect(routerSource).toContain("path: '/docs'")
    expect(routerSource).toContain("component: () => import('@/views/user/DocsView.vue')")
    expect(sidebarSource).toContain("{ path: '/docs', label: t('nav.docs'), icon: DocsIcon }")
  })
})

describe('DocsView interactions', () => {
  beforeEach(() => {
    copyToClipboard.mockReset()
  })

  it('shows the Claude quick-start guide by default', () => {
    const wrapper = mountView()

    expect(wrapper.text()).toContain('https://suyai.fun/')
    expect(wrapper.text()).toContain('配置并启动 Claude Code')
    expect(wrapper.text()).toContain('ANTHROPIC_BASE_URL=https://suyai.fun')
    expect(wrapper.text()).toContain('YOUR_MODEL_ID')
  })

  it('switches to the Codex and PowerShell configuration', async () => {
    const wrapper = mountView()

    await buttonByText(wrapper, 'ChatGPT / Codex').trigger('click')
    await buttonByText(wrapper, 'PowerShell').trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('https://suyai.fun/v1')
    expect(wrapper.text()).toContain('配置并启动 Codex CLI')
    expect(wrapper.text()).toContain('model_provider = "suyai"')
    expect(wrapper.text()).toContain('name = "SuyAI"')
    expect(wrapper.text()).toContain('Windows PowerShell 配置文件')
  })

  it('switches to the beginner guide without adding external brand links', async () => {
    const wrapper = mountView()

    await buttonByText(wrapper, '小白教程').trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('安装运行环境 Node.js')
    expect(wrapper.text()).toContain('安装 Claude Code')
    expect(wrapper.text()).toContain('配置接入地址与密钥')
    expect(wrapper.findAll('a')).toHaveLength(0)
  })

  it('copies the selected code block through the shared clipboard helper', async () => {
    const wrapper = mountView()

    await wrapper.get('[aria-label="复制Windows CMD 临时配置"]').trigger('click')

    expect(copyToClipboard).toHaveBeenCalledOnce()
    expect(copyToClipboard).toHaveBeenCalledWith(
      expect.stringContaining('ANTHROPIC_BASE_URL=https://suyai.fun'),
      '内容已复制',
    )
  })
})
