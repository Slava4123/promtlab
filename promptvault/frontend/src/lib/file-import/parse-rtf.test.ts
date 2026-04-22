import { describe, it, expect } from "vitest"
import { parseRtfFile } from "./parse-rtf"
import { FileImportError } from "./types"

function makeRtfFile(content: string, name = "doc.rtf"): File {
  return new File([content], name, { type: "application/rtf" })
}

describe("parseRtfFile", () => {
  it("простой RTF → чистый текст", async () => {
    const rtf = "{\\rtf1\\ansi\\ansicpg1251\\deff0 Hello world!}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.kind).toBe("rtf")
    expect(result.content).toContain("Hello world!")
  })

  it("удаляет control words (\\b \\i \\fs24)", async () => {
    const rtf = "{\\rtf1\\ansi \\b Bold\\b0  text \\i italic\\i0 .}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).not.toContain("\\b")
    expect(result.content).not.toContain("\\i")
    expect(result.content).toContain("Bold")
    expect(result.content).toContain("italic")
  })

  it("\\par → newline", async () => {
    const rtf = "{\\rtf1 line1\\par line2\\par line3}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("line1")
    expect(result.content).toContain("line2")
    expect(result.content).toContain("line3")
  })

  it("hex-escape \\'XX декодируется как cp1251 (RU-текст)", async () => {
    // \'f0 = 0xF0 в cp1251 = 'р' (русская буква, U+0440)
    // \'e0 = 0xE0 = 'а'
    // \'e1 = 0xE1 = 'б'
    const rtf = "{\\rtf1\\ansi \\'f0\\'e0\\'e1}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("раб")
  })

  it("\\uN декодируется как Unicode", async () => {
    // \u1040 с fallback '?' — это U+0410 + знак... но spec: \uN — signed 16-bit.
    // \u1074? = U+0434 "д"
    const rtf = "{\\rtf1 \\u1076? \\u1072?}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("д")
    expect(result.content).toContain("а")
  })

  it("fonttbl группа целиком пропускается", async () => {
    const rtf =
      "{\\rtf1{\\fonttbl{\\f0\\fnil\\fcharset204 Arial;}}\\f0\\fs24 visible text}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("visible text")
    expect(result.content).not.toContain("Arial")
  })

  it("escape literals \\{ \\} \\\\", async () => {
    const rtf = "{\\rtf1 brace: \\{ \\} slash: \\\\}"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("{")
    expect(result.content).toContain("}")
    expect(result.content).toContain("\\")
  })

  it("не-RTF контент → PARSE_FAILED", async () => {
    const file = makeRtfFile("plain text, no RTF header")
    await expect(parseRtfFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseRtfFile(file)
    } catch (err) {
      expect((err as FileImportError).code).toBe("PARSE_FAILED")
    }
  })

  it("RTF без текста → EMPTY_RESULT", async () => {
    const rtf = "{\\rtf1\\ansi{\\fonttbl{\\f0 Arial;}}}"
    await expect(parseRtfFile(makeRtfFile(rtf))).rejects.toThrow(FileImportError)
    try {
      await parseRtfFile(makeRtfFile(rtf))
    } catch (err) {
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
    }
  })

  it("кавычки \\lquote \\rquote \\ldblquote \\rdblquote", async () => {
    const rtf = "{\\rtf1 \\ldblquote hi\\rdblquote  \\lquote x\\rquote }"
    const result = await parseRtfFile(makeRtfFile(rtf))
    expect(result.content).toContain("\u201c")
    expect(result.content).toContain("\u201d")
  })
})
