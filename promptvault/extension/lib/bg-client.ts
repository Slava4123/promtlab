// Type-safe обёртка над chrome.runtime.sendMessage, на стороне Side Panel.
// Преобразует BgResponse в Promise<T> или throw ApiError.

import { ApiError } from './types';
import type { BgRequest, BgResponse, BgResultMap } from './messages';

export async function sendBg<K extends BgRequest['type']>(
  msg: Extract<BgRequest, { type: K }>,
): Promise<BgResultMap[K]> {
  const response = (await chrome.runtime.sendMessage(msg)) as BgResponse<BgResultMap[K]>;
  if (!response) {
    throw new ApiError('empty response from background', 0, 'network');
  }
  if (!response.ok) {
    const status = errorToStatus(response.error);
    throw new ApiError(response.message ?? response.error, status, response.error);
  }
  return response.data;
}

function errorToStatus(code: string): number {
  switch (code) {
    case 'unauthorized':
      return 401;
    case 'not_found':
      return 404;
    case 'rate_limited':
      return 429;
    case 'network':
      return 0;
    case 'no_target':
      return 0;
    default:
      return 500;
  }
}
