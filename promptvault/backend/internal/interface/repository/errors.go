package repository

import "errors"

// ErrNotFound — запись не найдена в хранилище.
// Используется вместо gorm.ErrRecordNotFound чтобы usecase слой не зависел от gorm.
var ErrNotFound = errors.New("record not found")

// ErrConflict — конкурентная запись или нарушение оптимистичного guard.
// Возвращается когда conditional UPDATE затронул 0 строк (например, попытка
// пометить юзера прошедшим онбординг, когда он уже помечен другой транзакцией).
// Usecase слой маппит её в свою доменную ошибку (ErrAlreadyCompleted и т.п.).
var ErrConflict = errors.New("conflict")
