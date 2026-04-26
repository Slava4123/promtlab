# ADR 0002 — PostgreSQL FTS вместо embeddings для search

**Дата:** 2026-04-26
**Статус:** Accepted
**Phase:** 15 W3

## Контекст

FEATURES.md #21 заявляет "Semantic search понимает синонимы, контекст,
намерение". Текущая реализация `prompt_repo.SearchByQuery` строит
`%query%` и фильтрует через `ILIKE` — substring-only:

- `q = "резюмировать"` НЕ находит промпт со словом "резюме";
- `q = "writing"` НЕ находит промпт "wrote", "writes".

Для production-grade поиска нужно как минимум **stem-aware** retrieval.
Изначально план предполагал self-hosted Python sidecar с
`multilingual-e5-small` (118M params, ~500MB RAM) + pgvector — но это
добавляет новый сервис в docker-compose, Node-зависимость для build CI
mjml-style, и Python runtime в production.

## Решение

**Использовать PostgreSQL Full-Text Search (`tsvector` + GIN индекс)
вместо embeddings.** Без новых сервисов, всё в существующей PG18.

Конкретно:

1. **`prompts.search_tsv`** — `GENERATED ALWAYS AS STORED tsvector`,
   конкатенация `setweight('A')` для title и `setweight('B')` для content,
   через два text search config (`russian_unaccent` + `english`).
2. **`russian_unaccent`** — копия `russian` с дополнительным `unaccent`
   filter для ё/е симметрии. Создаётся в миграции через `DO $$ ... $$`.
3. **GIN индекс** `idx_prompts_search_tsv WHERE deleted_at IS NULL`,
   `STATISTICS 1000` для лучших оценок селективности.
4. **`websearch_to_tsquery`** для user input — устойчив к спецсимволам,
   обрабатывает кавычки/OR/минусы как Google. `to_tsquery` отвергнут —
   падает на любом неправильном вводе.
5. **Гибрид FTS + pg_trgm** — `WHERE search_tsv @@ tsq OR title %% q`.
   Trigram-similarity ловит опечатки, FTS — словоформы.
6. **Weighted score** — `ts_rank_cd × 0.7 + similarity × 0.3`,
   двухфазный SELECT (raw для score+ids → GORM Find с CASE-ordering).

## Альтернативы рассмотрены

- **A. Self-hosted Python sidecar (e5-small) + pgvector.** Полная семантика
  ("составь CV" ↔ "напиши резюме"). Минус: новый сервис ~500MB RAM,
  сложнее ops, build pipeline зависит от HuggingFace cache. Для масштаба
  ПромтЛаба (≤10k промптов на Max-юзера) — overkill.
- **B. YandexGPT/GigaChat Embeddings (PaaS).** Доступны из РФ. Минус:
  внешняя зависимость от российских облаков, оплата по tokens, нарушает
  self-hosted принцип.
- **C. Внешний search engine (Meilisearch / Typesense / Elasticsearch).**
  Лучшее качество и поиск по синонимам. Минус: новый сервис, ops overhead,
  нужно поддерживать sync с PG. Для масштаба избыточно.
- **D. Per-row `lang` column + dynamic `regconfig`.** Точнее, но требует
  детекции языка на write-pass. ПромтЛаб промпты часто bilingual
  (русский ↔ английский в одном промпте) — детекция фейлится.

## Последствия

- "резюмировать" находит "резюме", "writing" находит "wrote".
- "ёжик" находит "ежик" (через unaccent в russian_unaccent config).
- Опечатки типа "резмюе" → "резюме" работают через pg_trgm `%%` оператор.
- FEATURES.md #21 заявленный "понимает синонимы" — НЕ покрывается
  (требует embeddings). "Составь CV" не найдёт "напиши резюме". Это
  acceptable trade-off для self-hosted без ML-инфраструктуры.
- При недоступности расширения `unaccent` — миграция использует
  `IF NOT EXISTS` для config, fallback graceful (russian stemmer
  работает без unaccent, ё/е симметрия теряется, остальное работает).

## Источники

- `backend/internal/infrastructure/postgres/migrations/000051_prompts_fts.up.sql`
- `backend/internal/infrastructure/postgres/repository/prompt_repo.go:SearchByQuery`
- [PostgreSQL 18 §12 Full Text Search](https://www.postgresql.org/docs/18/textsearch.html)
- Commit: `feat(search): PostgreSQL FTS вместо ILIKE для prompts (RU+EN stemming)`
