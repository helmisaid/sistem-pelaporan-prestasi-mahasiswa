package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterStudentRoutes(router fiber.Router, studentSvc service.IStudentService, achSvc service.IAchievementService) {
	students := router.Group("/students")
	students.Use(middleware.AuthProtected())

	// Get All Students 
	students.Get("/", middleware.PermissionCheck("student:read"), func(c *fiber.Ctx) error {
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		search := c.Query("search", "")
		sortBy := c.Query("sortBy", "u.created_at")
		sortOrder := c.Query("order", "desc")

		result, err := studentSvc.GetAll(c.Context(), page, limit, search, sortBy, sortOrder)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Daftar mahasiswa berhasil diambil", result)
	})

	// Get Student Detail By ID
	students.Get("/:id", middleware.PermissionCheck("student:read"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		student, err := studentSvc.GetByID(c.Context(), id)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Detail mahasiswa berhasil diambil", student)
	})

	// Update Student Advisor
	students.Put("/:id/advisor", middleware.PermissionCheck("student:update"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req model.UpdateAdvisorRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		err := studentSvc.UpdateAdvisor(c.Context(), id, req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Dosen wali berhasil diupdate", nil)
	})

	// Get student achievements
	students.Get("/:id/achievements", middleware.PermissionCheck("student:read"), func(c *fiber.Ctx) error {
        targetUserID := c.Params("id")
        
        viewerUserID := c.Locals("user_id").(string)
        viewerRole := c.Locals("role").(string)

        page := c.QueryInt("page", 1)
        limit := c.QueryInt("limit", 10)
        status := c.Query("status", "")

        result, err := achSvc.GetByStudent(c.Context(), targetUserID, viewerUserID, viewerRole, page, limit, status)
        if err != nil {
            return helper.HandleError(c, err)
        }

        return helper.Success(c, "Daftar prestasi mahasiswa berhasil diambil", result)
    })
}
