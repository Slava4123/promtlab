// Smoke test для chains/run page (Phase 16).
// Проверяет что страница рендерится с loading-состоянием — start mutation
// должен быть вызван автоматически через useEffect для chainID > 0.
import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"

import ChainRunPage from "./run"
import { renderWithProviders } from "@/test/render"

const mutateAsyncStart = vi.fn().mockResolvedValue({ id: 0 })

vi.mock("@/hooks/use-chains", () => ({
  useStartExecution: () => ({
    mutateAsync: mutateAsyncStart,
    isPending: false,
  }),
  useExecution: () => ({
    data: undefined,
    isLoading: true,
  }),
  useAdvanceStep: () => ({
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}))

describe("ChainRunPage", () => {
  it("рендерится в loading-состоянии без crash", () => {
    renderWithProviders(<ChainRunPage />, { route: "/chains/1/run" })
    // Заголовок страницы виден сразу — loading state показывает «Запуск цепочки»
    expect(screen.getByRole("heading", { name: /запуск цепочки/i })).toBeInTheDocument()
  })
})
