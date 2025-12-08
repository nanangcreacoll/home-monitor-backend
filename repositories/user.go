package repositories

import (
	"context"
	"home-monitor-backend/database"
	"home-monitor-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	UserFindByUsername(ctx context.Context, username string) (*models.User, error)
	UserFindByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error)
	UserFindByID(ctx context.Context, id uint) (*models.User, error)
	UserCreate(ctx context.Context, user *models.User) error
	UserUpdate(ctx context.Context, user *models.User) error
	UserDelete(ctx context.Context, user *models.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.DB}
}

func (r *userRepository) UserFindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserFindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserFindByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserCreate(ctx context.Context, user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UserUpdate(ctx context.Context, user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UserDelete(ctx context.Context, user *models.User) error {
	return r.db.Delete(user).Error
}
