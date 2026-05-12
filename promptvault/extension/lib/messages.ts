// Type-safe message passing между Side Panel, Background Service Worker, Content Scripts.

import type {
  CollectionDTO,
  MeResponse,
  PaginatedPrompts,
  Prompt,
  SearchResult,
  StreakDTO,
  TagDTO,
  TeamDTO,
  PromptVersion,
  ShareLink,
  TrashListResponse,
  Collection,
  Tag,
  Team,
  TeamDetail,
  UsageSummary,
  Plan,
  Subscription,
  APIKeyListResponse,
  CreatedAPIKey,
  CreateAPIKeyRequest,
  BadgeListResponse,
  ChangelogResponse,
  StreakResponse,
  ChainListResponse,
  ChainDetail,
  ChainExecution,
  ChainExecutionListResponse,
  TeamInvitation,
  FeedbackRequest,
  FeedbackResponse,
} from './types';
import type {
  CreatePromptBody,
  UpdatePromptBody,
  CreateCollectionBody,
  CreateTagBody,
  CreateTeamBody,
  UpdateProfileBody,
  UsageHistoryResponse,
  PersonalAnalyticsResponse,
  InsightsResponse,
  AnalyticsRange,
} from './api';

// ===== Side Panel → Background =====

export interface PromptFilterMessage {
  teamId?: number | null;
  collectionId?: number | null;
  tagIds?: number[];
  favorite?: boolean;
  search?: string;
}

export type BgRequest =
  // --- Auth / Me ---
  | { type: 'api.getMe' }
  | { type: 'api.validateKey'; key: string }
  | { type: 'api.health' }
  | { type: 'api.updateProfile'; body: UpdateProfileBody }
  | { type: 'api.changePassword'; oldPassword: string; newPassword: string }
  // --- Prompts list ---
  | {
      type: 'api.fetchPrompts'
      page?: number
      pageSize?: number
      filter?: PromptFilterMessage
    }
  | { type: 'api.searchPrompts'; q: string; filter?: PromptFilterMessage }
  | { type: 'api.getPrompt'; id: number }
  | { type: 'api.getPinned'; limit?: number; filter?: PromptFilterMessage }
  | { type: 'api.getRecent'; limit?: number; filter?: PromptFilterMessage }
  // --- Prompts mutations ---
  | { type: 'api.createPrompt'; body: CreatePromptBody }
  | { type: 'api.updatePrompt'; id: number; body: UpdatePromptBody }
  | { type: 'api.deletePrompt'; id: number }
  | { type: 'api.duplicatePrompt'; id: number }
  | { type: 'api.incrementUsage'; promptId: number }
  | { type: 'api.toggleFavorite'; promptId: number }
  | { type: 'api.togglePin'; promptId: number }
  // --- Versions ---
  | { type: 'api.listVersions'; promptId: number; limit?: number; offset?: number }
  | { type: 'api.revertVersion'; promptId: number; versionId: number }
  // --- Trash ---
  | { type: 'api.listTrash' }
  | { type: 'api.restoreTrashPrompt'; id: number }
  | { type: 'api.restoreTrashCollection'; id: number }
  | { type: 'api.permanentDeleteTrashPrompt'; id: number }
  | { type: 'api.permanentDeleteTrashCollection'; id: number }
  | { type: 'api.emptyTrash' }
  // --- Share ---
  | { type: 'api.getShareLink'; promptId: number }
  | { type: 'api.createShareLink'; promptId: number }
  | { type: 'api.deactivateShareLink'; promptId: number }
  // --- History / Analytics ---
  | { type: 'api.listUsageHistory'; limit?: number; offset?: number }
  | { type: 'api.getPersonalAnalytics'; range?: AnalyticsRange }
  | { type: 'api.getInsights' }
  | { type: 'api.refreshInsights' }
  // --- Collections CRUD ---
  | { type: 'api.listCollections'; teamId?: number | null }
  | { type: 'api.createCollection'; body: CreateCollectionBody }
  | { type: 'api.updateCollection'; id: number; body: Partial<CreateCollectionBody> }
  | { type: 'api.deleteCollection'; id: number }
  // --- Tags CRUD ---
  | { type: 'api.listTags'; teamId?: number | null }
  | { type: 'api.createTag'; body: CreateTagBody }
  | { type: 'api.deleteTag'; id: number }
  // --- Teams ---
  | { type: 'api.listTeams' }
  | { type: 'api.getTeam'; slug: string }
  | { type: 'api.createTeam'; body: CreateTeamBody }
  | { type: 'api.updateTeam'; slug: string; body: Partial<CreateTeamBody> }
  | { type: 'api.deleteTeam'; slug: string }
  | {
      type: 'api.inviteTeamMember'
      slug: string
      email: string
      role: 'editor' | 'viewer'
    }
  | { type: 'api.removeTeamMember'; slug: string; memberId: number }
  | {
      type: 'api.updateTeamMemberRole'
      slug: string
      memberId: number
      role: 'owner' | 'editor' | 'viewer'
    }
  // --- Invitations ---
  | { type: 'api.listMyInvitations' }
  | { type: 'api.acceptInvitation'; invitationId: number }
  | { type: 'api.declineInvitation'; invitationId: number }
  // --- Feedback ---
  | { type: 'api.submitFeedback'; body: FeedbackRequest }
  // --- Streak / Badges / Changelog ---
  | { type: 'api.getStreak' }
  | { type: 'api.getStreakDetail' }
  | { type: 'api.listBadges' }
  | { type: 'api.getChangelog' }
  | { type: 'api.markChangelogRead' }
  // --- Subscription ---
  | { type: 'api.listPlans' }
  | { type: 'api.getCurrentSubscription' }
  | { type: 'api.getUsageSummary' }
  | { type: 'api.cancelSubscription' }
  | { type: 'api.pauseSubscription' }
  | { type: 'api.resumeSubscription' }
  // --- API Keys ---
  | { type: 'api.listApiKeys' }
  | { type: 'api.createApiKey'; body: CreateAPIKeyRequest }
  | { type: 'api.deleteApiKey'; id: number }
  // --- Chains ---
  | { type: 'api.listChains'; teamId?: number | null }
  | { type: 'api.getChain'; id: number }
  | { type: 'api.startChainExecution'; chainId: number; initialVars: Record<string, string> }
  | { type: 'api.getExecution'; execId: number }
  | {
      type: 'api.advanceChainStep'
      execId: number
      stepOutput: string
      chosenBranchIndex?: number
    }
  | { type: 'api.listExecutions'; chainId: number }
  // --- Content commands ---
  | { type: 'cmd.insertPrompt'; text: string }
  | { type: 'cmd.insertPromptAll'; text: string }
  | { type: 'cmd.undoInsert' }
  | { type: 'cmd.getActiveHost' };

export type BgResponse<T = unknown> =
  | { ok: true; data: T }
  | { ok: false; error: BgError; message?: string };

export type BgError =
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'conflict'
  | 'validation'
  | 'quota_exceeded'
  | 'network'
  | 'rate_limited'
  | 'unknown'
  | 'no_target'
  | 'no_history';

export interface PinResultMessage {
  pinned: boolean;
  team_wide: boolean;
}

export interface BgResultMap {
  // --- Auth / Me ---
  'api.getMe': MeResponse;
  'api.validateKey': MeResponse;
  'api.health': { ok: true };
  'api.updateProfile': MeResponse;
  'api.changePassword': { ok: true };
  // --- Prompts list ---
  'api.fetchPrompts': PaginatedPrompts;
  'api.searchPrompts': SearchResult;
  'api.getPrompt': Prompt;
  'api.getPinned': Prompt[];
  'api.getRecent': Prompt[];
  // --- Prompts mutations ---
  'api.createPrompt': Prompt;
  'api.updatePrompt': Prompt;
  'api.deletePrompt': { ok: true };
  'api.duplicatePrompt': Prompt;
  'api.incrementUsage': { ok: true };
  'api.toggleFavorite': Prompt;
  'api.togglePin': PinResultMessage;
  // --- Versions ---
  'api.listVersions': { items: PromptVersion[]; total: number; has_more: boolean };
  'api.revertVersion': Prompt;
  // --- Trash ---
  'api.listTrash': TrashListResponse;
  'api.restoreTrashPrompt': Prompt;
  'api.restoreTrashCollection': Collection;
  'api.permanentDeleteTrashPrompt': { ok: true };
  'api.permanentDeleteTrashCollection': { ok: true };
  'api.emptyTrash': { ok: true };
  // --- Share ---
  'api.getShareLink': ShareLink | null;
  'api.createShareLink': ShareLink;
  'api.deactivateShareLink': { ok: true };
  // --- History / Analytics ---
  'api.listUsageHistory': UsageHistoryResponse;
  'api.getPersonalAnalytics': PersonalAnalyticsResponse;
  'api.getInsights': InsightsResponse;
  'api.refreshInsights': InsightsResponse;
  // --- Collections CRUD ---
  'api.listCollections': CollectionDTO[];
  'api.createCollection': Collection;
  'api.updateCollection': Collection;
  'api.deleteCollection': { ok: true };
  // --- Tags CRUD ---
  'api.listTags': TagDTO[];
  'api.createTag': Tag;
  'api.deleteTag': { ok: true };
  // --- Teams ---
  'api.listTeams': TeamDTO[];
  'api.getTeam': TeamDetail;
  'api.createTeam': Team;
  'api.updateTeam': Team;
  'api.deleteTeam': { ok: true };
  'api.inviteTeamMember': { ok: true };
  'api.removeTeamMember': { ok: true };
  'api.updateTeamMemberRole': { ok: true };
  // --- Invitations ---
  'api.listMyInvitations': TeamInvitation[];
  'api.acceptInvitation': { ok: true };
  'api.declineInvitation': { ok: true };
  // --- Feedback ---
  'api.submitFeedback': FeedbackResponse;
  // --- Streak / Badges / Changelog ---
  'api.getStreak': StreakDTO;
  'api.getStreakDetail': StreakResponse;
  'api.listBadges': BadgeListResponse;
  'api.getChangelog': ChangelogResponse;
  'api.markChangelogRead': { ok: true };
  // --- Subscription ---
  'api.listPlans': Plan[];
  'api.getCurrentSubscription': Subscription | null;
  'api.getUsageSummary': UsageSummary;
  'api.cancelSubscription': { ok: true };
  'api.pauseSubscription': { ok: true };
  'api.resumeSubscription': { ok: true };
  // --- API Keys ---
  'api.listApiKeys': APIKeyListResponse;
  'api.createApiKey': CreatedAPIKey;
  'api.deleteApiKey': { ok: true };
  // --- Chains ---
  'api.listChains': ChainListResponse;
  'api.getChain': ChainDetail;
  'api.startChainExecution': ChainExecution;
  'api.getExecution': ChainExecution;
  'api.advanceChainStep': ChainExecution;
  'api.listExecutions': ChainExecutionListResponse;
  // --- Content ---
  'cmd.insertPrompt': { strategy: InsertStrategy };
  'cmd.insertPromptAll': { count: number; successes: number };
  'cmd.undoInsert': { ok: true };
  'cmd.getActiveHost': { host: string | null; supported: boolean };
}

// ===== Background → Content Script =====

export type ContentCommand =
  | { type: 'content.ping' }
  | { type: 'content.insert'; text: string }
  | { type: 'content.undo' }
  | { type: 'content.captureLastAIResponse' };

export type ContentResponse =
  | { type: 'content.pong'; host: string }
  | { type: 'content.inserted'; strategy: InsertStrategy }
  | { type: 'content.undone' }
  | { type: 'content.notFound' }
  | { type: 'content.failed'; reason: string }
  | { type: 'content.captured'; text: string };

export type InsertStrategy = 'nativeSetter' | 'execCommand' | 'paste' | 'fallback';

export const HOST_LABELS: Record<string, string> = {
  'chatgpt.com': 'ChatGPT',
  'claude.ai': 'Claude',
  'gemini.google.com': 'Gemini',
  'www.perplexity.ai': 'Perplexity',
  'alice.yandex.ru': 'Yandex GPT',
  'ya.ru': 'Yandex GPT',
  'yandex.ru': 'Yandex GPT',
  'giga.chat': 'GigaChat',
  'developers.sber.ru': 'GigaChat',
  'chat.deepseek.com': 'DeepSeek',
  'chat.mistral.ai': 'Mistral Le Chat',
  'le-chat.mistral.ai': 'Mistral Le Chat',
  'chat.qwen.ai': 'Qwen',
};

export function hostLabel(host: string | null): string | null {
  if (!host) return null;
  return HOST_LABELS[host] ?? host;
}

export function isSupportedHost(host: string | null): boolean {
  if (!host) return false;
  return host in HOST_LABELS;
}
