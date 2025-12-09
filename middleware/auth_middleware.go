package middleware

import (
	"sistem-pelaporan-prestasi-mahasiswa/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// validasi token JWT
func AuthProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token tidak ditemukan",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Format token salah",
			})
		}

		tokenString := parts[1]

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token tidak valid atau kadaluwarsa",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}

// RBAC
func PermissionCheck(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permsString, ok := c.Locals("permissions").([]string)
        
        if !ok {
             if permsInterface, ok := c.Locals("permissions").([]interface{}); ok {
                 for _, p := range permsInterface {
                     if p.(string) == requiredPermission {
                         return c.Next() 
                     }
                 }
             }
             return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Akses ditolak"})
        }

		for _, p := range permsString {
			if p == requiredPermission {
				return c.Next() 
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden: Anda tidak memiliki permission '" + requiredPermission + "'",
		})
	}
}