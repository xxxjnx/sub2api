<template>
  <AppLayout>
    <div class="docs-page mx-auto max-w-7xl space-y-6">
      <header>
        <h1 class="page-title">使用文档</h1>
        <p class="page-description">选择模型平台、教程模式和操作系统，按步骤完成接入。</p>
      </header>

      <section class="card p-4 sm:p-5">
        <div class="grid min-w-0 gap-4 xl:grid-cols-[auto_auto_1fr] xl:items-end">
          <div class="min-w-0">
            <p class="docs-control-label">模型平台</p>
            <div class="docs-segment">
              <button
                type="button"
                :class="{ active: platform === 'claude' }"
                :aria-pressed="platform === 'claude'"
                @click="platform = 'claude'"
              >
                Claude
              </button>
              <button
                type="button"
                :class="{ active: platform === 'codex' }"
                :aria-pressed="platform === 'codex'"
                @click="platform = 'codex'"
              >
                ChatGPT / Codex
              </button>
            </div>
          </div>

          <div class="min-w-0">
            <p class="docs-control-label">教程模式</p>
            <div class="docs-segment">
              <button
                type="button"
                :class="{ active: tutorialMode === 'quick' }"
                :aria-pressed="tutorialMode === 'quick'"
                @click="tutorialMode = 'quick'"
              >
                快速接入
              </button>
              <button
                type="button"
                :class="{ active: tutorialMode === 'beginner' }"
                :aria-pressed="tutorialMode === 'beginner'"
                @click="tutorialMode = 'beginner'"
              >
                小白教程
              </button>
            </div>
          </div>

          <div class="min-w-0 xl:justify-self-end">
            <p class="docs-control-label">系统 / 终端</p>
            <div class="docs-segment docs-os-segment">
              <button
                type="button"
                :class="{ active: operatingSystem === 'cmd' }"
                :aria-pressed="operatingSystem === 'cmd'"
                @click="operatingSystem = 'cmd'"
              >
                Windows CMD
              </button>
              <button
                type="button"
                :class="{ active: operatingSystem === 'powershell' }"
                :aria-pressed="operatingSystem === 'powershell'"
                @click="operatingSystem = 'powershell'"
              >
                PowerShell
              </button>
              <button
                type="button"
                :class="{ active: operatingSystem === 'unix' }"
                :aria-pressed="operatingSystem === 'unix'"
                @click="operatingSystem = 'unix'"
              >
                macOS / Linux
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="item in overviewItems"
          :key="item.label"
          class="card p-5"
        >
          <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ item.label }}</p>
          <p class="mt-2 break-all font-mono text-sm font-semibold text-gray-900 dark:text-white">
            {{ item.value }}
          </p>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ item.hint }}</p>
        </article>
      </section>

      <section v-if="tutorialMode === 'quick'" class="card overflow-hidden">
        <div class="docs-section-head">
          <div>
            <span class="docs-step-badge">CLI</span>
            <h2>{{ cliTitle }}</h2>
            <p>{{ cliDescription }}</p>
          </div>
          <span class="docs-platform-tag">{{ operatingSystemLabel }}</span>
        </div>
        <div class="p-4 sm:p-6">
          <DocsCodeBlock
            :code="quickStartCode"
            :label="quickStartLabel"
            @copy="copyCode"
          />
          <div class="mt-5 rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm leading-6 text-gray-600 dark:border-dark-700 dark:bg-dark-900/50 dark:text-dark-300">
            <strong class="text-gray-900 dark:text-white">验证方式：</strong>
            {{ validationText }}
          </div>
        </div>
      </section>

      <section v-else class="space-y-4">
        <article
          v-for="(step, index) in beginnerSteps"
          :key="step.title"
          class="card overflow-hidden"
        >
          <div class="docs-section-head">
            <div class="flex min-w-0 items-start gap-4">
              <span class="docs-step-number">{{ index + 1 }}</span>
              <div class="min-w-0">
                <h2>{{ step.title }}</h2>
                <p>{{ step.description }}</p>
              </div>
            </div>
            <span v-if="step.badge" class="docs-platform-tag">{{ step.badge }}</span>
          </div>
          <div class="space-y-4 p-4 sm:p-6">
            <div
              v-if="step.note"
              class="rounded-xl border border-primary-100 bg-primary-50/70 p-4 text-sm leading-6 text-primary-800 dark:border-primary-900/60 dark:bg-primary-950/30 dark:text-primary-200"
            >
              {{ step.note }}
            </div>
            <DocsCodeBlock
              v-for="block in step.blocks"
              :key="block.label"
              :code="block.code"
              :label="block.label"
              @copy="copyCode"
            />
          </div>
        </article>
      </section>

      <section class="card overflow-hidden">
        <div class="docs-section-head">
          <div>
            <span class="docs-step-badge">DESKTOP</span>
            <h2>{{ desktopTitle }}</h2>
            <p>{{ desktopGuide.description }}</p>
          </div>
        </div>
        <div class="grid gap-4 p-4 sm:p-6 lg:grid-cols-3">
          <article
            v-for="(step, index) in desktopGuide.steps"
            :key="step.title"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/50"
          >
            <span class="text-xs font-bold text-primary-600 dark:text-primary-400">
              0{{ index + 1 }}
            </span>
            <h3 class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">{{ step.title }}</h3>
            <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ step.text }}</p>
          </article>
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="docs-section-head">
          <div>
            <span class="docs-step-badge">API</span>
            <h2>在你自己的程序里调用 API</h2>
            <p>{{ apiDescription }}</p>
          </div>
          <div class="docs-segment docs-code-tabs">
            <button
              v-for="tab in codeTabs"
              :key="tab.value"
              type="button"
              :class="{ active: codeTab === tab.value }"
              :aria-pressed="codeTab === tab.value"
              @click="codeTab = tab.value"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>
        <div class="space-y-4 p-4 sm:p-6">
          <p v-if="sdkInstallCommand" class="text-sm text-gray-500 dark:text-dark-400">
            先安装官方 SDK：
            <code class="ml-1 rounded bg-gray-100 px-2 py-1 text-primary-700 dark:bg-dark-700 dark:text-primary-300">
              {{ sdkInstallCommand }}
            </code>
          </p>
          <DocsCodeBlock
            :code="apiExample"
            label="完整调用示例"
            @copy="copyCode"
          />
          <div
            v-if="platform === 'codex'"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm leading-6 text-gray-600 dark:border-dark-700 dark:bg-dark-900/50 dark:text-dark-300"
          >
            多轮对话请把历史消息依次放入 <code>input</code> 数组；是否支持
            <code>store</code> 与 <code>previous_response_id</code> 取决于实际上游能力。
          </div>
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="docs-section-head">
          <div>
            <span class="docs-step-badge">TROUBLESHOOTING</span>
            <h2>常见错误码</h2>
            <p>快速定位鉴权、余额、模型、限流与上游连接问题。</p>
          </div>
        </div>
        <div class="grid gap-3 p-4 sm:grid-cols-2 sm:p-6 xl:grid-cols-3">
          <article
            v-for="item in errorCodes"
            :key="item.code"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/50"
          >
            <code class="text-sm font-bold text-primary-600 dark:text-primary-400">{{ item.code }}</code>
            <h3 class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>
            <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ item.hint }}</p>
          </article>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DocsCodeBlock from '@/components/docs/DocsCodeBlock.vue'
import { useClipboard } from '@/composables/useClipboard'

type Platform = 'claude' | 'codex'
type TutorialMode = 'quick' | 'beginner'
type OperatingSystem = 'cmd' | 'powershell' | 'unix'
type CodeTab = 'curl' | 'python' | 'node'

interface InstructionBlock {
  label: string
  code: string
}

interface BeginnerStep {
  title: string
  description: string
  badge?: string
  note?: string
  blocks: InstructionBlock[]
}

const SITE_ROOT = 'https://suyai.fun'
const SITE_ROOT_DISPLAY = `${SITE_ROOT}/`
const OPENAI_BASE_URL = `${SITE_ROOT}/v1`
const API_KEY_PLACEHOLDER = 'sk-你的API密钥'
const MODEL_ID_PLACEHOLDER = 'YOUR_MODEL_ID'

const platform = ref<Platform>('claude')
const tutorialMode = ref<TutorialMode>('quick')
const operatingSystem = ref<OperatingSystem>('cmd')
const codeTab = ref<CodeTab>('curl')
const { copyToClipboard } = useClipboard()

const operatingSystemLabel = computed(
  () =>
    ({
      cmd: 'Windows CMD',
      powershell: 'Windows PowerShell',
      unix: 'macOS / Linux',
    })[operatingSystem.value],
)

const overviewItems = computed(() =>
  platform.value === 'claude'
    ? [
        {
          label: 'Base URL',
          value: SITE_ROOT_DISPLAY,
          hint: 'Claude Code 和 Anthropic SDK 使用站点根地址。',
        },
        {
          label: '鉴权',
          value: 'x-api-key / AUTH_TOKEN',
          hint: '使用你的 API 密钥完成鉴权。',
        },
        {
          label: '协议',
          value: 'Anthropic Messages · SSE',
          hint: '原生端点 /v1/messages。',
        },
        {
          label: '模型',
          value: '以账户可用模型为准',
          hint: '请将示例中的模型占位符替换为实际可用模型 ID。',
        },
      ]
    : [
        {
          label: 'Base URL',
          value: OPENAI_BASE_URL,
          hint: 'Codex 和 OpenAI SDK 使用 /v1 地址。',
        },
        {
          label: '鉴权',
          value: 'Bearer / OPENAI_API_KEY',
          hint: '使用你的 API 密钥完成鉴权。',
        },
        {
          label: '协议',
          value: 'OpenAI Responses',
          hint: '主要端点 /v1/responses。',
        },
        {
          label: '模型',
          value: '以账户可用模型为准',
          hint: '请将示例中的模型占位符替换为实际可用模型 ID。',
        },
      ],
)

const cliTitle = computed(() =>
  platform.value === 'claude' ? '配置并启动 Claude Code' : '配置并启动 Codex CLI',
)

const cliDescription = computed(() =>
  platform.value === 'claude'
    ? '设置两个环境变量后直接启动，无需跳转官方登录。'
    : '创建 config.toml 与 auth.json 后直接启动，无需执行 codex login。',
)

const quickStartLabel = computed(() =>
  platform.value === 'claude'
    ? `${operatingSystemLabel.value} 临时配置`
    : `${operatingSystemLabel.value} 配置文件`,
)

const claudeQuickStart = computed<Record<OperatingSystem, string>>(() => ({
  cmd: `set ANTHROPIC_BASE_URL=${SITE_ROOT}
set ANTHROPIC_AUTH_TOKEN=${API_KEY_PLACEHOLDER}
claude`,
  powershell: `$env:ANTHROPIC_BASE_URL="${SITE_ROOT}"
$env:ANTHROPIC_AUTH_TOKEN="${API_KEY_PLACEHOLDER}"
claude`,
  unix: `export ANTHROPIC_BASE_URL=${SITE_ROOT}
export ANTHROPIC_AUTH_TOKEN=${API_KEY_PLACEHOLDER}
claude`,
}))

const codexConfig = computed(
  () => `model = "${MODEL_ID_PLACEHOLDER}"
model_provider = "suyai"
model_reasoning_effort = "high"

[model_providers.suyai]
name = "SuyAI"
base_url = "${OPENAI_BASE_URL}"
wire_api = "responses"
requires_openai_auth = true`,
)

const codexAuth = `{
  "OPENAI_API_KEY": "${API_KEY_PLACEHOLDER}"
}`

const quickStartCode = computed(() => {
  if (platform.value === 'claude') {
    return claudeQuickStart.value[operatingSystem.value]
  }

  return `${codexConfig.value}

# auth.json
${codexAuth}

# 保存后启动
codex`
})

const validationText = computed(() =>
  platform.value === 'claude'
    ? '看到 Claude Code 对话界面后发送一句测试消息；无需执行 claude login。'
    : '进入项目目录执行 codex，能打开交互界面即配置成功。',
)

const nodeInstallBlock = computed<InstructionBlock>(() => {
  if (operatingSystem.value === 'unix') {
    return {
      label: '安装并验证 Node.js',
      code: `# macOS 已安装 Homebrew
brew install node

# 验证
node -v
npm -v`,
    }
  }

  if (operatingSystem.value === 'powershell') {
    return {
      label: '安装并验证 Node.js',
      code: `winget install OpenJS.NodeJS.LTS

# 关闭并重新打开 PowerShell 后验证
node -v
npm -v`,
    }
  }

  return {
    label: '安装并验证 Node.js',
    code: `winget install OpenJS.NodeJS.LTS

REM 关闭并重新打开 CMD 后验证
node -v
npm -v`,
  }
})

const claudePermanentConfig = computed<InstructionBlock>(() => {
  if (operatingSystem.value === 'cmd') {
    return {
      label: '永久环境变量',
      code: `setx ANTHROPIC_BASE_URL "${SITE_ROOT}"
setx ANTHROPIC_AUTH_TOKEN "${API_KEY_PLACEHOLDER}"

REM 设置后重新打开 CMD`,
    }
  }

  if (operatingSystem.value === 'powershell') {
    return {
      label: '永久环境变量',
      code: `[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL", "${SITE_ROOT}", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN", "${API_KEY_PLACEHOLDER}", "User")

# 设置后重新打开 PowerShell`,
    }
  }

  return {
    label: '写入 Shell 配置',
    code: `# macOS 默认 zsh；bash 用户改为 ~/.bashrc
echo 'export ANTHROPIC_BASE_URL=${SITE_ROOT}' >> ~/.zshrc
echo 'export ANTHROPIC_AUTH_TOKEN=${API_KEY_PLACEHOLDER}' >> ~/.zshrc
source ~/.zshrc`,
  }
})

const codexConfigLocation = computed<InstructionBlock>(() => {
  if (operatingSystem.value === 'unix') {
    return {
      label: '配置文件位置',
      code: `mkdir -p ~/.codex
# 编辑 ~/.codex/config.toml
# 编辑 ~/.codex/auth.json`,
    }
  }

  if (operatingSystem.value === 'powershell') {
    return {
      label: '配置文件位置',
      code: `New-Item -ItemType Directory -Force "$env:USERPROFILE\\.codex"
notepad "$env:USERPROFILE\\.codex\\config.toml"
notepad "$env:USERPROFILE\\.codex\\auth.json"`,
    }
  }

  return {
    label: '配置文件位置',
    code: `if not exist "%USERPROFILE%\\.codex" mkdir "%USERPROFILE%\\.codex"
notepad "%USERPROFILE%\\.codex\\config.toml"
notepad "%USERPROFILE%\\.codex\\auth.json"`,
  }
})

const beginnerSteps = computed<BeginnerStep[]>(() => {
  const installRuntime: BeginnerStep = {
    title: '安装运行环境 Node.js',
    description: '请安装当前 Node.js LTS 版本。',
    badge: '首次必做',
    note: '安装完成后如果命令未识别，请完全关闭终端再重新打开，让 PATH 生效。',
    blocks: [nodeInstallBlock.value],
  }

  if (platform.value === 'claude') {
    return [
      installRuntime,
      {
        title: '安装 Claude Code',
        description: '使用 npm 安装官方 Claude Code CLI。',
        note: '安装完成后可先运行版本命令，确认终端已能识别 Claude Code。',
        blocks: [
          {
            label: '安装命令',
            code: `npm install -g @anthropic-ai/claude-code
claude --version`,
          },
        ],
      },
      {
        title: '配置接入地址与密钥',
        description: `让 Claude Code 通过 ${SITE_ROOT_DISPLAY} 调用当前网关。`,
        blocks: [
          claudePermanentConfig.value,
          {
            label: '仅当前窗口临时生效',
            code: claudeQuickStart.value[operatingSystem.value].replace(/\nclaude$/, ''),
          },
        ],
      },
      {
        title: '启动并开始对话',
        description: '进入你的项目目录启动 Claude Code。',
        note: '使用当前配置时不需要 claude login，也不会跳转官方登录网页。',
        blocks: [
          {
            label: '启动命令',
            code: `cd 你的项目目录
claude`,
          },
        ],
      },
    ]
  }

  return [
    installRuntime,
    {
      title: '安装 Codex CLI',
      description: '使用 npm 安装官方 Codex CLI。',
      note:
        operatingSystem.value === 'unix'
          ? 'macOS 也可执行 brew install codex，不需要 Node.js。'
          : undefined,
      blocks: [
        {
          label: '安装命令',
          code: `npm install -g @openai/codex
codex --version`,
        },
      ],
    },
    {
      title: '创建 Codex 配置目录',
      description: 'config.toml 和 auth.json 位于同一个 .codex 目录。',
      blocks: [codexConfigLocation.value],
    },
    {
      title: '填写接口配置',
      description: '将模型占位符换成当前账户实际可用的模型 ID。',
      blocks: [{ label: 'config.toml', code: codexConfig.value }],
    },
    {
      title: '填写密钥并启动',
      description: '保存 auth.json 后进入项目目录启动。',
      note: '配置完成后无需 codex login。auth.json 包含敏感凭据，请勿提交到代码仓库或分享。',
      blocks: [
        { label: 'auth.json', code: codexAuth },
        {
          label: '启动命令',
          code: `cd 你的项目目录
codex`,
        },
      ],
    },
  ]
})

const desktopTitle = computed(() =>
  platform.value === 'claude' ? '接入 Claude 桌面端' : '接入 Codex 桌面端',
)

const desktopGuide = computed(() =>
  platform.value === 'claude'
    ? {
        description: 'Claude 官方桌面客户端可通过第三方推理设置连接当前网关。',
        steps: [
          {
            title: '开启开发者模式',
            text: '在 Help → Troubleshooting 中启用 Developer Mode，然后按提示重启客户端。',
          },
          {
            title: '打开第三方推理',
            text: '重启后进入 Developer → Configure Third-Party Inference。',
          },
          {
            title: '填写连接参数',
            text: `Credential 选择 Static API key，Gateway Base URL 填 ${SITE_ROOT_DISPLAY}，粘贴你的 API Key，并添加实际可用的模型 ID。`,
          },
        ],
      }
    : {
        description: 'Codex 桌面端与 Codex CLI 共用同一套本地配置。',
        steps: [
          {
            title: '完成 CLI 配置',
            text: '先按上方教程创建 config.toml，并设置 Responses API 与实际可用模型。',
          },
          {
            title: '写入 auth.json',
            text: '在同一 .codex 目录写入 OPENAI_API_KEY。',
          },
          {
            title: '启动桌面端',
            text: '直接打开 Codex 桌面端即可读取配置，无需额外登录。',
          },
        ],
      },
)

const codeTabs: Array<{ value: CodeTab; label: string }> = [
  { value: 'curl', label: 'cURL' },
  { value: 'python', label: 'Python' },
  { value: 'node', label: 'Node.js' },
]

const sdkInstallCommand = computed(() => {
  if (codeTab.value === 'python') {
    return platform.value === 'claude' ? 'pip install anthropic' : 'pip install openai'
  }
  if (codeTab.value === 'node') {
    return platform.value === 'claude' ? 'npm install @anthropic-ai/sdk' : 'npm install openai'
  }
  return ''
})

const apiDescription = computed(() =>
  platform.value === 'claude'
    ? '使用 Anthropic 原生 Messages API。'
    : '使用 OpenAI Responses API，支持流式与非流式。',
)

const apiExamples = computed<Record<Platform, Record<CodeTab, string>>>(() => ({
  claude: {
    curl: `curl ${SITE_ROOT}/v1/messages \\
  -H "x-api-key: ${API_KEY_PLACEHOLDER}" \\
  -H "anthropic-version: 2023-06-01" \\
  -H "content-type: application/json" \\
  -d '{"model":"${MODEL_ID_PLACEHOLDER}","max_tokens":1024,"messages":[{"role":"user","content":"你好"}]}'`,
    python: `import anthropic

client = anthropic.Anthropic(
    api_key="${API_KEY_PLACEHOLDER}",
    base_url="${SITE_ROOT}",
)
message = client.messages.create(
    model="${MODEL_ID_PLACEHOLDER}",
    max_tokens=1024,
    messages=[{"role": "user", "content": "你好"}],
)
print(message.content[0].text)`,
    node: `import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic({
  apiKey: "${API_KEY_PLACEHOLDER}",
  baseURL: "${SITE_ROOT}",
});
const message = await client.messages.create({
  model: "${MODEL_ID_PLACEHOLDER}",
  max_tokens: 1024,
  messages: [{ role: "user", content: "你好" }],
});
console.log(message.content[0].text);`,
  },
  codex: {
    curl: `curl ${OPENAI_BASE_URL}/responses \\
  -H "Authorization: Bearer ${API_KEY_PLACEHOLDER}" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"${MODEL_ID_PLACEHOLDER}","input":"你好"}'`,
    python: `from openai import OpenAI

client = OpenAI(
    api_key="${API_KEY_PLACEHOLDER}",
    base_url="${OPENAI_BASE_URL}",
)
response = client.responses.create(
    model="${MODEL_ID_PLACEHOLDER}",
    input="你好",
)
print(response.output_text)`,
    node: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "${API_KEY_PLACEHOLDER}",
  baseURL: "${OPENAI_BASE_URL}",
});
const response = await client.responses.create({
  model: "${MODEL_ID_PLACEHOLDER}",
  input: "你好",
});
console.log(response.output_text);`,
  },
}))

const apiExample = computed(() => apiExamples.value[platform.value][codeTab.value])

const errorCodes = [
  {
    code: '400',
    title: '请求格式错误',
    hint: '检查 JSON、模型参数和所使用的 API 协议是否匹配。',
  },
  {
    code: '401',
    title: '密钥无效',
    hint: '检查 API Key 是否完整、启用且未过期，确认鉴权请求头正确。',
  },
  {
    code: '402',
    title: '余额或权益不足',
    hint: '补充余额、兑换权益码或购买有效套餐。',
  },
  {
    code: '403',
    title: '无模型或分组权限',
    hint: '当前密钥可能未绑定可用分组，或模型不在允许列表。',
  },
  {
    code: '404',
    title: '端点或模型不存在',
    hint: 'Claude 使用 /v1/messages，Codex 使用 /v1/responses；模型 ID 以账户实际可用列表为准。',
  },
  {
    code: '429',
    title: '请求过快',
    hint: '降低瞬时并发，等待 Retry-After 后重试，避免重复快速提交。',
  },
  {
    code: '5xx',
    title: '链路异常',
    hint: '稍后重试；持续出现时记录请求 ID 并联系管理员。',
  },
]

function copyCode(code: string): void {
  void copyToClipboard(code, '内容已复制')
}
</script>

<style scoped>
.docs-control-label {
  margin-bottom: 0.4375rem;
  color: #6b7280;
  font-size: 0.75rem;
  font-weight: 650;
}

.docs-segment {
  display: inline-flex;
  max-width: 100%;
  gap: 0.1875rem;
  overflow-x: auto;
  border: 1px solid #d1d5db;
  border-radius: 0.6875rem;
  background: #f3f4f6;
  padding: 0.1875rem;
}

.docs-segment button {
  flex: none;
  border: 0;
  border-radius: 0.5rem;
  background: transparent;
  padding: 0.5rem 0.8125rem;
  color: #6b7280;
  font-size: 0.75rem;
  font-weight: 650;
  white-space: nowrap;
  transition:
    background-color 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.docs-segment button:hover {
  color: #111827;
}

.docs-segment button.active {
  background: #fff;
  color: #0d9488;
  box-shadow: 0 1px 4px rgba(15, 23, 42, 0.12);
}

.docs-segment button:focus-visible {
  outline: 2px solid rgba(20, 184, 166, 0.45);
  outline-offset: 1px;
}

.docs-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid #e5e7eb;
  background: rgba(249, 250, 251, 0.78);
  padding: 1.125rem 1.375rem;
}

.docs-section-head h2 {
  margin-top: 0.375rem;
  color: #111827;
  font-size: 1.0625rem;
  font-weight: 700;
}

.docs-section-head p {
  margin-top: 0.3125rem;
  color: #6b7280;
  font-size: 0.8125rem;
  line-height: 1.6;
}

.docs-step-badge {
  color: #0d9488;
  font: 700 0.625rem ui-monospace, Menlo, Monaco, Consolas, monospace;
}

.docs-platform-tag {
  flex: none;
  border: 1px solid #99f6e4;
  border-radius: 100px;
  background: #f0fdfa;
  padding: 0.3125rem 0.5625rem;
  color: #0f766e;
  font-size: 0.6875rem;
  font-weight: 700;
}

.docs-step-number {
  display: grid;
  width: 2.125rem;
  height: 2.125rem;
  flex: none;
  place-items: center;
  border: 1px solid #99f6e4;
  border-radius: 50%;
  background: #f0fdfa;
  color: #0f766e;
  font-weight: 800;
}

.docs-code-tabs {
  align-self: center;
}

:global(.dark) .docs-control-label {
  color: #94a3b8;
}

:global(.dark) .docs-segment {
  border-color: #334155;
  background: #0f172a;
}

:global(.dark) .docs-segment button {
  color: #94a3b8;
}

:global(.dark) .docs-segment button:hover {
  color: #f8fafc;
}

:global(.dark) .docs-segment button.active {
  background: #334155;
  color: #5eead4;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:global(.dark) .docs-section-head {
  border-color: #334155;
  background: rgba(15, 23, 42, 0.55);
}

:global(.dark) .docs-section-head h2 {
  color: #fff;
}

:global(.dark) .docs-section-head p {
  color: #94a3b8;
}

:global(.dark) .docs-step-badge {
  color: #5eead4;
}

:global(.dark) .docs-platform-tag,
:global(.dark) .docs-step-number {
  border-color: rgba(20, 184, 166, 0.45);
  background: rgba(4, 47, 46, 0.65);
  color: #5eead4;
}

@media (max-width: 640px) {
  .docs-section-head {
    flex-direction: column;
    padding: 1rem;
  }

  .docs-section-head > .docs-segment {
    align-self: stretch;
  }

  .docs-section-head > .docs-segment button {
    flex: 1;
  }

  .docs-os-segment {
    display: flex;
    width: 100%;
  }
}
</style>
