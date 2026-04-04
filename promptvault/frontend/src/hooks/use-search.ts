import { useQuery } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { SearchResponse } from "@/api/types"

export function useSearch(query: string, teamId?: number | null) {
  return useQuery({
    queryKey: ["search", query, teamId],
    queryFn: () => {
      const params = new URLSearchParams({ q: query })
      if (teamId) params.set("team_id", String(teamId))
      return api<SearchResponse>(`/search?${params}`)
    },
    enabled: query.length >= 2,
    staleTime: 30_000,
    placeholderData: (prev) => prev,
  })
}
