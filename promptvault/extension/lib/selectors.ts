// Per-site селекторы input-полей. Атрибутные, re-query каждый раз
// (никогда не кэшировать references — SPA churn).

import { addBreadcrumb } from './sentry'

export type SupportedHost =
  | 'chatgpt.com'
  | 'claude.ai'
  | 'gemini.google.com'
  | 'www.perplexity.ai'
  | 'alice.yandex.ru'
  | 'ya.ru'
  | 'yandex.ru'
  | 'giga.chat'
  | 'developers.sber.ru'
  | 'chat.deepseek.com'
  | 'chat.mistral.ai'
  | 'le-chat.mistral.ai'
  | 'chat.qwen.ai'

export const SUPPORTED_HOSTS: SupportedHost[] = [
  'chatgpt.com',
  'claude.ai',
  'gemini.google.com',
  'www.perplexity.ai',
  'alice.yandex.ru',
  'ya.ru',
  'yandex.ru',
  'giga.chat',
  'developers.sber.ru',
  'chat.deepseek.com',
  'chat.mistral.ai',
  'le-chat.mistral.ai',
  'chat.qwen.ai',
]

export const INPUT_SELECTORS: Record<SupportedHost, string[]> = {
  'chatgpt.com': [
    '#prompt-textarea',
    'textarea[data-id="root"]',
    'div[contenteditable="true"][data-id]',
    'div.ProseMirror[contenteditable="true"]',
  ],
  'claude.ai': [
    'div[contenteditable="true"].ProseMirror',
    'div[contenteditable="true"][data-testid]',
    'div[contenteditable="true"][aria-label]',
  ],
  'gemini.google.com': [
    'rich-textarea .ql-editor',
    'rich-textarea .ProseMirror',
    'rich-textarea [contenteditable="true"]',
    '.input-area [contenteditable="true"]',
  ],
  'www.perplexity.ai': [
    '#ask-input',
    'div[data-lexical-editor="true"]',
    'div[contenteditable="true"][role="textbox"]',
  ],
  // Yandex GPT — ProseMirror в большинстве чатов
  'alice.yandex.ru': [
    'textarea[data-testid="chat-input"]',
    'div[contenteditable="true"][role="textbox"]',
    'textarea.YpcInput',
    'div.ProseMirror[contenteditable="true"]',
  ],
  'ya.ru': [
    'textarea[data-testid="chat-input"]',
    'div[contenteditable="true"][role="textbox"]',
    'div.ProseMirror[contenteditable="true"]',
    'textarea',
  ],
  // yandex.ru/alice — новый редизайн Yandex AI Studio
  'yandex.ru': [
    'textarea[data-testid="chat-input"]',
    'textarea[placeholder*="прос" i]',
    'div[contenteditable="true"][role="textbox"]',
    'div.ProseMirror[contenteditable="true"]',
    'textarea',
  ],
  // GigaChat
  'giga.chat': [
    'textarea[placeholder*="опросите" i]',
    'textarea[name="message"]',
    'textarea',
    'div[contenteditable="true"]',
  ],
  'developers.sber.ru': [
    'textarea[name="message"]',
    'textarea[placeholder]',
    'div[contenteditable="true"]',
  ],
  // DeepSeek
  'chat.deepseek.com': [
    '#chat-input',
    'textarea[placeholder*="message" i]',
    'textarea',
    'div[contenteditable="true"]',
  ],
  // Mistral Le Chat
  'chat.mistral.ai': [
    'div.ProseMirror[contenteditable="true"]',
    'textarea[name="message"]',
    'div[contenteditable="true"][role="textbox"]',
  ],
  'le-chat.mistral.ai': [
    'div.ProseMirror[contenteditable="true"]',
    'textarea[name="message"]',
    'div[contenteditable="true"][role="textbox"]',
  ],
  // Qwen
  'chat.qwen.ai': [
    'textarea[placeholder*="message" i]',
    'textarea#chat-input',
    'div[contenteditable="true"]',
  ],
}

/**
 * Находит input-элемент на странице по per-host селекторам, с fallback на
 * heuristic "largest contenteditable" (Agora 2026).
 *
 * B-17: при каждом fallback на heuristic пишем breadcrumb selector.miss
 * — мониторинг качества селекторов в проде. Если на каком-то хосте
 * фолбэк срабатывает часто — значит DOM сайта обновился и пора live recon.
 */
export function findTargetInput(host: string): HTMLElement | null {
  const list = INPUT_SELECTORS[host as SupportedHost] ?? []
  const tried: string[] = []
  for (const sel of list) {
    const el = document.querySelector<HTMLElement>(sel)
    if (el && isVisible(el)) {
      return el
    }
    tried.push(sel)
  }
  const fallback = findLargestContentEditable()
  addBreadcrumb('selector.miss', `input not found on ${host}`, {
    host,
    tried: tried.length,
    fallback_found: Boolean(fallback),
  }, fallback ? 'warning' : 'error')
  return fallback
}

function isVisible(el: HTMLElement): boolean {
  const rect = el.getBoundingClientRect()
  return rect.width > 0 && rect.height > 0
}

function findLargestContentEditable(): HTMLElement | null {
  let best: HTMLElement | null = null
  let bestArea = 0
  const nodes = document.querySelectorAll<HTMLElement>(
    '[contenteditable="true"], textarea',
  )
  for (const el of nodes) {
    if (!isVisible(el)) continue
    const rect = el.getBoundingClientRect()
    const area = rect.width * rect.height
    if (area > bestArea) {
      best = el
      bestArea = area
    }
  }
  return best
}
