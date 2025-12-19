package service

import (
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/utils"

	"github.com/gofiber/fiber/v2"
)

type IAuthService interface {
	Login(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	GetProfile(c *fiber.Ctx) error
}

type AuthService struct {
	repo repository.IAuthRepository
}

func NewAuthService(repo repository.IAuthRepository) IAuthService {
	return &AuthService{repo: repo}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with username and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} helper.Response{data=model.LoginResponse} "Login successful"
// @Failure 400 {object} helper.ErrorResponse "Invalid request format"
// @Failure 401 {object} helper.ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	if req.Username == "" || req.Password == "" {
		return helper.HandleError(c, model.ErrEmptyCredentials)
	}

	user, err := s.repo.GetUserByUsername(c.Context(), req.Username)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrInvalidCredentials)
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return helper.HandleError(c, model.NewAuthenticationError("username atau password salah"))
	}

	accessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Username, user.Role.Name, user.Permissions)
	if err != nil {
		return helper.HandleError(c, model.ErrTokenGenerationFailed)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID.String(), user.Username, user.Role.Name)
	if err != nil {
		return helper.HandleError(c, model.ErrTokenGenerationFailed)
	}

	resp := &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user.ToLoginDTO(),
	}

	return helper.Success(c, "Login berhasil", resp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} helper.Response{data=model.LoginResponse} "Token refreshed"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 401 {object} helper.ErrorResponse "Invalid or expired refresh token"
// @Router /auth/refresh [post]
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		return helper.HandleError(c, model.ErrInvalidToken)
	}

	user, err := s.repo.GetUserByID(c.Context(), claims.UserID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Username, user.Role.Name, user.Permissions)
	if err != nil {
		return helper.HandleError(c, model.ErrTokenGenerationFailed)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID.String(), user.Username, user.Role.Name)
	if err != nil {
		return helper.HandleError(c, model.ErrTokenGenerationFailed)
	}

	resp := &model.RefreshTokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return helper.Success(c, "Token berhasil diperbarui", resp)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current authenticated user's profile information
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=model.UserProfileDTO} "Profile retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Router /auth/profile [get]
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := s.repo.GetUserByID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	profileDTO := user.ToProfileDTO()
	return helper.Success(c, "Profil berhasil diambil", profileDTO)
}

// Logout godoc
// @Summary Logout user
// @Description Logout current user (placeholder for token invalidation)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response "Logout successful"
// @Router /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	return helper.Success(c, "Berhasil logout", nil)
}