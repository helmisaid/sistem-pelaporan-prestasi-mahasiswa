package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
)

func TestStudentService_GetAll(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}
	service := NewStudentService(mockRepo, mockLecturerSvc)

	app.Get("/students", service.GetAll)

	t.Run("GET - List Students Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/students?page=1&limit=10", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestStudentService_GetByID(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}
	service := NewStudentService(mockRepo, mockLecturerSvc)

	mockRepo.detailID = "student-1"

	app.Get("/students/:id", service.GetByID)

	t.Run("GET - Student By ID Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/students/student-1", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})

	t.Run("GET - Student By ID Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/students/nonexistent", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Errorf("Expected error status for non-existent student, got %d", resp.StatusCode)
		}
	})
}

func TestStudentService_UpdateAdvisor(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockStudentRepository{
		students: make(map[string]*model.StudentInfo),
	}
	mockLecturerSvc := &MockLecturerService{}
	service := NewStudentService(mockRepo, mockLecturerSvc)

	mockRepo.detailID = "student-1"
	mockLecturerSvc.exists = true

	app.Put("/students/:id/advisor", service.UpdateAdvisor)

	t.Run("PUT - Update Advisor Success", func(t *testing.T) {
		advisorID := "lecturer-uuid-1"
		reqBody := model.UpdateAdvisorRequest{AdvisorID: &advisorID}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/students/student-1/advisor", bytes.NewReader(body))
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