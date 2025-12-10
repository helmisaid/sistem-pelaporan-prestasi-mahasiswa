package helper

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
    if err == nil {
        return nil
    }

    // Validation 400
    if model.IsValidationError(err) {
        return BadRequest(c, err.Error(), nil) 
    }

    // Auth 401
    if model.IsAuthenticationError(err) {
        return Unauthorized(c, err.Error()) 
    }

    // Not Found 404
    if model.IsNotFoundError(err) {
        return NotFound(c, err.Error())
    }

    // Default 500
    return InternalServerError(c, "Terjadi kesalahan internal pada server")
}