package listquery

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type Options struct {
	Search   string
	Sort     string
	Order    string
	Page     int
	PageSize int
}

type Result[T any] struct {
	Items []T
	Total int
}

func Parse(values url.Values, allowedSorts map[string]bool, defaultSort string) (Options, error) {
	options := Options{
		Search:   strings.TrimSpace(values.Get("search")),
		Sort:     values.Get("sort"),
		Order:    strings.ToLower(values.Get("order")),
		Page:     1,
		PageSize: DefaultPageSize,
	}
	if options.Sort == "" {
		options.Sort = defaultSort
	}
	if !allowedSorts[options.Sort] {
		return Options{}, fmt.Errorf("sort must be one of %s", strings.Join(sortedKeys(allowedSorts), ", "))
	}
	if options.Order == "" {
		options.Order = "asc"
	}
	if options.Order != "asc" && options.Order != "desc" {
		return Options{}, fmt.Errorf("order must be asc or desc")
	}
	var err error
	if raw := values.Get("page"); raw != "" {
		options.Page, err = strconv.Atoi(raw)
		if err != nil || options.Page < 1 {
			return Options{}, fmt.Errorf("page must be a positive integer")
		}
	}
	if raw := values.Get("pageSize"); raw != "" {
		options.PageSize, err = strconv.Atoi(raw)
		if err != nil || options.PageSize < 1 || options.PageSize > MaxPageSize {
			return Options{}, fmt.Errorf("pageSize must be between 1 and %d", MaxPageSize)
		}
	}
	return options, nil
}

func Apply[T any](items []T, options Options, matches func(T, string) bool, less func(T, T, string) bool) Result[T] {
	filtered := make([]T, 0, len(items))
	search := strings.ToLower(options.Search)
	for _, item := range items {
		if search == "" || matches(item, search) {
			filtered = append(filtered, item)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if options.Order == "desc" {
			return less(filtered[j], filtered[i], options.Sort)
		}
		return less(filtered[i], filtered[j], options.Sort)
	})

	total := len(filtered)
	start := (options.Page - 1) * options.PageSize
	if start >= total {
		return Result[T]{Items: []T{}, Total: total}
	}
	end := min(start+options.PageSize, total)
	return Result[T]{Items: filtered[start:end], Total: total}
}

func Contains(search string, fields ...string) bool {
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), search) {
			return true
		}
	}
	return false
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
