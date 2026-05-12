// Type-safe обёртка над chrome.runtime.sendMessage, на стороне Side Panel.
// Преобразует BgResponse в Promise<T> или throw ApiError.

import { ApiError } from './types';
import type { BgRequest, BgResponse, BgResultMap } from './messages';
import { useQuotaStore } from '../stores/quota-store';

export async function sendBg<K extends BgRequest['type']>(
  msg: Extract<BgRequest, { type: K }>,
): Promise<BgResultMap[K]> {
  const response = (await chrome.runtime.sendMessage(msg)) as BgResponse<BgResultMap[K]>;
  if (!response) {
    throw new ApiError('empty response from background', 0, 'network');
  }
  if (!response.ok) {
    const status = errorToStatus(response.error);
    // Глобальный quota dialog при 402 — показываем модалку, потом throw'им
    // дальше чтобы UI мог обработать локально.
    if (response.error === 'quota_exceeded') {
      try {
        useQuotaStore.getState().show({
          quotaType: 'unknown',
          message: response.message ?? 'Лимит исчерпан',
        });
      } catch {
        // Store недоступен (SSR/no-DOM) — ignore
      }
    }
    throw new ApiError(response.message ?? response.error, status, response.error);
  }
  return response.data;
}

function errorToStatus(code: string): number {
  switch (code) {
    case 'unauthorized':
      return 401;
    case 'forbidden':
      return 403;
    case 'not_found':
      return 404;
    case 'conflict':
      return 409;
    case 'validation':
      return 422;
    case 'quota_exceeded':
      return 402;
    case 'rate_limited':
      return 429;
    case 'network':
      return 0;
    case 'no_target':
      return 0;
    case 'no_history':
      return 0;
    default:
      return 500;
  }
}
