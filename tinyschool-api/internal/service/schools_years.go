package service

import (
	"context"
	"strings"
	"time"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
)

func (a *App) Overview(ctx context.Context) (dto.Overview, error) {
	value, err := a.storage.Overview(ctx)
	if err != nil {
		return dto.Overview{}, err
	}
	upcoming, err := a.storage.Upcoming(ctx, value.AcademicYear.ID, a.now().Format("2006-01-02"), 5)
	if err != nil {
		return dto.Overview{}, err
	}
	upcomingDTO := make([]dto.UpcomingItem, len(upcoming))
	for index, item := range upcoming {
		upcomingDTO[index] = dto.UpcomingItem{
			ID: item.ID, Kind: item.Kind, Title: item.Name, Date: item.Date,
			StudentCount: item.StudentCount,
		}
		if item.Class != nil {
			upcomingDTO[index].ClassName = item.Class.Name
		}
	}
	return dto.Overview{
		Students: value.Students, Classes: value.Classes, Assignments: value.Assignments, Exams: value.Exams,
		School: referenceDTO(value.School), AcademicYear: referenceDTO(value.AcademicYear), Upcoming: upcomingDTO,
	}, nil
}

func (a *App) ListSchools(ctx context.Context, input dto.ListOptions) (dto.Page[dto.School], error) {
	options, err := listOptions(input, map[string]bool{"name": true}, "name")
	if err != nil {
		return dto.Page[dto.School]{}, err
	}
	items, total, err := a.storage.ListSchools(ctx, options)
	if err != nil {
		return dto.Page[dto.School]{}, err
	}
	result := make([]dto.School, len(items))
	for index := range items {
		result[index] = schoolDTO(items[index])
	}
	return dto.Page[dto.School]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetSchool(ctx context.Context, id string) (dto.School, error) {
	item, err := a.storage.School(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.School{}, translate(err, "school")
	}
	return schoolDTO(item), nil
}

func (a *App) CreateSchool(ctx context.Context, input dto.SchoolRequest) (dto.School, error) {
	item, err := a.schoolFromInput("", input)
	if err != nil {
		return dto.School{}, err
	}
	item.ID, err = a.newID("sch")
	if err != nil {
		return dto.School{}, err
	}
	created, err := a.storage.CreateSchool(ctx, item)
	if err != nil {
		return dto.School{}, translate(err, "school")
	}
	return schoolDTO(created), nil
}

func (a *App) UpdateSchool(ctx context.Context, id string, input dto.SchoolRequest) (dto.School, error) {
	if _, err := a.storage.School(ctx, strings.TrimSpace(id)); err != nil {
		return dto.School{}, translate(err, "school")
	}
	item, err := a.schoolFromInput(strings.TrimSpace(id), input)
	if err != nil {
		return dto.School{}, err
	}
	updated, err := a.storage.UpdateSchool(ctx, item)
	if err != nil {
		return dto.School{}, translate(err, "school")
	}
	return schoolDTO(updated), nil
}

func (a *App) DeleteSchool(ctx context.Context, id string) error {
	return translate(a.storage.DeleteSchool(ctx, strings.TrimSpace(id)), "school")
}

func (a *App) schoolFromInput(id string, input dto.SchoolRequest) (model.School, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return model.School{}, validation("name is required")
	}
	grades, err := uniqueTrimmed(input.Grades, "grades")
	if err != nil {
		return model.School{}, err
	}
	if len(grades) == 0 {
		return model.School{}, validation("at least one grade is required")
	}
	return model.School{ID: id, Name: input.Name, Grades: grades, IsActive: input.IsActive}, nil
}

func (a *App) ListAcademicYears(ctx context.Context, input dto.ListOptions) (dto.Page[dto.AcademicYear], error) {
	options, err := listOptions(input, map[string]bool{
		"name": true, "startDate": true, "durationDays": true,
	}, "startDate")
	if err != nil {
		return dto.Page[dto.AcademicYear]{}, err
	}
	items, total, err := a.storage.ListAcademicYears(ctx, options)
	if err != nil {
		return dto.Page[dto.AcademicYear]{}, err
	}
	result := make([]dto.AcademicYear, len(items))
	for index := range items {
		result[index] = academicYearDTO(items[index])
	}
	return dto.Page[dto.AcademicYear]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetAcademicYear(ctx context.Context, id string) (dto.AcademicYear, error) {
	item, err := a.storage.AcademicYear(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.AcademicYear{}, translate(err, "academic year")
	}
	return academicYearDTO(item), nil
}

func (a *App) CreateAcademicYear(ctx context.Context, input dto.AcademicYearRequest) (dto.AcademicYear, error) {
	item, err := a.academicYearFromInput("", input)
	if err != nil {
		return dto.AcademicYear{}, err
	}
	if _, err := a.storage.School(ctx, item.SchoolID); err != nil {
		return dto.AcademicYear{}, translate(err, "school")
	}
	item.ID, err = a.newID("ay")
	if err != nil {
		return dto.AcademicYear{}, err
	}
	for index := range item.Segments {
		item.Segments[index].ID, err = a.newID("seg")
		if err != nil {
			return dto.AcademicYear{}, err
		}
	}
	created, err := a.storage.CreateAcademicYear(ctx, item)
	if err != nil {
		return dto.AcademicYear{}, translate(err, "academic year")
	}
	return academicYearDTO(created), nil
}

func (a *App) UpdateAcademicYear(ctx context.Context, id string, input dto.AcademicYearRequest) (dto.AcademicYear, error) {
	if _, err := a.storage.AcademicYear(ctx, strings.TrimSpace(id)); err != nil {
		return dto.AcademicYear{}, translate(err, "academic year")
	}
	item, err := a.academicYearFromInput(strings.TrimSpace(id), input)
	if err != nil {
		return dto.AcademicYear{}, err
	}
	if _, err := a.storage.School(ctx, item.SchoolID); err != nil {
		return dto.AcademicYear{}, translate(err, "school")
	}
	for index := range item.Segments {
		item.Segments[index].ID, err = a.newID("seg")
		if err != nil {
			return dto.AcademicYear{}, err
		}
	}
	updated, err := a.storage.UpdateAcademicYear(ctx, item)
	if err != nil {
		return dto.AcademicYear{}, translate(err, "academic year")
	}
	return academicYearDTO(updated), nil
}

func (a *App) DeleteAcademicYear(ctx context.Context, id string) error {
	return translate(a.storage.DeleteAcademicYear(ctx, strings.TrimSpace(id)), "academic year")
}

func (a *App) academicYearFromInput(id string, input dto.AcademicYearRequest) (model.AcademicYear, error) {
	input.SchoolID = strings.TrimSpace(input.SchoolID)
	input.Name = strings.TrimSpace(input.Name)
	if input.SchoolID == "" {
		return model.AcademicYear{}, validation("schoolId is required")
	}
	if input.Name == "" {
		return model.AcademicYear{}, validation("name is required")
	}
	start, err := date(input.StartDate, "startDate")
	if err != nil {
		return model.AcademicYear{}, err
	}
	if len(input.Segments) == 0 {
		return model.AcademicYear{}, validation("at least one segment is required")
	}
	segments := make([]model.AcademicSegment, len(input.Segments))
	cursor := start
	totalDays := 0
	for index, segment := range input.Segments {
		segment.Name = strings.TrimSpace(segment.Name)
		segment.Type = strings.ToLower(strings.TrimSpace(segment.Type))
		if segment.Name == "" {
			return model.AcademicYear{}, validation("segment name is required")
		}
		if segment.Type != "term" && segment.Type != "vacation" {
			return model.AcademicYear{}, validation("segment type must be term or vacation")
		}
		if segment.DurationDays <= 0 {
			return model.AcademicYear{}, validation("segment durationDays must be positive")
		}
		end := cursor.AddDate(0, 0, segment.DurationDays-1)
		segments[index] = model.AcademicSegment{
			Name: segment.Name, Type: segment.Type, DurationDays: segment.DurationDays,
			StartDate: cursor.Format(time.DateOnly), EndDate: end.Format(time.DateOnly),
		}
		totalDays += segment.DurationDays
		cursor = end.AddDate(0, 0, 1)
	}
	return model.AcademicYear{
		ID: id, SchoolID: input.SchoolID, Name: input.Name, StartDate: start.Format(time.DateOnly),
		EndDate: cursor.AddDate(0, 0, -1).Format(time.DateOnly), DurationDays: totalDays,
		IsCurrent: input.IsCurrent, Segments: segments,
	}, nil
}

func schoolDTO(item model.School) dto.School {
	grades := append([]string(nil), item.Grades...)
	if grades == nil {
		grades = []string{}
	}
	return dto.School{ID: item.ID, Name: item.Name, Grades: grades, IsActive: item.IsActive}
}

func academicYearDTO(item model.AcademicYear) dto.AcademicYear {
	segments := make([]dto.AcademicSegment, len(item.Segments))
	for index, segment := range item.Segments {
		segments[index] = dto.AcademicSegment{
			ID: segment.ID, Name: segment.Name, Type: segment.Type,
			DurationDays: segment.DurationDays, StartDate: segment.StartDate, EndDate: segment.EndDate,
		}
	}
	return dto.AcademicYear{
		ID: item.ID, SchoolID: item.SchoolID, Name: item.Name, StartDate: item.StartDate,
		EndDate: item.EndDate, DurationDays: item.DurationDays, IsCurrent: item.IsCurrent,
		Segments: segments,
	}
}

func referenceDTO(item model.Reference) dto.Reference {
	return dto.Reference{ID: item.ID, Name: item.Name}
}
