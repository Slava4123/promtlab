# Security audit checklist перед открытием репо

Этот чек-лист надо прогнать **один раз локально** перед тем, как делать
`Slava4123/promtlab` публичным. Результаты сканеров НЕ коммитить.

## 1. Установка сканеров (Windows, Git Bash)

```bash
# Gitleaks — быстрый scanner Go-бинарь
curl -L https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks_8.21.2_windows_x64.zip -o gitleaks.zip
unzip gitleaks.zip -d "$HOME/bin"

# TruffleHog — более глубокое сканирование (entropy + regex)
curl -L https://github.com/trufflesecurity/trufflehog/releases/latest/download/trufflehog_3.82.6_windows_amd64.tar.gz | tar -xz -C "$HOME/bin"
```

## 2. Сканирование всей истории

```bash
cd C:/GolandProjects/awesomeProject/test

# Gitleaks — быстрый проход
gitleaks detect --source . --report-path gitleaks-report.json --redact

# TruffleHog — глубже
trufflehog git file://. --only-verified --json > trufflehog-report.json

# Посмотреть сколько найдено
jq '. | length' gitleaks-report.json
jq 'select(.Verified==true)' trufflehog-report.json
```

## 3. Если что-то найдено

Зависит от типа находки:

### Случай А: секрет всё ещё активен
1. **Немедленно ротируйте** credentials (OpenRouter, T-Bank, SMTP, JWT
   secret, `MCP_DNS_PRIVATE_KEY`).
2. Обновите `.env.prod` на VPS + GitHub Secrets.
3. Продолжайте с шагом Б.

### Случай Б: убрать из истории через git-filter-repo
```bash
pip install git-filter-repo

# Заменить конкретный файл целиком:
git filter-repo --path path/to/leaked/file --invert-paths

# Или заменить текст паттерном:
echo 'actual-leaked-value==>REDACTED' > replacements.txt
git filter-repo --replace-text replacements.txt

# После filter-repo origin удаляется — добавить заново:
git remote add origin git@github.com:Slava4123/promtlab.git
git push --force --all
git push --force --tags
```

⚠️ Force push перезапишет историю на GitHub. Все кто клонировал раньше —
останутся со старой версией. Для приватного репо с одним мейнтейнером — OK.

## 4. Чек-лист готовности к открытию

- [ ] `gitleaks detect` → 0 findings
- [ ] `trufflehog git --only-verified` → 0 verified leaks
- [ ] Все credentials, которые когда-либо могли попасть в репо, ротированы
- [ ] `.env.prod` на VPS обновлён новыми значениями
- [ ] GitHub Secrets актуальны (`MCP_DNS_PRIVATE_KEY`, `SMITHERY_TOKEN`,
      `AWESOME_MCP_PAT`, `VPS_SSH_KEY`, `SENTRY_AUTH_TOKEN`, и т.д.)
- [ ] `gosec ./...` в `promptvault/backend/` — 0 critical
- [ ] `staticcheck ./...` — 0 warnings в критичных пакетах
- [ ] `npm audit --audit-level=high` в `promptvault/frontend/` — 0 high
- [ ] `LICENSE.md` (FSL 1.1) присутствует
- [ ] `SECURITY.md` присутствует
- [ ] `README.md` опубличен и не содержит внутренних TODO
- [ ] Корневой `CLAUDE.md` и `promptvault/CLAUDE.md` проверены на утечки
      внутренних решений/реквизитов (банковские данные, личные ИНН в
      .local.md не должны быть в git)
- [ ] `docs/archive/` проверен — нет PII, токенов, внутренней переписки
- [ ] GitHub → Settings → Security → Enable: Dependabot alerts, Code
      scanning (CodeQL), Secret scanning + Push Protection

## 5. После открытия

- Enable GitHub Advanced Security features (бесплатны для public).
- Добавить badge в README: CI status, Go report card, лицензия.
- Follow-up: bug bounty / responsible disclosure через `SECURITY.md`.
