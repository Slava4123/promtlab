// Type-safe обёртка над chrome.runtime.sendMessage, на стороне Side Panel.
// Преобразует BgResponse в Promise<T> или throw ApiError.

import { ApiError } from './types';
import type { BgError, BgRequest, BgResponse, BgResultMap } from './messages';
import { useQuotaStore } from '../stores/quota-store';
import { addBreadcrumb, captureException } from './sentry';

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
    addBreadcrumb('bg.empty_response', msg.type, undefined, 'warning');
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
      } catch (e) {
        // Side Panel — DOM-context, поэтому реальные причины здесь
        // нетривиальны (Zustand bug, exception в .show, потеря reactivity).
        // Раньше silent ignore → юзер при 402 не видел диалог, ApiError
        // всё равно throw'ался, но UX-сигнал терялся.
        addBreadcrumb('bg.quota.show_failed', String(e), undefined, 'warning');
        captureException(e, { context: 'quota_dialog', msgType: msg.type });
      }
    }
    throw new ApiError(response.message ?? response.error, status, response.error, response.details);
  }
  return response.data;
}

// Типизированная мапа: добавление нового кода в BgError → TS-ошибка тут,
// пока не пропишешь HTTP-эквивалент. Раньше был switch с default:500 —
// пропущенный case проходил незамеченным.
const BG_ERROR_STATUS: Record<BgError, number> = {
  unauthorized: 401,
  forbidden: 403,
  not_found: 404,
  conflict: 409,
  validation: 422,
  quota_exceeded: 402,
  payload_too_large: 413,
  unsupported_media_type: 415,
  rate_limited: 429,
  client_error: 400,
  network: 0,
  unknown: 500,
  no_target: 0,
  no_history: 0,
};

function errorToStatus(code: BgError): number {
  return BG_ERROR_STATUS[code];
}
