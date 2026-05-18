import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import Collections from "./collections"
import * as collectionsHooks from "@/hooks/use-collections"
import * as emptyHooks from "@/hooks/use-empty-collections"
import * as teamHooks from "@/hooks/use-current-team"

vi.mock("@/hooks/use-collections")
vi.mock("@/hooks/use-empty-collections")
vi.mock("@/hooks/use-current-team")

function wrap(node: ReactNode, initial: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter initialEntries={[initial]}>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("CollectionsPage filter=empty", () => {
  beforeEach(() => {
    cleanup()
    // useCurrentTeam → personal workspace.
    vi.mocked(teamHooks.useCurrentTeam).mockReturnValue(null)
    // Mutations: noop. Page вызывает их во время render (для disabled-логики
    // в Save), без stub'а vi.mock возвращает undefined → TypeError.
    vi.mocked(collectionsHooks.useCreateCollection).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof collectionsHooks.useCreateCollection>)
    vi.mocked(collectionsHooks.useUpdateCollection).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof collectionsHooks.useUpdateCollection>)
    vi.mocked(collectionsHooks.useDeleteCollection).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof collectionsHooks.useDeleteCollection>)
    // Default-stub чтобы default-кейс (без ?filter=empty) не падал.
    vi.mocked(collectionsHooks.useCollections).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof collectionsHooks.useCollections>)
    vi.mocked(emptyHooks.useEmptyCollections).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof emptyHooks.useEmptyCollections>)
  })

  it("default shows all collections", () => {
    vi.mocked(collectionsHooks.useCollections).mockReturnValue({
      data: [
        { id: 1, name: "Активная", description: "", color: "#a78bfa", icon: "folder", prompt_count: 3 },
        { id: 2, name: "Заброшенная", description: "", color: "#fbbf24", icon: "folder", prompt_count: 0 },
      ],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof collectionsHooks.useCollections>)
    render(wrap(<Collections />, "/collections"))
    expect(screen.getByText("Активная")).toBeInTheDocument()
    expect(screen.getByText("Заброшенная")).toBeInTheDocument()
  })

  it("?filter=empty shows only empty collections + filter description", () => {
    vi.mocked(emptyHooks.useEmptyCollections).mockReturnValue({
      data: [{ id: 9, name: "Заброшенная" }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof emptyHooks.useEmptyCollections>)
    render(wrap(<Collections />, "/collections?filter=empty"))
    expect(screen.getByText("Заброшенная")).toBeInTheDocument()
    // Хедер/описание filter-режима — должно быть хотя бы одно упоминание
    // «без активных промптов» (заголовок или описание).
    expect(screen.getAllByText(/без активных промптов/i).length).toBeGreaterThan(0)
  })
})
