package services

import (
	"errors"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"home-monitor-backend/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	UserRegister(input models.UserRegisterRequest, userUUID uuid.UUID) (*models.User, error)
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

func (s *userService) UserRegister(input models.UserRegisterRequest, userUUID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return nil, errors.New("only admin can register new users")
	}

	_, err = s.userRepo.UserFindByUsername(input.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	newUser := &models.User{
		UUID:     uuid.New(),
		Username: input.Username,
		Password: input.Password,
		Role:     models.UserRoleUser,
	}

	if err := s.userRepo.UserCreate(newUser); err != nil {
		return nil, err
	}
	return newUser, nil
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

	if userUpdate.Username == "" && userUpdate.Password == "" {
		return user, errors.New("need to provide username or password to update")
	}

	userToUpdate, err := s.userRepo.UserFindByUsername(userUpdate.Username)
	if err != nil {
		return user, errors.New("username or password is incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userToUpdate.Password), []byte(userUpdate.Password)); err != nil {
		return user, errors.New("username or password is incorrect")
	}

	if userUpdate.NewUsername == "" && userUpdate.NewPassword == "" && userUpdate.Role == nil {
		return user, errors.New("need to provide at least one field to update")
	}

	if userUpdate.NewUsername != "" {
		userToUpdate.Username = userUpdate.NewUsername
	}

	if userUpdate.NewPassword != "" {
		userToUpdate.Password = userUpdate.NewPassword
		err = userToUpdate.HashPassword()
		if err != nil {
			return user, errors.New("failed to update")
		}
	}

	if userUpdate.Role != nil {
		if user.Role != models.UserRoleAdmin {
			return user, errors.New("only admin can update role")
		} else if user.Role == models.UserRoleAdmin && user.UUID == userToUpdate.UUID {
			return user, errors.New("admin cannot change their own role")
		}
		userToUpdate.Role = *userUpdate.Role
	}

	return userToUpdate, s.userRepo.UserUpdate(userToUpdate)
}
