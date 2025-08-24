package repositories

import (
	"home-monitor-backend/database"
	"home-monitor-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	UserFindByUsername(username string) (*models.User, error)
	UserFindByUUID(uuid uuid.UUID) (*models.User, error)
	UserCreate(user *models.User) error
	UserUpdate(user *models.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.DB}
}

func (r *userRepository) UserFindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserFindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserFindByUUID(uuid uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UserCreate(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UserUpdate(user *models.User) error {
	return r.db.Save(user).Error
}
