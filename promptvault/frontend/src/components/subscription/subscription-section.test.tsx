// Smoke test для SubscriptionSection.
// Проверяет что секция рендерится с loading-skeleton'ом без crash, и что
// заголовок «Подписка» виден сразу — независимо от состояния hooks.
import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"

import { SubscriptionSection } from "./subscription-section"
import { renderWithProviders } from "@/test/render"

vi.mock("react-router", async (importActual) => {
  const actual = await importActual<typeof import("react-router")>()
  return { ...actual, useNavigate: () => vi.fn() }
})

vi.mock("@/hooks/use-subscription", () => ({
  useSubscription: () => ({ data: undefined, isLoading: true }),
  useUsage: () => ({ data: undefined, isLoading: true }),
  useCancelSubscription: () => ({ mutate: vi.fn(), isPending: false }),
  usePauseSubscription: () => ({ mutate: vi.fn(), isPending: false }),
  useResumeSubscription: () => ({ mutate: vi.fn(), isPending: false }),
  useSetAutoRenew: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock("@/stores/auth-store", () => ({
  useAuthStore: <T,>(sel: (s: { user: { plan_id: string } | null }) => T) =>
    sel({ user: { plan_id: "free" } }),
}))

describe("SubscriptionSection", () => {
  it("рендерится с заголовком «Подписка» в loading-состоянии", () => {
    renderWithProviders(<SubscriptionSection />)
    expect(
      screen.getByRole("heading", { name: /подписка/i }),
    ).toBeInTheDocument()
  })
})
