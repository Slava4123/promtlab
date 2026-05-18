import { describe, it, expect } from "vitest"
import { formatDayShort } from "./date-format"

describe("formatDayShort", () => {
  it("formats ISO date as 'D MMM' in Russian", () => {
    expect(formatDayShort("2026-05-07")).toBe("7 мая")
    expect(formatDayShort("2026-05-16")).toBe("16 мая")
    // Intl ru-RU short month: "1 дек." with trailing period for short months
    expect(formatDayShort("2026-12-01")).toMatch(/^1 дек/)
  })

  it("returns empty string for invalid input", () => {
    expect(formatDayShort("")).toBe("")
    expect(formatDayShort("not-a-date")).toBe("")
  })

  it("handles full ISO timestamps", () => {
    expect(formatDayShort("2026-05-07T12:34:56Z")).toBe("7 мая")
  })
})
