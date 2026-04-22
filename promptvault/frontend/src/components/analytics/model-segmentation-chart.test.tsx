import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ModelSegmentationChart } from "./model-segmentation-chart"

// Регрессия на B.7 (segmentation по AI-моделям): пустой массив не должен
// крашить, известные модели получают русскую обёртку «Без модели» для "",
// хвост (>6) агрегируется в «Другие».

afterEach(() => cleanup())

describe("ModelSegmentationChart", () => {
  it("пустые данные — показывает fallback", () => {
    render(<ModelSegmentationChart data={[]} />)
    expect(screen.getByText("Использование по моделям")).toBeInTheDocument()
    expect(screen.getByText(/Пока нет данных/)).toBeInTheDocument()
  })

  it("рендерит легенду с процентами", () => {
    render(
      <ModelSegmentationChart
        data={[
          { model: "claude-sonnet-4", uses: 60 },
          { model: "gpt-4o", uses: 30 },
          { model: "deepseek", uses: 10 },
        ]}
      />,
    )
    expect(screen.getByText("claude-sonnet-4")).toBeInTheDocument()
    expect(screen.getByText("gpt-4o")).toBeInTheDocument()
    expect(screen.getByText("deepseek")).toBeInTheDocument()
    // Проценты
    expect(screen.getByText(/60.*60%/)).toBeInTheDocument()
    expect(screen.getByText(/30.*30%/)).toBeInTheDocument()
    expect(screen.getByText(/10.*10%/)).toBeInTheDocument()
  })

  it("пустая model заменяется на «Без модели»", () => {
    render(<ModelSegmentationChart data={[{ model: "", uses: 5 }]} />)
    expect(screen.getByText("Без модели")).toBeInTheDocument()
  })

  it("более 6 моделей — хвост агрегируется в «Другие»", () => {
    const data = Array.from({ length: 9 }, (_, i) => ({
      model: `model-${i}`,
      uses: 10,
    }))
    render(<ModelSegmentationChart data={data} />)
    // Первые 6 названий видны, «Другие» для 7-9.
    expect(screen.getByText("model-0")).toBeInTheDocument()
    expect(screen.getByText("model-5")).toBeInTheDocument()
    expect(screen.getByText("Другие")).toBeInTheDocument()
    // model-7/8 не показаны отдельно
    expect(screen.queryByText("model-7")).toBeNull()
  })

  it("кастомный title рендерится в заголовке", () => {
    render(
      <ModelSegmentationChart
        title="Моё распределение"
        data={[{ model: "claude", uses: 1 }]}
      />,
    )
    expect(screen.getByText("Моё распределение")).toBeInTheDocument()
  })
})
