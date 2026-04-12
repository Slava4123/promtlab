import { useQuery } from "@tanstack/react-query"
import { fetchBadges } from "@/api/badges"

/**
 * useBadges — TanStack Query хук для GET /api/badges.
 *
 * Кеширует на 5 минут (бейджи меняются только при mutations, которые сами
 * invalidate этот query через useBadgeUnlocks onSuccess хук).
 */
export function useBadges() {
  return useQuery({
    queryKey: ["badges"],
    queryFn: fetchBadges,
    staleTime: 5 * 60 * 1000,
  })
}
