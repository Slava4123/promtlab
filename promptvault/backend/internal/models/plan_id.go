package models

// PlanID — тарифный план юзера. Hashring-устойчивый enum-style тип:
// валидация через IsValid(), специализированные предикаты IsMax/IsPaid.
//
// MJ-25: до этого fix'а PlanID был голым `string` во всех слоях:
// - models.User.PlanID, models.Subscription.PlanID — string
// - usecases/subscription magic strings ("free", "pro", "max")
// - delivery/http/subscription/request.go без `oneof` validator
//
// Тип-alias на string совместим с существующими GORM-тегами и JSON
// сериализацией — не требует миграции схемы или массового rewrite
// callsites. Constants заменяют magic strings; helpers инкапсулируют
// проверки tier-gate (см. quotas.IsMaxTierUser).
type PlanID string

// Допустимые plan IDs. Должны совпадать с записями в `subscription_plans`
// таблице (миграция 000045). Список синхронизирован с oneof в
// delivery/http/subscription/request.go.
const (
	PlanFree     PlanID = "free"
	PlanPro      PlanID = "pro"
	PlanMax      PlanID = "max"
	PlanProYear  PlanID = "pro_yearly"
	PlanMaxYear  PlanID = "max_yearly"
)

// IsValid возвращает true если plan ID — один из допустимых.
// Используется как defensive guard перед сохранением (например, в audit).
func (p PlanID) IsValid() bool {
	switch p {
	case PlanFree, PlanPro, PlanMax, PlanProYear, PlanMaxYear:
		return true
	}
	return false
}

// IsMax возвращает true для тарифов Max-уровня (включая годовые).
// Используется в tier-gate'ах (Phase 16 Conditional Chains, fork-share,
// extended retention, Smart Insights).
func (p PlanID) IsMax() bool {
	return p == PlanMax || p == PlanMaxYear
}

// IsPaid возвращает true для платных планов (любой кроме free).
func (p PlanID) IsPaid() bool {
	return p != PlanFree
}

// String — для удобного логирования и интерполяции в SQL/JSON.
func (p PlanID) String() string { return string(p) }
