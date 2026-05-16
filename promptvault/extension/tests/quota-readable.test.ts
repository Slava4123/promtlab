// readableQuotaType — мапит quota_type из backend в русский label.
// Brittle hot-path: backend может слать новые quota_type'ы или менять copy,
// поэтому fallback угадывает по тексту backend-сообщения, а в крайнем
// случае добавляет Sentry breadcrumb для SRE-видимости. Тест охраняет от
// регрессий при смене копирайта или добавлении новых типов ресурсов.

import { describe, expect, it, vi, beforeEach } from 'vitest';

// Mock'аем sentry-модуль чтобы проверить, что addBreadcrumb вызывается
// для unknown quota types (раньше проверяли console.warn).
vi.mock('../lib/sentry', () => ({
  addBreadcrumb: vi.fn(),
}));

import { readableQuotaType } from '../lib/quota-labels';
import { addBreadcrumb } from '../lib/sentry';

beforeEach(() => {
  vi.mocked(addBreadcrumb).mockClear();
});

describe('readableQuotaType — техкей-маппинг', () => {
  it.each([
    // Источник истины — backend usecases/quota/quota.go::newQuotaExceeded +
    // team/branding_handler.go ("branding") + analytics/errors.go.
    ['prompts', 'Промпты'],
    ['collections', 'Коллекции'],
    ['chains', 'Цепочки'],
    ['teams', 'Команды'],
    ['ext_daily', 'Вставки сегодня'],
    ['mcp_daily', 'MCP-вызовы сегодня'],
    ['team_prompts', 'Промпты команды'],
    ['team_collections', 'Коллекции команды'],
    ['team_chains', 'Цепочки команды'],
    ['team_members', 'Участники команды'],
    ['chain_steps', 'Шаги в цепочке'],
    ['branding', 'Брендинг команды'],
    ['insights', 'Smart Insights (Max)'],
    ['export', 'Экспорт CSV (Pro)'],
    // Алиасы из UsageSummary endpoint
    ['ext_uses_today', 'Вставки сегодня'],
    ['mcp_uses_today', 'MCP-вызовы сегодня'],
  ])('маппит %s → %s', (key, expected) => {
    expect(readableQuotaType(key, null)).toBe(expected);
  });

  it('игнорирует "unknown" и идёт по fallback', () => {
    expect(readableQuotaType('unknown', null)).toBe('Лимит ресурса');
  });

  it('игнорирует неизвестный quotaType и идёт по fallback', () => {
    expect(readableQuotaType('totally_new_resource', null)).toBe('Лимит ресурса');
  });
});

describe('readableQuotaType — fuzzy-match по сообщению', () => {
  it.each([
    ['Лимит цепочек исчерпан', 'Цепочки'],
    ['Цепочка не может быть создана', 'Цепочки'],
    ['Превышен лимит промптов', 'Промпты'],
    ['промпт нельзя создать', 'Промпты'],
    ['Достигнут предел коллекций', 'Коллекции'],
    ['Лимит команд достигнут', 'Команды'],
    ['Превышено число вставок сегодня', 'Вставки сегодня'],
    ['Лимит использований API', 'Вставки сегодня'],
    ['Лимит MCP-вызовов', 'MCP-вызовы сегодня'],
    ['Smart Insights требует Max', 'Smart Insights (Max)'],
    ['Экспорт CSV — только Pro', 'Экспорт CSV (Pro)'],
    ['Загрузка логотипа недоступна', 'Брендинг команды'],
  ])('"%s" → %s', (message, expected) => {
    expect(readableQuotaType('unknown', message)).toBe(expected);
  });

  it('case-insensitive', () => {
    expect(readableQuotaType('unknown', 'ЛИМИТ ЦЕПОЧЕК')).toBe('Цепочки');
  });
});

describe('readableQuotaType — fallback и Sentry breadcrumb', () => {
  it('возвращает fallback при null+null', () => {
    expect(readableQuotaType(null, null)).toBe('Лимит ресурса');
  });

  it('не флудит breadcrumb если оба аргумента пустые', () => {
    readableQuotaType(null, null);
    expect(addBreadcrumb).not.toHaveBeenCalled();
  });

  it('пишет breadcrumb если quotaType не пуст, но не распознан', () => {
    readableQuotaType('mystery_resource', null);
    expect(addBreadcrumb).toHaveBeenCalledTimes(1);
    expect(addBreadcrumb).toHaveBeenCalledWith(
      'quota.unknown_type',
      'fallback used',
      expect.objectContaining({ quotaType: 'mystery_resource' }),
      'warning',
    );
  });

  it('пишет breadcrumb если message не пуст, но не угадан', () => {
    readableQuotaType('unknown', 'Какое-то новое сообщение без знакомых слов');
    expect(addBreadcrumb).toHaveBeenCalledTimes(1);
  });

  it('тех-кей побеждает невнятный message (не уходит в fallback)', () => {
    expect(readableQuotaType('chains', 'произвольный текст')).toBe('Цепочки');
  });

  it('возвращает строку, не null (тип-гарантия)', () => {
    const result = readableQuotaType(null, 'абракадабра');
    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
  });
});
