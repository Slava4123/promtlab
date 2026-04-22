import { useSearchParams } from "react-router-dom"
import { useMemo } from "react"

// useAnalyticsFilter — читает drill-down фильтры из URL query params.
// Поддерживает ?tag=:id и ?collection=:id для фильтрации всех метрик
// analytics-страниц. Задача #9 бэклога: drill-down по тегам/коллекциям.
//
// Использование:
//   const { tagId, collectionId, setTagId } = useAnalyticsFilter()
//   const { data } = usePersonalAnalytics(range, { tagId, collectionId })
//
// URL синхронизация: setter использует setSearchParams (replace).
export type AnalyticsFilterValue = {
  tagId: number | null
  collectionId: number | null
}

export function useAnalyticsFilter() {
  const [params, setParams] = useSearchParams()

  const value = useMemo<AnalyticsFilterValue>(() => {
    const tagRaw = params.get("tag")
    const colRaw = params.get("collection")
    return {
      tagId: tagRaw ? Number(tagRaw) : null,
      collectionId: colRaw ? Number(colRaw) : null,
    }
  }, [params])

  function setTagId(id: number | null) {
    const next = new URLSearchParams(params)
    if (id == null) next.delete("tag")
    else next.set("tag", String(id))
    setParams(next, { replace: true })
  }

  function setCollectionId(id: number | null) {
    const next = new URLSearchParams(params)
    if (id == null) next.delete("collection")
    else next.set("collection", String(id))
    setParams(next, { replace: true })
  }

  function reset() {
    const next = new URLSearchParams(params)
    next.delete("tag")
    next.delete("collection")
    setParams(next, { replace: true })
  }

  return { ...value, setTagId, setCollectionId, reset }
}
