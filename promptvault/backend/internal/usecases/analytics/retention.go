package analytics

import (
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/usecases/subscription"
)

// MaxRangeDays возвращает максимально допустимое окно аналитики по тарифу.
// Phase 14: Free=7, Pro=90, Max=365. Используется в retention middleware
// и clamp-логике перед запросом в AnalyticsRepository.
func MaxRangeDays(planID string) int {
	switch subscription.Tier(planID) {
	case "max":
		return 365
	case "pro":
		return 90
	default:
		return 7
	}
}

// rangeToDays — mapping от RangeID к числу дней.
func rangeToDays(r RangeID) int {
	switch r {
	case Range7d:
		return 7
	case Range30d:
		return 30
	case Range90d:
		return 90
	case Range365d:
		return 365
	default:
		return 7
	}
}

// ClampRange возвращает RangeID, обрезанный по максимальному окну тарифа.
// Пример: Free + Range365d → Range7d. Pro + Range365d → Range90d.
func ClampRange(requested RangeID, planID string) RangeID {
	req := rangeToDays(requested)
	maxDays := MaxRangeDays(planID)
	if req <= maxDays {
		return requested
	}
	switch {
	case maxDays >= 365:
		return Range365d
	case maxDays >= 90:
		return Range90d
	case maxDays >= 30:
		return Range30d
	default:
		return Range7d
	}
}

// BuildDateRange возвращает полуоткрытый DateRange [now - N days, now)
// для заданного RangeID. Используется как единая точка "сейчас" в Service,
// чтобы тесты могли подменить now через параметр.
func BuildDateRange(id RangeID, now time.Time) repo.DateRange {
	days := rangeToDays(id)
	return repo.DateRange{
		From: now.AddDate(0, 0, -days),
		To:   now,
	}
}

// BuildPreviousRange — предыдущий период той же длины для сравнения.
// Например 7d сейчас → 7d до него: [now-14d, now-7d).
func BuildPreviousRange(id RangeID, now time.Time) repo.DateRange {
	days := rangeToDays(id)
	return repo.DateRange{
		From: now.AddDate(0, 0, -2*days),
		To:   now.AddDate(0, 0, -days),
	}
}

// sumPoints — суммирует counts по массиву точек.
// Используется для расчёта Totals из per-day arrays.
func sumPoints(points []repo.UsagePoint) int64 {
	var total int64
	for _, p := range points {
		total += p.Count
	}
	return total
}
