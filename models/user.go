package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id" validate:"required"`
	Username  string    `gorm:"unique" json:"username" validate:"required,lte=255"`
	Password  string    `json:"password,omitempty" validate:"required,lte=255"`
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
	ID        uint      `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required,lte=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type UserLoginResponse struct {
	ID       uint   `json:"id" validate:"required"`
	Username string `json:"username" validate:"required,lte=255"`
	Token    string `json:"token" validate:"required"`
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
