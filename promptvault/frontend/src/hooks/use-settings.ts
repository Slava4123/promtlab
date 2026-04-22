import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api } from "@/api/client"
import type { LinkedAccount, UpdateProfileRequest, ChangePasswordRequest } from "@/api/types"

export function useLinkedAccounts() {
  return useQuery({
    queryKey: ["linked-accounts"],
    queryFn: () => api<LinkedAccount[]>("/auth/linked-accounts"),
  })
}

export function useUpdateProfile() {
  return useMutation({
    mutationFn: (data: UpdateProfileRequest) =>
      api<unknown>("/auth/profile", { method: "PUT", body: JSON.stringify(data) }),
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось обновить профиль")
    },
  })
}

export function useInitiateSetPassword() {
  return useMutation({
    mutationFn: () =>
      api<{ message: string }>("/auth/set-password/initiate", { method: "POST" }),
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось отправить код")
    },
  })
}

export function useConfirmSetPassword() {
  return useMutation({
    mutationFn: (data: { code: string; password: string }) =>
      api<{ message: string }>("/auth/set-password/confirm", { method: "POST", body: JSON.stringify(data) }),
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось установить пароль")
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (data: ChangePasswordRequest) =>
      api<{ message: string }>("/auth/password", { method: "PUT", body: JSON.stringify(data) }),
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось изменить пароль")
    },
  })
}

export function useUnlinkProvider() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (provider: string) =>
      api<{ message: string }>(`/auth/unlink/${provider}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["linked-accounts"] })
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось отвязать аккаунт")
    },
  })
}

// Phase 14 M-10: opt-in toggle для email-digest по Smart Insights.
// PATCH /api/auth/notifications/insights — { enabled: boolean }.
// Отвечает { insight_emails_enabled: boolean }. Инвалидация "me" чтобы
// обновить user.insight_emails_enabled в auth-store.
export function useSetInsightEmails() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (enabled: boolean) =>
      api<{ insight_emails_enabled: boolean }>(
        "/auth/notifications/insights",
        { method: "PATCH", body: JSON.stringify({ enabled }) },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["me"] })
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось обновить настройки уведомлений")
    },
  })
}
