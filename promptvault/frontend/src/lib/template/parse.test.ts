import { extractVariables, hasVariables, renderTemplate } from "./parse"

describe("extractVariables", () => {
  it("extracts a single variable", () => {
    expect(extractVariables("Привет, {{name}}!")).toEqual(["name"])
  })

  it("extracts multiple variables in order of first appearance", () => {
    expect(extractVariables("{{lang}} код: {{code}}")).toEqual(["lang", "code"])
  })

  it("de-duplicates repeated variables preserving first-appearance order", () => {
    expect(extractVariables("{{x}} and {{y}} and {{x}} again")).toEqual(["x", "y"])
  })

  it("returns empty array when there are no variables", () => {
    expect(extractVariables("Plain text with no placeholders.")).toEqual([])
  })

  it("ignores invalid identifier forms as literals", () => {
    const content =
      "literal {{my-var}} {{ name }} {{1bad}} {{}} {{has space}}"
    expect(extractVariables(content)).toEqual([])
  })

  it("accepts underscore and digits after the first char", () => {
    expect(extractVariables("{{_private}} {{var_1}} {{MixedCase42}}")).toEqual([
      "_private",
      "var_1",
      "MixedCase42",
    ])
  })

  it("accepts cyrillic identifiers", () => {
    expect(
      extractVariables("Привет, {{имя}}! Компания: {{компания}}, версия {{имя_1}}"),
    ).toEqual(["имя", "компания", "имя_1"])
  })

  it("accepts CJK and mixed-script identifiers", () => {
    expect(extractVariables("{{名前}} и {{ИмяКомпании}}")).toEqual([
      "名前",
      "ИмяКомпании",
    ])
  })

  it("rejects identifier starting with a digit regardless of script", () => {
    // Leading digit rule applies universally — Latin, Cyrillic, any.
    expect(extractVariables("{{1name}} {{1имя}}")).toEqual([])
  })
})

describe("hasVariables", () => {
  it("returns true when content has at least one valid variable", () => {
    expect(hasVariables("hello {{name}}")).toBe(true)
  })

  it("returns false for plain text", () => {
    expect(hasVariables("plain text")).toBe(false)
  })

  it("returns false for invalid identifier forms", () => {
    expect(hasVariables("{{ name }} and {{1}} and {{my-var}}")).toBe(false)
  })

  it("is safe to call multiple times (no lastIndex drift)", () => {
    const s = "hello {{x}}"
    // Regression guard: a naive global regex with test() would flip between runs.
    expect(hasVariables(s)).toBe(true)
    expect(hasVariables(s)).toBe(true)
    expect(hasVariables(s)).toBe(true)
  })
})

describe("renderTemplate", () => {
  it("substitutes a single variable", () => {
    expect(renderTemplate("Привет, {{name}}!", { name: "Аня" })).toBe(
      "Привет, Аня!",
    )
  })

  it("substitutes multiple variables", () => {
    expect(
      renderTemplate("{{lang}} code: {{code}}", { lang: "Go", code: "fmt.Println" }),
    ).toBe("Go code: fmt.Println")
  })

  it("replaces every occurrence of a repeated variable", () => {
    expect(renderTemplate("{{x}} and {{x}} and {{x}}", { x: "hi" })).toBe(
      "hi and hi and hi",
    )
  })

  it("renders empty string for empty-value keys", () => {
    expect(renderTemplate("before{{name}}after", { name: "" })).toBe("beforeafter")
  })

  it("renders empty string for missing keys", () => {
    expect(renderTemplate("{{a}} {{b}}", { a: "x" })).toBe("x ")
  })

  it("does NOT recursively render substituted values (single-pass)", () => {
    // {{a}} -> "{{b}}"; the literal {{b}} must survive, not be re-scanned.
    expect(renderTemplate("{{a}}", { a: "{{b}}" })).toBe("{{b}}")
  })

  it("does NOT interpret regex metacharacters in values", () => {
    // $1, $&, \, $` are all valid String.replace() back-refs when passed as a
    // string. Using a function replacer avoids that.
    const value = "$1 $& $` \\ $$"
    expect(renderTemplate("before {{x}} after", { x: value })).toBe(
      `before ${value} after`,
    )
  })

  it("leaves invalid identifier forms as literals in the output", () => {
    const template = "{{ name }} and {{my-var}} and {{1}}"
    expect(renderTemplate(template, { name: "X" })).toBe(template)
  })

  it("handles unicode in values (cyrillic, japanese, emoji)", () => {
    expect(
      renderTemplate("{{greeting}} {{lang}} {{emoji}}", {
        greeting: "Привет",
        lang: "日本語",
        emoji: "🚀",
      }),
    ).toBe("Привет 日本語 🚀")
  })

  it("renders cyrillic variable names end-to-end", () => {
    expect(
      renderTemplate(
        "Напиши письмо для {{имя}} из компании {{компания}}. Тон: {{тон}}.",
        { имя: "Иван", компания: "Ростелеком", тон: "деловой" },
      ),
    ).toBe("Напиши письмо для Иван из компании Ростелеком. Тон: деловой.")
  })

  it("parses a large template with many variables quickly", () => {
    // Construct a ~10 KB template with 50 unique variables.
    const parts: string[] = []
    const values: Record<string, string> = {}
    for (let i = 0; i < 50; i++) {
      parts.push(`section ${i}: {{var${i}}} `.padEnd(200, "x"))
      values[`var${i}`] = `v${i}`
    }
    const template = parts.join("")

    const start = performance.now()
    const vars = extractVariables(template)
    const rendered = renderTemplate(template, values)
    const duration = performance.now() - start

    expect(vars).toHaveLength(50)
    expect(rendered).toContain("v0")
    expect(rendered).toContain("v49")
    expect(rendered).not.toContain("{{var0}}")
    // Generous bound — on any modern machine this is microseconds.
    expect(duration).toBeLessThan(50)
  })
})
