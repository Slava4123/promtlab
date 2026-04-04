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
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  })
}

export function useDeleteTag() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/tags/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
    },
  })
}
