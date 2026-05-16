import { describe, it, expect, vi, beforeEach } from "vitest"
import { cleanup, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { RecurrentConsent } from "./recurrent-consent"
import type { Plan } from "@/api/types"

// Минимальный fixture: тестируемый компонент использует только
// id / price_kop / period_days. Остальные поля Plan заполнены нулями,
// явно через `as Plan` чтобы TS не требовал каждое поле в тестах.
function makePlan(overrides: Partial<Plan> = {}): Plan {
  return {
    id: "pro",
    name: "Pro",
    price_kop: 59900,
    period_days: 30,
    max_prompts: 0,
    max_collections: 0,
    max_teams: 0,
    max_team_members: 0,
    max_ext_uses_daily: 0,
    max_mcp_uses_daily: 0,
    max_chains: 0,
    max_steps_per_chain: 0,
    max_saved_executions: 0,
    max_team_prompts: 0,
    max_team_collections: 0,
    max_team_chains: 0,
    features: [],
    sort_order: 0,
    is_active: true,
    ...overrides,
  } as Plan
}

describe("RecurrentConsent", () => {
  beforeEach(() => cleanup())

  it("показывает сумму и период списания для monthly-плана", () => {
    render(<RecurrentConsent plan={makePlan()} checked={false} onChange={vi.fn()} />)
    expect(screen.getByText(/599\s*₽/)).toBeInTheDocument()
    // «месяц» в тексте «за каждый месяц»
    expect(screen.getByText(/месяц/i)).toBeInTheDocument()
  })

  it("для yearly-плана показывает период «год»", () => {
    const yearly = makePlan({ id: "pro_yearly", price_kop: 647_00 * 12, period_days: 365 })
    render(<RecurrentConsent plan={yearly} checked={false} onChange={vi.fn()} />)
    expect(screen.getByText(/год/i)).toBeInTheDocument()
  })

  it("вызывает onChange(true) при клике на чек-бокс", async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<RecurrentConsent plan={makePlan()} checked={false} onChange={onChange} />)
    await user.click(screen.getByRole("checkbox"))
    expect(onChange).toHaveBeenCalledWith(true)
  })

  it("отражает checked=true состояние", () => {
    render(<RecurrentConsent plan={makePlan()} checked={true} onChange={vi.fn()} />)
    expect(screen.getByRole("checkbox")).toBeChecked()
  })

  it("содержит ссылки на оферту и страницу управления подпиской", () => {
    render(<RecurrentConsent plan={makePlan()} checked={false} onChange={vi.fn()} />)
    expect(screen.getByRole("link", { name: /оферт/i })).toHaveAttribute("href", "/legal/offer")
    expect(screen.getByRole("link", { name: /настройки/i })).toHaveAttribute("href", "/settings/subscription")
  })

  it("каждый чек-бокс получает уникальный htmlFor при разных idSuffix", () => {
    const { container } = render(
      <>
        <RecurrentConsent plan={makePlan()} checked={false} onChange={vi.fn()} idSuffix="pricing" />
        <RecurrentConsent plan={makePlan()} checked={false} onChange={vi.fn()} idSuffix="quota" />
      </>,
    )
    const inputs = container.querySelectorAll("input[type=checkbox]")
    expect(inputs).toHaveLength(2)
    expect(inputs[0].id).not.toBe(inputs[1].id)
  })
})
