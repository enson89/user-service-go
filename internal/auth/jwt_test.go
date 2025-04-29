package auth_test

import (
	"github.com/enson89/user-service-go/internal/auth"
	"github.com/enson89/user-service-go/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	u := &model.User{ID: 123, Role: "user"}
	secret := []byte("s3cr3t")
	expire := 5 * time.Minute

	// Generate token
	tokStr, err := auth.GenerateToken(u, secret, expire)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokStr)

	// Parse token
	token, err := jwt.Parse(tokStr, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method is HMAC
		if token.Method != jwt.SigningMethodHS256 {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(123), claims["sub"])
	assert.Equal(t, "user", claims["role"])

	expVal := int64(claims["exp"].(float64))
	now := time.Now().Unix()
	assert.Greater(t, expVal, now)
	assert.LessOrEqual(t, expVal, now+int64(expire.Seconds())+1)
}
