import { useInfiniteQuery } from "@tanstack/react-query"
import { fetchTeamActivity, type ActivityFilters } from "@/api/activity"

const PAGE_SIZE = 50

// Infinite scroll — TanStack Query useInfiniteQuery с offset pagination.
// Backend возвращает has_more — именно от него зависит getNextPageParam.
export function useTeamActivity(slug: string, filters?: ActivityFilters) {
  return useInfiniteQuery({
    queryKey: ["activity", "team", slug, filters ?? {}],
    queryFn: ({ pageParam }) => fetchTeamActivity(slug, pageParam as number, PAGE_SIZE, filters),
    initialPageParam: 1,
    getNextPageParam: (last) => (last.has_more ? last.page + 1 : undefined),
    enabled: slug.length > 0,
  })
}
