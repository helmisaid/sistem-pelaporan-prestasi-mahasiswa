package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(router fiber.Router, userSvc service.IUserService) {
	users := router.Group("/users")
	users.Use(middleware.AuthProtected())

	users.Get("/", middleware.PermissionCheck("user:read"), userSvc.GetAll)
	users.Get("/:id", middleware.PermissionCheck("user:read"), userSvc.GetByID)
	users.Post("/", middleware.PermissionCheck("user:create"), userSvc.Create)
	users.Put("/:id", middleware.PermissionCheck("user:update"), userSvc.Update)
	users.Delete("/:id", middleware.PermissionCheck("user:delete"), userSvc.Delete)
	users.Put("/:id/role", middleware.PermissionCheck("user:update"), userSvc.UpdateRole)
}