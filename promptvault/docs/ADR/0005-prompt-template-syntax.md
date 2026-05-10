# ADR 0005 — Custom `{{var}}` template parser, не Mustache/Handlebars

**Дата:** 2026-04-27
**Статус:** Accepted
**Phase:** 15 polish wave 2 (UX-аудит)

## Контекст

Промпты в PromptVault поддерживают переменные вида `{{name}}`. Парсер
живёт в двух местах:

- `frontend/src/lib/template/parse.ts` — для редактора, превью и
  `use-prompt-dialog`.
- `backend/internal/template/template.go` — для MCP `use_prompt`
  (рендер при выдаче промпта внешним клиентам через MCP).

Грамматика идентична на обеих сторонах (BNF задокументирован в
обоих файлах):

```
variable   ::= "{{" identifier "}}"
identifier ::= (letter | "_") (letter | digit | "_")*
letter     ::= any Unicode letter (\p{L})
```

При UX-аудите Phase 15 встал вопрос: оставить custom-парсер или
заменить на стандарт индустрии (Mustache, Handlebars, Jinja2,
LangChain f-string)?

## Решение

Оставляем custom-парсер. Расширения синтаксиса не вводим.

## Альтернативы (рассмотрены и отклонены)

- **A. Mustache (`mustache.js` + `cbroglie/mustache`).** Logic-less
  templating, де-факто стандарт LLM-промптов (LangChain, Anthropic
  prompt-cookbook). Минус: парсеры на двух языках имеют расхождения
  по whitespace handling и lambdas; реальных use-case для секций
  (`{{#var}} … {{/var}}`) и partials у нас нет.
- **B. Handlebars.** Мощнее Mustache (helpers, conditionals). Минусы:
  +30 KB gzipped в bundle frontend, нет официального паритетного Go-парсера,
  для простой substitution — over-engineering.
- **C. LangChain f-string (одинарные `{var}`).** Конфликт с уже принятой
  пользователями нотацией `{{}}`, требует миграции существующих промптов
  в БД, ломает MCP-клиентов (Cursor / Cline уже видят `{{}}`).
- **D. Jinja2-подобный (`{% if %}`, фильтры).** Опасно для пользовательского
  ввода — известный вектор RCE при рендере на сервере (см. LangChain docs,
  явное предупреждение про `template_format='jinja2'`).

## Последствия

**Плюсы:**

- Нулевая зависимость от сторонних либ → нулевой риск supply-chain.
- 56 unit-тестов уже покрывают edge-cases (Latin / Cyrillic / CJK
  identifiers, dedup, single-pass, regex-meta safety).
- Идентичная грамматика на обеих сторонах — гарантия, что фронт и
  MCP видят одни и те же переменные.
- Никаких операторов, которые могли бы выполнять код → невозможен RCE
  через шаблон (фундаментальное свойство «logic-less data-binding»).
- Implicit escape (`{{ name }}` с пробелом → литерал) — простое,
  понятное и не требует дополнительного синтаксиса вида `\{{`.

**Минусы:**

- Нет циклов / условных. Если завтра появится use-case вида
  «вставить список из N items» — придётся либо добавить функционал
  в парсер (fork), либо мигрировать на Mustache. Это явный trigger
  для пересмотра ADR.

## Когда пересматривать

Любой из триггеров:

1. Появился реальный use-case для секций `{{#each items}}` или
   условных `{{#if x}}` — тогда Mustache (вариант A) становится
   оправданным.
2. Backend начнёт рендерить шаблоны в HTML email или public-share
   странице (сейчас всё рендерится как plain text → `react-markdown`
   с `rehype-sanitize` на выводе).
3. MCP-клиенты начнут массово использовать literal `{{` в значениях
   переменных — implicit escape перестанет быть достаточным.

## Источники

- `frontend/src/lib/template/parse.ts:1-71` — frontend parser + BNF.
- `backend/internal/template/template.go:1-81` — backend parser + BNF.
- `frontend/src/lib/template/parse.test.ts` — 40+ test cases.
- `backend/internal/template/template_test.go` — 16 test cases.
- [LangChain Prompt Template format guide](https://docs.langchain.com/langsmith/prompt-template-format)
  — почему Mustache/`{{}}` идиоматичен для prompt-templates.
- [Latitude blog — Template Syntax Basics for LLM Prompts](https://latitude-blog.ghost.io/blog/template-syntax-basics-for-llm-prompts/).
- Phase 15 UX-audit план: `~/.claude/plans/indexed-weaving-otter.md` §2.3.
