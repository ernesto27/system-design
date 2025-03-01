package utils

import (
	"time"
	"userservice/config"
	"userservice/models"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret []byte

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWTSecret)
}

func GenerateJWT(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
