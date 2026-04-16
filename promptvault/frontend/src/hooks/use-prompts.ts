import { useQuery, useInfiniteQuery, useMutation, useQueryClient, type InfiniteData } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import { captureException } from "@/lib/sentry"
import type { Prompt, PaginatedResponse, PinResult, IncrementUsageResponse } from "@/api/types"
import { useBadgeUnlocks } from "./use-badge-toast"

interface PromptFilters {
  q?: string
  favorite?: boolean
  collection_id?: number
  tag_ids?: number[]
  team_id?: number | null
}

type PromptsPage = PaginatedResponse<Prompt>
type PromptsInfiniteData = InfiniteData<PromptsPage>

// Предикат инвалидации/оптимистичных апдейтов: только главный grid /prompts,
// не побочные queries (["prompts","pinned"], ["prompts","recent"]).
const isMainPromptsQuery = (queryKey: unknown): boolean => {
  if (!Array.isArray(queryKey)) return true
  const second = queryKey[1]
  return second !== "pinned" && second !== "recent"
}

const PAGE_SIZE = 18

export function usePrompts(filters: PromptFilters = {}) {
  return useInfiniteQuery({
    queryKey: ["prompts", filters],
    queryFn: ({ pageParam }) => {
      const params = new URLSearchParams()
      params.set("page", String(pageParam))
      params.set("page_size", String(PAGE_SIZE))
      if (filters.q) params.set("q", filters.q)
      if (filters.favorite) params.set("favorite", "true")
      if (filters.collection_id) params.set("collection_id", String(filters.collection_id))
      if (filters.tag_ids?.length) params.set("tag_ids", filters.tag_ids.join(","))
      if (filters.team_id) params.set("team_id", String(filters.team_id))
      return api<PromptsPage>(`/prompts?${params.toString()}`)
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.page + 1 : undefined,
  })
}

export function usePrompt(id: number) {
  return useQuery({
    queryKey: ["prompt", id],
    queryFn: () => api<Prompt>(`/prompts/${id}`),
    enabled: id > 0,
  })
}

export function useCreatePrompt() {
  const qc = useQueryClient()
  const handleBadges = useBadgeUnlocks()
  return useMutation({
    mutationFn: (data: { title: string; content: string; model?: string; team_id?: number | null; collection_ids?: number[]; tag_ids?: number[] }) =>
      api<Prompt>("/prompts", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: (data) => {
      // Узкая инвалидация: только активные списки. Коллекции/теги
      // могли измениться (счётчики), поэтому трогаем их тоже.
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["collections"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["tags"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["streak"], refetchType: "active" })
      handleBadges(data.newly_unlocked_badges)
    },
  })
}

export function useUpdatePrompt() {
  const qc = useQueryClient()
  const handleBadges = useBadgeUnlocks()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; title?: string; content?: string; model?: string; change_note?: string; collection_ids?: number[]; tag_ids?: number[]; is_public?: boolean }) =>
      api<Prompt>(`/prompts/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: (data, vars) => {
      // Активные списки + карточка. Коллекции/теги только если они в payload'е
      // (could change), streak обновится если это первое использование в день.
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["prompt", vars.id] })
      if (vars.collection_ids !== undefined) {
        qc.invalidateQueries({ queryKey: ["collections"], refetchType: "active" })
        qc.invalidateQueries({ queryKey: ["collection"], refetchType: "active" })
      }
      if (vars.tag_ids !== undefined) {
        qc.invalidateQueries({ queryKey: ["tags"], refetchType: "active" })
      }
      qc.invalidateQueries({ queryKey: ["streak"], refetchType: "active" })
      handleBadges(data.newly_unlocked_badges)
    },
  })
}

export function useDeletePrompt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/prompts/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["collections"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["collection"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["tags"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["trash"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["trash-count"], refetchType: "active" })
    },
  })
}

export function useIncrementUsage() {
  const qc = useQueryClient()
  const handleBadges = useBadgeUnlocks()
  return useMutation({
    mutationFn: (id: number) => api<IncrementUsageResponse>(`/prompts/${id}/use`, { method: "POST" }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["prompts", "recent"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["streak"], refetchType: "active" })
      handleBadges(data.newly_unlocked_badges)
    },
    onError: (err) => {
      // Fire-and-forget: log but do not surface to user — copy already succeeded.
      console.error("Failed to increment usage", err)
      captureException(err instanceof Error ? err : new Error(String(err)), { tags: { feature: "increment-usage" } })
    },
  })
}

export function usePinnedPrompts(teamId?: number | null) {
  const params = new URLSearchParams()
  if (teamId) params.set("team_id", String(teamId))
  params.set("limit", "20")
  return useQuery({
    queryKey: ["prompts", "pinned", teamId ?? null],
    queryFn: () => api<{ items: Prompt[]; total: number }>(`/prompts/pinned?${params.toString()}`),
  })
}

export function useRecentPrompts(teamId?: number | null) {
  const params = new URLSearchParams()
  if (teamId) params.set("team_id", String(teamId))
  params.set("limit", "6")
  return useQuery({
    queryKey: ["prompts", "recent", teamId ?? null],
    queryFn: () => api<{ items: Prompt[]; total: number }>(`/prompts/recent?${params.toString()}`),
  })
}

export function useTogglePin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, teamWide = false }: { id: number; teamWide?: boolean }) =>
      api<PinResult>(`/prompts/${id}/pin`, {
        method: "POST",
        body: JSON.stringify({ team_wide: teamWide }),
      }),
    onMutate: async ({ id, teamWide = false }) => {
      const filter = { queryKey: ["prompts"] as const, predicate: (q: { queryKey: unknown }) => isMainPromptsQuery(q.queryKey) }
      await qc.cancelQueries({ ...filter, type: "all" })
      const prev = qc.getQueriesData<PromptsInfiniteData>(filter)
      qc.setQueriesData<PromptsInfiniteData>(filter, (old) => {
        if (!old?.pages) return old
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            items: page.items.map((p) =>
              p.id === id
                ? {
                    ...p,
                    pinned_personal: teamWide ? p.pinned_personal : !p.pinned_personal,
                    pinned_team: teamWide ? !p.pinned_team : p.pinned_team,
                  }
                : p
            ),
          })),
        }
      })
      return { prev }
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) {
        for (const [key, data] of context.prev) {
          qc.setQueryData(key, data)
        }
      }
    },
    onSettled: (_data, _err, { id }) => {
      // После оптимистик-апдейта background-rollup только для активных страниц,
      // чтобы sync с сервером без лавины рефетчей неактивных фильтров.
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["prompt", id] })
    },
  })
}

export function useToggleFavorite() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api<Prompt>(`/prompts/${id}/favorite`, { method: "POST" }),
    onMutate: async (id) => {
      const filter = { queryKey: ["prompts"] as const }
      await qc.cancelQueries(filter)
      const prev = qc.getQueriesData<PromptsInfiniteData>(filter)
      qc.setQueriesData<PromptsInfiniteData>(filter, (old) => {
        if (!old?.pages) return old
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            items: page.items.map((p) =>
              p.id === id ? { ...p, favorite: !p.favorite } : p
            ),
          })),
        }
      })
      return { prev }
    },
    onError: (_err, _id, context) => {
      if (context?.prev) {
        for (const [key, data] of context.prev) {
          qc.setQueryData(key, data)
        }
      }
    },
    onSettled: (_data, _err, id) => {
      qc.invalidateQueries({ queryKey: ["prompts"], refetchType: "active" })
      qc.invalidateQueries({ queryKey: ["prompt", id] })
    },
  })
}
