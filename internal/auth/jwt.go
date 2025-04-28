package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
	"user-service/internal/model"
)

// GenerateToken creates a signed JWT for the user.
func GenerateToken(u *model.User, secret []byte, expire time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"exp":  time.Now().Add(expire).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
