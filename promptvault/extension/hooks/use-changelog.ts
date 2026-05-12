import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { ChangelogResponse } from "../lib/types"

export function useChangelog() {
  return useQuery<ChangelogResponse>({
    queryKey: qk.changelog,
    queryFn: () => sendBg({ type: "api.getChangelog" }),
    staleTime: 5 * 60_000,
  })
}

export function useMarkChangelogRead() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, void>({
    mutationFn: () => sendBg({ type: "api.markChangelogRead" }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.changelog }),
  })
}
