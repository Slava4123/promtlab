# Security Follow-up

Документ фиксирует 2 security findings низкого приоритета из REVIEW_2026-05-07 (Low severity), которые не закрываются code-only fix'ом и требуют отдельной работы — архитектурного дизайна, ops-инфраструктуры или процессного решения.

## 1. Refresh abuse detection

### Контекст

Сейчас защита refresh-токенов:

- `subtle.ConstantTimeCompare` для nonce при validate (`backend/internal/usecases/auth/auth.go:656`) — закрыто финальной волной v1.
- Refresh-cookie HttpOnly + AllowCredentials с whitelist origins (CORS).
- `users.token_nonce` обновляется при logout/InvalidateTokens/ResetPassword/ChangePassword.
- Rate limiting per IP на `/auth/*` (20 rpm).

Что **не покрыто:** аномалии при использовании refresh-токенов с разных устройств/IP одновременно. Пример:

1. Юзер логинится с домашнего IP → получает refresh-токен.
2. Атакующий захватывает refresh-токен (XSS отсутствует, но возможен другой вектор — leakproxy, физический доступ к устройству).
3. Атакующий refresh'ит с другого IP/UA в течение TTL (7 дней).
4. Сейчас система не различает "юзер сменил Wi-Fi" vs "second device hijack".

### Предлагаемое решение

Event log + аномалия-детектор:

1. **Auth event log** — таблица `auth_events`:
   ```
   id, user_id, event_type (login|refresh|logout), ip, user_agent_hash,
   timestamp, asn (через MaxMind GeoLite2 — geolocation lookup IP →
   ASN/country, бесплатная база, обновляется ежемесячно)
   ```
2. **Anomaly detector** — фоновый loop (`auth_anomaly_loop.go`):
   - Refresh с разных ASN в течение 1 часа → flag.
   - Refresh с разных стран в течение 24 часов → flag.
   - >5 refresh'ев за 1 минуту с разных IP → invalidate token.
3. **Notify юзера** — email при detection с возможностью "это был я" / "это не я → invalidate all sessions".
4. **Алерты в admin panel** — список flagged users.

### Trade-offs

- **Pro:** detect realistic ATO scenarios, восстанавливает trust.
- **Contra:** false positives для легитимных юзеров (поездки, VPN-юзеры). Email storm risk.
- **Effort:** **L** (~1 неделя) — миграция + loop + email шаблон + admin UI + GeoLite2 интеграция.
- **Risk:** низкий — opt-out flag + threshold tuning после первой недели.

### Decision

**Откладывается** до момента, когда у нас будет:

1. Дашборд аналитики login event'ов (для baseline нормальной активности).
2. Подтверждённый случай ATO (или high-value account, который требует это active'но).

Tracked в SECURITY backlog. Owner — TBD.

---

## 2. dev-secret для staging

### Контекст

`backend/internal/infrastructure/config/loader.go:212-221` отвергает `dev-secret-change-me` как JWT_SECRET в production:

```go
if cfg.Server.IsProd() && secret == "dev-secret-change-me" {
    return fmt.Errorf("JWT_SECRET must be changed in production")
}
```

Однако `IsProd()` определяется по `cfg.Server.Env == "production"`. Если staging environment деплоится с `Env=staging`, валидация **пропускает** dev-secret для staging. Это значит:

- Если на staging выкатить тот же `dev-secret-change-me` — JWT-токены staging теоретически валидны и в production (если случайно пересеклись по `Env`).
- Атакующий с доступом к staging logs (более слабая защита) может извлечь токены и попытаться использовать их в prod.

### Предлагаемое решение

**Ops-task:** создать отдельный `.env.staging` с distinct `JWT_SECRET` ≥ 32 chars.

1. **Сейчас:** добавить `cfg.Server.Env == "staging"` в guard `validateProdJWTSecret` — staging должен использовать non-default secret.
2. **Deploy infra:** `.env.staging` файл (не в репо!) с `JWT_SECRET=<32-char random>`. Хранится в secret manager (либо ручной copy на staging VPS).
3. **Документация:** обновить `docs/DEPLOY.md` с секцией про staging-specific env.

### Trade-offs

- **Pro:** изолирует staging cryptographic boundary от prod, даже при leaked staging secrets.
- **Contra:** требует staging environment, который **сейчас не существует**. Нет dedicated staging VPS / domain.
- **Effort:** **S code** (5 строк в loader.go) + **M ops** (создать staging deployment).

### Decision

**Откладывается** до момента создания staging environment. Currently project использует:

- **Local dev** — `.env.dev` с `dev-secret-change-me` (OK по замыслу).
- **Production** — `.env.prod` с настоящим JWT_SECRET (OK).
- **Staging** — пока не существует.

Когда staging environment будет создан, открыть PR на:

1. Изменение `validateProdJWTSecret` чтобы покрыть `Env in {production, staging}`.
2. Добавление `.env.staging.example` в репо с pattern (без реального secret).
3. Обновление `docs/DEPLOY.md` с staging onboarding steps.

Tracked в DEPLOY backlog. Owner — TBD.

---

## Замечание

Оба finding'а — **Low severity** в REVIEW_2026-05-07. Production endpoint защищён (ProdJWTSecret валидация работает, refresh constant-time compare работает). Эти улучшения — defense-in-depth, не блокеры релиза.
