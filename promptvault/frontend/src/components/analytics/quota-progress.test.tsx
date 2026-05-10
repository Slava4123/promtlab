import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { QuotaProgress } from "./quota-progress"

afterEach(() => cleanup())

// Регрессии для quota-progress: малые значения должны иметь видимый штрих,
// overflow помечается badge «Превышено», amber-цвет включается на 80%+.
// Если что-то из этого ломается, юзер либо не видит, что у него есть
// расход (тонкая полоска незаметна), либо не понимает, что превысил лимит.

describe("QuotaProgress", () => {
  it("рендерит title и formatted value (1 / 500)", () => {
    render(<QuotaProgress title="Промпты" quota={{ used: 1, limit: 500 }} />)
    expect(screen.getByText("Промпты")).toBeInTheDocument()
    expect(screen.getByText("1 / 500")).toBeInTheDocument()
  })

  it("0 / 0 — не падает, нет «Превышено», нет amber", () => {
    render(<QuotaProgress title="Лимит" quota={{ used: 0, limit: 0 }} />)
    expect(screen.queryByText(/Превышено/i)).toBeNull()
    const valueSpan = screen.getByText("0 / 0")
    expect(valueSpan.className).not.toMatch(/amber|rose/)
  })

  it("0 / 500 — нет минимального штриха (used=0)", () => {
    const { container } = render(
      <QuotaProgress title="Промпты" quota={{ used: 0, limit: 500 }} />,
    )
    const indicator = container.querySelector("[data-slot='progress-indicator']")
    expect(indicator).not.toBeNull()
    // base-ui хранит value в data-aria-valuenow на корне. При used=0 это 0.
    const root = container.querySelector("[data-slot='progress']")
    expect(root?.getAttribute("aria-valuenow")).toBe("0")
  })

  it("1 / 500 — минимальный штрих 2% даже при pct=0.2%", () => {
    const { container } = render(
      <QuotaProgress title="Промпты" quota={{ used: 1, limit: 500 }} />,
    )
    const root = container.querySelector("[data-slot='progress']")
    expect(root?.getAttribute("aria-valuenow")).toBe("2")
  })

  it("80 / 100 — amber, без «Превышено»", () => {
    render(<QuotaProgress title="Лимит" quota={{ used: 80, limit: 100 }} />)
    expect(screen.queryByText(/Превышено/i)).toBeNull()
    const valueSpan = screen.getByText("80 / 100")
    expect(valueSpan.className).toMatch(/amber/)
  })

  it("100 / 100 — amber, без «Превышено» (ровно на лимите)", () => {
    render(<QuotaProgress title="Лимит" quota={{ used: 100, limit: 100 }} />)
    expect(screen.queryByText(/Превышено/i)).toBeNull()
    const valueSpan = screen.getByText("100 / 100")
    expect(valueSpan.className).toMatch(/amber/)
  })

  it("600 / 500 — badge «Превышено», rose-цвет, бар на 100%", () => {
    const { container } = render(
      <QuotaProgress title="Лимит" quota={{ used: 600, limit: 500 }} />,
    )
    expect(screen.getByText("Превышено")).toBeInTheDocument()
    const valueSpan = screen.getByText("600 / 500")
    expect(valueSpan.className).toMatch(/rose/)
    const root = container.querySelector("[data-slot='progress']")
    // Бар не уходит за 100% — clamp.
    expect(root?.getAttribute("aria-valuenow")).toBe("100")
  })

  it("кастомный format применяется", () => {
    render(
      <QuotaProgress
        title="Лимит"
        quota={{ used: 5, limit: 50 }}
        format={(u, l) => `${u} из ${l} в день`}
      />,
    )
    expect(screen.getByText("5 из 50 в день")).toBeInTheDocument()
  })
})
