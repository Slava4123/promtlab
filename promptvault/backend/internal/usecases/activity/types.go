package activity

import "encoding/json"

// Event — описание одного события для записи в team_activity_log.
//
// Обязательно: TeamID, EventType, TargetType.
// Обязательно хотя бы один из способов идентификации актора:
//   - ActorID != 0 → ActorEmail/ActorName подтянутся автоматически через UserRepo;
//   - ActorEmail явно — тогда ActorID и не нужен (например, system-events).
//
// Metadata сериализуется в JSONB. nil/пустая map → NULL в БД.
type Event struct {
	TeamID      uint
	ActorID     uint   // 0 — актор не user (system), требуется ActorEmail
	ActorEmail  string // если пусто и ActorID != 0 → резолв через UserRepo
	ActorName   string
	EventType   string
	TargetType  string
	TargetID    *uint
	TargetLabel string
	Metadata    map[string]any
}

// marshalMetadata возвращает nil для пустой map, чтобы JSONB в БД остался NULL.
func marshalMetadata(m map[string]any) (json.RawMessage, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}
