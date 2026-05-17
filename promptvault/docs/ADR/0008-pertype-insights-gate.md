# ADR-0008: Per-type Smart Insights gating (Pro teaser)

**Дата:** 2026-05-17
**Статус:** Принято

## Контекст
Pricing iteration v3 требует чтобы Pro юзеры видели **2 типа** Smart Insights
(unused + duplicates), а Max — все 7. До этой итерации использовался master-gate
`IsMax(planID)` (analytics/service.go:111) — Free/Pro получали 402, Max — всё.

## Решение
Per-type filter через hardcoded `proAllowedInsights []string` константу в
`usecases/analytics`. `insightsForPlan(planID)` возвращает разрешённый набор;
`GetInsightsGated` filter'ит результат `GetInsights` в памяти; `ComputeInsights`
принимает `allowed []string` параметр и пропускает SQL для не-allowed типов.

## Альтернативы
- **JSONB-колонка `subscription_plans.allowed_insight_types TEXT[]`** — гибкость,
  можно менять без релиза. Минус: overengineering для 3 tier'ов.
- **Env feature flag `PRO_INSIGHTS_TYPES=unused,duplicates`** — можно менять без
  релиза. Минус: конфигурация state в env vs БД — антипаттерн.
- **Const в коде (выбран)** — 1 место правды, типобезопасный.

## Trade-offs
✅ Простой, читаемый.
✅ Тестируется table-driven case'ами.
❌ Добавление 4-го tier'а требует кодовых правок (acceptable — 3 tier'а 1.5 года).

## Импликации
- Feature flag `PRO_INSIGHTS_TEASER_ENABLED` — для kill-switch в первую неделю
  после deploy.
- `ListMaxUsers` репозиторий переименован в `ListPaidUsers` (включает Pro).
- Loop делает extra-roundtrip в `users.GetByID(uid)` для определения `allowed`.
