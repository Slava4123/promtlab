import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { Collection } from "@/api/types"

export function useCollection(id: number) {
  return useQuery({
    queryKey: ["collection", id],
    queryFn: () => api<Collection>(`/collections/${id}`),
    enabled: id > 0,
  })
}

export function useCollections(teamId?: number | null) {
  return useQuery({
    queryKey: ["collections", { teamId }],
    queryFn: () => {
      const params = teamId ? `?team_id=${teamId}` : ""
      return api<Collection[]>(`/collections${params}`)
    },
  })
}

export function useCreateCollection() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; description?: string; color?: string; icon?: string; team_id?: number | null }) =>
      api<Collection>("/collections", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["collections"] }),
  })
}

export function useUpdateCollection() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; name: string; description?: string; color?: string; icon?: string }) =>
      api<Collection>(`/collections/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["collection", vars.id] })
    },
  })
}

export function useDeleteCollection() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/collections/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["collections"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
    },
  })
}
