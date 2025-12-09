package helper

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	if model.IsValidationError(err) {
		return BadRequest(c, err.Error(), nil)
	}

	if model.IsAuthenticationError(err) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	if model.IsNotFoundError(err) {
		return NotFound(c, err.Error())
	}

	return InternalServerError(c, "Terjadi kesalahan internal pada server")
}