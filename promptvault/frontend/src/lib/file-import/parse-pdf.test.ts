import { describe, it, expect, vi, beforeEach } from "vitest"
import { FileImportError } from "./types"

// Мокаем pdfjs-dist целиком. Fixture-генерация валидного PDF для честного
// интеграционного теста нетривиальна (нужен PDF-писатель), оставляем для E2E.
// Здесь проверяем нашу обёртку вокруг pdfjs-API.

const mockGetDocument = vi.fn()

vi.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  getDocument: mockGetDocument,
  GlobalWorkerOptions: { workerSrc: "" },
}))

vi.mock("pdfjs-dist/legacy/build/pdf.worker.min.mjs?url", () => ({
  default: "mock-worker-url",
}))

async function getParsePdf() {
  return (await import("./parse-pdf")).parsePdfFile
}

function makePdfFile(size = 1000): File {
  const buf = new Uint8Array(size)
  // PDF magic-bytes
  buf[0] = 0x25
  buf[1] = 0x50
  buf[2] = 0x44
  buf[3] = 0x46
  buf[4] = 0x2d
  return new File([buf], "test.pdf", { type: "application/pdf" })
}

function mockPdfDoc(
  pages: Array<{ items: Array<{ str: string; transform?: number[]; hasEOL?: boolean }> }>,
) {
  const doc = {
    numPages: pages.length,
    getPage: vi.fn(async (n: number) => ({
      getTextContent: vi.fn(async () => ({
        items: pages[n - 1].items,
      })),
      cleanup: vi.fn(),
    })),
    destroy: vi.fn(),
  }
  mockGetDocument.mockReturnValueOnce({
    promise: Promise.resolve(doc),
  })
  return doc
}

beforeEach(() => {
  mockGetDocument.mockReset()
})

describe("parsePdfFile", () => {
  it("извлекает текст из одной страницы", async () => {
    mockPdfDoc([
      {
        items: [
          { str: "Hello", transform: [0, 0, 0, 0, 0, 100], hasEOL: false },
          { str: " world", transform: [0, 0, 0, 0, 100, 100], hasEOL: true },
        ],
      },
    ])
    const parsePdfFile = await getParsePdf()
    const result = await parsePdfFile(makePdfFile())
    expect(result.kind).toBe("pdf")
    expect(result.content).toContain("Hello")
    expect(result.content).toContain("world")
    expect(result.pages).toBe(1)
  })

  it("multi-page: страницы склеиваются через \\n\\n", async () => {
    mockPdfDoc([
      { items: [{ str: "page 1 content", hasEOL: false }] },
      { items: [{ str: "page 2 content", hasEOL: false }] },
      { items: [{ str: "page 3 content", hasEOL: false }] },
    ])
    const parsePdfFile = await getParsePdf()
    const result = await parsePdfFile(makePdfFile())
    expect(result.content).toContain("page 1 content")
    expect(result.content).toContain("page 2 content")
    expect(result.content).toContain("page 3 content")
    expect(result.pages).toBe(3)
  })

  it("scan-only PDF (нет items с str) → EMPTY_RESULT", async () => {
    mockPdfDoc([{ items: [] }, { items: [] }])
    const parsePdfFile = await getParsePdf()
    try {
      await parsePdfFile(makePdfFile())
      expect.fail("должно было бросить")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
      expect((err as Error).message).toMatch(/скан/)
    }
  })

  it(">500 страниц → SIZE_EXCEEDED", async () => {
    const mockDoc = {
      numPages: 501,
      getPage: vi.fn(),
      destroy: vi.fn(),
    }
    mockGetDocument.mockReturnValueOnce({ promise: Promise.resolve(mockDoc) })
    const parsePdfFile = await getParsePdf()
    try {
      await parsePdfFile(makePdfFile())
      expect.fail("должно было бросить")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("SIZE_EXCEEDED")
    }
  })

  it("getDocument rejects → PARSE_FAILED", async () => {
    mockGetDocument.mockReturnValueOnce({
      promise: Promise.reject(new Error("corrupt pdf")),
    })
    const parsePdfFile = await getParsePdf()
    try {
      await parsePdfFile(makePdfFile())
      expect.fail("должно было бросить")
    } catch (err) {
      expect(err).toBeInstanceOf(FileImportError)
      expect((err as FileImportError).code).toBe("PARSE_FAILED")
    }
  })

  it("hasEOL=true → добавляет \\n", async () => {
    mockPdfDoc([
      {
        items: [
          { str: "line1", hasEOL: true },
          { str: "line2", hasEOL: false },
        ],
      },
    ])
    const parsePdfFile = await getParsePdf()
    const result = await parsePdfFile(makePdfFile())
    expect(result.content).toContain("line1\nline2")
  })

  it("destroy() вызывается (cleanup ресурсов worker'а)", async () => {
    const doc = mockPdfDoc([{ items: [{ str: "x", hasEOL: false }] }])
    const parsePdfFile = await getParsePdf()
    await parsePdfFile(makePdfFile())
    expect(doc.destroy).toHaveBeenCalled()
  })
})
