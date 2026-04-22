package subscription

// Tier возвращает tier-ступень для plan_id: "free", "pro" или "max".
// Используется везде, где нужно принять решение "платный ли план" или
// "какой именно платный". Замена небезопасным prefix-проверкам
// вида strings.HasPrefix(planID, "pro") — те срабатывали бы на
// будущие planID вроде "professional", "proto", "maximus".
//
// Источник правды для plan_id — subscription_plans.id в БД.
// Известные значения на апрель 2026: free, pro, pro_yearly, max, max_yearly.
// Любой неизвестный plan_id трактуется как free (safe fallback).
func Tier(planID string) string {
	switch planID {
	case "pro", "pro_yearly":
		return "pro"
	case "max", "max_yearly":
		return "max"
	default:
		return "free"
	}
}

// IsPaid возвращает true для всех не-free тарифов.
func IsPaid(planID string) bool {
	return Tier(planID) != "free"
}

// IsMax возвращает true только для Max-тарифов.
func IsMax(planID string) bool {
	return Tier(planID) == "max"
}
