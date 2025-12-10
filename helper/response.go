package helper

import "github.com/gofiber/fiber/v2"

type MetaInfo struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Success returns 200 OK response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(MetaInfo{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Created returns 201 Created response
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(MetaInfo{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// BadRequest returns 400 Bad Request response
func BadRequest(c *fiber.Ctx, message string, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(MetaInfo{
		Status:  "error",
		Message: message,
		Errors:  errors,
	})
}

// Unauthorized returns 401 Unauthorized response
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

// Forbidden returns 403 Forbidden response
func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

// NotFound returns 404 Not Found response
func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

// InternalServerError returns 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}
