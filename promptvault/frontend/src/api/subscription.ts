import { api, publicApi } from "./client"
import type {
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

export function postCancelSubscription(): Promise<void> {
  return api<void>("/subscription/cancel", { method: "POST" })
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
