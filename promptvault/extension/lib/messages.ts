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
} from './types';

// ===== Side Panel → Background =====

export interface PromptFilterMessage {
  teamId?: number | null;
  collectionId?: number | null;
  tagIds?: number[];
}

export type BgRequest =
  | { type: 'api.fetchPrompts'; page?: number; pageSize?: number; filter?: PromptFilterMessage }
  | { type: 'api.searchPrompts'; q: string; filter?: PromptFilterMessage }
  | { type: 'api.getPrompt'; id: number }
  | { type: 'api.getPinned'; limit?: number; filter?: PromptFilterMessage }
  | { type: 'api.getRecent'; limit?: number; filter?: PromptFilterMessage }
  | { type: 'api.incrementUsage'; promptId: number }
  | { type: 'api.toggleFavorite'; promptId: number }
  | { type: 'api.togglePin'; promptId: number }
  | { type: 'api.getMe' }
  | { type: 'api.validateKey'; key: string }
  | { type: 'api.health' }
  | { type: 'api.listTeams' }
  | { type: 'api.listCollections'; teamId?: number | null }
  | { type: 'api.listTags'; teamId?: number | null }
  | { type: 'api.getStreak' }
  | { type: 'api.createShareLink'; promptId: number }
  | { type: 'cmd.insertPrompt'; text: string }
  | { type: 'cmd.insertPromptAll'; text: string }
  | { type: 'cmd.undoInsert' }
  | { type: 'cmd.getActiveHost' };

export type BgResponse<T = unknown> =
  | { ok: true; data: T }
  | { ok: false; error: BgError; message?: string };

export type BgError =
  | 'unauthorized'
  | 'not_found'
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
  'api.fetchPrompts': PaginatedPrompts;
  'api.searchPrompts': SearchResult;
  'api.getPrompt': Prompt;
  'api.getPinned': Prompt[];
  'api.getRecent': Prompt[];
  'api.incrementUsage': { ok: true };
  'api.toggleFavorite': Prompt;
  'api.togglePin': PinResultMessage;
  'api.getMe': MeResponse;
  'api.validateKey': MeResponse;
  'api.health': { ok: true };
  'api.listTeams': TeamDTO[];
  'api.listCollections': CollectionDTO[];
  'api.listTags': TagDTO[];
  'api.getStreak': StreakDTO;
  'api.createShareLink': { token: string; url: string };
  'cmd.insertPrompt': { strategy: InsertStrategy };
  'cmd.insertPromptAll': { count: number; successes: number };
  'cmd.undoInsert': { ok: true };
  'cmd.getActiveHost': { host: string | null; supported: boolean };
}

// ===== Background → Content Script =====

export type ContentCommand =
  | { type: 'content.ping' }
  | { type: 'content.insert'; text: string }
  | { type: 'content.undo' };

export type ContentResponse =
  | { type: 'content.pong'; host: string }
  | { type: 'content.inserted'; strategy: InsertStrategy }
  | { type: 'content.undone' }
  | { type: 'content.notFound' }
  | { type: 'content.failed'; reason: string };

export type InsertStrategy = 'nativeSetter' | 'execCommand' | 'paste' | 'fallback';

export const HOST_LABELS: Record<string, string> = {
  'chatgpt.com': 'ChatGPT',
  'claude.ai': 'Claude',
  'gemini.google.com': 'Gemini',
  'www.perplexity.ai': 'Perplexity',
};

export function hostLabel(host: string | null): string | null {
  if (!host) return null;
  return HOST_LABELS[host] ?? host;
}

export function isSupportedHost(host: string | null): boolean {
  if (!host) return false;
  return host in HOST_LABELS;
}
