package prompt

import (
	"time"

	badgehttp "promptvault/internal/delivery/http/badge"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type PromptResponse struct {
	ID             uint                 `json:"id"`
	Title          string               `json:"title"`
	Content        string               `json:"content"`
	Model          string               `json:"model,omitempty"`
	Favorite       bool                 `json:"favorite"`
	PinnedPersonal bool                 `json:"pinned_personal"`
	PinnedTeam     bool                 `json:"pinned_team"`
	UsageCount     int                  `json:"usage_count"`
	LastUsedAt     *time.Time           `json:"last_used_at,omitempty"`
	Tags           []TagResponse        `json:"tags"`
	Collections    []CollectionResponse `json:"collections"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	// NewlyUnlockedBadges — заполняется в mutating handlers (Create/Update/Favorite/Revert/etc)
	// после badges.Evaluate. omitempty гарантирует backward-compat: при отсутствии unlocks
	// поле не появляется в JSON, и старый клиент ничего не заметит.
	NewlyUnlockedBadges []badgehttp.BadgeSummary `json:"newly_unlocked_badges,omitempty"`
}

// IncrementUsageResponse — typed response для POST /api/prompts/{id}/use.
// Раньше возвращался map[string]string{"message":"ok"}, теперь typed — для
// консистентности с остальными mutating endpoints, несущими newly_unlocked_badges.
type IncrementUsageResponse struct {
	Message             string                   `json:"message"`
	NewlyUnlockedBadges []badgehttp.BadgeSummary `json:"newly_unlocked_badges,omitempty"`
}

type TagResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CollectionResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon,omitempty"`
}

func NewPromptResponse(p models.Prompt, pinStatus ...repo.PinStatus) PromptResponse {
	tags := make([]TagResponse, 0, len(p.Tags))
	for _, t := range p.Tags {
		tags = append(tags, TagResponse{ID: t.ID, Name: t.Name, Color: t.Color})
	}

	cols := make([]CollectionResponse, 0, len(p.Collections))
	for _, c := range p.Collections {
		cols = append(cols, CollectionResponse{ID: c.ID, Name: c.Name, Color: c.Color, Icon: c.Icon})
	}

	resp := PromptResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		Model:       p.Model,
		Favorite:    p.Favorite,
		UsageCount:  p.UsageCount,
		LastUsedAt:  p.LastUsedAt,
		Tags:        tags,
		Collections: cols,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if len(pinStatus) > 0 {
		resp.PinnedPersonal = pinStatus[0].PinnedPersonal
		resp.PinnedTeam = pinStatus[0].PinnedTeam
	}

	return resp
}

func NewPromptListResponse(prompts []models.Prompt, pinStatuses map[uint]repo.PinStatus) []PromptResponse {
	res := make([]PromptResponse, 0, len(prompts))
	for _, p := range prompts {
		res = append(res, NewPromptResponse(p, pinStatuses[p.ID]))
	}
	return res
}

type VersionResponse struct {
	ID            uint      `json:"id"`
	VersionNumber uint      `json:"version_number"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Model         string    `json:"model,omitempty"`
	ChangeNote    string    `json:"change_note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewVersionResponse(v models.PromptVersion) VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		VersionNumber: v.VersionNumber,
		Title:         v.Title,
		Content:       v.Content,
		Model:         v.Model,
		ChangeNote:    v.ChangeNote,
		CreatedAt:     v.CreatedAt,
	}
}

type UsageLogResponse struct {
	ID       uint           `json:"id"`
	PromptID uint           `json:"prompt_id"`
	Prompt   PromptResponse `json:"prompt"`
	UsedAt   time.Time      `json:"used_at"`
}

func NewUsageLogListResponse(logs []models.PromptUsageLog) []UsageLogResponse {
	res := make([]UsageLogResponse, 0, len(logs))
	for _, l := range logs {
		res = append(res, UsageLogResponse{
			ID:       l.ID,
			PromptID: l.PromptID,
			Prompt:   NewPromptResponse(l.Prompt),
			UsedAt:   l.UsedAt,
		})
	}
	return res
}

func NewVersionListResponse(versions []models.PromptVersion) []VersionResponse {
	res := make([]VersionResponse, 0, len(versions))
	for _, v := range versions {
		res = append(res, NewVersionResponse(v))
	}
	return res
}
