import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { MetricCard } from "./metric-card"

// Регрессия на B.2 (compare previous period): DeltaBadge должен корректно
// различать рост/падение/zero/no-base. Если кто-то сломает маппинг —
// юзер увидит не ту индикацию «стало лучше / стало хуже».

afterEach(() => cleanup())

describe("MetricCard DeltaBadge", () => {
  it("рендерит title и value", () => {
    render(<MetricCard title="Использований" value={42} />)
    expect(screen.getByText("Использований")).toBeInTheDocument()
    expect(screen.getByText("42")).toBeInTheDocument()
  })

  it("положительная delta — ↑ со знаком +", () => {
    render(<MetricCard title="Uses" value={100} delta={12} />)
    expect(screen.getByText(/↑/)).toBeInTheDocument()
    expect(screen.getByText(/\+12%/)).toBeInTheDocument()
  })

  it("отрицательная delta — ↓ без знака +", () => {
    render(<MetricCard title="Uses" value={50} delta={-25} />)
    expect(screen.getByText(/↓/)).toBeInTheDocument()
    expect(screen.getByText(/-25%/)).toBeInTheDocument()
  })

  it("нулевая delta — ≡ 0%", () => {
    render(<MetricCard title="Uses" value={10} delta={0} />)
    expect(screen.getByText(/≡ 0%/)).toBeInTheDocument()
  })

  it("null delta (нет базы для сравнения) — тире", () => {
    render(<MetricCard title="Uses" value={5} delta={null} />)
    expect(screen.getByText("—")).toBeInTheDocument()
  })

  it("undefined delta — badge не рендерится вообще", () => {
    render(<MetricCard title="Uses" value={5} />)
    // Ни стрелок, ни «—», ни «0%».
    expect(screen.queryByText(/[↑↓≡]/)).toBeNull()
  })

  it("subtitle показывается рядом с delta", () => {
    render(<MetricCard title="Uses" value={50} subtitle="за 7 дней" delta={5} />)
    expect(screen.getByText("за 7 дней")).toBeInTheDocument()
    expect(screen.getByText(/\+5%/)).toBeInTheDocument()
  })
})
