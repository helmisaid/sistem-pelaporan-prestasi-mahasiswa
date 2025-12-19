package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterStudentRoutes(router fiber.Router, studentSvc service.IStudentService, achSvc service.IAchievementService) {
	students := router.Group("/students")
	students.Use(middleware.AuthProtected())

	students.Get("/", middleware.PermissionCheck("student:read"), studentSvc.GetAll)
	students.Get("/:id", middleware.PermissionCheck("student:read"), studentSvc.GetByID)
	students.Put("/:id/advisor", middleware.PermissionCheck("student:update"), studentSvc.UpdateAdvisor)

	students.Get("/:id/achievements", middleware.PermissionCheck("student:read"), achSvc.GetByStudent)
}
