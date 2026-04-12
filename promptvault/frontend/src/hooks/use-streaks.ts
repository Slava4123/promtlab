import { useQuery } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { StreakResponse } from "@/api/types"

export function useStreak() {
  return useQuery({
    queryKey: ["streak"],
    queryFn: () => api<StreakResponse>("/streaks"),
    staleTime: 60_000,
  })
}
