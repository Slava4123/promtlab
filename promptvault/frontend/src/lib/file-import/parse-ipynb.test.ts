import { describe, it, expect } from "vitest"
import { parseIpynbFile } from "./parse-ipynb"
import { FileImportError } from "./types"

function makeIpynbFile(notebook: unknown, name = "notebook.ipynb"): File {
  return new File([JSON.stringify(notebook)], name, { type: "application/json" })
}

describe("parseIpynbFile", () => {
  it("конвертирует markdown-ячейки в markdown", async () => {
    const nb = {
      cells: [
        { cell_type: "markdown", source: "# Heading" },
        { cell_type: "markdown", source: ["paragraph\n", "line 2"] },
      ],
      nbformat: 4,
    }
    const file = makeIpynbFile(nb)
    const result = await parseIpynbFile(file)
    expect(result.kind).toBe("ipynb")
    expect(result.content).toContain("# Heading")
    expect(result.content).toContain("paragraph\nline 2")
  })

  it("code-ячейки оборачиваются в fenced-блок с языком из metadata", async () => {
    const nb = {
      cells: [
        {
          cell_type: "code",
          source: "print('hi')",
        },
      ],
      metadata: { kernelspec: { language: "python" } },
      nbformat: 4,
    }
    const file = makeIpynbFile(nb)
    const result = await parseIpynbFile(file)
    expect(result.content).toContain("```python")
    expect(result.content).toContain("print('hi')")
    expect(result.content).toContain("```")
  })

  it("raw-ячейки пропускаются", async () => {
    const nb = {
      cells: [
        { cell_type: "markdown", source: "keep" },
        { cell_type: "raw", source: "skip this" },
        { cell_type: "markdown", source: "also keep" },
      ],
      nbformat: 4,
    }
    const result = await parseIpynbFile(makeIpynbFile(nb))
    expect(result.content).toContain("keep")
    expect(result.content).toContain("also keep")
    expect(result.content).not.toContain("skip this")
  })

  it("outputs code-ячейки игнорируются (важно: не попадают в content)", async () => {
    const nb = {
      cells: [
        {
          cell_type: "code",
          source: "print('x')",
          outputs: [{ output_type: "stream", text: "x\n" }],
        },
      ],
      metadata: { language_info: { name: "python" } },
      nbformat: 4,
    }
    const result = await parseIpynbFile(makeIpynbFile(nb))
    expect(result.content).toContain("print('x')")
    // Output "x\n" не должен попасть в content
    expect(result.content.split("\n").filter((l) => l === "x").length).toBe(0)
  })

  it("пустые source пропускаются", async () => {
    const nb = {
      cells: [
        { cell_type: "markdown", source: "" },
        { cell_type: "markdown", source: "real" },
      ],
      nbformat: 4,
    }
    const result = await parseIpynbFile(makeIpynbFile(nb))
    expect(result.content).toBe("real")
  })

  it("notebook без валидных ячеек → EMPTY_RESULT", async () => {
    const nb = { cells: [{ cell_type: "raw", source: "boring" }], nbformat: 4 }
    await expect(parseIpynbFile(makeIpynbFile(nb))).rejects.toThrow(FileImportError)
    try {
      await parseIpynbFile(makeIpynbFile(nb))
    } catch (err) {
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
    }
  })

  it("invalid JSON → PARSE_FAILED", async () => {
    const file = new File(["not json"], "bad.ipynb", { type: "application/json" })
    await expect(parseIpynbFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseIpynbFile(file)
    } catch (err) {
      expect((err as FileImportError).code).toBe("PARSE_FAILED")
    }
  })

  it("JSON без cells → PARSE_FAILED (Zod reject)", async () => {
    const file = makeIpynbFile({ foo: "bar" })
    await expect(parseIpynbFile(file)).rejects.toThrow(FileImportError)
  })

  it("per-cell language_info переопределяет default kernelspec language", async () => {
    const nb = {
      cells: [
        {
          cell_type: "code",
          source: "SELECT 1",
          metadata: { language_info: { name: "sql" } },
        },
      ],
      metadata: { kernelspec: { language: "python" } },
      nbformat: 4,
    }
    const result = await parseIpynbFile(makeIpynbFile(nb))
    expect(result.content).toContain("```sql")
  })

  it("content > 100k → truncated=true", async () => {
    const big = "x".repeat(100_500)
    const nb = { cells: [{ cell_type: "markdown", source: big }], nbformat: 4 }
    const result = await parseIpynbFile(makeIpynbFile(nb))
    expect(result.truncated).toBe(true)
    expect(result.content.length).toBe(100_000)
  })
})
