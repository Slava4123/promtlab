import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  fetchUnused,
  fetchDuplicates,
  fetchTrending,
  fetchDeclining,
  fetchMostEdited,
  mergePrompts,
} from "@/api/prompt-insights"

export function useUnusedPrompts() {
  return useQuery({ queryKey: ["prompt-insights", "unused"], queryFn: fetchUnused })
}

export function useDuplicates() {
  return useQuery({ queryKey: ["prompt-insights", "duplicates"], queryFn: fetchDuplicates })
}

export function useTrending() {
  return useQuery({ queryKey: ["prompt-insights", "trending"], queryFn: fetchTrending })
}

export function useDeclining() {
  return useQuery({ queryKey: ["prompt-insights", "declining"], queryFn: fetchDeclining })
}

export function useMostEdited() {
  return useQuery({ queryKey: ["prompt-insights", "most-edited"], queryFn: fetchMostEdited })
}

export function useMergePrompts() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ keepID, mergeID }: { keepID: number; mergeID: number }) =>
      mergePrompts(keepID, mergeID),
    onSuccess: () => {
      // Полная инвалидация prompt-insights — merge затрагивает 5 списков
      // (unused, duplicates, trending, declining, most_edited), invalidate
      // на parent key обновит всё одним движением.
      qc.invalidateQueries({ queryKey: ["prompt-insights"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
      // Backend пересчитывает 7 affected типов Smart Insights inline после
      // merge — обновляем analytics панель сразу вместо ожидания cron.
      qc.invalidateQueries({ queryKey: ["analytics", "insights"] })
    },
  })
}
