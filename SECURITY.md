# Security Policy

Спасибо, что помогаете сделать ПромтЛаб безопаснее.

## Supported versions

Обновления безопасности выпускаются **только для последней стабильной
версии**, опубликованной в [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab).
Self-hosted инстансы рекомендуется обновлять в течение 7 дней после релиза
патча.

## Сообщить об уязвимости

**Не создавайте публичный GitHub Issue для уязвимостей.** Вместо этого:

- **Email:** slava0gpt@gmail.com с темой `[SECURITY] краткое описание`
- **GitHub Security Advisory:**
  https://github.com/Slava4123/promtlab/security/advisories/new
  (приватный канал, видят только мейнтейнеры)

Чем больше деталей — тем быстрее фикс:

1. Описание уязвимости и потенциального воздействия.
2. PoC или шаги воспроизведения.
3. Версия ПромтЛаб (`v?.?.?`), окружение (self-hosted / promtlabs.ru).
4. Ваши контакты для обратной связи.

## Что считается уязвимостью

Уязвимости, которые принимаются к рассмотрению:

- Authentication / authorization bypass
- SQL injection, SSRF, XXE
- RCE, privilege escalation
- Утечка приватных данных пользователей (промпты, API-ключи)
- OAuth flow hijacking, PKCE bypass
- CSRF с высоким impact
- Rate limit bypass на критичных endpoints
- Secrets leak в логах или error messages

Не считается уязвимостью (please do не сообщайте об этом):

- Missing security headers на публичных страницах
- Clickjacking на не-аутентифицированных pages
- CSRF на logout
- Self-XSS
- Теоретические уязвимости без PoC
- DoS через массовый флуд (решается за rate-limit)
- Output из автоматических сканеров без верификации

## Процесс обработки

| Шаг | Срок |
|-----|------|
| Подтверждение получения | 48 часов |
| Первичная оценка severity | 5 рабочих дней |
| Патч для critical/high | 14 дней |
| Публичный disclosure | После выпуска патча + 30 дней пользователям на обновление |

## Благодарности

Reporters, помогшие улучшить безопасность, упоминаются в release notes
(по желанию, с атрибуцией или анонимно). Денежного bug bounty пока нет.

## Safe harbor

Обязуемся не инициировать юридические действия против research, проведённого
в рамках этой политики и без вреда нашим пользователям и инфраструктуре.
