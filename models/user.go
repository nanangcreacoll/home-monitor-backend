package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole uint8

const (
	UserRoleAdmin UserRole = 1
	UserRoleUser  UserRole = 2
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id" validate:"required"`
	UUID      uuid.UUID `gorm:"unique" json:"uuid" validate:"required,uuid"`
	Username  string    `gorm:"unique" json:"username" validate:"required,lte=255"`
	Password  string    `json:"password,omitempty" validate:"required,lte=255"`
	Role      UserRole  `gorm:"type:TINYINT;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

type UserRegisterResponse struct {
	UUID      uuid.UUID `json:"uuid" validate:"required,uuid"`
	Username  string    `json:"username" validate:"required,lte=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type UserLoginResponse struct {
	UUID     uuid.UUID `json:"uuid" validate:"required,uuid"`
	Username string    `json:"username" validate:"required,lte=255"`
	Token    string    `json:"token" validate:"required"`
}

type UserProfileResponse struct {
	UUID      uuid.UUID `json:"uuid" validate:"required,uuid"`
	Username  string    `json:"username" validate:"required,lte=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type UserUpdateRequest struct {
	Username    string    `json:"username" binding:"omitempty,min=3,max=255" validate:"required"`
	NewUsername string    `json:"new_username" binding:"omitempty,min=3,max=255"`
	Password    string    `json:"password" binding:"omitempty,min=6,max=255" validate:"required"`
	NewPassword string    `json:"new_password" binding:"omitempty,min=6,max=255"`
	Role        *UserRole `json:"role" binding:"omitempty,oneof=1 2"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}

	if u.Username == "" {
		return errors.New("username cannot be empty")
	}

	if u.Password == "" {
		return errors.New("password cannot be empty")
	}

	if u.Role != UserRoleAdmin && u.Role != UserRoleUser {
		u.Role = UserRoleUser
	}

	if err := u.HashPassword(); err != nil {
		return err
	}

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
