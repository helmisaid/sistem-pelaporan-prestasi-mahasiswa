package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
)

func TestAchievementService_Create(t *testing.T) {
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{
		achRefs: make(map[string]*model.AchievementReference),
	}
	mockStudentRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}

	service := NewAchievementService(mockAchRepo, mockStudentRepo, mockLecturerSvc)

	userID := "user-mhs-1"
	studentID := "student-1"
	mockStudentRepo.students[userID] = &model.StudentInfo{ID: studentID}

	app.Post("/achievements", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return service.Create(c)
	})

	t.Run("POST - Create Achievement Success", func(t *testing.T) {
		reqBody := model.CreateAchievementRequest{
			Title:           "Juara 1 Hackathon",
			AchievementType: "Nasional",
			Description:     "Test description",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusCreated && resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 201 or 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestAchievementService_Submit(t *testing.T) {
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{
		achRefs: make(map[string]*model.AchievementReference),
	}
	mockStudentRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}

	service := NewAchievementService(mockAchRepo, mockStudentRepo, mockLecturerSvc)

	userID := "user-mhs-1"
	studentID := "student-1"
	achID := "ach-1"

	mockStudentRepo.students[userID] = &model.StudentInfo{ID: studentID}
	mockAchRepo.achRefs[achID] = &model.AchievementReference{
		ID: achID, StudentID: studentID, Status: "draft",
	}

	app.Post("/achievements/:id/submit", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return service.Submit(c)
	})

	t.Run("POST - Submit Achievement Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/achievements/"+achID+"/submit", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestAchievementService_Verify(t *testing.T) {
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{
		achRefs: make(map[string]*model.AchievementReference),
	}
	mockStudentRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}

	service := NewAchievementService(mockAchRepo, mockStudentRepo, mockLecturerSvc)

	lecturerUserID := "user-dosen-1"
	lecturerID := "dosen-1"
	achID := "ach-1"

	mockLecturerSvc.lecturerInfo = &model.LecturerInfo{ID: lecturerID}
	mockAchRepo.achRefs[achID] = &model.AchievementReference{ID: achID, Status: "submitted"}
	mockAchRepo.achDetail = &model.AchievementDetailDTO{
		Student: model.StudentListDTO{AdvisorID: &lecturerID},
	}

	app.Post("/achievements/:id/verify", func(c *fiber.Ctx) error {
		c.Locals("user_id", lecturerUserID)
		return service.Verify(c)
	})

	t.Run("POST - Verify Achievement by Advisor", func(t *testing.T) {
		reqBody := model.VerifyAchievementRequest{Points: 100}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/achievements/"+achID+"/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestAchievementService_Delete(t *testing.T) {
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{
		achRefs: make(map[string]*model.AchievementReference),
	}
	mockStudentRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}

	service := NewAchievementService(mockAchRepo, mockStudentRepo, mockLecturerSvc)

	userID := "user-mhs-1"
	studentID := "student-1"
	achID := "ach-1"

	mockStudentRepo.students[userID] = &model.StudentInfo{ID: studentID}
	mockAchRepo.achRefs[achID] = &model.AchievementReference{
		ID: achID, StudentID: studentID, Status: "draft",
	}

	app.Delete("/achievements/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return service.Delete(c)
	})

	t.Run("DELETE - Soft Delete Achievement Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/achievements/"+achID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}