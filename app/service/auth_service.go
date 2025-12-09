package service

import (
	"context"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/utils"
)

type IAuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	Logout(ctx context.Context) error
	GetProfile(ctx context.Context, userID string) (*model.UserProfileDTO, error)
}

type AuthService struct {
	repo repository.IAuthRepository
}

func NewAuthService(repo repository.IAuthRepository) IAuthService {
	return &AuthService{repo: repo}
}

// Login
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, model.ErrEmptyCredentials
	}

	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil { 
		return nil, model.ErrDatabaseError
	}
	if user == nil { 
		return nil, model.ErrInvalidCredentials
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, model.NewAuthenticationError("username atau password salah")
	}

	accessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Username, user.Role.Name, user.Permissions)
	if err != nil { 
		return nil, model.ErrTokenGenerationFailed
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID.String(), user.Username, user.Role.Name)
	if err != nil { 
		return nil, model.ErrTokenGenerationFailed
	}

	return &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user.ToLoginDTO(),
	}, nil
}

// Refresh Token
func (s *AuthService) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, model.ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Username, user.Role.Name, user.Permissions)
	if err != nil { 
		return nil, model.ErrTokenGenerationFailed
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID.String(), user.Username, user.Role.Name)
	if err != nil { 
		return nil, model.ErrTokenGenerationFailed
	}

	return &model.RefreshTokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Get Profile
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.UserProfileDTO, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}
	profileDTO := user.ToProfileDTO()
	return &profileDTO, nil
}

func (s *AuthService) Logout(ctx context.Context) error {
	return nil
}