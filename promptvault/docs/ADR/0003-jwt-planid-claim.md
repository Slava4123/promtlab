# ADR 0003 — Добавление PlanID в JWT claims (M9)

**Дата:** 2026-04-26
**Статус:** Accepted
**Phase:** 15 W2.8

## Контекст

Каждый `/api/analytics/*` endpoint делает отдельный
`SELECT users WHERE id=?` только ради `plan_id` (для retention clamp:
Free 7d / Pro 90d / Max 365d). 9 callsites в `analytics/service.go`.
На дашбордах юзер открывает несколько endpoint'ов — DB-hit на каждый.

Задача M9 в plan: убрать лишний DB lookup без потери корректности.

## Решение

**Положить `PlanID` в JWT access-токен. Читать из ctx с graceful fallback
на DB lookup для legacy-JWT.**

1. **`Claims.PlanID string`** — добавлено в `usecases/auth/types.go`.
   Заполняется в `generateTokenPair(userID, nonce, planID)` только для
   access-токена (refresh-токен переоформит свежий plan при rotate).
2. **`middleware/auth.Middleware`** — кладёт весь `*Claims` в ctx через
   новый ключ `ClaimsKey` рядом с существующим `UserIDKey` (без breaking
   change для существующих handlers).
3. **`analytics.Service.SetPlanFromCtx(fn)`** — callback для чтения plan
   из ctx. Реализация в `app.go` делает type-assert на
   `*authuc.Claims` через `authmw.ClaimsKey`. Analytics-сервис не
   импортирует middleware/auth напрямую — избегаем cyclic import
   (middleware → usecases/auth → analytics → middleware).
4. **`analytics.Service.lookupPlanID(ctx, userID)`** — read-through helper:
   сначала `planFromCtx(ctx)`, потом `users.GetByID(ctx, userID)`.
   Все 9 callsites переписаны на этот helper.
5. **Старые JWT** (выпущенные до Phase 15) имеют `PlanID == ""`. Helper
   делает graceful fallback на DB. Через 7 дней (refresh TTL) все JWT
   обновятся с полем — DB hits сами пропадают.

## Альтернативы рассмотрены

- **A. Middleware preload `*User`** в ctx через дополнительный
  `users.GetByID` в auth middleware. Плюс: всегда свежие данные. Минус:
  лишний DB-hit на КАЖДЫЙ запрос, не только analytics — хуже текущего
  состояния.
- **B. In-memory cache (`sync.Map[uint]string`).** Минус: complexity
  invalidation, race conditions при admin override через ChangeTier,
  TTL choice (`ttl=15m` ≈ access-токен TTL — но почему не использовать
  сам токен).
- **C. Redis cache.** Новая зависимость только ради plan_id —
  непропорционально.

## Trade-offs

**Stale data в течение access-TTL (15 минут).** Если admin меняет тариф
юзера через `/api/admin/users/:id/tier` — у юзера в JWT остаётся старый
PlanID до следующего refresh. Это приемлемо для retention clamp:

- Понижение Max → Free: юзер до 15 мин видит дашборд за 365d вместо 7d.
  Не реальная утечка — данные уже посчитаны, ничего нового не приходит.
- Повышение Free → Pro: юзер до 15 мин видит дашборд за 7d. Frontend
  получит свежие данные на следующем refresh — приемлемая UX.

При желании немедленного эффекта admin может попросить юзера
relogin'нуться — это force-генерирует новый JWT.

## Последствия

- Минус 1 DB-hit на каждый /api/analytics/* запрос для юзеров с свежим JWT.
- В Prometheus можно мониторить: `users.GetByID` calls со scope analytics
  должны падать после 7-дневного rolling window после деплоя.
- Аналогичный паттерн можно применить для других hot-path запросов
  (apikey scope, role) если возникнут.

## Источники

- `backend/internal/usecases/auth/types.go:Claims`
- `backend/internal/usecases/auth/claims_helper.go`
- `backend/internal/usecases/analytics/service.go:lookupPlanID`
- [RFC 8725 §2.5 JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- Commit: `feat(analytics): убрать users.GetByID из hot-path через PlanID в JWT (M9)`
