# Sentry/GlitchTip — следующие опциональные шаги

> **Статус базовой интеграции:** ✅ Завершена end-to-end (см. `DEPLOY.md` Шаг 11)
> Backend + Frontend Sentry SDK активны, source maps загружаются в CI, release tracking работает.
>
> Этот документ описывает **опциональные расширения** (PR 5, PR 7), которые имеет смысл включать **только при определённых триггерах**, не сразу.

---

## PR 5 — Performance Monitoring (tracing)

### Что это

Помимо ловли **ошибок** (errors), Sentry умеет собирать данные о **скорости работы** приложения через **transactions** и **spans**.

### Что добавляется при `SENTRY_TRACES_SAMPLE_RATE=0.1`

**Backend (Go, через `sentry-go` v0.44):**
- Каждый 10-й HTTP запрос обёрнут в **transaction**
- Внутри transaction автоматически создаются **spans** для:
  - HTTP request handling (overall duration)
  - DB queries через GORM (если интегрировать)
  - HTTP outbound calls (например, OpenRouter API)
- Sentry собирает: latency p50/p75/p95/p99, throughput (req/s), apdex score

**Frontend (React, через `@sentry/react` v10 + `browserTracingIntegration`):**
- Каждый 10-й page navigation создаёт **transaction**
- Spans для:
  - Initial page load (TTFB, FCP, LCP, FID, CLS — Web Vitals)
  - Route changes
  - Fetch/XHR requests (auto-instrumented) с длительностью каждого
- **Distributed tracing** — связь frontend transaction ↔ backend transaction через `sentry-trace` header

### Что появится в GlitchTip UI

Вкладка **Performance** покажет:
- Список endpoints (`GET /api/prompts`, `POST /api/auth/login`) с p95/p99 latency
- Slowest transactions (трейсы которые тормозили)
- Throughput графики
- Web Vitals для frontend (LCP, FID, CLS)

### Use cases

| Сценарий | Как помогает |
|---|---|
| **"У меня тормозит Dashboard"** | Performance UI → видишь что `GET /api/prompts` имеет p95 = 2.5s → находишь проблемный endpoint без догадок |
| **"Юзер жалуется на медленную загрузку"** | Web Vitals покажут конкретно где тормозит (LCP большой = картинки/шрифты, FID = JS блокирует main thread) |
| **N+1 queries в GORM** | Span для DB query внутри transaction покажет 50 одинаковых SQL за один request |
| **Долгие AI запросы** | Span для OpenRouter API покажет реальное время ответа модели |

### ❌ Когда PR 5 НЕ нужен (текущее состояние)

1. **Friends-and-family beta, низкий трафик** — у нас ~10-30 юзеров. Performance проблем нет, и при таком объёме нет данных для статистически значимого анализа (10 событий из 100 = слишком мало)
2. **Нет SLA** — никто не страдает если страница грузится 1.5s vs 1.0s
3. **GlitchTip Performance UI ограничен** — он принимает transactions, но dashboard для них **базовый** (не такой как у Sentry.io). Мало смысла собирать данные которые сложно анализировать
4. **Расходует ресурсы**:
   - Backend: каждая transaction добавляет ~5-10ms overhead на request
   - Frontend: bundle size +20-30 KB (browserTracingIntegration + утилиты)
   - Network: каждый event летит на `sentry.promtlabs.ru`, увеличивает нагрузку на GlitchTip
5. **Шум в UI** — десятки transactions в день для одинаковых endpoints не дают новой инфы

### ✅ Триггеры для включения PR 5

Включай когда **любое** из:
- **> 100 активных юзеров в день** — есть статистическая база
- **Появилась реальная жалоба на скорость** — нужны конкретные данные
- **Запускается новая функциональность** где важно понимать latency (например, AI streaming)
- **Появился платящий клиент** — performance критична для retention
- **Готов настроить дашборды и alert rules** на p95 latency

### Как активировать (когда время придёт)

**Шаг 1: на VPS** обновить `.env.prod`:

```bash
ssh root@85.239.39.45
cd /root/promtlab/promptvault
sed -i 's|^SENTRY_TRACES_SAMPLE_RATE=0.0|SENTRY_TRACES_SAMPLE_RATE=0.1|' .env.prod
docker compose -f docker-compose.prod.yml restart api
```

Backend начнёт семплировать 10% запросов сразу после restart.

**Шаг 2: GitHub Variable** для frontend:

```bash
gh variable set VITE_SENTRY_TRACES_SAMPLE_RATE --body "0.1" --repo Slava4123/promtlab
```

**Шаг 3: Trigger frontend rebuild** через push:

```bash
git commit --allow-empty -m "chore: enable performance monitoring (10% sample rate)"
git push origin main
```

CI пересоберёт frontend image с новым `VITE_SENTRY_TRACES_SAMPLE_RATE=0.1`, задеплоит на VPS.

**Шаг 4: проверка** через GlitchTip UI:

1. Открыть `https://sentry.promtlabs.ru/promtlab/performance`
2. Подождать ~10 минут трафика
3. Должны появиться transactions

### Откат PR 5

Если performance monitoring создаёт проблемы (overhead, шум):

```bash
# На VPS
sed -i 's|^SENTRY_TRACES_SAMPLE_RATE=0.1|SENTRY_TRACES_SAMPLE_RATE=0.0|' .env.prod
docker compose -f docker-compose.prod.yml restart api

# В GitHub
gh variable set VITE_SENTRY_TRACES_SAMPLE_RATE --body "0.0" --repo Slava4123/promtlab
git commit --allow-empty -m "chore: disable performance monitoring"
git push origin main
```

---

## PR 7 — Alert routing (Email + Telegram)

### Что это

GlitchTip ловит ошибки, но **не уведомляет** о новых issues автоматически — нужно настроить **alert rules** + **notification channels**.

### Что GlitchTip умеет из коробки

| Channel | Что нужно | Сложность |
|---|---|---|
| **Email** | `EMAIL_URL` в `.env.glitchtip` (SMTP) | 5-10 минут |
| **Webhook (generic)** | URL endpoint который принимает POST с JSON | зависит от приёмника |
| **Discord webhook** | Discord webhook URL (специальный формат) | 5 минут |
| **Slack webhook** | Slack incoming webhook URL | 5 минут |
| **Microsoft Teams** | Teams webhook | 5 минут |

**Telegram напрямую — НЕТ** в GlitchTip. Нужен либо:
1. Webhook → промежуточный сервис (Telegram bot который слушает HTTP) → Telegram message
2. Или один из existing relays: [Shoutrrr](https://github.com/containrrr/shoutrrr), [Apprise](https://github.com/caronc/apprise), n8n

### Сравнение Email vs Telegram

| Аспект | Email | Telegram |
|---|---|---|
| Время доставки | 5-30 секунд | 1-2 секунды |
| Setup сложность | 5-10 минут (если SMTP уже есть) | 30-60 минут (нужен relay сервис) |
| Риск spam filter | Есть (особенно первые письма) | Нет |
| Архив | Inbox навсегда, можно искать | История в чате, можно искать |
| Заметность | Можно пропустить в потоке | Push-нотификация телефона = моментально |
| Несколько получателей | Легко (CC, multiple addresses) | Нужны group chats или несколько chat_id |
| Mobile UX | Зависит от email-клиента | Отлично из коробки |

### Рекомендация

**Email сейчас, Telegram потом** — потому что:
- SMTP уже настроен в проекте (Gmail App Password для верификации юзеров)
- Можно переиспользовать те же creds для GlitchTip
- 10 минут работы вместо 30+
- Telegram добавишь когда email окажется недостаточно быстрым

**Идеально (когда время есть):** оба параллельно, разные правила:
- Email — для всех новых issues (low-priority)
- Telegram — только для critical (`> 10 errors/min`, `issue frequency spike`)

---

### Вариант A: Email Setup (рекомендую начать с него)

#### Шаг 1: получить SMTP credentials из `.env.prod`

```bash
ssh root@85.239.39.45
cd /root/promtlab/promptvault
grep "^SMTP_" .env.prod
```

Должны быть:
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=promstlab@gmail.com
SMTP_PASSWORD=<gmail app password>
SMTP_FROM=promstlab@gmail.com
```

#### Шаг 2: составить `EMAIL_URL` для GlitchTip

GlitchTip использует [django-environ](https://django-environ.readthedocs.io/en/latest/types.html#environ-env-email-url) формат:

```
smtp+tls://USER:PASSWORD@HOST:PORT
```

Для Gmail:
- `USER` = `promstlab@gmail.com` (URL-encoded если есть `@` → `%40`)
- `PASSWORD` = Gmail App Password (без пробелов!)
- `HOST` = `smtp.gmail.com`
- `PORT` = `587` для STARTTLS, `465` для SSL

**Пример (с URL-encoding для @):**
```
EMAIL_URL=smtp+tls://promstlab%40gmail.com:APPPASSWORDWITHOUTSPACES@smtp.gmail.com:587
```

**Важно:**
- Gmail требует **App Password**, не основной пароль (нужен 2FA на аккаунте)
- App Password обычно показывается с пробелами `xxxx xxxx xxxx xxxx` — **убери пробелы** перед записью в EMAIL_URL
- `@` в email нужно URL-encode как `%40`
- Спецсимволы в пароле (`/`, `?`, `#`) тоже нужно URL-encode

#### Шаг 3: обновить `.env.glitchtip`

```bash
nano .env.glitchtip
# Заменить строку:
# EMAIL_URL=consolemail://
# на:
# EMAIL_URL=smtp+tls://promstlab%40gmail.com:PASSWORD@smtp.gmail.com:587
```

#### Шаг 4: перезапустить GlitchTip контейнеры

```bash
docker compose -f docker-compose.prod.yml restart glitchtip-web glitchtip-worker
```

Они прочитают новый `.env.glitchtip` (~30 секунд downtime UI, не влияет на PromptLab).

Проверка логов:
```bash
docker compose -f docker-compose.prod.yml logs glitchtip-worker --tail 30 | grep -i email
```

#### Шаг 5: настроить уведомления в UI

1. Открыть `https://sentry.promtlabs.ru`
2. Profile → **Notifications** → включить **"New issue notification"** для project `promptvault-frontend` и `promptvault-backend`
3. Сохранить

**Альтернатива через Alert Rules** (более точная настройка):

1. Открыть project `promptvault-frontend` → **Project Settings** → **Project Alerts**
2. **Create New Alert**:
   - **Name:** `New issue email alert`
   - **Trigger:** `When a new issue is created`
   - **Action:** `Send email to project members`
3. Сохранить
4. Повторить для `promptvault-backend`

#### Шаг 6: Smoke test

В браузере на `https://promtlabs.ru` → DevTools Console:
```js
const err = new Error("EMAIL ALERT TEST " + Date.now());
window.onerror(err.message, window.location.href, 0, 0, err);
```

Через ~30 секунд:
- Issue появится в GlitchTip UI ✅
- На `promstlab@gmail.com` должно прийти письмо от `noreply@promtlabs.ru` (или от твоего SMTP_FROM) с темой типа `[promptvault-frontend] New error: EMAIL ALERT TEST ...`

**Если не пришло:**
- Проверить логи: `docker compose -f docker-compose.prod.yml logs glitchtip-worker --tail 50 | grep -iE "email|smtp|fail"`
- Проверить **Spam folder** в Gmail (первые письма часто туда попадают)
- Проверить что **App Password правильный** (без пробелов, актуальный)
- Проверить что Gmail аккаунт **разрешает SMTP** (Less secure apps disabled, App Password is the way)

#### Шаг 7: пометить как Not Spam (если попало туда)

Чтобы Gmail не помечал последующие алерты как spam:
1. Найти первое письмо в Spam
2. Открыть → **"Report not spam"**
3. Создать правило: From `noreply@promtlabs.ru` → Never send to spam
4. Следующие алерты должны приходить в Inbox

---

### Вариант B: Telegram Setup (сложнее)

#### Шаг 1: создать Telegram bot

1. Открыть [@BotFather](https://t.me/BotFather) в Telegram
2. `/newbot`
3. Name: `PromtLab Alerts` (или любое)
4. Username: `promtlab_alerts_bot` (должно заканчиваться на `_bot`)
5. Получить **token** формата `1234567890:AAxxxxxxxxxxxxxxxxxxxxxxxx`
6. **Сохранить токен** в password manager

#### Шаг 2: получить chat_id

1. Отправить любое сообщение своему боту в Telegram
2. Открыть в браузере: `https://api.telegram.org/bot<TOKEN>/getUpdates`
3. Найти `"chat":{"id":XXXXXXXXX,...}`
4. **Сохранить chat_id** (число)

#### Шаг 3: поднять webhook relay

GlitchTip умеет webhook, но Telegram bot API требует **специфический формат** запроса. Нужен промежуточный сервис.

**Самый простой вариант — `containrrr/shoutrrr` контейнер:**

Добавить в `docker-compose.prod.yml`:

```yaml
  shoutrrr:
    image: containrrr/shoutrrr:latest
    command: serve
    networks:
      - app
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 32M
```

Затем в `.env.prod`:
```env
TELEGRAM_BOT_TOKEN=1234567890:AAxxxxxxxxxxxxxxxxxxxxxxxx
TELEGRAM_CHAT_ID=XXXXXXXXX
```

И HTTP endpoint shoutrrr будет: `http://shoutrrr:8080/api/v1/notify`

С body: `{"router": "telegram://TOKEN@telegram?chats=CHAT_ID", "message": "{{ .text }}"}`

**Альтернатива — простой Python webhook (~30 строк):**

```python
# webhook_relay.py
from flask import Flask, request
import requests
import os

app = Flask(__name__)
TOKEN = os.environ['TELEGRAM_BOT_TOKEN']
CHAT_ID = os.environ['TELEGRAM_CHAT_ID']

@app.route('/glitchtip-webhook', methods=['POST'])
def webhook():
    data = request.json
    text = f"🚨 GlitchTip Alert\n{data.get('text', 'Unknown error')}\n{data.get('url', '')}"
    requests.post(
        f'https://api.telegram.org/bot{TOKEN}/sendMessage',
        json={'chat_id': CHAT_ID, 'text': text, 'parse_mode': 'Markdown'}
    )
    return 'ok', 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8081)
```

В docker-compose:
```yaml
  webhook-relay:
    build: ./scripts/webhook-relay
    env_file: .env.prod
    networks:
      - app
    restart: unless-stopped
```

#### Шаг 4: настроить webhook в GlitchTip

1. Открыть project → **Project Alerts** → **Create New Alert**
2. **Trigger:** New issue / Issue frequency > N
3. **Action:** Webhook
4. **URL:** `http://webhook-relay:8081/glitchtip-webhook` (внутренний docker network URL)
5. Сохранить

#### Шаг 5: Smoke test

Триггерить ошибку → проверить что Telegram bot прислал сообщение в чат.

#### Шаг 6 (опционально): фильтрация

В Python relay можно добавить логику чтобы шла **только critical errors**:

```python
@app.route('/glitchtip-webhook', methods=['POST'])
def webhook():
    data = request.json
    level = data.get('level', 'error')
    # Только error/fatal в Telegram, warnings/info - пропускаем
    if level in ('error', 'fatal'):
        # ... send to Telegram
    return 'ok', 200
```

---

## Сводная таблица команд

### Быстрая активация PR 5

```bash
# VPS
ssh root@85.239.39.45
cd /root/promtlab/promptvault
sed -i 's|^SENTRY_TRACES_SAMPLE_RATE=0.0|SENTRY_TRACES_SAMPLE_RATE=0.1|' .env.prod
docker compose -f docker-compose.prod.yml restart api

# Локально
gh variable set VITE_SENTRY_TRACES_SAMPLE_RATE --body "0.1" --repo Slava4123/promtlab
git commit --allow-empty -m "chore: enable performance monitoring"
git push origin main
```

### Быстрая активация PR 7 (Email только)

```bash
# VPS
ssh root@85.239.39.45
cd /root/promtlab/promptvault
nano .env.glitchtip
# Заменить EMAIL_URL=consolemail:// на smtp+tls://...

docker compose -f docker-compose.prod.yml restart glitchtip-web glitchtip-worker
```

Затем в UI настроить notifications или alert rules.

---

## Когда вернуться к этому документу

| Триггер | Что включить |
|---|---|
| > 100 active users / день | PR 5 (performance) |
| Жалоба "тормозит" | PR 5 |
| Появился платящий клиент | PR 5 + PR 7 (alerts критичны) |
| Хочу узнавать о новых ошибках | PR 7 (минимум email) |
| Команда > 1 разработчика | PR 7 (telegram + email) |
| После настройки SMTP в `.env.glitchtip` | PR 7 email — 5 минут активации |

---

## Связанные документы

- `docs/DEPLOY.md` Шаг 11 — базовая настройка GlitchTip + security TODO
- `.claude/plans/silly-zooming-pumpkin.md` — полный plan интеграции (10-секционный senior architect документ)
- `CLAUDE.md` — стек проекта + ключевые архитектурные решения

---

**Дата создания:** 2026-04-08
**Статус:** Reference — активировать только при триггерах из таблицы выше
