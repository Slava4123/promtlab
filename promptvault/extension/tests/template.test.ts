import { describe, expect, it } from 'vitest';
import { extractVariables, renderTemplate } from '../lib/template';

describe('extractVariables', () => {
  it('extracts simple ASCII variables', () => {
    expect(extractVariables('Hello {{name}}, you are {{age}}')).toEqual(['name', 'age']);
  });

  it('deduplicates repeated variables', () => {
    expect(extractVariables('{{a}} {{b}} {{a}} {{a}} {{b}}')).toEqual(['a', 'b']);
  });

  it('preserves first-occurrence order', () => {
    expect(extractVariables('{{x}} {{y}} {{x}} {{z}} {{y}}')).toEqual(['x', 'y', 'z']);
  });

  it('supports Cyrillic identifiers', () => {
    expect(extractVariables('Привет {{имя}}, тебе {{возраст}} лет')).toEqual([
      'имя',
      'возраст',
    ]);
  });

  it('supports mixed scripts', () => {
    expect(extractVariables('{{name}} vs {{имя}}')).toEqual(['name', 'имя']);
  });

  it('supports underscores and digits (not first char)', () => {
    expect(extractVariables('{{user_name}} {{Var2}} {{_private}}')).toEqual([
      'user_name',
      'Var2',
      '_private',
    ]);
  });

  it('ignores placeholders with spaces', () => {
    expect(extractVariables('{{ name }} {{  a  }}')).toEqual([]);
  });

  it('ignores placeholders starting with digit', () => {
    expect(extractVariables('{{1name}} {{2var}}')).toEqual([]);
  });

  it('returns empty for empty input', () => {
    expect(extractVariables('')).toEqual([]);
  });

  it('returns empty when no placeholders', () => {
    expect(extractVariables('Просто текст без placeholders')).toEqual([]);
  });
});

describe('renderTemplate', () => {
  it('substitutes basic values', () => {
    expect(
      renderTemplate('Hello {{name}}, you are {{age}}', { name: 'John', age: '30' }),
    ).toBe('Hello John, you are 30');
  });

  it('empty string for missing values', () => {
    expect(
      renderTemplate('Hello {{name}}, you are {{age}}', { age: '30' }),
    ).toBe('Hello , you are 30');
  });

  it('handles Cyrillic values', () => {
    expect(
      renderTemplate('Привет {{имя}}, тебе {{возраст}} лет', {
        имя: 'Иван',
        возраст: '25',
      }),
    ).toBe('Привет Иван, тебе 25 лет');
  });

  it('escapes regex metacharacters in values', () => {
    expect(renderTemplate('Price: {{val}}', { val: '$100 & 50% off' })).toBe(
      'Price: $100 & 50% off',
    );
  });

  it('does not interpret $1 as back-reference', () => {
    expect(renderTemplate('Got {{val}}', { val: '$1 prize' })).toBe('Got $1 prize');
  });

  it('repeats same variable multiple times', () => {
    expect(
      renderTemplate('hi {{name}}, bye {{name}}, hey {{name}}', { name: 'John' }),
    ).toBe('hi John, bye John, hey John');
  });

  it('preserves newlines', () => {
    expect(renderTemplate('line1\n{{name}}\nline3', { name: 'John' })).toBe(
      'line1\nJohn\nline3',
    );
  });
});
