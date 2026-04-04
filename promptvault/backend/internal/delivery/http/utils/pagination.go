package utils

import "net/http"

type PaginatedResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	HasMore  bool  `json:"has_more"`
}

func WritePaginated[T any](w http.ResponseWriter, items []T, total int64, page, pageSize int) {
	resp := PaginatedResponse[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  int64(page*pageSize) < total,
	}
	WriteOK(w, resp)
}
