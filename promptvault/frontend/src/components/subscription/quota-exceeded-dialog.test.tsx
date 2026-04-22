import { describe, it, expect, beforeEach, vi } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"

import { QuotaExceededDialog } from "./quota-exceeded-dialog"
import { useQuotaStore } from "@/stores/quota-store"

// Регрессия-тест на BUG #3 (QA-сессия 2026-04-22):
// Модал показывал сырой feature_type (`daily_shares`) вместо русского лейбла.
// quotaLabels не покрывал все ключи, которые шлёт backend — fallback
// отдавал raw key.
//
// Этот тест гарантирует: каждый известный backend-ключ имеет русский label
// и benefit в QuotaExceededDialog.

// Хуки checkout и useAuthStore подтягивают react-query и cookie — замокаем
// минимально, чтобы rendering не падал.
vi.mock("@/hooks/use-subscription", () => ({
  useCheckout: () => ({ mutateAsync: vi.fn(), isPending: false }),
}))
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: <T,>(sel: (s: { user: { plan_id: string } }) => T) =>
    sel({ user: { plan_id: "free" } }),
}))

function renderDialog() {
  return render(
    <MemoryRouter>
      <QuotaExceededDialog />
    </MemoryRouter>,
  )
}

describe("QuotaExceededDialog localization", () => {
  beforeEach(() => {
    useQuotaStore.getState().dismiss()
    cleanup()
  })

  // Известные backend-ключи: prompts, collections, teams, team_members,
  // share_links (legacy), daily_shares (Phase 14), ext_daily, mcp_daily.
  // Если бэкенд добавит новый feature_type, этот список нужно расширить —
  // вместе с quotaLabels в самом компоненте.
  const cases: Array<{ quotaType: string; expectSubstring: string }> = [
    { quotaType: "prompts", expectSubstring: "Лимит промптов" },
    { quotaType: "collections", expectSubstring: "Лимит коллекций" },
    { quotaType: "teams", expectSubstring: "Лимит команд" },
    { quotaType: "team_members", expectSubstring: "Лимит участников команды" },
    { quotaType: "share_links", expectSubstring: "Лимит публичных ссылок" },
    { quotaType: "daily_shares", expectSubstring: "Лимит публичных ссылок в день" },
    { quotaType: "ext_daily", expectSubstring: "Лимит вставок" },
    { quotaType: "mcp_daily", expectSubstring: "Лимит MCP-вызовов" },
  ]

  cases.forEach(({ quotaType, expectSubstring }) => {
    it(`рендерит русский label для ${quotaType}`, () => {
      useQuotaStore.getState().show({
        quotaType,
        message: "msg",
        used: 10,
        limit: 10,
        plan: "free",
      })
      renderDialog()
      // Заголовок: `Лимит <ресурс>: used/limit`
      expect(screen.getByRole("heading")).toHaveTextContent(expectSubstring)
      // Не должно быть сырого ключа в тексте (кроме случая где expectSubstring сам его содержит).
      if (!expectSubstring.toLowerCase().includes(quotaType.toLowerCase())) {
        expect(screen.getByRole("heading")).not.toHaveTextContent(quotaType)
      }
    })
  })

  it("не падает на неизвестный quotaType и показывает его как fallback", () => {
    useQuotaStore.getState().show({
      quotaType: "future_unknown_quota",
      message: "",
      used: 1,
      limit: 1,
      plan: "free",
    })
    renderDialog()
    // Fallback на raw key — не красиво, но не должно падать.
    expect(screen.getByRole("heading")).toHaveTextContent(/Лимит/)
  })
})
