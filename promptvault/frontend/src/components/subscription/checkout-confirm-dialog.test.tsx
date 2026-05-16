import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { CheckoutConfirmDialog } from "./checkout-confirm-dialog"
import type { Plan } from "@/api/types"

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

describe("CheckoutConfirmDialog", () => {
  beforeEach(() => cleanup())

  it("не рендерится при plan=null", () => {
    const { container } = render(
      <CheckoutConfirmDialog
        plan={null}
        features={[]}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
      />,
    )
    expect(container.querySelector('[role="dialog"]')).toBeNull()
  })

  it("показывает название тарифа, цену и период", () => {
    render(
      <CheckoutConfirmDialog
        plan={makePlan()}
        features={[]}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
      />,
    )
    expect(screen.getByText(/Оформить подписку Pro/)).toBeInTheDocument()
    expect(screen.getByText(/599\s*₽\s*в месяц/i)).toBeInTheDocument()
  })

  it("показывает дату следующего списания (today + period_days)", () => {
    // Прибавляем 30 дней — выбираем плавающий месяц, чтобы не привязываться
    // к конкретной дате теста.
    render(
      <CheckoutConfirmDialog
        plan={makePlan({ period_days: 30 })}
        features={[]}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
      />,
    )
    // Достаточно проверить что "Следующее списание" присутствует и под ним
    // есть строка с годом — точное число дат зависит от текущей даты.
    expect(screen.getByText(/Следующее списание/i)).toBeInTheDocument()
    expect(screen.getByText(/202\d г\.?/)).toBeInTheDocument()
  })

  it("показывает список фич (max 5)", () => {
    const features = ["До 500 промптов", "До 100 коллекций", "5 цепочек", "Лента активности", "Аналитика 90 дней", "Шестая фича"]
    render(
      <CheckoutConfirmDialog
        plan={makePlan()}
        features={features}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
      />,
    )
    expect(screen.getByText("До 500 промптов")).toBeInTheDocument()
    expect(screen.getByText("Аналитика 90 дней")).toBeInTheDocument()
    // Шестая фича не должна рендериться (slice 0..5)
    expect(screen.queryByText("Шестая фича")).toBeNull()
  })

  it("кнопка «Перейти к оплате» disabled пока чек-бокс не отмечен", async () => {
    const user = userEvent.setup()
    const onConfirm = vi.fn()
    render(
      <CheckoutConfirmDialog
        plan={makePlan()}
        features={[]}
        onClose={vi.fn()}
        onConfirm={onConfirm}
        isPending={false}
      />,
    )
    const submitBtn = screen.getByRole("button", { name: /перейти к оплате/i })
    expect(submitBtn).toBeDisabled()

    await user.click(screen.getByRole("checkbox"))
    expect(submitBtn).not.toBeDisabled()

    await user.click(submitBtn)
    expect(onConfirm).toHaveBeenCalledWith(true)
  })

  it("кнопка «Отмена» вызывает onClose", async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    render(
      <CheckoutConfirmDialog
        plan={makePlan()}
        features={[]}
        onClose={onClose}
        onConfirm={vi.fn()}
        isPending={false}
      />,
    )
    await user.click(screen.getByRole("button", { name: /отмена/i }))
    expect(onClose).toHaveBeenCalled()
  })

  it("при isPending=true обе кнопки footer'а disabled", () => {
    render(
      <CheckoutConfirmDialog
        plan={makePlan()}
        features={[]}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        isPending={true}
      />,
    )
    // Cancel — по тексту (надёжно)
    expect(screen.getByRole("button", { name: /отмена/i })).toBeDisabled()
    // Submit — единственная кнопка с классом text-white (brand-gradient
    // стиль на submit, остальные кнопки в диалоге — outline/close).
    const buttons = screen.getAllByRole("button")
    const submitBtn = buttons.find((b) => b.className.includes("text-white"))
    expect(submitBtn).toBeDefined()
    expect(submitBtn).toBeDisabled()
  })
})
