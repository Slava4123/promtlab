import { describe, it, expect } from "vitest"
import { buildNarrative } from "./analytics-narrative"
import type { PersonalDashboard, Insight } from "@/api/analytics"

const baseDashboard: PersonalDashboard = {
  range: "7d",
  usage_per_day: [],
  top_prompts: [],
  prompts_created_per_day: [],
  prompts_updated_per_day: [],
  share_views_per_day: [],
  top_shared: [],
  totals_current: { uses: 234, created: 12, updated: 0, share_views: 89 },
  totals_previous: { uses: 190, created: 10, updated: 0, share_views: 96 },
  usage_by_model: [
    { model: "claude-3-opus", uses: 145 },
    { model: "gpt-4", uses: 65 },
    { model: "gemini-pro", uses: 24 },
  ],
}

describe("buildNarrative", () => {
  it("includes period and delta in summary for non-zero uses", () => {
    const result = buildNarrative(baseDashboard, null)
    expect(result.summary).toContain("234")
    expect(result.summary).toMatch(/\+23%|↑23%/)
  })

  it("returns quiet copy for zero uses", () => {
    const empty: PersonalDashboard = {
      ...baseDashboard,
      totals_current: { uses: 0, created: 0, updated: 0, share_views: 0 },
      totals_previous: { uses: 0, created: 0, updated: 0, share_views: 0 },
    }
    const result = buildNarrative(empty, null)
    expect(result.summary).toMatch(/тих|пуст/i)
  })

  it("returns topModel as Claude with percentage when dominant", () => {
    const result = buildNarrative(baseDashboard, null)
    expect(result.topModel).toMatch(/Claude/i)
    expect(result.topModel).toMatch(/62/)
  })

  it("returns null topModel when usage_by_model is empty", () => {
    const empty: PersonalDashboard = { ...baseDashboard, usage_by_model: [] }
    const result = buildNarrative(empty, null)
    expect(result.topModel).toBeNull()
  })

  it("returns actionHint when insights contain unused_prompts and possible_duplicates", () => {
    const insights: Insight[] = [
      { type: "unused_prompts", payload: [1, 2, 3, 4, 5], computed_at: "" },
      { type: "possible_duplicates", payload: [1, 2], computed_at: "" },
    ]
    const result = buildNarrative(baseDashboard, insights)
    expect(result.actionHint).toMatch(/5/)
    expect(result.actionHint).toMatch(/2/)
  })

  it("returns null actionHint for empty insights", () => {
    const result = buildNarrative(baseDashboard, [])
    expect(result.actionHint).toBeNull()
  })
})
