package service

import (
	"context"
	"database/sql"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
)

type IStudentService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateStudentProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateStudentProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.StudentInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateStudentID(ctx context.Context, studentID string, excludeUserID *string) error
}

type StudentService struct {
	studentRepo repository.IStudentRepository
}

func NewStudentService(
	studentRepo repository.IStudentRepository,
) IStudentService {
	return &StudentService{
		studentRepo: studentRepo,
	}
}

func (s *StudentService) CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateStudentProfileRequest) error {
	if err := s.ValidateStudentID(ctx, req.StudentID, nil); err != nil {
		return err
	}

	return s.studentRepo.Create(ctx, tx, userID, req.StudentID, req.ProgramStudy, req.AcademicYear, req.AdvisorID)
}

func (s *StudentService) UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateStudentProfileRequest) error {
	if req.StudentID != nil && *req.StudentID != "" {
		if err := s.ValidateStudentID(ctx, *req.StudentID, &userID); err != nil {
			return err
		}
	}

	return s.studentRepo.Update(ctx, tx, userID, req.StudentID, req.ProgramStudy, req.AcademicYear, req.AdvisorID)
}

func (s *StudentService) GetProfile(ctx context.Context, userID string) (*model.StudentInfo, error) {
	return s.studentRepo.GetByUserID(ctx, userID)
}

func (s *StudentService) DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error {
	return s.studentRepo.Delete(ctx, tx, userID)
}

func (s *StudentService) ValidateStudentID(ctx context.Context, studentID string, excludeUserID *string) error {
	exists, err := s.studentRepo.CheckStudentIDExists(ctx, studentID, excludeUserID)
	if err != nil {
		return model.ErrDatabaseError
	}
	if exists {
		return model.NewValidationError("student ID sudah digunakan")
	}
	return nil
}
