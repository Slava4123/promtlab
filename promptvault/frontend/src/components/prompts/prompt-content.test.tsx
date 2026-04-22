import { describe, it, expect } from "vitest"
import { render } from "@testing-library/react"
import { PromptContent } from "./prompt-content"

describe("PromptContent", () => {
  it("renders **bold** as <strong>", () => {
    const { container } = render(<PromptContent content="**hello**" />)
    const strong = container.querySelector("strong")
    expect(strong).not.toBeNull()
    expect(strong?.textContent).toBe("hello")
  })

  it("renders GFM table", () => {
    const md = "| a | b |\n|---|---|\n| 1 | 2 |"
    const { container } = render(<PromptContent content={md} />)
    expect(container.querySelector("table")).not.toBeNull()
    expect(container.querySelector("thead")).not.toBeNull()
    expect(container.querySelectorAll("tbody tr").length).toBe(1)
  })

  it("renders GFM task-list with disabled checkbox", () => {
    const md = "- [x] done\n- [ ] todo"
    const { container } = render(<PromptContent content={md} />)
    const checkboxes = container.querySelectorAll('input[type="checkbox"]')
    expect(checkboxes.length).toBe(2)
    checkboxes.forEach((cb) => {
      expect((cb as HTMLInputElement).disabled).toBe(true)
    })
    expect((checkboxes[0] as HTMLInputElement).checked).toBe(true)
  })

  it("applies hljs classes to fenced code blocks", () => {
    const md = "```ts\nconst x = 1\n```"
    const { container } = render(<PromptContent content={md} />)
    const code = container.querySelector("code")
    expect(code).not.toBeNull()
    expect(code?.className).toContain("language-ts")
    // rehype-highlight добавляет hljs-классы на токены
    expect(container.querySelector(".hljs-keyword")).not.toBeNull()
  })

  it("strips <script> tags (XSS prevention)", () => {
    const md = "Safe text\n\n<script>alert(1)</script>"
    const { container } = render(<PromptContent content={md} />)
    expect(container.querySelector("script")).toBeNull()
    expect(container.textContent).toContain("Safe text")
  })

  it("strips inline event handlers like onerror", () => {
    const md = "![x](http://x/a.png)\n\n<img src=x onerror=alert(1) />"
    const { container } = render(<PromptContent content={md} />)
    const imgs = container.querySelectorAll("img")
    imgs.forEach((img) => {
      expect(img.getAttribute("onerror")).toBeNull()
    })
  })

  it("blocks javascript: URIs in links", () => {
    const md = "[click](javascript:alert(1))"
    const { container } = render(<PromptContent content={md} />)
    const link = container.querySelector("a")
    // rehype-sanitize удаляет href с запрещённым протоколом;
    // в результате атрибут href отсутствует или пустой
    if (link) {
      const href = link.getAttribute("href")
      expect(href === null || href === "" || !href.startsWith("javascript:")).toBe(true)
    }
  })

  it("adds target=_blank + rel to http links", () => {
    const md = "[go](https://example.com)"
    const { container } = render(<PromptContent content={md} />)
    const link = container.querySelector("a")
    expect(link?.getAttribute("target")).toBe("_blank")
    expect(link?.getAttribute("rel")).toContain("noopener")
    expect(link?.getAttribute("rel")).toContain("noreferrer")
  })
})
