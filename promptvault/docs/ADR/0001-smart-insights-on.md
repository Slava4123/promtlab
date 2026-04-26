# ADR 0001 — Включить Smart Insights (4 расширенных типа) с Phase 15

**Дата:** 2026-04-26
**Статус:** Accepted
**Phase:** 15 (доделка базового функционала)

## Контекст

С Phase 14 backend `analytics.Service.ComputeInsights` считает 7 типов Smart
Insights, но 4 из них (`most_edited`, `possible_duplicates`, `orphan_tags`,
`empty_collections`) скрыты за `experimentalInsights=false` config flag.
Frontend `InsightsPanel` уже умеет рендерить все 7 типов (icons + labels).

Max-tier юзеры платят 1299₽/мес, а получают 3 из 7 заявленных в FEATURES
типов аналитики. Это пробел между ценой и ценностью.

`possible_duplicates` единственный тип, требующий PG-расширения `pg_trgm`.
На managed PG (Timeweb) расширения могут быть недоступны без admin-доступа —
миграция `000048_analytics_m8.up.sql` использует `IF NOT EXISTS`, но
неприменённое расширение приведёт к runtime-ошибкам.

## Решение

1. **Сменить default `experimental_insights` с `false` на `true`** —
   все 4 расширенных типа включены для Max-юзеров по умолчанию.
2. **Сохранить флаг как kill-switch.** Установка
   `ANALYTICS_EXPERIMENTAL_INSIGHTS=false` в `.env` мгновенно отключает
   все 4 расширенных типа без редеплоя.
3. **Отделить зависимость от pg_trgm в отдельный runtime probe.**
   `postgres.DetectExtensions(ctx, db)` проверяет наличие через
   `SELECT COUNT FROM pg_extension WHERE extname = 'pg_trgm'`.
   `analytics.Service.SetTrgmAvailable(v bool)` принимает результат,
   `ComputeInsights` оборачивает только `PossibleDuplicates` в
   `if s.trgmAvailable`. Остальные 3 расширенных типа от pg_trgm не зависят.

## Альтернативы рассмотрены

- **A. Удалить флаг полностью.** Минус: нет возможности экстренно отключить
  функцию без деплоя при найденной проблеме (например, медленный SQL для
  `MostEdited` на больших объёмах данных).
- **B. Считать pg_trgm обязательным, падать на старте если расширение
  недоступно.** Минус: блокирует деплой на managed PG без admin-доступа,
  что нарушает self-hosted принцип. Принцип fail-open для не-критичной
  фичи лучше.
- **C. Использовать существующий флаг `experimentalInsights` для гейта
  возможности pg_trgm.** Минус: смешивает две независимые концепции
  (kill-switch функции vs. наличие расширения), затрудняет дебаг
  ("Почему `most_edited` не работает на стейдже?" — потому что
  `experimental` = false ИЛИ потому что нет pg_trgm?).

## Последствия

- Max-юзеры получают полный Smart Insights дашборд после деплоя.
- Если pg_trgm недоступен — 6 из 7 типов работают, в логах при старте
  видно `postgres.capabilities trgm=false`. Юзер не видит ошибок.
- В случае инцидента (e.g. медленный SQL на проде) можно мгновенно
  отключить все 4 расширенных типа через `.env`.

## Источники

- `backend/internal/usecases/analytics/insights.go:54-96`
- `backend/internal/infrastructure/postgres/postgres.go:DetectExtensions`
- `backend/internal/infrastructure/config/analytics.go`
- Commit: `feat(analytics): включить Smart Insights (4 расширенных типа) + pg_trgm probe`
