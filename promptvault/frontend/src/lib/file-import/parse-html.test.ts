import { describe, it, expect } from "vitest"
import { parseHtmlFile } from "./parse-html"
import { FileImportError } from "./types"

function makeHtmlFile(html: string, name = "page.html"): File {
  return new File([html], name, { type: "text/html" })
}

describe("parseHtmlFile — XSS sanitize", () => {
  it("<script> удалён из output", async () => {
    const html = `<p>safe</p><script>alert('xss')</script><p>after</p>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).not.toContain("<script>")
    expect(result.content).not.toContain("alert")
    expect(result.content).toContain("safe")
    expect(result.content).toContain("after")
  })

  it("<iframe> удалён", async () => {
    const html = `<p>before</p><iframe src="http://evil.com"></iframe><p>after</p>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).not.toContain("iframe")
    expect(result.content).not.toContain("evil.com")
  })

  it("on*-атрибуты удалены", async () => {
    const html = `<a href="http://ex.com" onclick="alert(1)" onmouseover="evil()">link</a>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    // markdown-ссылка сохранилась без on-атрибутов
    expect(result.content).toContain("link")
    expect(result.content).toContain("http://ex.com")
    expect(result.content).not.toContain("alert")
    expect(result.content).not.toContain("onmouseover")
    expect(result.content).not.toContain("onclick")
  })

  it("javascript: URL в href удалены", async () => {
    const html = `<a href="javascript:alert(1)">click</a>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).not.toContain("javascript:")
    expect(result.content).toContain("click")
  })

  it("<style> удалён", async () => {
    const html = `<style>body{color:red}</style><p>visible</p>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("visible")
    expect(result.content).not.toContain("color:red")
  })
})

describe("parseHtmlFile — markdown output", () => {
  it("<h1>/<h2> → # / ##", async () => {
    const html = `<h1>Title</h1><h2>Subtitle</h2>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("# Title")
    expect(result.content).toContain("## Subtitle")
  })

  it("<strong>/<em>/<del> → ** / * / ~~", async () => {
    const html = `<p><strong>bold</strong> <em>italic</em> <del>strike</del></p>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("**bold**")
    expect(result.content).toContain("*italic*")
    expect(result.content).toContain("~~strike~~")
  })

  it("<ul>/<li> → bullet list", async () => {
    const html = `<ul><li>one</li><li>two</li></ul>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    // Turndown добавляет 3 пробела после маркера для совместимости с parsers
    expect(result.content).toMatch(/-\s+one/)
    expect(result.content).toMatch(/-\s+two/)
  })

  it("<code>/<pre> → fenced code block", async () => {
    const html = `<pre><code>const x = 42</code></pre>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("```")
    expect(result.content).toContain("const x = 42")
  })

  it("<table> → GFM markdown table", async () => {
    const html = `<table>
      <thead><tr><th>Name</th><th>Age</th></tr></thead>
      <tbody><tr><td>Alice</td><td>30</td></tr></tbody>
    </table>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("| Name | Age |")
    expect(result.content).toContain("| --- | --- |")
    expect(result.content).toContain("| Alice | 30 |")
  })

  it("<a href> → inline markdown link", async () => {
    const html = `<p><a href="https://example.com">click</a></p>`
    const result = await parseHtmlFile(makeHtmlFile(html))
    expect(result.content).toContain("[click](https://example.com)")
  })

  it("пустой HTML → EMPTY_RESULT", async () => {
    const file = makeHtmlFile("")
    await expect(parseHtmlFile(file)).rejects.toThrow(FileImportError)
    try {
      await parseHtmlFile(file)
    } catch (err) {
      expect((err as FileImportError).code).toBe("EMPTY_RESULT")
    }
  })

  it("HTML из одних опасных тегов → EMPTY_RESULT", async () => {
    const html = `<script>alert(1)</script><style>body{}</style>`
    await expect(parseHtmlFile(makeHtmlFile(html))).rejects.toThrow(FileImportError)
  })
})
