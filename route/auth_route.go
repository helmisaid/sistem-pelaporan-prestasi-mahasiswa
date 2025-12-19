package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(router fiber.Router, authSvc service.IAuthService) {
	auth := router.Group("/auth")

	auth.Post("/login", authSvc.Login)
	auth.Post("/refresh", authSvc.RefreshToken)
	auth.Post("/logout", middleware.AuthProtected(), authSvc.Logout)
	auth.Get("/profile", middleware.AuthProtected(), authSvc.GetProfile)
}