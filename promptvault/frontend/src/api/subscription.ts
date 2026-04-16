import { api, publicApi } from "./client"
import type {
  CancelReason,
  Plan,
  Subscription,
  UsageSummary,
  CheckoutResponse,
} from "./types"

export function fetchPlans(): Promise<Plan[]> {
  return publicApi<Plan[]>("/plans")
}

export function fetchSubscription(): Promise<Subscription | null> {
  return api<Subscription | null>("/subscription")
}

export function fetchUsage(): Promise<UsageSummary> {
  return api<UsageSummary>("/subscription/usage")
}

export function postCheckout(planId: string): Promise<CheckoutResponse> {
  return api<CheckoutResponse>("/subscription/checkout", {
    method: "POST",
    body: JSON.stringify({ plan_id: planId }),
  })
}

export interface CancelSubscriptionInput {
  reason?: CancelReason
  other_text?: string
}

export function postCancelSubscription(input?: CancelSubscriptionInput): Promise<void> {
  const body = input?.reason ? JSON.stringify({
    reason: input.reason,
    other_text: input.reason === "other" ? input.other_text ?? "" : undefined,
  }) : undefined
  return api<void>("/subscription/cancel", { method: "POST", body })
}

export function postPauseSubscription(months: 1 | 2 | 3): Promise<void> {
  return api<void>("/subscription/pause", {
    method: "POST",
    body: JSON.stringify({ months }),
  })
}

export function postResumeSubscription(): Promise<void> {
  return api<void>("/subscription/resume", { method: "POST" })
}

export function postDowngrade(): Promise<void> {
  return api<void>("/subscription/downgrade", { method: "POST" })
}

export function postSetAutoRenew(autoRenew: boolean): Promise<void> {
  return api<void>("/subscription/auto-renew", {
    method: "POST",
    body: JSON.stringify({ auto_renew: autoRenew }),
  })
}
