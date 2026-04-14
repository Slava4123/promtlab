package repository

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

const planCacheTTL = 5 * time.Minute

type planRepo struct {
	db *gorm.DB

	mu        sync.RWMutex
	cache     []models.SubscriptionPlan
	cacheTime time.Time
}

func NewPlanRepository(db *gorm.DB) *planRepo {
	return &planRepo{db: db}
}

func (r *planRepo) cachedPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	r.mu.RLock()
	if time.Since(r.cacheTime) < planCacheTTL && r.cache != nil {
		plans := r.cache
		r.mu.RUnlock()
		return plans, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// double-check after lock
	if time.Since(r.cacheTime) < planCacheTTL && r.cache != nil {
		return r.cache, nil
	}

	var plans []models.SubscriptionPlan
	if err := r.db.WithContext(ctx).Order("sort_order").Find(&plans).Error; err != nil {
		return nil, err
	}

	r.cache = plans
	r.cacheTime = time.Now()
	return plans, nil
}

func (r *planRepo) GetAll(ctx context.Context) ([]models.SubscriptionPlan, error) {
	return r.cachedPlans(ctx)
}

func (r *planRepo) GetByID(ctx context.Context, id string) (*models.SubscriptionPlan, error) {
	plans, err := r.cachedPlans(ctx)
	if err != nil {
		return nil, err
	}
	for i := range plans {
		if plans[i].ID == id {
			return &plans[i], nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *planRepo) GetActive(ctx context.Context) ([]models.SubscriptionPlan, error) {
	plans, err := r.cachedPlans(ctx)
	if err != nil {
		return nil, err
	}
	active := make([]models.SubscriptionPlan, 0, len(plans))
	for _, p := range plans {
		if p.IsActive {
			active = append(active, p)
		}
	}
	return active, nil
}
