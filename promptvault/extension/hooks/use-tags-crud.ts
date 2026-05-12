import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import { useWorkspaceStore } from "../stores/workspace-store"
import type { Tag, TagDTO } from "../lib/types"
import type { CreateTagBody } from "../lib/api"

export function useTags() {
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  return useQuery<TagDTO[]>({
    queryKey: [...qk.tags, teamId],
    queryFn: () => sendBg({ type: "api.listTags", teamId }),
    staleTime: 60_000,
  })
}

export function useCreateTag() {
  const qc = useQueryClient()
  return useMutation<Tag, Error, CreateTagBody>({
    mutationFn: (body) => sendBg({ type: "api.createTag", body }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.tags }),
  })
}

export function useDeleteTag() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.deleteTag", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.tags })
      void qc.invalidateQueries({ queryKey: qk.prompts })
    },
  })
}
