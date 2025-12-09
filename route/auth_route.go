package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(router fiber.Router, authSvc service.IAuthService) {
	auth := router.Group("/auth")

	// Login
	auth.Post("/login", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		resp, err := authSvc.Login(c.Context(), req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Login berhasil", resp)
	})

	// Refresh Token
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		var req model.RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		resp, err := authSvc.RefreshToken(c.Context(), req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Token berhasil diperbarui", resp)
	})

	// Logout
	auth.Post("/logout", middleware.AuthProtected(), func(c *fiber.Ctx) error {
		_ = authSvc.Logout(c.Context())
		return helper.Success(c, "Berhasil logout", nil)
	})

	// Profile
	auth.Get("/profile", middleware.AuthProtected(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		user, err := authSvc.GetProfile(c.Context(), userID)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Profil berhasil diambil", user)
	})
}