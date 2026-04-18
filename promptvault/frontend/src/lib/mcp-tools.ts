// Все 24 MCP-tool'а + list_prompt_vars (фича B). Синхронизировано с
// backend/internal/usecases/apikey/constants.go.
// Используется в API-keys форме для выбора allowed_tools.

export interface McpTool {
  name: string
  group: "read" | "write" | "destructive"
  label: string
}

export const MCP_TOOLS: McpTool[] = [
  // Read
  { name: "search_prompts", group: "read", label: "Поиск промптов" },
  { name: "list_prompts", group: "read", label: "Список промптов" },
  { name: "get_prompt", group: "read", label: "Получить промпт" },
  { name: "list_collections", group: "read", label: "Список коллекций" },
  { name: "list_tags", group: "read", label: "Список тегов" },
  { name: "get_prompt_versions", group: "read", label: "История версий" },
  { name: "prompt_list_pinned", group: "read", label: "Закреплённые" },
  { name: "prompt_list_recent", group: "read", label: "Недавние" },
  { name: "collection_get", group: "read", label: "Получить коллекцию" },
  { name: "search_suggest", group: "read", label: "Автодополнение поиска" },
  { name: "list_prompt_vars", group: "read", label: "Переменные промпта" },
  // Write
  { name: "create_prompt", group: "write", label: "Создать промпт" },
  { name: "update_prompt", group: "write", label: "Обновить промпт" },
  { name: "create_tag", group: "write", label: "Создать тег" },
  { name: "create_collection", group: "write", label: "Создать коллекцию" },
  { name: "collection_update", group: "write", label: "Обновить коллекцию" },
  { name: "prompt_favorite", group: "write", label: "Избранное" },
  { name: "prompt_pin", group: "write", label: "Закрепить" },
  { name: "prompt_increment_usage", group: "write", label: "Отметить использование" },
  { name: "share_create", group: "write", label: "Создать share-ссылку" },
  { name: "prompt_revert", group: "write", label: "Откатить к версии" },
  // Destructive
  { name: "delete_prompt", group: "destructive", label: "Удалить промпт (в корзину)" },
  { name: "delete_collection", group: "destructive", label: "Удалить коллекцию" },
  { name: "share_deactivate", group: "destructive", label: "Отключить share-ссылку" },
  { name: "tag_delete", group: "destructive", label: "Удалить тег" },
]

export const MCP_TOOL_NAMES = MCP_TOOLS.map((t) => t.name)
