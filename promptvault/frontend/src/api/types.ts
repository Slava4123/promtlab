export type UserRole = "user" | "admin"
export type UserStatus = "active" | "frozen"

export interface User {
  id: number
  email: string
  name: string
  username?: string
  avatar_url?: string
  email_verified: boolean
  has_password: boolean
  plan_id?: PlanID
  role?: UserRole
  status?: UserStatus
  onboarding_completed_at?: string | null
  has_unread_changelog?: boolean
  // Phase 14 M-10: opt-in email digest по Smart Insights.
  insight_emails_enabled?: boolean
}

export interface Tag {
  id: number
  name: string
  color: string
}

export interface Collection {
  id: number
  name: string
  description: string
  color: string
  icon: string
  prompt_count: number
  created_at: string
  updated_at: string
}

export interface Prompt {
  id: number
  title: string
  content: string
  model?: string
  favorite: boolean
  pinned_personal: boolean
  pinned_team: boolean
  usage_count: number
  last_used_at?: string
  tags: Tag[]
  collections: Collection[]
  created_at: string
  updated_at: string
  is_public: boolean
  slug?: string
  // newly_unlocked_badges заполняется только в ответах на mutating endpoints
  // (POST/PUT/POST use). Отсутствует в GET responses (omitempty на backend).
  newly_unlocked_badges?: BadgeSummary[]
}

export interface CollectionResponse extends Collection {
  newly_unlocked_badges?: BadgeSummary[]
}

export interface IncrementUsageResponse {
  message: string
  newly_unlocked_badges?: BadgeSummary[]
}

export interface PinResult {
  pinned: boolean
  team_wide: boolean
}

export interface UsageLogEntry {
  id: number
  prompt_id: number
  prompt: Prompt
  used_at: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

export interface TokenPair {
  access_token: string
  expires_in: number
}

export interface AuthResponse {
  user: User
  tokens: TokenPair
}

// Admin login step (когда backend решает, что нужен TOTP).
export interface AdminLoginStepResponse {
  // Flow 1: у admin confirmed TOTP — UI показывает TOTP input.
  totp_required?: boolean
  pre_auth_token?: string

  // Flow 2: admin без TOTP enrollment — UI ведёт на /admin/totp wizard.
  // AccessToken отдан, юзер залогинен.
  totp_enrollment_required?: boolean
  access_token?: string
  expires_in?: number

  user: User
}

export interface VerifyTOTPResponse {
  access_token: string
  expires_in: number
  user: User
  used_backup_code: boolean
  remaining_backup_codes: number
}

// Admin TOTP management responses (для /admin/totp/* endpoints).
export interface TOTPEnrollResponse {
  secret: string
  qr_url: string
  backup_codes: string[]
}

export interface TOTPStatusResponse {
  enrolled: boolean
  confirmed: boolean
}

export interface PromptVersion {
  id: number
  version_number: number
  title: string
  content: string
  model?: string
  change_note?: string
  created_at: string
  /** Phase 14: автор версии. Для записей до миграции 000039 поля могут быть пустыми. */
  changed_by_id?: number
  changed_by_email?: string
  changed_by_name?: string
}

// Settings (Phase 10)

export interface LinkedAccount {
  id: number
  provider: string
}

export interface UpdateProfileRequest {
  name: string
  username?: string
  avatar_url?: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

// Search (Phase 11)

export interface SearchResult {
  id: number
  type: "prompt" | "collection" | "tag"
  title: string
  description?: string
  color?: string
  icon?: string
}

export interface SearchResponse {
  prompts: SearchResult[]
  collections: SearchResult[]
  tags: SearchResult[]
}

// Streaks

export interface StreakResponse {
  current_streak: number
  longest_streak: number
  last_active_date: string
  active_today: boolean
}

// Search Suggest

export interface Suggestion {
  text: string
  type: "prompt" | "collection" | "tag"
}

export interface SuggestResponse {
  suggestions: Suggestion[]
}

// Feedback

export interface FeedbackRequest {
  type: "bug" | "feature" | "other"
  message: string
  page_url?: string
}

export interface FeedbackResponse {
  id: number
  message: string
}

// Changelog

export interface ChangelogEntry {
  version: string
  date: string
  title: string
  category: string
  description: string
}

export interface ChangelogResponse {
  entries: ChangelogEntry[]
  has_unread: boolean
}

// Teams (Phase 9)

export type TeamRole = "owner" | "editor" | "viewer"

export interface Team {
  id: number
  slug: string
  name: string
  description?: string
  role: TeamRole
  member_count: number
  created_at: string
  updated_at: string
}

export interface TeamDetail extends Team {
  members: TeamMember[]
}

export interface TeamMember {
  user_id: number
  name: string
  username?: string
  email: string
  avatar_url?: string
  role: TeamRole
}

export interface TeamInvitation {
  id: number
  team_id: number
  team_name: string
  team_slug: string
  role: TeamRole
  inviter_name: string
  status: "pending" | "accepted" | "declined"
  created_at: string
}

export interface PendingInvitation {
  id: number
  email: string
  name: string
  username?: string
  role: TeamRole
  status: "pending" | "accepted" | "declined"
}

export interface UserSearchResult {
  id: number
  name: string
  username: string
  avatar_url?: string
  email: string
}

// Starter templates (onboarding wizard)

export interface StarterCategory {
  id: string
  name: string
  description: string
  icon: string
  use_cases: string[]
}

export interface StarterTemplate {
  id: string
  category: string
  title: string
  content: string
  model: string
}

export interface StarterCatalog {
  version: number
  lang: string
  categories: StarterCategory[]
  templates: StarterTemplate[]
}

export interface CompleteOnboardingRequest {
  install: string[]
}

export interface CompleteOnboardingResponse {
  installed: Prompt[]
  onboarding_completed_at: string
}

// Trash

export interface TrashPrompt {
  id: number
  title: string
  content: string
  model?: string
  favorite: boolean
  tags: Tag[]
  collections: { id: number; name: string; color: string }[]
  deleted_at: string
  created_at: string
  days_left: number
}

export interface TrashCollection {
  id: number
  name: string
  description?: string
  color: string
  icon?: string
  deleted_at: string
  days_left: number
}

export interface TrashCounts {
  prompts: number
  collections: number
  tags: number
}

export interface TrashListResponse {
  prompts: PaginatedResponse<TrashPrompt>
  collections: TrashCollection[]
}

// API Keys
export interface APIKey {
  id: number
  name: string
  key_prefix: string
  last_used_at?: string | null
  created_at: string
  read_only: boolean
  team_id?: number | null
  allowed_tools?: string[] | null
  expires_at?: string | null
}

export interface APIKeyListResponse {
  keys: APIKey[]
  max_keys: number
}

export interface CreateAPIKeyRequest {
  name: string
  read_only?: boolean
  team_id?: number | null
  allowed_tools?: string[]
  expires_at?: string | null
}

// Share Links

export interface ShareLink {
  id: number
  token: string
  url: string
  is_active: boolean
  view_count: number
  last_viewed_at?: string | null
  created_at: string
}

export interface PublicPrompt {
  title: string
  content: string
  model?: string
  tags: { name: string; color: string }[]
  author: { name: string; avatar_url?: string }
  created_at: string
  updated_at: string
  /** Phase 14 D: Branded share pages (Max-only). undefined для Free/Pro и не-team промптов. */
  branding?: {
    logo_url?: string
    /** Phase 16-X: 'url' | 'file' | 'none'. Резолвинг между внешним URL и uploaded-file. */
    logo_source?: "url" | "file" | "none"
    /** Phase 16-X: готовый src для <img> (бэк уже выбрал между logo_url и /api/.../branding/logo). */
    effective_logo_url?: string
    tagline?: string
    website?: string
    primary_color?: string
  }
}

export interface CreatedAPIKey {
  id: number
  name: string
  key: string
  key_prefix: string
  created_at: string
  read_only: boolean
  team_id?: number | null
  allowed_tools?: string[] | null
  expires_at?: string | null
}

// Badges

export type BadgeCategory = "personal" | "team" | "milestone" | "streak"

export interface BadgeSummary {
  id: string
  title: string
  icon: string
}

export interface Badge {
  id: string
  title: string
  description: string
  icon: string
  category: BadgeCategory
  unlocked: boolean
  unlocked_at?: string
  progress: number
  target: number
}

export interface BadgeListResponse {
  items: Badge[]
  total_count: number
  total_unlocked: number
}

// Subscription / Billing

export type PlanID = "free" | "pro" | "max" | "pro_yearly" | "max_yearly"

// MN-60: runtime guard вместо `as PlanID` cast. Сейчас юзер из БД может прийти
// с любым plan_id (например, после rollback миграции, ручной правки строки или
// future tier — `enterprise`). Без этого guard'а компонент молча сломал бы
// rendering, потому что PlanBadge не знает unknown plan.
const VALID_PLAN_IDS: ReadonlySet<string> = new Set([
  "free",
  "pro",
  "max",
  "pro_yearly",
  "max_yearly",
])

export function isPlanID(v: unknown): v is PlanID {
  return typeof v === "string" && VALID_PLAN_IDS.has(v)
}

// asPlanID — runtime-проверенная замена `as PlanID`. fallback="free" даёт
// безопасный дефолт, если значение из API/store не валидно (free всегда
// поддерживается).
export function asPlanID(v: unknown, fallback: PlanID = "free"): PlanID {
  return isPlanID(v) ? v : fallback
}
export type SubscriptionStatus = "active" | "past_due" | "paused" | "cancelled" | "expired"

export type CancelReason =
  | "too_expensive"
  | "not_using"
  | "missing_feature"
  | "found_alternative"
  | "other"

export interface Plan {
  id: PlanID
  name: string
  price_kop: number
  period_days: number
  max_prompts: number
  max_collections: number
  max_teams: number
  max_team_members: number
  // Phase 16-Y: max_share_links / max_daily_shares УДАЛЕНЫ. Share-ссылки
  // живут по TTL (default 30 дней, до 1 года; Max — «без срока»). Анти-абуз
  // покрывает общий per-user rate-limit на /api.
  max_ext_uses_daily: number
  max_mcp_uses_daily: number
  /** Phase 16: tier-лимиты для Prompt Chains. Free 1/3/0, Pro 5/10/10, Max 100/50/1000. */
  max_chains: number
  max_steps_per_chain: number
  max_saved_executions: number
  features: string[]
  sort_order: number
  is_active: boolean
}

export interface Subscription {
  id: number
  plan_id: PlanID
  status: SubscriptionStatus
  current_period_start: string
  current_period_end: string
  cancel_at_period_end: boolean
  cancelled_at?: string
  auto_renew: boolean
  paused_at?: string
  paused_until?: string
  plan: Plan
}

export interface QuotaInfo {
  used: number
  limit: number
}

export interface UsageSummary {
  plan_id: PlanID
  prompts: QuotaInfo
  collections: QuotaInfo
  teams: QuotaInfo
  // Phase 16-Y: share_links и daily_shares_today УДАЛЕНЫ — на share-ссылки
  // больше нет квот, только TTL.
  ext_uses_today: QuotaInfo
  mcp_uses_today: QuotaInfo
  /** Phase 16: общее количество цепочек (не soft-deleted). */
  chains: QuotaInfo
}

export interface CheckoutResponse {
  payment_url: string
}

// --- Phase 16: Prompt Chains ---

/** Источник значения переменной шага. var_name — для type=chain_var. */
export interface VariableSource {
  type: "manual" | "chain_var"
  var_name?: string
}

export type VariableMapping = Record<string, VariableSource>

export interface Chain {
  id: number
  user_id: number
  team_id?: number
  name: string
  description: string
  created_at: string
  updated_at: string
  // Phase 16 UI polish — расширенный list-эндпойнт /api/chains
  // (отсутствуют в одиночных GET /api/chains/{id} ответах).
  step_count?: number
  has_branching?: boolean
  saved_runs_count?: number
  steps_preview?: ChainStepPreview[]
}

/** Облегчённое представление шага для рендера mini-graph. Только position и step_type. */
export interface ChainStepPreview {
  position: number
  step_type: ChainStepType
}

// Phase 16 v2 (Tree-canvas): manual fork с label-based branches.
// Юзер сам выбирает ветку в run-mode по её Label.
export type ChainStepType = "prompt" | "fork"

export interface ConditionBranch {
  /** Отображается юзеру как кнопка выбора пути в run-mode. ≤200 символов, уникален в шаге. */
  label: string
  /** id шага, на который перейти при выборе этой ветки. null = конец цепочки. */
  next_step_id?: number | null
}

export interface ChainConditions {
  branches: ConditionBranch[]
}

/** Облегчённое представление промпта внутри шага цепочки. */
export interface ChainStepPromptSummary {
  id: number
  title: string
  content: string
}

export interface ChainStep {
  id: number
  chain_id: number
  position: number
  /** prompt_id обязателен у prompt-шагов; у fork-шагов отсутствует (контейнер с ветками). */
  prompt_id?: number | null
  name: string
  /** JSONB raw: парсится в VariableMapping на use-site через JSON.parse если string. */
  variable_mapping: VariableMapping
  manual_checkpoint: boolean
  /** Phase 16 v2: 'prompt' | 'fork'. По умолчанию 'prompt'. */
  step_type: ChainStepType
  /** Только при step_type='fork'. */
  conditions?: ChainConditions
  /** Phase 16 v3: явный переход для prompt-шагов. null/undefined = конец ветки/цепочки.
   *  Игнорируется для fork-шагов — у них переход через conditions.branches[chosen].next_step_id. */
  next_step_id?: number | null
  /** Preloaded prompt для отображения title в Canvas + content в hover-preview.
   *  Может быть undefined если promпt soft-deleted. */
  prompt?: ChainStepPromptSummary
  created_at: string
}

export interface ChainDetail extends Chain {
  steps: ChainStep[]
}

export interface ChainListResponse {
  items: Chain[]
  total: number
  limit: number
  offset: number
}

/** Снимок цепочки + контент промптов на момент StartExecution.
 *  Используется фронтом для рендеринга текущего шага без обращения к /prompts. */
export interface ChainSnapshot {
  chain: Chain
  steps: ChainStep[]
  prompt_contents: Record<number, string>
}

export type ChainExecutionStatus = "in_progress" | "completed" | "abandoned"

/** Компактная сводка execution для страницы истории — без больших JSONB-полей. */
export interface ChainExecutionSummary {
  id: number
  chain_id: number
  user_id: number
  current_step: number
  status: ChainExecutionStatus
  started_at: string
  completed_at?: string | null
  updated_at: string
}

export interface ChainExecutionListResponse {
  items: ChainExecutionSummary[]
  limit: number
}

export interface ChainExecution {
  id: number
  chain_id: number
  user_id: number
  current_step: number
  /** JSONB: значения chain-level переменных (initial_vars). */
  variables: Record<string, string>
  /** JSONB: накопленные outputs шагов. Ключ "step_<id>". */
  step_outputs: Record<string, string>
  /** JSONB: ChainSnapshot. На фронте парсится через JSON.parse при необходимости. */
  chain_snapshot: ChainSnapshot
  status: ChainExecutionStatus
  started_at: string
  completed_at?: string | null
  updated_at: string
}
