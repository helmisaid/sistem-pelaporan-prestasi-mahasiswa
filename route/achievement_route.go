package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAchievementRoutes(router fiber.Router, achSvc service.IAchievementService) {
	ach := router.Group("/achievements")
	ach.Use(middleware.AuthProtected())

	// GET List 
	ach.Get("/", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		roleName := c.Locals("role").(string)

		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		search := c.Query("search", "")
		status := c.Query("status", "")

		result, err := achSvc.GetAll(c.Context(), userID, roleName, page, limit, search, status)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Daftar prestasi berhasil diambil", result)
	})

	// GET Detail
	ach.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		userID := c.Locals("user_id").(string)
		roleName := c.Locals("role").(string)

		result, err := achSvc.GetDetail(c.Context(), id, userID, roleName)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Detail prestasi berhasil diambil", result)
	})

	// Create Achievement
	ach.Post("/", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		var req model.CreateAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		result, err := achSvc.Create(c.Context(), userID, req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Created(c, "Prestasi berhasil dibuat (Draft). Silakan upload bukti.", result)
	})

	// Edit Achievement 
	ach.Put("/:id", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		userID := c.Locals("user_id").(string)

		var req model.UpdateAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format data tidak valid", nil)
		}

		err := achSvc.Edit(c.Context(), id, userID, req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Prestasi berhasil diperbarui", nil)
	})

	// Submit Achievement 
	ach.Post("/:id/submit", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		userID := c.Locals("user_id").(string)

		err := achSvc.Submit(c.Context(), id, userID)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Prestasi berhasil disubmit ke Dosen Wali", nil)
	})

	// Delete Achievement
	ach.Delete("/:id", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		userID := c.Locals("user_id").(string)

		err := achSvc.Delete(c.Context(), id, userID)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Prestasi berhasil dihapus", nil)
	})

	// Verify Achievement
    ach.Post("/:id/verify", middleware.PermissionCheck("achievement:verify"), func(c *fiber.Ctx) error {
        id := c.Params("id")
        userID := c.Locals("user_id").(string)

        var req model.VerifyAchievementRequest
        if err := c.BodyParser(&req); err != nil {
            return helper.BadRequest(c, "Format data tidak valid (poin diperlukan)", nil)
        }

        err := achSvc.Verify(c.Context(), id, userID, req)
        if err != nil {
            return helper.HandleError(c, err)
        }

        return helper.Success(c, "Prestasi berhasil diverifikasi dan poin disimpan", nil)
    })

    // Reject Achievement
    ach.Post("/:id/reject", middleware.PermissionCheck("achievement:verify"), func(c *fiber.Ctx) error {
        id := c.Params("id")
        userID := c.Locals("user_id").(string)

        var req model.RejectAchievementRequest
        if err := c.BodyParser(&req); err != nil {
            return helper.BadRequest(c, "Format data tidak valid (catatan penolakan diperlukan)", nil)
        }

        err := achSvc.Reject(c.Context(), id, userID, req)
        if err != nil {
            return helper.HandleError(c, err)
        }

        return helper.Success(c, "Prestasi ditolak dan dikembalikan ke mahasiswa", nil)
    })

	// Upload Attachment 
	ach.Post("/:id/attachments", middleware.PermissionCheck("achievement:create"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		userID := c.Locals("user_id").(string)

		fileHeader, err := c.FormFile("file")
		if err != nil {
			return helper.BadRequest(c, "File tidak ditemukan.", nil)
		}

		result, err := achSvc.UploadAttachment(c.Context(), id, userID, fileHeader)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "File berhasil diupload", result)
	})
}

