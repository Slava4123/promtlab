// MN-66 — typed query-keys factory для TanStack Query.
//
// Раньше каждый хук писал raw `["prompts", id]` — опечатка `["promtps"]`
// в invalidateQueries не давала ошибки на compile-time, баг находился только
// при ручном тестировании ("почему данные не обновляются"). Теперь:
//
//   queryKeys.prompts.detail(42)        // ['prompts', 'detail', 42]
//   queryKeys.prompts.list({...})        // ['prompts', 'list', filters]
//   queryKeys.admin.users.detail(1)      // ['admin', 'users', 'detail', 1]
//
// invalidate один scope:
//   qc.invalidateQueries({ queryKey: queryKeys.prompts.all })
//
// Структура inspired by https://tkdodo.eu/blog/effective-react-query-keys
// — root → list/detail/all для consistent invalidation hierarchy.

// AnalyticsRange — period для трендов аналитики.
export type AnalyticsRange = "7d" | "30d" | "90d" | "365d"

// AnalyticsFilter — опциональные фильтры для analytics queries.
export type AnalyticsFilter = {
  tagId?: number | null
  collectionId?: number | null
}

// Generic helper для filter в queryKey — не используем undefined, нормализуем
// в null чтобы JSON-ключ был стабилен (undefined превращается в "null" в JSON).
function norm<T>(v: T | undefined | null): T | null {
  return v ?? null
}

export const queryKeys = {
  prompts: {
    all: ["prompts"] as const,
    list: (filter: Record<string, unknown>) => ["prompts", "list", filter] as const,
    detail: (id: number) => ["prompts", "detail", id] as const,
    recent: (teamId: number | null) => ["prompts", "recent", norm(teamId)] as const,
    pinned: (teamId: number | null) => ["prompts", "pinned", norm(teamId)] as const,
    versions: (promptId: number) => ["prompts", "versions", promptId] as const,
    usageHistory: (filter: Record<string, unknown>) => ["prompts", "usage-history", filter] as const,
  },

  collections: {
    all: ["collections"] as const,
    list: (filter?: Record<string, unknown>) => ["collections", "list", norm(filter)] as const,
    detail: (id: number) => ["collections", "detail", id] as const,
  },

  tags: {
    all: ["tags"] as const,
    list: (teamId: number | null) => ["tags", "list", norm(teamId)] as const,
  },

  chains: {
    all: ["chains"] as const,
    list: (filter?: Record<string, unknown>) => ["chains", "list", norm(filter)] as const,
    detail: (id: number) => ["chains", "detail", id] as const,
    executions: (chainId: number) => ["chains", "executions", chainId] as const,
    execution: (executionId: number) => ["chains", "execution", executionId] as const,
  },

  analytics: {
    all: ["analytics"] as const,
    personal: (range: AnalyticsRange, filter?: AnalyticsFilter) =>
      ["analytics", "personal", range, norm(filter?.tagId), norm(filter?.collectionId)] as const,
    team: (teamId: number, range: AnalyticsRange, filter?: AnalyticsFilter) =>
      ["analytics", "team", teamId, range, norm(filter?.tagId), norm(filter?.collectionId)] as const,
    prompt: (promptId: number) => ["analytics", "prompt", promptId] as const,
    insights: () => ["analytics", "insights"] as const,
  },

  admin: {
    all: ["admin"] as const,
    users: {
      list: (filter: Record<string, unknown>) => ["admin", "users", filter] as const,
      detail: (id: number) => ["admin", "user", id] as const,
    },
    audit: (filter: Record<string, unknown>) => ["admin", "audit", filter] as const,
    health: () => ["admin", "health"] as const,
    feedbacks: {
      list: (filter: Record<string, unknown>) => ["admin", "feedbacks", filter] as const,
      detail: (id: number) => ["admin", "feedback", id] as const,
    },
  },

  teams: {
    all: ["teams"] as const,
    list: () => ["teams", "list"] as const,
    detail: (slug: string) => ["teams", "detail", slug] as const,
    branding: (slug: string) => ["branding", slug] as const,
    activity: (teamSlug: string, filter?: Record<string, unknown>) =>
      ["activity", teamSlug, norm(filter)] as const,
  },

  shares: {
    all: ["shares"] as const,
    list: (promptId: number) => ["shares", "list", promptId] as const,
  },

  apiKeys: {
    all: ["api-keys"] as const,
    list: () => ["api-keys"] as const,
  },

  trash: {
    all: ["trash"] as const,
    list: () => ["trash"] as const,
    count: () => ["trash-count"] as const,
  },

  changelog: ["changelog"] as const,
  streak: ["streak"] as const,
  badges: ["badges"] as const,
  linkedAccounts: ["linked-accounts"] as const,
  plans: ["plans"] as const,
  subscription: ["subscription"] as const,
  quota: ["quota"] as const,

  // 'me' хранится в auth-store (Zustand), не в TanStack cache.
  // НЕ используйте queryKey ["me"] — invalidate ничего не сделает.
  // Вместо этого: useAuthStore.getState().fetchMe()
} as const
