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
	UserCreate(ctx context.Context, user *models.User) error
	UserUpdate(ctx context.Context, user *models.User) error
	UserDelete(ctx context.Context, user *models.User) error
	UserList(ctx context.Context, length int) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.DB}
}

const defaultUserListLimit = 50

func (r *userRepository) UserFindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserFindByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserCreate(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) UserUpdate(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UserDelete(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Delete(user).Error
}

func (r *userRepository) UserList(ctx context.Context, length int) ([]models.User, error) {
	var users []models.User
	query := r.db.WithContext(ctx)
	if length > 0 {
		query = query.Limit(length)
	} else {
		query = query.Limit(defaultUserListLimit)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
