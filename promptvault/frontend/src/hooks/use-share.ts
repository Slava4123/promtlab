import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, publicApi, ApiError } from "@/api/client"
import type { ShareLink, PublicPrompt } from "@/api/types"

const retryUnless404 = (_count: number, error: Error) =>
  !(error instanceof ApiError && error.status === 404)

export function useShareLink(promptId: number) {
  return useQuery({
    queryKey: ["share-link", promptId],
    queryFn: () => api<ShareLink>(`/prompts/${promptId}/share`),
    enabled: promptId > 0,
    retry: retryUnless404,
  })
}

export function useCreateShareLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (promptId: number) =>
      api<ShareLink>(`/prompts/${promptId}/share`, { method: "POST" }),
    onSuccess: (_, promptId) => {
      qc.invalidateQueries({ queryKey: ["share-link", promptId] })
    },
  })
}

export function useDeleteShareLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (promptId: number) =>
      api<void>(`/prompts/${promptId}/share`, { method: "DELETE" }),
    onSuccess: (_, promptId) => {
      qc.invalidateQueries({ queryKey: ["share-link", promptId] })
    },
  })
}

export function usePublicPrompt(token: string) {
  return useQuery({
    queryKey: ["public-prompt", token],
    queryFn: () => publicApi<PublicPrompt>(`/s/${token}`),
    enabled: !!token,
    retry: retryUnless404,
  })
}
