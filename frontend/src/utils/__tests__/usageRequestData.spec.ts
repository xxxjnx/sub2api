import { describe, expect, it } from 'vitest'
import { extractUsageUserMessages } from '../usageRequestData'

describe('extractUsageUserMessages', () => {
  it('extracts only user messages from a Responses API request', () => {
    const raw = JSON.stringify({
      instructions: 'system instructions',
      input: [
        { role: 'developer', content: [{ type: 'input_text', text: 'developer data' }] },
        { role: 'user', content: [{ type: 'input_text', text: 'hi' }] },
        { role: 'assistant', content: [{ type: 'output_text', text: 'hello' }] },
        {
          role: 'user',
          content: [{ type: 'input_text', text: '你觉得1+1在什么时候等于3' }],
        },
      ],
      tools: [{ description: 'tool data' }],
    })

    expect(extractUsageUserMessages(raw)).toEqual([
      { text: 'hi', position: 1 },
      { text: '你觉得1+1在什么时候等于3', position: 2 },
    ])
  })

  it('supports chat, Gemini and direct prompt request shapes', () => {
    expect(extractUsageUserMessages(JSON.stringify({
      messages: [{ role: 'user', content: 'chat question' }],
    }))[0]?.text).toBe('chat question')

    expect(extractUsageUserMessages(JSON.stringify({
      contents: [{ role: 'user', parts: [{ text: 'gemini question' }] }],
    }))[0]?.text).toBe('gemini question')

    expect(extractUsageUserMessages(JSON.stringify({ prompt: 'plain prompt' }))[0]?.text)
      .toBe('plain prompt')
  })

  it('returns no structured messages for invalid JSON', () => {
    expect(extractUsageUserMessages('raw non-json request')).toEqual([])
  })
})
