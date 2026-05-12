// Lightweight Sentry/GlitchTip wrapper без @sentry/browser зависимости.
// Используется реальный envelope POST через sentry-envelope.ts,
// если задан WXT_SENTRY_DSN через env. Иначе — только локальный console.error.

import { generateEventID, sendEnvelope } from './sentry-envelope';

type Level = 'info' | 'warning' | 'error';

interface Breadcrumb {
  message: string;
  category: string;
  level: Level;
  ts: number;
  data?: Record<string, unknown>;
}

const MAX_BREADCRUMBS = 20;
const breadcrumbs: Breadcrumb[] = [];

let enabled = false;
let dsn = '';
let release = 'unknown';

export function initSentry(opts: { enabled: boolean; release?: string; dsn?: string }): void {
  enabled = opts.enabled;
  if (opts.release) release = opts.release;
  if (opts.dsn) dsn = opts.dsn;
  if (!enabled) {
    console.info('[sentry] init skipped (disabled)');
    return;
  }
  if (!dsn) {
    console.warn('[sentry] init: enabled but DSN missing — events will only print to console');
  } else {
    console.info('[sentry] init', { release, hasDSN: true });
  }
}

export function addBreadcrumb(
  category: string,
  message: string,
  data?: Record<string, unknown>,
  level: Level = 'info',
): void {
  breadcrumbs.push({ category, message, data, level, ts: Date.now() });
  if (breadcrumbs.length > MAX_BREADCRUMBS) breadcrumbs.shift();
}

export function captureException(err: unknown, context?: Record<string, unknown>): void {
  const payload = {
    event_id: generateEventID(),
    error: serializeError(err),
    context: scrubPII(context ?? {}),
    breadcrumbs: breadcrumbs.map((b) => ({ ...b, data: scrubPII(b.data ?? {}) })),
    release,
    ts: Date.now(),
  };

  if (!enabled) {
    console.error('[pv-error]', payload);
    return;
  }

  console.error('[sentry]', payload);
  if (!dsn) return;

  // sendEnvelope ожидает узкий тип error (string | {name,message,stack}).
  // serializeError может вернуть произвольный object — приводим к строке.
  const env = {
    ...payload,
    error:
      typeof payload.error === 'object' &&
      payload.error !== null &&
      'name' in payload.error &&
      'message' in payload.error
        ? (payload.error as { name: string; message: string; stack?: string })
        : typeof payload.error === 'string'
          ? payload.error
          : JSON.stringify(payload.error),
  };
  void sendEnvelope(dsn, env).then((result) => {
    if (!result.sent) {
      console.warn('[sentry] envelope not sent', result.reason);
    }
  });
}

/**
 * Сериализация unknown-error в читаемое представление.
 * Раньше `String(err)` давал «[object Object]» — теряли всю инфу о баге.
 */
function serializeError(err: unknown): unknown {
  if (err instanceof Error) {
    return { name: err.name, message: err.message, stack: err.stack };
  }
  if (typeof err === 'string') return err;
  if (err && typeof err === 'object') {
    try {
      // Сохраняем .message и .code если есть — типичные для нашего code: 'no_target'.
      const obj = err as Record<string, unknown>;
      const out: Record<string, unknown> = {};
      if (typeof obj.message === 'string') out.message = obj.message;
      if (typeof obj.code === 'string') out.code = obj.code;
      if (typeof obj.name === 'string') out.name = obj.name;
      // Остальное — best-effort JSON.
      try {
        const stringified = JSON.stringify(err);
        if (stringified !== '{}') out.raw = stringified;
      } catch {
        // ignore circular refs
      }
      return out;
    } catch {
      return String(err);
    }
  }
  return String(err);
}

/**
 * Удаляет потенциально чувствительные данные перед логированием.
 */
function scrubPII(data: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(data)) {
    const key = k.toLowerCase();
    if (key.includes('key') || key.includes('token') || key.includes('password') || key.includes('secret')) {
      out[k] = '[REDACTED]';
      continue;
    }
    if (key === 'content' || key === 'text') {
      // Не логируем содержимое промптов / значения переменных
      out[k] = typeof v === 'string' ? `[${v.length} chars]` : '[REDACTED]';
      continue;
    }
    out[k] = v;
  }
  return out;
}
