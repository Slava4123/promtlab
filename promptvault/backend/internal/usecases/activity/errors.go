package activity

import "errors"

var (
	ErrMissingTeam      = errors.New("activity: team_id is required")
	ErrMissingEventType = errors.New("activity: event_type and target_type are required")
	ErrMissingActor     = errors.New("activity: actor not resolvable (no actor_id and no actor_email)")
)
