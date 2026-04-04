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

### Переименование
- ПромтХаб → ПромтЛаб во всех файлах
