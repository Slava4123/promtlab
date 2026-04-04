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

export function useToggleFavorite() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api<Prompt>(`/prompts/${id}/favorite`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["prompts"] }),
  })
}
