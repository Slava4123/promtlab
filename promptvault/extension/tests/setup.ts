// Глобальная инициализация тестового окружения. Подменяем chrome.* APIs
// которые недоступны в happy-dom — extension-код их активно использует
// в lib/storage.ts, entrypoints/background.ts и др.

import { vi, beforeEach } from 'vitest';

interface StorageBackend {
  data: Record<string, unknown>;
  listeners: Array<
    (
      changes: Record<string, { oldValue?: unknown; newValue?: unknown }>,
      area: 'local' | 'sync' | 'session' | 'managed',
    ) => void
  >;
}

const local: StorageBackend = { data: {}, listeners: [] };

function notify(area: 'local', changes: Record<string, { oldValue?: unknown; newValue?: unknown }>): void {
  for (const fn of local.listeners) fn(changes, area);
}

const chromeMock = {
  storage: {
    local: {
      get: vi.fn(async (keys?: string | string[] | Record<string, unknown> | null) => {
        if (keys === null || keys === undefined) return { ...local.data };
        if (typeof keys === 'string') return { [keys]: local.data[keys] };
        if (Array.isArray(keys)) {
          const out: Record<string, unknown> = {};
          for (const k of keys) out[k] = local.data[k];
          return out;
        }
        const out: Record<string, unknown> = {};
        for (const k of Object.keys(keys)) {
          out[k] = local.data[k] ?? (keys as Record<string, unknown>)[k];
        }
        return out;
      }),
      set: vi.fn(async (items: Record<string, unknown>) => {
        const changes: Record<string, { oldValue?: unknown; newValue?: unknown }> = {};
        for (const [k, v] of Object.entries(items)) {
          changes[k] = { oldValue: local.data[k], newValue: v };
          local.data[k] = v;
        }
        notify('local', changes);
      }),
      remove: vi.fn(async (keys: string | string[]) => {
        const arr = Array.isArray(keys) ? keys : [keys];
        const changes: Record<string, { oldValue?: unknown; newValue?: unknown }> = {};
        for (const k of arr) {
          changes[k] = { oldValue: local.data[k] };
          delete local.data[k];
        }
        notify('local', changes);
      }),
    },
  },
  runtime: {
    getManifest: vi.fn(() => ({ version: '0.0.0-test' })),
    sendMessage: vi.fn(),
    onMessage: { addListener: vi.fn(), removeListener: vi.fn() },
    onInstalled: { addListener: vi.fn() },
  },
};

// Подключаем listener API к нашему backend.
(chromeMock.storage as unknown as { onChanged: unknown }).onChanged = {
  addListener: (fn: (typeof local.listeners)[number]) => {
    local.listeners.push(fn);
  },
  removeListener: (fn: (typeof local.listeners)[number]) => {
    const idx = local.listeners.indexOf(fn);
    if (idx >= 0) local.listeners.splice(idx, 1);
  },
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
(globalThis as any).chrome = chromeMock;

// Сброс состояния между тестами — изоляция.
beforeEach(() => {
  local.data = {};
  local.listeners = [];
  vi.clearAllMocks();
});
