import { describe, it, expect } from 'vitest';
import { getSettings } from '../lib/storage';

// Note: cache + invalidation тестируется только на уровне «повторный вызов
// не бьёт chrome.storage». Полноценный invalidation-test через setApiKey →
// notify требует resetModules между cases (модульный singleton живёт между
// тестами и листенер регистрируется в backend ровно один раз). Это не
// блокер для смока — реальная инвалидация уже верифицируется через
// background flow ручным smoke в Chrome.

describe('getSettings cache', () => {
  it('возвращает defaults для пустого хранилища', async () => {
    const s = await getSettings();
    expect(s.apiKey).toBeNull();
    expect(s.theme).toBe('system');
    expect(s.apiBase).toMatch(/^https?:\/\//);
  });

  it('повторные вызовы не бьют chrome.storage.local', async () => {
    const spy = chrome.storage.local.get as unknown as ReturnType<
      typeof import('vitest').vi.fn
    >;
    await getSettings();
    const before = spy.mock.calls.length;
    await getSettings();
    await getSettings();
    await getSettings();
    expect(spy.mock.calls.length).toBe(before);
  });
});
