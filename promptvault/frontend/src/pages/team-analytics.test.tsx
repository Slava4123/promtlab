import { render, screen } from "@testing-library/react"
import { describe, it, expect, vi, beforeEach } from "vitest"
import { MemoryRouter, Routes, Route } from "react-router-dom"
import TeamAnalyticsPage from "./team-analytics"

// Полная data-shape TeamDashboard содержит множество вложенных полей;
// тестируем «loading → есть заголовок команды» и «Free юзер» (UpgradeGate),
// которые обеспечивают ключевые UX-контракты без знания всей data.
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}))
vi.mock("@/hooks/use-teams", () => ({
  useTeam: vi.fn(),
}))
vi.mock("@/hooks/use-analytics", () => ({
  useTeamAnalytics: vi.fn(),
}))

import { useAuthStore } from "@/stores/auth-store"
import { useTeam } from "@/hooks/use-teams"
import { useTeamAnalytics } from "@/hooks/use-analytics"

function renderWithRouter() {
  return render(
    <MemoryRouter initialEntries={["/teams/acme/analytics"]}>
      <Routes>
        <Route path="/teams/:slug/analytics" element={<TeamAnalyticsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe("TeamAnalyticsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("Free-юзер видит заголовок команды (upgrade-gate блокирует метрики)", () => {
    vi.mocked(useAuthStore).mockImplementation(((selector: (s: unknown) => unknown) =>
      selector({ user: { plan_id: "free" } })) as never)
    vi.mocked(useTeam).mockReturnValue({
      data: { id: 10, slug: "acme", name: "Acme" },
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useTeam>)
    vi.mocked(useTeamAnalytics).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useTeamAnalytics>)

    renderWithRouter()
    expect(screen.getByRole("heading", { name: /Аналитика команды/i, level: 1 })).toBeInTheDocument()
  })

  it("Loading-состояние рендерится без падения", () => {
    vi.mocked(useAuthStore).mockImplementation(((selector: (s: unknown) => unknown) =>
      selector({ user: { plan_id: "max" } })) as never)
    vi.mocked(useTeam).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as unknown as ReturnType<typeof useTeam>)
    vi.mocked(useTeamAnalytics).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as ReturnType<typeof useTeamAnalytics>)

    // Падает ли рендер при loading? — не должен.
    renderWithRouter()
  })
})
