package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestAuthService_Login(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockAuthRepository{users: make(map[string]*model.User)}
	service := NewAuthService(mockRepo)

	uID := uuid.New()
	mockRepo.users[uID.String()] = &model.User{
		ID:           uID,
		Username:     "testuser",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Role:         model.Role{Name: "Mahasiswa"},
	}

	app.Post("/login", service.Login)

	t.Run("POST - Login Empty Credentials", func(t *testing.T) {
		reqBody := map[string]string{
			"username": "",
			"password": "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusBadRequest && resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("Expected 400 or 401 status, got %d", resp.StatusCode)
		}
	})

	t.Run("POST - Login Unknown User", func(t *testing.T) {
		reqBody := map[string]string{
			"username": "unknown",
			"password": "any",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusUnauthorized && resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected 401 or 400 status, got %d", resp.StatusCode)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockAuthRepository{users: make(map[string]*model.User)}
	service := NewAuthService(mockRepo)

	app.Post("/refresh", service.RefreshToken)

	t.Run("POST - Invalid Refresh Token", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": "invalid-token-format",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode == fiber.StatusOK {
			t.Error("Expected error status for invalid token, got 200")
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockAuthRepository{users: make(map[string]*model.User)}
	service := NewAuthService(mockRepo)

	app.Post("/logout", func(c *fiber.Ctx) error {
		c.Locals("user_id", "test-user-id")
		return service.Logout(c)
	})

	t.Run("POST - Logout Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected 200 status, got %d", resp.StatusCode)
		}
	})
}

func TestAuthService_GetProfile(t *testing.T) {
	app := fiber.New()
	mockRepo := &MockAuthRepository{users: make(map[string]*model.User)}
	service := NewAuthService(mockRepo)

	uID := uuid.New()
	mockRepo.users[uID.String()] = &model.User{
		ID:           uID,
		Username:     "testuser",
		PasswordHash: "$2a$10$dummyhash",
		Role:         model.Role{Name: "Mahasiswa"},
		FullName:     "Test User",
		Email:        "test@example.com",
	}

	app.Get("/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", uID.String())
		return service.GetProfile(c)
	})

	t.Run("GET - Profile Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)

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