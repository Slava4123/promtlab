# Runbook: TeamBrandingLogoUploadErrorRate

**Severity:** P2
**Pager:** Telegram silenceable

## Symptom

Юзеры жалуются «логотип не загружается» / 4xx-5xx на `POST /api/teams/{slug}/branding/logo`. Алерт срабатывает когда доля reject'ов > 20% за 10 минут.

## Impact

Только Max-юзеры с активной командой. Без брендинга публичные share-страницы продолжают работать (в URL-режиме или без логотипа). Backward path: юзер может переключиться в URL-режим в форме `/teams/:slug/branding` — это не требует bytea storage и работает независимо.

## Investigation

```bash
# 1. Распределение reject'ов по типу
# Prometheus query (за последние 30 минут):
#   sum by (result) (increase(team_branding_logo_uploads_total[30m]))
#
# Ожидаемое в health: result="success" >> остальные.
# 'too_large' / 'bad_format' от юзеров — норма (юзер пробует пока не подгонит).
# 'forbidden' — норма для не-Max и не-owner попыток.
# Скачок 'other' = регрессия в usecase или БД.

# 2. Логи API контейнера
docker logs promptvault-api-1 --tail 200 | grep -E "team\.branding\.logo|team/logo"

# 3. БД — есть ли place для bytea (TOAST overhead, free space)
docker exec promptvault-postgres-1 psql -U postgres -d promptvault -c "
  SELECT
    pg_size_pretty(pg_total_relation_size('team_logo_files')) AS table_size,
    COUNT(*) AS files,
    AVG(size_bytes)::int AS avg_bytes,
    MAX(size_bytes) AS max_bytes
  FROM team_logo_files;
"

# 4. Sentry — события 5xx (label result="other")
# Поиск по теме "team/logo" или "TeamBrandingLogoUploadErrorRate"
```

## Mitigation (immediate)

1. **Если result="too_large" / "bad_format" доминируют (юзер-error):** алерт ложный, `silence` на 24h, эскалация не нужна. Возможный follow-up: improve UX (показывать лимит до выбора файла — уже сделано в `frontend/src/components/teams/logo-uploader.tsx`).
2. **Если result="other" доминирует:** регрессия в коде. Проверить недавний deploy:
   ```bash
   git log --oneline -10 promptvault/backend/internal/usecases/team/logo.go promptvault/backend/internal/delivery/http/team/logo_handler.go
   ```
   Откат через `docker compose pull && docker compose up -d` на предыдущий тег image.
3. **Если ошибка не в нашем коде (БД переполнена):** добавить место к Postgres VOLUME, рассмотреть переход на FS-стратегию из ADR 0006 (план Б).

## Resolution (long-term)

- Если bytea-хранилище упирается в quota: реализовать миграцию на FS+nginx (см. ADR 0006 «Миграционные пути»).
- Если decode-ошибки повторяются на конкретном формате: написать unit-тест в `backend/internal/usecases/team/logo_test.go` с реальным файлом-кейсом, поправить magic-byte detection.

## Post-mortem

Если incident `> 30 min` или impact на >5% Max-юзеров — заполнить post-mortem template и приложить к incident'у в /admin.
