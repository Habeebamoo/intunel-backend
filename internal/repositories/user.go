package repositories

import (
	"context"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindOrCreate(ctx context.Context, oauthUser *models.OAuthUser) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindOrCreate(ctx context.Context, oauthUser *models.OAuthUser) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).
		Where("email = ?", oauthUser.Email).
		First(&user).Error

	if err == nil {
		return &user, nil
	}

	// User doesn't exist, create them
	user = models.User{
		Name:       oauthUser.Name,
		Email:      oauthUser.Email,
		Avatar:     oauthUser.Avatar,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}