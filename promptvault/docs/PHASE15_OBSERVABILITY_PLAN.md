# Observability end-to-end: Prometheus + Alertmanager + Grafana

> **Status: ✅ Closed (2026-04-26).** Phase 15 завершена и фактически
> расширена до Phase 16 объёма: помимо Prometheus + Alertmanager + Grafana
> добавлены node-exporter, postgres-exporter, cAdvisor v0.52.1 (cgroup v2),
> blackbox-exporter, Loki + Promtail (logs), Tempo (traces) и SLO
> multi-burn-rate alerts. Финальный scope и runbook — в `docs/OBSERVABILITY.md`.
> Этот документ оставлен как исторический план, разделы ниже отражают
> исходное намерение Phase 15.

## Context

Код observability частично готов с Phase 14.3: зарегистрировано 3 counter-метрики в `backend/internal/infrastructure/metrics/metrics.go` (`share_quota_increment_failed_total`, `analytics_insights_refresh_total{result}`, `analytics_cleanup_deleted_total{table}`), написаны 4 alert rules в `promptvault/infra/prometheus/alerts.yaml`, есть feature-flag `SERVER_METRICS_ENABLED`. **Но end-to-end pipeline не развёрнут**: `/metrics` endpoint за nginx отдаёт SPA fallback (нет location), нет scraper'а, нет Alertmanager'а, alert rules лежат мёртвым грузом. Сейчас prod мониторит только GlitchTip (exceptions), метрик — нет.

Цель волны: развернуть production-grade monitoring stack (Prometheus + Alertmanager + Grafana) на prod VPS, подключить к `/metrics`, завести alert rules в evaluation, настроить Telegram-уведомления по существующим 4 alert'ам, открыть Grafana UI через `grafana.promtlabs.ru` с basic auth. Пользователь сознательно выбрал upgrade VPS 2→4 GB + полный стандартный стек (не lightweight VictoriaMetrics) — ориентир «как в книгах», запас памяти на будущее.

## 1. Резюме

Развёртываем классический Prometheus stack в 3 контейнерах (`prometheus`, `alertmanager`, `grafana`), scraping api через Docker network `app` по `http://api:8080/metrics` с защитой IP-allowlist middleware в Go. Prometheus UI и Alertmanager UI биндятся на `127.0.0.1` и доступны только через SSH tunnel. Grafana публично через `grafana.promtlabs.ru` c Let's Encrypt, nginx reverse proxy, basic auth. Alert routing: Telegram (critical+warning) + SMTP как Phase 2. Pre-condition: upgrade VPS до 4 GB через Timeweb console, создание Telegram бота, DNS A-запись для Grafana. Ожидаемый memory footprint monitoring-стека: 430-500 MB, укладывается в 4 GB VPS с запасом ~1.8 GB.

## 2. Архитектурные решения

### Storage/scraper: Prometheus (не VictoriaMetrics)

- **Выбрано:** `prom/prometheus:v3.0.1`, scrape_interval 30s, retention 30d (time) + 2GB (size cap).
- **Альтернатива 1:** VictoriaMetrics `vmsingle` — 40-80 MB RAM против 300 MB. Отклонена: user выбрал upgrade VPS + стандартный стек, запас ресурса есть.
- **Альтернатива 2:** Grafana Cloud push (нулевой RAM на VPS). Отклонена: external egress из РФ, гео-риск, [ПРЕДПОЛОЖЕНИЕ] возможные санкции.
- **Источник:** `context7 /prometheus/prometheus` — default retention 15d, поддерживает `--storage.tsdb.retention.size=2GB` + `--storage.tsdb.retention.time=30d`.

### Alert manager: официальный Prometheus Alertmanager

- **Выбрано:** `prom/alertmanager:v0.27.0`.
- **Альтернатива:** vmalert (встроен в VM) — отклонена вместе с VM.
- **Источник:** `context7 /prometheus/alertmanager` — native Telegram receiver с `bot_token`+`chat_id`, parse_mode HTML default, route c group_wait 30s/group_interval 5m/repeat_interval 4h.

### Grafana: да, persistent dashboards

- **Выбрано:** `grafana/grafana:11.3.0`, auto-provision Prometheus datasource через `infra/grafana/provisioning/datasources/prometheus.yml`.
- **Альтернатива:** только Prometheus Web UI без Grafana. Отклонена: user выбрал полный стек.

### Memory plan: upgrade 2→4 GB, сохраняем все текущие сервисы

Бюджет после upgrade (4 GB VPS):

| Сервис | mem_limit | expected | source |
|---|---|---|---|
| api | 512M | ~100M | существует |
| frontend (nginx) | 128M | ~30M | существует |
| glitchtip-web | 768M | ~500M | существует |
| glitchtip-worker | 512M | ~250M | существует |
| glitchtip-valkey | 192M | ~80M | существует |
| **prometheus (new)** | **384M** | **~200M** | планируется |
| **alertmanager (new)** | **96M** | **~40M** | планируется |
| **grafana (new)** | **192M** | **~80M** | планируется |
| OS + Docker daemon | — | ~400M | оценка |
| **Ожидаемое использование** | | **~1680M** | |
| **Лимит на 4 GB VPS** | | **4096M** | |
| **Запас** | | **~2400M** | |

### Ingress/access: смешанный

- **Prometheus UI (`:9090`) и Alertmanager UI (`:9093`):** bind `127.0.0.1`, SSH tunnel только. Причина: utility для редкого использования админом, не оправдывает доп. attack surface.
- **Grafana (`:3000`):** публичный через `https://grafana.promtlabs.ru` + basic auth. Причина: user выбрал «как в книгах», Grafana рассчитана на регулярный UI-доступ.

### `/metrics` защита: IP-allowlist на уровне Go-роутера

- **Выбрано:** middleware `ipallowlist.New("172.0.0.0/8,127.0.0.1")` на `/metrics` route.
- Альтернатива: allow/deny на nginx. Отклонена: `/metrics` не экспозится наружу через nginx вообще — scraping происходит внутри Docker network по `http://api:8080/metrics`, минуя nginx. Защита в Go страхует на случай ошибочного nginx route.

### Notification: Telegram (primary) + SMTP через Alertmanager (Phase 2)

- **Phase 1:** Telegram receiver с `bot_token` + `chat_id`, route `group_by: [alertname, severity]`, `repeat_interval: 4h`. Critical и warning в один чат (флаг severity в тексте сообщения).
- **Phase 2 (отложено):** SMTP receiver через существующий `cfg.SMTP`, routing critical-only на email.
- **Источник:** `context7 /prometheus/alertmanager`, официальная схема `telegram_configs`.

### Правка `alerts.yaml`: добавить `for: 1h` к `InsightsComputeLoopDead`

- Текущий expr `absent_over_time(...[48h])` без `for:` даёт **гарантированный false positive** при первом запуске Prometheus (у vmalert/prometheus нет истории метрики 48h назад → `absent` срабатывает сразу).
- Добавляем `for: 1h` — страхует от всплеска alerts при старте и коротких рестартах.

## 3. Изменения в коде

### Новые файлы

- `promptvault/infra/prometheus/prometheus.yml` — scrape_config для job `promptvault-api`, targets `['api:8080']`, metrics_path `/metrics`, scrape_interval 30s. Плюс `rule_files: ['/etc/prometheus/rules/*.yaml']`, `alerting.alertmanagers: [{static_configs: [targets: ['alertmanager:9093']]}]`.
- `promptvault/infra/alertmanager/alertmanager.yml` — route c `group_by: [alertname, severity]`, `group_wait: 30s`, `group_interval: 5m`, `repeat_interval: 4h`, один receiver `telegram` с шаблоном HTML (severity + summary + description + firing/resolved).
- `promptvault/infra/grafana/provisioning/datasources/prometheus.yml` — auto-provision Prometheus datasource по URL `http://prometheus:9090`.
- `promptvault/nginx/grafana.conf.template` (или добавление в существующий nginx config) — server block для `grafana.promtlabs.ru` с `proxy_pass http://grafana:3000`, Let's Encrypt SSL, basic auth через `htpasswd` файл (монтируется из secret).

### Модификации

- `promptvault/backend/internal/infrastructure/config/server.go` — добавить поле `MetricsAllowlist string` с `koanf:"metrics_allowlist"`, env `SERVER_METRICS_ALLOWLIST`, default `"172.0.0.0/8,127.0.0.1"`.
- `promptvault/backend/internal/app/routes.go:39` — обернуть `/metrics` route в `ipallowlist.New(allowlist, false)` middleware. Использовать существующий `promptvault/backend/internal/middleware/ipallowlist/ipallowlist.go` (CIDR parsing готов).
- `promptvault/infra/prometheus/alerts.yaml` — добавить `for: 1h` к `InsightsComputeLoopDead` rule.
- `promptvault/docker-compose.prod.yml` — 3 новых service (`prometheus`, `alertmanager`, `grafana`), 3 новых volumes (`prometheus-data`, `alertmanager-data`, `grafana-data`), update `CERTBOT_DOMAINS` env на `${DOMAIN},${SENTRY_DOMAIN},${GRAFANA_DOMAIN}` для Let's Encrypt.
- `promptvault/.env.prod` (на VPS) и `promptvault/.env.example` — новые переменные:
  - `SERVER_METRICS_ENABLED=true`
  - `SERVER_METRICS_ALLOWLIST=172.0.0.0/8,127.0.0.1`
  - `GRAFANA_DOMAIN=grafana.promtlabs.ru`
  - `GRAFANA_ADMIN_PASSWORD=<random 32 chars>`
  - `ALERTMANAGER_TELEGRAM_BOT_TOKEN=<bot token>`
  - `ALERTMANAGER_TELEGRAM_CHAT_ID=<chat id>`
  - `GF_INSTALL_PLUGINS=` (пустое, опционально)
- `promptvault/docs/OBSERVABILITY.md` — добавить секцию «Текущий статус» (все ✅), подсекции Access (SSH tunnel + Grafana URL), Runbook (как диагностировать упавший alert).
- `promptvault/docs/DEPLOY.md` §11.9 memory check — обновить таблицу с 3 новыми контейнерами и зафиксировать upgrade 2→4 GB.

### Контракты между слоями (новые/изменённые)

- Go config получает 1 новое поле `Server.MetricsAllowlist` (обратно совместимо, default заполнен).
- `/metrics` endpoint получает middleware — поведение для не-allowlist IP: 403 Forbidden (как в ipallowlist.go).
- Docker network `app` остаётся единственной — `prometheus`/`alertmanager`/`grafana` подключаются к ней для доступа к `api`.

## 4. Модель данных / Volumes

Новых таблиц в PostgreSQL нет. Новые Docker volumes:

- `prometheus-data` — TSDB (ожидаемый размер 30d × 3 counter-метрики × 1 instance ≈ 50-200 MB).
- `alertmanager-data` — state (active alerts, silences), ~5 MB.
- `grafana-data` — SQLite с dashboards + session store, ~20-50 MB.

## 5. API-контракт

| Endpoint | Доступность | Метод |
|---|---|---|
| `http://api:8080/metrics` (Docker-internal) | Docker network `app` only + IP-allowlist | scrape изнутри Prometheus |
| `http://prometheus:9090` (Docker-internal) | Docker network `app` | Grafana datasource |
| `http://alertmanager:9093` (Docker-internal) | Docker network `app` | Prometheus → alertmanager push |
| `127.0.0.1:9090` (bind VPS loopback) | SSH tunnel | админ через `ssh -L 9090:127.0.0.1:9090` |
| `127.0.0.1:9093` (bind VPS loopback) | SSH tunnel | `ssh -L 9093:127.0.0.1:9093` |
| `https://grafana.promtlabs.ru` | Публичный HTTPS + basic auth | регулярный UI-доступ |

Никаких новых public HTTP endpoints от самого приложения.

## 6. План тестирования

### Unit
- `config/server_test.go` (если существует) — проверить что `MetricsAllowlist` парсится из env и `.env.example` default работает.
- `middleware/ipallowlist/ipallowlist_test.go` — уже покрыт (не трогаем).

### Integration
- После шага A: `docker exec -it api wget -qO- http://api:8080/metrics` внутри контейнера → 200 и вывод text/plain. `curl -I https://promtlabs.ru/metrics` снаружи → 404 (nginx не маршрутизирует).
- После шага C: `ssh -L 9090:127.0.0.1:9090 deploy@...`, браузер `http://localhost:9090/targets` → `promptvault-api` в state `UP`, last scrape < 1min.
- После шага D: временно установить `for: 1m` в одном из alert rules, подождать 3 мин → Telegram получает сообщение. Откатить `for:`.
- После шага E: `https://grafana.promtlabs.ru` → basic auth login, datasource Prometheus → `Test` → `Success`, пустой dashboard показывает все 3 counter'а через Explore.

### E2E smoke
- `curl -I https://promtlabs.ru/api/health` → 200, TTFB не деградировал (nginx не трогали).
- `curl -I https://sentry.promtlabs.ru/` → 200/302 (GlitchTip живой).
- `docker stats --no-stream` → суммарно < 2.5 GB RSS, запас > 1.5 GB.

### Что НЕ тестируем
- Нагрузочное тестирование Prometheus (overkill для 3 метрик + малый трафик).
- HA Alertmanager (single replica достаточно для single VPS).
- Grafana dashboard authoring — пока только default Explore.

## 7. План внедрения пошаговый

### Pre-шаги (ручные, ответственность пользователя, до любого push'а)

**P1 — Upgrade VPS 2→4 GB:**
- Timeweb Cloud console → VPS `85.239.39.45` → Change plan → 4 GB RAM.
- Reboot (Timeweb делает автоматически при смене плана).
- SSH проверка: `free -h` → `Mem: 3.8Gi total` (или похожее).

**P2 — Telegram bot setup:**
1. В Telegram: написать `@BotFather` → `/newbot` → имя `PromtLabs Alerts Bot` → username например `promtlabs_alerts_bot` → получить token формата `1234567890:ABC...`.
2. Написать боту любое сообщение (чтобы активировать чат).
3. Написать `@userinfobot` в Telegram, переслать ему любое сообщение от своего аккаунта → получить numeric `Your user ID: 123456789`.
4. Сохранить `BOT_TOKEN` и `CHAT_ID` — вставить в `.env.prod` на шаге D.

**P3 — DNS для Grafana:**
- Timeweb панель → DNS zone `promtlabs.ru` → добавить A-запись: `grafana` → `85.239.39.45`, TTL 300.
- Дождаться propagation: `dig +short grafana.promtlabs.ru` → `85.239.39.45`.

### Шаг A — Backend config + feature flag + правка alerts.yaml (один PR)

**Файлы:**
- `backend/internal/infrastructure/config/server.go` — добавить `MetricsAllowlist`.
- `backend/internal/app/routes.go` — обернуть `/metrics` в ipallowlist.
- `infra/prometheus/alerts.yaml` — `for: 1h` к `InsightsComputeLoopDead`.
- `.env.example` — `SERVER_METRICS_ENABLED=true`, `SERVER_METRICS_ALLOWLIST=172.0.0.0/8,127.0.0.1`.

**Критерий готовности:**
- `go test -short ./... && golangci-lint run` green.
- Merge в main → CI зелёный → deploy проходит (prod обновил api контейнер, `docker exec api wget -qO- http://api:8080/metrics` внутри контейнера даёт 200).

**Диффа:** ~25 строк.

### Шаг B — Infra config файлы (один PR, файлы + commit без активации)

**Файлы:**
- `infra/prometheus/prometheus.yml` (новый).
- `infra/alertmanager/alertmanager.yml` (новый, с `<CHAT_ID>`/`<TOKEN>` placeholder'ами — реальные значения на шаге D через env).
- `infra/grafana/provisioning/datasources/prometheus.yml` (новый).
- `docs/OBSERVABILITY.md` — добавить секцию "Текущий статус" (пока с отметкой "Rollout в процессе").

**Критерий готовности:**
- `promtool check config infra/prometheus/prometheus.yml` локально даёт OK (опционально — если promtool не установлен, пропустить).
- Merge → CI проходит (docs-only изменения не триггерят docker builds для api/frontend).

**Диффа:** ~60 строк.

### Шаг C — Prometheus в docker-compose (один PR, активация)

**Файлы:**
- `docker-compose.prod.yml` — добавить service `prometheus` + volume `prometheus-data`. Mounts: `./infra/prometheus:/etc/prometheus/rules:ro`, `./infra/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro`. Bind `127.0.0.1:9090:9090`. `restart: unless-stopped`. Network `app`.

**Deploy:**
- После merge → CI deploy workflow → на VPS `docker compose -f docker-compose.prod.yml up -d prometheus`.

**Критерий готовности:**
- `ssh -L 9090:127.0.0.1:9090 deploy@...` → браузер `http://localhost:9090/targets` → `promptvault-api` = UP.
- Query в UI `share_quota_increment_failed_total` → возвращает серию (0 или актуальное значение).
- `docker stats prometheus` → RSS < 300 MB.

**Диффа:** ~15 строк.

### Шаг D — Alertmanager + Telegram receiver (один PR + env update на VPS)

**Файлы:**
- `docker-compose.prod.yml` — добавить service `alertmanager` + volume `alertmanager-data`. Bind `127.0.0.1:9093:9093`. Network `app`. Env `TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID` прокидываются в контейнер.
- `infra/alertmanager/alertmanager.yml` — заменить placeholder'ы на `${TELEGRAM_BOT_TOKEN}` / `${TELEGRAM_CHAT_ID}` (Alertmanager поддерживает env interpolation с `$(env)` или через sed при старте — выберем sed в entrypoint если alertmanager не умеет).

**Env update на VPS (вручную):**
- `echo 'ALERTMANAGER_TELEGRAM_BOT_TOKEN=1234:ABC...' >> .env.prod`
- `echo 'ALERTMANAGER_TELEGRAM_CHAT_ID=123456789' >> .env.prod`
- `chmod 600 .env.prod`

**Deploy:**
- После merge → CI → `docker compose up -d alertmanager`.
- Smoke test: временно изменить `for: 1m` у `ShareQuotaIncrementLeak` или создать тестовый alert через `amtool` (`docker exec alertmanager amtool alert add alertname=TestAlert severity=critical ...`). Подождать 2 мин. Telegram приходит.
- Откатить временное изменение.

**Критерий готовности:**
- Telegram получает тестовое уведомление.
- `http://localhost:9093/#/alerts` (SSH tunnel) — UI доступен.

**Диффа:** ~20 строк compose + ~50 строк alertmanager.yml.

### Шаг E — Grafana + nginx vhost + Let's Encrypt (один PR)

**Файлы:**
- `docker-compose.prod.yml` — service `grafana` + volume `grafana-data`. Mounts provisioning. Env `GF_SECURITY_ADMIN_PASSWORD`, `GF_SERVER_ROOT_URL=https://grafana.promtlabs.ru`, `GF_INSTALL_PLUGINS=`. Bind `127.0.0.1:3000:3000`.
- `nginx/grafana.conf` (новый) — server block для `grafana.promtlabs.ru`, proxy_pass на `http://grafana:3000`, basic auth через `auth_basic_user_file /etc/nginx/grafana.htpasswd`. Letsencrypt сертификат добавляется через `CERTBOT_DOMAINS` в compose.
- `Dockerfile.frontend-nginx` / frontend nginx config template — добавить include новых .conf.
- `docker-compose.prod.yml` — update `CERTBOT_DOMAINS=${DOMAIN},${SENTRY_DOMAIN},${GRAFANA_DOMAIN}`.

**Env update на VPS:**
- `GRAFANA_DOMAIN=grafana.promtlabs.ru`
- `GRAFANA_ADMIN_PASSWORD=<random 32 chars>` → через `openssl rand -base64 32`
- `htpasswd -c /home/deploy/promtvault/nginx/grafana.htpasswd admin` → запрос пароль, тот же или другой.

**Deploy:**
- `docker compose up -d grafana` + frontend rebuild (если nginx config в frontend image).
- Let's Encrypt получает сертификат автоматически для нового домена (~30s после старта).

**Критерий готовности:**
- `curl -I https://grafana.promtlabs.ru` → 401 (basic auth prompt).
- Браузер с login: admin / <GRAFANA_ADMIN_PASSWORD> → Grafana UI.
- Datasource Prometheus → Test → Success.
- Explore → query `analytics_insights_refresh_total` → график.

**Диффа:** ~30 строк compose + ~40 строк nginx config.

### Шаг F — Документация (один PR)

**Файлы:**
- `docs/OBSERVABILITY.md` — обновить «Текущий статус» на все ✅, добавить подсекции: Access (SSH tunnel команды + Grafana URL), Runbook (по каждому из 4 alert'ов — что делать когда сработало), Retention policy.
- `docs/DEPLOY.md` §11.9 — актуализировать memory таблицу (3 новых контейнера, 4 GB VPS), убрать warning про «недостаточно 2 GB».

**Диффа:** ~80 строк.

## 8. Риски и митигации

| Риск | Вероятность | Impact | Mitigation |
|---|---|---|---|
| Timeweb upgrade с downtime > 5 мин | Низкая | High | Upgrade обычно hot (без reboot при change plan). Если reboot — все сервисы стартуют через docker compose restart-policy. Общий downtime ≤ 2 мин. Запланировать upgrade в низком трафике. |
| Let's Encrypt rate-limit при новом домене | Низкая | Middle | Certbot retry default, плюс grafana.promtlabs.ru — единственный новый домен, лимит 20/неделя недостижим. |
| Prometheus OOM при cardinality взрыва | Низкая | Middle | `mem_limit: 384M` в compose — Docker kill только prometheus, api/frontend не затронуты. Counter-метрики с фиксированными labels — cardinality статичная. |
| Telegram bot flooded несколькими алертами | Средняя | Low | Alertmanager `group_interval: 5m`, `repeat_interval: 4h` → максимум 1 сообщение в 5 мин на группу, повтор только через 4ч. |
| `InsightsComputeLoopDead` false positive на свежем prometheus | **Средняя без fix** | Middle | **Шаг A добавляет `for: 1h`** — страхует. |
| vmalert vs prometheus несовместимость (если выберем vmalert) | N/A | — | Выбрали prometheus native — риск снят. |
| Падение prometheus блокирует api | Нулевая | — | Prometheus не в dependency chain api. `restart: unless-stopped`. Scraping pull-model. |
| Grafana admin password утёк через git | Низкая | High | `.env.prod` в gitignore (проверено), `GF_SECURITY_ADMIN_PASSWORD` только через env. |

## 9. Метрики успеха

### Технические (через 24ч после deploy)
- `up{job="promptvault-api"} == 1` непрерывно (scrape success ratio > 99%).
- Все 4 alert rules в Prometheus UI state `inactive` (ни один не firing).
- `docker stats` total RSS < 2.2 GB (запас > 1.8 GB на 4 GB VPS).
- Prod latency `/api/health` p99 не выросла против baseline.
- Telegram не получает сообщений (означает ни один real alert не firing).

### Качественные
- Grafana доступна владельцу через UI в браузере без SSH.
- При намеренно вызванном alert (временный `for: 1m`) уведомление в Telegram < 3 мин.
- Runbook в OBSERVABILITY.md понятен — новый разработчик может за 10 мин понять как проверить состояние.

### Что покажет что fallback сработал
- OOM event на prometheus → Docker рестартует его → в Telegram не приходит alert (прямая зависимость), но api и frontend не затронуты (prod up).

## 10. Открытые вопросы

1. **nginx config для Grafana: template или static?** В репо frontend image собирается из template с env interpolation (см. `Dockerfile.frontend-nginx`). Новый vhost для Grafana нужно добавить либо в тот же template, либо как отдельный `.conf.template`. Решается на шаге E через чтение текущей структуры nginx config.

2. **Htpasswd файл: commit или generate on VPS?** Рекомендация — generate on VPS через `htpasswd -c ...`, монтировать в nginx контейнер. Не коммитить хэш (даже bcrypt) в public репо.

3. **Grafana anonymous access для view-only dashboard?** По умолчанию нет (обязательный login). Можно включить через `GF_AUTH_ANONYMOUS_ENABLED=true` + `GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer`. Предлагаю **не включать** — нет use-case, добавит атаку поверхности.

4. **Retention 30d достаточно?** Для `absent_over_time(...[48h])` нужно минимум 48h. 30d это 15x запас. Ок.

5. **Alert rules pre-flight:** прогнать локально `promtool check rules infra/prometheus/alerts.yaml` перед шагом C — если promtool не установлен, рекомендация поставить через `docker run --rm -v ./infra/prometheus:/r prom/prometheus:v3.0.1 promtool check rules /r/alerts.yaml`.

6. **Phase 2 (SMTP alerts):** отложить в отдельный план или сразу включить в этот? Решение: отложить — Telegram покрывает 1 владельца, SMTP станет актуальным при втором админе или команде.

## Critical files

- `promptvault/backend/internal/infrastructure/config/server.go` (добавить `MetricsAllowlist`)
- `promptvault/backend/internal/app/routes.go:39` (обернуть `/metrics` в ipallowlist middleware)
- `promptvault/backend/internal/middleware/ipallowlist/ipallowlist.go` (переиспользуем as-is, уже готов)
- `promptvault/backend/internal/infrastructure/metrics/metrics.go` (не трогаем, готов)
- `promptvault/infra/prometheus/alerts.yaml` (добавить `for: 1h` к `InsightsComputeLoopDead`)
- `promptvault/docker-compose.prod.yml` (3 новых service + 3 volumes + update CERTBOT_DOMAINS)
- `promptvault/nginx/*` (добавить grafana vhost — структура проверить на шаге E)
- `promptvault/.env.prod` (на VPS, вручную) + `promptvault/.env.example` (в git)
- `promptvault/docs/OBSERVABILITY.md` (обновить статус)
- `promptvault/docs/DEPLOY.md` §11.9 (обновить memory budget)

## Verification (end-to-end runbook)

Выполняется после шага F:

```
# 1. VPS живой после upgrade
free -h  # total 3.8Gi (было 1.9Gi)
docker stats --no-stream  # все сервисы up, total RSS < 2.2 GB

# 2. API /metrics защищён и отдаёт данные (внутри network)
docker exec api wget -qO- http://api:8080/metrics | head -20
# → Prometheus text exposition format

docker exec api wget -qO- http://api:8080/metrics 2>&1
# extern curl через nginx → 404 (no route)
curl -I https://promtlabs.ru/metrics  # → HTTP/2 404

# 3. Prometheus scrape success
ssh -L 9090:127.0.0.1:9090 deploy@85.239.39.45 &
# browser: http://localhost:9090/targets
# → promptvault-api = UP, last scrape < 30s

# 4. All 3 metrics visible
# http://localhost:9090/graph → query:
#   share_quota_increment_failed_total
#   analytics_insights_refresh_total
#   analytics_cleanup_deleted_total

# 5. Alertmanager live
ssh -L 9093:127.0.0.1:9093 deploy@85.239.39.45 &
# http://localhost:9093 → UI shows 0 active alerts, config loaded

# 6. Grafana live
curl -I https://grafana.promtlabs.ru  # → 401
# browser login → UI → Explore → query → график

# 7. Alert channel test (один раз после deploy)
# На VPS: временно поменять в infra/prometheus/alerts.yaml
#   for: 25h → for: 1m у ShareQuotaIncrementLeak
# Docker compose exec prometheus kill -HUP 1  (reload config)
# Отправить testовый request который incrementит counter
# Через 3 мин — сообщение в Telegram
# Откатить for:, reload

# 8. Prod health не деградировал
curl -I https://promtlabs.ru/api/health  # 200, TTFB < 2s
curl -I https://sentry.promtlabs.ru/     # 200/302
```

## Scope-контроль

**Включено:** Prometheus + Alertmanager + Grafana, IP-allowlist /metrics, Telegram нотификации, Grafana через subdomain с basic auth, правка `for:` в одном alert, обновление OBSERVABILITY.md и DEPLOY.md.

**Не включено (вне scope):**
- SMTP alert receiver (Phase 2).
- Custom Grafana dashboards (только default datasource).
- Новые метрики в коде (latency histograms, business KPI).
- Log aggregation (Loki/Vector).
- HA Prometheus.
- Anonymous Grafana access.
- Grafana OIDC/SAML.
- Экспорт prometheus метрик из PostgreSQL (pg_stat exporter) — отдельная волна.
- Upgrade actions на Node.js 24 в GitHub Actions (отдельный мелкий техдолг).
