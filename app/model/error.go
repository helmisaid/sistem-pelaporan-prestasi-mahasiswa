package model

import "errors"

var (
	ErrValidationFailed      = errors.New("validasi gagal")
	ErrEmptyCredentials      = errors.New("kredensial kosong")
	ErrInvalidCredentials    = errors.New("kredensial tidak valid")
	ErrUserNotFound          = errors.New("user tidak ditemukan")
	ErrDatabaseError         = errors.New("database error")
	ErrInvalidToken          = errors.New("token tidak valid")
	ErrTokenGenerationFailed = errors.New("gagal membuat token")
)

type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationError(message string) error {
	return &ValidationError{
		Message: message,
		Err:     ErrValidationFailed,
	}
}

type AuthenticationError struct {
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

func NewAuthenticationError(message string) error {
	return &AuthenticationError{
		Message: message,
		Err:     ErrInvalidCredentials,
	}
}

type NotFoundError struct {
	Message string
	Err     error
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(message string) error {
	return &NotFoundError{
		Message: message,
		Err:     ErrUserNotFound,
	}
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrEmptyCredentials)
}

func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInvalidToken)
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
