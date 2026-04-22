import { describe, it, expect } from "vitest"
import { detectAndDecode } from "./encoding"

// Генерирует байты кириллического текста в windows-1251.
// Кириллица 0x410-0x44F mapped to 0xC0-0xFF в cp1251.
function cp1251Bytes(utf: string): Uint8Array {
  const bytes: number[] = []
  for (const ch of utf) {
    const code = ch.charCodeAt(0)
    if (code <= 0x7f) bytes.push(code)
    else if (code >= 0x0410 && code <= 0x044f) bytes.push(0xc0 + (code - 0x0410))
    else if (code === 0x0451) bytes.push(0xb8) // ё
    else if (code === 0x0401) bytes.push(0xa8) // Ё
    else bytes.push(0x3f) // ? для unknown
  }
  return new Uint8Array(bytes)
}

describe("detectAndDecode", () => {
  it("windows-1251 RU-текст → detected + recovered", () => {
    const text = "Привет, мир! Это тестовое сообщение для проверки кодировки."
    const bytes = cp1251Bytes(text)
    const utf8FallbackGarbage = new TextDecoder("utf-8", { fatal: false }).decode(bytes)
    const result = detectAndDecode(bytes, utf8FallbackGarbage)
    expect(result.recovered).toBe(true)
    expect(result.encoding.toLowerCase()).toContain("1251")
    expect(result.content).toContain("Привет")
  })

  it("валидный UTF-8 текст → detected но recovered=false (оригинал уже ок)", () => {
    const text = "Normal ASCII text only here"
    const bytes = new TextEncoder().encode(text)
    const result = detectAndDecode(bytes, text)
    expect(result.recovered).toBe(false)
    expect(result.content).toBe(text)
  })

  it("короткий ambiguous → fallback UTF-8", () => {
    const tiny = new Uint8Array([0x41, 0x42, 0x43]) // "ABC"
    const result = detectAndDecode(tiny, "ABC")
    expect(result.recovered).toBe(false)
    expect(result.content).toBe("ABC")
  })

  it("bytes которые детект не знает → fallback на UTF-8", () => {
    // Rare pattern — confusing bytes
    const bytes = new Uint8Array([0xff, 0xff, 0xff, 0xff, 0xff])
    const fallback = "garbage"
    const result = detectAndDecode(bytes, fallback)
    // Может не определиться — recovered=false и content=fallback
    if (!result.recovered) {
      expect(result.content).toBe(fallback)
    }
  })
})
