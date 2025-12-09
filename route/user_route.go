package route

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(router fiber.Router, userSvc service.IUserService) {
	users := router.Group("/users")
	users.Use(middleware.AuthProtected())

	// Get All Users
	users.Get("/", middleware.PermissionCheck("user:read"), func(c *fiber.Ctx) error {
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		search := c.Query("search", "")
		sortBy := c.Query("sortBy", "created_at")
		sortOrder := c.Query("order", "desc")

		result, err := userSvc.GetAll(c.Context(), page, limit, search, sortBy, sortOrder)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Daftar user berhasil diambil", result)
	})

	// Get User By ID
	users.Get("/:id", middleware.PermissionCheck("user:read"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		user, err := userSvc.GetByID(c.Context(), id)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Detail user berhasil diambil", user)
	})

	// Create User
	users.Post("/", middleware.PermissionCheck("user:create"), func(c *fiber.Ctx) error {
		var req model.CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		user, err := userSvc.Create(c.Context(), req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Created(c, "User berhasil dibuat", user)
	})

	// Update User Profile
	users.Put("/:id", middleware.PermissionCheck("user:update"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req model.UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		user, err := userSvc.Update(c.Context(), id, req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "User berhasil diupdate", user)
	})

	// Delete User (Soft Delete)
	users.Delete("/:id", middleware.PermissionCheck("user:delete"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		err := userSvc.Delete(c.Context(), id)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "User berhasil dihapus", nil)
	})

	// Update User Role 
	users.Put("/:id/role", middleware.PermissionCheck("user:update"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req model.UpdateRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest(c, "Format request tidak valid", nil)
		}

		err := userSvc.UpdateRole(c.Context(), id, req)
		if err != nil {
			return helper.HandleError(c, err)
		}

		return helper.Success(c, "Role user berhasil diupdate", nil)
	})
}