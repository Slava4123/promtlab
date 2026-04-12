import { render, screen } from "@testing-library/react"
import { describe, it, expect } from "vitest"

import { BadgeCard } from "./badge-card"
import type { Badge } from "@/api/types"

const lockedBadge: Badge = {
  id: "architect",
  title: "Архитектор",
  description: "Создай 10 личных промптов",
  icon: "🏗️",
  category: "personal",
  unlocked: false,
  progress: 3,
  target: 10,
}

const unlockedBadge: Badge = {
  id: "first_prompt",
  title: "Первопроходец",
  description: "Создай первый личный промпт",
  icon: "🎯",
  category: "personal",
  unlocked: true,
  unlocked_at: "2026-04-01T12:00:00Z",
  progress: 1,
  target: 1,
}

describe("BadgeCard", () => {
  it("рендерит locked бейдж с прогрессом", () => {
    render(<BadgeCard badge={lockedBadge} />)
    expect(screen.getByText("Архитектор")).toBeInTheDocument()
    expect(screen.getByText("3/10")).toBeInTheDocument()
    expect(screen.getByText("Прогресс")).toBeInTheDocument()
  })

  it("рендерит unlocked бейдж с датой", () => {
    render(<BadgeCard badge={unlockedBadge} />)
    expect(screen.getByText("Первопроходец")).toBeInTheDocument()
    expect(screen.getByText(/Разблокировано/)).toBeInTheDocument()
    expect(screen.queryByText("Прогресс")).not.toBeInTheDocument()
  })
})
