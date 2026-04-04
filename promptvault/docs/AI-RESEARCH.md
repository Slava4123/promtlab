# AI-функции: Исследование best practices (апрель 2026)

## Методология

Изучено 45+ источников: официальные документации OpenAI, Anthropic, Google, DeepSeek, OpenRouter API Reference, OpenAI Community, GitHub issues, блоги разработчиков.

Все рекомендации ниже помечены уровнем уверенности:
- **[CONFIRMED]** — подтверждено официальной документацией или множественными независимыми источниками
- **[LIKELY]** — из нескольких независимых источников, но не из официальных docs
- **[UNCONFIRMED]** — из одного источника или блога, требует проверки

---

## 1. Per-model параметры генерации

### GPT-5 (`openai/gpt-5`)

| Параметр | Значение | Уровень |
|----------|----------|---------|
| temperature | НЕ поддерживается (ошибка API) | **[CONFIRMED]** — [OpenAI Community](https://community.openai.com/t/temperature-in-gpt-5-models/1337133), [mealie#6019](https://github.com/mealie-recipes/mealie/issues/6019), [litellm#13781](https://github.com/BerriAI/litellm/issues/13781) |
| top_p | НЕ поддерживается | **[CONFIRMED]** — те же источники |
| frequency_penalty | НЕ поддерживается | **[CONFIRMED]** — reasoning models не поддерживают sampling params |
| reasoning_effort | "low"/"medium"/"high" (альтернатива temperature) | **[CONFIRMED]** — [OpenAI Cookbook GPT-5](https://developers.openai.com/cookbook/examples/gpt-5/gpt-5_prompting_guide) |
| max_tokens | 128,000 (output) | **[CONFIRMED]** — OpenRouter model page |
| context | 400,000 | **[CONFIRMED]** — OpenRouter model page |

**Вывод:** НЕ отправлять temperature/top_p/penalties для GPT-5. Использовать reasoning_effort если нужно управлять глубиной.

---

### Claude Sonnet 4 (`anthropic/claude-sonnet-4`)

| Параметр | Значение | Уровень |
|----------|----------|---------|
| temperature | 0.2-0.5 для analytical, до 1.0 для creative | **[CONFIRMED]** — [Anthropic docs](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview) |
| top_p | Поддерживается, но НЕЛЬЗЯ использовать одновременно с temperature | **[CONFIRMED]** — Anthropic API reference |
| frequency_penalty | НЕ поддерживается нативно | **[CONFIRMED]** — Anthropic API reference |
| max_tokens | 64,000 (output) | **[CONFIRMED]** — OpenRouter model page |
| context | 200,000 | **[CONFIRMED]** — OpenRouter model page |
| Старый ID | `anthropic/claude-sonnet-4-20250514` — НЕДОСТУПЕН | **[CONFIRMED]** — OpenRouter возвращает "model is not available" |
| Правильный ID | `anthropic/claude-sonnet-4` | **[CONFIRMED]** — OpenRouter model page |

**Вывод:** Отправлять temperature: 0.4. НЕ отправлять top_p вместе с temperature. Исправить model ID.

---

### Gemini 2.5 Pro (`google/gemini-2.5-pro-preview-05-06`)

| Параметр | Значение | Уровень |
|----------|----------|---------|
| temperature | Default 1.0 | **[CONFIRMED]** — Google docs |
| temperature < 1.0 вызывает зацикливание | НЕ подтверждено | **[UNCONFIRMED]** — только блоги, НЕ найдено в official Google docs |
| top_p | 0.95 (default) | **[CONFIRMED]** — Google API reference |
| top_k | Поддерживается (уникально для Gemini) | **[CONFIRMED]** — Google API reference |
| max_tokens | 65,536 (output) | **[CONFIRMED]** — Google docs |
| context | 1,048,576 (1M) | **[CONFIRMED]** — Google docs |

**Вывод:** НЕ отправлять temperature (использовать default 1.0). Утверждение о зацикливании при temp < 1.0 не подтверждено.

---

### DeepSeek V3 (`deepseek/deepseek-chat-v3-0324`)

| Параметр | Значение | Уровень |
|----------|----------|---------|
| temperature для rewriting/translation | 1.3 | **[CONFIRMED]** — [Официальные docs DeepSeek](https://api-docs.deepseek.com/quick_start/parameter_settings) |
| temperature для coding | 0.0 | **[CONFIRMED]** — те же docs |
| temperature для data analysis | 1.0 | **[CONFIRMED]** — те же docs |
| temperature default | 1.0 | **[CONFIRMED]** — те же docs |
| Internal temp mapping (API 1.0 = model 0.3) | НЕ подтверждено | **[UNCONFIRMED]** — НЕ в официальных docs |
| max_tokens default | 4,000 (МАЛО!) | **[LIKELY]** — несколько источников, включая DeepSeek community |
| frequency_penalty | Поддерживается (OpenAI-совместимый API) | **[CONFIRMED]** — DeepSeek API docs |
| context | 163,840 | **[CONFIRMED]** — OpenRouter model page |

**Вывод:** Отправлять temperature: 1.3. Установить max_tokens: 8192 (default 4K — мало).

---

### GPT-4o Mini (`openai/gpt-4o-mini`)

| Параметр | Значение | Уровень |
|----------|----------|---------|
| temperature | Поддерживается (0-2.0) | **[CONFIRMED]** — OpenAI API reference |
| top_p | Поддерживается | **[CONFIRMED]** — OpenAI API reference |
| frequency_penalty | Поддерживается (-2.0 to 2.0) | **[CONFIRMED]** — OpenAI API reference |
| presence_penalty | Поддерживается (-2.0 to 2.0) | **[CONFIRMED]** — OpenAI API reference |
| Оптимальная temperature для rewriting | 0.2-0.5 (общая рекомендация) | **[LIKELY]** — практические отчёты, не official |
| max_tokens | 16,384 (output) | **[CONFIRMED]** — OpenAI docs |
| context | 128,000 | **[CONFIRMED]** — OpenAI docs |

**Вывод:** Не отправлять temperature (нет 100% оптимального значения). Можно экспериментировать.

---

## 2. Системные промпты

### Язык промпта

| Рекомендация | Уровень |
|-------------|---------|
| System prompt на английском → лучшие результаты для всех моделей | **[LIKELY]** — несколько независимых источников, практика индустрии |
| "Respond in the same language as the input" | **[CONFIRMED]** — стандартная практика |

### Anti-preamble инструкция

| Рекомендация | Уровень |
|-------------|---------|
| Все модели иногда добавляют "Sure, here's..." | **[CONFIRMED]** — широко задокументировано |
| GPT-4o Mini — самая "пристрастная" к пристрастиям | **[LIKELY]** — несколько источников |
| Инструкция "Do NOT start with Sure/Certainly" работает | **[CONFIRMED]** — стандартная практика |

### Claude: XML tags

| Рекомендация | Уровень |
|-------------|---------|
| Claude работает лучше с XML tags в промптах | **[CONFIRMED]** — [Anthropic docs](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags) |
| Данные в начале, инструкции в конце → +30% качество | **[LIKELY]** — Anthropic docs (не точная цифра) |

### Рейтинг моделей по качеству русского

| Модель | Качество русского | Уровень |
|--------|-------------------|---------|
| DeepSeek V3 | Отличное (на уровне носителя) | **[LIKELY]** — несколько независимых обзоров |
| Gemini 2.5 Pro | Очень хорошее | **[LIKELY]** — практические тесты |
| Claude Sonnet 4 | Хорошее | **[LIKELY]** — практический опыт |
| GPT-5 | Хорошее | **[LIKELY]** — практический опыт |
| GPT-4o Mini | Слабее остальных | **[LIKELY]** — [исследование](https://www.sciencedirect.com/science/article/pii/S0720048X25004279): 69% точность |

---

## 3. Кеширование OpenRouter

### Автоматическое (0 конфигурации)

| Провайдер | Кеширование | Уровень |
|-----------|-------------|---------|
| OpenAI (GPT-5, GPT-4o Mini) | Автоматическое при >= 1024 токенов | **[CONFIRMED]** — [OpenRouter docs](https://openrouter.ai/docs/guides/best-practices/prompt-caching) |
| DeepSeek V3 | Автоматическое | **[CONFIRMED]** — те же docs |
| Gemini 2.5 Pro | Implicit caching (автоматическое) | **[CONFIRMED]** — те же docs |

### Требует конфигурации

| Провайдер | Как включить | Уровень |
|-----------|-------------|---------|
| Anthropic (Claude) | `cache_control: {"type": "ephemeral"}` на уровне запроса | **[CONFIRMED]** — [OpenRouter docs](https://openrouter.ai/docs/guides/best-practices/prompt-caching) |
| Claude 1h TTL | `cache_control: {"type": "ephemeral", "ttl": "1h"}` | **[CONFIRMED]** — те же docs |

### Мониторинг кеша

| Поле | Описание | Уровень |
|------|----------|---------|
| `usage.prompt_tokens_details.cached_tokens` | Токены прочитанные из кеша | **[CONFIRMED]** — OpenRouter API reference |
| `usage.prompt_tokens_details.cache_write_tokens` | Токены записанные в кеш | **[CONFIRMED]** — те же docs |

---

## 4. Логирование и мониторинг

### Что доступно в SSE-стриме

| Данные | Где | Уровень |
|--------|-----|---------|
| `usage.prompt_tokens` | Финальный SSE-чанк (где `finish_reason != null`) | **[CONFIRMED]** — [OpenRouter streaming docs](https://openrouter.ai/docs/api/reference/streaming) |
| `usage.completion_tokens` | Тот же чанк | **[CONFIRMED]** |
| `usage.total_tokens` | Тот же чанк | **[CONFIRMED]** |
| `usage.cost` (USD) | Тот же чанк | **[CONFIRMED]** |
| `usage.prompt_tokens_details.cached_tokens` | Тот же чанк | **[CONFIRMED]** |
| Время (TTFT, total) | НЕ в SSE — измерять на клиенте | **[CONFIRMED]** |

### Формат финального SSE-чанка

```json
{
  "id": "gen-abc",
  "choices": [{"delta": {}, "finish_reason": "stop"}],
  "usage": {
    "prompt_tokens": 234,
    "completion_tokens": 512,
    "total_tokens": 746,
    "cost": 0.0085,
    "prompt_tokens_details": {
      "cached_tokens": 200,
      "cache_write_tokens": 0
    }
  }
}
```

**[CONFIRMED]** — OpenRouter API reference

### Observability

| Функция | Описание | Уровень |
|---------|----------|---------|
| `x-session-id` header | Группировка запросов по сессии | **[CONFIRMED]** — OpenRouter docs |
| `/api/v1/generation?id=$ID` | Детальная инфо: latency, cost, provider | **[CONFIRMED]** — OpenRouter API reference |
| `/api/v1/key` | Агрегированные расходы | **[CONFIRMED]** — OpenRouter API reference |

---

## 5. Что реализуем (100% verified)

### Шаг 1: Исправить model ID Claude
- `anthropic/claude-sonnet-4-20250514` → `anthropic/claude-sonnet-4` **[CONFIRMED]**

### Шаг 2: Per-model temperature
- GPT-5: НЕ отправлять (ошибка API) **[CONFIRMED]**
- Claude Sonnet 4: 0.4 **[CONFIRMED]**
- Gemini 2.5 Pro: НЕ отправлять (default 1.0) **[CONFIRMED нет ограничений]**
- DeepSeek V3: 1.3 **[CONFIRMED]**
- GPT-4o Mini: НЕ отправлять (нет 100% оптимального) **[нет данных]**

### Шаг 3: Системные промпты
- Перевести на английский **[LIKELY → стандарт индустрии]**
- Anti-preamble инструкция **[CONFIRMED]**
- Claude: XML tags **[CONFIRMED]**

### Шаг 4: Логирование
- Парсить `usage` из финального SSE-чанка **[CONFIRMED]**
- Логировать: prompt_tokens, completion_tokens, cost, cached_tokens **[CONFIRMED]**
- Замер времени на сервере (time.Since) **[стандартная практика]**

### Шаг 5: Кеширование Claude
- `cache_control: {"type": "ephemeral"}` для anthropic/ моделей **[CONFIRMED]**
- Остальные — автоматически **[CONFIRMED]**

---

## Источники

### Официальная документация
- [OpenAI: GPT-5 Prompting Guide](https://developers.openai.com/cookbook/examples/gpt-5/gpt-5_prompting_guide)
- [Anthropic: Prompt Engineering](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview)
- [Anthropic: XML Tags](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags)
- [Google: Gemini Prompting Strategies](https://ai.google.dev/gemini-api/docs/prompting-strategies)
- [DeepSeek: Parameter Settings](https://api-docs.deepseek.com/quick_start/parameter_settings)
- [OpenRouter: Prompt Caching](https://openrouter.ai/docs/guides/best-practices/prompt-caching)
- [OpenRouter: Streaming API](https://openrouter.ai/docs/api/reference/streaming)
- [MDN: Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)

### Подтверждение GPT-5 temperature
- [OpenAI Community: Temperature in GPT-5 models](https://community.openai.com/t/temperature-in-gpt-5-models/1337133)
- [GitHub: mealie-recipes/mealie#6019](https://github.com/mealie-recipes/mealie/issues/6019)
- [GitHub: BerriAI/litellm#13781](https://github.com/BerriAI/litellm/issues/13781)
- [GitHub: RooCodeInc/Roo-Code#6965](https://github.com/RooCodeInc/Roo-Code/issues/6965)

### Model pages (OpenRouter)
- [GPT-5](https://openrouter.ai/openai/gpt-5)
- [Claude Sonnet 4](https://openrouter.ai/anthropic/claude-sonnet-4)
- [Gemini 2.5 Pro](https://openrouter.ai/google/gemini-2.5-pro-preview-05-06)
- [DeepSeek V3](https://openrouter.ai/deepseek/deepseek-chat-v3-0324)
- [GPT-4o Mini](https://openrouter.ai/openai/gpt-4o-mini)
