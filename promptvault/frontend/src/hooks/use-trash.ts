import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { TrashListResponse, TrashCounts } from "@/api/types"

interface TrashFilters {
  team_id?: number | null
  page?: number
  page_size?: number
}

export function useTrash(filters: TrashFilters = {}) {
  const params = new URLSearchParams()
  if (filters.page) params.set("page", String(filters.page))
  if (filters.page_size) params.set("page_size", String(filters.page_size))
  if (filters.team_id) params.set("team_id", String(filters.team_id))

  return useQuery({
    queryKey: ["trash", filters],
    queryFn: () => api<TrashListResponse>(`/trash?${params.toString()}`),
  })
}

export function useTrashCount(teamId?: number | null) {
  const params = teamId ? `?team_id=${teamId}` : ""
  return useQuery({
    queryKey: ["trash-count", teamId],
    queryFn: () => api<TrashCounts>(`/trash/count${params}`),
  })
}

export function useRestoreItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ type, id }: { type: "prompt" | "collection" | "tag"; id: number }) =>
      api<{ status: string }>(`/trash/${type}/${id}/restore`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["trash"] })
      qc.invalidateQueries({ queryKey: ["trash-count"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["tags"] })
    },
  })
}

export function usePermanentDelete() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ type, id }: { type: "prompt" | "collection" | "tag"; id: number }) =>
      apiVoid(`/trash/${type}/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["trash"] })
      qc.invalidateQueries({ queryKey: ["trash-count"] })
    },
  })
}

export function useEmptyTrash() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (teamId?: number | null) => {
      const params = teamId ? `?team_id=${teamId}` : ""
      return api<{ deleted: number }>(`/trash${params}`, { method: "DELETE" })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["trash"] })
      qc.invalidateQueries({ queryKey: ["trash-count"] })
    },
  })
}
