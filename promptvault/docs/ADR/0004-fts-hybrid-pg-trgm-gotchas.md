# ADR 0004 — FTS hybrid с pg_trgm: подводные камни

**Дата:** 2026-04-27
**Статус:** Accepted (post-mortem)
**Phase:** 15 W3.9 + QA findings

## Контекст

В Phase 15 W3.9 (PostgreSQL FTS вместо ILIKE) реализовали гибридный поиск:

```sql
WHERE search_tsv @@ websearch_to_tsquery('russian_unaccent', ?)
   OR title %% ?     -- pg_trgm fuzzy для опечаток
```

Юнит-тесты Go (с моками) и self code review через 3 агентов прошли —
но **QA UI testing нашёл production crash**: `/api/search` возвращает
HTTP 500 при включённом pg_trgm.

```
ERROR: operator does not exist: character varying %% unknown (SQLSTATE 42883)
```

## Корневая причина

Несколько gotchas наложились друг на друга:

1. **`%%` ≠ `%`.** Я ошибочно использовал двойной процент, думая что
   pg_trgm имеет такой оператор. На самом деле один `%` — similarity threshold,
   `%>` — word_similarity_op, `<%>` — strict word similarity. **`%%` не
   существует** как pg_trgm оператор.
2. **`varchar` vs `text`.** pg_trgm операторы и `similarity()` определены
   только для типа `text`. `prompts.title varchar(300)` без явного cast
   вызывает `operator does not exist: character varying % unknown`.
3. **Параметр `?` без cast тоже падает.** Даже после `title::text %`,
   placeholder `?` парсится как `unknown` тип → новая ошибка
   `operator does not exist: text % unknown`.
4. **GORM `Raw()` и `%` в SQL string.** Один `%` может быть ambiguously
   интерпретирован GORM как format-string escape. Чтобы избежать
   двусмысленности — использовать функцию вместо оператора.

## Решение

Использовать **функцию `similarity(a::text, b::text) > threshold`**
вместо оператора `%`:

```sql
-- БЫЛО (broken):
WHERE search_tsv @@ websearch_to_tsquery('russian_unaccent', ?)
   OR title %% ?

-- СТАЛО (works):
WHERE search_tsv @@ websearch_to_tsquery('russian_unaccent', ?)
   OR similarity(title::text, ?::text) > 0.3
```

Преимущества функционального синтаксиса:
- Явный type cast обеих сторон;
- Threshold (0.3) виден в коде (не глобальный `pg_trgm.similarity_threshold`);
- Нет ambiguity с GORM `%` escape;
- Единая стилистика с `similarity()` в ORDER BY.

## Альтернативы рассмотрены

- **A. Изменить тип `title` на `text`.** Минус: схема migration на проде
  (ALTER TABLE ... ALTER COLUMN TYPE text), не оправдано ради FTS-edge-case.
- **B. Использовать `<<%`/`%>>` (word similarity)** — ближе к pg_trgm
  best-practices для autocomplete. Минус: меньше выборка, может пропустить
  valid typos. Для resume-style queries `similarity()` универсальнее.
- **C. Убрать pg_trgm ветку вообще** — тогда теряется fuzzy match для
  опечаток. Self-degrade при `trgmAvailable=false` уже работает.

## Уроки на будущее

1. **Юнит-тесты с моками не ловят SQL-уровневые регрессии.**
   `prompt_repo.SearchByQuery` тестировался через mock'и AnalyticsRepository
   и т.д., реальный SQL не запускался. **Нужен integration-тест с
   testcontainers + `CREATE EXTENSION pg_trgm + unaccent + russian_unaccent`**.
2. **Self code review агенты предупреждали о hard-fail на PG без pg_trgm**
   — но реальность оказалась хуже: **падает И с pg_trgm**, потому что
   проблема в другом (varchar vs text). Recommendation от агента
   "пробросить probe в prompt_repo" был верен, но недостаточен.
3. **GORM `Raw()` — поверхность для ambiguity** с спец-символами SQL.
   Когда возможно — использовать billable-builder API, либо функции
   вместо операторов.

## TODO для CI (не блокер этого PR)

- Добавить integration-тест `prompt_repo_search_test.go` через testcontainers,
  который реально вызывает `SearchByQuery` с pg_trgm на свежей PG18:
  - happy path: stemming находит словоформы;
  - typo path: `similarity > 0.3` находит опечатки в title;
  - graceful degrade: при `trgmAvailable=false` запрос проходит без ошибок.

## Источники

- `backend/internal/infrastructure/postgres/repository/prompt_repo.go:163-186`
- [PostgreSQL pg_trgm operators](https://www.postgresql.org/docs/18/pgtrgm.html#PGTRGM-OP-TABLE)
- Commit: `fix(search): pg_trgm требует cast varchar→text + similarity() вместо оператора %`
- QA report Phase 15 (Bug #1, Critical)
