import { describe, it, expect, vi } from "vitest"
import { FileImportError } from "./types"

// Мокаем mammoth/mammoth.browser целиком — полноценный DOCX-fixture требует
// сгенерированного zip-архива, что усложняет тестирование. Достаточно проверить
// что наш wrapper корректно обрабатывает результаты/ошибки mammoth.
vi.mock("mammoth/mammoth.browser.js", () => ({
  default: {
    convertToMarkdown: vi.fn(),
    images: {
      imgElement: vi.fn(() => "mock-img-handler"),
    },
  },
}))

async function getMammoth() {
  return (await import("mammoth/mammoth.browser.js")).default as unknown as {
    convertToMarkdown: ReturnType<typeof vi.fn>
  }
}

async function getParseDocx() {
  return (await import("./parse-docx")).parseDocxFile
}

function makeDocxFile(size = 1000): File {
  const buf = new Uint8Array(size)
  // zip magic PK\x03\x04 для реализма
  buf[0] = 0x50
  buf[1] = 0x4b
  buf[2] = 0x03
  buf[3] = 0x04
  return new File([buf], "test.docx", {
    type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  })
}

describe("parseDocxFile", () => {
  it("обычный DOCX → ParseResult с markdown", async () => {
    const mammoth = await getMammoth()
    mammoth.convertToMarkdown.mockResolvedValueOnce({
      value: "# Title\n\nParagraph text.",
      messages: [],
    })
    const parseDocxFile = await getParseDocx()
    const result = await parseDocxFile(makeDocxFile())
    expect(result.kind).toBe("docx")
    expect(result.content).toContain("# Title")
    expect(result.content).toContain("Paragraph text")
    expect(result.warnings).toEqual([])
  })

  it("mammoth warnings → попадают в result.warnings (first 3)", async () => {
    const mammoth = await getMammoth()
    mammoth.convertToMarkdown.mockResolvedValueOnce({
      value: "text",
      messages: [
        { type: "warning", message: "Unrecognized style X" },
        { type: "warning", message: "Unrecognized style Y" },
        { type: "info", message: "whatever" },
      ],
    })
    const parseDocxFile = await getParseDocx()
    const result = await parseDocxFile(makeDocxFile())
    expect(result.warnings.length).toBeGreaterThan(0)
    expect(result.warnings[0]).toContain("Unrecognized style X")
    expect(result.warnings[0]).toContain("Unrecognized style Y")
    expect(result.warnings[0]).not.toContain("whatever")
  })

  it("mammoth бросает → FileImportError(PARSE_FAILED)", async () => {
    const mammoth = await getMammoth()
    mammoth.convertToMarkdown.mockRejectedValueOnce(new Error("malformed docx"))
    const parseDocxFile = await getParseDocx()
    await expect(parseDocxFile(makeDocxFile())).rejects.toThrow(FileImportError)
  })

  it("пустой markdown → FileImportError(EMPTY_RESULT)", async () => {
    const mammoth = await getMammoth()
    mammoth.convertToMarkdown.mockResolvedValue({ value: "   \n  \n", messages: [] })
    const parseDocxFile = await getParseDocx()
    try {
      await parseDocxFile(makeDocxFile())
      expect.fail("должно было бросить")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
    }
  })

  it("zip-bomb защита: вывод > 500k → SIZE_EXCEEDED", async () => {
    const mammoth = await getMammoth()
    const huge = "x".repeat(600_000)
    mammoth.convertToMarkdown.mockResolvedValue({ value: huge, messages: [] })
    const parseDocxFile = await getParseDocx()
    try {
      await parseDocxFile(makeDocxFile())
      expect.fail("должно было бросить")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("SIZE_EXCEEDED")
    }
  })

  it("content > 100k → truncated=true", async () => {
    const mammoth = await getMammoth()
    mammoth.convertToMarkdown.mockResolvedValueOnce({
      value: "a".repeat(100_500),
      messages: [],
    })
    const parseDocxFile = await getParseDocx()
    const result = await parseDocxFile(makeDocxFile())
    expect(result.truncated).toBe(true)
    expect(result.content.length).toBe(100_000)
  })
})
