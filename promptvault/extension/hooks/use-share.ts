import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { ShareLink } from "../lib/types"

export function useShareLink(promptId: number | null) {
  return useQuery<ShareLink | null>({
    queryKey: promptId ? qk.share(promptId) : ["share", "none"],
    queryFn: () => {
      if (!promptId) throw new Error("no prompt id")
      return sendBg({ type: "api.getShareLink", promptId })
    },
    enabled: promptId !== null,
    staleTime: 60_000,
  })
}

export function useCreateShareLink(promptId: number) {
  const qc = useQueryClient()
  return useMutation<ShareLink, Error, void>({
    mutationFn: () => sendBg({ type: "api.createShareLink", promptId }),
    onSuccess: (link) => {
      qc.setQueryData(qk.share(promptId), link)
    },
  })
}

export function useDeactivateShareLink(promptId: number) {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, void>({
    mutationFn: () => sendBg({ type: "api.deactivateShareLink", promptId }),
    onSuccess: () => {
      qc.setQueryData(qk.share(promptId), null)
    },
  })
}
