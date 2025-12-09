package model

import (
	"time"

	"github.com/google/uuid"
)


type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` 
	FullName     string    `json:"full_name"`
	RoleID       uuid.UUID `json:"role_id"`
	Role         Role      `json:"role"` 
	Permissions  []string  `json:"permissions"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Role struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}



type UserLoginDTO struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	Role        Role      `json:"role"`
	Permissions []string  `json:"permissions"`
}


type UserProfileDTO struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	Role        Role      `json:"role"`
	Permissions []string  `json:"permissions"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}



type CreateUserRequest struct {
	Username     string  `json:"username" validate:"required"`
	Email        string  `json:"email" validate:"required,email"`
	Password     string  `json:"password" validate:"required,min=8"`
	FullName     string  `json:"full_name" validate:"required"`
	RoleID       string  `json:"role_id" validate:"required,uuid"`
	
	// For students
	StudentID    *string `json:"student_id,omitempty"`
	ProgramStudy *string `json:"program_study,omitempty"`
	AcademicYear *string `json:"academic_year,omitempty"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
	
	// For lecturers
	LecturerID   *string `json:"lecturer_id,omitempty"`
	Department   *string `json:"department,omitempty"`
}


type UpdateUserRequest struct {
	Email        *string `json:"email,omitempty" validate:"omitempty,email"`
	FullName     *string `json:"full_name,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	
	// For students
	ProgramStudy *string `json:"program_study,omitempty"`
	AcademicYear *string `json:"academic_year,omitempty"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
	
	// For lecturers
	Department   *string `json:"department,omitempty"`
}


type UpdateRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid"`
}


type UserListDTO struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     Role      `json:"role"`
	IsActive bool      `json:"is_active"`
}


type UserDetailDTO struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	Role      Role       `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	

	Student   *StudentInfo   `json:"student,omitempty"`
	

	Lecturer  *LecturerInfo  `json:"lecturer,omitempty"`
}


type StudentInfo struct {
	StudentID    string  `json:"student_id"`
	ProgramStudy string  `json:"program_study"`
	AcademicYear string  `json:"academic_year"`
	AdvisorID    *string `json:"advisor_id,omitempty"`
	AdvisorName  *string `json:"advisor_name,omitempty"`
}


type LecturerInfo struct {
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}


type PaginatedUsers struct {
	Data       []UserListDTO `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}



func (u *User) ToLoginDTO() UserLoginDTO {
	return UserLoginDTO{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		Role:        u.Role,
		Permissions: u.Permissions,
	}
}


func (u *User) ToProfileDTO() UserProfileDTO {
	return UserProfileDTO{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		Role:        u.Role,
		Permissions: u.Permissions,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
	}
}


func (u *User) ToListDTO() UserListDTO {
	return UserListDTO{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}


func (u *User) ToDetailDTO() UserDetailDTO {
	return UserDetailDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}