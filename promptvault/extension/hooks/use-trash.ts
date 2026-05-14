import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { TrashListResponse } from "../lib/types"

export function useTrash() {
  return useQuery<TrashListResponse>({
    queryKey: qk.trash,
    queryFn: () => sendBg({ type: "api.listTrash" }),
    staleTime: 30_000,
  })
}

export function useRestoreTrashPrompt() {
  const qc = useQueryClient()
  return useMutation<unknown, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.restoreTrashPrompt", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.trash })
      void qc.invalidateQueries({ queryKey: qk.prompts })
      // Prompts count растёт обратно.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}

export function useRestoreTrashCollection() {
  const qc = useQueryClient()
  return useMutation<unknown, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.restoreTrashCollection", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.trash })
      void qc.invalidateQueries({ queryKey: qk.collections })
      // Collections count растёт обратно.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}

export function usePermanentDeletePrompt() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.permanentDeleteTrashPrompt", id }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.trash }),
  })
}

export function usePermanentDeleteCollection() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.permanentDeleteTrashCollection", id }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.trash }),
  })
}

export function useEmptyTrash() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, void>({
    mutationFn: () => sendBg({ type: "api.emptyTrash" }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.trash }),
  })
}
