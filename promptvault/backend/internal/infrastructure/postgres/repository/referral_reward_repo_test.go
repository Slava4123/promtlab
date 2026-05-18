package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"promptvault/internal/models"
)

// TestReferralRewardRepo_CRUD — интеграционный тест полного жизненного цикла:
// Create → дубль (UNIQUE violation) → FindByReferee → ListEligible (до/после
// «refund-окна») → Delete. Использует реальные миграции через setupTestDB
// (CR-10), поэтому проверяет UNIQUE index по referee_id, FK на users/payments
// и индекс по eligible_at.
func TestReferralRewardRepo_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("requires postgres testcontainer")
	}
	ctx := context.Background()
	db := setupTestDB(t)
	r := NewReferralRewardRepository(db)

	// Сидим 2 юзеров + payment.
	referrer := &models.User{
		Email:        "referrer-reward@test.local",
		Name:         "Referrer",
		PlanID:       "pro",
		ReferralCode: "REFREF01",
	}
	referee := &models.User{
		Email:  "referee-reward@test.local",
		Name:   "Referee",
		PlanID: "pro",
	}
	require.NoError(t, db.Create(referrer).Error)
	require.NoError(t, db.Create(referee).Error)

	payment := &models.Payment{
		UserID:         referee.ID,
		AmountKop:      59900,
		Status:         models.PaymentSucceeded,
		Currency:       models.CurrencyRUB,
		Provider:       models.PaymentProviderTBank,
		ExternalID:     "ext-reward-1",
		IdempotencyKey: "idem-reward-1",
	}
	require.NoError(t, db.Create(payment).Error)

	now := time.Now()
	eligibleAt := now.Add(14 * 24 * time.Hour)

	// Create
	pending := &models.ReferralPendingReward{
		ReferrerID: referrer.ID,
		RefereeID:  referee.ID,
		PaymentID:  payment.ID,
		EligibleAt: eligibleAt,
	}
	require.NoError(t, r.Create(ctx, pending))
	require.NotZero(t, pending.ID)

	// Create idempotent (UNIQUE on referee_id) — повторный INSERT должен вернуть error.
	dup := &models.ReferralPendingReward{
		ReferrerID: referrer.ID,
		RefereeID:  referee.ID,
		PaymentID:  payment.ID,
		EligibleAt: eligibleAt,
	}
	err := r.Create(ctx, dup)
	require.Error(t, err, "должен быть UNIQUE violation на referee_id")

	// FindByReferee
	found, err := r.FindByReferee(ctx, referee.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, pending.ID, found.ID)
	require.Equal(t, referrer.ID, found.ReferrerID)
	require.Equal(t, payment.ID, found.PaymentID)

	// FindByReferee — not found: nil + nil (контракт для idempotency-check).
	notFound, err := r.FindByReferee(ctx, 9999)
	require.NoError(t, err)
	require.Nil(t, notFound)

	// ListEligible — eligibleAt в будущем (14d) → пусто.
	eligible, err := r.ListEligible(ctx, now, 10)
	require.NoError(t, err)
	require.Empty(t, eligible)

	// Симулируем «прошло 14 дней» — двигаем eligible_at в прошлое.
	require.NoError(t, db.Model(pending).Update("eligible_at", now.Add(-1*time.Minute)).Error)

	eligible, err = r.ListEligible(ctx, now, 10)
	require.NoError(t, err)
	require.Len(t, eligible, 1)
	require.Equal(t, pending.ID, eligible[0].ID)

	// Delete
	require.NoError(t, r.Delete(ctx, pending.ID))
	found, err = r.FindByReferee(ctx, referee.ID)
	require.NoError(t, err)
	require.Nil(t, found)
}
