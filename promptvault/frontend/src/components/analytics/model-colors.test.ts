import { describe, it, expect } from "vitest"
import { colorFor, labelFor, MODEL_COLORS, DEFAULT_COLOR, UNKNOWN_MODEL_HINT } from "./model-colors"

describe("model-colors", () => {
  it("matches Claude variants to Anthropic orange", () => {
    expect(colorFor("claude-3-opus")).toBe("#cc7a3e")
    expect(colorFor("Claude")).toBe("#cc7a3e")
  })

  it("matches GPT variants to OpenAI green", () => {
    expect(colorFor("gpt-4-turbo")).toBe("#10a37f")
  })

  it("returns DEFAULT_COLOR for unknown models", () => {
    expect(colorFor("custom-llm")).toBe(DEFAULT_COLOR)
    expect(colorFor("")).toBe(DEFAULT_COLOR)
  })

  it("labels empty model as «Модель не указана»", () => {
    expect(labelFor("")).toBe("Модель не указана")
    expect(labelFor("claude")).toBe("claude")
  })

  it("exposes UNKNOWN_MODEL_HINT", () => {
    expect(UNKNOWN_MODEL_HINT).toContain("при создании")
  })

  it("exports MODEL_COLORS array", () => {
    expect(MODEL_COLORS.length).toBeGreaterThanOrEqual(6)
  })
})
