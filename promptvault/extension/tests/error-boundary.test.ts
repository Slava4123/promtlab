// Тесты для isChunkLoadError — pure function, ловит chunk-load-errors
// после rebuild расширения. Без авто-recovery юзер увидит generic
// «Что-то пошло не так» в момент когда ему нужен reload.
//
// Reload-flow самого ErrorBoundary тестировать нет смысла: location.reload()
// уничтожает JS-контекст, поведение можно проверить только E2E.

import { describe, expect, it } from 'vitest';

import { isChunkLoadError } from '../components/error-boundary';

describe('isChunkLoadError', () => {
  it('true для err.name === "ChunkLoadError" (Webpack/Firefox/WebKit)', () => {
    const err = new Error('whatever');
    err.name = 'ChunkLoadError';
    expect(isChunkLoadError(err)).toBe(true);
  });

  it.each([
    'Failed to fetch dynamically imported module: http://...',
    'Importing a module script failed',
    'error loading dynamically imported module',
    'Loading chunk 42 failed',
    'Loading CSS chunk main failed',
    'Unable to preload CSS for /assets/x.css',
    'Unable to preload module for /assets/y.js',
  ])('true для "%s"', (msg) => {
    expect(isChunkLoadError(new Error(msg))).toBe(true);
  });

  it('case-insensitive', () => {
    expect(isChunkLoadError(new Error('LOADING CHUNK 1 FAILED'))).toBe(true);
  });

  it('false для обычной TypeError', () => {
    expect(isChunkLoadError(new TypeError('Cannot read property of undefined'))).toBe(false);
  });

  it('false для обычной Error без узнаваемого message', () => {
    expect(isChunkLoadError(new Error('Network request failed'))).toBe(false);
  });

  it('false для пустого message', () => {
    expect(isChunkLoadError(new Error(''))).toBe(false);
  });

  it('true когда name=ChunkLoadError даже при пустом message', () => {
    const err = new Error('');
    err.name = 'ChunkLoadError';
    expect(isChunkLoadError(err)).toBe(true);
  });

  it('false для синтетического message со словом "chunk", не подпадающего под паттерны', () => {
    // Гарантируем, что мы не ловим случайные совпадения вроде "chunk-encoding".
    expect(isChunkLoadError(new Error('chunk-encoding mismatch'))).toBe(false);
  });
});
