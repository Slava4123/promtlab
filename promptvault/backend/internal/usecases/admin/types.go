package admin

import (
	repo "promptvault/internal/interface/repository"
)

// UserListFilter — параметры admin.Service.ListUsers. Тонкая обёртка над
// repo.UserListFilter для развязки HTTP-слоя от DB-слоя (в будущем можем
// добавить валидацию/normalization в конверторе).
type UserListFilter = repo.UserListFilter

// UserListResult — результат ListUsers.
type UserListResult struct {
	Items    []repo.UserSummary
	Total    int64
	Page     int
	PageSize int
}
