import { api } from "./client"
import type { Tag } from "./types"

// fetchOrphanTags — GET /api/tags/orphan (B10).
// «orphan» = тег, который не привязан ни к одному активному (не soft-deleted)
// промпту. Бэк возвращает {items: Tag[]} envelope; защищаемся от пустого
// ответа `{}` (real-case на чистой БД у нового юзера) через `?? []`.
interface OrphanTagsEnvelope {
  items?: Tag[]
}

export async function fetchOrphanTags(): Promise<Tag[]> {
  const body = await api<OrphanTagsEnvelope>("/tags/orphan")
  return body.items ?? []
}
