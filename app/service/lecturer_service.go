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

type ILecturerService interface {
	CreateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.CreateLecturerProfileRequest) error
	UpdateProfile(ctx context.Context, tx *sql.Tx, userID string, req model.UpdateLecturerProfileRequest) error
	GetProfile(ctx context.Context, userID string) (*model.LecturerInfo, error)
	DeleteProfile(ctx context.Context, tx *sql.Tx, userID string) error
	ValidateLecturerID(ctx context.Context, lecturerID string, excludeUserID *string) error
	CheckExistsByID(ctx context.Context, id string) (bool, error)

	GetAll(c *fiber.Ctx) error
	GetAdvisees(c *fiber.Ctx) error
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

// GetAll godoc
// @Summary List all lecturers
// @Description Get paginated list of lecturers (Admin only)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by name or lecturer ID"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} helper.Response{data=model.PaginatedLecturers} "Lecturers retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Forbidden - Admin only"
// @Router /lecturers [get]
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	viewerRole := c.Locals("role").(string)
	if viewerRole != "Admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Anda tidak memiliki hak akses.",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sort_by", "u.created_at")
	sortOrder := c.Query("sort_order", "DESC")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize >  100 {
		pageSize = 10
	}

	validSortColumns := map[string]bool{
		"u.full_name":   true,
		"l.lecturer_id": true,
		"l.department":  true,
		"u.created_at":  true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "u.created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	data, total, err := s.lecturerRepo.GetAll(c.Context(), page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedLecturers{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar dosen berhasil diambil", result)
}

// GetAdvisees godoc
// @Summary Get lecturer's advisees
// @Description Get paginated list of students under a lecturer's guidance
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} helper.Response{data=model.PaginatedStudents} "Advisees retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Forbidden - Not your advisees"
// @Router /lecturers/{id}/advisees [get]
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)

	viewerUserID := c.Locals("user_id").(string)
	viewerRole := c.Locals("role").(string)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	exists, err := s.lecturerRepo.CheckExistsByID(c.Context(), lecturerID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if !exists {
		return helper.HandleError(c, model.NewNotFoundError("Dosen tidak ditemukan"))
	}

	if viewerRole == "Dosen Wali" {
		viewerLecturer, err := s.lecturerRepo.GetByUserID(c.Context(), viewerUserID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if viewerLecturer == nil {
			return helper.HandleError(c, model.NewValidationError("Profil dosen tidak ditemukan"))
		}
		if viewerLecturer.ID != lecturerID {
			return helper.HandleError(c, model.NewValidationError("Anda hanya dapat melihat mahasiswa bimbingan Anda sendiri"))
		}
	}

	data, total, err := s.lecturerRepo.GetAdvisees(c.Context(), lecturerID, page, pageSize)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedStudents{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar mahasiswa bimbingan berhasil diambil", result)
}
