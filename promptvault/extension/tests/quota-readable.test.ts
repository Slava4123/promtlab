// readableQuotaType — мапит quota_type из backend в русский label.
// Brittle hot-path: backend сейчас отдаёт 'unknown' (bg-client), поэтому
// fallback угадывает по тексту backend-сообщения. Тест охраняет от
// регрессий при смене копирайта или добавлении новых типов ресурсов.

import { describe, expect, it, vi, beforeEach } from 'vitest';

import { readableQuotaType } from '../components/subscription/quota-exceeded-dialog';

beforeEach(() => {
  vi.spyOn(console, 'warn').mockImplementation(() => {});
});

describe('readableQuotaType — техкей-маппинг', () => {
  it.each([
    ['prompts', 'Промпты'],
    ['collections', 'Коллекции'],
    ['chains', 'Цепочки'],
    ['teams', 'Команды'],
    ['ext_uses_today', 'Вставки сегодня'],
    ['mcp_uses_today', 'MCP-вызовы сегодня'],
    ['api_keys', 'API-ключи'],
    ['share_links', 'Публичные ссылки'],
    ['team_members', 'Участники команды'],
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
    ['API-ключи: лимит достигнут', 'API-ключи'],
    ['Создание api ключа невозможно', 'API-ключи'],
  ])('"%s" → %s', (message, expected) => {
    expect(readableQuotaType('unknown', message)).toBe(expected);
  });

  it('case-insensitive', () => {
    expect(readableQuotaType('unknown', 'ЛИМИТ ЦЕПОЧЕК')).toBe('Цепочки');
  });
});

describe('readableQuotaType — fallback и логирование', () => {
  it('возвращает fallback при null+null', () => {
    expect(readableQuotaType(null, null)).toBe('Лимит ресурса');
  });

  it('не флудит warn если оба аргумента пустые', () => {
    const warn = vi.spyOn(console, 'warn');
    readableQuotaType(null, null);
    expect(warn).not.toHaveBeenCalled();
  });

  it('пишет warn если quotaType не пуст, но не распознан', () => {
    const warn = vi.spyOn(console, 'warn');
    readableQuotaType('mystery_resource', null);
    expect(warn).toHaveBeenCalledTimes(1);
    expect(warn).toHaveBeenCalledWith(
      '[QuotaDialog] не распознан тип квоты',
      expect.objectContaining({ quotaType: 'mystery_resource' }),
    );
  });

  it('пишет warn если message не пуст, но не угадан', () => {
    const warn = vi.spyOn(console, 'warn');
    readableQuotaType('unknown', 'Какое-то новое сообщение без знакомых слов');
    expect(warn).toHaveBeenCalledTimes(1);
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
