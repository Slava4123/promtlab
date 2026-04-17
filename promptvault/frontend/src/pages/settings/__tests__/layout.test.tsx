import { describe, it, expect, beforeEach, vi } from "vitest"
import { render, screen, within } from "@testing-library/react"
import { createMemoryRouter, RouterProvider, Navigate } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

import SettingsLayout from "../layout"
import SettingsProfile from "../profile"
import SettingsSecurity from "../security"
import SettingsAccounts from "../accounts"
import SettingsSubscription from "../subscription"
import SettingsReferral from "../referral"
import SettingsIntegrations from "../integrations"
import SettingsAppearance from "../appearance"
import { NAV_ITEMS } from "../_nav-config"
import { useAuthStore } from "@/stores/auth-store"

vi.mock("@/api/client", () => ({
  api: vi.fn().mockResolvedValue([]),
  ApiError: class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
    }
  },
}))

vi.mock("@/components/subscription/subscription-section", () => ({
  SubscriptionSection: () => <div data-testid="subscription-section">subscription</div>,
}))
vi.mock("@/components/settings/referral-section", () => ({
  ReferralSection: () => <div data-testid="referral-section">referral</div>,
}))
vi.mock("@/components/settings/extension-promo-section", () => ({
  ExtensionPromoSection: () => <div data-testid="extension-section">extension</div>,
}))
vi.mock("@/components/settings/api-keys-section", () => ({
  APIKeysSection: () => <div data-testid="api-keys-section">api-keys</div>,
}))

const TEST_USER = {
  id: 1,
  email: "test@example.com",
  name: "Test User",
  username: "",
  avatar_url: undefined,
  email_verified: true,
  has_password: true,
  default_model: "anthropic/claude-sonnet-4",
  plan_id: "free",
  role: "user",
  status: "active",
  has_unread_changelog: false,
} as const

function renderAt(initialPath: string) {
  // Замена роутерной части App.tsx — проверяем именно конфиг nested routes
  const routes = [
    {
      path: "/settings",
      element: <SettingsLayout />,
      children: [
        { index: true, element: <Navigate to="profile" replace /> },
        { path: "profile", element: <SettingsProfile /> },
        { path: "security", element: <SettingsSecurity /> },
        { path: "accounts", element: <SettingsAccounts /> },
        { path: "subscription", element: <SettingsSubscription /> },
        { path: "referral", element: <SettingsReferral /> },
        { path: "integrations", element: <SettingsIntegrations /> },
        { path: "appearance", element: <SettingsAppearance /> },
      ],
    },
  ]
  const router = createMemoryRouter(routes, { initialEntries: [initialPath] })
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe("SettingsLayout", () => {
  beforeEach(() => {
    useAuthStore.setState({ user: { ...TEST_USER }, isLoading: false })
  })

  it("рендерит все пункты nav из NAV_ITEMS", () => {
    renderAt("/settings/profile")
    const nav = screen.getByRole("navigation", { name: /разделы настроек/i })
    expect(nav).toBeInTheDocument()
    for (const item of NAV_ITEMS) {
      // Каждый пункт встречается дважды (desktop + mobile вариант), достаточно getAllByRole
      expect(within(nav).getAllByRole("link", { name: new RegExp(item.title, "i") }).length).toBeGreaterThan(0)
    }
  })

  it("на /settings/security ссылка 'Безопасность' помечена aria-current=page", () => {
    renderAt("/settings/security")
    const links = screen.getAllByRole("link", { name: /безопасность/i })
    expect(links.some((l) => l.getAttribute("aria-current") === "page")).toBe(true)
  })

  it("на /settings (index) редиректит на /settings/profile", () => {
    renderAt("/settings")
    // Index → Navigate to=profile → активным становится Профиль
    const links = screen.getAllByRole("link", { name: /профиль/i })
    expect(links.some((l) => l.getAttribute("aria-current") === "page")).toBe(true)
    // И заголовок sub-страницы Профиля виден
    expect(screen.getByRole("heading", { level: 2, name: /профиль/i })).toBeInTheDocument()
  })

  it("Outlet монтирует sub-страницу: /settings/appearance показывает заголовок Оформление", () => {
    renderAt("/settings/appearance")
    expect(screen.getByRole("heading", { level: 2, name: /оформление/i })).toBeInTheDocument()
  })

  it("если user=null — рендерит null (не падает)", () => {
    useAuthStore.setState({ user: null, isLoading: false })
    const { container } = renderAt("/settings/profile")
    // Layout возвращает null → пустой контейнер
    expect(container.querySelector("nav")).toBeNull()
  })
})
