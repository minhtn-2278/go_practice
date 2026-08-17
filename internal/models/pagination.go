package models

import "errors"

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

var ErrInvalidPagination = errors.New("page must be at least 1 and limit must be between 1 and 100")

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewPagination(page, limit int) (Pagination, error) {
	pagination := Pagination{Page: page, Limit: limit}
	if page < 1 || limit < 1 || limit > MaxLimit {
		return Pagination{}, ErrInvalidPagination
	}

	return pagination, nil
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p Pagination) WithTotal(total int64) Pagination {
	if total < 0 {
		total = 0
	}

	p.Total = total
	if total == 0 {
		return p
	}

	p.TotalPages = (total + int64(p.Limit) - 1) / int64(p.Limit)
	return p
}

type PaginatedResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func NewPaginatedResult[T any](items []T, pagination Pagination, total int64) PaginatedResult[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return PaginatedResult[T]{
		Data:       items,
		Pagination: pagination.WithTotal(total),
	}
}
