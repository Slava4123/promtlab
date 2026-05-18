import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { CompactQuotas } from "./compact-quotas"
import type { UsageSummary } from "@/api/analytics"

afterEach(() => cleanup())

describe("CompactQuotas", () => {
  const baseQuota: UsageSummary = {
    plan_id: "pro",
    prompts: { used: 230, limit: 500 },
    collections: { used: 30, limit: 100 },
    teams: { used: 1, limit: 5 },
    ext_uses_today: { used: 0, limit: 50 },
    mcp_uses_today: { used: 0, limit: 50 },
  }

  it("renders prompts/collections/mcp usage", () => {
    render(<CompactQuotas quotas={baseQuota} />)
    expect(screen.getByText(/230/)).toBeInTheDocument()
    expect(screen.getByText(/500/)).toBeInTheDocument()
    expect(screen.getByText(/30 \/ 100/)).toBeInTheDocument()
    expect(screen.getByText(/100/)).toBeInTheDocument()
  })

  it("renders nothing when quotas is undefined", () => {
    const { container } = render(<CompactQuotas quotas={undefined} />)
    expect(container.firstChild).toBeNull()
  })
})
