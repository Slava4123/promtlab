import { describe, it, expect } from 'vitest';
import { QUOTA_KEYS, quotaByKey, type UsageSummary } from '@pv/shared/types';

const summary: UsageSummary = {
  plan_id: 'pro',
  prompts: { used: 10, limit: 100 },
  collections: { used: 2, limit: 50 },
  teams: { used: 1, limit: 5 },
  chains: { used: 0, limit: 5 },
  ext_uses_today: { used: 42, limit: 200 },
  mcp_uses_today: { used: 0, limit: 100 },
};

describe('quotaByKey', () => {
  it('returns QuotaInfo for every key in QUOTA_KEYS', () => {
    for (const key of QUOTA_KEYS) {
      const info = quotaByKey(summary, key);
      expect(info).toHaveProperty('used');
      expect(info).toHaveProperty('limit');
    }
  });

  it('preserves exact reference (no copy)', () => {
    expect(quotaByKey(summary, 'prompts')).toBe(summary.prompts);
  });

  it('all known keys covered (no typos vs UsageSummary fields)', () => {
    // Защита от drift'а: если в UsageSummary добавили новое поле, тут забыли —
    // снимок ключей фиксирован, и расхождение видно сразу.
    expect([...QUOTA_KEYS].sort()).toEqual(
      ['chains', 'collections', 'ext_uses_today', 'mcp_uses_today', 'prompts', 'teams'].sort(),
    );
  });
});
