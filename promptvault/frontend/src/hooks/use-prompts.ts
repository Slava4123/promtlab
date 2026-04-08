import { useQuery, useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { Prompt, PaginatedResponse } from "@/api/types"

interface PromptFilters {
  q?: string
  favorite?: boolean
  collection_id?: number
  tag_ids?: number[]
  team_id?: number | null
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
      return api<PaginatedResponse<Prompt>>(`/prompts?${params.toString()}`)
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
  return useMutation({
    mutationFn: (data: { title: string; content: string; model?: string; team_id?: number | null; collection_ids?: number[]; tag_ids?: number[] }) =>
      api<Prompt>("/prompts", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["tags"] })
    },
  })
}

export function useUpdatePrompt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; title?: string; content?: string; model?: string; change_note?: string; collection_ids?: number[]; tag_ids?: number[] }) =>
      api<Prompt>(`/prompts/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["prompt", vars.id] })
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["collection"] })
    },
  })
}

export function useDeletePrompt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/prompts/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["collection"] })
      qc.invalidateQueries({ queryKey: ["tags"] })
    },
  })
}

export function useIncrementUsage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/prompts/${id}/use`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["prompts"] })
    },
    onError: (err) => {
      // Fire-and-forget: log but do not surface to user — copy already succeeded.
      console.error("Failed to increment usage", err)
    },
  })
}

export function useToggleFavorite() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api<Prompt>(`/prompts/${id}/favorite`, { method: "POST" }),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ["prompts"] })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const prev = qc.getQueriesData<any>({ queryKey: ["prompts"] })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      qc.setQueriesData<any>({ queryKey: ["prompts"] }, (old: any) => {
        if (!old?.pages) return old
        return {
          ...old,
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          pages: old.pages.map((page: any) => ({
            ...page,
            items: page.items.map((p: Prompt) =>
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
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["prompt", id] })
    },
  })
}
