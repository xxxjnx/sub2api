export interface UsageUserMessage {
  text: string
  position: number
}

type UnknownRecord = Record<string, unknown>

function isRecord(value: unknown): value is UnknownRecord {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function extractTextParts(value: unknown): string[] {
  if (typeof value === 'string') {
    return value.trim() ? [value.trim()] : []
  }
  if (Array.isArray(value)) {
    return value.flatMap(extractTextParts)
  }
  if (!isRecord(value)) {
    return []
  }

  if (typeof value.text === 'string') {
    return value.text.trim() ? [value.text.trim()] : []
  }
  if ('content' in value) {
    return extractTextParts(value.content)
  }
  if ('parts' in value) {
    return extractTextParts(value.parts)
  }
  return []
}

function appendRoleMessages(
  target: UsageUserMessage[],
  value: unknown,
): void {
  if (!Array.isArray(value)) return

  for (const item of value) {
    if (!isRecord(item) || item.role !== 'user') continue
    const text = extractTextParts(item.content ?? item.parts)
      .join('\n\n')
      .trim()
    if (!text) continue
    target.push({ text, position: target.length + 1 })
  }
}

export function extractUsageUserMessages(rawRequest: string | null | undefined): UsageUserMessage[] {
  if (!rawRequest) return []

  let payload: unknown
  try {
    payload = JSON.parse(rawRequest)
  } catch {
    return []
  }
  if (!isRecord(payload)) return []

  const messages: UsageUserMessage[] = []
  appendRoleMessages(messages, payload.input)
  appendRoleMessages(messages, payload.messages)
  appendRoleMessages(messages, payload.contents)

  if (messages.length === 0) {
    const directInput = typeof payload.input === 'string' ? payload.input.trim() : ''
    const prompt = typeof payload.prompt === 'string' ? payload.prompt.trim() : ''
    const text = directInput || prompt
    if (text) messages.push({ text, position: 1 })
  }

  return messages
}
