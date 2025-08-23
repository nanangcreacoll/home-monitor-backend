package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = os.Getenv("JWT_SECRET")

func GenerateJWT(userID uint) (string, error) {
	Expiration, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil {
		return "", err
	}
	ExpirationTime := time.Now().Add(time.Duration(Expiration) * time.Minute).Unix()

	claim := jwt.MapClaims{
		"user_id": userID,
		"exp":     ExpirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(JWTSecret))
}

func ValidateJWT(tokenString string) (uint, error) {
	tokenString = tokenString[len("Bearer "):]
	if tokenString == "" {
		return 0, jwt.ErrSignatureInvalid
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return uint(claims["user_id"].(float64)), nil
	}

	return 0, err
}
