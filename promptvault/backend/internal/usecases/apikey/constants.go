package apikey

// KnownTools — полный список имён MCP-tool'ов, разрешённых в allowed_tools.
// Должен совпадать с tools.go регистрацией в mcpserver (25 имён после
// добавления list_prompt_vars в v1.1).
// При добавлении нового tool — обновить здесь, в mcpserver/tools.go и в
// frontend/src/lib/mcp-tools.ts (иначе UI-форма не покажет tool в выборе).
var KnownTools = map[string]bool{
	// Read
	"search_prompts":      true,
	"list_prompts":        true,
	"get_prompt":          true,
	"list_collections":    true,
	"list_tags":           true,
	"get_prompt_versions": true,
	"prompt_list_pinned":  true,
	"prompt_list_recent":  true,
	"collection_get":      true,
	"search_suggest":      true,
	"list_prompt_vars":    true, // добавлен в фиче B
	// Write
	"create_prompt":          true,
	"update_prompt":          true,
	"delete_prompt":          true,
	"create_tag":             true,
	"create_collection":      true,
	"delete_collection":      true,
	"prompt_favorite":        true,
	"prompt_pin":             true,
	"prompt_increment_usage": true,
	"share_create":           true,
	"collection_update":      true,
	"prompt_revert":          true,
	"share_deactivate":       true,
	"tag_delete":             true,
}

// IsKnownTool проверяет, существует ли tool с таким именем.
func IsKnownTool(name string) bool {
	return KnownTools[name]
}
