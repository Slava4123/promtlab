// Mutations для CRUD промптов (Phase 1). Хуки query для list/pinned/recent
// остаются в существующем hooks/use-prompts.ts.

import { useMutation, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { CreatePromptBody, UpdatePromptBody } from "../lib/api"
import type { Prompt } from "../lib/types"

export function useCreatePrompt() {
  const qc = useQueryClient()
  return useMutation<Prompt, Error, CreatePromptBody>({
    mutationFn: (body) => sendBg({ type: "api.createPrompt", body }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.prompts })
      void qc.invalidateQueries({ queryKey: qk.trash })
    },
  })
}

export function useUpdatePrompt(id: number | null) {
  const qc = useQueryClient()
  return useMutation<Prompt, Error, UpdatePromptBody>({
    mutationFn: (body) => {
      if (id === null) throw new Error("no prompt id")
      return sendBg({ type: "api.updatePrompt", id, body })
    },
    onSuccess: (updated) => {
      void qc.invalidateQueries({ queryKey: qk.prompts })
      void qc.invalidateQueries({ queryKey: qk.promptDetail(updated.id) })
      void qc.invalidateQueries({ queryKey: qk.versions(updated.id) })
    },
  })
}

export function useDeletePrompt() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.deletePrompt", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.prompts })
      void qc.invalidateQueries({ queryKey: qk.trash })
    },
  })
}

export function useDuplicatePrompt() {
  const qc = useQueryClient()
  return useMutation<Prompt, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.duplicatePrompt", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.prompts })
    },
  })
}
