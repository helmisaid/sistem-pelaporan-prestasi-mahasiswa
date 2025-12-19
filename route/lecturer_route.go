package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterLecturerRoutes(router fiber.Router, lecturerSvc service.ILecturerService) {
	lecturers := router.Group("/lecturers", middleware.AuthProtected())

	lecturers.Get("/", lecturerSvc.GetAll)
	lecturers.Get("/:id/advisees", middleware.PermissionCheck("lecturer:read"), lecturerSvc.GetAdvisees)
}
