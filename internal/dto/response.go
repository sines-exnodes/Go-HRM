package dto

import "math"

// Response is the standard success envelope. T is the data type.
//   { "success": true, "message": "...", "data": T }
type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// PaginatedData wraps a list result with pagination metadata. Embed inside a
// Response[PaginatedData[T]] for the canonical paginated payload.
type PaginatedData[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// NewResponse builds a successful Response[T] with the given data and an
// optional message.
func NewResponse[T any](data T, message string) Response[T] {
	return Response[T]{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewPaginatedResponse builds the canonical paginated envelope. page and
// pageSize must be positive; the caller is responsible for validating them
// before invoking this helper.
func NewPaginatedResponse[T any](items []T, total int64, page, pageSize int) Response[PaginatedData[T]] {
	if pageSize <= 0 {
		pageSize = len(items)
		if pageSize == 0 {
			pageSize = 1
		}
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	return Response[PaginatedData[T]]{
		Success: true,
		Data: PaginatedData[T]{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
}
