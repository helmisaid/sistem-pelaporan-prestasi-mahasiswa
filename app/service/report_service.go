package service

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"

	"github.com/gofiber/fiber/v2"
)

type IReportService interface {
	GetDashboardStats(c *fiber.Ctx) error
	GetStudentReport(c *fiber.Ctx) error
}

type ReportService struct {
	reportRepo  repository.IReportRepository
	studentRepo repository.IStudentRepository
	lecturerSvc ILecturerService
}

func NewReportService(
	reportRepo repository.IReportRepository,
	studentRepo repository.IStudentRepository,
	lecturerSvc ILecturerService,
) IReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		studentRepo: studentRepo,
		lecturerSvc: lecturerSvc,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get global statistics for the dashboard (Admin only)
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=model.DashboardStatistics} "Statistics retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Router /reports/statistics [get]
func (s *ReportService) GetDashboardStats(c *fiber.Ctx) error {
	result, err := s.reportRepo.GetGlobalStats(c.Context())
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Statistik dashboard berhasil diambil", result)
}

// GetStudentReport godoc
// @Summary Get student report
// @Description Get achievement report for a specific student with RBAC
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student User ID"
// @Success 200 {object} helper.Response{data=model.StudentReportDTO} "Student report retrieved"
// @Failure 403 {object} helper.ErrorResponse "Forbidden - Not authorized to view this report"
// @Failure 404 {object} helper.ErrorResponse "Student not found"
// @Router /reports/student/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	targetUserID := c.Params("id")
	viewerUserID := c.Locals("user_id").(string)
	viewerRole := c.Locals("role").(string)

	targetStudent, err := s.studentRepo.GetByUserID(c.Context(), targetUserID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if targetStudent == nil {
		return helper.HandleError(c, model.NewNotFoundError("Mahasiswa tidak ditemukan"))
	}

	if viewerRole == "Mahasiswa" {
		if viewerUserID != targetUserID {
			return helper.HandleError(c, model.NewValidationError("Anda tidak berhak melihat laporan mahasiswa lain"))
		}
	} else if viewerRole == "Dosen Wali" {
		lecturer, err := s.lecturerSvc.GetProfile(c.Context(), viewerUserID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}

		if targetStudent.AdvisorID == nil || *targetStudent.AdvisorID != lecturer.ID {
			return helper.HandleError(c, model.NewValidationError("Mahasiswa ini bukan bimbingan Anda"))
		}
	}

	studentDetail, err := s.studentRepo.GetDetailByID(c.Context(), targetStudent.ID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentDetail == nil {
		return helper.HandleError(c, model.NewNotFoundError("Detail mahasiswa tidak ditemukan"))
	}

	report, err := s.reportRepo.GetStudentReport(c.Context(), targetStudent.ID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	report.StudentProfile = model.StudentListDTO{
		ID:           studentDetail.ID,
		StudentID:    studentDetail.StudentID,
		FullName:     studentDetail.FullName,
		Email:        studentDetail.Email,
		ProgramStudy: studentDetail.ProgramStudy,
		AcademicYear: studentDetail.AcademicYear,
		AdvisorID:    studentDetail.AdvisorID,
		AdvisorName:  studentDetail.AdvisorName,
		IsActive:     studentDetail.IsActive,
	}

	return helper.Success(c, "Laporan prestasi mahasiswa berhasil diambil", report)
}
