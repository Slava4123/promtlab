import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, apiVoid } from "@/api/client"
import type { APIKeyListResponse, CreateAPIKeyRequest, CreatedAPIKey } from "@/api/types"

export function useAPIKeys() {
  return useQuery({
    queryKey: ["api-keys"],
    queryFn: () => api<APIKeyListResponse>("/api-keys"),
  })
}

export function useCreateAPIKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateAPIKeyRequest) =>
      api<CreatedAPIKey>("/api-keys", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["api-keys"] })
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось создать API-ключ")
    },
  })
}

export function useRevokeAPIKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      apiVoid(`/api-keys/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["api-keys"] })
      toast.success("API-ключ отозван")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось отозвать API-ключ")
    },
  })
}
