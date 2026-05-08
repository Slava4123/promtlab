import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { ChangelogResponse } from "@/api/types"

export function useChangelog() {
  return useQuery({
    queryKey: ["changelog"],
    queryFn: () => api<ChangelogResponse>("/changelog"),
    // MJ-24: changelog обновляется при релизе фич (раз в неделю-две);
    // 30 минут staleTime защищает от повторного fetch'а на каждый mount
    // компонента ChangelogBadge.
    staleTime: 30 * 60_000,
  })
}

export function useMarkChangelogSeen() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiVoid("/changelog/seen", { method: "POST" }),
    onSuccess: () => {
      // MN-49: ["auth"] не существует как queryKey (мы используем ["me"]).
      // Использую существующий ключ + changelog для инвалидации badge'а.
      qc.invalidateQueries({ queryKey: ["changelog"] })
      qc.invalidateQueries({ queryKey: ["me"] })
    },
  })
}
