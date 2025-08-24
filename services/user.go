package services

import (
	"errors"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"home-monitor-backend/utils"
	"log"

	"github.com/google/uuid"
)

type UserService interface {
	UserRegister(input models.UserRegisterRequest) (*models.User, error)
	UserLogin(input models.UserLoginRequest) (*models.User, string, error)
	UserProfile(userUUID uuid.UUID) (*models.User, error)
	UserUpdate(userUUID uuid.UUID, userUpdate *models.UserUpdateRequest) (*models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) UserRegister(input models.UserRegisterRequest) (*models.User, error) {
	_, err := s.userRepo.UserFindByUsername(input.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	user := &models.User{
		UUID:     uuid.New(),
		Username: input.Username,
		Password: input.Password,
	}

	if err := s.userRepo.UserCreate(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UserLogin(input models.UserLoginRequest) (*models.User, string, error) {
	user, err := s.userRepo.UserFindByUsername(input.Username)
	if err != nil || !user.CheckPassword(input.Password) {
		return nil, "", errors.New("invalid username or password")
	}

	token, err := utils.GenerateJWT(user.UUID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *userService) UserProfile(userUUID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) UserUpdate(userUUID uuid.UUID, userUpdate *models.UserUpdateRequest) (*models.User, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return user, errors.New("user not found")
	}

	if userUpdate.Username != "" {
		existingUser, err := s.userRepo.UserFindByUsername(userUpdate.Username)
		if err == nil && existingUser.ID != user.ID {
			return user, errors.New("username already exists")
		}

		if userUpdate.Username == user.Username {
			return user, errors.New("new username must be different from the current one")
		}

		user.Username = userUpdate.Username
	}

	if userUpdate.Password != "" {
		if user.CheckPassword(userUpdate.Password) {
			return user, errors.New("new password must be different from the current one")
		}

		user.Password = userUpdate.Password
		if err := user.HashPassword(); err != nil {
			log.Println("Error hashing password:", err)
			return user, errors.New("failed to update password")
		}

	}

	return user, s.userRepo.UserUpdate(user)
}
