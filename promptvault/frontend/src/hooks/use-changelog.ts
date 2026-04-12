import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { ChangelogResponse } from "@/api/types"

export function useChangelog() {
  return useQuery({
    queryKey: ["changelog"],
    queryFn: () => api<ChangelogResponse>("/changelog"),
    staleTime: 5 * 60_000,
  })
}

export function useMarkChangelogSeen() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiVoid("/changelog/seen", { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["changelog"] })
      qc.invalidateQueries({ queryKey: ["auth"] })
    },
  })
}
