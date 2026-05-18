package config

// ReferralConfig — Pricing iteration v3 (ADR-0009). Реферальная награда
// +30 дней Pro пригласившему после первого платежа реферри.
//
// При false: subscription webhook не пишет pending'ы в referral_pending_rewards,
// а ReferralRewardLoop не запускается (lifecycle.go). Включается отдельным
// flag'ом после 1 недели QA — независимо от Pricing Wave 1 (Free quota)
// и Wave 2 (Pro insights teaser), чтобы при необходимости можно было
// откатить только реферальную программу.
type ReferralConfig struct {
	RewardEnabled bool `koanf:"reward_enabled"`
}
