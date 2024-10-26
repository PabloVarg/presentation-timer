package filters

import (
	"math"
	"net/http"
	"strings"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
	"github.com/PabloVarg/presentation-timer/internal/validation"
)

type Filters struct {
	Page           int32
	PageSize       int32
	SortBy         string
	SafeSortFields []string
}

type PageInfo struct {
	TotalPages int64 `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
}

func FromRequest(r *http.Request, defaultPageSize int32, safeSortFields ...string) (Filters, validation.Validator) {
	v := validation.New()

	page, err := helpers.QueryInt32(r, "page", 1)
	if err != nil {
		v.AddErrors("page", "not a valid number")
	}

	pageSize, err := helpers.QueryInt32(r, "page_size", defaultPageSize)
	if err != nil {
		v.AddErrors("page_size", "not a valid number")
	}

	result := Filters{
		Page:           page,
		PageSize:       pageSize,
		SortBy:         r.URL.Query().Get("sort_by"),
		SafeSortFields: safeSortFields,
	}
	result.Validate(v)

	return result, v
}

func (f Filters) Validate(v validation.Validator) {
	v.Check(
		"page",
		f.Page,
		validation.IntCheckNatural("can not be negative"),
		validation.IntCheckMax(10_000_000, "exceeded maximum page limit"),
	)
	v.Check(
		"page_size",
		f.PageSize,
		validation.IntCheckNatural("can not be negative"),
		validation.IntCheckMax(100, "maximum page size is 100"),
	)
	if f.SortBy == "" {
		return
	}
	v.Check(
		"sort_by",
		strings.TrimPrefix(f.SortBy, "-"),
		validation.StringCheckIn(f.SafeSortFields, "invalid value"),
	)
}

func (f Filters) PageInfo(totalRows int64) PageInfo {
	return PageInfo{
		TotalPages: int64(math.Ceil(float64(totalRows) / float64(f.PageSize))),
		TotalItems: totalRows,
	}
}

func (f Filters) QueryLimit() int32 {
	return f.PageSize
}

func (f Filters) QueryOffset() int32 {
	return (f.Page - 1) * f.PageSize
}

func (f Filters) QuerySortDirection() string {
	if strings.HasPrefix(f.SortBy, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) QuerySortBy() string {
	return strings.TrimPrefix(f.SortBy, "-")
}
