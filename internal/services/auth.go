package services

import (
	"context"
	"fmt"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/repositories"
	"github.com/Habeebamoo/intunel-backend/internal/utils"
)

type AuthService interface {
	HandleOAuth(ctx context.Context, oauthUser *models.OAuthUser) (*models.AuthResponse, error)
}

type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *authService) HandleOAuth(ctx context.Context, oauthUser *models.OAuthUser) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindOrCreate(ctx, oauthUser)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	token, err := utils.GenerateToken(user.UserID.String(), user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{Token: token, User: *user}, nil
}