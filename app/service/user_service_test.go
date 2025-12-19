package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestUserService_Create(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockUserRepository{
		users: make(map[string]*model.User),
		roles: make(map[string]*model.Role),
	}
	mockStudentSvc := &MockStudentService{}
	mockLecturerSvc := &MockLecturerService{}

	service := NewUserService(mockRepo, mockStudentSvc, mockLecturerSvc, &sql.DB{})

	roleID := uuid.New().String()
	mockRepo.roles[roleID] = &model.Role{ID: uuid.MustParse(roleID), Name: "Mahasiswa"}
	mockRepo.roleExists = true

	app.Post("/users", service.Create)

	t.Run("POST - Create User Duplicate Username", func(t *testing.T) {
		mockRepo.usernameExists = true

		reqBody := model.CreateUserRequest{
			Username: "existinguser",
			Email:    "test@example.com",
			Password: "password123",
			FullName: "Test User",
			RoleID:   roleID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected error status for duplicate username, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("POST - Create User Email Already Exists", func(t *testing.T) {
		mockRepo.usernameExists = false
		mockRepo.emailExists = true

		reqBody := model.CreateUserRequest{
			Username: "newuser",
			Email:    "existing@example.com",
			Password: "password123",
			FullName: "Test User",
			RoleID:   roleID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusCreated {
			t.Errorf("Expected error status for duplicate email, got %d", resp.StatusCode)
		}
	})
}

func TestUserService_Delete(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockUserRepository{
		users: make(map[string]*model.User),
		roles: make(map[string]*model.Role),
	}
	mockStudentSvc := &MockStudentService{}
	mockLecturerSvc := &MockLecturerService{}

	service := NewUserService(mockRepo, mockStudentSvc, mockLecturerSvc, &sql.DB{})

	uID := uuid.New().String()
	mockRepo.users[uID] = &model.User{
		ID:       uuid.MustParse(uID),
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	app.Delete("/users/:id", service.Delete)

	t.Run("DELETE - User Soft Delete Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/"+uID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 200 status, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("DELETE - User Not Found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		req := httptest.NewRequest("DELETE", "/users/"+nonExistentID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Errorf("Expected error status for non-existent user, got %d", resp.StatusCode)
		}
	})
}

func TestUserService_GetByID(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockUserRepository{
		users: make(map[string]*model.User),
		roles: make(map[string]*model.Role),
	}
	mockStudentSvc := &MockStudentService{}
	mockLecturerSvc := &MockLecturerService{}

	service := NewUserService(mockRepo, mockStudentSvc, mockLecturerSvc, &sql.DB{})

	uID := uuid.New().String()
	roleID := uuid.New()
	mockRepo.users[uID] = &model.User{
		ID:       uuid.MustParse(uID),
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		RoleID:   roleID,
		Role:     model.Role{ID: roleID, Name: "Admin"},
	}

	app.Get("/users/:id", service.GetByID)

	t.Run("GET - User By ID Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/"+uID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 200 status, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("GET - User By ID Not Found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		req := httptest.NewRequest("GET", "/users/"+nonExistentID, nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Errorf("Expected error status for non-existent user, got %d", resp.StatusCode)
		}
	})
}

func TestUserService_UpdateRole(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockUserRepository{
		users: make(map[string]*model.User),
		roles: make(map[string]*model.Role),
	}
	mockStudentSvc := &MockStudentService{}
	mockLecturerSvc := &MockLecturerService{}

	service := NewUserService(mockRepo, mockStudentSvc, mockLecturerSvc, &sql.DB{})

	uID := uuid.New().String()
	roleID := uuid.New()
	mockRepo.users[uID] = &model.User{
		ID:       uuid.MustParse(uID),
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		RoleID:   roleID,
		Role:     model.Role{ID: roleID, Name: "Mahasiswa"},
	}

	newRoleID := uuid.New().String()
	mockRepo.roles[newRoleID] = &model.Role{ID: uuid.MustParse(newRoleID), Name: "Admin"}

	app.Put("/users/:id/role", service.UpdateRole)

	t.Run("PUT - Update Role Success", func(t *testing.T) {
		reqBody := model.UpdateRoleRequest{
			RoleID: newRoleID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/users/"+uID+"/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 200 status, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})
}