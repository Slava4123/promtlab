package changelog

// Entry — одна запись changelog'а. Хранится в embedded JSON.
type Entry struct {
	Version     string `json:"version"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// Changelog — десериализованный changelog.json.
type Changelog struct {
	Entries []Entry `json:"entries"`
}

// ChangelogOutput — ответ для клиента: список записей + флаг непрочитанных.
type ChangelogOutput struct {
	Entries   []Entry `json:"entries"`
	HasUnread bool    `json:"has_unread"`
}
