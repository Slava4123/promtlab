import { describe, it, expect, beforeAll, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { UsageChart } from "./usage-chart"
import { createUsageChartConfig } from "./usage-chart-config"
import { formatDayShort } from "@/lib/date-format"

// jsdom не имеет ResizeObserver, на котором держится recharts ResponsiveContainer.
// Достаточно no-op стаба, чтобы компонент не крашился при mount с непустым data.
beforeAll(() => {
  if (typeof globalThis.ResizeObserver === "undefined") {
    globalThis.ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    } as unknown as typeof ResizeObserver
  }
})

afterEach(() => cleanup())

describe("createUsageChartConfig", () => {
  it("вписывает переданный лейбл в config.count.label", () => {
    const config = createUsageChartConfig("Создано")
    expect(config.count.label).toBe("Создано")
  })

  it("сохраняет цветовую переменную, идентичную всем инстансам", () => {
    const a = createUsageChartConfig("Использования")
    const b = createUsageChartConfig("Создано")
    expect(a.count.color).toBe("var(--chart-1)")
    expect(b.count.color).toBe("var(--chart-1)")
  })

  it("разные лейблы не интерферируют между инстансами", () => {
    const a = createUsageChartConfig("Использования")
    const b = createUsageChartConfig("Создано")
    expect(a.count.label).not.toBe(b.count.label)
  })
})

describe("UsageChart", () => {
  it("рендерит title и empty-label при пустых данных", () => {
    render(<UsageChart title="Использование по дням" data={[]} />)
    expect(screen.getByText("Использование по дням")).toBeInTheDocument()
    expect(screen.getByText(/Пока нет данных/i)).toBeInTheDocument()
  })

  it("принимает кастомный emptyLabel", () => {
    render(<UsageChart title="Создано" data={[]} emptyLabel="Нет точек данных" />)
    expect(screen.getByText("Нет точек данных")).toBeInTheDocument()
  })

  it("монтируется при непустых данных и custom chartConfig без exception", () => {
    // Recharts в jsdom не отрисовывает SVG (нет getBoundingClientRect размеров),
    // зато падает с exception при некорректной конфигурации. Достаточно того,
    // что title и Card видны — это значит компонент дошёл до children без throw.
    render(
      <UsageChart
        title="Создание промптов по дням"
        data={[
          { day: "2026-04-12T00:00:00Z", count: 1 },
          { day: "2026-04-13T00:00:00Z", count: 2 },
        ]}
        chartConfig={createUsageChartConfig("Создано")}
      />,
    )
    expect(screen.getByText("Создание промптов по дням")).toBeInTheDocument()
  })

  it("formatDayShort преобразует ISO в русский короткий формат '7 мая'", () => {
    // Recharts в jsdom не отрисовывает <text> внутри SVG (нет размеров ResponsiveContainer),
    // поэтому container.textContent пуст на тиках. Контракт XAxis tickFormatter
    // покрываем через прямой вызов formatDayShort — той же функции, которую
    // компонент передаёт в <XAxis tickFormatter={formatDayShort} />.
    expect(formatDayShort("2026-05-07")).toBe("7 мая")
    expect(formatDayShort("2026-05-16")).toBe("16 мая")
  })

  it("x-axis tickFormatter в коде использует formatDayShort (а не v.slice(5))", async () => {
    // Регресс-гард: source-of-truth — содержание usage-chart.tsx. Если кто-то
    // вернёт `tickFormatter={(v) => v.slice(5)}`, тест упадёт. Достаточно lightweight,
    // потому что jsdom + ResponsiveContainer не дают надёжно проверить SVG-тики.
    const src = await import("./usage-chart.tsx?raw").then((m) => m.default as string)
    expect(src).toMatch(/tickFormatter=\{formatDayShort\}/)
    expect(src).not.toMatch(/tickFormatter=\{\(v\)\s*=>\s*v\.slice\(5\)/)
  })
})
