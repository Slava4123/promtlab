import { useQuery } from "@tanstack/react-query"
import { fetchOrphanTags } from "@/api/tag-orphan"

// useOrphanTags — fetcher для overlay /tags?filter=orphan. Отдельный query
// от useTags потому что (а) полный tags-список инвалидируется на любой
// create/delete/rename, (б) бэк делает дополнительный LEFT JOIN с prompts
// для поиска «orphan», что дороже простого list.
export function useOrphanTags() {
  return useQuery({ queryKey: ["tags", "orphan"], queryFn: fetchOrphanTags })
}
