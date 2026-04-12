import { useInfiniteQuery } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { UsageLogEntry, PaginatedResponse } from "@/api/types"

const PAGE_SIZE = 20

export function useHistory(teamId?: number | null) {
  return useInfiniteQuery({
    queryKey: ["history", teamId ?? null],
    queryFn: ({ pageParam }) => {
      const params = new URLSearchParams()
      params.set("page", String(pageParam))
      params.set("page_size", String(PAGE_SIZE))
      if (teamId) params.set("team_id", String(teamId))
      return api<PaginatedResponse<UsageLogEntry>>(`/prompts/history?${params.toString()}`)
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.page + 1 : undefined,
  })
}
