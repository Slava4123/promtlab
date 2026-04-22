import { describe, it, expect } from "vitest"
import { parseFile, FileImportError } from "./parsers"

function makeFile(parts: (string | Uint8Array)[], name: string, type = ""): File {
  return new File(parts as BlobPart[], name, { type })
}

describe("parseFile (router)", () => {
  it("текстовый файл роутится в parseTextFile", async () => {
    const file = makeFile(["hello"], "test.txt")
    const result = await parseFile(file)
    expect(result.kind).toBe("text")
    expect(result.content).toBe("hello")
  })

  it("markdown файл роутится в parseTextFile с kind=markdown", async () => {
    const file = makeFile(["# heading"], "doc.md")
    const result = await parseFile(file)
    expect(result.kind).toBe("markdown")
  })

  it("JSON файл роутится в parseJsonFile", async () => {
    const file = makeFile(['{"content": "hi"}'], "p.json")
    const result = await parseFile(file)
    expect(result.kind).toBe("json")
    expect(result.content).toBe("hi")
  })

  it("пустой файл → FileImportError(EMPTY_RESULT)", async () => {
    const file = makeFile([""], "empty.txt")
    await expect(parseFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseFile(file)
    } catch (err) {
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
    }
  })

  it("файл больше MAX_UPLOAD_BYTES → FileImportError(SIZE_EXCEEDED)", async () => {
    // Моделируем "большой" файл через Blob с fake size через Object.defineProperty.
    // 11 MB > 10 MB hardcap.
    const big = new Uint8Array(11 * 1024 * 1024)
    const file = new File([big], "big.txt", { type: "text/plain" })
    await expect(parseFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseFile(file)
    } catch (err) {
      expect((err as FileImportError).code).toBe("SIZE_EXCEEDED")
    }
  })

  it("PDF роутится в parse-pdf (lazy-chunk)", async () => {
    // С minimal-PDF магическими байтами pdfjs упадёт на "Invalid PDF structure",
    // это ожидаемо — нам важен факт маршрутизации (FileImportError, но не UNSUPPORTED).
    const pdfMagic = new Uint8Array([0x25, 0x50, 0x44, 0x46, 0x2d])
    const file = makeFile([pdfMagic], "doc.pdf")
    try {
      await parseFile(file)
      expect.fail("должно было бросить (пустой PDF)")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).not.toBe("UNSUPPORTED")
    }
  })

  it("DOCX роутится в parse-docx (lazy-chunk)", async () => {
    const zipMagic = new Uint8Array([0x50, 0x4b, 0x03, 0x04])
    const file = makeFile([zipMagic], "doc.docx")
    try {
      await parseFile(file)
      expect.fail("должно было бросить (минимальный zip, не валидный docx)")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).not.toBe("UNSUPPORTED")
    }
  })

  it(".ipynb роутится в parse-ipynb (не в generic-json)", async () => {
    const nb = JSON.stringify({
      cells: [{ cell_type: "markdown", source: "# hello" }],
      nbformat: 4,
    })
    const file = makeFile([nb], "notebook.ipynb")
    const result = await parseFile(file)
    expect(result.kind).toBe("ipynb")
    expect(result.content).toContain("# hello")
  })
})
