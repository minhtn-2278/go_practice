package utils

import (
	"net/url"
	"strconv"

	"employee-management/internal/models"
)

func ParsePagination(query url.Values) (models.Pagination, error) {
	page, err := queryInt(query, "page", models.DefaultPage)
	if err != nil {
		return models.Pagination{}, err
	}

	limit, err := queryInt(query, "limit", models.DefaultLimit)
	if err != nil {
		return models.Pagination{}, err
	}

	return models.NewPagination(page, limit)
}

func queryInt(query url.Values, key string, fallback int) (int, error) {
	value := query.Get(key)
	if value == "" {
		return fallback, nil
	}

	return strconv.Atoi(value)
}
