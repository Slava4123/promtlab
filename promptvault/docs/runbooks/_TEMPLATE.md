# Runbook: <AlertName>

**Severity:** P0 / P1 / P2
**Pager:** Telegram + Email | Telegram only | Telegram silenceable

## Symptom (что увидит пользователь)

Что заметит пользователь приложения / админ. Внешние симптомы.

## Impact

Кого затронет: всех / только Pro/Max юзеров / тестовые env. Severity rationale.

## Investigation

Команды для диагностики (выполнять по порядку):

```bash
# 1. Проверить состояние сервиса
docker ps | grep <service>
docker logs <service> --tail 50

# 2. Метрики в Prometheus / Grafana
# Query: <prometheus query>

# 3. Связанные логи
# Loki: {container="<service>"} |= "ERROR"
```

## Mitigation (immediate)

Что сделать сейчас чтобы остановить кровотечение:

1. Шаг 1
2. Шаг 2

## Resolution (long-term)

Что fix'ить чтобы alert не повторился:

- Pull request с fix'ом
- Test coverage добавление
- Architectural change

## Post-mortem

Если incident `> 5 min` impact — заполнить post-mortem template.
