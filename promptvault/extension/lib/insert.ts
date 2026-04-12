// Cascade insertion strategy — один файл для всех target-сайтов.
// Подтверждено live reconnaissance 2026-04-11:
//   Claude (ProseMirror)    → execCommand('insertText') ✅
//   Gemini (Quill)          → execCommand('insertText') ✅
//   Perplexity (Lexical)    → execCommand('insertText') ✅
//   ChatGPT (textarea)      → nativeInputValueSetter ✅
//   ChatGPT (ProseMirror)   → execCommand('insertText') ✅

import type { InsertStrategy } from './messages';

export interface InsertResult {
  success: boolean;
  strategy: InsertStrategy;
  reason?: string;
}

export function insertPrompt(el: HTMLElement, text: string): InsertResult {
  try {
    el.focus();
  } catch {
    return { success: false, strategy: 'fallback', reason: 'cannot focus' };
  }

  // Strategy 1: native input/textarea — React-aware setter
  if (el instanceof HTMLTextAreaElement || el instanceof HTMLInputElement) {
    return insertIntoNativeInput(el, text);
  }

  // Strategy 2/3/4: contenteditable cascade
  if (el.isContentEditable) {
    return insertIntoContentEditable(el, text);
  }

  return { success: false, strategy: 'fallback', reason: 'unsupported element' };
}

function insertIntoNativeInput(
  el: HTMLTextAreaElement | HTMLInputElement,
  text: string,
): InsertResult {
  const proto =
    el instanceof HTMLTextAreaElement
      ? window.HTMLTextAreaElement.prototype
      : window.HTMLInputElement.prototype;

  const descriptor = Object.getOwnPropertyDescriptor(proto, 'value');
  const setter = descriptor?.set;
  if (!setter) {
    return { success: false, strategy: 'nativeSetter', reason: 'no setter' };
  }

  setter.call(el, text);
  el.dispatchEvent(new Event('input', { bubbles: true, composed: true }));
  el.dispatchEvent(new Event('change', { bubbles: true }));

  if (el.value === text) {
    return { success: true, strategy: 'nativeSetter' };
  }
  return { success: false, strategy: 'nativeSetter', reason: 'value mismatch' };
}

function insertIntoContentEditable(el: HTMLElement, text: string): InsertResult {
  // Clear existing content
  try {
    const sel = window.getSelection();
    if (sel) {
      sel.selectAllChildren(el);
      document.execCommand('delete');
    }
  } catch {
    // best effort
  }

  // Primary: execCommand('insertText')
  try {
    const ok = document.execCommand('insertText', false, text);
    if (ok && verifyContains(el, text)) {
      return { success: true, strategy: 'execCommand' };
    }
  } catch {
    // fallthrough
  }

  // Fallback: simulated paste (works on Lexical, ProseMirror, Quill via their paste handlers)
  try {
    const dt = new DataTransfer();
    dt.setData('text/plain', text);
    const evt = new ClipboardEvent('paste', {
      clipboardData: dt,
      bubbles: true,
      cancelable: true,
    });
    el.dispatchEvent(evt);
    if (verifyContains(el, text)) {
      return { success: true, strategy: 'paste' };
    }
  } catch {
    // fallthrough
  }

  return { success: false, strategy: 'fallback', reason: 'all strategies failed' };
}

function verifyContains(el: HTMLElement, text: string): boolean {
  const probe = text.slice(0, Math.min(40, text.length)).trim();
  if (!probe) return true;
  const content = (el.innerText ?? el.textContent ?? '').trim();
  return content.includes(probe);
}
