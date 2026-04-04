export interface User {
  id: number
  email: string
  name: string
  username?: string
  avatar_url?: string
  email_verified: boolean
  has_password: boolean
  default_model: string
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
  usage_count: number
  tags: Tag[]
  collections: Collection[]
  created_at: string
  updated_at: string
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

export interface PromptVersion {
  id: number
  version_number: number
  title: string
  content: string
  model?: string
  change_note?: string
  created_at: string
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

// AI (Phase 8)

export type AIAction = "enhance" | "rewrite" | "analyze" | "variations"

export interface AIModel {
  id: string
  name: string
  provider: string
  description: string
  max_tokens: number
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
