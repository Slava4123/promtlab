import { describe, expect, it } from 'vitest';
import { INPUT_SELECTORS, SUPPORTED_HOSTS } from '../lib/selectors';

describe('selectors', () => {
  it('covers all 4 target sites', () => {
    expect(SUPPORTED_HOSTS).toEqual([
      'chatgpt.com',
      'claude.ai',
      'gemini.google.com',
      'www.perplexity.ai',
    ]);
  });

  it('has at least 2 selectors per site (primary + fallback)', () => {
    for (const host of SUPPORTED_HOSTS) {
      expect(INPUT_SELECTORS[host].length).toBeGreaterThanOrEqual(2);
    }
  });

  it('perplexity primary selector is confirmed Lexical #ask-input', () => {
    expect(INPUT_SELECTORS['www.perplexity.ai'][0]).toBe('#ask-input');
  });

  it('chatgpt primary selector is #prompt-textarea', () => {
    expect(INPUT_SELECTORS['chatgpt.com'][0]).toBe('#prompt-textarea');
  });

  it('gemini primary selector targets rich-textarea .ql-editor (Quill)', () => {
    expect(INPUT_SELECTORS['gemini.google.com'][0]).toBe('rich-textarea .ql-editor');
  });
});
