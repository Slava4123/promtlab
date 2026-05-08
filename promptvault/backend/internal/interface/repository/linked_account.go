package repository

import (
	"context"

	"promptvault/internal/models"
)

type LinkedAccountRepository interface {
	Create(ctx context.Context, la *models.LinkedAccount) error
	GetByUserID(ctx context.Context, userID uint) ([]models.LinkedAccount, error)
	GetByProviderID(ctx context.Context, provider, providerID string) (*models.LinkedAccount, error)
	Delete(ctx context.Context, userID uint, provider string) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)

	// DeleteIfMethodsRemain атомарно удаляет linked_account по (userID,
	// provider), но только если у юзера остаётся хотя бы один способ
	// входа после удаления.
	//
	// MJ-14: до этого метода Service.UnlinkProvider делал check-then-delete
	// (CountByUserID → arithmetic → Delete) — два concurrent unlink-запроса
	// могли пройти оба check одновременно и оставить юзера без login methods.
	//
	// hasPassword — true если у юзера есть password_hash (его наличие
	// означает что после удаления linked_account останется ≥ 1 способ
	// войти — через email+пароль).
	//
	// Возвращает (true, nil) если delete прошёл; (false, nil) если запись
	// не существовала ИЛИ удаление оставило бы юзера без login methods —
	// в этом случае caller должен вернуть ErrCannotUnlinkLast.
	DeleteIfMethodsRemain(ctx context.Context, userID uint, provider string, hasPassword bool) (bool, error)
}
