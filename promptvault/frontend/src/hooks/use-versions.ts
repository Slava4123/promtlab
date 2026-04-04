import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { Prompt, PromptVersion, PaginatedResponse } from "@/api/types"

const PAGE_SIZE = 20

export function useVersions(promptId: number) {
  return useInfiniteQuery({
    queryKey: ["versions", promptId],
    queryFn: ({ pageParam }) =>
      api<PaginatedResponse<PromptVersion>>(`/prompts/${promptId}/versions?page=${pageParam}&page_size=${PAGE_SIZE}`),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.page + 1 : undefined,
    enabled: promptId > 0,
  })
}

export function useRevertVersion() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ promptId, versionId }: { promptId: number; versionId: number }) =>
      api<Prompt>(`/prompts/${promptId}/revert/${versionId}`, { method: "POST" }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ["prompts"], exact: false })
      qc.invalidateQueries({ queryKey: ["prompt", vars.promptId] })
      qc.invalidateQueries({ queryKey: ["versions", vars.promptId] })
    },
  })
}
