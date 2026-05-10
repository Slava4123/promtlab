// Smoke test для Dashboard page.
// Проверяет что страница рендерится без crash при undefined data из всех hooks
// и что главный поисковый input виден сразу (даже в loading state).
import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"

import Dashboard from "./dashboard"
import { renderWithProviders } from "@/test/render"

vi.mock("@/hooks/use-prompts", () => ({
  usePrompts: () => ({
    data: undefined,
    isLoading: true,
    isFetchingNextPage: false,
    hasNextPage: false,
    fetchNextPage: vi.fn(),
  }),
  usePinnedPrompts: () => ({ data: undefined, error: null }),
  useRecentPrompts: () => ({ data: undefined, error: null }),
  useToggleFavorite: () => ({ mutate: vi.fn() }),
  useTogglePin: () => ({ mutate: vi.fn() }),
  useDeletePrompt: () => ({ isPending: false, mutate: vi.fn() }),
  useIncrementUsage: () => ({ mutate: vi.fn() }),
}))

vi.mock("@/hooks/use-trash", () => ({
  useRestoreItem: () => ({ mutate: vi.fn() }),
}))

vi.mock("@/hooks/use-tags", () => ({
  useTags: () => ({ data: undefined }),
}))

vi.mock("@/hooks/use-collections", () => ({
  useCollections: () => ({ data: undefined, isLoading: false }),
}))

vi.mock("@/hooks/use-streaks", () => ({
  useStreak: () => ({ data: undefined, isLoading: false }),
}))

vi.mock("@/stores/workspace-store", () => ({
  useWorkspaceStore: <T,>(sel: (s: { team: null }) => T) => sel({ team: null }),
}))

describe("Dashboard page", () => {
  it("рендерится с поисковой строкой в loading-состоянии", () => {
    renderWithProviders(<Dashboard />, { route: "/dashboard" })
    // Search input — главный CTA dashboard'а, виден всегда
    expect(
      screen.getByPlaceholderText(/поиск|поиск промптов/i),
    ).toBeInTheDocument()
  })
})
