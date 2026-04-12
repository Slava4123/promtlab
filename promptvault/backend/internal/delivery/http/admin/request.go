package admin

// ListUsersRequest — параметры GET /api/admin/users (query string).
// Page начинается с 1, PageSize clamped в usecase/repo (20 default, 100 max).
type ListUsersRequest struct {
	Query    string `form:"q"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	SortBy   string `form:"sort"`
	SortDesc bool   `form:"desc"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// ChangeTierRequest — тело POST /api/admin/users/{id}/tier.
// Tier — один из заранее известных значений ("free", "pro", "max").
// В MVP handler возвращает 501 т.к. subscription system отсутствует.
type ChangeTierRequest struct {
	Tier     string `json:"tier" validate:"required,oneof=free pro max"`
	TOTPCode string `json:"totp_code" validate:"required"`
}

// TOTPCodeRequest — базовая структура для destructive actions, требующих
// fresh TOTP verification (sudo mode). Re-verification на каждое destructive
// action вместо TTL-based approach — проще и соответствует GitHub sudo pattern.
type TOTPCodeRequest struct {
	TOTPCode string `json:"totp_code" validate:"required"`
}
