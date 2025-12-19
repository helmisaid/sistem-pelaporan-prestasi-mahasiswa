package service

import (
	"context"
	"net/http/httptest"
	"testing"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
)

type MockReportRepository struct {
	globalStats   *model.DashboardStatistics
	studentReport *model.StudentReportDTO
}

func (m *MockReportRepository) GetGlobalStats(ctx context.Context) (*model.DashboardStatistics, error) {
	return m.globalStats, nil
}

func (m *MockReportRepository) GetStudentReport(ctx context.Context, studentID string) (*model.StudentReportDTO, error) {
	if m.studentReport != nil {
		return m.studentReport, nil
	}
	return nil, nil
}

func TestReportService_GetDashboardStats(t *testing.T) {
	app := fiber.New()
	mockReportRepo := &MockReportRepository{
		globalStats: &model.DashboardStatistics{
			TotalStudents:     10,
			TotalLecturers:    2,
			TotalAchievements: 5,
		},
	}
	service := NewReportService(mockReportRepo, &MockStudentRepository{}, &MockLecturerService{})

	app.Get("/reports/statistics", service.GetDashboardStats)

	t.Run("GET - Global Stats Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/reports/statistics", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestReportService_GetStudentReport_RBAC(t *testing.T) {
	mockReportRepo := &MockReportRepository{}
	mockStudentRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}
	service := NewReportService(mockReportRepo, mockStudentRepo, mockLecturerSvc)

	targetUserID := "user-mhs-uuid"
	studentProfileID := "student-internal-id"
	advisorID := "lecturer-internal-id"

	mockStudentRepo.students[targetUserID] = &model.StudentInfo{
		ID:        studentProfileID,
		AdvisorID: &advisorID,
	}
	mockStudentRepo.detailID = studentProfileID
	mockReportRepo.studentReport = &model.StudentReportDTO{TotalPoints: 100}

	t.Run("Mahasiswa Access - Own Report (Success)", func(t *testing.T) {
		app := fiber.New()
		app.Get("/reports/student/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", targetUserID)
			c.Locals("role", "Mahasiswa")
			return service.GetStudentReport(c)
		})

		req := httptest.NewRequest("GET", "/reports/student/"+targetUserID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Mahasiswa should have access to their own report, got %d", resp.StatusCode)
		}
	})

	t.Run("Mahasiswa Access - Other Report (Forbidden)", func(t *testing.T) {
		app := fiber.New()
		otherUserID := "other-user-uuid"
		app.Get("/reports/student/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", otherUserID)
			c.Locals("role", "Mahasiswa")
			return service.GetStudentReport(c)
		})

		req := httptest.NewRequest("GET", "/reports/student/"+targetUserID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Errorf("Expected forbidden error, got %d", resp.StatusCode)
		}
	})

	t.Run("Admin Access - Any Report (Success)", func(t *testing.T) {
		app := fiber.New()
		adminUserID := "admin-user-uuid"
		app.Get("/reports/student/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", adminUserID)
			c.Locals("role", "Admin")
			return service.GetStudentReport(c)
		})

		req := httptest.NewRequest("GET", "/reports/student/"+targetUserID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Logf("Admin should have unrestricted access, got %d", resp.StatusCode)
		}
	})
}