# .github/scripts

Вспомогательные скрипты для GitHub Actions workflows.

## `add_awesome_mcp.py`

Идемпотентно вставляет запись про `Slava4123/promptvault` в README.md форка
`punkpeye/awesome-mcp-servers` под категорию «🧠 Knowledge & Memory» с сохранением
алфавитного порядка.

**Коды возврата:**
- `0` — строка добавлена, файл изменён
- `78` — запись уже присутствует, изменений нет (sentinel для workflow)
- `2` — ошибка аргументов или отсутствует README

**Локальный запуск:**
```bash
git clone https://github.com/Slava4123/awesome-mcp-servers /tmp/awesome
python3 .github/scripts/add_awesome_mcp.py /tmp/awesome
```
