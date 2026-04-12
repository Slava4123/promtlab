package changelog

import changeloguc "promptvault/internal/usecases/changelog"

// ChangelogResponse — ответ GET /api/changelog.
type ChangelogResponse struct {
	Entries   []EntryResponse `json:"entries"`
	HasUnread bool            `json:"has_unread"`
}

// EntryResponse — одна запись changelog'а в ответе.
type EntryResponse struct {
	Version     string `json:"version"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func NewChangelogResponse(out *changeloguc.ChangelogOutput) ChangelogResponse {
	entries := make([]EntryResponse, 0, len(out.Entries))
	for _, e := range out.Entries {
		entries = append(entries, EntryResponse{
			Version:     e.Version,
			Date:        e.Date,
			Title:       e.Title,
			Category:    e.Category,
			Description: e.Description,
		})
	}
	return ChangelogResponse{
		Entries:   entries,
		HasUnread: out.HasUnread,
	}
}
