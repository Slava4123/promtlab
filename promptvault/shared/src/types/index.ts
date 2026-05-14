// Shared TypeScript types for PromptVault — single source of truth для
// frontend и extension. Source mirror: backend/internal/delivery/http/*/response.go.

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
  insight_emails_enabled?: boolean
  has_legacy_quotas?: boolean
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

export interface AdminLoginStepResponse {
  totp_required?: boolean
  pre_auth_token?: string
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
  changed_by_id?: number
  changed_by_email?: string
  changed_by_name?: string
}

// Settings

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

// Search

export interface SearchResultItem {
  id: number
  type: "prompt" | "collection" | "tag"
  title: string
  description?: string
  color?: string
  icon?: string
}

export interface SearchResponse {
  prompts: SearchResultItem[]
  collections: SearchResultItem[]
  tags: SearchResultItem[]
}

// Streaks

export interface StreakResponse {
  current_streak: number
  longest_streak: number
  last_active_date: string
  active_today: boolean
}

// Suggest

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

// Teams

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

// Starter templates

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
  branding?: {
    logo_url?: string
    logo_source?: "url" | "file" | "none"
    effective_logo_url?: string
    tagline?: string
    website?: string
    primary_color?: string
  }
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
  max_ext_uses_daily: number
  max_mcp_uses_daily: number
  max_chains: number
  max_steps_per_chain: number
  max_saved_executions: number
  max_team_prompts: number
  max_team_collections: number
  max_team_chains: number
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
  ext_uses_today: QuotaInfo
  mcp_uses_today: QuotaInfo
  chains: QuotaInfo
}

// Все ключи UsageSummary, указывающие на QuotaInfo. Используются обоими
// клиентами (extension и frontend) для итерации по квотам в unread-count
// и notifications. Типизированный accessor избавляет от cast'а
// `as unknown as Record<string, QuotaInfo | undefined>`.
export const QUOTA_KEYS = [
  "prompts",
  "collections",
  "teams",
  "chains",
  "ext_uses_today",
  "mcp_uses_today",
] as const
export type QuotaKey = (typeof QUOTA_KEYS)[number]

export function quotaByKey(usage: UsageSummary, key: QuotaKey): QuotaInfo {
  return usage[key]
}

export interface TeamUsageSummary {
  team_id: number
  team_name: string
  owner_plan_id: PlanID
  prompts: QuotaInfo
  collections: QuotaInfo
  chains: QuotaInfo
}

export interface CheckoutResponse {
  payment_url: string
}

// Prompt Chains (Phase 16)

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
  step_count?: number
  has_branching?: boolean
  saved_runs_count?: number
  steps_preview?: ChainStepPreview[]
}

export interface ChainStepPreview {
  position: number
  step_type: ChainStepType
}

export type ChainStepType = "prompt" | "fork"

export interface ConditionBranch {
  label: string
  next_step_id?: number | null
}

export interface ChainConditions {
  branches: ConditionBranch[]
}

export interface ChainStepPromptSummary {
  id: number
  title: string
  content: string
}

export interface ChainStep {
  id: number
  chain_id: number
  position: number
  prompt_id?: number | null
  name: string
  variable_mapping: VariableMapping
  manual_checkpoint: boolean
  step_type: ChainStepType
  conditions?: ChainConditions
  next_step_id?: number | null
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

export interface ChainSnapshot {
  chain: Chain
  steps: ChainStep[]
  prompt_contents: Record<number, string>
}

export type ChainExecutionStatus = "in_progress" | "completed" | "abandoned"

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
  variables: Record<string, string>
  step_outputs: Record<string, string>
  chain_snapshot: ChainSnapshot
  status: ChainExecutionStatus
  started_at: string
  completed_at?: string | null
  updated_at: string
}

// ApiError — runtime class, не type-only. Используется обеими сторонами.
//
// `details` хранит сырой body ответа backend'а для structured-ошибок:
//   402 quota_exceeded: { quota_type, used, limit, plan, upgrade_url }
//   422 validation: { errors: [...] }
//   и др.
// Без этого quota dialog в UI терял quota_type и показывал generic fallback,
// даже когда backend явно слал "quota_type": "prompts".

export class ApiError extends Error {
  status: number
  code?: string
  details?: Record<string, unknown>

  constructor(message: string, status: number, code?: string, details?: Record<string, unknown>) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.code = code
    this.details = details
  }
}

// Legacy aliases (для обратной совместимости с extension/lib/types.ts).

export type TagDTO = Tag
export type CollectionDTO = Pick<Collection, "id" | "name" | "color" | "icon"> & {
  prompts_count?: number
}
export type TeamDTO = Pick<Team, "id" | "slug" | "name" | "description"> & {
  role?: string
}
export interface StreakDTO {
  current_streak: number
  longest_streak: number
  last_activity_at?: string | null
}
export interface PaginatedPrompts {
  items: Prompt[]
  total: number
  page?: number
  page_size?: number
  has_more?: boolean
}
export interface MeResponse {
  id: number
  email: string
  username?: string
  name?: string
  role?: string
}
