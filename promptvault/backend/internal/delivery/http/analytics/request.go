package analytics

import (
	analyticsuc "promptvault/internal/usecases/analytics"
)

// parseRange — query param "range" → RangeID с фолбэком "7d".
// Неизвестные значения также обрезаются в Range7d (safe default).
func parseRange(raw string) analyticsuc.RangeID {
	switch raw {
	case "30d":
		return analyticsuc.Range30d
	case "90d":
		return analyticsuc.Range90d
	case "365d":
		return analyticsuc.Range365d
	default:
		return analyticsuc.Range7d
	}
}
