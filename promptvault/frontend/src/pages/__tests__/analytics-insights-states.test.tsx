import { render, screen } from "@testing-library/react"
import { describe, it, expect, vi, beforeEach } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router-dom"
import AnalyticsPage from "../analytics"
import type { PersonalDashboard, InsightsResponse } from "@/api/analytics"

// Pricing iteration v3 + Bento redesign (Task A14): three-state Smart Insights UI.
// Free → UpgradeGate Pro; Pro → InsightsPanel + 5 locked-карточек;
// Max → 7 insights без locked'ов.
//
// Стратегия мока: useInsights принимает isPaid и возвращает фиктивные
// данные с разным числом insight-типов в зависимости от плана. Это
// эмулирует backend, который per-type гейтит по plan_id (Task 5/6).

let mockPlanId: "free" | "pro" | "max" = "free"

// Мокаем все analytics-hooks, которые использует страница.
// usePersonalAnalytics — возвращает минимальный dashboard, чтобы render
// дошёл до Smart Insights блока (Free/Pro/Max теста).
vi.mock("@/hooks/use-analytics", () => {
  const emptyDashboard: PersonalDashboard = {
    range: "7d",
    usage_per_day: [],
    top_prompts: [],
    prompts_created_per_day: [],
    prompts_updated_per_day: [],
    share_views_per_day: [],
    top_shared: [],
    totals_current: { uses: 0, created: 0, updated: 0, share_views: 0 },
    totals_previous: { uses: 0, created: 0, updated: 0, share_views: 0 },
    usage_by_model: [],
  }

  // Pro юзер получает 2 типа от backend; Max — 7. Free → useInsights
  // не enabled, фейкаем data: undefined.
  const insightsByPlan = (): InsightsResponse | undefined => {
    if (mockPlanId === "max") {
      return {
        items: [
          { type: "unused_prompts", payload: [], computed_at: "" },
          { type: "trending", payload: [], computed_at: "" },
          { type: "declining", payload: [], computed_at: "" },
          { type: "most_edited", payload: [], computed_at: "" },
          { type: "possible_duplicates", payload: [], computed_at: "" },
          { type: "orphan_tags", payload: [], computed_at: "" },
          { type: "empty_collections", payload: [], computed_at: "" },
        ],
      }
    }
    if (mockPlanId === "pro") {
      return {
        items: [
          { type: "unused_prompts", payload: [], computed_at: "" },
          { type: "possible_duplicates", payload: [], computed_at: "" },
        ],
      }
    }
    return undefined
  }

  return {
    usePersonalAnalytics: () => ({
      data: emptyDashboard,
      isLoading: false,
      isError: false,
    }),
    useInsights: (isPaid: boolean) => ({
      data: isPaid ? insightsByPlan() : undefined,
      isLoading: false,
      isError: false,
    }),
    useRefreshInsights: () => ({
      mutateAsync: vi.fn(),
      isPending: false,
    }),
  }
})

// Task A13 добавил useStreak в analytics.tsx — мокаем, иначе TanStack Query
// уйдёт за реальным /streaks fetch и зависнет.
vi.mock("@/hooks/use-streaks", () => ({
  useStreak: () => ({ data: undefined, isLoading: false, isError: false }),
}))

vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (selector: (s: { user: { plan_id: string } }) => unknown) =>
    selector({ user: { plan_id: mockPlanId } }),
}))

function renderAnalytics() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <AnalyticsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe("Analytics insights — three states", () => {
  beforeEach(() => {
    mockPlanId = "free"
  })

  it("Free → UpgradeGate Pro (нет locked-карточек, нет InsightsPanel)", () => {
    mockPlanId = "free"
    renderAnalytics()
    // UpgradeGate с правильным заголовком — это invariant Free-state.
    expect(screen.getByText(/Подсказки — на тарифе Pro/i)).toBeInTheDocument()
    // Locked-карточки НЕ показываются Free-юзеру (новые названия после A11/A12).
    expect(screen.queryByText("Растёт")).not.toBeInTheDocument()
    expect(screen.queryByText("Падает")).not.toBeInTheDocument()
    // «Доступно в Max» CTA из locked-card не должно быть на Free.
    expect(screen.queryByText(/Доступно в Max/i)).not.toBeInTheDocument()
  })

  it("Pro → InsightsPanel + 5 locked-карточек (нет UpgradeGate Pro)", () => {
    mockPlanId = "pro"
    renderAnalytics()
    // Нет Pro-teaser'а — Pro-юзер уже залогинен.
    expect(screen.queryByText(/Подсказки — на тарифе Pro/i)).not.toBeInTheDocument()
    // 5 locked-карточек для Max-only типов. Стрелка теперь icon (ArrowRight),
    // поэтому regex проверяет только текстовое начало «Доступно в Max».
    const lockedLinks = screen.getAllByText(/Доступно в Max/i)
    expect(lockedLinks).toHaveLength(5)
    // Заголовки 5 Max-only типов (новый набор после refactor insights-locked-card).
    expect(screen.getByText("Растёт")).toBeInTheDocument()
    expect(screen.getByText("Падает")).toBeInTheDocument()
    expect(screen.getByText("Часто правят")).toBeInTheDocument()
    expect(screen.getByText("Теги без промптов")).toBeInTheDocument()
    expect(screen.getByText("Пустые коллекции")).toBeInTheDocument()
  })

  it("Max → 7 insights в InsightsPanel, нет locked-карточек", () => {
    mockPlanId = "max"
    renderAnalytics()
    // Нет Pro-teaser'а.
    expect(screen.queryByText(/Подсказки — на тарифе Pro/i)).not.toBeInTheDocument()
    // НЕТ locked-карточек — Max видит все 7 типов в полном объёме.
    expect(screen.queryByText(/Доступно в Max/i)).not.toBeInTheDocument()
  })
})
