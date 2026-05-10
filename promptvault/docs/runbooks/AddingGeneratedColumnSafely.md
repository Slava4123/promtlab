# Adding GENERATED Column Safely (3-step rollout pattern)

> **CR-15 готовый template.** Используется когда нужно добавить `GENERATED ALWAYS AS ... STORED` колонку на существующую большую таблицу (>10k rows). Делает миграцию zero-downtime.

## Зачем

`ALTER TABLE foo ADD COLUMN x GENERATED ALWAYS AS (...) STORED` берёт `AccessExclusiveLock` и переписывает ВСЮ таблицу — все API-запросы зависают. На 100K rows это 30-90 секунд downtime.

3-шаговый rollout разделяет: добавление nullable column (быстро, lock метаданных), backfill в background (chunks по 1000 row), переключение на `GENERATED ALWAYS` (быстро).

---

## Шаг 1 — миграция NNNN_add_column.up.sql

```sql
-- ALTER TABLE с ADD COLUMN nullable — берёт AccessExclusive только на метаданные.
-- Быстрый ALTER (миллисекунды) — не блокирует читателей/писателей.
ALTER TABLE prompts
  ADD COLUMN IF NOT EXISTS search_tsv_v2 tsvector;

-- Optional: trigger чтобы НОВЫЕ INSERT/UPDATE заполняли v2 колонку сразу.
-- Без trigger — backfill code должен сам обновлять row'ы которые юзер изменил
-- между ADD COLUMN и backfill completion.
CREATE OR REPLACE FUNCTION fill_search_tsv_v2() RETURNS TRIGGER AS $$
BEGIN
  NEW.search_tsv_v2 :=
    setweight(to_tsvector('russian_unaccent', coalesce(NEW.title, '')),   'A') ||
    setweight(to_tsvector('russian_unaccent', coalesce(NEW.content, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prompts_search_tsv_v2_fill
  BEFORE INSERT OR UPDATE OF title, content ON prompts
  FOR EACH ROW
  EXECUTE FUNCTION fill_search_tsv_v2();
```

## Шаг 1 — миграция NNNN_add_column.down.sql

```sql
DROP TRIGGER IF EXISTS prompts_search_tsv_v2_fill ON prompts;
DROP FUNCTION IF EXISTS fill_search_tsv_v2();
ALTER TABLE prompts DROP COLUMN IF EXISTS search_tsv_v2;
```

---

## Шаг 2 — Go backfill (cmd/backfill_search_tsv_v2/main.go)

```go
// cmd/backfill_search_tsv_v2 — fills prompts.search_tsv_v2 для существующих
// row'ов чанками по 1000. Безопасен для повторного запуска (idempotent —
// WHERE search_tsv_v2 IS NULL).
//
// Запуск: go run ./cmd/backfill_search_tsv_v2 --batch=1000 --pause-ms=100
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/postgres"
)

func main() {
	batch := flag.Int("batch", 1000, "row count per chunk")
	pauseMS := flag.Int("pause-ms", 100, "pause between chunks (ms)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config.load", "error", err)
		os.Exit(1)
	}
	db, err := postgres.NewGORM(cfg.Database)
	if err != nil {
		slog.Error("postgres.connect", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	totalUpdated := int64(0)
	for {
		select {
		case <-ctx.Done():
			slog.Info("backfill.cancelled", "total_updated", totalUpdated)
			return
		default:
		}

		// UPDATE с self-LIMIT через CTE — атомарный chunk.
		res := db.WithContext(ctx).Exec(`
			WITH next_batch AS (
				SELECT id FROM prompts
				WHERE search_tsv_v2 IS NULL
				LIMIT ?
				FOR UPDATE SKIP LOCKED
			)
			UPDATE prompts SET search_tsv_v2 =
				setweight(to_tsvector('russian_unaccent', coalesce(title, '')),   'A') ||
				setweight(to_tsvector('russian_unaccent', coalesce(content, '')), 'B')
			WHERE id IN (SELECT id FROM next_batch)
		`, *batch)
		if res.Error != nil {
			slog.Error("backfill.exec", "error", res.Error)
			os.Exit(1)
		}
		updated := res.RowsAffected
		totalUpdated += updated
		if updated == 0 {
			slog.Info("backfill.done", "total_updated", totalUpdated)
			return
		}
		slog.Info("backfill.chunk", "updated", updated, "total", totalUpdated)
		time.Sleep(time.Duration(*pauseMS) * time.Millisecond)
	}
}

var _ = fmt.Sprintf // suppress unused import warning if any
```

Запуск в prod:

```bash
docker exec -it promtlab-api ./backfill_search_tsv_v2 --batch=1000 --pause-ms=100
```

---

## Шаг 3 — миграция NNNN+1_promote_to_generated.up.sql

```sql
-- Удаляем trigger — больше не нужен, GENERATED займётся auto-fill.
DROP TRIGGER IF EXISTS prompts_search_tsv_v2_fill ON prompts;
DROP FUNCTION IF EXISTS fill_search_tsv_v2();

-- Переключаем на GENERATED ALWAYS. На этой стадии все row'ы уже заполнены
-- (backfill завершён) — ALTER занимает миллисекунды, не делает rewrite.
ALTER TABLE prompts
  ALTER COLUMN search_tsv_v2
  SET GENERATED ALWAYS AS (
    setweight(to_tsvector('russian_unaccent', coalesce(title, '')),   'A') ||
    setweight(to_tsvector('russian_unaccent', coalesce(content, '')), 'B')
  ) STORED;
```

## Шаг 3 — миграция NNNN+1_promote_to_generated.down.sql

```sql
-- Понижаем обратно до nullable + восстанавливаем trigger.
ALTER TABLE prompts ALTER COLUMN search_tsv_v2 DROP EXPRESSION;

CREATE OR REPLACE FUNCTION fill_search_tsv_v2() RETURNS TRIGGER AS $$
BEGIN
  NEW.search_tsv_v2 :=
    setweight(to_tsvector('russian_unaccent', coalesce(NEW.title, '')),   'A') ||
    setweight(to_tsvector('russian_unaccent', coalesce(NEW.content, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prompts_search_tsv_v2_fill
  BEFORE INSERT OR UPDATE OF title, content ON prompts
  FOR EACH ROW
  EXECUTE FUNCTION fill_search_tsv_v2();
```

---

## Шаг 4 — индекс CONCURRENTLY (если нужен)

```sql
-- migrate:no-transaction — CREATE INDEX CONCURRENTLY запрещён в transaction.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_search_tsv_v2
  ON prompts USING GIN (search_tsv_v2)
  WHERE deleted_at IS NULL;
```

---

## Чек-лист применения

1. [ ] Шаг 1 миграция применена в prod (`migrate up`).
2. [ ] Запущен `backfill_search_tsv_v2` — дождаться `total_updated` логи и `done`.
3. [ ] Verify: `SELECT COUNT(*) FROM prompts WHERE search_tsv_v2 IS NULL` → 0.
4. [ ] Шаг 3 миграция применена.
5. [ ] Шаг 4 (индекс) применён вручную (потому что `no-transaction` директива; golang-migrate v4 поддерживает её через `-- migrate:no-transaction` first-line comment).

Total downtime: 0 (все ALTER'ы быстрые, backfill в фоне, INDEX CONCURRENTLY не блокирует writes).
