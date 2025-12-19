package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAchievementRoutes(router fiber.Router, achSvc service.IAchievementService) {
	ach := router.Group("/achievements")
	ach.Use(middleware.AuthProtected())

	ach.Get("/", achSvc.GetAll)
	ach.Get("/:id", achSvc.GetDetail)
	ach.Post("/", middleware.PermissionCheck("achievement:create"), achSvc.Create)
	ach.Put("/:id", middleware.PermissionCheck("achievement:create"), achSvc.Edit)
	ach.Post("/:id/submit", middleware.PermissionCheck("achievement:create"), achSvc.Submit)
	ach.Delete("/:id", middleware.PermissionCheck("achievement:create"), achSvc.Delete)
	ach.Post("/:id/verify", middleware.PermissionCheck("achievement:verify"), achSvc.Verify)
	ach.Post("/:id/reject", middleware.PermissionCheck("achievement:verify"), achSvc.Reject)
	ach.Post("/:id/attachments", middleware.PermissionCheck("achievement:create"), achSvc.UploadAttachment)
	ach.Get("/:id/history", achSvc.GetByStudent)
}
