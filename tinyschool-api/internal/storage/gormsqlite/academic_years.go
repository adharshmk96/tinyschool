package gormsqlite

import (
	"context"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListAcademicYears(ctx context.Context, options storage.ListOptions) ([]model.AcademicYear, int64, error) {
	query := s.db.WithContext(ctx).Model(&academicYearRecord{}).Where("academic_years.user_id = ?", userID(ctx))
	if options.Search != "" {
		pattern := contains(options.Search)
		query = query.Where("LOWER(name) LIKE ? OR start_date LIKE ?", pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []academicYearRecord
	err := paginate(order(query, options, map[string]string{"name": "name", "startDate": "start_date", "durationDays": "duration_days"}, "start_date"), options).
		Preload("Segments", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).Find(&records).Error
	items := make([]model.AcademicYear, len(records))
	for i := range records {
		items[i] = yearModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) AcademicYear(ctx context.Context, id string) (model.AcademicYear, error) {
	var record academicYearRecord
	err := s.db.WithContext(ctx).Preload("Segments", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error
	return yearModel(record), storageError(err)
}

func yearRecord(ownerID string, year model.AcademicYear) academicYearRecord {
	record := academicYearRecord{ID: year.ID, UserID: ownerID, SchoolID: year.SchoolID, Name: year.Name, StartDate: year.StartDate, EndDate: year.EndDate, DurationDays: year.DurationDays, IsCurrent: year.IsCurrent}
	for i, segment := range year.Segments {
		record.Segments = append(record.Segments, academicSegmentRecord{ID: segment.ID, AcademicYearID: year.ID, Name: segment.Name, Type: segment.Type, DurationDays: segment.DurationDays, StartDate: segment.StartDate, EndDate: segment.EndDate, Position: i})
	}
	return record
}

func saveYear(tx *gorm.DB, ownerID string, year model.AcademicYear, create bool) error {
	if year.IsCurrent {
		if err := tx.Model(&academicYearRecord{}).Where("user_id = ? AND school_id = ? AND id <> ?", ownerID, year.SchoolID, year.ID).Update("is_current", false).Error; err != nil {
			return err
		}
	}
	record := yearRecord(ownerID, year)
	if create {
		return tx.Create(&record).Error
	}
	result := tx.Model(&academicYearRecord{}).Where("id = ? AND user_id = ?", year.ID, ownerID).Updates(map[string]any{
		"school_id": year.SchoolID, "name": year.Name, "start_date": year.StartDate, "end_date": year.EndDate,
		"duration_days": year.DurationDays, "is_current": year.IsCurrent,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if err := tx.Where("academic_year_id = ?", year.ID).Delete(&academicSegmentRecord{}).Error; err != nil {
		return err
	}
	for i := range record.Segments {
		if err := tx.Create(&record.Segments[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateAcademicYear(ctx context.Context, year model.AcademicYear) (model.AcademicYear, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return saveYear(tx, userID(ctx), year, true) })
	return year, storageError(err)
}

func (s *Store) UpdateAcademicYear(ctx context.Context, year model.AcademicYear) (model.AcademicYear, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return saveYear(tx, userID(ctx), year, false) })
	return year, storageError(err)
}

func (s *Store) DeleteAcademicYear(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&academicYearRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
