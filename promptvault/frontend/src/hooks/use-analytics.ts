import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  fetchPersonalAnalytics,
  fetchTeamAnalytics,
  fetchPromptAnalytics,
  fetchInsights,
  refreshInsights,
  type AnalyticsRange,
  type InsightsResponse,
  type PersonalAnalyticsFilter,
} from "@/api/analytics"

export function usePersonalAnalytics(range: AnalyticsRange, filter?: PersonalAnalyticsFilter) {
  // queryKey включает фильтры, чтобы drill-down инвалидировал кеш
  // при смене tag/collection.
  return useQuery({
    queryKey: ["analytics", "personal", range, filter?.tagId ?? null, filter?.collectionId ?? null],
    queryFn: () => fetchPersonalAnalytics(range, filter),
  })
}

export function useTeamAnalytics(teamId: number | undefined, range: AnalyticsRange, filter?: PersonalAnalyticsFilter) {
  return useQuery({
    queryKey: ["analytics", "team", teamId, range, filter?.tagId ?? null, filter?.collectionId ?? null],
    queryFn: () => fetchTeamAnalytics(teamId!, range, filter),
    enabled: typeof teamId === "number" && teamId > 0,
  })
}

export function usePromptAnalytics(promptId: number) {
  return useQuery({
    queryKey: ["analytics", "prompt", promptId],
    queryFn: () => fetchPromptAnalytics(promptId),
    enabled: promptId > 0,
  })
}

export function useInsights(enabled = true) {
  return useQuery({
    queryKey: ["analytics", "insights"],
    queryFn: () => fetchInsights(),
    enabled,
  })
}

// useRefreshInsights — mutation-хук для кнопки «Обновить сейчас» в InsightsPanel.
// После успеха кэширует свежий ответ как результат useInsights — UI перерисуется
// без дополнительного fetch.
export function useRefreshInsights() {
  const qc = useQueryClient()
  return useMutation<InsightsResponse, Error, void>({
    mutationFn: () => refreshInsights(),
    onSuccess: (data) => {
      qc.setQueryData(["analytics", "insights"], data)
    },
  })
}
