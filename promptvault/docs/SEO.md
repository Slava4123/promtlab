# SEO + Discovery для публичных промптов

Production-ready инфраструктура для индексации публичных промптов поисковиками (Yandex, Google, Bing) и rich-превью при шаринге (Telegram, VK, Twitter, Slack, Discord).

## Архитектура

```
                    ┌─────────────────────────────┐
                    │      nginx (443 SSL)        │
                    │                             │
   bot User-Agent ──┤  map $http_user_agent $is_bot
   (Yandexbot etc.) │       ↓                     │
                    │  /p/<slug>:                 │
                    │    if $is_bot → api:8080    │──────┐
                    │    else → SPA index.html    │      │
                    │                             │      │
   обычный юзер ────┤  /sitemap.xml → api:8080    │──────┤
                    │  /api/og/* → api:8080       │──────┤
                    └─────────────────────────────┘      │
                                                         ▼
                                            ┌────────────────────────┐
                                            │  Go API (seo package)  │
                                            │  - PromptHTML (server- │
                                            │    rendered template)  │
                                            │  - Sitemap (XML cache) │
                                            │  - OGImage (PNG via    │
                                            │    fogleman/gg)        │
                                            └────────────────────────┘
```

## Endpoints

| Endpoint | Назначение | Cache |
|---|---|---|
| `/p/{slug}` (для bot-UA) | Server-rendered HTML с `<head>` (title, OG, Twitter Card, JSON-LD Article) и `<body>` с контентом промпта | 5 мин |
| `/sitemap.xml` | Список всех публичных промптов с `<loc>` + `<lastmod>` (формат sitemaps.org/0.9) | 1 ч in-memory |
| `/api/og/prompts/{slug}.png` | Динамический OG-image 1200×630 с заголовком промпта на фирменном градиенте | 24 ч + ETag |

## Bot detection

`nginx/nginx.conf::map $http_user_agent $is_bot` содержит regex для:
- **Search engines:** Googlebot, Bingbot, Yandexbot, DuckDuckBot, Baiduspider, Applebot
- **Social previews:** facebookexternalhit, Twitterbot, LinkedInBot, Slackbot, Telegrambot, WhatsApp, vkShare, DiscordBot, Pinterestbot, Tumblr, Redditbot, Embedly
- **AI crawlers (2026):** GPTBot, ClaudeBot, Claude-Web, PerplexityBot, Amazonbot

Чтобы добавить нового бота — открыть [Prerender.io official list](https://docs.prerender.io/docs/how-to-add-additional-bots) и расширить регексы в `nginx.conf`.

## Деплой и one-time ops

После первого деплоя в production выполнить вручную:

### 1. Yandex Webmaster
1. Открыть https://webmaster.yandex.ru/
2. Добавить домен `promtlabs.ru`
3. Подтвердить владение (HTML-файл или meta-tag)
4. Indexing → Sitemap files → Add → `https://promtlabs.ru/sitemap.xml`
5. Indexing → JS rendering → выбрать **«Recommend rendering»** (для SPA-страниц помимо `/p/*`)

### 2. Google Search Console
1. Открыть https://search.google.com/search-console
2. Add property → URL prefix → `https://promtlabs.ru`
3. Подтвердить (DNS TXT или HTML upload)
4. Sitemaps → Add a new sitemap → `sitemap.xml` → Submit
5. URL Inspection → ввести любой `/p/<slug>` → Request indexing (для прогрева)

### 3. Validation
- [Google Rich Results Test](https://search.google.com/test/rich-results?url=https://promtlabs.ru/p/<slug>) → должен распознать **Article** schema
- [Twitter Card Validator](https://cards-dev.twitter.com/validator)
- [Facebook Sharing Debugger](https://developers.facebook.com/tools/debug/) → проверить OG для Telegram/VK
- [VK Share Preview](https://share.vk.com/?url=https://promtlabs.ru/p/<slug>)

### 4. Telegram cache invalidation

При изменении title/content уже опубликованного промпта:
- Боты Telegram кэшируют OG **до 7 дней**
- Force-refresh: написать `@WebpageBot` ссылку → бот сбросит кэш
- Альтернатива: сменить slug (новая URL — новая запись в кэше)

## Verification (curl smoke)

```bash
# Sitemap
curl -sS https://promtlabs.ru/sitemap.xml | head -20
curl -sS -I https://promtlabs.ru/sitemap.xml | grep -E "Cache-Control|Content-Type"

# Server-HTML для bot-UA
curl -sS -A "Mozilla/5.0 (compatible; Yandexbot/3.0)" https://promtlabs.ru/p/<slug> \
  | grep -E '<title>|og:title|application/ld\+json'

# Server-HTML НЕ для bot-UA — должен быть пустой <div id=root>
curl -sS -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36" \
  https://promtlabs.ru/p/<slug> | grep -c '<div id="root">'  # → 1

# OG-image
curl -sS -I https://promtlabs.ru/api/og/prompts/<slug>.png | grep -E "Content-Type|ETag|Cache-Control"
```

## Метрики и логи

slog-события (доступны через GlitchTip + grep по prod-логам):

| Event | Уровень | Поля |
|---|---|---|
| `seo.html.served` | INFO | slug, ua, duration_ms |
| `seo.html.not_found` | WARN | slug |
| `seo.html.lookup_failed` | ERROR | slug, error |
| `seo.html.render_failed` | ERROR | slug, error |
| `seo.sitemap.served` | INFO | bytes, duration_ms |
| `seo.sitemap.failed` | ERROR | error |
| `seo.sitemap.size_warning` | WARN | count, max (когда близко к 50K) |

5xx-ошибки автоматически захватываются Sentry middleware (через `RespondWithRequest`) с user-context при наличии.

## Лимиты и масштаб

- **Sitemap:** 50K URL / 50MB (sitemaps.org spec). Текущий лимит репо `ListPublic(10000)`. При приближении — мигрировать на sitemap-index с chunked-файлами.
- **OG-image:** ETag сильно снижает реальный рендер. Без ETag типичный рендер — ~50ms. Память: ~16MB на горутину при рендере (gradient + truetype).
- **Rate limit:** 60 req/min/IP на каждом SEO endpoint (`byIP(60)` middleware).

## Future work (не реализовано в MVP)

- **410 Gone** для приватизированных промптов (сейчас возвращается 404). Требует дополнительный query "найди slug регardless of is_public", ~30 строк.
- **Author** в JSON-LD — сейчас только `publisher: ПромтЛаб`. Добавить `Preload("User")` в `GetPublicBySlug` + поле `author: Person` в schema. ~15 строк.
- **Sitemap-index** при росте >10K публичных промптов. Один день работы при необходимости.
- **OG-image для коллекций / профилей** — когда появятся публичные коллекции и профили авторов.
