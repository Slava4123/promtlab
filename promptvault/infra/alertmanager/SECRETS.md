# Alertmanager secrets — VPS setup

Alertmanager читает SMTP пароль из файла в `/etc/alertmanager/secrets/`
(на хосте — `infra/alertmanager/secrets/`, монтируется read-only в контейнер).
Сама папка `secrets/` в `.gitignore` (строка 50 корневого `.gitignore`).

## Email-only deployment

С прод VPS Timeweb (РФ) Telegram Bot API недоступен — пакеты к
`api.telegram.org:443` блокируются на IPv4. Поэтому в `alertmanager.yml`
сконфигурирован один email-receiver, Telegram удалён.

Gmail SMTP (`smtp.gmail.com:587`) работает через IPv6 — Docker bridge
network даёт IPv6 fallback автоматически.

## Setup smtp_password (Gmail App Password)

1. У аккаунта `promstlab@gmail.com` (значение `smtp_auth_username`) включить
   2FA: Google Account → Security.
2. Сгенерировать App Password: <https://myaccount.google.com/apppasswords>,
   название — «Alertmanager VPS». 16 символов без пробелов.
3. Положить на VPS:

```bash
ssh root@<vps>
cd /root/promtlab/promptvault
mkdir -p infra/alertmanager/secrets
printf '%s' '<APP_PASSWORD_16_CHARS>' > infra/alertmanager/secrets/smtp_password

# КРИТИЧНО: alertmanager image runs as UID 65534 (nobody).
# Файл с владельцем root:root mode 600 контейнер прочитать НЕ сможет —
# ловится в логах как 'permission denied: smtp_password' и email не уйдёт.
chown 65534:65534 infra/alertmanager/secrets/smtp_password
chmod 600 infra/alertmanager/secrets/smtp_password

docker compose -f docker-compose.prod.yml restart alertmanager
docker logs promptvault-alertmanager-1 --tail 30
# В логах не должно быть `failed to read smtp_password` или `auth failed`.
```

## Тест end-to-end

Один раз, перед закрытием Phase 15 / после ротации App Password:

1. Скопировать любое существующее warning-правило в test-копию или временно
   изменить выражение на `vector(1)` и `for: 1m`. Положить как
   `/etc/prometheus/rules/test_alert.yaml` через `docker cp` (UID nobody
   не позволяет писать в bind-mount директории напрямую).
2. Reload Prometheus: `curl -X POST http://localhost:9090/-/reload`.
3. Через ~90 секунд (`for: 1m` + `group_wait: 30s`): проверить inbox
   `promstlab@gmail.com` — должно прийти `[CRITICAL] TestAlertName`.
4. Откатить тестовое изменение, ещё раз reload.
5. Через `resolve_timeout` (5 мин) придёт `[RESOLVED]` уведомление.

## Permission troubleshooting

- `permission denied: smtp_password` в логах → `chown 65534:65534`.
- `auth failed: 535-5.7.8 Username and Password not accepted` → App
  Password протух или Google заблокировал. Перегенерировать App Password.
- `dial tcp ... 587: connect: connection timed out` через IPv4 — Roskomnadzor
  блокирует Gmail SMTP по IPv4. Docker bridge должен иметь IPv6: проверить
  `docker network inspect promptvault_app | grep -i ipv6`.
