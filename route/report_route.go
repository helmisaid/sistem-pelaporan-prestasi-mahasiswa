package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterReportRoutes(router fiber.Router, reportSvc service.IReportService) {
	rep := router.Group("/reports", middleware.AuthProtected())

	rep.Get("/statistics", middleware.PermissionCheck("report:view_global"), func(c *fiber.Ctx) error {
		result, err := reportSvc.GetDashboardStats(c.Context())
		if err != nil {
			return helper.HandleError(c, err)
		}
		return helper.Success(c, "Statistik dashboard berhasil diambil", result)
	})

	rep.Get("/student/:id", middleware.PermissionCheck("report:view_student"), func(c *fiber.Ctx) error {
		targetUserID := c.Params("id")
		viewerUserID := c.Locals("user_id").(string)
		viewerRole := c.Locals("role").(string)

		result, err := reportSvc.GetStudentReport(c.Context(), targetUserID, viewerUserID, viewerRole)
		if err != nil {
			return helper.HandleError(c, err)
		}
		return helper.Success(c, "Laporan prestasi mahasiswa berhasil diambil", result)
	})
}