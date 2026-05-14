import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import { useWorkspace } from "./use-workspace"
import type { Collection, CollectionDTO } from "../lib/types"
import type { CreateCollectionBody } from "../lib/api"

export function useCollections() {
  // Workspace через chrome.storage (синхронно с WorkspaceSelector в Home).
  const { workspaceId } = useWorkspace()
  return useQuery<CollectionDTO[]>({
    queryKey: [...qk.collections, workspaceId],
    queryFn: () => sendBg({ type: "api.listCollections", teamId: workspaceId }),
    staleTime: 60_000,
  })
}

export function useCreateCollection() {
  const qc = useQueryClient()
  return useMutation<Collection, Error, CreateCollectionBody>({
    mutationFn: (body) => sendBg({ type: "api.createCollection", body }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.collections })
      // Collections count в Подписке растёт.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}

export function useUpdateCollection() {
  const qc = useQueryClient()
  return useMutation<Collection, Error, { id: number; body: Partial<CreateCollectionBody> }>({
    mutationFn: ({ id, body }) => sendBg({ type: "api.updateCollection", id, body }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.collections }),
  })
}

export function useDeleteCollection() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (id) => sendBg({ type: "api.deleteCollection", id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.collections })
      void qc.invalidateQueries({ queryKey: qk.trash })
      // Промпты в коллекции теряют привязку — у них меняется отображение.
      void qc.invalidateQueries({ queryKey: qk.prompts })
      // Collections count уменьшается.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}
