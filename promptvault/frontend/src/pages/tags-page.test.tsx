import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import TagsPage from "./tags-page"
import * as tagHooks from "@/hooks/use-tags"
import * as orphanHooks from "@/hooks/use-orphan-tags"

vi.mock("@/hooks/use-tags")
vi.mock("@/hooks/use-orphan-tags")

function wrap(node: ReactNode, initial: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter initialEntries={[initial]}>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("TagsPage", () => {
  beforeEach(() => {
    cleanup()
    vi.mocked(tagHooks.useDeleteTag).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof tagHooks.useDeleteTag>)
    // Default-stub orphan hook так, чтобы default-кейс (без ?filter=orphan)
    // не падал на `data.length` если страница случайно сделает unconditional
    // вызов хука (vi.mock возвращает undefined без explicit ReturnValue).
    vi.mocked(orphanHooks.useOrphanTags).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof orphanHooks.useOrphanTags>)
    vi.mocked(tagHooks.useTags).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof tagHooks.useTags>)
  })

  it("default shows all tags", () => {
    vi.mocked(tagHooks.useTags).mockReturnValue({
      data: [
        { id: 1, name: "feature", color: "#fff" },
        { id: 2, name: "old", color: "#fff" },
      ],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof tagHooks.useTags>)
    render(wrap(<TagsPage />, "/tags"))
    expect(screen.getByText("feature")).toBeInTheDocument()
    expect(screen.getByText("old")).toBeInTheDocument()
  })

  it("?filter=orphan shows only orphan tags + filter description", () => {
    vi.mocked(orphanHooks.useOrphanTags).mockReturnValue({
      data: [{ id: 2, name: "old", color: "#fff" }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof orphanHooks.useOrphanTags>)
    render(wrap(<TagsPage />, "/tags?filter=orphan"))
    // Используем getAllByText т.к. фраза «без активных промптов» появляется
    // и в заголовке («Теги без активных промптов»), и в описании
    // («…не привязаны ни к одному активному промпту…») — оба валидны.
    expect(screen.getAllByText(/без активных промптов/i).length).toBeGreaterThan(0)
    expect(screen.getByText("old")).toBeInTheDocument()
  })
})
