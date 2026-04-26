# Alertmanager secrets — VPS setup

Alertmanager читает чувствительные значения из файлов в `/etc/alertmanager/secrets/`
(на хосте — `infra/alertmanager/secrets/`, монтируется read-only в контейнер).
Сама папка `secrets/` в `.gitignore` (см. строку 50 корневого `.gitignore`).

## Файлы

| Файл | Назначение | Source |
| --- | --- | --- |
| `bot_token` | Telegram Bot API token | https://t.me/BotFather → `/newbot` или `/token` |
| `smtp_password` | Gmail App Password для email-critical receiver (Phase 2) | https://myaccount.google.com/apppasswords |

## Telegram bot_token (Шаг D Phase 15)

```bash
ssh deploy@<vps>
cd /home/deploy/promtvault
mkdir -p infra/alertmanager/secrets
printf '%s' '<TELEGRAM_BOT_TOKEN>' > infra/alertmanager/secrets/bot_token
chmod 600 infra/alertmanager/secrets/bot_token

# chat_id у Alertmanager v0.27.0 нельзя читать из файла — заменяем sed'ом
# placeholder в alertmanager.yml. v0.28+ поддерживает chat_id_file.
sed -i "s/chat_id: 1$/chat_id: <TELEGRAM_CHAT_ID>/" infra/alertmanager/alertmanager.yml

docker compose -f docker-compose.prod.yml restart alertmanager
docker logs alertmanager --tail 30
```

## SMTP smtp_password (Phase 2 — email backup для critical)

1. У аккаунта `promstlab@gmail.com` (значение `smtp_auth_username` в alertmanager.yml)
   включить 2FA в Google Account → Security.
2. Сгенерировать App Password: https://myaccount.google.com/apppasswords →
   название «Alertmanager VPS». 16 символов без пробелов.
3. Положить на VPS:

```bash
ssh deploy@<vps>
cd /home/deploy/promtvault
printf '%s' '<APP_PASSWORD_16_CHARS>' > infra/alertmanager/secrets/smtp_password
chmod 600 infra/alertmanager/secrets/smtp_password

docker compose -f docker-compose.prod.yml restart alertmanager
docker logs alertmanager --tail 30
# В логах не должно быть `failed to read smtp_password` или `auth failed`.
```

## Тест end-to-end

Один раз, перед закрытием Phase 15:

1. Скопировать `alerts.yaml` в test-копию или временно поменять выражение
   у любого warning alert на always-true (`vector(1) > 0`) с `for: 1m`.
2. Reload Prometheus: `docker compose exec prometheus kill -HUP 1`.
3. Через 2-3 минуты:
   - Telegram-канал — пришло уведомление.
   - Inbox `promstlab@gmail.com` — пришло email уведомление (только для
     `severity: critical`, через `continue: true` в route).
4. Откатить тестовое изменение, ещё раз reload. Должны прийти `[RESOLVED]`
   уведомления в оба канала.

## Permission troubleshooting

Alertmanager image (`prom/alertmanager:v0.27.0`) запускается как UID `nobody` (65534).
Если файлы на VPS принадлежат `root:root` (по умолчанию после `printf > file`),
read-only mount всё равно работает: nobody может читать `chmod 600` файлы
владельца root, потому что docker mount не нарушает Linux DAC. Если в логах
видно `permission denied` — проверить SELinux/AppArmor контекст хоста.
