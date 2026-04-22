import { useQuery } from "@tanstack/react-query"
import {
  fetchPersonalAnalytics,
  fetchTeamAnalytics,
  fetchPromptAnalytics,
  fetchInsights,
  type AnalyticsRange,
} from "@/api/analytics"

export function usePersonalAnalytics(range: AnalyticsRange) {
  return useQuery({
    queryKey: ["analytics", "personal", range],
    queryFn: () => fetchPersonalAnalytics(range),
  })
}

export function useTeamAnalytics(teamId: number | undefined, range: AnalyticsRange) {
  return useQuery({
    queryKey: ["analytics", "team", teamId, range],
    queryFn: () => fetchTeamAnalytics(teamId!, range),
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
