// Минимальная отправка Sentry envelope (NDJSON POST) совместимая с GlitchTip.
// Без зависимости от @sentry/browser — работает в MV3 service worker без eval.
//
// Spec: https://develop.sentry.dev/sdk/data-model/envelopes/
//
// DSN format: https://<publicKey>@<host>/<projectId>
// Endpoint:   https://<host>/api/<projectId>/envelope/?sentry_key=<publicKey>

const SDK_NAME = 'promtlab.extension';
const SDK_VERSION = '0.1.0';

const RATE_LIMIT_MAX = 10; // events
const RATE_LIMIT_WINDOW_MS = 60_000; // per minute

let recentTimestamps: number[] = [];

interface ParsedDSN {
  publicKey: string;
  host: string;
  projectId: string;
  envelopeURL: string;
}

export function parseDSN(dsn: string): ParsedDSN | null {
  try {
    const u = new URL(dsn);
    if (!u.username) return null;
    const projectId = u.pathname.replace(/^\//, '');
    if (!projectId) return null;
    return {
      publicKey: u.username,
      host: u.host,
      projectId,
      envelopeURL: `${u.protocol}//${u.host}/api/${projectId}/envelope/?sentry_key=${u.username}`,
    };
  } catch {
    return null;
  }
}

export interface EnvelopePayload {
  event_id: string;
  release: string;
  error: { name: string; message: string; stack?: string } | string;
  breadcrumbs: ReadonlyArray<unknown>;
  context: Record<string, unknown>;
  ts: number;
}

function checkRateLimit(now: number): boolean {
  recentTimestamps = recentTimestamps.filter((t) => now - t < RATE_LIMIT_WINDOW_MS);
  if (recentTimestamps.length >= RATE_LIMIT_MAX) return false;
  recentTimestamps.push(now);
  return true;
}

// Только для тестов — сбросить состояние rate-limit между runs.
export function _resetRateLimitForTest(): void {
  recentTimestamps = [];
}

export interface SendOptions {
  fetchImpl?: typeof fetch; // для тестов
  now?: () => number;        // для тестов
}

export async function sendEnvelope(
  dsn: string,
  payload: EnvelopePayload,
  opts: SendOptions = {},
): Promise<{ sent: boolean; reason?: string }> {
  const parsed = parseDSN(dsn);
  if (!parsed) return { sent: false, reason: 'invalid_dsn' };

  const fetchImpl = opts.fetchImpl ?? globalThis.fetch;
  const now = (opts.now ?? Date.now)();

  if (!checkRateLimit(now)) {
    return { sent: false, reason: 'rate_limited' };
  }

  const sentAt = new Date(now).toISOString();

  const envelopeHeader = JSON.stringify({
    event_id: payload.event_id,
    sent_at: sentAt,
    sdk: { name: SDK_NAME, version: SDK_VERSION },
    dsn,
  });

  const itemHeader = JSON.stringify({ type: 'event' });

  const event = {
    event_id: payload.event_id,
    timestamp: now / 1000,
    platform: 'javascript',
    sdk: { name: SDK_NAME, version: SDK_VERSION },
    release: payload.release,
    environment: 'production',
    exception:
      typeof payload.error === 'string'
        ? { values: [{ type: 'Error', value: payload.error }] }
        : {
            values: [
              {
                type: payload.error.name,
                value: payload.error.message,
                stacktrace: payload.error.stack
                  ? { frames: parseStack(payload.error.stack) }
                  : undefined,
              },
            ],
          },
    breadcrumbs: { values: payload.breadcrumbs },
    extra: payload.context,
  };

  const itemPayload = JSON.stringify(event);
  const body = `${envelopeHeader}\n${itemHeader}\n${itemPayload}\n`;

  try {
    const ctl = new AbortController();
    const timeout = setTimeout(() => ctl.abort(), 5000);
    const resp = await fetchImpl(parsed.envelopeURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-sentry-envelope',
        'X-Sentry-Auth': `Sentry sentry_version=7, sentry_key=${parsed.publicKey}, sentry_client=${SDK_NAME}/${SDK_VERSION}`,
      },
      body,
      signal: ctl.signal,
    });
    clearTimeout(timeout);
    if (!resp.ok) {
      return { sent: false, reason: `http_${resp.status}` };
    }
    return { sent: true };
  } catch (err) {
    return { sent: false, reason: err instanceof Error ? err.message : 'fetch_failed' };
  }
}

// Очень простой парсер stack-trace: "at fn (file:line:col)" → frames.
// Лучшее усилие — Sentry/GlitchTip переварит даже сырую строку в exception.value.
function parseStack(stack: string): Array<{ function?: string; filename?: string; lineno?: number; colno?: number }> {
  const lines = stack.split('\n').slice(1, 21);
  const out: Array<{ function?: string; filename?: string; lineno?: number; colno?: number }> = [];
  for (const line of lines) {
    const m = line.match(/at\s+(?:(.+?)\s+\()?(.+?):(\d+):(\d+)\)?/);
    if (!m) continue;
    out.push({
      function: m[1] || undefined,
      filename: m[2],
      lineno: Number(m[3]),
      colno: Number(m[4]),
    });
  }
  return out;
}

export function generateEventID(): string {
  // 32-hex chars (без дефисов, как требует Sentry).
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}
