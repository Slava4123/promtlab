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
  type PromptVersion,
  type ShareLink,
  type TrashListResponse,
  type Collection,
  type Tag,
  type Team,
  type TeamDetail,
  type UsageSummary,
  type Plan,
  type Subscription,
  type APIKeyListResponse,
  type CreatedAPIKey,
  type CreateAPIKeyRequest,
  type BadgeListResponse,
  type ChangelogResponse,
  type StreakResponse,
  type ChainListResponse,
  type ChainDetail,
  type ChainExecution,
  type ChainExecutionListResponse,
  type TeamInvitation,
  type FeedbackRequest,
  type FeedbackResponse,
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
  // X-Client-Source отдельный заголовок, его backend проверяет в
  // prompt/handler.go::IncrementUsage чтобы инкрементить daily_feature_usage
  // (квоту «Вставки сегодня»). Без него юзер вставляет промпты, а счётчик
  // в Подписке остаётся 0/500.
  headers.set('X-Client-Source', 'extension');
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
    // Извлекаем сообщение из тела для всех 4xx — backend пишет {"error": "..."}
    // или {"message": "..."} в delivery/http/errors. Без этого юзер видел
    // «http 400» вместо нормального текста ошибки.
    const body = await safeJson(res);
    const msg = body?.message ?? body?.error;
    if (res.status === 401) throw new ApiError(msg ?? 'unauthorized', 401, 'unauthorized');
    if (res.status === 402) throw new ApiError(msg ?? 'quota exceeded', 402, 'quota_exceeded');
    if (res.status === 403) throw new ApiError(msg ?? 'forbidden', 403, 'forbidden');
    if (res.status === 404) throw new ApiError(msg ?? 'not found', 404, 'not_found');
    if (res.status === 409) throw new ApiError(msg ?? 'conflict', 409, 'conflict');
    if (res.status === 422) throw new ApiError(msg ?? 'validation', 422, 'validation');
    if (res.status === 429) throw new ApiError(msg ?? 'rate limited', 429, 'rate_limited');
    if (res.status >= 400 && res.status < 500) {
      // Для непознанных 4xx (418/451/410 etc.) код 'client_error', а не
      // 'validation' — иначе form-handlers вроде prompt-editor решат, что
      // это ошибка валидации полей и покажут неуместное сообщение.
      throw new ApiError(msg ?? `http ${res.status}`, res.status, 'client_error');
    }
    throw new ApiError(msg ?? `http ${res.status}`, res.status, 'network');
  }

  try {
    return (await res.json()) as T;
  } catch {
    throw new ApiError('invalid json response', 500, 'network');
  }
}

async function safeJson(res: Response): Promise<{ message?: string; error?: string } | null> {
  try {
    return await res.json();
  } catch {
    return null;
  }
}

// --- Auth/Me ---

export async function getMe(): Promise<MeResponse> {
  return request<MeResponse>('/api/auth/me');
}

export interface UpdateProfileBody {
  name?: string;
  username?: string;
  avatar_url?: string;
}

export async function updateProfile(body: UpdateProfileBody): Promise<MeResponse> {
  return request<MeResponse>('/api/auth/profile', {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  await request<void>('/api/auth/password', {
    method: 'PUT',
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
  });
}

export async function health(): Promise<{ ok: true }> {
  await request('/api/auth/me');
  return { ok: true };
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

// --- Prompts list / filters ---

export interface PromptFilter {
  teamId?: number | null;
  collectionId?: number | null;
  tagIds?: number[];
  favorite?: boolean;
  search?: string;
}

function applyFilter(params: URLSearchParams, filter?: PromptFilter): void {
  if (!filter) return;
  if (filter.teamId) params.set('team_id', String(filter.teamId));
  if (filter.collectionId) params.set('collection_id', String(filter.collectionId));
  if (filter.tagIds && filter.tagIds.length > 0) {
    params.set('tag_ids', filter.tagIds.join(','));
  }
  if (filter.favorite) params.set('favorite_only', 'true');
  if (filter.search) params.set('q', filter.search);
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

// --- Prompts CRUD (Phase 1) ---

export interface CreatePromptBody {
  title: string;
  content: string;
  model?: string;
  collection_ids?: number[];
  tag_ids?: number[];
  team_id?: number | null;
  is_public?: boolean;
  description?: string;
}

export async function createPrompt(body: CreatePromptBody): Promise<Prompt> {
  return request<Prompt>('/api/prompts', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export interface UpdatePromptBody extends Partial<CreatePromptBody> {
  change_note?: string;
}

export async function updatePrompt(id: number, body: UpdatePromptBody): Promise<Prompt> {
  return request<Prompt>(`/api/prompts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function deletePrompt(id: number): Promise<void> {
  await request<void>(`/api/prompts/${id}`, { method: 'DELETE' });
}

// NOTE: backend не имеет /duplicate endpoint — реализуем client-side через
// GET + POST. Сохранение в той же команде/коллекциях/тегах, is_public сброшен.
export async function duplicatePrompt(id: number): Promise<Prompt> {
  const original = await getPrompt(id);
  return createPrompt({
    title: `${original.title} (копия)`,
    content: original.content,
    model: original.model,
    collection_ids: original.collections.map((c) => c.id),
    tag_ids: original.tags.map((t) => t.id),
    is_public: false,
  });
}

// --- History (usage timeline) ---

// Mirror backend response (см. backend/.../prompt/response.go UsageLogResponse).
// `prompt` — nested полный PromptResponse; берём отсюда title и model.
export interface UsageHistoryItem {
  id: number;
  prompt_id: number;
  prompt: {
    id: number;
    title: string;
    model?: string;
    tags?: Tag[];
  };
  used_at: string;
}

export interface UsageHistoryResponse {
  items: UsageHistoryItem[];
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
}

export async function listUsageHistory(
  limit = 50,
  offset = 0,
): Promise<UsageHistoryResponse> {
  // Backend pagination через page/page_size — конвертируем из limit/offset.
  const page = Math.floor(offset / limit) + 1;
  const params = new URLSearchParams({ page: String(page), page_size: String(limit) });
  return request<UsageHistoryResponse>(`/api/prompts/history?${params}`);
}

// --- Analytics ---

export type AnalyticsRange = '7d' | '30d' | '90d' | '365d';

export interface UsageByDayPoint {
  date: string;
  count: number;
}

export interface TopPromptItem {
  id: number;
  title: string;
  usage_count: number;
}

export interface PersonalAnalyticsResponse {
  range: AnalyticsRange;
  totals: {
    uses: number;
    created: number;
    share_views: number;
  };
  usage_by_day: UsageByDayPoint[];
  top_prompts: TopPromptItem[];
  model_segmentation?: { model: string; count: number }[];
}

export async function getPersonalAnalytics(
  range: AnalyticsRange = '30d',
): Promise<PersonalAnalyticsResponse> {
  return request<PersonalAnalyticsResponse>(`/api/analytics/personal?range=${range}`);
}

export interface Insight {
  type: string;
  title: string;
  description: string;
  data?: Record<string, unknown>;
}

export interface InsightsResponse {
  items: Insight[];
  generated_at: string;
}

export async function getInsights(): Promise<InsightsResponse> {
  return request<InsightsResponse>('/api/analytics/insights');
}

export async function refreshInsights(): Promise<InsightsResponse> {
  return request<InsightsResponse>('/api/analytics/insights/refresh', { method: 'POST' });
}

// --- Versions ---

export async function listVersions(
  promptId: number,
  limit = 20,
  offset = 0,
): Promise<{ items: PromptVersion[]; total: number; has_more: boolean }> {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  return request<{ items: PromptVersion[]; total: number; has_more: boolean }>(
    `/api/prompts/${promptId}/versions?${params}`,
  );
}

export async function revertVersion(promptId: number, versionId: number): Promise<Prompt> {
  return request<Prompt>(`/api/prompts/${promptId}/revert/${versionId}`, {
    method: 'POST',
  });
}

// --- Trash ---

export async function listTrash(): Promise<TrashListResponse> {
  return request<TrashListResponse>('/api/trash');
}

// Backend ожидает singular ItemType — "prompt" / "collection" (см.
// backend/internal/usecases/trash/types.go). Plural форма даёт 400.
export async function restoreTrashPrompt(id: number): Promise<Prompt> {
  return request<Prompt>(`/api/trash/prompt/${id}/restore`, { method: 'POST' });
}

export async function restoreTrashCollection(id: number): Promise<Collection> {
  return request<Collection>(`/api/trash/collection/${id}/restore`, { method: 'POST' });
}

export async function permanentDeleteTrashPrompt(id: number): Promise<void> {
  await request<void>(`/api/trash/prompt/${id}`, { method: 'DELETE' });
}

export async function permanentDeleteTrashCollection(id: number): Promise<void> {
  await request<void>(`/api/trash/collection/${id}`, { method: 'DELETE' });
}

export async function emptyTrash(): Promise<void> {
  await request<void>('/api/trash', { method: 'DELETE' });
}

// --- Share Links ---

export async function getShareLink(promptId: number): Promise<ShareLink | null> {
  try {
    return await request<ShareLink>(`/api/prompts/${promptId}/share`);
  } catch (err) {
    if (err instanceof ApiError && err.code === 'not_found') return null;
    throw err;
  }
}

export async function createShareLink(promptId: number): Promise<ShareLink> {
  return request<ShareLink>(`/api/prompts/${promptId}/share`, { method: 'POST' });
}

export async function deactivateShareLink(promptId: number): Promise<void> {
  await request<void>(`/api/prompts/${promptId}/share`, { method: 'DELETE' });
}

// --- Collections CRUD (Phase 2) ---

export async function listCollections(teamId?: number | null): Promise<CollectionDTO[]> {
  const params = new URLSearchParams();
  if (teamId) params.set('team_id', String(teamId));
  const suffix = params.toString() ? `?${params}` : '';
  const data = await request<CollectionDTO[] | { items: CollectionDTO[] }>(
    `/api/collections${suffix}`,
  );
  return Array.isArray(data) ? data : (data.items ?? []);
}

export interface CreateCollectionBody {
  name: string;
  description?: string;
  color: string;
  icon: string;
  team_id?: number | null;
}

export async function createCollection(body: CreateCollectionBody): Promise<Collection> {
  return request<Collection>('/api/collections', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function updateCollection(
  id: number,
  body: Partial<CreateCollectionBody>,
): Promise<Collection> {
  return request<Collection>(`/api/collections/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function deleteCollection(id: number): Promise<void> {
  await request<void>(`/api/collections/${id}`, { method: 'DELETE' });
}

// --- Tags CRUD ---

export async function listTags(teamId?: number | null): Promise<TagDTO[]> {
  const params = new URLSearchParams();
  if (teamId) params.set('team_id', String(teamId));
  const suffix = params.toString() ? `?${params}` : '';
  const data = await request<TagDTO[] | { items: TagDTO[] }>(`/api/tags${suffix}`);
  return Array.isArray(data) ? data : (data.items ?? []);
}

export interface CreateTagBody {
  name: string;
  color: string;
  team_id?: number | null;
}

export async function createTag(body: CreateTagBody): Promise<Tag> {
  return request<Tag>('/api/tags', { method: 'POST', body: JSON.stringify(body) });
}

export async function deleteTag(id: number): Promise<void> {
  await request<void>(`/api/tags/${id}`, { method: 'DELETE' });
}

// --- Teams ---

export async function listTeams(): Promise<TeamDTO[]> {
  const data = await request<TeamDTO[] | { items: TeamDTO[] }>('/api/teams');
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function getTeam(slug: string): Promise<TeamDetail> {
  return request<TeamDetail>(`/api/teams/${slug}`);
}

export interface CreateTeamBody {
  name: string;
  description?: string;
}

export async function createTeam(body: CreateTeamBody): Promise<Team> {
  return request<Team>('/api/teams', { method: 'POST', body: JSON.stringify(body) });
}

export async function updateTeam(slug: string, body: Partial<CreateTeamBody>): Promise<Team> {
  return request<Team>(`/api/teams/${slug}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function deleteTeam(slug: string): Promise<void> {
  await request<void>(`/api/teams/${slug}`, { method: 'DELETE' });
}

export async function inviteTeamMember(
  slug: string,
  body: { email: string; role: 'editor' | 'viewer' },
): Promise<void> {
  await request<void>(`/api/teams/${slug}/invitations`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function removeTeamMember(slug: string, memberId: number): Promise<void> {
  await request<void>(`/api/teams/${slug}/members/${memberId}`, { method: 'DELETE' });
}

export async function updateTeamMemberRole(
  slug: string,
  memberId: number,
  role: 'owner' | 'editor' | 'viewer',
): Promise<void> {
  await request<void>(`/api/teams/${slug}/members/${memberId}`, {
    method: 'PUT',
    body: JSON.stringify({ role }),
  });
}

// --- Streak / Stats ---

export async function getStreak(): Promise<StreakDTO> {
  return request<StreakDTO>('/api/streaks');
}

export async function getStreakDetail(): Promise<StreakResponse> {
  return request<StreakResponse>('/api/streaks');
}

// --- Badges ---

export async function listBadges(): Promise<BadgeListResponse> {
  return request<BadgeListResponse>('/api/badges');
}

// --- Changelog ---

export async function getChangelog(): Promise<ChangelogResponse> {
  return request<ChangelogResponse>('/api/changelog');
}

export async function markChangelogRead(): Promise<void> {
  await request<void>('/api/changelog/seen', { method: 'POST' });
}

// --- Subscription ---

export async function listPlans(): Promise<Plan[]> {
  const data = await request<Plan[] | { items: Plan[] }>('/api/plans');
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function getCurrentSubscription(): Promise<Subscription | null> {
  try {
    return await request<Subscription>('/api/subscription');
  } catch (err) {
    if (err instanceof ApiError && err.code === 'not_found') return null;
    throw err;
  }
}

export async function getUsageSummary(): Promise<UsageSummary> {
  return request<UsageSummary>('/api/subscription/usage');
}

export async function cancelSubscription(): Promise<void> {
  await request<void>('/api/subscription/cancel', { method: 'POST' });
}

// Backend требует body { months: 1|2|3 } — default 1.
export async function pauseSubscription(months = 1): Promise<void> {
  await request<void>('/api/subscription/pause', {
    method: 'POST',
    body: JSON.stringify({ months }),
  });
}

export async function resumeSubscription(): Promise<void> {
  await request<void>('/api/subscription/resume', { method: 'POST' });
}

// --- Use / favorite / pin ---

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

// --- API Keys ---

export async function listApiKeys(): Promise<APIKeyListResponse> {
  return request<APIKeyListResponse>('/api/api-keys');
}

export async function createApiKey(body: CreateAPIKeyRequest): Promise<CreatedAPIKey> {
  return request<CreatedAPIKey>('/api/api-keys', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function deleteApiKey(id: number): Promise<void> {
  await request<void>(`/api/api-keys/${id}`, { method: 'DELETE' });
}

// --- Chains (Phase 3) ---

export async function listChains(teamId?: number | null): Promise<ChainListResponse> {
  const params = new URLSearchParams();
  if (teamId) params.set('team_id', String(teamId));
  const suffix = params.toString() ? `?${params}` : '';
  return request<ChainListResponse>(`/api/chains${suffix}`);
}

export async function getChain(id: number): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${id}`);
}

export async function startChainExecution(
  chainId: number,
  initialVars: Record<string, string> = {},
): Promise<ChainExecution> {
  return request<ChainExecution>(`/api/chains/${chainId}/executions`, {
    method: 'POST',
    body: JSON.stringify({ initial_vars: initialVars }),
  });
}

export async function getExecution(execId: number): Promise<ChainExecution> {
  return request<ChainExecution>(`/api/executions/${execId}`);
}

export async function advanceChainStep(
  execId: number,
  stepOutput: string,
  chosenBranchIndex?: number,
): Promise<ChainExecution> {
  const body: Record<string, unknown> = { step_output: stepOutput };
  if (chosenBranchIndex !== undefined) body.chosen_branch_index = chosenBranchIndex;
  return request<ChainExecution>(`/api/executions/${execId}/advance`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function listExecutions(chainId: number): Promise<ChainExecutionListResponse> {
  return request<ChainExecutionListResponse>(`/api/chains/${chainId}/executions`);
}

// --- Chain CRUD ---

export interface CreateChainBody {
  name: string;
  description?: string;
  team_id?: number | null;
}

export async function createChain(body: CreateChainBody): Promise<ChainDetail> {
  return request<ChainDetail>('/api/chains', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function updateChain(
  id: number,
  body: { name?: string; description?: string },
): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function deleteChain(id: number): Promise<void> {
  await request<void>(`/api/chains/${id}`, { method: 'DELETE' });
}

// --- Chain Steps ---

export interface AddStepBody {
  prompt_id?: number | null;
  name?: string;
  variable_mapping?: Record<string, { type: string; var_name?: string }>;
  manual_checkpoint?: boolean;
  step_type?: 'prompt' | 'fork';
  conditions?: unknown;
  after_step_id?: number;
  parent_fork_id?: number;
  branch_index?: number;
}

export async function addChainStep(chainId: number, body: AddStepBody): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${chainId}/steps`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export interface UpdateStepBody {
  name?: string;
  variable_mapping?: Record<string, { type: string; var_name?: string }>;
  manual_checkpoint?: boolean;
  step_type?: 'prompt' | 'fork';
  conditions?: unknown;
}

export async function updateChainStep(
  chainId: number,
  stepId: number,
  body: UpdateStepBody,
): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${chainId}/steps/${stepId}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function removeChainStep(chainId: number, stepId: number): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${chainId}/steps/${stepId}`, {
    method: 'DELETE',
  });
}

export async function moveStepUp(chainId: number, stepId: number): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${chainId}/steps/${stepId}/move-up`, {
    method: 'POST',
  });
}

export async function moveStepDown(chainId: number, stepId: number): Promise<ChainDetail> {
  return request<ChainDetail>(`/api/chains/${chainId}/steps/${stepId}/move-down`, {
    method: 'POST',
  });
}

// --- Phase 6: Email/password auth ---

// Phase 6 — после успешного login/register сразу создаём API-key
// "Chrome Extension" через access_token и работаем дальше как с pvlt_* ключом.
// Refresh_token через HttpOnly cookie в extension не работает, а access живёт
// только 15 минут — это решает проблему долгоживущей сессии.

export interface AuthLoginResponse {
  apiKey: string;
  user: MeResponse;
}

async function unauthFetch(
  apiBase: string,
  path: string,
  body: unknown,
): Promise<Response> {
  const headers = new Headers({
    'Content-Type': 'application/json',
    'X-Client': `chrome-extension/${EXTENSION_VERSION}`,
  });
  return fetch(`${apiBase}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });
}

async function createKeyWithAccessToken(
  apiBase: string,
  accessToken: string,
  name: string,
): Promise<string> {
  const res = await fetch(`${apiBase}/api/api-keys`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Client': `chrome-extension/${EXTENSION_VERSION}`,
      'Authorization': `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ name, read_only: false }),
  });
  if (!res.ok) {
    throw new ApiError(`не удалось создать API-key: ${res.status}`, res.status, 'unknown');
  }
  const data = (await res.json()) as { key: string };
  return data.key;
}

export async function loginEmailPassword(
  email: string,
  password: string,
): Promise<AuthLoginResponse> {
  const { apiBase } = await getSettings();
  const res = await unauthFetch(apiBase, '/api/auth/login', { email, password });
  if (!res.ok) {
    if (res.status === 401) throw new ApiError('Неверный email или пароль', 401, 'unauthorized');
    if (res.status === 403) {
      const body = await safeJson(res);
      throw new ApiError(body?.message ?? 'Доступ заблокирован', 403, 'forbidden');
    }
    if (res.status === 429) throw new ApiError('Слишком много попыток', 429, 'rate_limited');
    throw new ApiError(`Ошибка входа: ${res.status}`, res.status, 'network');
  }
  const data = (await res.json()) as { user: MeResponse; tokens: { access_token: string } };
  const apiKey = await createKeyWithAccessToken(apiBase, data.tokens.access_token, 'Chrome Extension');
  return { apiKey, user: data.user };
}

export async function registerEmailPassword(body: {
  email: string;
  password: string;
  name: string;
  referredBy?: string;
}): Promise<AuthLoginResponse> {
  const { apiBase } = await getSettings();
  const res = await unauthFetch(apiBase, '/api/auth/register', {
    email: body.email,
    password: body.password,
    name: body.name,
    referred_by: body.referredBy,
  });
  if (!res.ok) {
    if (res.status === 409) throw new ApiError('Email уже зарегистрирован', 409, 'conflict');
    if (res.status === 422) {
      const data = await safeJson(res);
      throw new ApiError(data?.message ?? 'Ошибка валидации', 422, 'validation');
    }
    throw new ApiError(`Ошибка регистрации: ${res.status}`, res.status, 'network');
  }
  const data = (await res.json()) as { user: MeResponse; tokens: { access_token: string } };
  const apiKey = await createKeyWithAccessToken(apiBase, data.tokens.access_token, 'Chrome Extension');
  return { apiKey, user: data.user };
}

export async function forgotPassword(email: string): Promise<void> {
  const { apiBase } = await getSettings();
  const res = await unauthFetch(apiBase, '/api/auth/forgot-password', { email });
  if (!res.ok) {
    if (res.status === 429) throw new ApiError('Слишком много попыток', 429, 'rate_limited');
    // Backend для безопасности всегда возвращает 200, чтобы не раскрывать
    // существование аккаунта. Если получили не-200 — fail-safe ошибка.
    throw new ApiError(`Ошибка: ${res.status}`, res.status, 'network');
  }
}

export async function resetPassword(body: {
  email: string;
  code: string;
  newPassword: string;
}): Promise<void> {
  const { apiBase } = await getSettings();
  const res = await unauthFetch(apiBase, '/api/auth/reset-password', {
    email: body.email,
    code: body.code,
    new_password: body.newPassword,
  });
  if (!res.ok) {
    if (res.status === 400 || res.status === 422) {
      const data = await safeJson(res);
      throw new ApiError(data?.message ?? 'Неверный код или email', res.status, 'validation');
    }
    if (res.status === 410) throw new ApiError('Код истёк', 410, 'validation');
    throw new ApiError(`Ошибка: ${res.status}`, res.status, 'network');
  }
}

// --- Team Invitations ---

export async function listMyInvitations(): Promise<TeamInvitation[]> {
  const data = await request<TeamInvitation[] | { items: TeamInvitation[] }>(
    '/api/invitations',
  );
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function acceptInvitation(invitationId: number): Promise<void> {
  await request<void>(`/api/invitations/${invitationId}/accept`, { method: 'POST' });
}

export async function declineInvitation(invitationId: number): Promise<void> {
  await request<void>(`/api/invitations/${invitationId}/decline`, { method: 'POST' });
}

// --- Feedback ---

export async function submitFeedback(body: FeedbackRequest): Promise<FeedbackResponse> {
  return request<FeedbackResponse>('/api/feedback', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// --- Notifications settings ---

export async function setInsightEmails(enabled: boolean): Promise<{ insight_emails_enabled: boolean }> {
  return request<{ insight_emails_enabled: boolean }>(
    '/api/auth/notifications/insights',
    {
      method: 'PATCH',
      body: JSON.stringify({ enabled }),
    },
  );
}

// --- Linked accounts ---

export interface LinkedAccountDTO {
  id: number;
  provider: string;
}

export async function listLinkedAccounts(): Promise<LinkedAccountDTO[]> {
  const data = await request<LinkedAccountDTO[] | { items: LinkedAccountDTO[] }>(
    '/api/auth/linked-accounts',
  );
  return Array.isArray(data) ? data : (data.items ?? []);
}

export async function unlinkProvider(provider: string): Promise<void> {
  await request<void>(`/api/auth/unlink/${provider}`, { method: 'DELETE' });
}

// --- Referral ---

export interface ReferralInfo {
  code: string;
  invited_count: number;
  referred_by?: string;
  reward_granted: boolean;
}

export async function getReferral(): Promise<ReferralInfo> {
  return request<ReferralInfo>('/api/auth/referral');
}

// --- Team Branding ---

export interface TeamBranding {
  logo_url?: string;
  logo_source?: string;
  effective_logo_url?: string;
  tagline?: string;
  website?: string;
  primary_color?: string;
}

export interface UpdateBrandingBody {
  tagline?: string;
  website?: string;
  primary_color?: string;
  logo_url?: string;
}

export async function getTeamBranding(slug: string): Promise<TeamBranding> {
  return request<TeamBranding>(`/api/teams/${slug}/branding`);
}

export async function updateTeamBranding(
  slug: string,
  body: UpdateBrandingBody,
): Promise<TeamBranding> {
  return request<TeamBranding>(`/api/teams/${slug}/branding`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export async function deleteTeamLogo(slug: string): Promise<void> {
  await request<void>(`/api/teams/${slug}/branding/logo`, { method: 'DELETE' });
}

// uploadTeamLogo — multipart upload. Делается напрямую из side-panel (не через
// background switch), потому что FormData не сериализуется через chrome.runtime.sendMessage.
export async function uploadTeamLogoDirect(
  slug: string,
  file: File,
): Promise<TeamBranding> {
  const { apiKey, apiBase } = await getSettings();
  if (!apiKey) throw new ApiError('missing api key', 401, 'unauthorized');

  const formData = new FormData();
  formData.append('logo', file);

  const headers = new Headers();
  headers.set('Authorization', `Bearer ${apiKey}`);
  headers.set('X-Client', `chrome-extension/${EXTENSION_VERSION}`);

  const res = await fetch(`${apiBase}/api/teams/${slug}/branding/logo`, {
    method: 'POST',
    headers,
    body: formData,
  });

  if (!res.ok) {
    if (res.status === 401) throw new ApiError('unauthorized', 401, 'unauthorized');
    if (res.status === 402) {
      const body = await safeJson(res);
      throw new ApiError(body?.message ?? 'upgrade required', 402, 'quota_exceeded');
    }
    if (res.status === 413) throw new ApiError('file too large', 413, 'validation');
    if (res.status === 415) throw new ApiError('unsupported format', 415, 'validation');
    throw new ApiError(`upload failed: ${res.status}`, res.status, 'network');
  }

  return (await res.json()) as TeamBranding;
}

// --- Team Analytics ---

export interface AnalyticsUsagePoint {
  day: string;
  count: number;
}

export interface AnalyticsTopPrompt {
  prompt_id: number;
  title: string;
  uses: number;
}

export interface AnalyticsContributor {
  user_id: number;
  email: string;
  name?: string;
  prompts_created: number;
  prompts_edited: number;
  uses: number;
}

export interface AnalyticsModelUsage {
  model: string;
  uses: number;
}

export interface AnalyticsTotals {
  uses: number;
  created: number;
  updated: number;
  share_views: number;
}

export interface TeamDashboardResponse {
  range: AnalyticsRange;
  usage_per_day: AnalyticsUsagePoint[];
  top_prompts: AnalyticsTopPrompt[];
  prompts_created_per_day: AnalyticsUsagePoint[];
  prompts_updated_per_day: AnalyticsUsagePoint[];
  contributors: AnalyticsContributor[];
  totals_current: AnalyticsTotals;
  totals_previous: AnalyticsTotals;
  usage_by_model: AnalyticsModelUsage[];
}

export async function getTeamAnalytics(
  teamId: number,
  range: AnalyticsRange = '30d',
): Promise<TeamDashboardResponse> {
  return request<TeamDashboardResponse>(
    `/api/analytics/teams/${teamId}?range=${range}`,
  );
}

// --- Team Activity ---

export interface ActivityItem {
  id: number;
  actor_id?: number;
  actor_email?: string;
  actor_name?: string;
  event_type: string;
  target_type: string;
  target_id?: number;
  target_label?: string;
  metadata?: unknown;
  created_at: string;
}

export interface ActivityResponse {
  items: ActivityItem[];
  page: number;
  page_size: number;
  has_more: boolean;
}

export async function getTeamActivity(
  slug: string,
  page = 1,
  pageSize = 50,
): Promise<ActivityResponse> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  return request<ActivityResponse>(`/api/teams/${slug}/activity?${params}`);
}

// --- API Keys are dual-purpose — Type-only re-exports kept above ---
// (chains module section ends here)

// Re-export для backward-compat existing callers.
export {
  type APIKey,
  type APIKeyListResponse,
  type CreatedAPIKey,
  type CreateAPIKeyRequest,
} from './types';
