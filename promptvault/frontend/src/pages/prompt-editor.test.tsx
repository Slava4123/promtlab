// Smoke test для PromptEditor page.
// Покрывает create-режим (params.id отсутствует) — самый простой путь без
// существующего prompt'а. Heavy components (PromptSplitEditor, FileImport*)
// мокаются как простые stubs.
import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"

import PromptEditor from "./prompt-editor"
import { renderWithProviders } from "@/test/render"

vi.mock("@/hooks/use-prompts", () => ({
  usePrompt: () => ({ data: undefined, isLoading: false }),
  useCreatePrompt: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useUpdatePrompt: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useIncrementUsage: () => ({ mutate: vi.fn() }),
  useDeletePrompt: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock("@/hooks/use-collections", () => ({
  useCollections: () => ({ data: [], isLoading: false }),
}))

vi.mock("@/stores/workspace-store", () => ({
  useWorkspaceStore: <T,>(sel: (s: { team: null }) => T) => sel({ team: null }),
}))

// Heavy components — render как stubs без своих deps
vi.mock("@/components/prompts/prompt-split-editor", () => ({
  PromptSplitEditor: () => <div data-testid="prompt-split-editor" />,
}))
vi.mock("@/components/prompts/file-import-button", () => ({
  FileImportButton: () => null,
}))
vi.mock("@/components/prompts/file-import-drop-zone", () => ({
  FileImportDropZone: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}))
vi.mock("@/components/prompts/collections-combobox", () => ({
  CollectionsCombobox: () => <div data-testid="collections-combobox" />,
}))
vi.mock("@/components/tags/tag-input", () => ({
  TagInput: () => <div data-testid="tag-input" />,
}))

describe("PromptEditor page", () => {
  it("рендерится в create-mode с заголовком", () => {
    renderWithProviders(<PromptEditor />, { route: "/prompts/new" })
    // Заголовок страницы — постоянный элемент в любом режиме
    expect(
      screen.getByRole("heading", { name: /новый промпт|редактирование/i }),
    ).toBeInTheDocument()
  })
})
