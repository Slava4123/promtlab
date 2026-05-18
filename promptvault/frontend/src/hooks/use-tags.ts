import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { Tag } from "@/api/types"

export function useTags(teamId?: number | null) {
  return useQuery({
    queryKey: ["tags", { teamId }],
    queryFn: () => {
      const params = teamId ? `?team_id=${teamId}` : ""
      return api<Tag[]>(`/tags${params}`)
    },
  })
}

export function useCreateTag() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; color?: string; team_id?: number | null }) =>
      api<Tag>("/tags", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] })
      // Backend пересчитывает orphan_tags после Create (новый тег не имеет
      // привязок к промптам → гарантированно orphan). Обновляем analytics
      // панель и /tags?filter=orphan overlay сразу.
      qc.invalidateQueries({ queryKey: ["analytics", "insights"] })
      qc.invalidateQueries({ queryKey: ["tags", "orphan"] })
    },
  })
}

export function useDeleteTag() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/tags/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
      // Backend пересчитывает orphan_tags insight inline после DELETE —
      // обновляем analytics панель сразу вместо ожидания ночного cron.
      qc.invalidateQueries({ queryKey: ["analytics", "insights"] })
    },
  })
}
