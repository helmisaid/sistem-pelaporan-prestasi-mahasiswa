package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterReportRoutes(router fiber.Router, reportSvc service.IReportService) {
	rep := router.Group("/reports", middleware.AuthProtected())

	rep.Get("/statistics", middleware.PermissionCheck("report:view_global"), reportSvc.GetDashboardStats)
	rep.Get("/student/:id", middleware.PermissionCheck("report:view_student"), reportSvc.GetStudentReport)
}