# Runbook: InstanceDown (Blackbox external probe failure)

**Severity:** P0 (critical)
**Pager:** Telegram + Email immediate

## Symptom

`probe_success == 0` от blackbox-exporter для одного из:
- `https://promtlabs.ru/` — главный сайт
- `https://promtlabs.ru/api/health` — backend API
- `https://grafana.promtlabs.ru/` — Grafana
- `https://sentry.promtlabs.ru/` — GlitchTip

5+ минут подряд. Извне сервис недоступен.

## Impact

**P0**. Если `promtlabs.ru` или `/api/health` — юзеры не могут пользоваться сервисом. Полный outage.

Для `grafana` / `sentry` — административный impact (мы не можем мониторить), не клиентский.

## Investigation

```bash
# 1. VPS жив вообще?
ping 85.239.39.45

# 2. SSH работает?
ssh root@85.239.39.45 "uptime; free -h; docker ps"

# 3. Какой именно endpoint упал?
# Grafana → ПромтЛаб — Производительность приложения → External uptime probes
# Видно какой instance у которого probe_success=0.

# 4. Если api endpoint:
ssh root@85.239.39.45 "docker logs promptvault-api-1 --tail 100"
ssh root@85.239.39.45 "docker logs promptvault-frontend-1 --tail 50"

# 5. nginx-certbot — cert expired?
# Если SSL handshake fails — cert проблема:
ssh root@85.239.39.45 "docker exec promptvault-frontend-1 openssl x509 -in /etc/letsencrypt/live/promtlabs.ru/cert.pem -noout -enddate"

# 6. DNS — возможно DNS-провайдер flap?
dig +short promtlabs.ru
```

## Mitigation

**Если контейнер crash:**
1. `docker compose -f docker-compose.prod.yml up -d <service>` — рестарт.
2. Проверить `docker stats` — не OOM ли.

**Если cert expired:**
1. nginx-certbot должен auto-renew. Проверить `docker logs promptvault-frontend-1 | grep certbot`.
2. Manual renew: `docker exec promptvault-frontend-1 certbot renew --force-renewal`.

**Если VPS down полностью:**
1. Timeweb Cloud console → VPS status. Reboot если нужно.

## Resolution

- Investigate root cause: OOM, panic, network issue, cert auto-renew failed.
- Если повторяется — наладить healthchecks + restart policies.
- Если VPS reliability — рассмотреть failover region.

## Post-mortem

**ОБЯЗАТЕЛЬНО** при downtime > 5 мин.
