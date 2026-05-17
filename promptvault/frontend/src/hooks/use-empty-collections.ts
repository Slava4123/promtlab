import { useQuery } from "@tanstack/react-query"
import { fetchEmptyCollections } from "@/api/collection-empty"

// useEmptyCollections — fetcher для overlay /collections?filter=empty. Отдельный
// query от useCollections потому что (а) полный collections-список
// инвалидируется на любой create/update/delete, (б) бэк делает LEFT JOIN с
// prompts для подсчёта «пустоты», что дороже простого list.
export function useEmptyCollections() {
  return useQuery({ queryKey: ["collections", "empty"], queryFn: fetchEmptyCollections })
}
