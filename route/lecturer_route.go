package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterLecturerRoutes(router fiber.Router, lecturerSvc service.ILecturerService) {
	lecturers := router.Group("/lecturers", middleware.AuthProtected())

	// GET /api/v1/lecturers - List all lecturers (Admin only)
	lecturers.Get("/", func(c *fiber.Ctx) error {
		// Only Admin can list all lecturers
		viewerRole := c.Locals("role").(string)
		if viewerRole != "Admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Anda tidak memiliki hak akses.",
			})
		}

		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "u.created_at")
		sortOrder := c.Query("sort_order", "DESC")

		result, err := lecturerSvc.GetAll(c.Context(), page, limit, search, sortBy, sortOrder)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Daftar dosen berhasil diambil", result)
	})

	// GET /api/v1/lecturers/:id/advisees - Get students supervised by lecturer
	lecturers.Get("/:id/advisees", middleware.PermissionCheck("lecturer:read"), func(c *fiber.Ctx) error {
		lecturerID := c.Params("id")
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)

		viewerUserID := c.Locals("user_id").(string)
		viewerRole := c.Locals("role").(string)

		result, err := lecturerSvc.GetAdvisees(c.Context(), lecturerID, viewerUserID, viewerRole, page, limit)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Daftar mahasiswa bimbingan berhasil diambil", result)
	})
}
