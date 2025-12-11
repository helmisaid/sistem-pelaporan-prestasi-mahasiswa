package model

import "github.com/google/uuid"

type CreateLecturerProfileRequest struct {
	LecturerID string `json:"lecturer_id" validate:"required"`
	Department string `json:"department" validate:"required"`
}

type UpdateLecturerProfileRequest struct {
	LecturerID *string `json:"lecturer_id,omitempty"`
	Department *string `json:"department,omitempty"`
}

type LecturerInfo struct {
	ID         string `json:"id"`          
	LecturerID string `json:"lecturer_id"` 
	Department string `json:"department"`
}

// DTOs for Lecturer Endpoints
type LecturerListDTO struct {
	ID         uuid.UUID `json:"id"`
	LecturerID string    `json:"lecturer_id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Department string    `json:"department"`
	IsActive   bool      `json:"is_active"`
}

type PaginatedLecturers struct {
	Data       []LecturerListDTO `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}
