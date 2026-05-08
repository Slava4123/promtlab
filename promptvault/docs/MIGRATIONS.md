# Database Migrations — Best Practices

> Создан в рамках **CR-15 + CR-16** из `REVIEW_2026-05-07.md` после обнаружения, что:
> - все 61 миграция в `backend/internal/infrastructure/postgres/migrations/` создавали индексы **без `CONCURRENTLY`** — каждый `CREATE INDEX` на горячей таблице (`prompts`, `prompt_usage_log`, `share_view_log`, `team_activity_log`, `audit_log`, `prompt_chain_executions`) брал `ShareLock`, блокируя все INSERT/UPDATE/DELETE на время билдинга индекса;
> - миграция `000051_prompts_fts` делала `ALTER TABLE prompts ADD COLUMN search_tsv tsvector GENERATED ALWAYS AS (...) STORED` — полный rewrite таблицы под `AccessExclusiveLock`. Сейчас prod выживает только потому, что `prompts` маленькая.

Этот документ — обязательное чтение перед PR'ом, который добавляет SQL-миграцию.

---

## Правила

### 1. Индексы на горячих таблицах — ВСЕГДА `CONCURRENTLY`

**Горячие таблицы:**
- `prompts`, `prompts_versions`
- `prompt_usage_log`, `share_view_log`, `team_activity_log`, `audit_log`
- `prompt_chain_executions`, `prompt_chain_steps`
- `users` (10K+ rows)

**Bad:**
```sql
CREATE INDEX idx_prompts_team_updated ON prompts (team_id, updated_at DESC);
```

**Good:**
```sql
-- migrate:no-transaction
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_team_updated
    ON prompts (team_id, updated_at DESC);
```

**Зачем:**
- `CREATE INDEX` без `CONCURRENTLY` берёт `ShareLock` — блокирует все INSERT/UPDATE/DELETE на время построения индекса.
- На проде с заметным RPS это ведёт к stall запросов, тайм-аутам, накоплению connection pool, провалу health-check после deploy и rollback.
- `CONCURRENTLY` строит индекс параллельно, без блокировки writes. Цена — индекс строится ~2× медленнее.

**Гatchas:**
- `CONCURRENTLY` нельзя использовать внутри transaction. `golang-migrate/migrate` v4 по дефолту оборачивает каждую up.sql в транзакцию — поэтому **обязательна директива `-- migrate:no-transaction` в первой строке файла**, чтобы migrate отключил BEGIN/COMMIT.
- Без `IF NOT EXISTS` повторный запуск миграции после партial failure (например, OOM в середине) даст ошибку. Всегда писать `IF NOT EXISTS` для CONCURRENTLY-индексов.
- Drop'ать тоже через `DROP INDEX CONCURRENTLY IF EXISTS` в .down.sql.

### 2. `ADD COLUMN GENERATED ALWAYS AS … STORED` на существующих таблицах — 3-шаговый rollout

**Bad** (миграция 000051 в её текущем виде):
```sql
ALTER TABLE prompts ADD COLUMN search_tsv tsvector
    GENERATED ALWAYS AS (to_tsvector('russian_unaccent', title || ' ' || content)) STORED;
```

→ `AccessExclusiveLock` + полный rewrite таблицы. На 100K промптов = 30-90 секунд downtime API. Сейчас `prompts` маленькая, риск низкий, но будущий scaling упрётся в это.

**Good** (CR-15 pattern):
```sql
-- Миграция 000062: ADD COLUMN nullable, БЕЗ GENERATED — быстрая, lock метаданных.
ALTER TABLE prompts ADD COLUMN search_tsv tsvector;
```

```go
// Backfill в Go — chunks по 1000 rows, между chunks pg_sleep(50ms) если нужно.
for {
    res := db.Exec(`
        UPDATE prompts
        SET search_tsv = to_tsvector('russian_unaccent', title || ' ' || content)
        WHERE id IN (
            SELECT id FROM prompts WHERE search_tsv IS NULL ORDER BY id LIMIT 1000
        )
    `)
    if res.RowsAffected == 0 { break }
}
```

```sql
-- Миграция 000063: переключаем на GENERATED ALWAYS — теперь таблица полная,
-- никакого rewrite не происходит, только смена column attribute.
ALTER TABLE prompts ALTER COLUMN search_tsv
    SET GENERATED ALWAYS AS (to_tsvector('russian_unaccent', title || ' ' || content)) STORED;
```

```sql
-- migrate:no-transaction
-- Миграция 000064: индекс отдельно, через CONCURRENTLY.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_search_tsv ON prompts USING GIN (search_tsv);
```

**Альтернатива для маленьких таблиц (<10K rows):** использовать одношаговую миграцию + зафиксировать в runbook что nuжен low-traffic window для applying.

### 3. `ALTER COLUMN … SET NOT NULL` — двухшаг с `NOT VALID`

**Bad:**
```sql
ALTER TABLE subscriptions ALTER COLUMN paused_until SET NOT NULL;
```

→ Полный scan всех существующих rows под `AccessExclusiveLock`.

**Good:**
```sql
-- Шаг 1: добавляем CHECK constraint NOT VALID — мгновенно, новые insert'ы будут проверяться.
ALTER TABLE subscriptions ADD CONSTRAINT chk_paused_until_not_null CHECK (paused_until IS NOT NULL) NOT VALID;
-- Шаг 2 (отдельная миграция, после backfill'а NULL значений):
ALTER TABLE subscriptions VALIDATE CONSTRAINT chk_paused_until_not_null;
-- Шаг 3 (опционально, для performance): SET NOT NULL — теперь scan не нужен,
-- PG использует наш VALID constraint как доказательство.
ALTER TABLE subscriptions ALTER COLUMN paused_until SET NOT NULL;
ALTER TABLE subscriptions DROP CONSTRAINT chk_paused_until_not_null;
```

### 4. `ALTER TABLE` с TYPE conversion (varchar → text, int4 → int8 etc.) — partial UPDATE + CHECK

Любой type change в существующей колонке = rewrite таблицы. Если таблица крупная — заменять через `ADD COLUMN new_col` + backfill + `RENAME` swap.

### 5. FK на горячую таблицу — `NOT VALID` + `VALIDATE CONSTRAINT`

```sql
ALTER TABLE prompt_chain_steps ADD CONSTRAINT fk_chain_steps_prompt
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE RESTRICT NOT VALID;
-- Затем в отдельной миграции (после ручного scan'а на orphans):
ALTER TABLE prompt_chain_steps VALIDATE CONSTRAINT fk_chain_steps_prompt;
```

`NOT VALID` пропускает scan существующих rows (только проверяет новые), что позволяет применить FK мгновенно. `VALIDATE CONSTRAINT` потом запускает scan под `ShareUpdateExclusiveLock` (не блокирует writes).

### 6. `DROP COLUMN` — не удалять немедленно

Удаление колонки моментально только в metadata (PG помечает как dropped), но space реально освобождается только после `VACUUM FULL` (требует exclusive lock) или при следующем CLUSTER. Лучше:

1. Phase A: deprecate в коде (приложение перестаёт писать/читать). Wait 1-2 deploy cycles.
2. Phase B: `ALTER TABLE foo DROP COLUMN bar` в migration.
3. Phase C: `VACUUM (FULL, VERBOSE) foo` в плановом maintenance window.

---

## CHECK list для каждой миграции

Перед merge'ом PR'а с новой `*.up.sql`/`.down.sql`:

- [ ] Открывает ли миграция `AccessExclusiveLock` на горячую таблицу? Если да — split на безопасные шаги.
- [ ] `CREATE INDEX` на горячей таблице помечен `CONCURRENTLY` + `-- migrate:no-transaction`?
- [ ] `IF NOT EXISTS` / `IF EXISTS` для idempotency (повторный запуск)?
- [ ] `.down.sql` обратимо разворачивает изменения (или явно помечен как `-- IRREVERSIBLE: explanation`)?
- [ ] Проверена локально через `docker compose -f docker-compose.dev.yml up -d --build` (не забыть `--build`!).
- [ ] Если миграция меняет схему GORM-модели — модель в `internal/models/` и интерфейс репозитория обновлены в том же PR'е.
- [ ] Если новый CHECK/FK constraint — рассмотрено `NOT VALID` + отдельный VALIDATE.
- [ ] EXPLAIN ANALYZE-протестировано на realistic dataset (5K+ rows для горячих таблиц).

---

## Feature-флаги для опасных миграций

Если миграция требует длительного backfill'а (минуты-часы), feature-flag в env позволяет деплоить код заранее, потом включать функциональность отдельно:

```go
if cfg.Search.UseTSV {
    // новый FTS path
} else {
    // legacy ILIKE
}
```

Это защищает от `mig.deploy → blocking ALTER → rollback` сценария: код знает оба state'а схемы и работает в обоих.

См. примеры:
- `CHAINS_ENABLED` — Phase 16 dark-launch
- `ANALYTICS_EXPERIMENTAL_INSIGHTS` — Smart Insights kill-switch

---

## Реальные примеры из истории проекта

| Миграция | Проблема | Урок |
|---|---|---|
| `000051_prompts_fts.up.sql` | `ADD COLUMN ... GENERATED STORED` сразу | На 100K rows = 30-90s downtime — split на 4 шага (CR-15) |
| `000048_analytics_m8.up.sql` | дублирует `idx_prompts_content_trgm` из 000026 | `IF NOT EXISTS` спасает, но мусор в diff (MN-81) |
| `000059_subscriptions_status_paused` | `ALTER TABLE ... DROP/ADD CONSTRAINT` без `NOT VALID` | full table scan под `AccessExclusiveLock` (MN-80) |
| `000053_prompt_chains` (DEFERRABLE) | `DEFERRABLE INITIALLY DEFERRED` глобально | Лучше `INITIALLY IMMEDIATE` + явный `SET CONSTRAINTS DEFERRED` в transaction (MJ-40) |

---

## Ссылки

- [PostgreSQL CREATE INDEX docs](https://www.postgresql.org/docs/current/sql-createindex.html)
- [Generated Columns docs](https://www.postgresql.org/docs/current/ddl-generated-columns.html)
- [Brandur — Postgres Migration Patterns](https://brandur.org/postgres-atomicity)
- [Stripe — How we built rapid migrations](https://stripe.com/blog/online-migrations) — паттерн dual-write/dual-read.
