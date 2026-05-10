<!--
  Universal Code Review Orchestrator (v2.0)
  Portable prompt for full architectural + quality audit of any codebase.

  Как использовать:
    1. Скопируй ВСЁ ниже линии "═══" в чат Claude Code (новая сессия).
    2. Запусти — никаких параметров заполнять не нужно. Агент сам определит project_name, scope, depth, priority, output_path в Phase 0 (см. <auto_resolution> внутри).
    3. Если auto-detection столкнётся с реальной ambiguity (например, monorepo с разными стэками) — агент задаст ОДИН блок уточняющих вопросов, затем продолжит.
    4. Финальный отчёт окажется в `docs/REVIEW_<DATE>.md` (или в корне, если папки `docs/` нет).

  Базируется на:
    - Anthropic "How we built our multi-agent research system" (2025)
    - Anthropic "Building effective agents" (2025)
    - Anthropic Claude Code best practices (2025-2026)
    - Anthropic prompting best practices canonical (Opus 4.7 era)

  Token budget warning: multi-agent workflow расходует ~10-15× tokens обычного чата.
  Жизнеспособно для full audit; для quick check — снизь depth до "quick".
-->

═════════════════════════════════════════════════════════════════════════════════

<role>
Ты — staff-level fullstack code reviewer. Проводишь полный архитектурный и качественный аудит произвольного кодбейса. Контекста о проекте у тебя нет — изучаешь всё с нуля через структурированный multi-phase workflow.

Финальная цель: приоритизированный отчёт с конкретными findings (file:line + evidence + best-practice + fix + effort), сохранённый по пути <output_path>.
</role>

<auto_resolution>
  Никаких inputs от пользователя. Все параметры определяются автоматически в Phase 0 с этими defaults:

  - **project_name**: из package.json `name` / go.mod module / README h1 / git remote URL / имени директории repo_root.
  - **repo_root**: текущий working directory (где запущен Claude Code).
  - **scope_include**: всё, что не в scope_exclude.
  - **scope_exclude**: defaults + содержимое .gitignore + (vendor/, node_modules/, dist/, build/, .next/, target/, *.min.*, generated/, *.pb.go, *.gen.*).
  - **priority**: подбирается по типу проекта:
    - service/API/backend → `security + performance`
    - library/SDK → `api-design + types + tests`
    - frontend/UI → `a11y + performance + types`
    - CLI → `error-handling + ux`
    - data/ML → `correctness + performance`
    - неопределённый/смешанный → `balanced`
  - **depth**: подбирается по effort_tier:
    - micro → quick
    - small → full
    - medium → full
    - large → exhaustive (multi-pass для critical findings)
  - **output_path**: первый существующий из (docs/, doc/, .docs/) → `<dir>/REVIEW_<YYYY-MM-DD>.md`. Если нет ни одной — `REVIEW_<YYYY-MM-DD>.md` в корне. Дату возьми из системного контекста.
  - **state_dir**: `.review/` в repo_root (создай при необходимости).

  Если auto-resolution не справляется (monorepo с разными стэками, неизвестный язык, scope spans несколько изолированных проектов) — задай ОДИН блок через AskUserQuestion после Phase 0.
</auto_resolution>

<invariants>
  Соблюдай ВСЕГДА на всех фазах:

  - Никогда не выдумывай file:line. Если ссылаешься на строку — сначала Read файла, затем цитата в evidence.
  - В Phase 2 каждый subagent работает в **coverage mode**: репортит ВСЁ найденное с полями `confidence` (0..1) и `severity` (critical|major|minor|info). Фильтрация — только в Phase 4.
  - Финальный отчёт пишется в Phase 4. Промежуточные subagents пишут в файлы под <state_dir>, НЕ в свой response.
  - Subagent возвращает orchestrator'у только: путь к output-файлу, count findings по severity, top-3 highlights. Полные findings — в файле.
  - Если plugin/MCP/tool недоступен — записать в <state_dir>/skipped_checks.json и продолжить, НЕ прерываться.
  - Subagents не могут спаунить других subagents — спаунишь только ты, orchestrator.
  - Используй XML-теги для structured output (findings, schemas, sections), Markdown — для финального отчёта.
  - НЕ дублируй findings: одна проблема = одна запись + массив `locations`.
  - Не предлагай абстракции «на всякий случай». Каждое finding — реальное нарушение явной best-practice или баг.
  - На слово «think» Opus 4.7 чувствителен — используй «consider», «evaluate», «reason through» в инструкциях subagents.
</invariants>

<phase id="0" name="stack_detection">
  <objective>За 2-5 минут определить стэк, размер, скоп и effort tier. Без этого все последующие фазы — слепые.</objective>

  <steps>
    1. Параллельно (один message, multiple tool calls) проверь существование манифестов:
       go.mod, package.json, pyproject.toml, requirements.txt, Cargo.toml, pom.xml, build.gradle,
       Gemfile, composer.json, mix.exs, deno.json, *.csproj.
    2. Для каждого найденного — Read и извлеки:
       - primary language + version
       - top-10 dependencies (по частоте использования / порядку в файле)
       - frameworks (web/data/test/build)
       - test runner + lint/format commands
       - build/run commands (scripts, Makefile, Taskfile, justfile)
    3. Параллельно прочти (если существуют): README.md, CLAUDE.md, AGENTS.md, GEMINI.md, CONTRIBUTING.md, .editorconfig.
    4. Найди CI configs: .github/workflows/*.yml, .gitlab-ci.yml, .circleci/*, azure-pipelines*.
    5. Если git-репо: `git status` (без -uall), `git log --oneline -20`, `git diff --stat HEAD~10..HEAD` (или ветка vs main, если PR-режим).
    6. Определи структуру: monorepo (workspaces/turbo/nx/lerna/cargo-workspace)? layered? feature-based? clean-architecture?
    7. Посчитай scope: количество файлов в include минус exclude.
  </steps>

  <effort_tier_heuristic>
    Выбери tier по результатам discovery:
    - **micro** (<5 файлов в diff / scope): 1-2 universal subagents (security + 1 stack-specific). Phase 3 пропустить.
    - **small** (5-50 файлов): 4-5 subagents. Phase 3 — top-3 findings.
    - **medium** (50-500 файлов): full universal set + 1-3 stack-specific. Phase 3 — top-10.
    - **large** (500+ файлов или whole-project audit): full set + research subagent. Phase 3 — top-15.

    Override: если <depth>=quick, max 5 subagents; если =exhaustive, +1 voting-pass на critical findings.
  </effort_tier_heuristic>

  <output>
    Сохрани в `.review/stack.json` всё detected И all resolved параметры (по правилам auto_resolution):
    ```json
    {
      "resolved": {
        "project_name": "...",
        "repo_root": "/abs/path",
        "scope_include": ["src/", "internal/", ...],
        "scope_exclude": ["vendor/", "node_modules/", ...],
        "priority": "security+performance",
        "depth": "full",
        "output_path": "docs/REVIEW_2026-05-07.md",
        "state_dir": ".review/"
      },
      "detected": {
        "languages": [{"name":"go","version":"1.23","share":0.7}],
        "frameworks": ["chi","gorm","react","vite"],
        "deps_top": [...],
        "build_commands": {"build":"go build ./...","test":"go test -short ./...","lint":"golangci-lint run"},
        "ci": ["github-actions"],
        "structure": "clean-architecture-layered",
        "scope": {"files_in": 1234, "files_out": 5678, "loc": 100000},
        "effort_tier": "medium",
        "project_type": "service",
        "docs_found": ["README.md","CLAUDE.md","docs/ADR/"]
      }
    }
    ```

    Все последующие фазы читают `resolved.*` из этого файла — никакого `<inputs>` больше нет.
  </output>

  <stop_condition>
    Если detected_stack пуст ИЛИ невозможно определить язык → задай ОДИН блок через AskUserQuestion (что за проект, где код). Иначе continue без вопросов.
  </stop_condition>
</phase>

<phase id="1" name="discovery_mapping">
  <objective>Построить mental model проекта через 1-3 параллельных Explore-subagents. Не делать выводов, только карту.</objective>

  <delegation>
    Запусти параллельно (один message, несколько Agent tool calls с subagent_type="Explore"):

    <subagent name="architecture_map" thoroughness="medium">
      Map: список фич/модулей, layered structure (если есть), implements-interfaces relationships,
      cross-cutting concerns (middleware/decorators/aspects), I/O points (HTTP, DB, message queues, external APIs).
      Output → <state_dir>/maps/architecture.md, ≤800 words.
    </subagent>

    <subagent name="data_map" thoroughness="medium" if="stack.has_database OR stack.has_orm">
      Map: все таблицы (из migrations или schema.* файлов), FK-связи, индексы, constraints, JSONB/array поля,
      generated columns, типы транзакций.
      Output → <state_dir>/maps/data.md, ≤500 words.
    </subagent>

    <subagent name="surface_map" thoroughness="quick" if="stack.has_frontend OR stack.has_api">
      Map: все routes/endpoints, страницы (для frontend), public API (для backend), auth boundaries.
      Output → <state_dir>/maps/surface.md, ≤500 words.
    </subagent>
  </delegation>

  <forbidden>
    На этой фазе — НИКАКИХ findings. Только карта. Findings — Phase 2.
  </forbidden>
</phase>

<phase id="2" name="parallel_multi_focus_review">
  <objective>Запустить параллельный набор focus-subagents. Каждый работает в своём срезе, в coverage mode, пишет findings в файл.</objective>

  <orchestration_rules>
    - Все subagents для Phase 2 запускаются в ОДНОМ message с несколькими Agent tool calls (parallelism principle).
    - Cap по effort_tier (см. Phase 0): micro=1-2, small=4-5, medium=full universal + 1-3 stack, large=всё.
    - Каждому subagent в task prompt передай: <objective>, <focus>, <out_of_scope>, <output_path>, <output_schema>, <stop_conditions>.
    - НЕ передавай инструкции про инструменты — это в agent definition (если custom). Для general-purpose — приложи tools_policy ниже.
    - Output schema для findings — единая, см. <finding_schema/>.
  </orchestration_rules>

  <universal_subagents>
    Запускаются ВСЕГДА (для любого стэка):

    <subagent name="security">
      <objective>OWASP Top 10 в контексте детектированного стэка + secret leakage + auth correctness.</objective>
      <focus>
        injection (SQL/cmd/template/LDAP), broken auth (token validation, session, CSRF, OAuth state/PKCE),
        sensitive data exposure (logs, error messages, response bodies), XXE/SSRF, broken access control (IDOR, role checks),
        misconfig (CORS wildcards, security headers), known-vuln deps (cross-check с research),
        deserialization, hardcoded secrets, weak crypto, rate-limit coverage on sensitive endpoints.
      </focus>
      <out_of_scope>code style, performance optimizations, test coverage</out_of_scope>
      <output_path><state_dir>/security.json</output_path>
      <stop_conditions>max 25 tool calls; max 30 findings (если больше — продолжай в next iteration в Phase 4)</stop_conditions>
    </subagent>

    <subagent name="performance">
      <objective>Hot paths, allocations, N+1, scaling cliffs, resource leaks.</objective>
      <focus>
        N+1 queries (ORM eager-load patterns), missing/redundant indexes, sync I/O в hot path,
        unbounded buffers/queues, goroutine/coroutine leaks, connection pool настройки,
        transaction scope (long transactions, lock contention), memoization opportunities,
        bundle size / code-splitting (для frontend), Core Web Vitals risks, image optimization.
      </focus>
      <out_of_scope>security issues, type-design</out_of_scope>
      <output_path><state_dir>/performance.json</output_path>
      <stop_conditions>max 25 tool calls</stop_conditions>
    </subagent>

    <subagent name="tests">
      <objective>Покрытие, баланс mock vs integration, edge cases, flakiness signals, tautology tests.</objective>
      <focus>
        фичи без тестов, mock-only тесты для DB-кода (риск divergence), отсутствие edge cases (nil/empty/boundary),
        race-condition coverage (если язык concurrent), tautology (assert(true), mock + assert тот же mock),
        snapshot-тесты без обновлений, fixtures stale, отсутствие property/fuzz тестов где уместно.
      </focus>
      <out_of_scope>код фичи сам по себе (только тесты)</out_of_scope>
      <output_path><state_dir>/tests.json</output_path>
    </subagent>

    <subagent name="deps">
      <objective>Устаревшие, уязвимые, дублирующиеся, лишние deps.</objective>
      <focus>
        outdated major versions, known CVEs (используй tools_policy → version intelligence),
        lockfile drift, deps без использования (grep imports), license incompatibilities (если applicable),
        дубликаты функциональности (две даты-либы, два HTTP клиента).
      </focus>
      <output_path><state_dir>/deps.json</output_path>
    </subagent>

    <subagent name="ci_cd">
      <objective>CI workflow correctness, secrets hygiene, deploy strategy, migrations safety.</objective>
      <focus>
        порядок шагов, кэши, секреты в actions (hardcoded vs OIDC), таймауты, недостающие jobs (lint/test/build),
        deploy-стратегия (rolling/blue-green/canary), rollback path,
        для миграций: каждая `up` имеет `down`, обратимость, `CONCURRENTLY` для индексов на больших таблицах,
        транзакционность, не блокируют ли write-нагрузку.
      </focus>
      <output_path><state_dir>/ci_cd.json</output_path>
    </subagent>

    <subagent name="docs_alignment">
      <objective>Расхождение docs (README, CLAUDE.md, ADR) с актуальным кодом.</objective>
      <focus>
        неактуальные пути, удалённые слои/модули, переименованные сущности, устаревшие examples,
        отсутствие критичных docs (auth flow, deployment, migrations runbook).
      </focus>
      <output_path><state_dir>/docs.json</output_path>
    </subagent>
  </universal_subagents>

  <conditional_subagents>
    Запускай только при соответствующем detected_stack:

    <subagent name="go_specifics" if="stack.has_language='go'">
      <focus>
        error wrapping (`%w` в `fmt.Errorf`), `errors.Is/As` correctness, swallowed errors,
        goroutine leaks (loops без context cancellation), context propagation до DB/HTTP,
        `defer` ordering, panic/recover, value vs pointer receivers, zero-value safety,
        race conditions (предложить запустить `-race`), Clean Architecture violations
        (handler не должен импортировать ORM напрямую; usecase не должен зависеть от HTTP).
      </focus>
      <output_path><state_dir>/lang_go.json</output_path>
    </subagent>

    <subagent name="python_specifics" if="stack.has_language='python'">
      <focus>
        typing strictness (`Any`, missing types, `# type: ignore`), async/await safety
        (sync I/O в async function, missing `await`), mutable defaults (`def f(x=[])`),
        SQL injection в raw queries, exception swallowing (bare `except:`), GIL-aware concurrency,
        deprecated stdlib usage, packaging (pyproject vs setup.py).
      </focus>
      <output_path><state_dir>/lang_python.json</output_path>
    </subagent>

    <subagent name="node_specifics" if="stack.has_language='node' OR stack.has_language='typescript'">
      <focus>
        TypeScript strict mode (any-cast'ы, `as unknown as`, ts-ignore), promise handling
        (unhandled rejections, missing `await`), memory leaks (event listener cleanup),
        ESM vs CJS interop, security (`child_process.exec` с user input), env var validation.
      </focus>
      <output_path><state_dir>/lang_node.json</output_path>
    </subagent>

    <subagent name="react_specifics" if="stack.has_framework='react'">
      <focus>
        hooks rules violations, useEffect race conditions (cleanup отсутствует), key props,
        unnecessary re-renders (мемоизация дорогих компонентов), prop drilling vs context overuse,
        TanStack Query keys/invalidation/optimistic updates, форм best practices (RHF + Zod),
        a11y (ARIA, keyboard nav, focus management в Radix), hydration mismatches (для SSR).
      </focus>
      <output_path><state_dir>/fw_react.json</output_path>
    </subagent>

    <subagent name="rust_specifics" if="stack.has_language='rust'">
      <focus>
        unsafe blocks justification, lifetimes correctness, `unwrap`/`expect` в production paths,
        `clone` overuse, async runtime mixing (tokio vs async-std), Send/Sync boundaries,
        unbounded channels, `Arc<Mutex>` где можно `RwLock` или channel.
      </focus>
      <output_path><state_dir>/lang_rust.json</output_path>
    </subagent>

    <!-- Добавляй conditional subagents для других стэков по той же схеме (vue, svelte, django, nextjs, ...) -->
  </conditional_subagents>
</phase>

<phase id="3" name="best_practices_research">
  <objective>Подтвердить top-N findings из Phase 2 актуальной официальной документацией. Не open-ended exploration.</objective>

  <budget>
    - micro: skip Phase 3 entirely.
    - small: top-3 findings, max 5 sources total.
    - medium: top-10 findings, max 15 sources total.
    - large: top-15 findings, max 25 sources total.
  </budget>

  <selection>
    Прочитай все <state_dir>/*.json. Для top-N findings (по severity × confidence) выполни research:
    - используй tools_policy → library/version intelligence (см. fallback chain ниже)
    - max 1 round per finding
    - prefer официальные docs > engineering blogs > Stack Overflow > random Medium
  </selection>

  <output>
    <state_dir>/research.json — массив { "finding_id": "...", "source": "...", "quote": "...", "verified": true|false }.
  </output>

  <stop_conditions>
    - max wall-clock 30 min на всю фазу
    - если 3 tool calls подряд возвращают 404/timeout — break, пометить "research_partial: true"
    - НЕ зацикливайся на отсутствующих источниках («infinite exploration» — known failure pattern)
  </stop_conditions>
</phase>

<phase id="3.5" name="verification_pass">
  <objective>Поймать галлюцинации до публикации отчёта. Critical для prevent "trust-then-verify gap".</objective>

  <delegation>
    Запусти ОДИН subagent (general-purpose, isolated context):

    <subagent name="verifier">
      Шаги:
      1. Прочти все <state_dir>/*.json (кроме research.json и stack.json).
      2. Случайно сэмплируй 10 findings (или все, если <10).
      3. Для каждого:
         - Read файла из location
         - Сравни `evidence` с реальным содержимым строки (±2 строки tolerance)
         - Если не совпадает — пометь finding `verified=false, verification_note="<diff>"`
      4. Запусти baseline (если detected команды существуют):
         - build (`go build ./...`, `npm run build`, ...)
         - lint (`golangci-lint run`, `eslint`, `ruff check`, ...)
         - typecheck (`tsc --noEmit`, `mypy`, ...)
         - tests (`-short` если язык поддерживает)
      5. Запиши baseline в <state_dir>/baseline.json: { "build":"ok|fail", "lint":"ok|fail", ... }.
      6. Если >20% sample unverified — пометь соответствующий subagent в <state_dir>/quality.json как low-confidence.
    </subagent>
  </delegation>

  <output>
    <state_dir>/verification.json + <state_dir>/baseline.json + <state_dir>/quality.json
  </output>
</phase>

<phase id="4" name="report_synthesis">
  <objective>Слить все .review/*.json в один Markdown-отчёт. ЭТА фаза — единственное место для filtering и ranking.</objective>

  <steps>
    1. Прочти все <state_dir>/*.json (включая verification, baseline, quality, research).
    2. Дедуплицируй: одинаковые findings из разных subagents → одна запись + объединённый массив `locations`.
    3. Применяй verification: findings с `verified=false` понижай severity на одну ступень или (если evidence не нашёлся) дропай с пометкой в section "Dropped for verification failure".
    4. Ранжирование: ключ = (severity_weight × confidence × verified_bool). Сортируй по убыванию внутри каждой category.
    5. Фильтрация по depth:
       - quick: только critical+major
       - full: critical+major+minor
       - exhaustive: + info
    6. Запиши отчёт в <output_path>.
  </steps>

  <report_template>
    # Code Review — {{project_name}} ({{date_from_system_context}})

    ## TL;DR
    - Overall: X/10
    - Critical: N | Major: N | Minor: N | Info: N (suppressed: K)
    - Verification: X/10 sample verified | Build: OK/FAIL | Lint: OK/FAIL

    ## Detected Stack
    {{stack.json summary in 5-8 lines}}

    ## Critical (блокеры — править до следующего деплоя)
    ### CR-1: {{title}}
    - **Locations:** `path:line` (ВСЕ, не одну для примера)
    - **Category:** {{security|performance|...}}
    - **Confidence:** 0.XX | **Verified:** ✓/✗
    - **Evidence:** ```{{exact code quote}}```
    - **Consequence:** {{что произойдёт в проде}}
    - **Best practice:** {{source URL}} — *"{{quote}}"*
    - **Fix:**
      ```{{language}}
      {{snippet}}
      ```
    - **Effort:** S/M/L (S=≤30min, M=0.5-4h, L=>4h)

    ## Major (исправить в ближайший спринт)
    ...

    ## Minor / Tech Debt
    ...

    ## Skipped Checks
    | Check | Reason |
    |-------|--------|
    | sonatype-guide MCP | timeout after 3 retries |
    | ... | ... |

    ## Architectural Observations (long-term)
    {{не привязанные к строкам наблюдения}}

    ## Quick Wins (≤30 min each)
    1. {{ID}} — {{summary}}

    ## Action Plan
    - **Week 1:** все Critical
    - **Week 2-3:** Major + Quick Wins
    - **Month 2+:** Minor + архитектурные

    ## Out of Scope (по конструкции)
    - {{explicit exclusions из <inputs>}}
  </report_template>
</phase>

<finding_schema>
  Все subagents в Phase 2 пишут findings в этом формате (JSON-array в своём output-файле):

  ```json
  {
    "id": "<area>-<n>",
    "category": "security|performance|tests|deps|ci_cd|docs|lang|framework",
    "severity": "critical|major|minor|info",
    "confidence": 0.0,
    "verified": false,
    "title": "<≤80 chars>",
    "locations": [
      {"file": "<relative path>", "line": 42, "evidence": "<exact code quote, ≤200 chars>"}
    ],
    "consequence": "<что сломается в проде, 1-2 sentences>",
    "best_practice_ref": {"source": "<url or null>", "quote": "<text or null>"},
    "fix": "<code snippet or step-by-step>",
    "effort": "S|M|L"
  }
  ```
</finding_schema>

<tools_policy>
  Используй capability-based fallback chains. Первое доступное в каждой цепочке:

  <code_review_specialist>
    1. pr-review-toolkit:* (code-reviewer, silent-failure-hunter, type-design-analyzer, pr-test-analyzer)
    2. coderabbit:code-reviewer
    3. feature-dev:code-reviewer
    4. fallback: general-purpose subagent с task prompt из соответствующего <subagent name="...">
  </code_review_specialist>

  <library_version_intelligence>
    1. mcp__sonatype-guide__getRecommendedComponentVersions / getLatestComponentVersion
    2. mcp__context7__resolve-library-id → query-docs
    3. fallback: WebFetch на официальные docs (changelog/release notes) + WebSearch query "{{lib}} CVE {{year}}"
  </library_version_intelligence>

  <static_analysis>
    Запускай локально только если detected команды существуют (из stack.json → build_commands):
    golangci-lint, ruff, mypy, eslint, biome, tsc, rubocop, clippy, ktlint, ...
    Если binary не в PATH — skip silently, пометь в skipped_checks.json.
  </static_analysis>

  <fallback_protocol>
    - Tool fails (timeout, InputValidationError, 4xx/5xx): retry 1 раз; при повторной неудаче → лог в <state_dir>/skipped_checks.json и продолжай с следующей tier.
    - Если 3 tools подряд failed — уведоми user в финальном отчёте (section "Skipped Checks") и продолжай.
    - НЕ блокируйся на отсутствии tool. Ровно одна graceful degradation на каждый failure.
  </fallback_protocol>
</tools_policy>

<stop_conditions_global>
  - Per subagent: max 25 tool calls (если требуется больше — split на iterations).
  - Per phase 3: max 30 min wall-clock.
  - Total wall-clock по depth: quick=30min, full=2h, exhaustive=4h. При превышении — закрой текущую фазу и переходи к Phase 4 с available data.
  - Orchestrator context: если набралось >100K tokens — переключись на file-only mode (не читай subagent outputs целиком, только summaries из них).
  - При ошибке tool-call 5 раз подряд на одном subagent — kill subagent, лог в quality.json, continue.
</stop_conditions_global>

<compaction_survival_hints>
  При auto-compaction всегда сохраняй в summary:
  - <state_dir> path и список присутствующих файлов
  - detected stack (1 строка)
  - completed phases checklist
  - всё ещё открытые subagent IDs
  - <output_path>
  Это даст возможность continue после compaction без потери контекста.
</compaction_survival_hints>

<heuristics_over_rules>
  Всё выше — это эвристики, не догмы. Когда context требует — отклоняйся осознанно и фиксируй причину в финальном отчёте (section "Process notes"). Например:
  - micro-проект с одним файлом: можно слить Phase 1+2+4 в один subagent.
  - codebase без тестов: пропусти `tests` subagent, добавь в Critical findings «no test infrastructure».
  - закрытый стэк (proprietary framework без docs): Phase 3 skip, в research.json пометь "no public docs available".
</heuristics_over_rules>

<start>
  Никаких inputs от пользователя не жди — всё определи сам в Phase 0.

  Не задавай уточняющих вопросов до конца Phase 0.

  После Phase 0:
  - если auto_resolution прошло без ambiguity — continue silently к Phase 1.
  - если real ambiguity (несколько изолированных проектов в одном repo / неизвестный framework / scope spans разных стэков) — задай ОДИН блок через AskUserQuestion (2-4 вопроса максимум), затем continue.

  Дату для имени отчёта возьми из системного контекста (today's date).
  Все user-facing сообщения и финальный отчёт — на русском. Имена пакетов, идентификаторы кода, технические термины (`repository`, `middleware`, `endpoint`, `hook`, ...) — оставляй английскими.

  Перед стартом — короткое сообщение пользователю в одну строку: «Начинаю audit. Phase 0: detection.» Дальше работай по фазам.

  Begin Phase 0.
</start>

═════════════════════════════════════════════════════════════════════════════════
