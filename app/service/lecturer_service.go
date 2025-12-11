package service

import (
	"context"
	"database/sql"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
)

type ILecturerService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateLecturerProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateLecturerProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.LecturerInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateLecturerID(ctx context.Context, lecturerID string, excludeUserID *string) error
	CheckExistsByID(ctx context.Context, id string) (bool, error)
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedLecturers, error)
	GetAdvisees(ctx context.Context, lecturerID, viewerUserID, viewerRole string, page, pageSize int) (*model.PaginatedStudents, error)
}

type LecturerService struct {
	lecturerRepo repository.ILecturerRepository
}

func NewLecturerService(
	lecturerRepo repository.ILecturerRepository,
) ILecturerService {
	return &LecturerService{
		lecturerRepo: lecturerRepo,
	}
}

func (s *LecturerService) CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateLecturerProfileRequest) error {
	if err := s.ValidateLecturerID(ctx, req.LecturerID, nil); err != nil {
		return err
	}

	return s.lecturerRepo.Create(ctx, tx, userID, req.LecturerID, req.Department)
}

func (s *LecturerService) UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateLecturerProfileRequest) error {
	// Validate LecturerID if provided
	if req.LecturerID != nil && *req.LecturerID != "" {
		if err := s.ValidateLecturerID(ctx, *req.LecturerID, &userID); err != nil {
			return err
		}
	}

	return s.lecturerRepo.Update(ctx, tx, userID, req.LecturerID, req.Department)
}

func (s *LecturerService) GetProfile(ctx context.Context, userID string) (*model.LecturerInfo, error) {
	return s.lecturerRepo.GetByUserID(ctx, userID)
}

func (s *LecturerService) DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error {
	return s.lecturerRepo.Delete(ctx, tx, userID)
}

func (s *LecturerService) ValidateLecturerID(ctx context.Context, lecturerID string, excludeUserID *string) error {
	exists, err := s.lecturerRepo.CheckLecturerIDExists(ctx, lecturerID, excludeUserID)
	if err != nil {
		return model.ErrDatabaseError
	}
	if exists {
		return model.NewValidationError("lecturer ID sudah digunakan")
	}
	return nil
}

func (s *LecturerService) CheckExistsByID(ctx context.Context, id string) (bool, error) {
	return s.lecturerRepo.CheckExistsByID(ctx, id)
}

// GetAll retrieves all lecturers with pagination
func (s *LecturerService) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedLecturers, error) {
	// Validate and normalize pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	validSortColumns := map[string]bool{
		"u.full_name":     true,
		"l.lecturer_id":   true,
		"l.department":    true,
		"u.created_at":    true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "u.created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	data, total, err := s.lecturerRepo.GetAll(ctx, page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedLecturers{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAdvisees retrieves students supervised by a lecturer with role-based access control
func (s *LecturerService) GetAdvisees(ctx context.Context, lecturerID, viewerUserID, viewerRole string, page, pageSize int) (*model.PaginatedStudents, error) {
	// Validate and normalize pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Check if lecturer exists
	exists, err := s.lecturerRepo.CheckExistsByID(ctx, lecturerID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if !exists {
		return nil, model.NewNotFoundError("Dosen tidak ditemukan")
	}

	// Role-based access control
	if viewerRole == "Dosen Wali" {
		// Dosen Wali can only view their own advisees
		viewerLecturer, err := s.lecturerRepo.GetByUserID(ctx, viewerUserID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if viewerLecturer == nil {
			return nil, model.NewValidationError("Profil dosen tidak ditemukan")
		}
		if viewerLecturer.ID != lecturerID {
			return nil, model.NewValidationError("Anda hanya dapat melihat mahasiswa bimbingan Anda sendiri")
		}
	}
	// Admin can view all advisees - no additional check needed

	data, total, err := s.lecturerRepo.GetAdvisees(ctx, lecturerID, page, pageSize)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedStudents{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
