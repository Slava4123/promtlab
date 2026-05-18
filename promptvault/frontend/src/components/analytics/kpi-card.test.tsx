import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { Activity } from "lucide-react"
import { KpiCard } from "./kpi-card"

afterEach(() => cleanup())

describe("KpiCard", () => {
  it("renders label, value, and icon", () => {
    render(<KpiCard label="Использования" value={234} icon={Activity} />)
    expect(screen.getByText("Использования")).toBeInTheDocument()
    expect(screen.getByText("234")).toBeInTheDocument()
  })

  it("shows up arrow for positive delta", () => {
    const { container } = render(
      <KpiCard label="X" value={100} delta={23} icon={Activity} />,
    )
    expect(screen.getByText(/23%/)).toBeInTheDocument()
    expect(container.querySelector(".text-emerald-600, .dark\\:text-emerald-400")).not.toBeNull()
  })

  it("shows down arrow for negative delta", () => {
    const { container } = render(
      <KpiCard label="X" value={100} delta={-8} icon={Activity} />,
    )
    expect(screen.getByText(/8%/)).toBeInTheDocument()
    expect(container.querySelector(".text-rose-600, .dark\\:text-rose-400")).not.toBeNull()
  })

  it("renders sparkline when points provided", () => {
    const { container } = render(
      <KpiCard label="X" value={100} sparkline={[1, 3, 5, 8]} icon={Activity} />,
    )
    expect(container.querySelector("svg")).not.toBeNull()
  })

  it("renders «—» when delta is null", () => {
    render(<KpiCard label="X" value={0} delta={null} icon={Activity} />)
    expect(screen.getByText("—")).toBeInTheDocument()
  })
})
