import { describe, it, expect } from "vitest"
import { isSafeHttpsUrl } from "./url"

describe("isSafeHttpsUrl", () => {
  it("принимает валидный https URL", () => {
    expect(isSafeHttpsUrl("https://example.com")).toBe(true)
    expect(isSafeHttpsUrl("https://example.com/path?q=1")).toBe(true)
    expect(isSafeHttpsUrl("https://sub.example.com:8443/")).toBe(true)
  })

  it("отбрасывает http (не https)", () => {
    expect(isSafeHttpsUrl("http://example.com")).toBe(false)
  })

  it("отбрасывает javascript: / data: / file: схемы", () => {
    expect(isSafeHttpsUrl("javascript:alert(1)")).toBe(false)
    expect(isSafeHttpsUrl("data:text/html,<script>alert(1)</script>")).toBe(false)
    expect(isSafeHttpsUrl("file:///etc/passwd")).toBe(false)
  })

  it("отбрасывает malformed URL и пустые значения", () => {
    expect(isSafeHttpsUrl("")).toBe(false)
    expect(isSafeHttpsUrl(null)).toBe(false)
    expect(isSafeHttpsUrl(undefined)).toBe(false)
    expect(isSafeHttpsUrl("not a url")).toBe(false)
    // "https:" без host даёт пустой u.host.
    expect(isSafeHttpsUrl("https:")).toBe(false)
  })
})
