import { useQuery } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { SuggestResponse } from "@/api/types"

export function useSuggest(query: string, teamId?: number | null) {
  return useQuery({
    queryKey: ["search-suggest", query, teamId],
    queryFn: () => {
      const params = new URLSearchParams({ q: query })
      if (teamId) params.set("team_id", String(teamId))
      return api<SuggestResponse>(`/search/suggest?${params}`)
    },
    enabled: query.length >= 2,
    staleTime: 60_000,
    placeholderData: (prev) => prev,
  })
}
