package gormsqlite

import (
	"sort"

	"tinyschool-api/internal/model"
)

func userModel(r userRecord) model.User {
	role := r.Role
	if role == "" {
		role = model.RoleUser
	}
	return model.User{
		ID: r.ID, Name: r.Name, Email: r.Email, PasswordHash: r.PasswordHash,
		Role: role, CreatedAt: r.CreatedAt, BlockedAt: r.BlockedAt,
	}
}

func schoolModel(r schoolRecord) model.School {
	m := model.School{ID: r.ID, Name: r.Name, IsActive: r.IsActive, Grades: make([]string, len(r.Grades))}
	for i := range r.Grades {
		m.Grades[i] = r.Grades[i].Grade
	}
	return m
}

func yearModel(r academicYearRecord) model.AcademicYear {
	m := model.AcademicYear{ID: r.ID, SchoolID: r.SchoolID, Name: r.Name, StartDate: r.StartDate, EndDate: r.EndDate, DurationDays: r.DurationDays, IsCurrent: r.IsCurrent}
	for _, item := range r.Segments {
		m.Segments = append(m.Segments, model.AcademicSegment{ID: item.ID, Name: item.Name, Type: item.Type, DurationDays: item.DurationDays, StartDate: item.StartDate, EndDate: item.EndDate})
	}
	return m
}

func studentModel(r studentRecord) model.Student {
	m := model.Student{
		ID: r.ID, SchoolID: r.SchoolID, FirstName: r.FirstName, LastName: r.LastName,
		Email: r.Email, Phone: r.Phone, GuardianName: r.GuardianName,
		GuardianEmail: r.GuardianEmail, GuardianPhone: r.GuardianPhone,
		ResidentAddress: r.ResidentAddress, PermanentAddress: r.PermanentAddress,
	}
	grades := append([]studentGradeRecord(nil), r.Grades...)
	sort.Slice(grades, func(i, j int) bool {
		return grades[i].AcademicYear.StartDate < grades[j].AcademicYear.StartDate
	})
	for _, grade := range grades {
		m.Grades = append(m.Grades, model.StudentGrade{
			AcademicYearID:   grade.AcademicYearID,
			AcademicYearName: grade.AcademicYear.Name,
			Grade:            grade.Grade,
			IsCurrent:        grade.AcademicYear.IsCurrent,
		})
	}
	return m
}

func classModel(r classRecord) model.Class {
	m := model.Class{ID: r.ID, SchoolID: r.SchoolID, AcademicYearID: r.AcademicYearID, Name: r.Name, Subject: r.Subject, Grade: r.Grade, Description: r.Description}
	for _, student := range r.Students {
		m.Students = append(m.Students, studentModel(student))
	}
	return m
}

func assignmentModel(r assignmentRecord) model.Assignment {
	m := model.Assignment{ID: r.ID, SchoolID: r.SchoolID, AcademicYearID: r.AcademicYearID, Name: r.Name, Type: r.Type, DueDate: r.DueDate, TotalScore: r.TotalScore, ClassID: r.ClassID}
	if r.Class != nil {
		m.Class = &model.Reference{ID: r.Class.ID, Name: r.Class.Name}
	}
	for _, item := range r.Students {
		m.Students = append(m.Students, model.AssignmentStudent{Student: studentModel(item.Student), Score: item.Score, CompletedAt: item.CompletedAt})
	}
	return m
}

func examModel(r examRecord) model.Exam {
	m := model.Exam{ID: r.ID, SchoolID: r.SchoolID, AcademicYearID: r.AcademicYearID, ClassID: r.ClassID, Name: r.Name, Type: r.Type, ExamDate: r.ExamDate, TotalScore: r.TotalScore, Class: model.Reference{ID: r.Class.ID, Name: r.Class.Name}}
	for _, item := range r.Students {
		m.Students = append(m.Students, model.ExamStudent{Student: studentModel(item.Student), Score: item.Score, MarkedAt: item.MarkedAt})
	}
	return m
}
