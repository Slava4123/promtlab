package badge

import "errors"

// ErrCatalogLoad — невалидный catalog.json (ошибка парсинга или валидации
// при старте). Возвращается из LoadCatalog/NewService, триггерит panic в
// app.go (bootstrap failure — как в starter/changelog usecases).
var ErrCatalogLoad = errors.New("badge catalog load failed")

// ErrUnknownCondition — в catalog.json попал неизвестный ConditionType.
// Валидируется при загрузке каталога, не возникает в runtime.
var ErrUnknownCondition = errors.New("unknown badge condition type")
