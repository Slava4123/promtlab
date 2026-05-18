# Analytics Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Заменить vertical-stack layout `/analytics` на современный Bento Grid с AI Narrative Banner, 4 KPI-карточек с inline sparklines, actionable Smart Insights ribbon, donut chart для моделей, GitHub-style activity heatmap, streak tracker и compact quotas footer. Все эмодзи → Lucide-иконки, semantic colors через CSS-token'ы.

**Architecture:** Frontend-only refactor. Backend hooks (`usePersonalAnalytics`, `useInsights`, `useStreak`) не меняются. Создаём 9 новых компонентов и pure-function `buildNarrative` в `frontend/src/components/analytics/` и `frontend/src/lib/`. Реструктурируем `frontend/src/pages/analytics.tsx`. Пipeline TDD: для каждого компонента — failing Vitest test → minimal impl → passing test → commit.

**Tech Stack:** React 19.2, Vite 8, TypeScript strict, Tailwind v4 (inline config в `index.css`), shadcn/ui (Card/Skeleton), Lucide React (single icon lib), Recharts (для donut через `@/components/ui/chart` wrapper), TanStack Query (existing). Vitest + Testing Library + jest-dom.

**Дизайн-спека:** [docs/superpowers/specs/2026-05-17-analytics-redesign-design.md](../specs/2026-05-17-analytics-redesign-design.md)

---

## File Structure

### Создаём

| Путь | Назначение |
|---|---|
| `frontend/src/components/analytics/sparkline.tsx` | Custom SVG polyline ~40 строк, без Recharts overhead |
| `frontend/src/components/analytics/sparkline.test.tsx` | Unit-тест (3 случая: empty, single-point, normal) |
| `frontend/src/components/analytics/kpi-card.tsx` | KPI card: label + icon + value + delta + sparkline |
| `frontend/src/components/analytics/kpi-card.test.tsx` | Unit-тест (with/without sparkline, delta states) |
| `frontend/src/components/analytics/insight-action-card.tsx` | Color-coded actionable card (warning/info/success), CTA с ArrowRight |
| `frontend/src/components/analytics/insight-action-card.test.tsx` | Unit-тест по тонам, count badge, href |
| `frontend/src/components/analytics/activity-heatmap.tsx` | GitHub-style 4w × 7d grid, opacity по count |
| `frontend/src/components/analytics/activity-heatmap.test.tsx` | Empty/partial/full points |
| `frontend/src/components/analytics/models-donut.tsx` | Recharts PieChart innerRadius=60% для моделей |
| `frontend/src/components/analytics/models-donut.test.tsx` | Empty + top-6 + others |
| `frontend/src/components/analytics/streak-tracker.tsx` | Current streak + 7d dots calendar |
| `frontend/src/components/analytics/streak-tracker.test.tsx` | Render с current/longest streak |
| `frontend/src/components/analytics/compact-quotas.tsx` | Однострочный quota footer, 3 inline progress |
| `frontend/src/components/analytics/compact-quotas.test.tsx` | Render с тремя quotas |
| `frontend/src/components/analytics/narrative-banner.tsx` | AI-style summary banner |
| `frontend/src/components/analytics/narrative-banner.test.tsx` | Render с / без insights |
| `frontend/src/lib/analytics-narrative.ts` | Pure function `buildNarrative(data, insights)` |
| `frontend/src/lib/analytics-narrative.test.ts` | Table-driven, 6+ кейсов |
| `frontend/src/components/analytics/model-colors.ts` | Экспортированные `MODEL_COLORS`, `DEFAULT_COLOR`, `colorFor`, `labelFor`, `UNKNOWN_MODEL_HINT` |

### Меняем

| Путь | Что меняем |
|---|---|
| `frontend/src/pages/analytics.tsx` | Полный rewrite JSX body, новый Bento Grid layout |
| `frontend/src/components/analytics/insights-panel.tsx` | Рендерит `InsightActionCard` для каждого `Insight.type` |
| `frontend/src/components/analytics/insights-locked-card.tsx` | Restyle под visual consistency (dashed border, lock-icon) |
| `frontend/src/components/analytics/model-segmentation-chart.tsx` | Импортирует из `model-colors.ts` (вместо локальных const) — для backward-compat если он ещё где-то используется |
| `frontend/src/pages/__tests__/analytics-insights-states.test.tsx` | Обновить селекторы под новые компоненты |

### Не трогаем (existing, переиспользуем)

- `frontend/src/components/analytics/usage-chart.tsx` — оставляем как есть, рисуется в Bento Grid (4×2).
- `frontend/src/components/analytics/top-prompts-table.tsx` — рисуется в Bento Grid (6×1).
- `frontend/src/components/analytics/range-picker.tsx` — header.
- `frontend/src/components/analytics/upgrade-gate.tsx` — Free three-state.
- `frontend/src/hooks/use-analytics.ts`, `use-streaks.ts` — no changes.

---

## Phase 1: Foundation (independent units)

### Task 1: Extract MODEL_COLORS to shared module

**Files:**
- Create: `frontend/src/components/analytics/model-colors.ts`
- Modify: `frontend/src/components/analytics/model-segmentation-chart.tsx` (импортировать вместо локальных const)

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/model-colors.test.ts`:

```ts
import { describe, it, expect } from "vitest"
import { colorFor, labelFor, MODEL_COLORS, DEFAULT_COLOR, UNKNOWN_MODEL_HINT } from "./model-colors"

describe("model-colors", () => {
  it("matches Claude variants to Anthropic orange", () => {
    expect(colorFor("claude-3-opus")).toBe("#cc7a3e")
    expect(colorFor("Claude")).toBe("#cc7a3e")
  })

  it("matches GPT variants to OpenAI green", () => {
    expect(colorFor("gpt-4-turbo")).toBe("#10a37f")
  })

  it("returns DEFAULT_COLOR for unknown models", () => {
    expect(colorFor("custom-llm")).toBe(DEFAULT_COLOR)
    expect(colorFor("")).toBe(DEFAULT_COLOR)
  })

  it("labels empty model as «Модель не указана»", () => {
    expect(labelFor("")).toBe("Модель не указана")
    expect(labelFor("claude")).toBe("claude")
  })

  it("exposes UNKNOWN_MODEL_HINT", () => {
    expect(UNKNOWN_MODEL_HINT).toContain("при создании")
  })

  it("exports MODEL_COLORS array", () => {
    expect(MODEL_COLORS.length).toBeGreaterThanOrEqual(6)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd C:/GolandProjects/awesomeProject/test/promptvault/frontend && npx vitest run model-colors`
Expected: FAIL — `model-colors` module not found.

- [ ] **Step 3: Create shared module**

Create `frontend/src/components/analytics/model-colors.ts`:

```ts
// Palette для известных семейств моделей. Неузнанные модели получают серый.
// Извлечено из model-segmentation-chart.tsx для переиспользования
// в models-donut.tsx (analytics redesign 2026-05-17).
export const MODEL_COLORS: Array<[RegExp, string]> = [
  [/^claude/i, "#cc7a3e"], // оранж-коричневый как Anthropic brand
  [/^gpt/i, "#10a37f"], // зелёный как OpenAI
  [/deepseek/i, "#4a7fff"], // синий
  [/gemini|google/i, "#8ab4f8"],
  [/llama|meta/i, "#0668e1"],
  [/mistral/i, "#ff7000"],
]

export const DEFAULT_COLOR = "#94a3b8" // серый для «Без модели» и неопознанных

export const UNKNOWN_MODEL_HINT =
  "Промпты, в которых при создании не указана target-модель в редакторе"

export function colorFor(model: string): string {
  for (const [re, color] of MODEL_COLORS) {
    if (re.test(model)) return color
  }
  return DEFAULT_COLOR
}

// Backend агрегирует строки с пустой `prompts.model` под пустой строкой.
// Для пользователя показываем расшифровку — это не legacy и не баг.
export function labelFor(model: string): string {
  return model === "" ? "Модель не указана" : model
}
```

- [ ] **Step 4: Update model-segmentation-chart.tsx to import**

Open `frontend/src/components/analytics/model-segmentation-chart.tsx`. Удалить local `MODEL_COLORS`, `DEFAULT_COLOR`, `colorFor`, `labelFor`, `UNKNOWN_MODEL_HINT` (строки ~12-40). Заменить на:

```ts
import { colorFor, labelFor, DEFAULT_COLOR, UNKNOWN_MODEL_HINT } from "./model-colors"
```

Сохранить остальное тело компонента.

- [ ] **Step 5: Run tests to verify pass**

```bash
npx vitest run model-colors model-segmentation-chart
```

Expected: PASS (oба зелёных).

- [ ] **Step 6: Commit**

```bash
cd C:/GolandProjects/awesomeProject/test
git add promptvault/frontend/src/components/analytics/model-colors.ts \
        promptvault/frontend/src/components/analytics/model-colors.test.ts \
        promptvault/frontend/src/components/analytics/model-segmentation-chart.tsx
git commit -m "refactor(analytics): extract MODEL_COLORS to shared model-colors.ts"
```

---

### Task 2: Sparkline component

**Files:**
- Create: `frontend/src/components/analytics/sparkline.tsx`
- Create: `frontend/src/components/analytics/sparkline.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/sparkline.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, cleanup } from "@testing-library/react"
import { Sparkline } from "./sparkline"

afterEach(() => cleanup())

describe("Sparkline", () => {
  it("renders nothing for empty points", () => {
    const { container } = render(<Sparkline points={[]} />)
    expect(container.querySelector("svg")).toBeNull()
  })

  it("renders SVG with polyline for normal points", () => {
    const { container } = render(<Sparkline points={[1, 3, 2, 5, 8]} />)
    const svg = container.querySelector("svg")
    expect(svg).not.toBeNull()
    expect(svg?.querySelector("polyline")).not.toBeNull()
  })

  it("uses emerald color for up trend", () => {
    const { container } = render(<Sparkline points={[1, 5]} trend="up" />)
    const polyline = container.querySelector("polyline")
    expect(polyline?.getAttribute("stroke")).toMatch(/emerald|#10b981/i)
  })

  it("uses rose color for down trend", () => {
    const { container } = render(<Sparkline points={[5, 1]} trend="down" />)
    const polyline = container.querySelector("polyline")
    expect(polyline?.getAttribute("stroke")).toMatch(/rose|#ef4444/i)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run sparkline`
Expected: FAIL — `Sparkline` not exported.

- [ ] **Step 3: Implement Sparkline**

Create `frontend/src/components/analytics/sparkline.tsx`:

```tsx
interface SparklineProps {
  points: number[]
  trend?: "up" | "down" | "neutral"
  width?: number
  height?: number
}

// Sparkline — мини-график для KPI-карточки. Чистый SVG, без Recharts overhead.
// trend влияет на цвет: up → emerald, down → rose, neutral → slate.
// fill — лёгкий gradient под линией (соответствует цвету stroke на 15%).
export function Sparkline({
  points,
  trend = "neutral",
  width = 120,
  height = 22,
}: SparklineProps) {
  if (points.length === 0) return null

  const stroke =
    trend === "up" ? "#10b981" : trend === "down" ? "#ef4444" : "#94a3b8"
  const fill =
    trend === "up"
      ? "rgba(16,185,129,0.15)"
      : trend === "down"
        ? "rgba(239,68,68,0.15)"
        : "rgba(148,163,184,0.15)"

  const max = Math.max(...points, 1)
  const min = Math.min(...points, 0)
  const range = max - min || 1
  const stepX = points.length > 1 ? width / (points.length - 1) : width

  const coords = points.map((p, i) => {
    const x = i * stepX
    const y = height - ((p - min) / range) * (height - 2) - 1
    return `${x},${y}`
  })

  const polyPoints = coords.join(" ")
  const areaPoints = `0,${height} ${polyPoints} ${width},${height}`

  return (
    <svg width={width} height={height} aria-hidden="true">
      <polyline points={areaPoints} fill={fill} stroke="none" />
      <polyline points={polyPoints} fill="none" stroke={stroke} strokeWidth="1.5" />
    </svg>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run sparkline`
Expected: PASS — 4/4.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/sparkline.tsx \
        promptvault/frontend/src/components/analytics/sparkline.test.tsx
git commit -m "feat(analytics): Sparkline component (custom SVG polyline)"
```

---

### Task 3: buildNarrative pure function

**Files:**
- Create: `frontend/src/lib/analytics-narrative.ts`
- Create: `frontend/src/lib/analytics-narrative.test.ts`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/lib/analytics-narrative.test.ts`:

```ts
import { describe, it, expect } from "vitest"
import { buildNarrative } from "./analytics-narrative"
import type { PersonalDashboard } from "@/api/analytics"
import type { Insight } from "@/api/types"

const baseDashboard: PersonalDashboard = {
  range: "7d",
  usage_per_day: [],
  top_prompts: [],
  prompts_created_per_day: [],
  prompts_updated_per_day: [],
  share_views_per_day: [],
  top_shared: [],
  totals_current: { uses: 234, created: 12, updated: 0, share_views: 89 },
  totals_previous: { uses: 190, created: 10, updated: 0, share_views: 96 },
  usage_by_model: [
    { model: "claude-3-opus", uses: 145 },
    { model: "gpt-4", uses: 65 },
    { model: "gemini-pro", uses: 24 },
  ],
}

describe("buildNarrative", () => {
  it("includes period and delta in summary for non-zero uses", () => {
    const result = buildNarrative(baseDashboard, null)
    expect(result.summary).toContain("234")
    expect(result.summary).toMatch(/\+23%|↑23%/)
  })

  it("returns quiet copy for zero uses", () => {
    const empty: PersonalDashboard = {
      ...baseDashboard,
      totals_current: { uses: 0, created: 0, updated: 0, share_views: 0 },
      totals_previous: { uses: 0, created: 0, updated: 0, share_views: 0 },
    }
    const result = buildNarrative(empty, null)
    expect(result.summary).toMatch(/тих|пуст/i)
  })

  it("returns topModel as Claude with percentage when dominant", () => {
    const result = buildNarrative(baseDashboard, null)
    expect(result.topModel).toMatch(/Claude/i)
    expect(result.topModel).toMatch(/62/)
  })

  it("returns null topModel when usage_by_model is empty", () => {
    const empty: PersonalDashboard = { ...baseDashboard, usage_by_model: [] }
    const result = buildNarrative(empty, null)
    expect(result.topModel).toBeNull()
  })

  it("returns actionHint when insights contain unused_prompts and possible_duplicates", () => {
    const insights: Insight[] = [
      { type: "unused_prompts", payload: [1, 2, 3, 4, 5], computed_at: "" },
      { type: "possible_duplicates", payload: [1, 2], computed_at: "" },
    ]
    const result = buildNarrative(baseDashboard, insights)
    expect(result.actionHint).toMatch(/5/)
    expect(result.actionHint).toMatch(/2/)
  })

  it("returns null actionHint for empty insights", () => {
    const result = buildNarrative(baseDashboard, [])
    expect(result.actionHint).toBeNull()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run analytics-narrative`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement buildNarrative**

Create `frontend/src/lib/analytics-narrative.ts`:

```ts
import type { PersonalDashboard } from "@/api/analytics"
import type { Insight } from "@/api/types"
import { formatRange } from "@/api/analytics"

export interface NarrativeSegments {
  summary: string
  topModel: string | null
  streak: string | null
  actionHint: string | null
}

// buildNarrative — template-функция для AI-style summary без LLM-вызовов.
// Принцип «без AI на нашей стороне» из CLAUDE.md: текст детерминирован.
// Каждый сегмент опциональный — может быть null если данных нет.
export function buildNarrative(
  data: PersonalDashboard,
  insights: Insight[] | null,
): NarrativeSegments {
  return {
    summary: buildSummary(data),
    topModel: buildTopModel(data),
    streak: null, // заполняется в narrative-banner.tsx через useStreak hook
    actionHint: buildActionHint(insights),
  }
}

function buildSummary(data: PersonalDashboard): string {
  const period = formatRange(data.range)
  const uses = data.totals_current.uses
  if (uses === 0) {
    return `За ${period} пока тихо — самое время попробовать новые промпты`
  }
  const prev = data.totals_previous.uses
  const deltaText = formatDelta(uses, prev)
  return `За ${period}: ${uses.toLocaleString("ru")} использований${deltaText}`
}

function buildTopModel(data: PersonalDashboard): string | null {
  if (data.usage_by_model.length === 0) return null
  const total = data.usage_by_model.reduce((s, r) => s + r.uses, 0)
  if (total === 0) return null
  const top = [...data.usage_by_model].sort((a, b) => b.uses - a.uses)[0]
  const pct = Math.round((top.uses / total) * 100)
  const name = top.model === "" ? "Без модели" : top.model
  return `топ-модель ${name} (${pct}%)`
}

function buildActionHint(insights: Insight[] | null): string | null {
  if (!insights || insights.length === 0) return null
  const parts: string[] = []
  for (const ins of insights) {
    const count = Array.isArray(ins.payload) ? ins.payload.length : 0
    if (count === 0) continue
    if (ins.type === "unused_prompts") parts.push(`${count} забытых`)
    else if (ins.type === "possible_duplicates") parts.push(`${count} дубликата`)
    else if (ins.type === "orphan_tags") parts.push(`${count} orphan-тегов`)
    else if (ins.type === "empty_collections") parts.push(`${count} пустых коллекций`)
  }
  if (parts.length === 0) return null
  return `${parts.join(" и ")} ждут уборки`
}

function formatDelta(current: number, previous: number): string {
  if (previous === 0) return ""
  const pct = Math.round(((current - previous) / previous) * 100)
  if (pct === 0) return " (без изменений)"
  return pct > 0 ? ` ↑${pct}%` : ` ↓${Math.abs(pct)}%`
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run analytics-narrative`
Expected: PASS — 6/6.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/lib/analytics-narrative.ts \
        promptvault/frontend/src/lib/analytics-narrative.test.ts
git commit -m "feat(analytics): buildNarrative pure-function (template, no LLM)"
```

---

### Task 4: NarrativeBanner component

**Files:**
- Create: `frontend/src/components/analytics/narrative-banner.tsx`
- Create: `frontend/src/components/analytics/narrative-banner.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/narrative-banner.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { NarrativeBanner } from "./narrative-banner"
import type { NarrativeSegments } from "@/lib/analytics-narrative"

afterEach(() => cleanup())

describe("NarrativeBanner", () => {
  it("renders summary text", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней: 234 использований ↑23%",
      topModel: null,
      streak: null,
      actionHint: null,
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/234 использований/)).toBeInTheDocument()
  })

  it("renders all 4 segments when provided", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней: 234 использований ↑23%",
      topModel: "топ-модель Claude (62%)",
      streak: "streak 5 дней",
      actionHint: "5 забытых ждут уборки",
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/Claude/)).toBeInTheDocument()
    expect(screen.getByText(/streak 5 дней/)).toBeInTheDocument()
    expect(screen.getByText(/5 забытых/)).toBeInTheDocument()
  })

  it("omits null segments gracefully", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней пока тихо",
      topModel: null,
      streak: null,
      actionHint: null,
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/тихо/)).toBeInTheDocument()
    expect(screen.queryByText(/streak/)).toBeNull()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run narrative-banner`
Expected: FAIL — `NarrativeBanner` not exported.

- [ ] **Step 3: Implement NarrativeBanner**

Create `frontend/src/components/analytics/narrative-banner.tsx`:

```tsx
import { Sparkles, ArrowRight } from "lucide-react"
import type { NarrativeSegments } from "@/lib/analytics-narrative"

interface NarrativeBannerProps {
  segments: NarrativeSegments
  href?: string
}

// NarrativeBanner — top-of-page AI-style summary без LLM-вызова.
// Сегменты собираются в buildNarrative() из existing data.
// Визуально: violet gradient + Sparkles icon + ArrowRight на CTA если есть actionHint.
export function NarrativeBanner({ segments, href = "/analytics" }: NarrativeBannerProps) {
  const hasAction = segments.actionHint !== null

  return (
    <div className="flex items-center gap-3 rounded-lg border border-violet-500/25 bg-gradient-to-r from-violet-500/10 to-violet-500/5 px-4 py-3">
      <Sparkles className="size-5 shrink-0 text-violet-500" aria-hidden="true" />
      <div className="flex-1 text-sm leading-relaxed">
        <span className="font-medium">{segments.summary}</span>
        {segments.topModel && (
          <>
            <span className="mx-1.5 text-muted-foreground">·</span>
            <span>{segments.topModel}</span>
          </>
        )}
        {segments.streak && (
          <>
            <span className="mx-1.5 text-muted-foreground">·</span>
            <span>{segments.streak}</span>
          </>
        )}
        {hasAction && (
          <div className="mt-0.5 text-xs text-muted-foreground">{segments.actionHint}</div>
        )}
      </div>
      {hasAction && (
        <a
          href={href}
          className="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
          aria-label="Подробнее об инсайтах"
        >
          <ArrowRight className="size-4" />
        </a>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run narrative-banner`
Expected: PASS — 3/3.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/narrative-banner.tsx \
        promptvault/frontend/src/components/analytics/narrative-banner.test.tsx
git commit -m "feat(analytics): NarrativeBanner component (Sparkles + 4 segments)"
```

---

## Phase 2: KPI и Smart Insights cards

### Task 5: KpiCard component

**Files:**
- Create: `frontend/src/components/analytics/kpi-card.tsx`
- Create: `frontend/src/components/analytics/kpi-card.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/kpi-card.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { Activity } from "lucide-react"
import { KpiCard } from "./kpi-card"

afterEach(() => cleanup())

describe("KpiCard", () => {
  it("renders label, value, and icon", () => {
    render(<KpiCard label="Использования" value={234} icon={Activity} />)
    expect(screen.getByText("Использования")).toBeInTheDocument()
    expect(screen.getByText("234")).toBeInTheDocument()
  })

  it("shows up arrow for positive delta", () => {
    const { container } = render(
      <KpiCard label="X" value={100} delta={23} icon={Activity} />,
    )
    expect(screen.getByText(/23%/)).toBeInTheDocument()
    expect(container.querySelector(".text-emerald-600, .dark\\:text-emerald-400")).not.toBeNull()
  })

  it("shows down arrow for negative delta", () => {
    const { container } = render(
      <KpiCard label="X" value={100} delta={-8} icon={Activity} />,
    )
    expect(screen.getByText(/8%/)).toBeInTheDocument()
    expect(container.querySelector(".text-rose-600, .dark\\:text-rose-400")).not.toBeNull()
  })

  it("renders sparkline when points provided", () => {
    const { container } = render(
      <KpiCard label="X" value={100} sparkline={[1, 3, 5, 8]} icon={Activity} />,
    )
    expect(container.querySelector("svg")).not.toBeNull()
  })

  it("renders «—» when delta is null", () => {
    render(<KpiCard label="X" value={0} delta={null} icon={Activity} />)
    expect(screen.getByText("—")).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run kpi-card`
Expected: FAIL — `KpiCard` not exported.

- [ ] **Step 3: Implement KpiCard**

Create `frontend/src/components/analytics/kpi-card.tsx`:

```tsx
import { ArrowUp, ArrowDown, type LucideIcon } from "lucide-react"
import { Card } from "@/components/ui/card"
import { Sparkline } from "./sparkline"
import { cn } from "@/lib/utils"

interface KpiCardProps {
  label: string
  value: string | number
  delta?: number | null
  sparkline?: number[]
  icon: LucideIcon
  className?: string
}

// KpiCard — расширение MetricCard: добавлены icon и sparkline.
// Layout: label сверху (uppercase muted) + icon справа, value крупно,
// delta inline с ArrowUp/Down, sparkline снизу.
// Цвета delta: raw emerald/rose для консистентности с metric-card (см. CLAUDE.md).
export function KpiCard({ label, value, delta, sparkline, icon: Icon, className }: KpiCardProps) {
  const trend = !delta ? "neutral" : delta > 0 ? "up" : "down"

  return (
    <Card className={cn("p-4", className)}>
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">{label}</span>
        <Icon className="size-4 text-muted-foreground" aria-hidden="true" />
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-bold tabular-nums">{value}</span>
        <DeltaInline delta={delta} />
      </div>
      {sparkline && sparkline.length > 0 && (
        <div className="mt-2">
          <Sparkline points={sparkline} trend={trend} />
        </div>
      )}
    </Card>
  )
}

function DeltaInline({ delta }: { delta: number | null | undefined }) {
  if (delta === null) return <span className="text-xs text-muted-foreground">—</span>
  if (delta === undefined || delta === 0)
    return <span className="text-xs text-muted-foreground">≡ 0%</span>
  const up = delta > 0
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 text-xs font-medium",
        up ? "text-emerald-600 dark:text-emerald-400" : "text-rose-600 dark:text-rose-400",
      )}
    >
      {up ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />}
      {Math.abs(delta)}%
    </span>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run kpi-card`
Expected: PASS — 5/5.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/kpi-card.tsx \
        promptvault/frontend/src/components/analytics/kpi-card.test.tsx
git commit -m "feat(analytics): KpiCard with sparkline + Lucide icon"
```

---

### Task 6: InsightActionCard component

**Files:**
- Create: `frontend/src/components/analytics/insight-action-card.tsx`
- Create: `frontend/src/components/analytics/insight-action-card.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/insight-action-card.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { AlertCircle, Copy, TrendingUp } from "lucide-react"
import { InsightActionCard } from "./insight-action-card"

afterEach(() => cleanup())

describe("InsightActionCard", () => {
  it("renders title, description, count, and CTA", () => {
    render(
      <InsightActionCard
        tone="warning"
        icon={AlertCircle}
        title="Забытые"
        description="5 промптов не использовались 30+ дней"
        href="/prompts?filter=unused"
        count={5}
        ctaLabel="Посмотреть"
      />,
    )
    expect(screen.getByText("Забытые")).toBeInTheDocument()
    expect(screen.getByText(/30\+ дней/)).toBeInTheDocument()
    expect(screen.getByText("5")).toBeInTheDocument()
    expect(screen.getByRole("link", { name: /Посмотреть/ })).toHaveAttribute("href", "/prompts?filter=unused")
  })

  it("applies warning tone classes", () => {
    const { container } = render(
      <InsightActionCard
        tone="warning" icon={AlertCircle} title="X" description="Y" href="#" ctaLabel="→"
      />,
    )
    expect(container.querySelector(".border-amber-500\\/30, .bg-amber-500\\/8, .text-amber-500")).not.toBeNull()
  })

  it("applies info tone classes", () => {
    const { container } = render(
      <InsightActionCard tone="info" icon={Copy} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.querySelector(".border-violet-500\\/30, .text-violet-500, .border-indigo-500\\/30, .text-indigo-500")).not.toBeNull()
  })

  it("applies success tone classes", () => {
    const { container } = render(
      <InsightActionCard tone="success" icon={TrendingUp} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.querySelector(".border-emerald-500\\/30, .text-emerald-500")).not.toBeNull()
  })

  it("omits count when not provided", () => {
    const { container } = render(
      <InsightActionCard tone="warning" icon={AlertCircle} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.textContent).not.toContain("0")
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run insight-action-card`
Expected: FAIL — module not exported.

- [ ] **Step 3: Implement InsightActionCard**

Create `frontend/src/components/analytics/insight-action-card.tsx`:

```tsx
import { ArrowRight, type LucideIcon } from "lucide-react"
import { Link } from "react-router-dom"
import { cn } from "@/lib/utils"

export type InsightTone = "warning" | "info" | "success"

interface InsightActionCardProps {
  tone: InsightTone
  icon: LucideIcon
  title: string
  description: string
  href: string
  ctaLabel: string
  count?: number
}

const TONE_CLASSES: Record<InsightTone, string> = {
  warning: "border-amber-500/30 bg-amber-500/8 text-amber-500",
  info: "border-violet-500/30 bg-violet-500/8 text-violet-500",
  success: "border-emerald-500/30 bg-emerald-500/8 text-emerald-500",
}

const TONE_LABELS: Record<InsightTone, string> = {
  warning: "Внимание",
  info: "Подсказка",
  success: "Растёт",
}

// InsightActionCard — actionable card для Smart Insights items.
// Цвет отражает tone: warning=amber, info=violet, success=emerald.
// CTA → React Router Link (deep link на конкретный фильтр/раздел).
export function InsightActionCard({
  tone,
  icon: Icon,
  title,
  description,
  href,
  ctaLabel,
  count,
}: InsightActionCardProps) {
  const toneClass = TONE_CLASSES[tone]
  return (
    <div className={cn("rounded-lg border p-4", toneClass.split(" ").slice(0, 2).join(" "))}>
      <div className="mb-1.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon className={cn("size-4", toneClass.split(" ")[2])} aria-hidden="true" />
          <span className={cn("text-[11px] font-semibold uppercase tracking-wide", toneClass.split(" ")[2])}>
            {title || TONE_LABELS[tone]}
          </span>
        </div>
        {count !== undefined && (
          <span className="rounded-full bg-foreground/10 px-2 py-0.5 text-[11px] font-medium tabular-nums">
            {count}
          </span>
        )}
      </div>
      <p className="mb-2 text-sm text-foreground/90">{description}</p>
      <Link
        to={href}
        className="inline-flex items-center gap-1 text-xs font-medium text-foreground/80 transition-colors hover:text-foreground"
      >
        {ctaLabel}
        <ArrowRight className="size-3" aria-hidden="true" />
      </Link>
    </div>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run insight-action-card`
Expected: PASS — 5/5.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/insight-action-card.tsx \
        promptvault/frontend/src/components/analytics/insight-action-card.test.tsx
git commit -m "feat(analytics): InsightActionCard with warning/info/success tones"
```

---

### Task 7: ActivityHeatmap component

**Files:**
- Create: `frontend/src/components/analytics/activity-heatmap.tsx`
- Create: `frontend/src/components/analytics/activity-heatmap.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/activity-heatmap.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ActivityHeatmap } from "./activity-heatmap"

afterEach(() => cleanup())

describe("ActivityHeatmap", () => {
  it("renders 28 cells for 4 weeks of data", () => {
    const points = Array.from({ length: 28 }, (_, i) => ({
      day: `2026-05-${String(i + 1).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBe(28)
  })

  it("renders empty state for no data", () => {
    render(<ActivityHeatmap points={[]} />)
    expect(screen.getByText(/нет активности/i)).toBeInTheDocument()
  })

  it("varies opacity by count", () => {
    const points = [
      { day: "2026-05-01", count: 0 },
      { day: "2026-05-02", count: 100 },
    ]
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    const opacities = Array.from(cells).map((c) => parseFloat((c as HTMLElement).style.opacity || "1"))
    expect(opacities[0]).toBeLessThan(opacities[1])
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run activity-heatmap`
Expected: FAIL — module not exported.

- [ ] **Step 3: Implement ActivityHeatmap**

Create `frontend/src/components/analytics/activity-heatmap.tsx`:

```tsx
import { Calendar } from "lucide-react"
import { Card } from "@/components/ui/card"
import type { UsagePoint } from "@/api/analytics"

interface ActivityHeatmapProps {
  points: UsagePoint[]
}

// ActivityHeatmap — GitHub-style 4-week grid (28 cells, 7 cols × 4 rows).
// Opacity по count, нормализуется по max в наборе.
// Используется data.usage_per_day из usePersonalAnalytics(range="30d") или похожего —
// берёт последние 28 точек.
export function ActivityHeatmap({ points }: ActivityHeatmapProps) {
  if (points.length === 0) {
    return (
      <Card className="p-4">
        <div className="mb-2 flex items-center gap-2">
          <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Активность 4 недели</h3>
        </div>
        <p className="text-xs text-muted-foreground">Пока нет активности — создайте промпт</p>
      </Card>
    )
  }

  const slice = points.slice(-28)
  const max = Math.max(...slice.map((p) => p.count), 1)

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center gap-2">
        <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Активность 4 недели</h3>
      </div>
      <div className="grid grid-cols-7 gap-1.5">
        {slice.map((p) => {
          const opacity = p.count === 0 ? 0.08 : 0.2 + (p.count / max) * 0.8
          return (
            <span
              key={p.day}
              data-cell
              title={`${p.day}: ${p.count}`}
              className="aspect-square rounded-sm bg-violet-500"
              style={{ opacity }}
            />
          )
        })}
      </div>
    </Card>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run activity-heatmap`
Expected: PASS — 3/3.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/activity-heatmap.tsx \
        promptvault/frontend/src/components/analytics/activity-heatmap.test.tsx
git commit -m "feat(analytics): ActivityHeatmap GitHub-style 4w×7d grid"
```

---

### Task 8: ModelsDonut component (Recharts PieChart)

**Files:**
- Create: `frontend/src/components/analytics/models-donut.tsx`
- Create: `frontend/src/components/analytics/models-donut.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/models-donut.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ModelsDonut } from "./models-donut"
import type { ModelUsageRow } from "@/api/analytics"

afterEach(() => cleanup())

describe("ModelsDonut", () => {
  it("renders empty state for no data", () => {
    render(<ModelsDonut data={[]} />)
    expect(screen.getByText(/нет данных/i)).toBeInTheDocument()
  })

  it("renders legend with percentages", () => {
    const data: ModelUsageRow[] = [
      { model: "claude-3-opus", uses: 60 },
      { model: "gpt-4", uses: 30 },
      { model: "gemini", uses: 10 },
    ]
    render(<ModelsDonut data={data} />)
    expect(screen.getByText(/60%/)).toBeInTheDocument()
    expect(screen.getByText(/30%/)).toBeInTheDocument()
    expect(screen.getByText(/10%/)).toBeInTheDocument()
  })

  it("collapses tail beyond top-6 into «Другие»", () => {
    const data: ModelUsageRow[] = Array.from({ length: 8 }, (_, i) => ({
      model: `model-${i}`,
      uses: 10,
    }))
    render(<ModelsDonut data={data} />)
    expect(screen.getByText(/Другие/)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run models-donut`
Expected: FAIL — module not exported.

- [ ] **Step 3: Implement ModelsDonut**

Create `frontend/src/components/analytics/models-donut.tsx`:

```tsx
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts"
import { PieChart as PieIcon } from "lucide-react"
import { Card } from "@/components/ui/card"
import type { ModelUsageRow } from "@/api/analytics"
import { colorFor, labelFor, DEFAULT_COLOR } from "./model-colors"

interface ModelsDonutProps {
  data: ModelUsageRow[]
}

// ModelsDonut — donut chart для распределения по моделям.
// Top-6 + «Другие» хвост. Используем shared MODEL_COLORS palette.
// Recharts PieChart с innerRadius=60% для donut эффекта.
export function ModelsDonut({ data }: ModelsDonutProps) {
  const total = data.reduce((s, r) => s + r.uses, 0)

  if (total === 0) {
    return (
      <Card className="p-4">
        <div className="mb-2 flex items-center gap-2">
          <PieIcon className="size-[18px] text-muted-foreground" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Модели</h3>
        </div>
        <p className="text-xs text-muted-foreground">Пока нет данных</p>
      </Card>
    )
  }

  const sorted = [...data].sort((a, b) => b.uses - a.uses)
  const top = sorted.slice(0, 6)
  const tail = sorted.slice(6)
  const tailTotal = tail.reduce((s, r) => s + r.uses, 0)
  const display: ModelUsageRow[] =
    tailTotal > 0 ? [...top, { model: "__other__", uses: tailTotal }] : top

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center gap-2">
        <PieIcon className="size-[18px] text-muted-foreground" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Модели</h3>
      </div>
      <div className="flex items-center gap-3">
        <div className="size-[90px] shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={display.map((r) => ({ name: r.model, value: r.uses }))}
                dataKey="value"
                innerRadius="60%"
                outerRadius="100%"
                paddingAngle={2}
              >
                {display.map((r) => (
                  <Cell
                    key={r.model}
                    fill={r.model === "__other__" ? DEFAULT_COLOR : colorFor(r.model)}
                  />
                ))}
              </Pie>
            </PieChart>
          </ResponsiveContainer>
        </div>
        <ul className="flex-1 space-y-1 text-xs">
          {display.map((row) => {
            const pct = Math.round((row.uses / total) * 100)
            const color = row.model === "__other__" ? DEFAULT_COLOR : colorFor(row.model)
            const label = row.model === "__other__" ? "Другие" : labelFor(row.model)
            return (
              <li key={row.model} className="flex items-center gap-2">
                <span className="size-2 rounded-full" style={{ backgroundColor: color }} />
                <span className="flex-1 truncate">{label}</span>
                <span className="tabular-nums text-muted-foreground">{pct}%</span>
              </li>
            )
          })}
        </ul>
      </div>
    </Card>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run models-donut`
Expected: PASS — 3/3.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/models-donut.tsx \
        promptvault/frontend/src/components/analytics/models-donut.test.tsx
git commit -m "feat(analytics): ModelsDonut chart (Recharts PieChart) with top-6 + Others"
```

---

### Task 9: StreakTracker component

**Files:**
- Create: `frontend/src/components/analytics/streak-tracker.tsx`
- Create: `frontend/src/components/analytics/streak-tracker.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/streak-tracker.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { StreakTracker } from "./streak-tracker"

afterEach(() => cleanup())

describe("StreakTracker", () => {
  it("renders current and longest streak", () => {
    render(<StreakTracker current={5} longest={12} activeToday={true} />)
    expect(screen.getByText("5")).toBeInTheDocument()
    expect(screen.getByText(/best 12/i)).toBeInTheDocument()
  })

  it("renders 7 dots for last 7 days", () => {
    const { container } = render(<StreakTracker current={3} longest={10} activeToday={false} />)
    const dots = container.querySelectorAll("[data-streak-dot]")
    expect(dots.length).toBe(7)
  })

  it("highlights filled dots up to current streak", () => {
    const { container } = render(<StreakTracker current={4} longest={10} activeToday={true} />)
    const filled = container.querySelectorAll("[data-streak-dot][data-filled='true']")
    expect(filled.length).toBe(4)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run streak-tracker`
Expected: FAIL — module not exported.

- [ ] **Step 3: Implement StreakTracker**

Create `frontend/src/components/analytics/streak-tracker.tsx`:

```tsx
import { Flame } from "lucide-react"
import { Card } from "@/components/ui/card"

interface StreakTrackerProps {
  current: number
  longest: number
  activeToday: boolean
}

// StreakTracker — current streak counter + last-7-days dots.
// Заполненные dots = current streak (capped at 7 для отображения).
// Не путать с долгосрочным streak — это микро-визуализация для KPI-card.
export function StreakTracker({ current, longest, activeToday }: StreakTrackerProps) {
  const filledCount = Math.min(current, 7)
  return (
    <Card className="p-4">
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">Streak</span>
        <Flame className="size-4 text-amber-500" aria-hidden="true" />
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-bold tabular-nums">{current}</span>
        <span className="text-xs text-muted-foreground">{`best ${longest}`}</span>
      </div>
      <div className="mt-2 flex gap-1">
        {Array.from({ length: 7 }, (_, i) => {
          const filled = i < filledCount
          return (
            <span
              key={i}
              data-streak-dot
              data-filled={filled}
              className={
                filled
                  ? "h-3 w-3 rounded-sm bg-amber-500"
                  : "h-3 w-3 rounded-sm bg-foreground/10"
              }
              aria-hidden="true"
            />
          )
        })}
      </div>
      {!activeToday && current > 0 && (
        <p className="mt-1 text-[11px] text-muted-foreground">Сегодня ещё нет активности</p>
      )}
    </Card>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run streak-tracker`
Expected: PASS — 3/3.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/streak-tracker.tsx \
        promptvault/frontend/src/components/analytics/streak-tracker.test.tsx
git commit -m "feat(analytics): StreakTracker — counter + 7-day dots"
```

---

### Task 10: CompactQuotas component

**Files:**
- Create: `frontend/src/components/analytics/compact-quotas.tsx`
- Create: `frontend/src/components/analytics/compact-quotas.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/analytics/compact-quotas.test.tsx`:

```tsx
import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { CompactQuotas } from "./compact-quotas"
import type { UsageSummary } from "@/api/analytics"

afterEach(() => cleanup())

describe("CompactQuotas", () => {
  const baseQuota: UsageSummary = {
    plan_id: "pro",
    prompts: { used: 230, limit: 500 },
    collections: { used: 30, limit: 100 },
    teams: { used: 1, limit: 5 },
    ext_uses_today: { used: 0, limit: 50 },
    mcp_uses_today: { used: 0, limit: 50 },
  }

  it("renders prompts/collections/mcp usage", () => {
    render(<CompactQuotas quotas={baseQuota} />)
    expect(screen.getByText(/230/)).toBeInTheDocument()
    expect(screen.getByText(/500/)).toBeInTheDocument()
    expect(screen.getByText(/30/)).toBeInTheDocument()
    expect(screen.getByText(/100/)).toBeInTheDocument()
  })

  it("renders nothing when quotas is undefined", () => {
    const { container } = render(<CompactQuotas quotas={undefined} />)
    expect(container.firstChild).toBeNull()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run compact-quotas`
Expected: FAIL — module not exported.

- [ ] **Step 3: Implement CompactQuotas**

Create `frontend/src/components/analytics/compact-quotas.tsx`:

```tsx
import { Card } from "@/components/ui/card"
import type { UsageSummary } from "@/api/analytics"

interface CompactQuotasProps {
  quotas: UsageSummary | undefined
}

// CompactQuotas — однострочный footer с тремя ключевыми quota:
// Промпты, Коллекции, MCP-вызовы сегодня.
// Заменяет 3 больших QuotaProgress блока, не занимает prime real-estate.
export function CompactQuotas({ quotas }: CompactQuotasProps) {
  if (!quotas) return null

  const items = [
    { label: "Промпты", used: quotas.prompts.used, limit: quotas.prompts.limit },
    { label: "Коллекции", used: quotas.collections.used, limit: quotas.collections.limit },
    { label: "MCP сегодня", used: quotas.mcp_uses_today.used, limit: quotas.mcp_uses_today.limit },
  ]

  return (
    <Card className="flex items-center gap-6 px-4 py-3">
      <span className="text-[11px] uppercase tracking-wide text-muted-foreground">Лимиты:</span>
      {items.map((item) => {
        const pct = item.limit > 0 ? (item.used / item.limit) * 100 : 0
        const isHigh = pct >= 90
        const isMid = pct >= 75 && pct < 90
        const barColor = isHigh ? "bg-rose-500" : isMid ? "bg-amber-500" : "bg-violet-500"
        return (
          <div key={item.label} className="flex flex-1 items-center gap-2">
            <span className="text-xs text-muted-foreground">{item.label}</span>
            <div className="flex-1">
              <div className="h-1.5 overflow-hidden rounded-full bg-foreground/10">
                <div
                  className={`h-full ${barColor}`}
                  style={{ width: `${Math.min(pct, 100)}%` }}
                />
              </div>
            </div>
            <span className="font-mono text-xs tabular-nums">
              {item.used.toLocaleString("ru")}/{item.limit.toLocaleString("ru")}
            </span>
          </div>
        )
      })}
    </Card>
  )
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `npx vitest run compact-quotas`
Expected: PASS — 2/2.

- [ ] **Step 5: Commit**

```bash
git add promptvault/frontend/src/components/analytics/compact-quotas.tsx \
        promptvault/frontend/src/components/analytics/compact-quotas.test.tsx
git commit -m "feat(analytics): CompactQuotas footer — 3 inline progress bars"
```

---

## Phase 3: Integration

### Task 11: Refactor insights-panel.tsx to use InsightActionCard

**Files:**
- Modify: `frontend/src/components/analytics/insights-panel.tsx`

- [ ] **Step 1: Read existing insights-panel.tsx**

Find current rendering logic. Note current mapping `Insight.type` → UI label / icon.

- [ ] **Step 2: Replace render with InsightActionCard map**

Replace the JSX that renders insights with:

```tsx
import { AlertCircle, Copy, TrendingUp, TrendingDown, Archive, Hash, FolderOpen, type LucideIcon } from "lucide-react"
import { InsightActionCard, type InsightTone } from "./insight-action-card"
import type { Insight } from "@/api/types"

const INSIGHT_META: Record<
  Insight["type"],
  { icon: LucideIcon; tone: InsightTone; title: string; href: string; descBuilder: (n: number) => string; ctaLabel: string }
> = {
  unused_prompts: {
    icon: AlertCircle, tone: "warning", title: "Забытые",
    href: "/prompts?filter=unused",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт не использовался" : "промптов не использовались"} 30+ дней`,
    ctaLabel: "Посмотреть",
  },
  possible_duplicates: {
    icon: Copy, tone: "info", title: "Дубликаты",
    href: "/prompts?filter=duplicates",
    descBuilder: (n) => `${n} ${n === 1 ? "пара" : "пары"} похожих промптов`,
    ctaLabel: "Объединить",
  },
  trending: {
    icon: TrendingUp, tone: "success", title: "Растут",
    href: "/prompts?filter=trending",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} растут в использовании`,
    ctaLabel: "Открыть",
  },
  declining: {
    icon: TrendingDown, tone: "warning", title: "Падают",
    href: "/prompts?filter=declining",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} используются всё реже`,
    ctaLabel: "Посмотреть",
  },
  most_edited: {
    icon: Archive, tone: "info", title: "Часто правят",
    href: "/prompts?sort=most-edited",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} с большим числом версий`,
    ctaLabel: "Открыть",
  },
  orphan_tags: {
    icon: Hash, tone: "warning", title: "Orphan-теги",
    href: "/tags",
    descBuilder: (n) => `${n} ${n === 1 ? "тег" : "тегов"} без промптов`,
    ctaLabel: "Очистить",
  },
  empty_collections: {
    icon: FolderOpen, tone: "warning", title: "Пустые коллекции",
    href: "/collections",
    descBuilder: (n) => `${n} ${n === 1 ? "коллекция" : "коллекций"} без промптов`,
    ctaLabel: "Очистить",
  },
}

interface InsightsPanelProps {
  insights: Insight[]
}

export function InsightsPanel({ insights }: InsightsPanelProps) {
  if (insights.length === 0) {
    return <p className="text-sm text-muted-foreground">Пока нет инсайтов. Возвращайтесь завтра.</p>
  }
  return (
    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
      {insights.map((ins) => {
        const meta = INSIGHT_META[ins.type]
        if (!meta) return null
        const count = Array.isArray(ins.payload) ? ins.payload.length : 0
        return (
          <InsightActionCard
            key={ins.type}
            tone={meta.tone}
            icon={meta.icon}
            title={meta.title}
            description={meta.descBuilder(count)}
            href={meta.href}
            ctaLabel={meta.ctaLabel}
            count={count}
          />
        )
      })}
    </div>
  )
}
```

- [ ] **Step 3: Verify existing test still passes**

Run: `npx vitest run insights-panel analytics-insights-states`
Expected: PASS. If `analytics-insights-states.test.tsx` selectors broke, mark for fix in Task 14.

- [ ] **Step 4: Commit**

```bash
git add promptvault/frontend/src/components/analytics/insights-panel.tsx
git commit -m "refactor(analytics): InsightsPanel uses InsightActionCard per type"
```

---

### Task 12: Restyle insights-locked-card.tsx

**Files:**
- Modify: `frontend/src/components/analytics/insights-locked-card.tsx`

- [ ] **Step 1: Update component to match InsightActionCard layout**

Replace the existing component body:

```tsx
import { Lock, ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

interface InsightsLockedCardProps {
  title: string
  description: string
}

// InsightsLockedCard — Pro teaser locked card (для Max-only insight types).
// Визуально соответствует InsightActionCard, но dashed border и lock icon.
// CTA — ссылка на /pricing.
export function InsightsLockedCard({ title, description }: InsightsLockedCardProps) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-foreground/2 p-4">
      <div className="mb-1.5 flex items-center gap-2">
        <Lock className="size-4 text-muted-foreground" aria-hidden="true" />
        <span className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
          {title}
        </span>
      </div>
      <p className="mb-2 text-sm text-foreground/70">{description}</p>
      <Link
        to="/pricing"
        className="inline-flex items-center gap-1 text-xs font-medium text-violet-600 dark:text-violet-400"
      >
        Доступно в Max
        <ArrowRight className="size-3" aria-hidden="true" />
      </Link>
    </div>
  )
}
```

- [ ] **Step 2: Verify existing analytics-insights-states test still passes**

Run: `npx vitest run analytics-insights-states`

Expected: PASS (test ищет text `Доступно в Max →` или `Доступно в Max`, новая разметка содержит этот текст).

If test fails — adjust either component text или test assertions; record fix for Task 14.

- [ ] **Step 3: Commit**

```bash
git add promptvault/frontend/src/components/analytics/insights-locked-card.tsx
git commit -m "refactor(analytics): InsightsLockedCard restyle to match InsightActionCard"
```

---

### Task 13: Restructure analytics.tsx (Bento Grid)

**Files:**
- Modify: `frontend/src/pages/analytics.tsx`

- [ ] **Step 1: Backup intent before rewrite**

The existing page structure is preserved logically (3-state Free/Pro/Max, range picker, CSV export, isPaid/isMax gating). Only JSX body changes.

- [ ] **Step 2: Rewrite analytics.tsx**

Replace entire file contents:

```tsx
import { useState, useMemo } from "react"
import { Loader2, Download, Activity, FileText, Eye, Trophy } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthStore } from "@/stores/auth-store"
import { usePersonalAnalytics, useInsights } from "@/hooks/use-analytics"
import { useStreak } from "@/hooks/use-streaks"
import { computeDelta, downloadAnalyticsCSV, type AnalyticsRange } from "@/api/analytics"
import { UsageChart } from "@/components/analytics/usage-chart"
import { TopPromptsTable } from "@/components/analytics/top-prompts-table"
import { RangePicker } from "@/components/analytics/range-picker"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { InsightsPanel } from "@/components/analytics/insights-panel"
import { InsightsLockedCard } from "@/components/analytics/insights-locked-card"
import { NarrativeBanner } from "@/components/analytics/narrative-banner"
import { KpiCard } from "@/components/analytics/kpi-card"
import { ActivityHeatmap } from "@/components/analytics/activity-heatmap"
import { ModelsDonut } from "@/components/analytics/models-donut"
import { StreakTracker } from "@/components/analytics/streak-tracker"
import { CompactQuotas } from "@/components/analytics/compact-quotas"
import { buildNarrative } from "@/lib/analytics-narrative"
import { toast } from "sonner"

// Phase 14 C.2 + analytics redesign 2026-05-17: /analytics — личный dashboard
// в формате Bento Grid. Three-state Pro Insights teaser сохранён:
//  - Free: 7-дневное окно, UpgradeGate Pro, без CSV, без Smart Insights
//  - Pro: до 90 дней, CSV export, 2 insight types (unused + duplicates)
//  - Max: до 365 дней, CSV, все 7 insight types
export default function AnalyticsPage() {
  const user = useAuthStore((s) => s.user)
  const planId = user?.plan_id ?? "free"
  const isMax = planId.startsWith("max")
  const isPaid = planId.startsWith("pro") || isMax

  const [range, setRange] = useState<AnalyticsRange>("7d")

  const { data, isLoading, isError } = usePersonalAnalytics(range)
  const insightsQuery = useInsights(isPaid)
  const streakQuery = useStreak()

  const usageSparkline = useMemo(
    () => data?.usage_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const createdSparkline = useMemo(
    () => data?.prompts_created_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const sharedSparkline = useMemo(
    () => data?.share_views_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const narrative = useMemo(
    () => (data ? buildNarrative(data, insightsQuery.data?.items ?? null) : null),
    [data, insightsQuery.data],
  )
  const streakSegment = streakQuery.data
    ? `streak ${streakQuery.data.current_streak} ${pluralStreak(streakQuery.data.current_streak)}`
    : null
  const narrativeFinal = narrative ? { ...narrative, streak: streakSegment } : null

  async function handleExport() {
    try {
      await downloadAnalyticsCSV("personal", range)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Не удалось скачать CSV")
    }
  }

  if (isError) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="mb-4 text-2xl font-bold">Аналитика</h1>
        <p className="text-destructive">Не удалось загрузить данные. Попробуйте обновить страницу.</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto space-y-4 px-4 py-8">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Аналитика</h1>
          <p className="text-sm text-muted-foreground">
            Ваше использование промптов и публичных ссылок
          </p>
        </div>
        <div className="flex items-center gap-2">
          <RangePicker value={range} onChange={setRange} planId={planId} />
          {isPaid && (
            <Button variant="outline" size="sm" onClick={handleExport}>
              <Download className="size-4" />
              CSV
            </Button>
          )}
        </div>
      </div>

      {isLoading || !data ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      ) : (
        <>
          {/* AI Narrative Banner */}
          {narrativeFinal && <NarrativeBanner segments={narrativeFinal} />}

          {/* KPI Strip — 4 cards с sparklines */}
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <KpiCard
              label="Использования"
              value={data.totals_current.uses.toLocaleString("ru")}
              delta={computeDelta(data.totals_current.uses, data.totals_previous.uses)}
              sparkline={usageSparkline}
              icon={Activity}
            />
            <KpiCard
              label="Новых промптов"
              value={data.totals_current.created.toLocaleString("ru")}
              delta={computeDelta(data.totals_current.created, data.totals_previous.created)}
              sparkline={createdSparkline}
              icon={FileText}
            />
            <KpiCard
              label="Просмотров ссылок"
              value={data.totals_current.share_views.toLocaleString("ru")}
              delta={
                isPaid
                  ? computeDelta(data.totals_current.share_views, data.totals_previous.share_views)
                  : null
              }
              sparkline={isPaid ? sharedSparkline : undefined}
              icon={Eye}
            />
            {streakQuery.data ? (
              <StreakTracker
                current={streakQuery.data.current_streak}
                longest={streakQuery.data.longest_streak}
                activeToday={streakQuery.data.active_today}
              />
            ) : (
              <KpiCard
                label="Топ-промпт"
                value={data.top_prompts[0]?.uses?.toLocaleString("ru") ?? "—"}
                icon={Trophy}
              />
            )}
          </div>

          {/* Smart Insights three-state */}
          {!isPaid && (
            <UpgradeGate
              title="Подсказки — на тарифе Pro"
              description="Забытые промпты и дубликаты помогут навести порядок. Полный набор — в Max."
              targetPlan="Pro"
            />
          )}

          {isPaid && insightsQuery.isLoading && (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {isPaid && insightsQuery.data && (
            <section className="space-y-3">
              <h2 className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                Стоит сделать сегодня
              </h2>
              <InsightsPanel insights={insightsQuery.data.items} />
              {!isMax && (
                <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                  <InsightsLockedCard title="Растёт" description="Промпты, использование которых выросло за 7 дней." />
                  <InsightsLockedCard title="Падает" description="Промпты, которые перестали активно использоваться." />
                  <InsightsLockedCard title="Часто правят" description="Топ промптов по количеству версий." />
                  <InsightsLockedCard title="Orphan-теги" description="Теги без промптов для уборки." />
                  <InsightsLockedCard title="Пустые коллекции" description="Коллекции без промптов." />
                </div>
              )}
            </section>
          )}

          {/* Bento Grid main charts */}
          <div className="grid gap-3 lg:grid-cols-6 lg:auto-rows-[90px]">
            <div className="lg:col-span-4 lg:row-span-3">
              <UsageChart title="Использование по дням" data={data.usage_per_day} />
            </div>
            <div className="lg:col-span-2 lg:row-span-3">
              <ActivityHeatmap points={data.usage_per_day} />
            </div>
            <div className="lg:col-span-2 lg:row-span-2">
              <ModelsDonut data={data.usage_by_model} />
            </div>
            <div className="lg:col-span-4 lg:row-span-2">
              <TopPromptsTable title="Топ-10 промптов" prompts={data.top_prompts} />
            </div>
          </div>

          {/* Compact Quotas */}
          <CompactQuotas quotas={data.quotas} />

          {/* Upgrade gate для Free — расширенная история */}
          {!isPaid && (
            <UpgradeGate
              title="Больше истории на Pro"
              description="До 90 дней на Pro, до 365 на Max. Плюс экспорт CSV и подробные метрики."
              targetPlan="Pro"
            />
          )}
        </>
      )}
    </div>
  )
}

function pluralStreak(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "день"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return "дня"
  return "дней"
}
```

- [ ] **Step 3: Build + run all analytics tests**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
npm run lint
npx tsc --noEmit
npx vitest run
```

Expected: ESLint clean, tsc clean, all tests pass (or fail on `analytics-insights-states` — Task 14).

- [ ] **Step 4: Commit**

```bash
git add promptvault/frontend/src/pages/analytics.tsx
git commit -m "feat(analytics): Bento Grid layout with NarrativeBanner + KpiCard + StreakTracker"
```

---

### Task 14: Update analytics-insights-states.test.tsx

**Files:**
- Modify: `frontend/src/pages/__tests__/analytics-insights-states.test.tsx`

- [ ] **Step 1: Update mocks for new hooks**

Add `useStreak` mock to existing mock setup. The test currently mocks `useInsights`/`usePersonalAnalytics`/`useRefreshInsights`. Add `useStreak`:

```tsx
vi.mock("@/hooks/use-streaks", () => ({
  useStreak: () => ({ data: undefined, isLoading: false, isError: false }),
}))
```

- [ ] **Step 2: Update assertions if selectors broke**

Verify each of three test cases (Free/Pro/Max) still passes. Common breakages:

- Free → `UpgradeGate Pro`: text `Подсказки — на тарифе Pro` is preserved in new analytics.tsx. Should still pass.
- Pro → 5 locked cards: text `Доступно в Max` is in `InsightsLockedCard` (Task 12). Should still pass.
- Max → no locked cards: assertion `queryByText("Доступно в Max →")` should still work — но новый layout рендерит `Доступно в Max` без `→` стрелки в тексте (стрелка как icon).

If Max test fails, update selector:

```tsx
// было:
expect(screen.queryByText("Доступно в Max →")).not.toBeInTheDocument()
// стало:
expect(screen.queryByText(/Доступно в Max/i)).not.toBeInTheDocument()
```

- [ ] **Step 3: Run test**

Run: `npx vitest run analytics-insights-states`

Expected: PASS — 3/3.

- [ ] **Step 4: Commit**

```bash
git add promptvault/frontend/src/pages/__tests__/analytics-insights-states.test.tsx
git commit -m "test(analytics): update three-state UI tests for Bento layout"
```

---

### Task 15: Manual browser smoke + screenshots

**Files:**
- No code changes. Run docker dev stack + screenshot 3 plans.

- [ ] **Step 1: Start stack**

Already running (`docker compose -f promptvault/docker-compose.dev.yml ps` should show postgres/api/frontend). If not:

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault
docker compose -f docker-compose.dev.yml up -d --build
```

- [ ] **Step 2: Login as each test user**

Login in browser at http://localhost:5173 as:

- `e2e-free@test.local` / `TestPass2026!`
- `e2e-pro@test.local` / `TestPass2026!`
- `e2e-max@test.local` / `TestPass2026!`

- [ ] **Step 3: Navigate to /analytics and take screenshots**

For each plan, open http://localhost:5173/analytics, verify visually:

- **Free:** UpgradeGate Pro карточка вместо Smart Insights. KPI strip с 4 cards. No CSV export button. Range picker — только 7d available.
- **Pro:** NarrativeBanner вверху. KPI strip + sparklines. Smart Insights ribbon с 2 InsightActionCard (unused + duplicates) + 5 locked. UsageChart 4-col + ActivityHeatmap 2-col side by side. ModelsDonut + TopPromptsTable. CompactQuotas внизу. CSV button visible.
- **Max:** All 7 InsightActionCard, no locked. Все остальные блоки.

Screenshots сохранить:

```
promptvault/docs/superpowers/screenshots/analytics-redesign-free.png
promptvault/docs/superpowers/screenshots/analytics-redesign-pro.png
promptvault/docs/superpowers/screenshots/analytics-redesign-max.png
```

- [ ] **Step 4: Commit screenshots**

```bash
mkdir -p promptvault/docs/superpowers/screenshots
git add promptvault/docs/superpowers/screenshots/analytics-redesign-*.png
git commit -m "docs(analytics): smoke screenshots for redesigned Bento layout"
```

---

### Task 16: Cleanup deprecated model-segmentation-chart (optional)

**Files:**
- Remove or deprecate `frontend/src/components/analytics/model-segmentation-chart.tsx`

- [ ] **Step 1: grep for consumers**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
grep -rn "ModelSegmentationChart\|model-segmentation-chart" src/
```

- [ ] **Step 2: Decision**

- Если consumers = 0 (после изменений в `analytics.tsx`) — удалить файл.
- Если есть consumers (например, в `/teams/:id/analytics`) — оставить, file uses shared `model-colors.ts` (Task 1).

- [ ] **Step 3: If removing, delete files + test**

```bash
rm promptvault/frontend/src/components/analytics/model-segmentation-chart.tsx
rm promptvault/frontend/src/components/analytics/model-segmentation-chart.test.tsx
```

Verify build:

```bash
npx tsc --noEmit
npx vitest run
```

- [ ] **Step 4: Commit (if removed)**

```bash
git add -A promptvault/frontend/src/components/analytics/
git commit -m "chore(analytics): remove deprecated model-segmentation-chart (replaced by ModelsDonut)"
```

---

## Final verification (before merge/PR)

After Task 16:

- [ ] **Final 1: Full test suite**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
npm run lint
npx tsc --noEmit
npx vitest run
npm run build
```

Expected: всё зелёное, build clean.

- [ ] **Final 2: Visual regression**

Manually scroll through http://localhost:5173/analytics, проверь responsive: desktop (1920px), laptop (1280px), tablet (768px), mobile (375px). На mobile карточки должны стекаться вертикально (через Tailwind `sm:` / `lg:` breakpoints).

- [ ] **Final 3: Bundle size delta**

Before/after `npm run build` сравнить `dist/assets/*.js` sizes. Допустимо +25 KB gzipped.

```bash
du -sh dist/assets/*.js | sort -rh | head -5
```

---

## Self-Review

### Spec coverage

| Spec section | Task |
|---|---|
| §1 NarrativeBanner | Tasks 3, 4 |
| §1 KPI sparklines | Tasks 2, 5 |
| §1 InsightActionCard tones | Task 6 |
| §1 ActivityHeatmap | Task 7 |
| §1 ModelsDonut | Task 8 |
| §1 StreakTracker | Task 9 |
| §1 CompactQuotas | Task 10 |
| §2 Решение 1 Sparkline | Task 2 |
| §2 Решение 2 Bento Grid | Task 13 |
| §2 Решение 3 narrative template | Tasks 3, 4 |
| §3 Файлы создаём | Tasks 1-10 |
| §3 Файлы меняем | Tasks 11-14 |
| §6 Никаких новых deps | All tasks (только lucide/recharts/shadcn) |
| §7 Unit-тесты | Tasks 1-10 — каждый с test |
| §7 Integration tests | Task 14 |
| §8 Observability N/A | acknowledged |
| §10 No feature flag | acknowledged |
| §12 Mobile fallback | Task 15 Final 2 |
| §13 Bundle size delta | Task 15 Final 3 |

**Gaps:** Нет. Все требования спеки покрыты.

### Placeholder scan

- Нет "TBD", "TODO", "implement later".
- Все code-снипеты — complete и используются.
- Type-references везде согласованы (`Insight.type`, `PersonalDashboard`, `useStreak`).

### Type consistency

- `Insight.type` (discriminator) — везде совпадает (Task 3, 11).
- `PersonalDashboard` (not `PersonalAnalyticsResponse`) — Task 3, 13.
- `useStreak` (singular, не `useStreaks`) — Tasks 13, 14.
- `KpiCard` props (`label`, `value`, `delta`, `sparkline`, `icon`) — same в Tasks 5 и 13.
- `InsightActionCard` props (`tone`, `icon`, `title`, `description`, `href`, `ctaLabel`, `count`) — same в Tasks 6 и 11.
- `InsightTone` — type определён в Task 6, импортируется в Task 11.
- `MODEL_COLORS` / `colorFor` / `labelFor` — экспортируется в Task 1, используется в Task 8.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-17-analytics-redesign.md`. Two execution options:

**1. Subagent-Driven (recommended)** — Fresh subagent per task + two-stage review между задачами. Подходит для длинных планов, защищает основной контекст.

**2. Inline Execution** — Batch execution в этой сессии через executing-plans skill, checkpoints для review.

Which approach?
