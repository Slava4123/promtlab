// Маленькие helpers для work с host labels — вынесены в отдельный файл чтобы
// избежать зависимости background.ts от messages.ts (которое тянет React-ориентированные типы).

import { HOST_LABELS } from './messages';

export const SUPPORTED_HOSTS_LIST: string[] = Object.keys(HOST_LABELS);

export function isSupportedHost(host: string | null): boolean {
  if (!host) return false;
  return host in HOST_LABELS;
}
