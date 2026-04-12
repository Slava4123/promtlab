// API-клиент для backend PromptVault. Работает только из background service worker.

import { getSettings } from './storage';
import {
  ApiError,
  type CollectionDTO,
  type MeResponse,
  type PaginatedPrompts,
  type Prompt,
  type SearchResult,
  type StreakDTO,
  type TagDTO,
  type TeamDTO,
} from './types';

const EXTENSION_VERSION = chrome.runtime.getManifest?.().version ?? '0.1.0';

function timezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone ?? '';
  } catch {
    return '';
  }
}

async function request<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const { apiKey, apiBase } = await getSettings();
  if (!apiKey) throw new ApiError('missing api key', 401, 'unauthorized');

  const headers = new Headers(init.headers);
  headers.set('Authorization', `Bearer ${apiKey}`);
  headers.set('X-Client', `chrome-extension/${EXTENSION_VERSION}`);
  headers.set('Accept', 'application/json');
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const tz = timezone();
  if (tz) headers.set('X-Timezone', tz);

  let res: Response;
  try {
    res = await fetch(`${apiBase}${path}`, { ...init, headers });
  } catch {
    throw new ApiError('network error', 0, 'network');
  }

  if (res.status === 204) {
    return undefined as T;
  }

  if (!res.ok) {
    if (res.status === 401) throw new ApiError('unauthorized', 401, 'unauthorized');
    if (res.status === 404) throw new ApiError('not found', 404, 'not_found');
    if (res.status === 429) throw new ApiError('rate limited', 429, 'rate_limited');
    throw new ApiError(`http ${res.status}`, res.status, 'network');
  }

  try {
    return (await res.json()) as T;
  } catch {
    throw new ApiError('invalid json response', 500, 'network');
  }
}

// --- API методы ---

export async function getMe(): Promise<MeResponse> {
  return request<MeResponse>('/api/auth/me');
}

export async function health(): Promise<{ ok: true }> {
  await request('/api/auth/me');
  return { ok: true };
}

export interface PromptFilter {
  teamId?: number | null;
  collectionId?: number | null;
  tagIds?: number[];
}

function applyFilter(params: URLSearchParams, filter?: PromptFilter): void {
  if (!filter) return;
  if (filter.teamId) params.set('team_id', String(filter.teamId));
  if (filter.collectionId) params.set('collection_id', String(filter.collectionId));
  if (filter.tagIds && filter.tagIds.length > 0) {
    params.set('tag_ids', filter.tagIds.join(','));
  }
}

export async function listPrompts(
  page = 1,
  pageSize = 100,
  filter?: PromptFilter,
): Promise<PaginatedPrompts> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  applyFilter(params, filter);
  return request<PaginatedPrompts>(`/api/prompts?${params}`);
}

export async function getPrompt(id: number): Promise<Prompt> {
  return request<Prompt>(`/api/prompts/${id}`);
}

export async function getPinnedPrompts(limit = 10, filter?: PromptFilter): Promise<Prompt[]> {
  const params = new URLSearchParams({ limit: String(limit) });
  applyFilter(params, filter);
  const data = await request<Prompt[] | PaginatedPrompts>(
    `/api/prompts/pinned?${params}`,
  );
  return Array.isArray(data) ? data : data.items;
}

export async function getRecentPrompts(limit = 10, filter?: PromptFilter): Promise<Prompt[]> {
  const params = new URLSearchParams({ limit: String(limit) });
  applyFilter(params, filter);
  const data = await request<Prompt[] | PaginatedPrompts>(
    `/api/prompts/recent?${params}`,
  );
  return Array.isArray(data) ? data : data.items;
}

export async function search(q: string, filter?: PromptFilter): Promise<SearchResult> {
  if (!q.trim()) return { prompts: [], collections: [], tags: [] };
  const params = new URLSearchParams({ q });
  applyFilter(params, filter);
  return request<SearchResult>(`/api/search?${params}`);
}

// ===== Teams, collections, tags, streaks =====

export async function listTeams(): Promise<TeamDTO[]> {
  const data = await request<TeamDTO[] | { items: TeamDTO[] }>('/api/teams');
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function listCollections(teamId?: number | null): Promise<CollectionDTO[]> {
  const params = new URLSearchParams();
  if (teamId) params.set('team_id', String(teamId));
  const suffix = params.toString() ? `?${params}` : '';
  const data = await request<CollectionDTO[] | { items: CollectionDTO[] }>(
    `/api/collections${suffix}`,
  );
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function listTags(teamId?: number | null): Promise<TagDTO[]> {
  const params = new URLSearchParams();
  if (teamId) params.set('team_id', String(teamId));
  const suffix = params.toString() ? `?${params}` : '';
  const data = await request<TagDTO[] | { items: TagDTO[] }>(`/api/tags${suffix}`);
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function getStreak(): Promise<StreakDTO> {
  return request<StreakDTO>('/api/streaks');
}

export async function createShareLink(promptId: number): Promise<{ token: string; url: string }> {
  return request<{ token: string; url: string }>(`/api/prompts/${promptId}/share`, {
    method: 'POST',
  });
}

export async function incrementUsage(promptId: number): Promise<void> {
  await request<{ message?: string }>(`/api/prompts/${promptId}/use`, {
    method: 'POST',
  });
}

export interface PinResult {
  pinned: boolean;
  team_wide: boolean;
}

export async function toggleFavorite(promptId: number): Promise<Prompt> {
  return request<Prompt>(`/api/prompts/${promptId}/favorite`, { method: 'POST' });
}

export async function togglePin(promptId: number): Promise<PinResult> {
  return request<PinResult>(`/api/prompts/${promptId}/pin`, { method: 'POST' });
}

export async function validateKey(
  apiKey: string,
  apiBase?: string,
): Promise<MeResponse> {
  const { apiBase: storedBase } = await getSettings();
  const base = apiBase ?? storedBase;
  const headers = new Headers();
  headers.set('Authorization', `Bearer ${apiKey}`);
  headers.set('X-Client', `chrome-extension/${EXTENSION_VERSION}`);
  headers.set('Accept', 'application/json');

  let res: Response;
  try {
    res = await fetch(`${base}/api/auth/me`, { headers });
  } catch {
    throw new ApiError('network error', 0, 'network');
  }
  if (res.status === 401) throw new ApiError('unauthorized', 401, 'unauthorized');
  if (!res.ok) throw new ApiError(`http ${res.status}`, res.status, 'network');
  return (await res.json()) as MeResponse;
}
