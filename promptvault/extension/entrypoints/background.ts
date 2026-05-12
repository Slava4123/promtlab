// Background service worker — брокер между Side Panel, Content Scripts и backend API.

import { defineBackground } from 'wxt/utils/define-background';
import {
  acceptInvitation,
  advanceChainStep,
  cancelSubscription,
  changePassword,
  createApiKey,
  declineInvitation,
  deleteTeamLogo,
  getReferral,
  getTeamActivity,
  getTeamAnalytics,
  getTeamBranding,
  listLinkedAccounts,
  setInsightEmails,
  unlinkProvider,
  updateTeamBranding,
  createCollection,
  createPrompt,
  createShareLink,
  createTag,
  createTeam,
  deactivateShareLink,
  deleteApiKey,
  deleteCollection,
  deletePrompt,
  deleteTag,
  deleteTeam,
  duplicatePrompt,
  emptyTrash,
  getChain,
  getChangelog,
  getCurrentSubscription,
  getExecution,
  getInsights,
  getMe,
  getPersonalAnalytics,
  getPinnedPrompts,
  getPrompt,
  getRecentPrompts,
  getShareLink,
  getStreak,
  getStreakDetail,
  getTeam,
  getUsageSummary,
  health,
  incrementUsage,
  inviteTeamMember,
  listApiKeys,
  listBadges,
  listChains,
  listCollections,
  listExecutions,
  listMyInvitations,
  listPlans,
  listPrompts,
  listTags,
  listTeams,
  listTrash,
  listUsageHistory,
  listVersions,
  markChangelogRead,
  pauseSubscription,
  permanentDeleteTrashCollection,
  permanentDeleteTrashPrompt,
  refreshInsights,
  removeTeamMember,
  restoreTrashCollection,
  restoreTrashPrompt,
  resumeSubscription,
  revertVersion,
  search,
  startChainExecution,
  submitFeedback,
  toggleFavorite,
  togglePin,
  updateCollection,
  updatePrompt,
  updateProfile,
  updateTeam,
  updateTeamMemberRole,
  validateKey,
} from '../lib/api';
import { ApiError } from '../lib/types';
import { isSupportedHost, SUPPORTED_HOSTS_LIST } from '../lib/messages-helpers';
import { initSentry, captureException, addBreadcrumb } from '../lib/sentry';
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
  initSentry({
    enabled: import.meta.env.WXT_SENTRY_DSN ? true : false,
    release: chrome.runtime.getManifest?.().version,
    dsn: import.meta.env.WXT_SENTRY_DSN,
  });

  chrome.sidePanel
    ?.setPanelBehavior?.({ openPanelOnActionClick: true })
    .catch((err) => console.warn('sidePanel.setPanelBehavior failed', err));

  chrome.runtime.onInstalled.addListener((details) => {
    console.info('PromptVault extension installed', details.reason);
    setupContextMenus();
    // Re-inject content scripts во все открытые AI-вкладки. Без этого
    // MV3 не обновляет существующие scripts при extension reload — юзер
    // получает stale content script → sendMessage → no_target errors.
    void reinjectContentScripts();
  });

  // Re-create при старте, на случай если onInstalled не выполнился.
  setupContextMenus();

  chrome.contextMenus?.onClicked.addListener(async (info, tab) => {
    if (info.menuItemId !== 'pv-save-selection') return;
    const selection = info.selectionText ?? '';
    const pageUrl = info.pageUrl ?? tab?.url ?? '';
    if (!selection.trim()) return;
    await chrome.storage.session?.set({
      'pv.pendingCapture': {
        content: selection,
        sourceUrl: pageUrl,
        capturedAt: Date.now(),
      },
    });
    try {
      if (tab?.windowId !== undefined) {
        await (chrome.sidePanel as unknown as { open: (o: { windowId: number }) => Promise<void> })?.open?.({
          windowId: tab.windowId,
        });
      }
    } catch {
      // ignore
    }
  });

  chrome.runtime.onMessage.addListener(
    (
      msg: BgRequest,
      _sender,
      sendResponse: (response: BgResponse) => void,
    ) => {
      addBreadcrumb('bg.request', msg.type, undefined, 'info');
      handleRequest(msg)
        .then((data) => sendResponse({ ok: true, data }))
        .catch((err: unknown) => {
          const resp = toErrorResponse(err);
          if (!resp.ok && resp.error === 'unknown') {
            captureException(err, { msgType: msg.type });
          }
          sendResponse(resp);
        });
      return true;
    },
  );
});

async function handleRequest(msg: BgRequest): Promise<unknown> {
  switch (msg.type) {
    // --- Auth / Me ---
    case 'api.getMe':
      return getMe();
    case 'api.validateKey':
      return validateKey(msg.key);
    case 'api.health':
      return health();
    case 'api.updateProfile':
      return updateProfile(msg.body);
    case 'api.changePassword':
      await changePassword(msg.oldPassword, msg.newPassword);
      return { ok: true };

    // --- Prompts list ---
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

    // --- Prompts mutations ---
    case 'api.createPrompt':
      return createPrompt(msg.body);
    case 'api.updatePrompt':
      return updatePrompt(msg.id, msg.body);
    case 'api.deletePrompt':
      await deletePrompt(msg.id);
      return { ok: true };
    case 'api.duplicatePrompt':
      return duplicatePrompt(msg.id);
    case 'api.incrementUsage':
      await incrementUsage(msg.promptId);
      return { ok: true };
    case 'api.toggleFavorite':
      return toggleFavorite(msg.promptId);
    case 'api.togglePin':
      return togglePin(msg.promptId);

    // --- Versions ---
    case 'api.listVersions':
      return listVersions(msg.promptId, msg.limit, msg.offset);
    case 'api.revertVersion':
      return revertVersion(msg.promptId, msg.versionId);

    // --- Trash ---
    case 'api.listTrash':
      return listTrash();
    case 'api.restoreTrashPrompt':
      return restoreTrashPrompt(msg.id);
    case 'api.restoreTrashCollection':
      return restoreTrashCollection(msg.id);
    case 'api.permanentDeleteTrashPrompt':
      await permanentDeleteTrashPrompt(msg.id);
      return { ok: true };
    case 'api.permanentDeleteTrashCollection':
      await permanentDeleteTrashCollection(msg.id);
      return { ok: true };
    case 'api.emptyTrash':
      await emptyTrash();
      return { ok: true };

    // --- Share ---
    case 'api.getShareLink':
      return getShareLink(msg.promptId);
    case 'api.createShareLink':
      return createShareLink(msg.promptId);
    case 'api.deactivateShareLink':
      await deactivateShareLink(msg.promptId);
      return { ok: true };

    // --- History / Analytics ---
    case 'api.listUsageHistory':
      return listUsageHistory(msg.limit, msg.offset);
    case 'api.getPersonalAnalytics':
      return getPersonalAnalytics(msg.range);
    case 'api.getInsights':
      return getInsights();
    case 'api.refreshInsights':
      return refreshInsights();

    // --- Collections CRUD ---
    case 'api.listCollections':
      return listCollections(msg.teamId ?? null);
    case 'api.createCollection':
      return createCollection(msg.body);
    case 'api.updateCollection':
      return updateCollection(msg.id, msg.body);
    case 'api.deleteCollection':
      await deleteCollection(msg.id);
      return { ok: true };

    // --- Tags CRUD ---
    case 'api.listTags':
      return listTags(msg.teamId ?? null);
    case 'api.createTag':
      return createTag(msg.body);
    case 'api.deleteTag':
      await deleteTag(msg.id);
      return { ok: true };

    // --- Teams ---
    case 'api.listTeams':
      return listTeams();
    case 'api.getTeam':
      return getTeam(msg.slug);
    case 'api.createTeam':
      return createTeam(msg.body);
    case 'api.updateTeam':
      return updateTeam(msg.slug, msg.body);
    case 'api.deleteTeam':
      await deleteTeam(msg.slug);
      return { ok: true };
    case 'api.inviteTeamMember':
      await inviteTeamMember(msg.slug, { email: msg.email, role: msg.role });
      return { ok: true };
    case 'api.removeTeamMember':
      await removeTeamMember(msg.slug, msg.memberId);
      return { ok: true };
    case 'api.updateTeamMemberRole':
      await updateTeamMemberRole(msg.slug, msg.memberId, msg.role);
      return { ok: true };

    // --- Invitations ---
    case 'api.listMyInvitations':
      return listMyInvitations();
    case 'api.acceptInvitation':
      await acceptInvitation(msg.invitationId);
      return { ok: true };
    case 'api.declineInvitation':
      await declineInvitation(msg.invitationId);
      return { ok: true };

    // --- Feedback ---
    case 'api.submitFeedback':
      return submitFeedback(msg.body);

    // --- Notifications / Linked accounts / Referral ---
    case 'api.setInsightEmails':
      return setInsightEmails(msg.enabled);
    case 'api.listLinkedAccounts':
      return listLinkedAccounts();
    case 'api.unlinkProvider':
      await unlinkProvider(msg.provider);
      return { ok: true };
    case 'api.getReferral':
      return getReferral();

    // --- Team Branding/Analytics/Activity ---
    case 'api.getTeamBranding':
      return getTeamBranding(msg.slug);
    case 'api.updateTeamBranding':
      return updateTeamBranding(msg.slug, msg.body);
    case 'api.deleteTeamLogo':
      await deleteTeamLogo(msg.slug);
      return { ok: true };
    case 'api.getTeamAnalytics':
      return getTeamAnalytics(msg.teamId, msg.range);
    case 'api.getTeamActivity':
      return getTeamActivity(msg.slug, msg.page, msg.pageSize);

    // --- Streak / Badges / Changelog ---
    case 'api.getStreak':
      return getStreak();
    case 'api.getStreakDetail':
      return getStreakDetail();
    case 'api.listBadges':
      return listBadges();
    case 'api.getChangelog':
      return getChangelog();
    case 'api.markChangelogRead':
      await markChangelogRead();
      return { ok: true };

    // --- Subscription ---
    case 'api.listPlans':
      return listPlans();
    case 'api.getCurrentSubscription':
      return getCurrentSubscription();
    case 'api.getUsageSummary':
      return getUsageSummary();
    case 'api.cancelSubscription':
      await cancelSubscription();
      return { ok: true };
    case 'api.pauseSubscription':
      await pauseSubscription();
      return { ok: true };
    case 'api.resumeSubscription':
      await resumeSubscription();
      return { ok: true };

    // --- API Keys ---
    case 'api.listApiKeys':
      return listApiKeys();
    case 'api.createApiKey':
      return createApiKey(msg.body);
    case 'api.deleteApiKey':
      await deleteApiKey(msg.id);
      return { ok: true };

    // --- Chains ---
    case 'api.listChains':
      return listChains(msg.teamId ?? null);
    case 'api.getChain':
      return getChain(msg.id);
    case 'api.startChainExecution':
      return startChainExecution(msg.chainId, msg.initialVars);
    case 'api.getExecution':
      return getExecution(msg.execId);
    case 'api.advanceChainStep':
      return advanceChainStep(msg.execId, msg.stepOutput, msg.chosenBranchIndex);
    case 'api.listExecutions':
      return listExecutions(msg.chainId);

    // --- Content commands ---
    case 'cmd.insertPrompt':
      return insertIntoActiveTab(msg.text);
    case 'cmd.insertPromptAll':
      return insertIntoAllSupportedTabs(msg.text);
    case 'cmd.undoInsert':
      return undoInActiveTab();
    case 'cmd.getActiveHost':
      return getActiveHost();
    default: {
      const exhaust: never = msg;
      throw new Error(`unknown message type: ${JSON.stringify(exhaust)}`);
    }
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

function setupContextMenus(): void {
  if (!chrome.contextMenus) return;
  chrome.contextMenus.removeAll(() => {
    chrome.contextMenus.create({
      id: 'pv-save-selection',
      title: 'ПромтЛаб: сохранить как промпт',
      contexts: ['selection'],
    });
  });
}

// Маппинг host → content-script file (берётся из output WXT после build).
// Re-inject обходит ограничение MV3 «scripts не обновляются при reload extension».
const HOST_SCRIPTS: Array<{ pattern: string; file: string }> = [
  { pattern: 'https://chatgpt.com/*', file: 'content-scripts/chatgpt.js' },
  { pattern: 'https://claude.ai/*', file: 'content-scripts/claude.js' },
  { pattern: 'https://gemini.google.com/*', file: 'content-scripts/gemini.js' },
  { pattern: 'https://www.perplexity.ai/*', file: 'content-scripts/perplexity.js' },
  { pattern: 'https://alice.yandex.ru/*', file: 'content-scripts/yandex.js' },
  { pattern: 'https://ya.ru/*', file: 'content-scripts/yandex.js' },
  { pattern: 'https://yandex.ru/alice*', file: 'content-scripts/yandex.js' },
  { pattern: 'https://giga.chat/*', file: 'content-scripts/gigachat.js' },
  { pattern: 'https://developers.sber.ru/*', file: 'content-scripts/gigachat.js' },
  { pattern: 'https://chat.deepseek.com/*', file: 'content-scripts/deepseek.js' },
  { pattern: 'https://chat.mistral.ai/*', file: 'content-scripts/mistral.js' },
  { pattern: 'https://le-chat.mistral.ai/*', file: 'content-scripts/mistral.js' },
  { pattern: 'https://chat.qwen.ai/*', file: 'content-scripts/qwen.js' },
];

async function reinjectContentScripts(): Promise<void> {
  if (!chrome.scripting?.executeScript) return;
  for (const { pattern, file } of HOST_SCRIPTS) {
    try {
      const tabs = await chrome.tabs.query({ url: pattern });
      for (const tab of tabs) {
        if (typeof tab.id !== 'number') continue;
        try {
          await chrome.scripting.executeScript({
            target: { tabId: tab.id },
            files: [file],
          });
          console.info(`reinjected ${file} into tab ${tab.id} (${tab.url})`);
        } catch (err) {
          // Игнорируем ошибки (например chrome-extension://, chrome://, etc.)
          console.warn(`reinject ${file} failed for tab ${tab.id}:`, err);
        }
      }
    } catch (err) {
      console.warn(`tabs.query failed for ${pattern}:`, err);
    }
  }
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
