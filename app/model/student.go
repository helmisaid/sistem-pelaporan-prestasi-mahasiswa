package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateStudentProfileRequest struct {
	StudentID    string  `json:"student_id" validate:"required"`
	ProgramStudy string  `json:"program_study" validate:"required"`
	AcademicYear string  `json:"academic_year" validate:"required"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateStudentProfileRequest struct {
	StudentID    *string `json:"student_id,omitempty"`
	ProgramStudy *string `json:"program_study,omitempty"`
	AcademicYear *string `json:"academic_year,omitempty"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateAdvisorRequest struct {
	AdvisorID *string `json:"advisor_id" validate:"omitempty,uuid"`
}

// DTOs for Student Endpoints
type StudentListDTO struct {
	ID           uuid.UUID `json:"id"`
	StudentID    string    `json:"student_id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	ProgramStudy string    `json:"program_study"`
	AcademicYear string    `json:"academic_year"`
	AdvisorID    *string   `json:"advisor_id,omitempty"`
	AdvisorName  *string   `json:"advisor_name,omitempty"`
	IsActive     bool      `json:"is_active"`
}

type StudentDetailDTO struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	FullName     string     `json:"full_name"`
	StudentID    string     `json:"student_id"`
	ProgramStudy string     `json:"program_study"`
	AcademicYear string     `json:"academic_year"`
	AdvisorID    *string    `json:"advisor_id,omitempty"`
	AdvisorName  *string    `json:"advisor_name,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PaginatedStudents struct {
	Data       []StudentListDTO `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type StudentInfo struct {
	ID           string  `json:"id"`            
	StudentID    string  `json:"student_id"`    
	ProgramStudy string  `json:"program_study"`
	AcademicYear string  `json:"academic_year"`
	AdvisorID    *string `json:"advisor_id"`
	AdvisorName  *string `json:"advisor_name"`
}
