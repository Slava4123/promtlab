# Analytics Page Redesign — Design Doc

**Дата:** 2026-05-17
**Owner:** Slava Kovalchuk
**Branch:** `feat/pricing-iteration-v3` (поверх pricing-итерации)
**Статус:** утверждён, готов к implementation plan

---

## Контекст

**Фича.** Полный визуальный redesign страницы `/analytics` (личный dashboard) под современные state-of-the-art-паттерны (Linear, Vercel/Tremor, Notion AI, Stripe). Backend hooks `usePersonalAnalytics` и `useInsights` остаются как есть; меняется только presentation layer.

**Для кого.** Юзеры всех тарифов (Free/Pro/Max), которые открывают `/analytics` чтобы понять «как я использую промпты». Текущая страница — линейный список карточек без визуальной иерархии. Цель — повысить engagement и читаемость основных метрик за счёт правильной типографики, иконок, sparklines, actionable Smart Insights и асимметричной Bento Grid.

**Зачем.**

- **Eye-tracking F-pattern.** В текущей версии главные метрики тонут в общем потоке. Hero-зона top-left (80-120px) сейчас отдана под header без actionable-данных.
- **Smart Insights в самом низу страницы** — actionable-контент после квот, под графиками. Должен быть в hero-зоне.
- **Эмодзи / отсутствие иконок.** Сейчас иконок практически нет, есть text-only metric titles. Lucide-иконки уже в `package.json`, не используются полно.
- **Hot insights без визуала.** Модели = горизонтальная полоса (не donut). Activity / streak / time-of-day patterns нигде не визуализированы.

**Жёсткие ограничения.**

- Frontend-only. Backend `usePersonalAnalytics`, `useInsights`, `useStreaks` оставляем как есть. Любой новый data-shape (time-heatmap, comparison overlay) — либо отложен в follow-up, либо реализуется на frontend из существующего ответа.
- Не ломать существующие unit-tests аналитики (`metric-card.test.tsx`, `usage-chart.test.tsx`, `quota-progress.test.tsx`, `model-segmentation-chart.test.tsx`).
- Three-state Pro Insights teaser из Pricing v3 (Free → `UpgradeGate`, Pro → 2 типа + 5 locked-карточек, Max → 7 типов) сохраняем точь-в-точь — только применяем новый визуальный стиль actionable-card.
- shadcn/ui + Tailwind + Recharts — никаких новых UI-libs. Не добавляем Tremor (Vercel-acquired), потому что у нас уже свой shadcn-stack; копируем паттерны из Tremor blocks вручную.
- Lucide React — единственная icon library. Никаких эмодзи в production UI.

**Заинтересованные стороны.**

- Исполнитель (Slava — сам имплементит).
- PR-ревьюер (он же).
- Аудитория страницы — все юзеры PromtLab; design должен работать на Free/Pro/Max с разными объёмами данных.

**Аудит ясности.** Open questions из brainstorming зафиксированы defaults:

1. **Time-heatmap (часы×дни)** — **отложен в follow-up**. Не делаем в этой итерации, требует нового backend-endpoint (`/api/analytics/usage-by-hour-dow`). Слот в Bento Grid используется компактнее (расширяет ShareFunnel или GroupedTopPrompts).
2. **Comparison overlay на UsageChart** — **в v1 не делаем**. Backend сейчас отдаёт `totals_previous` числом, но не массив по дням. В v1 рисуем одну линию + дельту в KPI-card; overlay на графике — follow-up после backend-расширения.
3. **Share-funnel conversion (views → sign-ups)** — **в v1 показываем только views/day** (текущие данные `share_views_per_day`). Реальная funnel-конверсия (sign-ups от share-link visitor) требует tracking referrer в auth-store и нового backend-endpoint — отдельная фича.

**Скоуп. Что вне scope:**

- Time-heatmap, comparison-overlay, share-funnel-conversion (см. выше).
- Team analytics (`/teams/:id/analytics`) — другая страница, отдельный redesign.
- Mobile layout — этот redesign приоритезирует desktop; mobile должен оставаться функциональным (≥ 360px), но без полной адаптации Bento Grid.
- Иллюстрации в empty-states — оставляем текстовый placeholder, иллюстрации в follow-up.

---

## Карта существующего кода

**Слои.** Frontend: `frontend/src/pages/analytics.tsx` собирает компоненты из `frontend/src/components/analytics/*`. Hook `usePersonalAnalytics(range)` (`frontend/src/hooks/use-analytics.ts`) возвращает `PersonalAnalyticsResponse` со всеми данными. Hook `useInsights(isPaid)` возвращает `{items: SmartInsight[]}`.

**Эталонные файлы.**

- `frontend/src/pages/analytics.tsx:31-235` — текущая страница, vertical-stack из 7 секций. Файл сохраняем, переписываем тело JSX полностью.
- `frontend/src/components/analytics/metric-card.tsx:13-30` — текущий `MetricCard` с title/value/subtitle/delta. Расширяем под sparkline.
- `frontend/src/components/analytics/model-segmentation-chart.tsx:44-131` — горизонтальная полоса. Полностью заменяется на donut. `MODEL_COLORS` palette переиспользуем.
- `frontend/src/components/analytics/usage-chart.tsx:14-63` — Recharts AreaChart. Оставляем, переиспользуем в Bento Grid.
- `frontend/src/components/analytics/insights-panel.tsx` — текущая отрисовка Smart Insights items. Возможно заменим на `insight-action-card.tsx` (см. §3).
- `frontend/src/components/analytics/insights-locked-card.tsx` — Pro teaser locked-карточка. Применяем тот же визуальный стиль actionable-card.
- `frontend/src/components/ui/card.tsx` — shadcn/ui Card. Используем везде.

**Тесты-эталоны.**

- `frontend/src/components/analytics/metric-card.test.tsx` — пример unit-теста на render-логику с delta. Копируем структуру для `kpi-card.test.tsx`.
- `frontend/src/components/analytics/model-segmentation-chart.test.tsx` — empty state + segments. Копируем для `models-donut.test.tsx`.
- `frontend/src/pages/__tests__/analytics-insights-states.test.tsx` (только что создан в Pricing v3 Task 10) — three-state UI Free/Pro/Max. Сохраняем, обновляем под новые компоненты.

**Свежий git log по `frontend/src/components/analytics/`.** Последние коммиты — `99dde8f three-state insights UI` (Pricing v3 Task 10), `eea9b85 team-scope Smart Insights`, `2cad5ad Phase 14.3 — тесты + observability`. Активный refactor — только Pricing v3, мы поверх него.

**Конвенции.** shadcn/ui Card/Skeleton, TanStack Query для всех API-вызовов, Vitest + Testing Library для тестов. Lucide React для иконок. Tailwind utility-классы (нет CSS modules). `tabular-nums` для чисел. Цвета через CSS-переменные (`var(--muted)`, `var(--accent)`, etc.) — наши brand-токены в `index.css`.

---

## 1. Резюме

Полный визуальный redesign `/analytics` без backend-изменений. Новая структура: AI Narrative Banner → 4 KPI-карточки с sparklines → Actionable Smart Insights ribbon → асимметричная Bento Grid из 6 чартов → Compact Quotas footer. Все эмодзи заменены на Lucide-иконки (Sparkles, Activity, FileText, Eye, Flame, AlertCircle, Copy, TrendingUp, LineChart, Calendar, PieChart, Share2, Trophy) с единым stroke-weight 1.75 и 2 строгими размерами (16px / 18px). Сохраняется three-state Pro Insights teaser (Free / Pro / Max) из Pricing v3.

**Ключевые технические решения.**

1. **KPI-карточка с inline-sparkline** — расширение `metric-card.tsx` через композицию: добавляется проп `sparkline?: number[]`, рендерится через новый компонент `Sparkline` (custom SVG ~40 строк, без Recharts overhead для мини-графиков).
2. **Bento Grid через CSS Grid + `grid-column: span N` + `grid-row: span N`** — без grid-libraries. Адаптивный breakpoint: на ≥ `lg` (1024px) — 6-col grid; на `md` — 2-col fallback; на `sm` — vertical-stack.
3. **AI Narrative Banner — template-based** на frontend, без LLM-вызова. Builder-функция собирает текст из `data.totals_current` / `totals_previous` / `top_prompts` / `usage_by_model` / `insights`. Single point of formatting в `narrative-banner.tsx`.
4. **Donut вместо горизонтальной полосы для моделей** — Recharts PieChart с innerRadius. Сохраняем `MODEL_COLORS` palette из существующего файла.

**Аудитория плана.** Исполнитель + PR-ревьюер. Глубина — implementation-ready: для каждого нового компонента указаны props, размеры, тесты.

---

## 2. Архитектурные решения

### Решение 1: Sparkline в KPI-карточке — custom SVG vs Recharts vs Tremor

- **Решение:** Custom SVG-компонент `<Sparkline points={[1,3,2,5,8,7,12]} trend="up" />`, ~40 строк. Polyline по нормализованным точкам, gradient-fill снизу, цвет по trend. Без Recharts overhead.
- **Альтернативы:**
  - (A) Recharts AreaChart с минимальным config (без осей, без tooltip, height=22px). Минус: ~80 KB на инстанс, медленный рендер для 4 sparklines одновременно.
  - (B) Tremor `SparkAreaChart`. Минус: добавление новой UI-библиотеки целиком ради одного компонента — overkill.
  - (C) Custom SVG **(выбран)** — нулевой bundle-cost, контроль над gradient, легко тестируется.
- **Trade-offs.**
  - ✅ Минимальный bundle impact.
  - ✅ Гибкость в цветах (можно ставить amber/emerald/rose без перенастройки Recharts theme).
  - ❌ Если в будущем понадобятся tooltips на sparkline — придётся переписать на Recharts. Acceptable: KPI-карточка с tooltip = full chart, не sparkline.
- **Источник.** Vercel KPI cards (vercel.com/dashboard), Stripe KPI (stripe.com/dashboard) — оба используют custom SVG polyline без библиотек.

### Решение 2: Bento Grid — pure CSS Grid vs grid-library

- **Решение:** Pure Tailwind CSS Grid: `grid-cols-6` + `grid-auto-rows: 90px` + `grid-column: span N; grid-row: span N` через утилиты `col-span-4 row-span-2`. Responsive через breakpoints.
- **Альтернативы:**
  - (A) react-grid-layout / react-mosaic — overkill, нужна draggable-функциональность которая нам не нужна.
  - (B) CSS Subgrid — нативно поддерживается всеми браузерами (2024+), но избыточно для нашего случая (выравнивание children).
  - (C) Pure CSS Grid + span **(выбран)** — простой, читаемый, никаких deps.
- **Trade-offs.**
  - ✅ Нет новых deps.
  - ✅ Полностью совместимо с Tailwind + shadcn/ui.
  - ❌ Адаптация под mobile требует ручного breakpoint'а: на `md` все карточки → `col-span-1` (vertical-stack). Прописываем явно.
- **Источник.** [Tremor Blocks dashboard examples](https://blocks.tremor.so/) — Bento patterns без library, на CSS Grid.

### Решение 3: Narrative Banner — frontend template vs backend AI-call

- **Решение:** Frontend template-функция `buildNarrative(data, insights)` собирает строку из существующего ответа. Без LLM-вызовов (принцип «без AI на нашей стороне» из CLAUDE.md). Сегменты: dynamic (period + delta), top_model (если есть), streak (если есть), action-hint (если insights не пусты).
- **Альтернативы:**
  - (A) Backend endpoint `/api/analytics/narrative` который вызывает GigaChat/YandexGPT для summary. Минус: ломает принцип проекта; +cost; latency.
  - (B) Hardcoded text всегда «Ваша аналитика за период». Минус: не использует возможность дать insight.
  - (C) Template на frontend **(выбран)** — детерминистично, нулевой cost, легко тестируется (table-driven по входным `data`).
- **Trade-offs.**
  - ✅ Соответствует принципу «без AI».
  - ✅ Тестируется через `expect(buildNarrative(mockData)).toContain('+23%')`.
  - ❌ Менее «живой» чем LLM, но юзеры не ждут narrative-essay — формат «За 7 дней: 234 использования +23%» им подходит.
- **Источник.** Linear dashboard «You shipped X issues this week» — template-based summary без LLM.

---

## 3. Изменения в коде

### Создаём

- `frontend/src/components/analytics/sparkline.tsx` — custom SVG polyline компонент.
- `frontend/src/components/analytics/kpi-card.tsx` — заменяет/обёртывает `metric-card.tsx`.
- `frontend/src/components/analytics/narrative-banner.tsx` — AI narrative summary (template-based).
- `frontend/src/components/analytics/insight-action-card.tsx` — actionable color-coded card для Smart Insights items.
- `frontend/src/components/analytics/activity-heatmap.tsx` — GitHub-style 4-week × 7-day grid.
- `frontend/src/components/analytics/models-donut.tsx` — Recharts PieChart для моделей (заменяет horizontal bar).
- `frontend/src/components/analytics/streak-tracker.tsx` — current streak + 7 days dots.
- `frontend/src/components/analytics/compact-quotas.tsx` — однострочный quota footer.
- `frontend/src/lib/analytics-narrative.ts` — pure-function `buildNarrative(data, insights)` для тестируемости.
- Unit-тесты для каждого нового компонента (`*.test.tsx`).

### Меняем

- `frontend/src/pages/analytics.tsx` — полностью переписываем тело JSX. Логика хуков (`usePersonalAnalytics`, `useInsights`, `isMax`, `isPaid`) сохраняется.
- `frontend/src/components/analytics/metric-card.tsx` — мигрируем на `kpi-card.tsx` или расширяем существующий props (решим в Task 1 of impl plan).
- `frontend/src/components/analytics/insights-panel.tsx` — заменяем grid из старых cards на `<InsightActionCard>` для каждого insight type. Сохраняем three-state логику (Free → UpgradeGate, Pro → 2 + 5 locked, Max → 7).
- `frontend/src/components/analytics/insights-locked-card.tsx` — рестайл под visual consistency с `insight-action-card.tsx` (тот же layout, dashed border вместо solid, lock-icon вместо action-icon).
- `frontend/src/pages/__tests__/analytics-insights-states.test.tsx` — обновляем под новые компоненты (selectors `getByText(...)` могут сломаться).

### Удаляем (опционально)

- `frontend/src/components/analytics/model-segmentation-chart.tsx` — заменяется `models-donut.tsx`. Если есть external consumers — оставляем как deprecated.
- `frontend/src/components/analytics/quota-progress.tsx` — теперь рендерится внутри `compact-quotas.tsx` как inline-bar. Сам файл можно оставить, если используется в `/teams/:id/analytics`.

### Сущности / типы

```typescript
// frontend/src/components/analytics/sparkline.tsx
interface SparklineProps {
  points: number[]              // нормализуется на input — без оси, без zero-axis
  trend?: "up" | "down" | "neutral"  // определяет color (emerald/rose/slate)
  width?: number                // default 120
  height?: number               // default 22
}
export function Sparkline(props: SparklineProps): JSX.Element

// frontend/src/components/analytics/kpi-card.tsx
interface KpiCardProps {
  label: string                 // короткий "Использования"
  value: string | number
  delta?: number | null         // %; null → "—" (нет базы)
  sparkline?: number[]
  icon: LucideIcon              // обязательная
}

// frontend/src/components/analytics/insight-action-card.tsx
type InsightTone = "warning" | "info" | "success"
interface InsightActionCardProps {
  type: string                  // models.InsightType (unused_prompts | ...)
  tone: InsightTone
  icon: LucideIcon
  title: string                 // "Забытые"
  description: string           // "5 промптов не использовались 30+ дней"
  href: string                  // куда вести по клику (deep link)
  count?: number                // показывается в углу
}

// frontend/src/lib/analytics-narrative.ts
export function buildNarrative(
  data: PersonalAnalyticsResponse,
  insights: SmartInsight[] | null,
): NarrativeSegments

export interface NarrativeSegments {
  summary: string               // "За 7 дней: 234 использования +23%"
  topModel: string | null       // "топ-модель Claude (62%)"
  streak: string | null         // "streak 5 дней"
  actionHint: string | null     // "5 забытых и 2 дубликата ждут уборки"
}
```

### Контракты слоёв

- **Page → Component:** `analytics.tsx` передаёт `data: PersonalAnalyticsResponse` (existing) в каждый компонент. Никаких новых hooks.
- **Component → Component:** все новые компоненты — pure presentational, никакого fetch/мутаций.
- **Lib → Component:** `buildNarrative()` чистая функция, импортируется в `narrative-banner.tsx`.

---

## 4. Модель данных

**N/A** — этот redesign не меняет схему БД и не вводит новые API-endpoints. Используем существующий ответ `PersonalAnalyticsResponse` из `frontend/src/api/analytics.ts`.

---

## 5. API контракт

**N/A** — не меняем существующие endpoints, не добавляем новые. Three open questions (time-heatmap endpoint, comparison-overlay daily breakdown, share-funnel conversion) явно отложены в follow-up.

---

## 6. Зависимости

- **lucide-react** — уже в `package.json`, версия в проекте.
- **recharts** — уже в `package.json`, для PieChart (`models-donut.tsx`) и существующих AreaChart.
- **Никаких новых deps.** Tremor / nivo / @visx — отвергнуто (см. §2 Решение 1, 2).

---

## 7. План тестирования

### Unit

- **`sparkline.test.tsx`** — render с empty `points`, single-point edge case, gradient color по `trend`.
- **`kpi-card.test.tsx`** — render с / без sparkline, render с null delta → "—", render с positive/negative delta → правильная иконка + цвет.
- **`insight-action-card.test.tsx`** — render для каждого tone (warning/info/success), CTA href корректный, count badge.
- **`activity-heatmap.test.tsx`** — нормализация opacity по count, render с пустым массивом → 28 пустых клеток.
- **`models-donut.test.tsx`** — table-driven по `usage_by_model`: empty → empty state, top-6 + others, цвета из `MODEL_COLORS`.
- **`narrative-banner.test.tsx`** + **`analytics-narrative.test.ts`** — table-driven по `buildNarrative` для 4 сегментов; pure-function, no DOM.
- **`compact-quotas.test.tsx`** — render с тремя progress, percentage расчёт, color при >90%.

### Integration

- **`analytics-insights-states.test.tsx`** (существующий, обновлённый) — three-state UI: Free→UpgradeGate, Pro→InsightsPanel с 2 типами + 5 InsightsLockedCard, Max→7 типов без lock'ов. Селекторы обновляются под новые компоненты.

### E2E

- **N/A** — Playwright не добавляем для этой итерации. Manual smoke в browser достаточен (как в Pricing v3).

### Не тестируем

- Pixel-perfect рендеринг (это design implementation, не behavior).
- Recharts internal рендеринг donut (тестируется upstream).
- Tailwind classes presence (низкий ROI на test).

---

## 8. Наблюдаемость

### Метрики

**N/A для backend.** Frontend-only redesign. Если хотим отслеживать engagement — отдельный analytics-tracking (Plausible / Mixpanel) — вне scope.

### Логи

- `slog.Warn("analytics.narrative.build_failed", ...)` — если builder упадёт на edge-case данных (empty totals, NaN delta). Throw guard в `buildNarrative` → return empty `summary`.

### Sentry / GlitchTip

- Existing error boundary в `frontend/src/components/error-boundary.tsx` отловит исключения. Дополнительная инструментация не нужна.

---

## 9. План внедрения

| ID | Шаг | Owner | Критерий готовности | Зависит |
|---|---|---|---|---|
| **S1** | `Sparkline` компонент + unit-test. | frontend | `npx vitest run sparkline` зелёный. | — |
| **S2** | `KpiCard` компонент (расширение MetricCard) + unit-test. | frontend | `npx vitest run kpi-card` зелёный; визуально совпадает с mockup. | S1 |
| **S3** | `buildNarrative()` lib-function + unit-test (table-driven). | frontend | 6+ test cases зелёные (empty, only_summary, with_streak, action_hint, full, edge_NaN). | — |
| **S4** | `NarrativeBanner` компонент. | frontend | Render с моком `data` показывает все 4 сегмента. | S3 |
| **S5** | `InsightActionCard` компонент + unit-test. | frontend | 3 tone cases (warning/info/success), href корректный. | — |
| **S6** | Обновить `insights-panel.tsx` для использования `InsightActionCard` per type. Сохранить three-state логику. | frontend | `analytics-insights-states.test.tsx` обновлён и зелёный для всех 3 plans. | S5 |
| **S7** | `ActivityHeatmap` компонент + unit-test. | frontend | Render с 28 точками показывает grid 4×7. | — |
| **S8** | `ModelsDonut` компонент (Recharts PieChart) + unit-test. | frontend | Empty state + top-6 + legend; цвета совпадают с `MODEL_COLORS`. | — |
| **S9** | `StreakTracker` компонент. | frontend | Render с current+7d dots. | — |
| **S10** | `CompactQuotas` компонент. | frontend | Render с тремя quota. | — |
| **S11** | Реструктура `analytics.tsx` — собрать всё в Bento Grid. | frontend | Browser smoke: все секции рендерятся в правильном порядке; mobile fallback на `md` breakpoint работает. | S1-S10 |
| **S12** | Обновить `analytics-insights-states.test.tsx` под новые селекторы. | frontend | Test зелёный. | S11 |
| **S13** | Manual browser smoke на 3 тарифах (Free/Pro/Max) через docker-dev. | frontend | Скриншоты сохранены в `docs/superpowers/screenshots/analytics-redesign-{free,pro,max}.png`. | S11 |
| **S14** | Cleanup: удалить deprecated компоненты (`model-segmentation-chart.tsx` если нет consumers). | frontend | grep по `model-segmentation-chart` показывает 0 imports. | S11 |

**Atomicity.** Каждый шаг — отдельный commit. S1-S10 — независимые компоненты, мержатся в любом порядке. S11 ждёт всех. S13 — manual, не CI-блокер.

**Размер диффов:**

- S1-S10: каждый ~50-150 строк (компонент + test).
- S11: ~200 строк (полный rewrite analytics.tsx JSX body).
- Остальные: < 50 строк.

---

## 10. Rollout и kill-switch

### Стратегия

- **Direct prod.** Frontend-only визуальный rewrite без feature flag. Существующие тесты + manual smoke на 3 тарифах достаточны.
- Если что-то пойдёт не так — revert одного коммита в `analytics.tsx` восстанавливает старый layout (старые компоненты не удаляются в этом шаге, см. S14 — cleanup отдельный).

### Feature flags

**N/A** — никаких runtime-флагов не вводим. Это чистый визуальный refactor, behavior не меняется.

### Kill-switch RTO

- Git revert + redeploy frontend: ~5-10 минут.

### Communication

- Changelog `/changelog`: «Обновили дизайн страницы Аналитика».
- In-app banner — не нужен, изменение визуальное.

---

## 11. Документация

- **README** — N/A.
- **ADR** — N/A. Решения 1-3 в §2 достаточно для контекста, не требуют отдельных ADR (нет долгосрочной архитектурной важности).
- **Runbook** — N/A (нет background-loops / cron).
- **CLAUDE.md** — добавить 1-2 строки про analytics page структуру (Bento Grid, KPI strip, NarrativeBanner) в раздел «Ключевые решения» или Frontend конвенции.

---

## 12. Риски и митигации

### Технические риски

- **Рекурсивная замена `MetricCard` → `KpiCard`.** Существующий `MetricCard` может использоваться в других местах (`/teams/:id/analytics`, dashboard widgets). **Митигация:** grep по `MetricCard` перед удалением; либо оставить `MetricCard` как `KpiCard` alias.
- **Bento Grid на mobile.** 6-col grid на 360px ширине ломается, нужен явный breakpoint. **Митигация:** на `<md` все карточки `col-span-1; row-span-1` — vertical-stack как сейчас.
- **Sparkline rendering performance.** 4 KPI cards × sparkline на mount = 4 SVG polyline. Поскольку sparkline — pure SVG, render < 16ms.
- **Recharts PieChart bundle impact.** Уже в bundle (используется в `usage-chart.tsx` через AreaChart), добавление PieChart — incremental, не значимый.
- **AI narrative text quality.** Template без LLM может звучать «казённо» на edge-cases (0 использований за период → «За 7 дней: 0 использований»). **Митигация:** для пустых данных показывать другой text «За 7 дней пока тихо — самое время попробовать новые промпты».

### Pre-mortem: «через 6 месяцев это сломалось»

1. **Юзеры не воспринимают новый layout как «лучше старого».** Engagement не вырос. **Митигация:** A/B-тест через feature flag (если уже есть инфра) — но мы решили без flag. Альтернатива: post-launch survey через `feedback`-механизм.
2. **Bento Grid не масштабируется когда добавляем новые виджеты.** Перепутаются col-span/row-span. **Митигация:** документировать grid layout в комментариях `analytics.tsx`.
3. **Lucide icon resolution в TS.** Если кто-то переименует `Trophy` → `TrophyIcon` upstream — typecheck упадёт. **Митигация:** lockfile + Renovate для контролируемых обновлений.

### Известные ограничения

- Time-heatmap (часы×дни) — не делаем, требует backend.
- Comparison-overlay — не делаем, требует backend daily breakdown.
- Share-funnel conversion — не делаем, требует tracking referrer.
- Mobile experience — degraded (vertical-stack), не полный mobile redesign.

---

## 13. Метрики успеха

### Бизнес

- **Engagement на /analytics** (если есть аналитика страничного traffic). Цель: time-on-page +30% за месяц.
- **Click-through Smart Insights actions.** Цель: > 15% юзеров кликают на actionable card в течение месяца.

### Технические

- Bundle size delta: не больше +25 KB gzipped (новые компоненты + custom SVG).
- LCP (Largest Contentful Paint) на `/analytics`: не вырастает > +100ms vs текущая.
- Lighthouse Accessibility: ≥ 95.

### Срок измерения

- **1 неделя** post-launch — bundle/performance metrics.
- **30 дней** — engagement metrics.

---

## 14. Открытые вопросы

1. **`MetricCard` deprecation** — оставляем как alias на `KpiCard` или удаляем после grep'а? Кому: исполнитель. Блокирует ли план: нет.
2. **`buildNarrative` empty-state copy** — какой текст для «0 использований за период»? Кому: пользователь (нужно решение по copy). Блокирует: нет.
3. **Иллюстрации в empty states** — оставляем плейн текст или подключаем существующий illustration pack? Кому: пользователь. Блокирует: нет.
4. **Включать ли follow-up time-heatmap в текущую ветку** — или отдельная фича после backend-доработки? Кому: пользователь. Блокирует: нет.

---

## Self-check

- [x] **Инвентаризация инструментов проведена.** Evidence: AskUserQuestion (1 раз для scope), Read (5 файлов: analytics.tsx, metric-card.tsx, usage-chart.tsx, model-segmentation-chart.tsx, insights-panel.tsx), Glob/Bash для ls компонентов, WebSearch (2 запроса — SaaS dashboards 2026, Tremor), TaskCreate/TaskUpdate (8 brainstorming tasks), visual companion server.
- [x] **Прочитан релевантный код.** Evidence: `pages/analytics.tsx:31-235`, `components/analytics/metric-card.tsx:13-30`, `components/analytics/usage-chart.tsx:14-63`, `components/analytics/model-segmentation-chart.tsx:44-131`.
- [x] **Внешняя документация.** Evidence: WebSearch — Tremor/Vercel acquisition, SaaS UI 2026 trends (F-pattern, KPI strip 80-120px), dashboard patterns. [НЕДОСТУПНО: context7 quota exceeded — обошлись WebSearch].
- [x] **Архитектурные решения с альтернативами.** Evidence: §2 Решение 1 (sparkline: custom SVG / Recharts / Tremor), Решение 2 (Bento Grid: pure CSS / react-grid-layout / Subgrid), Решение 3 (narrative: template / backend AI / hardcoded).
- [x] **Нет over-engineering.** Evidence: отвергнут Tremor (отдельная UI-lib ради одного компонента), отвергнут react-grid-layout (нет draggable требования), отвергнут backend AI-call для narrative.
- [x] **Edge cases / errors.** Evidence: §12 (empty data, mobile, bundle impact, icon resolution), `buildNarrative` empty-state copy в §14.
- [x] **Консистентность с проектом.** Evidence: shadcn/ui Card / Recharts / Lucide React уже в `package.json`, MODEL_COLORS из existing `model-segmentation-chart.tsx` переиспользуем, three-state Pro Insights teaser из Pricing v3 сохраняется.
- [x] **Критерии готовности конкретны.** Evidence: §9 — S1: vitest sparkline зелёный; S11: browser smoke + mobile fallback; S12: existing test обновлён.
- [x] **Допущения помечены.** Evidence: open questions §14 + 3 explicit «отложено» (time-heatmap, comparison-overlay, share-funnel).
- [x] **Scope discipline.** Evidence: явно отложены 3 фичи требующих backend; team analytics page вне scope; mobile redesign degraded.
