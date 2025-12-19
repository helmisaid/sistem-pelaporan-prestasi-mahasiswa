package service

import (
	"context"
	"database/sql"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"

	"github.com/gofiber/fiber/v2"
)

type IStudentService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateStudentProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateStudentProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.StudentInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateStudentID(ctx context.Context, studentID string, excludeUserID *string) error

	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	UpdateAdvisor(c *fiber.Ctx) error
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

// GetAll godoc
// @Summary List all students
// @Description Get paginated list of students with optional filtering and sorting
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by name or student ID"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} helper.Response{data=model.PaginatedStudents} "Students retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Router /students [get]
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "u.created_at")
	sortOrder := c.Query("order", "desc")

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

	students, total, err := s.studentRepo.GetAll(c.Context(), page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedStudents{
		Data:       students,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar mahasiswa berhasil diambil", result)
}

// GetByID godoc
// @Summary Get student by ID
// @Description Get detailed information about a specific student
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} helper.Response{data=model.StudentDetailDTO} "Student details retrieved"
// @Failure 404 {object} helper.ErrorResponse "Student not found"
// @Router /students/{id} [get]
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	detail, err := s.studentRepo.GetDetailByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if detail == nil {
		return helper.HandleError(c, model.NewNotFoundError("mahasiswa tidak ditemukan"))
	}

	return helper.Success(c, "Detail mahasiswa berhasil diambil", detail)
}

// UpdateAdvisor godoc
// @Summary Update student advisor
// @Description Assign or change academic advisor for a student
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param request body model.UpdateAdvisorRequest true "Advisor ID"
// @Success 200 {object} helper.Response "Advisor updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 404 {object} helper.ErrorResponse "Student or lecturer not found"
// @Router /students/{id}/advisor [put]
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	detail, err := s.studentRepo.GetDetailByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if detail == nil {
		return helper.HandleError(c, model.NewNotFoundError("Mahasiswa tidak ditemukan"))
	}

	if req.AdvisorID != nil && *req.AdvisorID != "" {
		exists, err := s.lecturerSvc.CheckExistsByID(c.Context(), *req.AdvisorID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if !exists {
			return helper.HandleError(c, model.NewValidationError("Dosen wali dengan ID tersebut tidak ditemukan"))
		}
	}

	err = s.studentRepo.UpdateAdvisor(c.Context(), id, req.AdvisorID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Dosen wali berhasil diupdate", nil)
}
