// Общий handler для всех content-script'ов.

import { findTargetInput, type SupportedHost } from './selectors';
import { insertPrompt } from './insert';
import type { ContentCommand, ContentResponse } from './messages';

// Хранит текст последней вставки — используется undo операцией
let lastInserted: { text: string; elementSelector: string | null } | null = null;

export function installContentHandler(host: SupportedHost): void {
  chrome.runtime.onMessage.addListener(
    (
      msg: ContentCommand,
      _sender,
      sendResponse: (response: ContentResponse) => void,
    ) => {
      if (msg?.type === 'content.ping') {
        sendResponse({ type: 'content.pong', host });
        return false;
      }

      if (msg?.type === 'content.insert') {
        const el = findTargetInput(host);
        if (!el) {
          sendResponse({ type: 'content.notFound' });
          return false;
        }
        const result = insertPrompt(el, msg.text);
        if (result.success) {
          lastInserted = {
            text: msg.text,
            elementSelector: el.id ? `#${el.id}` : null,
          };
          sendResponse({ type: 'content.inserted', strategy: result.strategy });
        } else {
          sendResponse({
            type: 'content.failed',
            reason: result.reason || 'unknown',
          });
        }
        return false;
      }

      if (msg?.type === 'content.undo') {
        const el = findTargetInput(host);
        if (!el || !lastInserted) {
          sendResponse({ type: 'content.notFound' });
          return false;
        }
        // Очищаем поле. Если user уже что-то дописал — очищаем только если
        // текущее содержимое ВСЁ ЕЩЁ содержит наш вставленный текст.
        const current =
          el instanceof HTMLTextAreaElement || el instanceof HTMLInputElement
            ? el.value
            : (el.innerText ?? el.textContent ?? '');
        if (!current.includes(lastInserted.text.slice(0, Math.min(40, lastInserted.text.length)))) {
          sendResponse({ type: 'content.failed', reason: 'edited' });
          return false;
        }
        // Заменяем на пустую строку через ту же cascade logic
        const clear = insertPrompt(el, '');
        if (clear.success) {
          lastInserted = null;
          sendResponse({ type: 'content.undone' });
        } else {
          sendResponse({ type: 'content.failed', reason: clear.reason || 'clear failed' });
        }
        return false;
      }

      return false;
    },
  );
}
