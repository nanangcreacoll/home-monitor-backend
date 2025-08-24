package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JWTSecret = os.Getenv("JWT_SECRET")

func GenerateJWT(userUUID uuid.UUID) (string, error) {
	Expiration, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil {
		return "", err
	}
	ExpirationTime := time.Now().Add(time.Duration(Expiration) * time.Minute).Unix()

	claim := jwt.MapClaims{
		"user_uuid": userUUID.String(),
		"exp":       ExpirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(JWTSecret))
}

func ValidateJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return uuid.Parse(claims["user_uuid"].(string))
	}

	return uuid.Nil, err
}
