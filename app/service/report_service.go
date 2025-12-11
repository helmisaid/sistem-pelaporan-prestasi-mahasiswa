package service

import (
	"context"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
)

type IReportService interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardStatistics, error)
	GetStudentReport(ctx context.Context, targetUserUUID, viewerUserID, viewerRole string) (*model.StudentReportDTO, error)
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

// GetDashboardStats
func (s *ReportService) GetDashboardStats(ctx context.Context) (*model.DashboardStatistics, error) {
	return s.reportRepo.GetGlobalStats(ctx)
}

// GetStudentReport
func (s *ReportService) GetStudentReport(ctx context.Context, targetUserUUID, viewerUserID, viewerRole string) (*model.StudentReportDTO, error) {
	targetStudent, err := s.studentRepo.GetByUserID(ctx, targetUserUUID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if targetStudent == nil {
		return nil, model.NewNotFoundError("Mahasiswa tidak ditemukan")
	}

	// Role-based access control
	if viewerRole == "Mahasiswa" {
		if viewerUserID != targetUserUUID {
			return nil, model.NewValidationError("Anda tidak berhak melihat laporan mahasiswa lain")
		}
	} else if viewerRole == "Dosen Wali" {
		lecturer, err := s.lecturerSvc.GetProfile(ctx, viewerUserID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}

		if targetStudent.AdvisorID == nil || *targetStudent.AdvisorID != lecturer.ID {
			return nil, model.NewValidationError("Mahasiswa ini bukan bimbingan Anda")
		}
	}

	studentDetail, err := s.studentRepo.GetDetailByID(ctx, targetStudent.ID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if studentDetail == nil {
		return nil, model.NewNotFoundError("Detail mahasiswa tidak ditemukan")
	}

	report, err := s.reportRepo.GetStudentReport(ctx, targetStudent.ID)
	if err != nil {
		return nil, model.ErrDatabaseError
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

	return report, nil
}
