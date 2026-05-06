# ADR 0006: Хранилище загруженных логотипов команды — bytea в Postgres

**Status:** Accepted (2026-05-06)
**Phase:** 16-X (Branding UX)

## Контекст

В Phase 14 D реализован брендинг публичных share-страниц команды (Max-only): логотип, подпись, сайт, основной цвет. Логотип хранился как `brand_logo_url VARCHAR(500)` — внешняя ссылка на CDN юзера. Это работает только для тех Max-юзеров, у кого свой публичный CDN; обычный пользователь не имеет такой ссылки и упирается в форму.

Задача: дать возможность загрузить файл прямо в форме на `/teams/:slug/branding` без зависимостей от внешнего CDN. Self-hosted в РФ; никакие платные стораджи (S3, CloudFront) недопустимы.

## Решение

Файл хранится как **`bytea` в новой таблице `team_logo_files`** (1:1 к teams через FK CASCADE). Раздаётся через `GET /api/teams/:slug/branding/logo` с `ETag: "<sha256>"` и `Cache-Control: public, max-age=86400`. Источник логотипа дискриминируется enum-колонкой `teams.brand_logo_source ∈ {url, file, none}`.

**Лимиты:** ≤1 МБ, форматы whitelist `image/png|jpeg|webp` (без SVG = XSS), pixel-dim ≤1024×1024.

**Validation:** магик-байты через `http.DetectContentType` + `image.DecodeConfig` (std lib + `golang.org/x/image/webp`). Polyglot файлы (например, PNG header + JPEG body) отвергаются на decode-уровне.

## Альтернативы

### А. Local FS + nginx serve

Volume `/var/lib/promptvault/uploads` в `docker-compose.prod.yml`, nginx location `/uploads/team-logos/...`.

**Плюсы:** Postgres не пухнет; раздача через nginx эффективнее single-row SELECT bytea.
**Минусы:**
- Новый volume в compose → новая операция: backup отдельно от БД.
- Risk orphan-файлов при rollback миграции (DROP TABLE удаляет metadata, файлы остаются в volume).
- Операция «удалить команду» становится cross-cutting: нужен hook на cleanup из `teamRepo.Delete`. Сейчас FK CASCADE решает это бесплатно.
- Нужен путь-генератор без collisions (uuid в имени) — ещё один movbody.

### Б. MinIO в docker-compose

S3-совместимый сервис в стеке.

**Плюсы:** presigned URL, готовая инфраструктура для будущих uploads (avatars, attachments).
**Минусы:**
- +200 МБ RAM, +50% сетевого overhead на bucket-bookkeeping.
- Доп. сервис в self-hosted VPS с 8 ГБ RAM = 2.5% от capacity на одну вспомогательную фичу.
- Lifecycle/auth конфигурация (IAM-policy, bucket policy) усложняют 100-line PR до 300+.
- Если у нас будет всего 1000 команд × 1 МБ — overkill.

### В. Bytea в Postgres (выбрано)

**Плюсы:**
- Ноль новой инфры. Никаких изменений в `docker-compose.prod.yml`/`nginx.conf`.
- Атомарность: upload + переключение `brand_logo_source` в одной транзакции (важно для consistency).
- Backup вместе с БД через существующий `pg_dump` pipeline.
- FK ON DELETE CASCADE автоматически чистит файлы при удалении команды — нулевой риск orphan'ов.
- Объём: max 1 МБ × ≤1000 команд = ≤1 ГБ; ~5% от размера текущей БД, амортизируется.

**Минусы:**
- TOAST overhead для blob'ов до 1 МБ — single-row SELECT остаётся <50ms p95 (см. SLO в плане), это не hot-path (≤1 раз в год per team).
- При 10× росте аудитории (10K команд × 1 МБ = 10 ГБ) BD-объём станет заметным — будем мигрировать на FS+nginx (план Б ниже).

## Last-mile: Cache strategy

`Cache-Control: public, max-age=86400` (24ч) — браузер кэширует, но **не immutable**: ETag меняется при замене файла, поэтому 24h revalidation пробуждает актуальное обновление. immutable дал бы более жёсткий cache-pin, но поломал бы UX «загрузил новый — не отображается».

`If-None-Match: "<sha256>"` → `304 Not Modified`. На прогретом сценарии (юзер открывает share-страницу второй раз) вся отдача = 304 без чтения bytea из БД.

## Защита от загрузок-эксплоитов

- `MaxBytesReader(1MiB)` на handler-level — обрывает соединение до multipart-парсинга.
- Magic-byte detection via std lib (НЕ ImageMagick — нет CVE-сюрпризов).
- `image.DecodeConfig` отвергает polyglot-файлы.
- Whitelist content-type, без SVG (XSS-вектор через `<script>` внутри SVG).
- Pixel-dim ≤1024×1024 — защита от «zip-bomb» PNG, который декодируется в гигабайты.
- Rate-limit 10/час/userID на upload — защита от DoS.

## Миграционные пути (если решение разонравится)

### Bytea → Local FS + nginx
1. Новая таблица `team_logo_files_v2` с колонкой `path TEXT` вместо `bytes BYTEA`.
2. Миграционный скрипт: `INSERT … VALUES (team_id, content_type, size_bytes, sha256, '/uploads/team-logos/' || team_id || '.' || ext, …) FROM team_logo_files`. Параллельно записать файлы в FS.
3. Переключение feature flag `LOGO_STORAGE=fs` в env. Handler `Serve` начинает отдавать redirect 302 → `nginx /uploads/...`.
4. После недели стабильности: `DROP TABLE team_logo_files`.

### Bytea → MinIO
Аналогично, но `path` = presigned URL; handler становится прокси-redirect.

## Последствия

- ✅ Maintenance: одна таблица, один backup, один сервис.
- ✅ Code: ~300 строк нового кода (handler 130 + usecase 150 + tests).
- ⚠️ Capacity: при 10× scale — нужно мигрировать (доступная дорожка описана выше).
- ⚠️ Производительность: bytea TOAST round-trip заметнее CDN-edge на масштабе тысяч RPS. Сейчас не на масштабе.

## Связанные документы

- План реализации: `~/.claude/plans/microsoft-windows-version-10-0-19045-645-idempotent-pizza.md`
- Миграция: `backend/internal/infrastructure/postgres/migrations/000060_team_logo_files.up.sql`
- Runbook: `docs/runbooks/TeamBrandingUploadErrors.md`
