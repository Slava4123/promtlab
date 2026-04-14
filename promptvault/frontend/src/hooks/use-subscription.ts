import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { useAuthStore } from "@/stores/auth-store"
import {
  fetchPlans,
  fetchSubscription,
  fetchUsage,
  postCheckout,
  postCancelSubscription,
  postDowngrade,
  postSetAutoRenew,
} from "@/api/subscription"

export function usePlans() {
  return useQuery({
    queryKey: ["plans"],
    queryFn: fetchPlans,
  })
}

export function useSubscription() {
  return useQuery({
    queryKey: ["subscription"],
    queryFn: fetchSubscription,
  })
}

export function useUsage() {
  return useQuery({
    queryKey: ["subscription", "usage"],
    queryFn: fetchUsage,
  })
}

export function useCheckout() {
  return useMutation({
    mutationFn: (planId: string) => postCheckout(planId),
    onSuccess: (data) => {
      // Сохраняем метку что был checkout — при возврате в приложение
      // (даже если T-Bank redirect идёт на лендинг, а не на localhost)
      // приложение рефетчит подписку.
      sessionStorage.setItem("pending_checkout", "true")
      window.location.href = data.payment_url
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось начать оплату")
    },
  })
}

export function useCancelSubscription() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)
  return useMutation({
    mutationFn: postCancelSubscription,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["subscription"] })
      qc.invalidateQueries({ queryKey: ["subscription", "usage"] })
      fetchMe()
      toast.success("Подписка будет отменена в конце периода")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось отменить подписку")
    },
  })
}

export function useDowngrade() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)
  return useMutation({
    mutationFn: postDowngrade,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["subscription"] })
      qc.invalidateQueries({ queryKey: ["subscription", "usage"] })
      fetchMe()
      toast.success("Вы перешли на бесплатный план")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось сменить план")
    },
  })
}

export function useSetAutoRenew() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (autoRenew: boolean) => postSetAutoRenew(autoRenew),
    onSuccess: (_, autoRenew) => {
      qc.invalidateQueries({ queryKey: ["subscription"] })
      toast.success(
        autoRenew
          ? "Автопродление включено"
          : "Автопродление отключено — подписка истечёт в конце периода",
      )
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось изменить автопродление")
    },
  })
}

/**
 * Результат polling'а подписки после оплаты.
 * - "updated" — plan обновился (webhook дошёл)
 * - "timeout" — за отведённое время не обновился (webhook задерживается/потерян)
 * - "already_pro" — план уже был не-free в момент вызова (upgrade Pro→Max и т.п.)
 */
export type RefreshResult = "updated" | "timeout" | "already_pro"

// Polling config: 40 × 3 сек = 2 минуты. T-Bank в sandbox иногда шлёт
// webhook с задержкой до 30-60 сек, поэтому 15 сек (старый лимит) было мало.
const POLL_INTERVAL_MS = 3000
const POLL_MAX_ATTEMPTS = 40

/**
 * Хук для обновления subscription-related данных после оплаты.
 * Поллит /auth/me пока plan_id не изменится или не истечёт таймаут.
 * Возвращает результат — вызывающий код может показать финальный toast.
 */
export function useRefreshSubscription() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)

  return async (): Promise<RefreshResult> => {
    const invalidateAll = () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: ["subscription"] }),
        qc.invalidateQueries({ queryKey: ["subscription", "usage"] }),
      ])

    await invalidateAll()
    await fetchMe()

    const initial = useAuthStore.getState().user
    if (initial?.plan_id && initial.plan_id !== "free") {
      return "already_pro"
    }

    for (let i = 0; i < POLL_MAX_ATTEMPTS; i++) {
      await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS))
      await fetchMe()
      const updated = useAuthStore.getState().user
      if (updated?.plan_id && updated.plan_id !== "free") {
        await invalidateAll()
        return "updated"
      }
    }
    return "timeout"
  }
}
