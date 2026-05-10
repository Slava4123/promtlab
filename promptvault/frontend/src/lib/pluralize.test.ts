import { describe, it, expect } from "vitest"
import { pluralizeRu } from "./pluralize"

// Эталон для русского pluralize. Если кто-то меняет логику — пускай поймёт,
// что 11-14 это many, а не few; что 21/22/25 распадаются на one/few/many и т.д.
// Если эти кейсы перестанут работать, юзер увидит «1 использований» или
// «11 использование» — оба смотрятся как ошибка локализации.

const FORMS = ["использование", "использования", "использований"] as const
const [ONE, FEW, MANY] = FORMS

function pl(n: number): string {
  return pluralizeRu(n, ONE, FEW, MANY)
}

describe("pluralizeRu (CLDR `ru` rules)", () => {
  it("0 → many", () => {
    expect(pl(0)).toBe(MANY)
  })

  it("1 → one", () => {
    expect(pl(1)).toBe(ONE)
  })

  it("2-4 → few", () => {
    expect(pl(2)).toBe(FEW)
    expect(pl(3)).toBe(FEW)
    expect(pl(4)).toBe(FEW)
  })

  it("5-10 → many", () => {
    for (const n of [5, 6, 7, 8, 9, 10]) {
      expect(pl(n)).toBe(MANY)
    }
  })

  it("11-14 → many (исключение из правила «оканчивается на 1/2-4»)", () => {
    expect(pl(11)).toBe(MANY)
    expect(pl(12)).toBe(MANY)
    expect(pl(13)).toBe(MANY)
    expect(pl(14)).toBe(MANY)
  })

  it("15-20 → many", () => {
    for (const n of [15, 16, 17, 18, 19, 20]) {
      expect(pl(n)).toBe(MANY)
    }
  })

  it("21 → one (заканчивается на 1, но не 11)", () => {
    expect(pl(21)).toBe(ONE)
  })

  it("22-24 → few", () => {
    expect(pl(22)).toBe(FEW)
    expect(pl(23)).toBe(FEW)
    expect(pl(24)).toBe(FEW)
  })

  it("25-30 → many", () => {
    for (const n of [25, 26, 27, 28, 29, 30]) {
      expect(pl(n)).toBe(MANY)
    }
  })

  it("100 → many; 101 → one; 102-104 → few; 111-114 → many", () => {
    expect(pl(100)).toBe(MANY)
    expect(pl(101)).toBe(ONE)
    expect(pl(102)).toBe(FEW)
    expect(pl(104)).toBe(FEW)
    expect(pl(111)).toBe(MANY)
    expect(pl(112)).toBe(MANY)
    expect(pl(114)).toBe(MANY)
  })

  it("1000+ — продолжает работать корректно", () => {
    expect(pl(1000)).toBe(MANY)
    expect(pl(1001)).toBe(ONE)
    expect(pl(1002)).toBe(FEW)
    expect(pl(1011)).toBe(MANY)
    expect(pl(1021)).toBe(ONE)
  })

  it("отрицательные нормализуются через Math.abs", () => {
    expect(pl(-1)).toBe(ONE)
    expect(pl(-2)).toBe(FEW)
    expect(pl(-5)).toBe(MANY)
  })

  it("дробные округляются вниз через Math.trunc", () => {
    expect(pl(1.5)).toBe(ONE)
    expect(pl(2.9)).toBe(FEW)
    expect(pl(0.5)).toBe(MANY)
  })
})
