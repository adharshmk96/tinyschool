package httpdelivery

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/service"
)

func listOptions(r *http.Request) (dto.ListOptions, error) {
	query := r.URL.Query()
	page, err := positiveQueryInteger(query.Get("page"), 1, "page")
	if err != nil {
		return dto.ListOptions{}, err
	}
	pageSize, err := positiveQueryInteger(query.Get("pageSize"), 10, "pageSize")
	if err != nil {
		return dto.ListOptions{}, err
	}
	if pageSize > 100 {
		return dto.ListOptions{}, fmt.Errorf("pageSize must be between 1 and 100")
	}
	return dto.ListOptions{
		Search:         strings.TrimSpace(query.Get("search")),
		AcademicYearID: strings.TrimSpace(query.Get("academicYearId")),
		Sort:           strings.TrimSpace(query.Get("sort")),
		Order:          strings.TrimSpace(query.Get("order")),
		Page:           page,
		PageSize:       pageSize,
	}, nil
}

func writeListError(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, service.ErrValidation) {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	writeServiceError(logger, w, r, err)
}

func positiveQueryInteger(value string, fallback int, name string) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
