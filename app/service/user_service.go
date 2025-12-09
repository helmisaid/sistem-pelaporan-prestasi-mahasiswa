package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/utils"

	"github.com/google/uuid"
)

type IUserService interface {
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedUsers, error)
	GetByID(ctx context.Context, id string) (*model.UserDetailDTO, error)
	Create(ctx context.Context, req model.CreateUserRequest) (*model.UserDetailDTO, error)
	Update(ctx context.Context, id string, req model.UpdateUserRequest) (*model.UserDetailDTO, error)
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, id string, req model.UpdateRoleRequest) error
}

type UserService struct {
	repo repository.IUserRepository
	db   *sql.DB
}

func NewUserService(repo repository.IUserRepository, db *sql.DB) IUserService {
	return &UserService{
		repo: repo,
		db:   db,
	}
}

// GetAllUsers
func (s *UserService) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	validSortColumns := map[string]bool{
		"username":   true,
		"email":      true,
		"full_name":  true,
		"created_at": true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	users, total, err := s.repo.GetAll(ctx, page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	userDTOs := make([]model.UserListDTO, len(users))
	for i, user := range users {
		userDTOs[i] = user.ToListDTO()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedUsers{
		Data:       userDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserDetailByID 
func (s *UserService) GetByID(ctx context.Context, id string) (*model.UserDetailDTO, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}

	dto := user.ToDetailDTO()

	// Get student
	if user.Role.Name == "Mahasiswa" {
		studentInfo, err := s.repo.GetStudentByUserID(ctx, id)
		if err == nil && studentInfo != nil {
			dto.Student = studentInfo
		}
	}

	// Get lecturer 
	if user.Role.Name == "Dosen Wali" {
		lecturerInfo, err := s.repo.GetLecturerByUserID(ctx, id)
		if err == nil && lecturerInfo != nil {
			dto.Lecturer = lecturerInfo
		}
	}

	return &dto, nil
}

// Create 
func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.UserDetailDTO, error) {
	exists, err := s.repo.CheckUsernameExists(ctx, req.Username, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if exists {
		return nil, fmt.Errorf("username sudah digunakan: %w", model.ErrValidationFailed)
	}

	exists, err = s.repo.CheckEmailExists(ctx, req.Email, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if exists {
		return nil, fmt.Errorf("email sudah digunakan: %w", model.ErrValidationFailed)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, model.ErrTokenGenerationFailed
	}

	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("format Role ID tidak valid: %w", model.ErrValidationFailed)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	defer tx.Rollback()

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       roleUUID,
		IsActive:     true,
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	if req.StudentID != nil && *req.StudentID != "" {
		err = s.repo.CreateStudent(
			ctx, tx,
			user.ID.String(),
			*req.StudentID,
			utils.GetStringOrDefault(req.ProgramStudy, ""),
			utils.GetStringOrDefault(req.AcademicYear, ""),
			req.AdvisorID,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat profil mahasiswa: %w", model.ErrDatabaseError)
		}
	}

	if req.LecturerID != nil && *req.LecturerID != "" {
		err = s.repo.CreateLecturer(
			ctx, tx,
			user.ID.String(),
			*req.LecturerID,
			utils.GetStringOrDefault(req.Department, ""),
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat profil dosen: %w", model.ErrDatabaseError)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, model.ErrDatabaseError
	}

	return s.GetByID(ctx, user.ID.String())
}

// Update 
func (s *UserService) Update(ctx context.Context, id string, req model.UpdateUserRequest) (*model.UserDetailDTO, error) {
	existingUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if existingUser == nil {
		return nil, model.ErrUserNotFound
	}

	if req.Email != nil && *req.Email != existingUser.Email {
		exists, err := s.repo.CheckEmailExists(ctx, *req.Email, &id)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if exists {
			return nil, fmt.Errorf("email sudah digunakan: %w", model.ErrValidationFailed)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	defer tx.Rollback()

	updatedUser := &model.User{
		Email:    utils.GetStringOrDefault(req.Email, existingUser.Email),
		FullName: utils.GetStringOrDefault(req.FullName, existingUser.FullName),
		IsActive: utils.GetBoolOrDefault(req.IsActive, existingUser.IsActive),
	}

	err = s.repo.Update(ctx, id, updatedUser)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	if existingUser.Role.Name == "Mahasiswa" {
		if req.ProgramStudy != nil || req.AcademicYear != nil || req.AdvisorID != nil {
			err = s.repo.UpdateStudent(ctx, tx, id, req.ProgramStudy, req.AcademicYear, req.AdvisorID)
			if err != nil {
				return nil, fmt.Errorf("gagal update profil mahasiswa: %w", model.ErrDatabaseError)
			}
		}
	}

	if existingUser.Role.Name == "Dosen Wali" {
		if req.Department != nil {
			err = s.repo.UpdateLecturer(ctx, tx, id, *req.Department)
			if err != nil {
				return nil, fmt.Errorf("gagal update profil dosen: %w", model.ErrDatabaseError)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, model.ErrDatabaseError
	}

	return s.GetByID(ctx, id)
}

// Soft deletes 
func (s *UserService) Delete(ctx context.Context, id string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if user == nil {
		return model.ErrUserNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}

	return nil
}

// UpdateRole 
func (s *UserService) UpdateRole(ctx context.Context, id string, req model.UpdateRoleRequest) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if user == nil {
		return model.ErrUserNotFound
	}

	err = s.repo.UpdateRole(ctx, id, req.RoleID)
	if err != nil {
		return model.ErrDatabaseError
	}

	return nil
}

