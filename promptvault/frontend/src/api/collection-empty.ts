import { api } from "./client"

// fetchEmptyCollections — GET /api/collections/empty (B11).
// «empty» = коллекция без активных (не soft-deleted) промптов. Бэк отдаёт
// упрощённый shape {id, name} (без color/icon/prompt_count) — этого
// достаточно для overlay в /collections?filter=empty: юзер пришёл из
// Smart Insights почистить мусор и кликает Удалить, цвет/иконка тут не нужны.
// Защищаемся от пустого ответа `{}` (real-case на чистой БД) через `?? []`.
export interface EmptyCollection {
  id: number
  name: string
}

interface EmptyCollectionsEnvelope {
  items?: EmptyCollection[]
}

export async function fetchEmptyCollections(): Promise<EmptyCollection[]> {
  const body = await api<EmptyCollectionsEnvelope>("/collections/empty")
  return body.items ?? []
}
