import { describe, it, expect } from "vitest"
import { parseTextFile, normalizeText } from "./parse-text"

function makeFile(parts: (string | Uint8Array)[], name = "test.txt"): File {
  return new File(parts as BlobPart[], name, { type: "text/plain" })
}

describe("normalizeText", () => {
  it("\\r\\n → \\n", () => {
    expect(normalizeText("line1\r\nline2\r\n")).toBe("line1\nline2\n")
  })

  it("одинокий \\r → \\n (legacy Mac)", () => {
    expect(normalizeText("line1\rline2\r")).toBe("line1\nline2\n")
  })

  it("микс \\r\\n и \\r корректно обрабатывается", () => {
    expect(normalizeText("a\r\nb\rc\nd")).toBe("a\nb\nc\nd")
  })

  it("NULL-байты (\\u0000) вырезаются", () => {
    expect(normalizeText("hello\u0000world")).toBe("helloworld")
  })
})

describe("parseTextFile", () => {
  it("читает обычный UTF-8 файл", async () => {
    const file = makeFile(["Привет, мир!"], "greeting.txt")
    const result = await parseTextFile(file, "text")
    expect(result.content).toBe("Привет, мир!")
    expect(result.kind).toBe("text")
    expect(result.filename).toBe("greeting.txt")
    expect(result.truncated).toBe(false)
    expect(result.warnings).toEqual([])
    expect(result.detectedEncoding).toBe("utf-8")
  })

  it("снимает UTF-8 BOM (EF BB BF)", async () => {
    const bom = new Uint8Array([0xef, 0xbb, 0xbf])
    const payload = new TextEncoder().encode("hello")
    const file = new File([bom, payload], "bom.txt", { type: "text/plain" })
    const result = await parseTextFile(file, "text")
    expect(result.content).toBe("hello")
    expect(result.content.charCodeAt(0)).not.toBe(0xfeff)
  })

  it("снимает UTF-16LE BOM (FF FE)", async () => {
    const bom = new Uint8Array([0xff, 0xfe])
    // "hi" в UTF-16LE: 0x68 0x00 0x69 0x00
    const payload = new Uint8Array([0x68, 0x00, 0x69, 0x00])
    const file = new File([bom, payload], "utf16.txt", { type: "text/plain" })
    const result = await parseTextFile(file, "text")
    expect(result.content).toBe("hi")
    expect(result.detectedEncoding).toBe("utf-16le")
  })

  it("нормализует \\r\\n в \\n", async () => {
    const file = makeFile(["line1\r\nline2\r\nline3"])
    const result = await parseTextFile(file, "text")
    expect(result.content).toBe("line1\nline2\nline3")
  })

  it("обрезает до MAX_PROMPT_CONTENT_LENGTH при превышении", async () => {
    const big = "a".repeat(100_500)
    const file = makeFile([big])
    const result = await parseTextFile(file, "text")
    expect(result.content.length).toBe(100_000)
    expect(result.truncated).toBe(true)
  })

  it("не обрезает если content ровно MAX_PROMPT_CONTENT_LENGTH", async () => {
    const exact = "a".repeat(100_000)
    const file = makeFile([exact])
    const result = await parseTextFile(file, "text")
    expect(result.content.length).toBe(100_000)
    expect(result.truncated).toBe(false)
  })

  it("детектит кракозябры и даёт warning", async () => {
    // 20 символов U+FFFD и 10 обычных — это > 10% replacement chars
    const garbage = "\uFFFD".repeat(20) + "normal"
    const file = makeFile([garbage])
    const result = await parseTextFile(file, "text")
    expect(result.warnings.length).toBeGreaterThan(0)
    expect(result.warnings[0]).toContain("кодировк")
  })

  it("markdown — kind=markdown, поведение то же что для text", async () => {
    const file = makeFile(["# Title\n\nParagraph"], "doc.md")
    const result = await parseTextFile(file, "markdown")
    expect(result.kind).toBe("markdown")
    expect(result.content).toContain("# Title")
  })
})
