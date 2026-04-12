package collection

import (
	"time"

	badgehttp "promptvault/internal/delivery/http/badge"
	"promptvault/internal/models"
)

// CollectionResponse — typed response для POST /api/collections и аналогичных
// mutating endpoints, которые должны возвращать newly_unlocked_badges.
// Другие endpoints (GetByID) продолжают возвращать map[string]any — менять их
// не требуется, пока нет unlock-логики на read-операциях.
type CollectionResponse struct {
	ID                  uint                     `json:"id"`
	Name                string                   `json:"name"`
	Description         string                   `json:"description,omitempty"`
	Color               string                   `json:"color"`
	Icon                string                   `json:"icon,omitempty"`
	TeamID              *uint                    `json:"team_id,omitempty"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	NewlyUnlockedBadges []badgehttp.BadgeSummary `json:"newly_unlocked_badges,omitempty"`
}

// NewCollectionResponse конвертит domain-объект в transport-DTO.
// newBadges может быть nil — omitempty скрывает поле из JSON.
func NewCollectionResponse(c models.Collection, newBadges []badgehttp.BadgeSummary) CollectionResponse {
	return CollectionResponse{
		ID:                  c.ID,
		Name:                c.Name,
		Description:         c.Description,
		Color:               c.Color,
		Icon:                c.Icon,
		TeamID:              c.TeamID,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		NewlyUnlockedBadges: newBadges,
	}
}
