# ПромтЛаб — Руководство по деплою

## Инфраструктура (уже создано)

| Компонент | Данные |
|-----------|--------|
| **VPS** | Timeweb Cloud, 2 CPU, 2 GB RAM, 40 GB NVMe, Ubuntu 24.04 |
| **IP сервера** | `85.239.39.45` |
| **Домен** | `promtlabs.ru` (купить и оплатить!) |
| **Managed PostgreSQL** | Timeweb Cloud, PostgreSQL 18 |
| **БД хост** | `fdf27c65a5c6ba390823dd0e.twc1.net` |
| **БД порт** | `5432` |
| **БД пользователь** | `gen_user` |
| **БД имя** | `promtlab` |
| **БД SSL** | `verify-full` (сертификат `ca.crt` в корне проекта) |

---

## Что уже сделано на VPS

```bash
# Система обновлена
apt update && apt upgrade -y

# Docker установлен
curl -fsSL https://get.docker.com | sh

# Пользователь deploy создан
adduser deploy
usermod -aG sudo deploy
usermod -aG docker deploy

# Firewall настроен
ufw allow 22/tcp   # SSH
ufw allow 80/tcp   # HTTP
ufw allow 443/tcp  # HTTPS
ufw enable

# fail2ban установлен
apt install -y fail2ban
systemctl enable fail2ban

# Автообновления безопасности
apt install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades
```

---

## Что осталось сделать

### Шаг 1: Купить и оплатить домен

Домен `promtlabs.ru` добавлен но не оплачен. Оплатить в панели Timeweb.

При покупке:
- SSL Timeweb Pro — **ВЫКЛ** (у нас бесплатный Let's Encrypt)
- Пакет "Старт" — **ВЫКЛ**
- Все допуслуги — **ВЫКЛ**
- Только домен за 169-200₽

### Шаг 2: Настроить DNS

В панели Timeweb → Домены → `promtlabs.ru` → DNS-записи:

| Тип | Имя | Значение |
|-----|-----|----------|
| A | `@` | `85.239.39.45` |
| A | `www` | `85.239.39.45` |

DNS распространяется 5-30 минут. Проверка: `ping promtlabs.ru` должен показать `85.239.39.45`.

### Шаг 3: Инициализировать Git и запушить код

На **локальной машине** в папке проекта:

```bash
cd C:\GolandProjects\awesomeProject\test\promptvault

# Инициализировать git
git init
git add .
git commit -m "Initial commit — ПромтЛаб"

# Создать private репозиторий на GitHub, затем:
git remote add origin git@github.com:YOUR_USERNAME/promtlabs.git
git push -u origin main
```

**ВАЖНО:** `.env.prod` и `.env.dev` в `.gitignore` — они НЕ попадут в репозиторий. Это правильно, секреты не должны быть в git.

### Шаг 4: Захардить SSH (на VPS)

```bash
# Подключиться к VPS
ssh root@85.239.39.45

# Настроить SSH ключи (на ЛОКАЛЬНОЙ машине, в другом терминале)
ssh-keygen -t ed25519 -f ~/.ssh/promtlabs
ssh-copy-id -i ~/.ssh/promtlabs deploy@85.239.39.45

# Захардить SSH (на VPS)
nano /etc/ssh/sshd_config
# Изменить:
#   PermitRootLogin no
#   PasswordAuthentication no
#   PubkeyAuthentication yes
systemctl restart sshd

# С этого момента вход только через SSH ключ, не пароль!
# Проверить вход ПЕРЕД закрытием текущей сессии:
# ssh -i ~/.ssh/promtlabs deploy@85.239.39.45
```

### Шаг 5: Склонировать проект на VPS

```bash
# Войти как deploy
ssh deploy@85.239.39.45

# Склонировать
git clone git@github.com:YOUR_USERNAME/promtlabs.git ~/promtlabs
cd ~/promtlabs
```

### Шаг 6: Создать .env.prod на VPS

```bash
nano .env.prod
```

Вставить содержимое (заменить CHANGE_ME значения на реальные):

```env
# Domain & SSL
DOMAIN=promtlabs.ru
CERTBOT_EMAIL=your-real-email@gmail.com

# Server
SERVER_PORT=8080
SERVER_ENVIRONMENT=production
SERVER_ALLOWED_ORIGINS=https://promtlabs.ru
SERVER_FRONTEND_URL=https://promtlabs.ru
SERVER_SECURE_COOKIES=true

# Database (Managed PostgreSQL от Timeweb Cloud)
DATABASE_HOST=fdf27c65a5c6ba390823dd0e.twc1.net
DATABASE_PORT=5432
DATABASE_USER=gen_user
DATABASE_PASSWORD=СЮДА_ПАРОЛЬ_БД
DATABASE_NAME=promtlab
DATABASE_SSLMODE=verify-full
DATABASE_SSLROOTCERT=/app/ca.crt
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5

# JWT (сгенерировать: openssl rand -hex 32)
JWT_SECRET=СЮДА_64_СИМВОЛА
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h

# OAuth (из консолей провайдеров, опционально для MVP)
OAUTH_CALLBACK_BASE=https://promtlabs.ru
OAUTH_GITHUB_CLIENT_ID=
OAUTH_GITHUB_CLIENT_SECRET=
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_CLIENT_SECRET=
OAUTH_YANDEX_CLIENT_ID=
OAUTH_YANDEX_CLIENT_SECRET=

# AI (из openrouter.ai)
AI_OPENROUTER_API_KEY=СЮДА_КЛЮЧ_OPENROUTER
AI_RATE_LIMIT_RPM=10

# SMTP (опционально — для email верификации)
SMTP_HOST=
SMTP_PORT=465
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=
```

Защитить файл:

```bash
chmod 600 .env.prod
```

Сгенерировать JWT_SECRET:

```bash
openssl rand -hex 32
```

### Шаг 7: Скопировать SSL сертификат БД на VPS

С **локальной машины**:

```bash
scp ca.crt deploy@85.239.39.45:~/promtlabs/ca.crt
```

### Шаг 8: Запустить деплой

```bash
cd ~/promtlabs
bash scripts/deploy.sh
```

Скрипт автоматически:
1. Проверит `.env.prod` на заполненность и placeholder'ы
2. Соберёт Docker образы
3. Запустит контейнеры
4. Проверит health endpoint

### Шаг 9: Проверить

```bash
# Статус контейнеров
docker compose -f docker-compose.prod.yml ps

# Логи
docker compose -f docker-compose.prod.yml logs api --tail 20
docker compose -f docker-compose.prod.yml logs frontend --tail 20

# Health check
curl https://promtlabs.ru/api/health
# Ожидаемый ответ: {"status":"ok"}
```

В браузере: `https://promtlabs.ru` — должна открыться лендинг страница с замком HTTPS.

---

## Архитектура в production

```
Пользователь → https://promtlabs.ru
    ↓
[Nginx + Let's Encrypt] (Docker: jonasal/nginx-certbot)
    ├── /api/* → [Go Backend] (Docker: alpine)
    │                ↓
    │        [Managed PostgreSQL] (Timeweb Cloud, внешний)
    │
    └── /* → [Static React SPA] (встроен в nginx образ)
```

## Полезные команды

```bash
# Пересобрать и перезапустить
docker compose -f docker-compose.prod.yml up -d --build

# Посмотреть логи в реальном времени
docker compose -f docker-compose.prod.yml logs -f

# Остановить
docker compose -f docker-compose.prod.yml down

# Обновить код
git pull
docker compose -f docker-compose.prod.yml up -d --build

# Проверить SSL сертификат
curl -vI https://promtlabs.ru 2>&1 | grep "SSL certificate"
```

## Шаг 10: CI/CD через GitHub Actions

Автоматический pipeline: **Lint → Тесты (Go + React параллельно) → Deploy → Health Check → Rollback при ошибке**.

Workflow файл: `.github/workflows/deploy.yml` (в корне репозитория).

### 10.1 Создать SSH ключ для GitHub Actions

На **VPS** (как deploy):

```bash
ssh-keygen -t ed25519 -f ~/.ssh/github-actions -N ""
cat ~/.ssh/github-actions.pub >> ~/.ssh/authorized_keys
cat ~/.ssh/github-actions  # скопировать приватный ключ → GitHub Secret
```

### 10.2 Настроить Deploy Key (для git fetch на VPS)

Чтобы VPS мог подтягивать код из приватного репозитория:

```bash
# На VPS (как deploy):
ssh-keygen -t ed25519 -f ~/.ssh/github-deploy -N ""
cat ~/.ssh/github-deploy.pub
# → Добавить в GitHub → Repo → Settings → Deploy keys (Allow read access)

# Настроить SSH config:
cat >> ~/.ssh/config << 'EOF'
Host github.com
  IdentityFile ~/.ssh/github-deploy
EOF
```

### 10.3 Добавить Secrets в GitHub

GitHub → Repository → Settings → Secrets and variables → Actions → New repository secret:

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | `85.239.39.45` |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | Содержимое `~/.ssh/github-actions` (приватный ключ) |
| `VPS_PORT` | `22` |
| `DOMAIN` | `promtlabs.ru` (для health check) |

### 10.4 Pipeline

```
Push to main (promptvault/** changed)
         │
    ┌────┴────┐
    │  Lint   │  golangci-lint v2
    └────┬────┘
         │
    ┌────┴────────────┐
    │                 │
┌───┴──────┐  ┌──────┴───────┐
│  Test Go │  │ Test React   │  (параллельно)
│  -short  │  │ vitest run   │
│  -race   │  │              │
└───┬──────┘  └──────┬───────┘
    └────┬───────────┘
         │
    ┌────┴─────┐
    │  Deploy  │  SSH → git pull → docker compose up --build
    └────┬─────┘
         │
    ┌────┴──────────┐
    │ Health Check  │  curl /api/health (5 попыток × 10с)
    └────┬──────────┘
         │
      SUCCESS ──── done
         │ (failure)
    ┌────┴──────┐
    │ Rollback  │  git reset --hard PREV → rebuild
    └───────────┘
```

### 10.5 Как это работает

1. Push в `main` → запускается pipeline
2. **Lint**: golangci-lint v2 проверяет Go-код
3. **Тесты** (параллельно): Go unit-тесты (`-short -race`) + React vitest
4. **Deploy**: SSH на VPS → `git fetch` + `docker compose up --build`
5. **Health check**: 5 попыток curl к `https://DOMAIN/api/health`
6. **Rollback**: при ошибке откат к предыдущему коммиту + пересборка
7. Всё кешируется: Go modules, npm, golangci-lint (~5-7 мин total)

### 10.6 Безопасность

- `permissions: contents: read` — минимальные права workflow
- `concurrency: cancel-in-progress: true` — один деплой одновременно
- `script_stop: true` — SSH-action падает при ошибке любой команды
- `.env.prod` живёт только на VPS, никогда не попадает в git
- Deploy key (read-only) для git fetch

---

## Стоимость в месяц

| Услуга | Цена |
|--------|:----:|
| VPS (2 CPU, 2 GB, 40 GB) | 800₽ |
| Публичный IP | 180₽ |
| Бэкапы VPS | 240₽ |
| Managed PostgreSQL | 790₽ |
| Публичный IP БД | 180₽ |
| Бэкапы БД | 120₽ |
| Домен promtlabs.ru | ~17₽ (200₽/год) |
| **Итого** | **~2 327₽/мес** |

## Что было сделано в проекте за сессию

### Аудит и исправления (5 раундов × 9 агентов)
- 67 исправлений кода (security, correctness, robustness, code quality)
- 259 тестов (213 backend unit + 24 integration + 22 frontend)
- Type design score: 6.9 → 9.1/10

### Новые фичи
- Username (@ник) — поле в профиле, поиск пользователей, инвайт по @username с autocomplete
- golang-migrate вместо AutoMigrate
- testcontainers для integration тестов
- vitest для frontend тестов

### Deploy инфраструктура
- HTTPS через jonasal/nginx-certbot (автоматический Let's Encrypt)
- Production-grade docker-compose (networks, limits, logging)
- Non-root Docker контейнеры
- deploy.sh с валидацией и health check
- Managed PostgreSQL с SSL (verify-full)

### Docker Image Registry (GHCR)

CI/CD собирает Docker-образы и публикует в GitHub Container Registry. Для VPS нужна одноразовая настройка:

```bash
# На VPS: авторизация в GHCR (одноразово)
# Создать PAT: GitHub → Settings → Developer settings → Personal access tokens → read:packages
echo "YOUR_PAT" | docker login ghcr.io -u slava4123 --password-stdin
```

Также добавь GitHub secret `DOMAIN` = `promtlabs.ru` (Settings → Secrets → Actions).

После этого CI автоматически:
1. Собирает образы с Docker layer caching
2. Пушит в `ghcr.io/slava4123/promtlab-api:latest` и `ghcr.io/slava4123/promtlab-frontend:latest`
3. На VPS делает `docker compose pull` + `up -d`

Для rollback: `docker compose pull` с конкретным SHA-тегом.

### Переименование
- ПромтХаб → ПромтЛаб во всех файлах

---

## Шаг 11: Sentry (GlitchTip) — Error Tracking + Monitoring

**Цель:** self-hosted GlitchTip (API-совместимый с Sentry SDK) для ловли runtime ошибок backend + frontend.

**Почему GlitchTip, а не Sentry.io:** Sentry.io заблокирован для РФ с сентября 2024 (OFAC санкции — `https://sentry.zendesk.com/hc/en-us/articles/28038067843739`). Sentry self-hosted требует 16+ GB RAM. GlitchTip — open-source, 100% совместим с `sentry-go` и `@sentry/react`, работает в ~500 MB RAM.

### 11.1. Создать отдельную БД `glitchtip` в Timeweb Cloud

1. Зайти в Timeweb Cloud console → Managed PostgreSQL → выбрать существующий инстанс `fdf27c65a5c6ba390823dd0e`
2. Вкладка **Databases** → **Add database**:
   - Name: `glitchtip`
   - Owner: `gen_user` (переиспользуем существующего)
3. Подождать ~10 секунд, БД готова. Миграции Django применятся автоматически при первом старте GlitchTip контейнера.

### 11.2. Настроить DNS для `sentry.promtlabs.ru`

В панели Timeweb → Домены → `promtlabs.ru` → DNS-записи:

| Тип | Имя | Значение |
|-----|-----|----------|
| A | `sentry` | `85.239.39.45` |

DNS распространяется 5-30 минут. Проверка: `ping sentry.promtlabs.ru` → `85.239.39.45`.

### 11.3. Создать `.env.glitchtip` на VPS

```bash
ssh deploy@85.239.39.45
cd ~/promtlabs/promptvault
cp .env.glitchtip.example .env.glitchtip
chmod 600 .env.glitchtip
nano .env.glitchtip
```

Заполнить реальными значениями:
- `SECRET_KEY` — `openssl rand -hex 32`
- `DATABASE_URL` — с реальным паролем (тот же что в `.env.prod` `DATABASE_PASSWORD`)
- `EMAIL_URL` — переиспользовать SMTP credentials из `.env.prod` в формате `smtp+tls://USER:PASSWORD@HOST:PORT`
- `GLITCHTIP_DOMAIN=https://sentry.promtlabs.ru`
- `DEFAULT_FROM_EMAIL=noreply@promtlabs.ru`

### 11.4. Добавить `SENTRY_DOMAIN` в `.env.prod`

```bash
nano .env.prod
# Добавить в конец:
SENTRY_DOMAIN=sentry.promtlabs.ru
```

Эта переменная используется `docker-compose.prod.yml` для:
1. Build-arg в frontend image (nginx envsubst для vhost)
2. `CERTBOT_DOMAINS=${DOMAIN},${SENTRY_DOMAIN}` для автоматического получения SSL через Let's Encrypt

### 11.5. Pull новых образов и запустить GlitchTip

```bash
cd ~/promtlabs/promptvault
docker compose -f docker-compose.prod.yml pull glitchtip-web glitchtip-worker glitchtip-valkey
docker compose -f docker-compose.prod.yml up -d glitchtip-valkey glitchtip-web glitchtip-worker
```

Первый старт займёт ~90 секунд (Django migrations + collectstatic). Проверить:

```bash
docker compose -f docker-compose.prod.yml logs glitchtip-web --tail 50
# Ждать: "Starting gunicorn" и отсутствие ERROR
```

### 11.6. Пересобрать frontend image для nginx с новым vhost

Nginx конфиг изменился (CSP + новый server block). Нужен rebuild frontend image:

```bash
# Локально в репо:
git add -A
git commit -m "feat: sentry (glitchtip) infrastructure"
git push origin main
# CI/CD автоматически пересоберёт и задеплоит новый frontend image.
# На VPS:
docker compose -f docker-compose.prod.yml pull frontend
docker compose -f docker-compose.prod.yml up -d frontend
```

Let's Encrypt получит сертификат для `sentry.promtlabs.ru` при первом запуске (~30 секунд).

### 11.7. Создать superuser для GlitchTip UI

```bash
docker compose -f docker-compose.prod.yml exec glitchtip-web ./manage.py createsuperuser
# Ввести email, password
```

### 11.8. Настроить GlitchTip через UI

1. Открыть `https://sentry.promtlabs.ru` — должна открыться login страница
2. Войти под superuser
3. Создать organization: `promtlab`
4. Создать **2 проекта**:
   - `promptvault-backend` (platform: **Go**)
   - `promptvault-frontend` (platform: **React**)
5. Скопировать **DSN** каждого проекта (Settings → Client Keys). Они понадобятся для PR 2 и PR 4:
   - `SENTRY_DSN` → backend
   - `VITE_SENTRY_DSN` → frontend
6. Создать **Auth Token** (User Settings → Auth Tokens → Create):
   - Scope: `project:releases`, `project:write`, `org:read`
   - Скопировать токен (будет использован в GitHub secrets как `SENTRY_AUTH_TOKEN` в PR 6)

### 11.9. Memory check после старта

**⚠️ Важно:** итоговые memory limits всех сервисов:

| Сервис | Limit |
|---|---|
| api | 512 MB |
| frontend (nginx) | 128 MB |
| glitchtip-web | 768 MB |
| glitchtip-worker | 512 MB |
| glitchtip-valkey | 192 MB |
| **Итого limits** | **2112 MB** |
| + OS / buffers | ~300-400 MB |
| **Требуется VPS** | **~2.5-3 GB RAM** |

**Текущий VPS 2 GB RAM недостаточен для полного стека.** Нужно **upgrade до 4 GB** в Timeweb Cloud console (примерно +400 ₽/мес) **перед** запуском GlitchTip.

Альтернатива — временно уменьшить limits (в `docker-compose.prod.yml`) до:
- glitchtip-web: `384M`
- glitchtip-worker: `256M`
- glitchtip-valkey: `128M`

И в `.env.glitchtip`: `GUNICORN_WORKERS=2`, `CELERY_WORKER_CONCURRENCY=1`. Но это sub-optimal, лучше делать upgrade.

Проверка после старта:

```bash
free -m
# Expected: available > 300 MB

docker stats --no-stream
# Expected (на 4 GB VPS):
#   glitchtip-web    ~400-600 MB
#   glitchtip-worker ~250-400 MB
#   glitchtip-valkey ~50-120 MB
#   api              ~80-150 MB
#   frontend         ~20-50 MB
```

Если `available < 300 MB` стабильно — upgrade VPS или вынос GlitchTip на отдельный сервер.

### 11.10. Smoke test (до начала PR 2)

```bash
# GlitchTip UI доступен
curl -I https://sentry.promtlabs.ru
# Expected: HTTP/2 200 или 302

# PromptLab всё ещё работает
curl https://promtlabs.ru/api/health
# Expected: {"status":"ok"}
```

### 11.11. Включение Sentry в production (PR 4 — errors only)

После запуска GlitchTip и получения DSN/Auth Token из UI, включить SDK в prod.

**Шаг 1:** Добавить GitHub Repository secrets (Settings → Secrets and variables → Actions → New repository secret):

| Secret | Значение | Назначение |
|---|---|---|
| `VITE_SENTRY_DSN` | Frontend DSN из GlitchTip (Project Settings → Client Keys) | Build-arg frontend image |
| `SENTRY_DSN` | Backend DSN из GlitchTip | Для документации, реально устанавливается в .env.prod |
| `SENTRY_DOMAIN` | `sentry.promtlabs.ru` | nginx vhost, CERTBOT |
| `SENTRY_URL` | `https://sentry.promtlabs.ru` | sentry-cli endpoint |
| `SENTRY_ORG` | slug организации (обычно `promtlab`) | sentry-cli |
| `SENTRY_PROJECT_FRONTEND` | `promptvault-frontend` | sentry-cli releases |
| `SENTRY_PROJECT_BACKEND` | `promptvault-backend` | (зарезервировано на будущее) |
| `SENTRY_AUTH_TOKEN` | Auth Token из GlitchTip (User Settings → Auth Tokens) | sentry-cli source maps upload |

Также в **Settings → Secrets and variables → Actions → Variables** (не secrets!):

| Variable | Значение |
|---|---|
| `VITE_SENTRY_ENABLED` | `true` |
| `VITE_SENTRY_TRACES_SAMPLE_RATE` | `0.0` (PR 4 — errors only, performance выключен) |

**Шаг 2:** На VPS обновить `.env.prod`:

```bash
ssh deploy@85.239.39.45
cd ~/promtlabs/promptvault
nano .env.prod
```

Добавить/обновить:
```env
SENTRY_ENABLED=true
SENTRY_DSN=<Backend DSN из GlitchTip UI>
SENTRY_ENVIRONMENT=production
SENTRY_TRACES_SAMPLE_RATE=0.0
SENTRY_DEBUG=false
SENTRY_RELEASE=placeholder
# SENTRY_RELEASE перезаписывается CI скриптом (deploy.yml) на актуальный GITHUB_SHA при каждом деплое.
```

**Шаг 3:** Rebuild frontend (через CI) — нужен чтобы `VITE_SENTRY_*` build-args попали в bundle. Любой push в `main` запустит pipeline:

```bash
# Локально в репо
git commit --allow-empty -m "chore: trigger rebuild to enable Sentry in prod"
git push origin main
```

CI соберёт frontend с `VITE_SENTRY_ENABLED=true`, загрузит source maps в GlitchTip, задеплоит на VPS, обновит `SENTRY_RELEASE` в `.env.prod` и перезапустит api контейнер.

**Шаг 4:** Smoke test (см. 11.13 ниже).

### 11.12. Включение Performance Monitoring (PR 5)

После подтверждения что error tracking работает (smoke test 11.13 прошёл):

**Шаг 1:** В GitHub Repository Variables изменить:
- `VITE_SENTRY_TRACES_SAMPLE_RATE` → `0.1` (10% транзакций семплируется)

**Шаг 2:** На VPS обновить `.env.prod`:
```bash
nano .env.prod
# Изменить:
SENTRY_TRACES_SAMPLE_RATE=0.1
```

**Шаг 3:** Rebuild frontend + restart backend:
```bash
git commit --allow-empty -m "chore: enable Sentry performance monitoring"
git push origin main
```

**Шаг 4:** Проверить после деплоя:
- GlitchTip UI → Performance — должны появляться transactions
- `/api/health` latency через `loggermw` логи не должна увеличиться более чем на 10%

Если видишь regression — rollback: `SENTRY_TRACES_SAMPLE_RATE=0.0` в .env.prod + `VITE_SENTRY_TRACES_SAMPLE_RATE=0.0` в GitHub variables, push empty commit.

### 11.13. Smoke test после включения Sentry в prod

**Backend error capture:**
```bash
# На VPS или локально с curl'ом
# Дождаться пока юзер вызовет ошибку 500 (или провоцировать вручную).
# В GlitchTip UI → promptvault-backend → Issues должен появиться event в течение 30 секунд.
# Event должен содержать:
# - release = <GITHUB_SHA>
# - environment = production
# - user.id (если ошибка произошла в protected endpoint)
# - stack trace Go
```

**Frontend error capture:**
1. Открыть `https://promtlabs.ru`
2. DevTools → Console
3. Выполнить: `throw new Error("e2e test " + Date.now())`
4. Через 10-30 сек в GlitchTip UI → promptvault-frontend → Issues должен быть event
5. Stack trace должен показывать **исходные TypeScript файлы** (если source maps загружены через PR 6), не `assets/index-XYZ.js`

**Release attribution check:**
- В GlitchTip UI → Releases → найти `<GITHUB_SHA>` — должен быть список attached commits + issues
- Клик на "View commit" → redirect на GitHub commit page

**PII scrubbing check:**
1. Провоцировать 500 ошибку на protected endpoint (с JWT токеном)
2. В GlitchTip event → Request Headers
3. **Ожидается:** нет Authorization header, нет Cookie header
4. Если видишь — BeforeSend scrubber не сработал, проверить `backend/cmd/server/main.go`

**Health check unaffected:**
```bash
curl https://promtlabs.ru/api/health
# Expected: {"status":"ok"}
```

Даже если GlitchTip недоступен, backend должен работать (fail-open паттерн `sentry.Init`).

### 11.14. Security TODO — перед переходом с beta на full production

⚠️ **Обязательно выполнить** когда появятся реальные платящие юзеры:

1. **Сменить пароль `glitchtip_user`** на random 32+ chars:
   - Timeweb Cloud console → Managed PG → promtlab → Пользователи → glitchtip_user → Изменить пароль
   - Обновить `DATABASE_URL` в `.env.glitchtip` на VPS
   - Перезапустить `glitchtip-web` и `glitchtip-worker`

2. **Изолировать привилегии `glitchtip_user`** — сейчас он имеет все права на **обе** БД (promtlab + glitchtip):
   - Timeweb Cloud UI → glitchtip_user → Привилегии → выключить "Использовать одинаковые привилегии для всех баз"
   - Оставить права только на `glitchtip`, убрать с `promtlab`
   - Проверить что GlitchTip продолжает работать

3. **Rotate `gen_user` пароль** если он попадал в shared docs, screenshots, git history:
   - Timeweb UI → gen_user → Изменить пароль
   - Обновить `DATABASE_PASSWORD` в `.env.prod` на VPS
   - Перезапустить api контейнер

4. **Rotate Sentry Auth Token** при подозрении на compromise:
   - GlitchTip UI → User Settings → Auth Tokens → Revoke + Create new
   - Обновить `SENTRY_AUTH_TOKEN` в GitHub secrets

5. **Проверить `.env.glitchtip` permissions:**
   ```bash
   chmod 600 .env.glitchtip
   ls -la .env.glitchtip
   # Expected: -rw------- 1 deploy deploy
   ```

6. **Включить `ENABLE_USER_REGISTRATION=false`** — уже установлено в `.env.glitchtip.example`, проверить что в реальном `.env.glitchtip` тоже так.

### 11.15. Дальнейшие шаги (опционально)

**PR 7 — Alert routing в Telegram:**
1. Создать Telegram bot через @BotFather, получить token
2. Получить chat_id (отправить сообщение боту, GET `https://api.telegram.org/bot<TOKEN>/getUpdates`)
3. В GlitchTip UI → Organization → Alerts → Create Alert:
   - Condition: "A new issue is created" или "Issue frequency > N"
   - Action: Webhook → URL: свой endpoint который пересылает в Telegram
   - Для простоты: использовать [Shoutrrr](https://github.com/containrrr/shoutrrr) или простой Python webhook-relay
4. Тест: триггерить ошибку, убедиться что алерт пришёл

**Monitoring VPS (опционально):**
```bash
# Cron job каждые 5 минут проверяет available RAM
echo '*/5 * * * * deploy test $(free -m | awk "/^Mem:/ {print \$7}") -lt 300 && curl -s -X POST "https://api.telegram.org/bot<TOKEN>/sendMessage" -d "chat_id=<CHAT_ID>&text=⚠️ VPS RAM low: $(free -m | awk \"/^Mem:/ {print \\\$7}\")MB available"' | crontab -
```

**План полностью описан в `.claude/plans/silly-zooming-pumpkin.md`.**
