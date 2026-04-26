import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  _resetRateLimitForTest,
  generateEventID,
  parseDSN,
  sendEnvelope,
} from '../lib/sentry-envelope';

const VALID_DSN = 'https://abc123@glitchtip.example.com/42';

describe('parseDSN', () => {
  it('parses a valid DSN', () => {
    const p = parseDSN(VALID_DSN);
    expect(p).not.toBeNull();
    expect(p?.publicKey).toBe('abc123');
    expect(p?.host).toBe('glitchtip.example.com');
    expect(p?.projectId).toBe('42');
    expect(p?.envelopeURL).toBe(
      'https://glitchtip.example.com/api/42/envelope/?sentry_key=abc123',
    );
  });

  it('returns null for missing publicKey', () => {
    expect(parseDSN('https://glitchtip.example.com/42')).toBeNull();
  });

  it('returns null for missing projectId', () => {
    expect(parseDSN('https://abc@glitchtip.example.com/')).toBeNull();
  });

  it('returns null for malformed input', () => {
    expect(parseDSN('not a url')).toBeNull();
  });
});

describe('generateEventID', () => {
  it('returns 32-hex string', () => {
    const id = generateEventID();
    expect(id).toMatch(/^[0-9a-f]{32}$/);
  });

  it('is unique per call', () => {
    expect(generateEventID()).not.toBe(generateEventID());
  });
});

describe('sendEnvelope', () => {
  beforeEach(() => {
    _resetRateLimitForTest();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('POSTs envelope NDJSON with X-Sentry-Auth header', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(new Response('', { status: 200 }));
    const result = await sendEnvelope(
      VALID_DSN,
      {
        event_id: 'a'.repeat(32),
        release: '1.0.0',
        error: { name: 'TypeError', message: 'boom', stack: 'TypeError: boom\n    at x (file.js:1:1)' },
        breadcrumbs: [],
        context: { foo: 'bar' },
        ts: 1_700_000_000_000,
      },
      { fetchImpl, now: () => 1_700_000_000_000 },
    );

    expect(result.sent).toBe(true);
    expect(fetchImpl).toHaveBeenCalledOnce();
    const [url, init] = fetchImpl.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('https://glitchtip.example.com/api/42/envelope/?sentry_key=abc123');
    expect((init.headers as Record<string, string>)['X-Sentry-Auth']).toContain(
      'sentry_key=abc123',
    );
    expect((init.headers as Record<string, string>)['Content-Type']).toBe(
      'application/x-sentry-envelope',
    );

    const body = init.body as string;
    const lines = body.trim().split('\n');
    expect(lines).toHaveLength(3);
    const envHeader = JSON.parse(lines[0]);
    expect(envHeader.event_id).toBe('a'.repeat(32));
    expect(envHeader.dsn).toBe(VALID_DSN);
    expect(JSON.parse(lines[1])).toEqual({ type: 'event' });
    const event = JSON.parse(lines[2]);
    expect(event.platform).toBe('javascript');
    expect(event.release).toBe('1.0.0');
    expect(event.exception.values[0].type).toBe('TypeError');
    expect(event.extra.foo).toBe('bar');
  });

  it('returns invalid_dsn for malformed DSN', async () => {
    const fetchImpl = vi.fn();
    const result = await sendEnvelope(
      'broken',
      {
        event_id: 'x',
        release: '1',
        error: 'boom',
        breadcrumbs: [],
        context: {},
        ts: 0,
      },
      { fetchImpl },
    );
    expect(result.sent).toBe(false);
    expect(result.reason).toBe('invalid_dsn');
    expect(fetchImpl).not.toHaveBeenCalled();
  });

  it('drops events past rate limit (10/min)', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(new Response('', { status: 200 }));
    const payload = {
      event_id: 'a'.repeat(32),
      release: '1',
      error: 'boom',
      breadcrumbs: [],
      context: {},
      ts: 0,
    };

    let now = 1_700_000_000_000;
    for (let i = 0; i < 10; i++) {
      const res = await sendEnvelope(VALID_DSN, payload, {
        fetchImpl,
        now: () => now,
      });
      expect(res.sent).toBe(true);
    }

    // 11-я попытка в том же окне → rate-limited
    const dropped = await sendEnvelope(VALID_DSN, payload, {
      fetchImpl,
      now: () => now,
    });
    expect(dropped.sent).toBe(false);
    expect(dropped.reason).toBe('rate_limited');
    expect(fetchImpl).toHaveBeenCalledTimes(10);

    // Через минуту окно сбрасывается
    now += 60_001;
    const after = await sendEnvelope(VALID_DSN, payload, {
      fetchImpl,
      now: () => now,
    });
    expect(after.sent).toBe(true);
  });

  it('returns reason when fetch throws', async () => {
    const fetchImpl = vi.fn().mockRejectedValue(new Error('network down'));
    const result = await sendEnvelope(
      VALID_DSN,
      {
        event_id: 'a'.repeat(32),
        release: '1',
        error: 'boom',
        breadcrumbs: [],
        context: {},
        ts: 0,
      },
      { fetchImpl },
    );
    expect(result.sent).toBe(false);
    expect(result.reason).toBe('network down');
  });

  it('returns http_NNN reason on non-2xx', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(new Response('', { status: 429 }));
    const result = await sendEnvelope(
      VALID_DSN,
      {
        event_id: 'a'.repeat(32),
        release: '1',
        error: 'boom',
        breadcrumbs: [],
        context: {},
        ts: 0,
      },
      { fetchImpl },
    );
    expect(result.sent).toBe(false);
    expect(result.reason).toBe('http_429');
  });
});
