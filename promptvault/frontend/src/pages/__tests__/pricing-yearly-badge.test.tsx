import { render, screen, fireEvent } from "@testing-library/react"
import { describe, it, expect, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router-dom"
import Pricing from "../pricing"

// Mock plans с новыми ценами −20% после миграции 000073.
// Pro yearly 5750 ₽ vs monthly 599×12 = 7188 ₽ → (7188-5750)/7188 ≈ 20.0%.
const mockPlans = [
  { id: "free", name: "Free", price_kop: 0, period_days: 0, max_prompts: 25, max_collections: 3, max_teams: 1, max_team_members: 1, max_ext_uses_daily: 5, max_mcp_uses_daily: 5, max_chains: 0, max_steps_per_chain: 0, max_saved_executions: 0, max_team_prompts: 50, max_team_collections: 10, max_team_chains: 3, features: [], sort_order: 0, is_active: true },
  { id: "pro", name: "Pro", price_kop: 59900, period_days: 30, max_prompts: 500, max_collections: 100, max_teams: 5, max_team_members: 10, max_ext_uses_daily: 50, max_mcp_uses_daily: 50, max_chains: 5, max_steps_per_chain: 10, max_saved_executions: 50, max_team_prompts: 2000, max_team_collections: 400, max_team_chains: 20, features: [], sort_order: 1, is_active: true },
  { id: "pro_yearly", name: "Pro (год)", price_kop: 575000, period_days: 365, max_prompts: 500, max_collections: 100, max_teams: 5, max_team_members: 10, max_ext_uses_daily: 50, max_mcp_uses_daily: 50, max_chains: 5, max_steps_per_chain: 10, max_saved_executions: 50, max_team_prompts: 2000, max_team_collections: 400, max_team_chains: 20, features: [], sort_order: 2, is_active: true },
  { id: "max", name: "Max", price_kop: 129900, period_days: 30, max_prompts: 10000, max_collections: 1000, max_teams: 50, max_team_members: 50, max_ext_uses_daily: 500, max_mcp_uses_daily: 500, max_chains: 100, max_steps_per_chain: 50, max_saved_executions: 1000, max_team_prompts: 50000, max_team_collections: 5000, max_team_chains: 500, features: [], sort_order: 3, is_active: true },
  { id: "max_yearly", name: "Max (год)", price_kop: 1247000, period_days: 365, max_prompts: 10000, max_collections: 1000, max_teams: 50, max_team_members: 50, max_ext_uses_daily: 500, max_mcp_uses_daily: 500, max_chains: 100, max_steps_per_chain: 50, max_saved_executions: 1000, max_team_prompts: 50000, max_team_collections: 5000, max_team_chains: 500, features: [], sort_order: 4, is_active: true },
]

vi.mock("@/hooks/use-subscription", () => ({
  useCheckout: () => ({ mutate: vi.fn(), isPending: false }),
  useDowngrade: () => ({ mutate: vi.fn(), isPending: false }),
  useDowngradePreview: () => ({ data: null, isFetching: false, refetch: vi.fn() }),
  usePlans: () => ({ data: mockPlans, isLoading: false, error: null }),
}))

vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (selector: (s: { user: { plan_id: string; has_legacy_quotas: boolean } }) => unknown) =>
    selector({ user: { plan_id: "free", has_legacy_quotas: false } }),
}))

function renderPricing() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <Pricing />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe("Pricing yearly badge — dynamic discount", () => {
  it("показывает -20% для текущих цен (Pro yearly 5750 vs monthly 599×12=7188)", () => {
    renderPricing()
    // Yearly tab контролирует badge — кликаем чтобы убедиться, что
    // badge виден даже если default = monthly для free-юзера.
    const yearlyTab = screen.getByRole("tab", { name: /Ежегодно/i })
    fireEvent.click(yearlyTab)

    // После клика — badge должен показать "−20%" (динамика, не хардкод "−10%").
    expect(screen.getByText("−20%")).toBeInTheDocument()
  })
})
