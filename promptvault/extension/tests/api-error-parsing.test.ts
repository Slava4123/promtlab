// Regression-тесты для request() из lib/api.ts — единый разбор тела ошибок
// для всех 4xx. Backend пишет либо {"message":"X"} (delivery/http/utils),
// либо {"error":"X"} (delivery/http/errors). Раньше расширение игнорировало
// тело и показывало "http 400" — фикс f54c69f покрывается этим тестом.

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getMe } from '../lib/api';
import { setApiBase, setApiKey } from '../lib/storage';
import { ApiError } from '../lib/types';

const API_BASE = 'https://example.test';

function jsonRes(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function textRes(status: number, text: string): Response {
  return new Response(text, {
    status,
    headers: { 'Content-Type': 'text/plain' },
  });
}

beforeEach(async () => {
  await setApiKey('test-key');
  await setApiBase(API_BASE);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('request() — body parsing for 4xx', () => {
  it('берёт message из тела для 401', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonRes(401, { message: 'Сессия истекла' })));
    await expect(getMe()).rejects.toMatchObject({
      status: 401,
      code: 'unauthorized',
      message: 'Сессия истекла',
    });
  });

  it('берёт error из тела когда message отсутствует', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonRes(403, { error: 'forbidden by policy' })));
    await expect(getMe()).rejects.toMatchObject({
      status: 403,
      code: 'forbidden',
      message: 'forbidden by policy',
    });
  });

  it('предпочитает message над error', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonRes(409, { message: 'conflict-msg', error: 'conflict-err' })),
    );
    await expect(getMe()).rejects.toMatchObject({
      status: 409,
      message: 'conflict-msg',
    });
  });

  it.each([
    [401, 'unauthorized'],
    [402, 'quota_exceeded'],
    [403, 'forbidden'],
    [404, 'not_found'],
    [409, 'conflict'],
    [422, 'validation'],
    [429, 'rate_limited'],
  ] as const)('подставляет fallback при пустом теле для %i → code=%s', async (status, expectedCode) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(textRes(status, '')));
    try {
      await getMe();
      expect.fail('should have thrown ApiError');
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(status);
      expect((err as ApiError).code).toBe(expectedCode);
      expect((err as ApiError).message).toBeTruthy(); // дефолтное «unauthorized»/«validation»/...
    }
  });

  it('catch-all для прочих 4xx возвращает status + msg + code=client_error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonRes(418, { message: "I'm a teapot" })));
    await expect(getMe()).rejects.toMatchObject({
      status: 418,
      code: 'client_error',
      message: "I'm a teapot",
    });
  });

  it('catch-all для 4xx без тела даёт generic http NNN + code=client_error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(textRes(418, '')));
    await expect(getMe()).rejects.toMatchObject({
      status: 418,
      code: 'client_error',
      message: 'http 418',
    });
  });

  it('5xx без тела → ApiError code=network', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(textRes(503, '')));
    await expect(getMe()).rejects.toMatchObject({
      status: 503,
      code: 'network',
    });
  });

  it('network throw → ApiError code=network status=0', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    await expect(getMe()).rejects.toMatchObject({
      status: 0,
      code: 'network',
    });
  });

  it('200 с невалидным JSON даёт ApiError code=network', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response('not-json', { status: 200, headers: { 'Content-Type': 'application/json' } }),
      ),
    );
    await expect(getMe()).rejects.toMatchObject({
      status: 500,
      code: 'network',
      message: 'invalid json response',
    });
  });
});

