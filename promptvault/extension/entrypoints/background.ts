// Background service worker — брокер между Side Panel, Content Scripts и backend API.

import { defineBackground } from 'wxt/utils/define-background';
import {
  createShareLink,
  getMe,
  getPinnedPrompts,
  getPrompt,
  getRecentPrompts,
  getStreak,
  health,
  incrementUsage,
  listCollections,
  listPrompts,
  listTags,
  listTeams,
  search,
  toggleFavorite,
  togglePin,
  validateKey,
} from '../lib/api';
import { ApiError } from '../lib/types';
import { isSupportedHost, SUPPORTED_HOSTS_LIST } from '../lib/messages-helpers';
import { clearLastInsert, getLastInsert, setLastInsert } from '../lib/storage';
import type {
  BgError,
  BgRequest,
  BgResponse,
  ContentCommand,
  ContentResponse,
  InsertStrategy,
} from '../lib/messages';

export default defineBackground(() => {
  chrome.sidePanel
    ?.setPanelBehavior?.({ openPanelOnActionClick: true })
    .catch((err) => console.warn('sidePanel.setPanelBehavior failed', err));

  chrome.runtime.onInstalled.addListener(() => {
    console.info('PromptVault extension installed');
  });

  chrome.runtime.onMessage.addListener(
    (
      msg: BgRequest,
      _sender,
      sendResponse: (response: BgResponse) => void,
    ) => {
      handleRequest(msg)
        .then((data) => sendResponse({ ok: true, data }))
        .catch((err: unknown) => sendResponse(toErrorResponse(err)));
      return true;
    },
  );
});

async function handleRequest(msg: BgRequest): Promise<unknown> {
  switch (msg.type) {
    case 'api.fetchPrompts':
      return listPrompts(msg.page ?? 1, msg.pageSize ?? 100, msg.filter ?? undefined);
    case 'api.searchPrompts':
      return search(msg.q, msg.filter ?? undefined);
    case 'api.getPrompt':
      return getPrompt(msg.id);
    case 'api.getPinned':
      return getPinnedPrompts(msg.limit, msg.filter ?? undefined);
    case 'api.getRecent':
      return getRecentPrompts(msg.limit, msg.filter ?? undefined);
    case 'api.incrementUsage':
      await incrementUsage(msg.promptId);
      return { ok: true };
    case 'api.toggleFavorite':
      return toggleFavorite(msg.promptId);
    case 'api.togglePin':
      return togglePin(msg.promptId);
    case 'api.getMe':
      return getMe();
    case 'api.validateKey':
      return validateKey(msg.key);
    case 'api.health':
      return health();
    case 'api.listTeams':
      return listTeams();
    case 'api.listCollections':
      return listCollections(msg.teamId ?? null);
    case 'api.listTags':
      return listTags(msg.teamId ?? null);
    case 'api.getStreak':
      return getStreak();
    case 'api.createShareLink':
      return createShareLink(msg.promptId);
    case 'cmd.insertPrompt':
      return insertIntoActiveTab(msg.text);
    case 'cmd.insertPromptAll':
      return insertIntoAllSupportedTabs(msg.text);
    case 'cmd.undoInsert':
      return undoInActiveTab();
    case 'cmd.getActiveHost':
      return getActiveHost();
    default:
      throw new Error('unknown message type');
  }
}

async function getActiveHost(): Promise<{ host: string | null; supported: boolean }> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.url) return { host: null, supported: false };
  try {
    const host = new URL(tab.url).host;
    return { host, supported: isSupportedHost(host) };
  } catch {
    return { host: null, supported: false };
  }
}

async function insertIntoActiveTab(
  text: string,
): Promise<{ strategy: InsertStrategy }> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab || typeof tab.id !== 'number') {
    throw Object.assign(new Error('no active tab'), { code: 'no_target' });
  }

  const cmd: ContentCommand = { type: 'content.insert', text };
  let response: ContentResponse | undefined;
  try {
    response = await chrome.tabs.sendMessage<ContentCommand, ContentResponse>(
      tab.id,
      cmd,
    );
  } catch (err) {
    throw Object.assign(new Error('content script not reachable'), {
      code: 'no_target',
      cause: err,
    });
  }

  if (!response) {
    throw Object.assign(new Error('no response from content script'), {
      code: 'no_target',
    });
  }
  if (response.type === 'content.inserted') {
    await setLastInsert({ promptId: 0, text, ts: Date.now() });
    return { strategy: response.strategy };
  }
  if (response.type === 'content.notFound') {
    throw Object.assign(new Error('target input not found'), { code: 'no_target' });
  }
  if (response.type === 'content.failed') {
    throw Object.assign(new Error(response.reason || 'insertion failed'), { code: 'unknown' });
  }
  throw Object.assign(new Error('unexpected content response'), { code: 'unknown' });
}

async function insertIntoAllSupportedTabs(
  text: string,
): Promise<{ count: number; successes: number }> {
  const patterns = SUPPORTED_HOSTS_LIST.map((h) => `*://${h}/*`);
  const tabs = await chrome.tabs.query({ url: patterns });
  let successes = 0;
  const cmd: ContentCommand = { type: 'content.insert', text };
  for (const tab of tabs) {
    if (typeof tab.id !== 'number') continue;
    try {
      const response = await chrome.tabs.sendMessage<ContentCommand, ContentResponse>(
        tab.id,
        cmd,
      );
      if (response?.type === 'content.inserted') {
        successes++;
      }
    } catch {
      // ignore unreachable tabs
    }
  }
  await setLastInsert({ promptId: 0, text, ts: Date.now() });
  return { count: tabs.length, successes };
}

async function undoInActiveTab(): Promise<{ ok: true }> {
  const last = await getLastInsert();
  if (!last) throw Object.assign(new Error('nothing to undo'), { code: 'no_history' });

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab || typeof tab.id !== 'number') {
    throw Object.assign(new Error('no active tab'), { code: 'no_target' });
  }

  const cmd: ContentCommand = { type: 'content.undo' };
  let response: ContentResponse | undefined;
  try {
    response = await chrome.tabs.sendMessage<ContentCommand, ContentResponse>(tab.id, cmd);
  } catch (err) {
    throw Object.assign(new Error('content script not reachable'), {
      code: 'no_target',
      cause: err,
    });
  }
  if (response?.type === 'content.undone') {
    await clearLastInsert();
    return { ok: true };
  }
  throw Object.assign(new Error('undo failed'), { code: 'unknown' });
}

function toErrorResponse(err: unknown): BgResponse {
  if (err instanceof ApiError) {
    const code = (err.code ?? 'unknown') as BgError;
    return { ok: false, error: code, message: err.message };
  }
  if (err && typeof err === 'object' && 'code' in err) {
    const code = ((err as { code?: string }).code ?? 'unknown') as BgError;
    const message = err instanceof Error ? err.message : 'error';
    return { ok: false, error: code, message };
  }
  const message = err instanceof Error ? err.message : 'unknown error';
  return { ok: false, error: 'unknown', message };
}
