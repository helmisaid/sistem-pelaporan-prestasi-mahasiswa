package helper

import "github.com/gofiber/fiber/v2"

type MetaInfo struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(MetaInfo{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(MetaInfo{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *fiber.Ctx, message string, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(MetaInfo{
		Status:  "error",
		Message: message,
		Errors:  errors,
	})
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}

func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(MetaInfo{
		Status:  "error",
		Message: message,
	})
}
