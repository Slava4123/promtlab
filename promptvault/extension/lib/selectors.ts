// Per-site селекторы input-полей. Атрибутные, re-query каждый раз
// (никогда не кэшировать references — Gemini Angular churn, ChatGPT A/B).
//
// Проверено live reconnaissance 2026-04-11 для Perplexity (Lexical).
// Остальные — на основе open-source расширений и статей 2025-2026.

export type SupportedHost =
  | 'chatgpt.com'
  | 'claude.ai'
  | 'gemini.google.com'
  | 'www.perplexity.ai';

export const SUPPORTED_HOSTS: SupportedHost[] = [
  'chatgpt.com',
  'claude.ai',
  'gemini.google.com',
  'www.perplexity.ai',
];

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
};

/**
 * Находит input-элемент на странице по per-host селекторам, с fallback на
 * heuristic "largest contenteditable" (Agora 2026).
 */
export function findTargetInput(host: string): HTMLElement | null {
  const list = INPUT_SELECTORS[host as SupportedHost] ?? [];
  for (const sel of list) {
    const el = document.querySelector<HTMLElement>(sel);
    if (el && isVisible(el)) return el;
  }
  return findLargestContentEditable();
}

function isVisible(el: HTMLElement): boolean {
  const rect = el.getBoundingClientRect();
  return rect.width > 0 && rect.height > 0;
}

function findLargestContentEditable(): HTMLElement | null {
  let best: HTMLElement | null = null;
  let bestArea = 0;
  const nodes = document.querySelectorAll<HTMLElement>('[contenteditable="true"]');
  for (const el of nodes) {
    if (!isVisible(el)) continue;
    const rect = el.getBoundingClientRect();
    const area = rect.width * rect.height;
    if (area > bestArea) {
      best = el;
      bestArea = area;
    }
  }
  return best;
}
