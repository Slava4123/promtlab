// Все 30 MCP-tool'ов в v1.2. Синхронизировано с
// backend/internal/usecases/apikey/constants.go.
// Используется в API-keys форме для выбора allowed_tools.

export interface McpTool {
  name: string
  group: "read" | "write" | "destructive"
  label: string
}

export const MCP_TOOLS: McpTool[] = [
  // Read
  { name: "whoami", group: "read", label: "Текущий пользователь" },
  { name: "search_prompts", group: "read", label: "Поиск промптов" },
  { name: "list_prompts", group: "read", label: "Список промптов" },
  { name: "get_prompt", group: "read", label: "Получить промпт" },
  { name: "list_collections", group: "read", label: "Список коллекций" },
  { name: "list_tags", group: "read", label: "Список тегов" },
  { name: "list_teams", group: "read", label: "Список команд" },
  { name: "list_trash", group: "read", label: "Содержимое корзины" },
  { name: "get_prompt_versions", group: "read", label: "История версий" },
  { name: "prompt_list_pinned", group: "read", label: "Закреплённые" },
  { name: "prompt_list_recent", group: "read", label: "Недавние" },
  { name: "collection_get", group: "read", label: "Получить коллекцию" },
  { name: "search_suggest", group: "read", label: "Автодополнение поиска" },
  { name: "list_prompt_vars", group: "read", label: "Переменные промпта" },
  { name: "team_activity_feed", group: "read", label: "Лента активности команды" },
  { name: "analytics_summary", group: "read", label: "Сводка аналитики (личная)" },
  { name: "analytics_team_summary", group: "read", label: "Сводка аналитики (команда)" },
  // Write
  { name: "create_prompt", group: "write", label: "Создать промпт" },
  { name: "update_prompt", group: "write", label: "Обновить промпт" },
  { name: "create_tag", group: "write", label: "Создать тег" },
  { name: "create_collection", group: "write", label: "Создать коллекцию" },
  { name: "collection_update", group: "write", label: "Обновить коллекцию" },
  { name: "prompt_favorite", group: "write", label: "Избранное" },
  { name: "prompt_pin", group: "write", label: "Закрепить" },
  { name: "prompt_increment_usage", group: "write", label: "Отметить использование" },
  { name: "share_create", group: "write", label: "Создать публичную ссылку" },
  { name: "prompt_revert", group: "write", label: "Откатить к версии" },
  { name: "restore_prompt", group: "write", label: "Восстановить из корзины" },
  // Destructive
  { name: "delete_prompt", group: "destructive", label: "Удалить промпт (в корзину)" },
  { name: "delete_collection", group: "destructive", label: "Удалить коллекцию" },
  { name: "share_deactivate", group: "destructive", label: "Отключить публичную ссылку" },
  { name: "tag_delete", group: "destructive", label: "Удалить тег" },
  { name: "purge_prompt", group: "destructive", label: "Удалить навсегда (из корзины)" },
]

export const MCP_TOOL_NAMES = MCP_TOOLS.map((t) => t.name)
