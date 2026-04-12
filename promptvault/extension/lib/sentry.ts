// Lightweight Sentry wrapper без @sentry/browser зависимости.
// Для production можно заменить на полноценный Sentry SDK — интерфейс останется.
//
// Текущий подход: структурированный логгер + optional POST в GlitchTip endpoint,
// только если задан VITE_SENTRY_DSN через env.

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
let release = 'unknown';

export function initSentry(opts: { enabled: boolean; release?: string }): void {
  enabled = opts.enabled;
  if (opts.release) release = opts.release;
  if (!enabled) {
    console.info('[sentry] init skipped (disabled)');
    return;
  }
  console.info('[sentry] init', { release });
}

export function addBreadcrumb(category: string, message: string, data?: Record<string, unknown>, level: Level = 'info'): void {
  breadcrumbs.push({ category, message, data, level, ts: Date.now() });
  if (breadcrumbs.length > MAX_BREADCRUMBS) breadcrumbs.shift();
}

export function captureException(err: unknown, context?: Record<string, unknown>): void {
  const payload = {
    error: err instanceof Error ? { name: err.name, message: err.message, stack: err.stack } : String(err),
    context: scrubPII(context ?? {}),
    breadcrumbs: breadcrumbs.map((b) => ({ ...b, data: scrubPII(b.data ?? {}) })),
    release,
    ts: Date.now(),
  };
  if (enabled) {
    console.error('[sentry]', payload);
    // TODO: реальная отправка в GlitchTip через fetch на DSN endpoint.
    // Для MVP — только структурированный лог; эти события видны в chrome://extensions service worker console.
  } else {
    console.error('[pv-error]', payload);
  }
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
