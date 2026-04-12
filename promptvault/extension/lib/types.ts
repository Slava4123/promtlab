// Типы API PromptVault, минимальный набор нужный extension.
// Точные shape-определения — см. backend/internal/delivery/http/prompt/response.go

export interface TagDTO {
  id: number;
  name: string;
  color: string;
}

export interface CollectionDTO {
  id: number;
  name: string;
  color?: string;
  icon?: string;
  prompts_count?: number;
}

export interface TeamDTO {
  id: number;
  slug: string;
  name: string;
  description?: string;
  role?: string;
}

export interface StreakDTO {
  current_streak: number;
  longest_streak: number;
  last_activity_at?: string | null;
}

export interface Prompt {
  id: number;
  title: string;
  content: string;
  model?: string;
  favorite: boolean;
  pinned_personal: boolean;
  pinned_team: boolean;
  usage_count: number;
  last_used_at?: string | null;
  tags: TagDTO[];
  collections: CollectionDTO[];
  created_at: string;
  updated_at: string;
}

export interface PaginatedPrompts {
  items: Prompt[];
  total: number;
  page?: number;
  page_size?: number;
  has_more?: boolean;
}

export interface SearchResult {
  prompts: Array<{ id: number; type: 'prompt'; title: string; description: string }>;
  collections: Array<{ id: number; type: 'collection'; title: string; color?: string; icon?: string }>;
  tags: Array<{ id: number; type: 'tag'; title: string; color?: string }>;
}

export interface MeResponse {
  id: number;
  email: string;
  username?: string;
  name?: string;
  role?: string;
}

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}
