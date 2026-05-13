import { describe, expect, it } from 'vitest';
import { INPUT_SELECTORS, SUPPORTED_HOSTS } from '../lib/selectors';

// AI-сайты, которые extension поддерживает — стабильный список, любое
// добавление/удаление должно требовать обновления теста + manifest host_permissions.
const EXPECTED_HOSTS = [
  'chatgpt.com',
  'claude.ai',
  'gemini.google.com',
  'www.perplexity.ai',
  'alice.yandex.ru',
  'ya.ru',
  'yandex.ru',
  'giga.chat',
  'developers.sber.ru',
  'chat.deepseek.com',
  'chat.mistral.ai',
  'le-chat.mistral.ai',
  'chat.qwen.ai',
] as const;

describe('selectors', () => {
  it('covers all target AI sites in stable order', () => {
    expect(SUPPORTED_HOSTS).toEqual(EXPECTED_HOSTS);
  });

  it('has at least 2 selectors per site (primary + fallback)', () => {
    for (const host of SUPPORTED_HOSTS) {
      expect(INPUT_SELECTORS[host].length).toBeGreaterThanOrEqual(2);
    }
  });

  it('chatgpt primary selector is #prompt-textarea', () => {
    expect(INPUT_SELECTORS['chatgpt.com'][0]).toBe('#prompt-textarea');
  });

  it('perplexity primary selector is Lexical #ask-input', () => {
    expect(INPUT_SELECTORS['www.perplexity.ai'][0]).toBe('#ask-input');
  });

  it('gemini primary selector targets rich-textarea .ql-editor (Quill)', () => {
    expect(INPUT_SELECTORS['gemini.google.com'][0]).toBe('rich-textarea .ql-editor');
  });

  it('claude primary selector targets ProseMirror contenteditable', () => {
    expect(INPUT_SELECTORS['claude.ai'][0]).toBe('div[contenteditable="true"].ProseMirror');
  });

  it('every entry in SUPPORTED_HOSTS has matching INPUT_SELECTORS', () => {
    for (const host of SUPPORTED_HOSTS) {
      expect(INPUT_SELECTORS[host]).toBeDefined();
      expect(Array.isArray(INPUT_SELECTORS[host])).toBe(true);
    }
  });
});
