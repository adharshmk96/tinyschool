package gormsqlite

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"tinyschool-api/internal/storage"
)

func storageError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return storage.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) {
		return fmt.Errorf("%w: %v", storage.ErrConflict, err)
	}
	return err
}

func paginate(db *gorm.DB, options storage.ListOptions) *gorm.DB {
	page, size := options.Page, options.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return db.Offset((page - 1) * size).Limit(size)
}

func order(db *gorm.DB, options storage.ListOptions, allowed map[string]string, fallback string) *gorm.DB {
	column, ok := allowed[options.Sort]
	if !ok {
		column = fallback
	}
	direction := "ASC"
	if strings.EqualFold(options.Order, "desc") {
		direction = "DESC"
	}
	return db.Order(column + " " + direction)
}

// inclusiveDays rebuilds the derived duration of a date range that is stored
// alongside its start and end dates. Imported files only carry the dates.
func inclusiveDays(startDate, endDate string) int {
	start, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		return 0
	}
	end, err := time.Parse(time.DateOnly, endDate)
	if err != nil {
		return 0
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days < 0 {
		return 0
	}
	return days
}

func contains(search string) string {
	return "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
}
