// Тесты для notifications-read-store — общий source-of-truth для
// прочитанных уведомлений. Раньше NotificationsPage и useUnreadCount
// независимо читали localStorage, и storage event не fires в той же tab
// после setItem — дрифт между компонентами.

import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useNotificationsReadStore } from '../stores/notifications-read-store';

// Изолируем stор между тестами — clear + сброс persist'а в storage.
beforeEach(() => {
  useNotificationsReadStore.getState().clear();
});

afterEach(() => {
  useNotificationsReadStore.persist?.clearStorage();
});

describe('notifications-read-store', () => {
  it('пустой стор: ids = []', () => {
    expect(useNotificationsReadStore.getState().ids).toEqual([]);
  });

  it('markRead добавляет id в список', () => {
    useNotificationsReadStore.getState().markRead('invitation-1');
    expect(useNotificationsReadStore.getState().ids).toEqual(['invitation-1']);
  });

  it('markRead дедуплицирует повторные id', () => {
    const { markRead } = useNotificationsReadStore.getState();
    markRead('quota-prompts');
    markRead('quota-prompts');
    markRead('quota-prompts');
    expect(useNotificationsReadStore.getState().ids).toEqual(['quota-prompts']);
  });

  it('markRead сохраняет порядок первого появления', () => {
    const { markRead } = useNotificationsReadStore.getState();
    markRead('a');
    markRead('b');
    markRead('c');
    expect(useNotificationsReadStore.getState().ids).toEqual(['a', 'b', 'c']);
  });

  it('markAllRead мерджит массив без дубликатов', () => {
    const { markRead, markAllRead } = useNotificationsReadStore.getState();
    markRead('a');
    markAllRead(['b', 'a', 'c']);
    expect(useNotificationsReadStore.getState().ids.sort()).toEqual(['a', 'b', 'c']);
  });

  it('markAllRead с пустым массивом ничего не меняет', () => {
    useNotificationsReadStore.getState().markRead('x');
    useNotificationsReadStore.getState().markAllRead([]);
    expect(useNotificationsReadStore.getState().ids).toEqual(['x']);
  });

  it('isRead true для известного id', () => {
    useNotificationsReadStore.getState().markRead('quota-chains');
    expect(useNotificationsReadStore.getState().isRead('quota-chains')).toBe(true);
  });

  it('isRead false для неизвестного id', () => {
    expect(useNotificationsReadStore.getState().isRead('never-marked')).toBe(false);
  });

  it('clear обнуляет ids', () => {
    const { markRead, clear } = useNotificationsReadStore.getState();
    markRead('a');
    markRead('b');
    clear();
    expect(useNotificationsReadStore.getState().ids).toEqual([]);
  });

  it('persist пишет под ключом pv-notifications-read', () => {
    useNotificationsReadStore.getState().markRead('persist-check');
    // chrome.storage mock из setup.ts не используется persist'ом — он работает
    // с localStorage. happy-dom предоставляет localStorage.
    const raw = localStorage.getItem('pv-notifications-read');
    expect(raw).not.toBeNull();
    const parsed = JSON.parse(raw ?? '{}');
    expect(parsed.state?.ids).toContain('persist-check');
  });
});
