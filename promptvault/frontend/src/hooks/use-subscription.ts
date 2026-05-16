import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { useAuthStore } from "@/stores/auth-store"
import {
  fetchDowngradePreview,
  fetchPlans,
  fetchSubscription,
  fetchUsage,
  fetchTeamUsage,
  postCheckout,
  postCancelSubscription,
  postDowngrade,
  postPauseSubscription,
  postResumeSubscription,
  postSetAutoRenew,
  type CancelSubscriptionInput,
} from "@/api/subscription"

// CHECKOUT_INTENT_KEY — пометка что юзер хотел upgrade но не был залогинен.
// После успешного login/register → sign-in.tsx вызывает popCheckoutIntent и
// сразу начинает checkout, не теряя upsell momentum (M-14).
//
// Phase 16-Z (16.05.2026): храним JSON {planId, consent}. consent — отметил
// ли юзер чек-бокс согласия на recurrent списания на pricing/quota-dialog.
// Без consent=true sign-in.tsx НЕ запускает checkout автоматически — иначе
// мы продадим подписку без явного согласия и нарушим требование T-Bank
// (acq_help@tbank.ru) для активации метода Charge.
const CHECKOUT_INTENT_KEY = "pv_checkout_intent"

export interface CheckoutIntent {
  planId: string
  consent: boolean
}

export function saveCheckoutIntent(intent: CheckoutIntent) {
  try {
    sessionStorage.setItem(CHECKOUT_INTENT_KEY, JSON.stringify(intent))
  } catch { /* disabled storage */ }
}

export function popCheckoutIntent(): CheckoutIntent | null {
  try {
    const raw = sessionStorage.getItem(CHECKOUT_INTENT_KEY)
    sessionStorage.removeItem(CHECKOUT_INTENT_KEY)
    if (!raw) return null
    // Backward-compat: легаси-формат — просто planId-строка без обёртки.
    // Тогда consent=false → sign-in потребует подтверждения через UI.
    if (!raw.startsWith("{")) {
      return { planId: raw, consent: false }
    }
    const parsed = JSON.parse(raw) as Partial<CheckoutIntent>
    if (typeof parsed.planId !== "string") return null
    return { planId: parsed.planId, consent: parsed.consent === true }
  } catch {
    return null
  }
}

function isAuthError(err: unknown): boolean {
  if (!(err instanceof Error)) return false
  const msg = err.message.toLowerCase()
  return msg.includes("сессия истекла") || msg.includes("войдите") || msg.includes("unauthorized")
}

export function usePlans() {
  return useQuery({
    queryKey: ["plans"],
    queryFn: fetchPlans,
    // MJ-24: планы меняются редко (релиз новой ценовой структуры), но
    // /api/plans в дефолтном 5min staleTime инвалидируется на каждое
    // переключение страницы. 1 час — разумный баланс между свежестью
    // и нагрузкой на /api/plans от админ-навигации.
    staleTime: 60 * 60_000,
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

// Pack TU: usage конкретной команды (team-pool квоты).
// enabled=false если slug пустой — для condense-loading случая.
export function useTeamUsage(slug: string | null | undefined) {
  return useQuery({
    queryKey: ["team", "usage", slug],
    queryFn: () => fetchTeamUsage(slug!),
    enabled: !!slug,
    staleTime: 30_000,
  })
}

// useDowngradePreview — lazy-load (enabled=false), вызывается через refetch()
// прямо перед открытием confirm-dialog на Free. Cache — query key учитывает
// targetPlanId, staleTime 0, чтобы при повторной попытке юзер видел актуальное.
export function useDowngradePreview(targetPlanId = "free") {
  return useQuery({
    queryKey: ["subscription", "downgrade-preview", targetPlanId],
    queryFn: () => fetchDowngradePreview(targetPlanId),
    enabled: false,
    staleTime: 0,
    gcTime: 0,
  })
}

export interface CheckoutArgs {
  planId: string
  // consent — пользователь явно отметил чек-бокс согласия на recurrent
  // списания на pricing/quota-dialog. Phase 16-Z: при auth-fail сохраняем
  // вместе с planId, чтобы sign-in.tsx не запускал checkout без согласия.
  consent: boolean
}

export function useCheckout() {
  return useMutation({
    mutationFn: ({ planId }: CheckoutArgs) => postCheckout(planId),
    onSuccess: (data) => {
      // Сохраняем метку что был checkout — при возврате в приложение
      // (даже если T-Bank redirect идёт на лендинг, а не на localhost)
      // приложение рефетчит подписку.
      sessionStorage.setItem("pending_checkout", "true")
      window.location.href = data.payment_url
    },
    onError: (err: Error, { planId, consent }) => {
      // Если юзер не залогинен — сохраняем intent (с consent) и ведём на
      // sign-in, чтобы после login продолжить checkout без повторной
      // проставки галочки (M-14 + T-Bank consent flow).
      if (isAuthError(err)) {
        saveCheckoutIntent({ planId, consent })
        toast.info("Войдите, чтобы оформить подписку")
        // Используем полный redirect (а не navigate): гарантированно сбрасывает
        // React Query cache и перехватывает из pricing-page state.
        window.location.href = `/sign-in?redirect=${encodeURIComponent("/pricing")}`
        return
      }
      toast.error(err.message || "Не удалось начать оплату")
    },
  })
}

export function useCancelSubscription() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)
  return useMutation({
    mutationFn: (input?: CancelSubscriptionInput) => postCancelSubscription(input),
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

export function usePauseSubscription() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)
  return useMutation({
    mutationFn: (months: 1 | 2 | 3) => postPauseSubscription(months),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["subscription"] })
      qc.invalidateQueries({ queryKey: ["subscription", "usage"] })
      fetchMe()
      toast.success("Подписка приостановлена")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось приостановить подписку")
    },
  })
}

export function useResumeSubscription() {
  const qc = useQueryClient()
  const fetchMe = useAuthStore((s) => s.fetchMe)
  return useMutation({
    mutationFn: postResumeSubscription,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["subscription"] })
      qc.invalidateQueries({ queryKey: ["subscription", "usage"] })
      fetchMe()
      toast.success("Подписка возобновлена")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Не удалось возобновить подписку")
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
