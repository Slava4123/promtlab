import { render, screen } from "@testing-library/react"
import { describe, it, expect, vi, beforeEach } from "vitest"
import { MemoryRouter, Routes, Route } from "react-router-dom"
import PromptAnalyticsPage from "./prompt-analytics"

// Мок хуков; полная data-shape сложная (prompt + analytics), поэтому тест
// сознательно покрывает состояния loading / error — гарантируют, что
// компонент корректно обрабатывает разорванный fetch.
vi.mock("@/hooks/use-analytics", () => ({
  usePromptAnalytics: vi.fn(),
}))
vi.mock("@/hooks/use-prompts", () => ({
  usePrompt: vi.fn(),
}))

import { usePromptAnalytics } from "@/hooks/use-analytics"
import { usePrompt } from "@/hooks/use-prompts"

function renderWithRouter() {
  return render(
    <MemoryRouter initialEntries={["/prompt/42/analytics"]}>
      <Routes>
        <Route path="/prompt/:id/analytics" element={<PromptAnalyticsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe("PromptAnalyticsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("рендерит заголовок-страницы в loading-состоянии", () => {
    vi.mocked(usePrompt).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as ReturnType<typeof usePrompt>)
    vi.mocked(usePromptAnalytics).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as ReturnType<typeof usePromptAnalytics>)

    renderWithRouter()
    expect(screen.getByText(/Аналитика промпта/i)).toBeInTheDocument()
  })

  it("показывает ошибку если данные не загрузились", () => {
    vi.mocked(usePrompt).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as ReturnType<typeof usePrompt>)
    vi.mocked(usePromptAnalytics).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as ReturnType<typeof usePromptAnalytics>)

    renderWithRouter()
    expect(screen.getByText(/Не удалось загрузить/i)).toBeInTheDocument()
  })
})
