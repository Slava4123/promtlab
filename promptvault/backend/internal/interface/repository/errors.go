package repository

import "errors"

// ErrNotFound — запись не найдена в хранилище.
// Используется вместо gorm.ErrRecordNotFound чтобы usecase слой не зависел от gorm.
var ErrNotFound = errors.New("record not found")
