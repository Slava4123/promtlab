package apikey

import "time"

type CreateRequest struct {
	Name         string     `json:"name" validate:"required,min=1,max=100"`
	ReadOnly     bool       `json:"read_only,omitempty"`
	TeamID       *uint      `json:"team_id,omitempty"`
	AllowedTools []string   `json:"allowed_tools,omitempty" validate:"omitempty,dive,min=1,max=64"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
