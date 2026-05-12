// shadcn/ui helper для конкатенации класов (clsx + tailwind-merge).
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/**
 * Возвращает URL frontend веб-приложения для открытия страниц SPA
 * (sign-up, settings/integrations, pricing и т.д.).
 *
 * В prod backend и frontend на одном домене (promtlabs.ru) → используем apiBase.
 * В dev backend на :8080, frontend на :5173 → подставляем 5173.
 */
export function deriveFrontendUrl(apiBase: string): string {
  const trimmed = apiBase.replace(/\/$/, '');
  if (/^https?:\/\/localhost:8080(\/|$)/i.test(trimmed)) {
    return 'http://localhost:5173';
  }
  if (/^https?:\/\/127\.0\.0\.1:8080(\/|$)/i.test(trimmed)) {
    return 'http://127.0.0.1:5173';
  }
  return trimmed;
}

/**
 * Открывает страницу веб-приложения PromptVault (не backend API) в новой вкладке.
 */
export function openWebPage(apiBase: string, path: string): void {
  const frontend = deriveFrontendUrl(apiBase);
  const url = path.startsWith('/') ? `${frontend}${path}` : `${frontend}/${path}`;
  chrome.tabs.create({ url });
}
