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
      qc.invalidateQueries({ queryKey: ["prompt-insights", "duplicates"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
    },
  })
}
