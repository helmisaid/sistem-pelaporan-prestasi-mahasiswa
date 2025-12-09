package model

import "errors"

// Custom error types
var (
	// Validation errors
	ErrValidationFailed = errors.New("validasi gagal")
	ErrEmptyCredentials = errors.New("username dan password wajib diisi")

	// Authentication errors
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrInvalidToken       = errors.New("token tidak valid atau kadaluwarsa")

	// Not found errors
	ErrUserNotFound = errors.New("user tidak ditemukan")

	// Internal errors
	ErrDatabaseError      = errors.New("terjadi kesalahan database")
	ErrTokenGenerationFailed = errors.New("gagal membuat token")
)

// IsValidationError 
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrEmptyCredentials)
}

// IsAuthenticationError 
func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInvalidToken)
}

// IsNotFoundError 
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
