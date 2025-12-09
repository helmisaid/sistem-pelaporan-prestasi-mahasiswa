package service

import (
	"context"
	"database/sql"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
)

type ILecturerService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateLecturerProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateLecturerProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.LecturerInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateLecturerID(ctx context.Context, lecturerID string, excludeUserID *string) error
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
