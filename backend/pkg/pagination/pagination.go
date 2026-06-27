// Package pagination provides offset-based pagination parsing and metadata.
package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// Params holds normalized, bounded pagination input.
type Params struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Offset is the SQL OFFSET for the current page.
func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }

// Limit is the SQL LIMIT (== PerPage).
func (p Params) Limit() int { return p.PerPage }

// FromQuery reads ?page= and ?per_page= and clamps them to safe bounds.
func FromQuery(c echo.Context) Params {
	page := atoiDefault(c.QueryParam("page"), defaultPage)
	perPage := atoiDefault(c.QueryParam("per_page"), defaultPerPage)

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return Params{Page: page, PerPage: perPage}
}

// Meta is the pagination block returned in the response envelope's meta.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta builds pagination metadata from params and a total count.
func NewMeta(p Params, total int64) Meta {
	totalPages := 0
	if p.PerPage > 0 {
		totalPages = int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	}
	return Meta{Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: totalPages}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
