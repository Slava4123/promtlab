import { describe, it, expect } from 'vitest';
import { extractVariables, hasVariables, renderTemplate } from '@pv/shared/template';

describe('extractVariables', () => {
  it('returns empty for plain text', () => {
    expect(extractVariables('hello world')).toEqual([]);
  });

  it('extracts single variable', () => {
    expect(extractVariables('Hi {{name}}')).toEqual(['name']);
  });

  it('de-duplicates preserving first-occurrence order', () => {
    expect(extractVariables('{{b}} {{a}} {{b}} {{c}} {{a}}')).toEqual(['b', 'a', 'c']);
  });

  it('supports Cyrillic identifiers (Unicode \\p{L})', () => {
    expect(extractVariables('Привет {{имя}}, твой возраст {{возраст}}.')).toEqual(['имя', 'возраст']);
  });

  it('rejects identifiers starting with digit', () => {
    expect(extractVariables('{{1var}}')).toEqual([]);
  });
});

describe('hasVariables', () => {
  it('true when content has at least one variable', () => {
    expect(hasVariables('say {{hi}}')).toBe(true);
  });

  it('false for plain text', () => {
    expect(hasVariables('no variables here')).toBe(false);
  });

  it('false for malformed braces', () => {
    expect(hasVariables('{single} or {{ }}')).toBe(false);
  });
});

describe('renderTemplate', () => {
  it('substitutes provided values', () => {
    expect(renderTemplate('Hi {{name}}', { name: 'Slava' })).toBe('Hi Slava');
  });

  it('missing key becomes empty string, NOT literal placeholder', () => {
    expect(renderTemplate('{{a}}/{{b}}', { a: '1' })).toBe('1/');
  });

  it('does NOT re-scan substituted values for {{...}} (single pass)', () => {
    expect(renderTemplate('{{x}}', { x: '{{y}}', y: 'recursed' })).toBe('{{y}}');
  });

  it('regex metacharacters in values are treated as literals', () => {
    expect(renderTemplate('{{re}}', { re: '$1 \\d+' })).toBe('$1 \\d+');
  });
});
