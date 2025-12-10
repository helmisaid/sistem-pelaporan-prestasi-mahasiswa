package model

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