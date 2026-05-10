-- MJ-26: Collection.Color size:20→7. До этого fix'а MCP-tool create_collection
-- мог сохранять до 20 символов в color (например `red; }<style/script>`),
-- размер давал место для XSS payload'а. HTTP-валидатор стоял только на handler'е,
-- MCP обходил его. Теперь:
--   1. Type HexColor + BeforeSave validation в Go-модели — independent защита.
--   2. Этот ALTER ужимает column до VARCHAR(7) — БД-уровень enforcement.
--
-- Backfill: невалидные значения (длина > 7 или не подходящие под #RRGGBB)
-- заменяются на дефолт #8b5cf6 (тот же что в models/collection.go).
-- Без backfill ALTER упадёт на существующих row'ах с Length(color) > 7.

UPDATE collections
SET color = '#8b5cf6'
WHERE color IS NOT NULL
  AND (length(color) > 7 OR color !~ '^#[0-9a-fA-F]{6}$');

ALTER TABLE collections
ALTER COLUMN color TYPE VARCHAR(7);
