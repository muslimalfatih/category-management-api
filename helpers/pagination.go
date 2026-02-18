package helpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// ParsePagination extracts page and limit from query parameters with
// sensible defaults and upper-bound clamping.
func ParsePagination(c *gin.Context) (page, limit int) {
	page = DefaultPage
	limit = DefaultLimit

	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
		if limit > MaxLimit {
			limit = MaxLimit
		}
	}

	return page, limit
}

// CalcTotalPages returns the total number of pages given a total item
// count and a per-page limit.
func CalcTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
