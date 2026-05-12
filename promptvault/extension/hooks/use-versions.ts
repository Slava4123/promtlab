import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { Prompt, PromptVersion } from "../lib/types"

interface VersionsResponse {
  items: PromptVersion[]
  total: number
  has_more: boolean
}

export function useVersions(promptId: number | null) {
  return useQuery<VersionsResponse>({
    queryKey: promptId ? qk.versions(promptId) : ["versions", "none"],
    queryFn: () => {
      if (!promptId) throw new Error("no prompt id")
      return sendBg({ type: "api.listVersions", promptId, limit: 50 })
    },
    enabled: promptId !== null,
    staleTime: 30_000,
  })
}

export function useRevertVersion(promptId: number) {
  const qc = useQueryClient()
  return useMutation<Prompt, Error, number>({
    mutationFn: (versionId) => sendBg({ type: "api.revertVersion", promptId, versionId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.promptDetail(promptId) })
      void qc.invalidateQueries({ queryKey: qk.versions(promptId) })
      void qc.invalidateQueries({ queryKey: qk.prompts })
    },
  })
}
