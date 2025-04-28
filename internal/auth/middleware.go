package auth

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type SessionStore interface {
	BlacklistToken(ctx context.Context, token string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// AuthMiddleware parses and validates the JWT, then checks blacklist.
func AuthMiddleware(secret []byte, store SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokStr := parts[1]
		tok, err := jwt.Parse(tokStr, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// blacklist check
		if black, _ := store.IsBlacklisted(c.Request.Context(), tokStr); black {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := tok.Claims.(jwt.MapClaims)
		c.Set("userID", int64(claims["sub"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

// RequireRole enforces a specific role in context.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
