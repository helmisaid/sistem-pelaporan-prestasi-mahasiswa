package service

import (
	"context"
	"database/sql"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
)

type IStudentService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateStudentProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateStudentProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.StudentInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateStudentID(ctx context.Context, studentID string, excludeUserID *string) error
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedStudents, error)
	GetByID(ctx context.Context, id string) (*model.StudentDetailDTO, error)
	UpdateAdvisor(ctx context.Context, id string, req model.UpdateAdvisorRequest) error
}

type StudentService struct {
	studentRepo repository.IStudentRepository
	lecturerSvc ILecturerService
}

func NewStudentService(
	studentRepo repository.IStudentRepository,
	lecturerSvc ILecturerService,
) IStudentService {
	return &StudentService{
		studentRepo: studentRepo,
		lecturerSvc: lecturerSvc,
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

func (s *StudentService) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedStudents, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	validSortColumns := map[string]bool{
		"u.full_name":     true,
		"s.student_id":    true,
		"s.program_study": true,
		"s.academic_year": true,
		"u.created_at":    true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "u.created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	students, total, err := s.studentRepo.GetAll(ctx, page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedStudents{
		Data:       students,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStudentByID
func (s *StudentService) GetByID(ctx context.Context, id string) (*model.StudentDetailDTO, error) {
	detail, err := s.studentRepo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if detail == nil {
		return nil, model.NewNotFoundError("mahasiswa tidak ditemukan")
	}
	return detail, nil
}

func (s *StudentService) UpdateAdvisor(ctx context.Context, id string, req model.UpdateAdvisorRequest) error {

	// Check if student exists
	detail, err := s.studentRepo.GetDetailByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if detail == nil {
		return model.NewNotFoundError("Mahasiswa tidak ditemukan")
	}

	// Validate advisor exists if provided
	if req.AdvisorID != nil && *req.AdvisorID != "" {
		exists, err := s.lecturerSvc.CheckExistsByID(ctx, *req.AdvisorID)
		if err != nil {
			return model.ErrDatabaseError
		}
		if !exists {
			return model.NewValidationError("Dosen wali dengan ID tersebut tidak ditemukan")
		}
	}

	err = s.studentRepo.UpdateAdvisor(ctx, id, req.AdvisorID)
	if err != nil {
		return model.ErrDatabaseError
	}

	return nil
}
