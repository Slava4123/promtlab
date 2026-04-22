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
  max_share_links: number
  /** Phase 14: Fixed-window лимит создаваемых share-ссылок в день (Free 10 / Pro 100 / Max 1000). */
  max_daily_shares: number
  max_ext_uses_daily: number
  max_mcp_uses_daily: number
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
  share_links: QuotaInfo
  /** Phase 14: fixed-window лимит создаваемых share-ссылок за сутки (UTC-полночь). */
  daily_shares_today: QuotaInfo
  ext_uses_today: QuotaInfo
  mcp_uses_today: QuotaInfo
}

export interface CheckoutResponse {
  payment_url: string
}
