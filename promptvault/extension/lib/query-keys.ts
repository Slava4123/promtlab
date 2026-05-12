// Централизованные query keys для cache invalidation. Иерархия:
//   ['prompts']                  — все списки промптов
//   ['prompts', 'list', filter]  — конкретный список
//   ['prompts', 'detail', id]    — отдельный промпт
//   ['prompts', 'pinned']
//   ['prompts', 'recent']
//   ['versions', promptId]
//   ['trash']
//   ['share', promptId]
//   ['collections']
//   ['tags']
//   ['teams']
//   ['streak'], ['badges'], ['changelog']
//   ['subscription', 'plans' | 'current' | 'usage']
//   ['apikeys']
//   ['chains']
//   ['executions', execId]

export const qk = {
  prompts: ["prompts"] as const,
  promptDetail: (id: number) => ["prompts", "detail", id] as const,
  promptList: (filter?: unknown) => ["prompts", "list", filter] as const,
  pinned: ["prompts", "pinned"] as const,
  recent: ["prompts", "recent"] as const,
  versions: (promptId: number) => ["versions", promptId] as const,
  trash: ["trash"] as const,
  share: (promptId: number) => ["share", promptId] as const,
  collections: ["collections"] as const,
  tags: ["tags"] as const,
  teams: ["teams"] as const,
  team: (slug: string) => ["teams", slug] as const,
  streak: ["streak"] as const,
  badges: ["badges"] as const,
  changelog: ["changelog"] as const,
  plans: ["subscription", "plans"] as const,
  subscription: ["subscription", "current"] as const,
  usage: ["subscription", "usage"] as const,
  apikeys: ["apikeys"] as const,
  chains: ["chains"] as const,
  chain: (id: number) => ["chains", id] as const,
  executions: (chainId: number) => ["executions", chainId] as const,
  execution: (execId: number) => ["execution", execId] as const,
  me: ["me"] as const,
} as const
