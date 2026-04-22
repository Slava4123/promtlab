import { describe, it, expect } from "vitest"
import { detectFileKind } from "./detect"

function makeFile(parts: (string | Uint8Array)[], name: string): File {
  return new File(parts as BlobPart[], name, { type: "" })
}

describe("detectFileKind", () => {
  it("распознаёт .txt по расширению (plain text без magic-bytes)", async () => {
    const file = makeFile(["hello world"], "prompt.txt")
    const result = await detectFileKind(file)
    expect(result.kind).toBe("text")
    expect(result.byExtension).toBe("text")
    expect(result.byMagicBytes).toBeUndefined()
    expect(result.mismatch).toBe(false)
  })

  it("распознаёт .md по расширению", async () => {
    const file = makeFile(["# hello"], "notes.md")
    expect((await detectFileKind(file)).kind).toBe("markdown")
  })

  it("распознаёт .markdown по расширению", async () => {
    const file = makeFile(["## chapter"], "README.markdown")
    expect((await detectFileKind(file)).kind).toBe("markdown")
  })

  it("распознаёт PDF по magic-bytes (%PDF-)", async () => {
    const pdfMagic = new Uint8Array([0x25, 0x50, 0x44, 0x46, 0x2d, 0x31, 0x2e, 0x34])
    const file = makeFile([pdfMagic], "document.pdf")
    const result = await detectFileKind(file)
    expect(result.kind).toBe("pdf")
    expect(result.byMagicBytes).toBe("pdf")
    expect(result.mismatch).toBe(false)
  })

  it("распознаёт DOCX по magic-bytes (PK\\x03\\x04 — ZIP)", async () => {
    const zipMagic = new Uint8Array([0x50, 0x4b, 0x03, 0x04, 0x0a, 0x00])
    const file = makeFile([zipMagic], "draft.docx")
    expect((await detectFileKind(file)).kind).toBe("docx")
  })

  it("распознаёт RTF по magic-bytes ({\\rtf)", async () => {
    const rtfMagic = new Uint8Array([0x7b, 0x5c, 0x72, 0x74, 0x66, 0x31])
    const file = makeFile([rtfMagic], "sample.rtf")
    expect((await detectFileKind(file)).kind).toBe("rtf")
  })

  it("распознаёт HTML по <!doctype html>", async () => {
    const file = makeFile(["<!DOCTYPE html><html><body>hi</body></html>"], "page.html")
    expect((await detectFileKind(file)).kind).toBe("html")
  })

  it("распознаёт HTML по <html (без doctype)", async () => {
    const file = makeFile(["<html><head></head></html>"], "page.htm")
    expect((await detectFileKind(file)).kind).toBe("html")
  })

  it("распознаёт JSON по первому { (по magic-heuristic)", async () => {
    const file = makeFile(['{"content": "hi"}'], "prompt.json")
    const result = await detectFileKind(file)
    expect(result.kind).toBe("json")
    // magic детект считает {...} как json (byMagicBytes=json), расширение тоже json — не mismatch
    expect(result.byMagicBytes).toBe("json")
  })

  it("MAGIC_MISMATCH: файл .txt с PDF-magic → kind=pdf + mismatch=true", async () => {
    const pdfMagic = new Uint8Array([0x25, 0x50, 0x44, 0x46, 0x2d])
    const file = makeFile([pdfMagic], "fake.txt")
    const result = await detectFileKind(file)
    // Доверяем magic-bytes — это pdf; но mismatch отмечаем.
    expect(result.kind).toBe("pdf")
    expect(result.mismatch).toBe(true)
  })

  it("неизвестное расширение → fallback на text", async () => {
    const file = makeFile(["some content"], "unknown.xyz")
    expect((await detectFileKind(file)).kind).toBe("text")
  })

  it("text vs markdown — не считается mismatch (они совместимы)", async () => {
    // Файл .txt с markdown-содержимым — kind=text (по расширению), без mismatch
    const file = makeFile(["# heading"], "notes.txt")
    const result = await detectFileKind(file)
    expect(result.mismatch).toBe(false)
  })
})
