package auth

import (
	"time"

	"github.com/enson89/user-service-go/internal/model"
	"github.com/golang-jwt/jwt/v5"
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
