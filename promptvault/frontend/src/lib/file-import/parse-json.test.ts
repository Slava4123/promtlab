import { describe, it, expect } from "vitest"
import { parseJsonFile } from "./parse-json"
import { FileImportError } from "./types"

function makeJsonFile(obj: unknown, name = "prompt.json"): File {
  return new File([JSON.stringify(obj)], name, { type: "application/json" })
}

describe("parseJsonFile", () => {
  it("prompt-JSON с content → содержимое в content + metadata", async () => {
    const file = makeJsonFile({
      title: "Code reviewer",
      content: "You are a senior code reviewer...",
      model: "anthropic/claude-sonnet",
    })
    const result = await parseJsonFile(file)
    expect(result.kind).toBe("json")
    expect(result.content).toBe("You are a senior code reviewer...")
    expect(result.metadata).toEqual({
      title: "Code reviewer",
      model: "anthropic/claude-sonnet",
    })
    expect(result.warnings).toEqual([])
  })

  it("prompt-JSON только с content (без title/model) → metadata отсутствует", async () => {
    const file = makeJsonFile({ content: "just the prompt" })
    const result = await parseJsonFile(file)
    expect(result.content).toBe("just the prompt")
    expect(result.metadata).toBeUndefined()
  })

  it("generic JSON → pretty-print в content + warning", async () => {
    const file = makeJsonFile({ foo: "bar", list: [1, 2, 3] })
    const result = await parseJsonFile(file)
    expect(result.content).toContain('"foo": "bar"')
    expect(result.content).toContain('"list": [')
    expect(result.warnings.length).toBeGreaterThan(0)
    expect(result.warnings[0]).toContain("формат")
  })

  it("массив JSON → pretty-print + warning (не prompt-shape)", async () => {
    const file = makeJsonFile([{ a: 1 }, { b: 2 }])
    const result = await parseJsonFile(file)
    expect(result.content).toContain('"a": 1')
    expect(result.warnings.length).toBeGreaterThan(0)
  })

  it("content не строка → generic JSON (warning)", async () => {
    const file = makeJsonFile({ content: 42 })
    const result = await parseJsonFile(file)
    expect(result.warnings.length).toBeGreaterThan(0)
  })

  it("invalid JSON → FileImportError(PARSE_FAILED)", async () => {
    const file = new File(["not valid { json"], "broken.json", {
      type: "application/json",
    })
    await expect(parseJsonFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseJsonFile(file)
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("PARSE_FAILED")
    }
  })

  it("content > 100k символов → truncated=true", async () => {
    const big = "x".repeat(100_500)
    const file = makeJsonFile({ content: big })
    const result = await parseJsonFile(file)
    expect(result.content.length).toBe(100_000)
    expect(result.truncated).toBe(true)
  })

  it("title > 300 символов → Zod rejection → fallback на generic JSON pretty-print", async () => {
    const longTitle = "T".repeat(500)
    const file = makeJsonFile({ content: "body", title: longTitle })
    const result = await parseJsonFile(file)
    // Zod отклонил структуру → fallback: pretty-print + warning, metadata пустое
    expect(result.metadata).toBeUndefined()
    expect(result.warnings.length).toBeGreaterThan(0)
    expect(result.content).toContain("body")
  })

  it("empty content строка → Zod rejection → generic JSON", async () => {
    const file = makeJsonFile({ content: "" })
    const result = await parseJsonFile(file)
    // content: "" провалит Zod min(1) → fallback
    expect(result.warnings.length).toBeGreaterThan(0)
    expect(result.metadata).toBeUndefined()
  })

  it("extra поля в prompt-JSON игнорируются (passthrough)", async () => {
    const file = makeJsonFile({
      content: "the prompt",
      title: "My title",
      extra_field: "ignored",
      tags: ["a", "b"],
    })
    const result = await parseJsonFile(file)
    expect(result.content).toBe("the prompt")
    expect(result.metadata?.title).toBe("My title")
    expect(result.warnings).toEqual([])
  })
})
