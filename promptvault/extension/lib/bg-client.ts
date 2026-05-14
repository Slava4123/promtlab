// Type-safe обёртка над chrome.runtime.sendMessage, на стороне Side Panel.
// Преобразует BgResponse в Promise<T> или throw ApiError.

import { ApiError } from './types';
import type { BgRequest, BgResponse, BgResultMap } from './messages';
import { useQuotaStore } from '../stores/quota-store';

function asString(v: unknown): string | undefined {
  return typeof v === 'string' ? v : undefined;
}
function asNumber(v: unknown): number | undefined {
  return typeof v === 'number' ? v : undefined;
}

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
        // Backend кладёт quota_type/used/limit/plan в body 402-ответа.
        // lib/api.ts::request пробрасывает их через ApiError.details →
        // background::toErrorResponse → response.details. Без этого
        // диалог не знает какой ресурс исчерпан и показывает generic fallback.
        const d = response.details ?? {};
        useQuotaStore.getState().show({
          quotaType: asString(d.quota_type) ?? 'unknown',
          message: response.message ?? asString(d.error) ?? 'Лимит исчерпан',
          used: asNumber(d.used),
          limit: asNumber(d.limit),
          plan: asString(d.plan),
        });
      } catch {
        // Store недоступен (SSR/no-DOM) — ignore
      }
    }
    throw new ApiError(response.message ?? response.error, status, response.error, response.details);
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
