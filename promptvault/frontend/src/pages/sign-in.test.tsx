// MN-14 — smoke render test для sign-in page.
// Не deep behavior (login flow), а базовое: страница рендерится без crash + основные UI элементы есть.
import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"

import SignIn from "./sign-in"
import { renderWithProviders } from "@/test/render"

// Мокаем auth-store — sign-in вызывает login() из store; в smoke тесте нам важна только
// рендер-часть, не реальный flow.
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: <T,>(sel: (s: {
    login: () => Promise<{ kind: "ok" }>
    verifyTOTP: () => Promise<unknown>
  }) => T) =>
    sel({
      login: vi.fn(),
      verifyTOTP: vi.fn(),
    }),
}))

describe("SignIn page", () => {
  it("рендерится без ошибок и показывает форму входа", () => {
    renderWithProviders(<SignIn />, { route: "/sign-in" })
    // Основные элементы — email input, password input, submit-кнопка.
    expect(screen.getByPlaceholderText(/example/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/пароль/i, { selector: "input" })).toBeInTheDocument()
  })

  it("показывает ссылку на регистрацию", () => {
    renderWithProviders(<SignIn />, { route: "/sign-in" })
    // Ссылка «Создать аккаунт» / «Зарегистрироваться» — критичный CTA.
    const links = screen.getAllByRole("link")
    const hasSignupLink = links.some((l) =>
      /регистр|создать|аккаунт/i.test(l.textContent || ""),
    )
    expect(hasSignupLink).toBe(true)
  })
})
