package starter

import "errors"

// ErrAlreadyCompleted — юзер уже прошёл онбординг. Клиент должен реагировать
// 409 Conflict — это защита от двойного клика и мутаций задним числом.
var ErrAlreadyCompleted = errors.New("онбординг уже пройден")

// ErrUnknownTemplate — переданный template_id отсутствует в catalog.json.
// Маппится на 400 Bad Request.
var ErrUnknownTemplate = errors.New("неизвестный шаблон")

// ErrUserNotFound — юзер из контекста не найден в БД (битый JWT?).
var ErrUserNotFound = errors.New("пользователь не найден")
