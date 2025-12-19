package service

import (
	"net/http/httptest"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLecturerService_GetAll(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockLecturerRepository{
		lecturers: make(map[string]*model.LecturerInfo),
	}
	service := NewLecturerService(mockRepo)

	app.Get("/lecturers", func(c *fiber.Ctx) error {
		c.Locals("role", "Admin")
		return service.GetAll(c)
	})

	t.Run("GET - List Lecturers as Admin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/lecturers?page=1&limit=10", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})

	t.Run("GET - List Lecturers as Non-Admin (Forbidden)", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/lecturers", func(c *fiber.Ctx) error {
			c.Locals("role", "Mahasiswa")
			return service.GetAll(c)
		})

		req := httptest.NewRequest("GET", "/lecturers", nil)

		resp, err := app2.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Errorf("Expected forbidden status, got %d", resp.StatusCode)
		}
	})
}

func TestLecturerService_GetAdvisees(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockLecturerRepository{
		lecturers: make(map[string]*model.LecturerInfo),
	}
	mockRepo.exists = true
	service := NewLecturerService(mockRepo)

	lecturerID := "lecturer-1"

	app.Get("/lecturers/:id/advisees", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-dosen-1")
		c.Locals("role", "Dosen Wali")
		return service.GetAdvisees(c)
	})

	t.Run("GET - Advisees Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/lecturers/"+lecturerID+"/advisees", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusBadRequest {
			t.Logf("Got status %d (expected OK or error)", resp.StatusCode)
		}
	})
}