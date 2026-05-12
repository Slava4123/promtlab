// Селекторы для извлечения текста последнего AI-ответа на странице.
// Используется в Phase 3 chains autocapture (content.captureLastAIResponse).
// Best-effort: при изменении DOM сайтов селекторы могут протухать; UI всегда
// даёт fallback на manual paste.

import type { SupportedHost } from './selectors'

const RESPONSE_SELECTORS: Record<SupportedHost, string[]> = {
  'chatgpt.com': [
    '[data-message-author-role="assistant"]:last-of-type .markdown',
    '[data-message-author-role="assistant"]:last-of-type',
    'div.group:last-of-type .markdown',
  ],
  'claude.ai': [
    'div[data-test-render-count] .font-claude-message:last-of-type',
    'div.font-claude-message:last-of-type',
    'div[data-is-streaming="false"]:last-of-type',
  ],
  'gemini.google.com': [
    'message-content:last-of-type .markdown',
    'model-response:last-of-type .markdown',
    '[data-test-id="conversation-turn"]:last-of-type',
  ],
  'www.perplexity.ai': [
    'div.prose:last-of-type',
    'div[id^="answer"]:last-of-type',
  ],
  'alice.yandex.ru': [
    '[data-testid="message-assistant"]:last-of-type',
    'div.message-assistant:last-of-type',
    'div.YpcMessage:last-of-type',
  ],
  'ya.ru': [
    '[data-testid="message-assistant"]:last-of-type',
    'div.message-assistant:last-of-type',
  ],
  'yandex.ru': [
    '[data-testid="message-assistant"]:last-of-type',
    'div.message-assistant:last-of-type',
    'div.YpcMessage:last-of-type',
  ],
  'giga.chat': [
    'div[data-role="assistant"]:last-of-type',
    'div.message-bot:last-of-type',
  ],
  'developers.sber.ru': [
    'div[data-role="assistant"]:last-of-type',
    'div.assistant-message:last-of-type',
  ],
  'chat.deepseek.com': [
    'div[data-role="assistant"]:last-of-type',
    'div.markdown:last-of-type',
  ],
  'chat.mistral.ai': [
    'div[data-role="assistant"]:last-of-type',
    'article.assistant:last-of-type',
  ],
  'le-chat.mistral.ai': [
    'div[data-role="assistant"]:last-of-type',
    'article.assistant:last-of-type',
  ],
  'chat.qwen.ai': [
    'div[data-role="assistant"]:last-of-type',
    'div.message-assistant:last-of-type',
  ],
}

export function extractLastAIResponse(host: string): string | null {
  const list = RESPONSE_SELECTORS[host as SupportedHost] ?? []
  for (const sel of list) {
    try {
      const el = document.querySelector<HTMLElement>(sel)
      if (el) {
        const text = (el.innerText || el.textContent || '').trim()
        if (text.length > 0) return text
      }
    } catch {
      // Невалидный селектор — продолжаем
    }
  }
  return null
}
