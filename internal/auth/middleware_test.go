package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/enson89/user-service-go/internal/auth"
	authmocks "github.com/enson89/user-service-go/internal/auth/mocks"
	"github.com/enson89/user-service-go/internal/model"
)

func TestAuthMiddleware_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	m := auth.AuthMiddleware([]byte("secret"), nil)
	m(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token")

	m := auth.AuthMiddleware([]byte("secret"), nil)
	m(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Blacklisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Generate a valid token
	u := &model.User{ID: 5, Role: "user"}
	secret := []byte("s3cr3t")
	tok, _ := auth.GenerateToken(u, secret, time.Minute)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tok)

	// Mock the store to return blacklisted
	store := new(authmocks.MockSessionStore)
	store.On("IsBlacklisted", mock.Anything, tok).Return(true, nil)

	m := auth.AuthMiddleware(secret, store)
	m(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	store.AssertExpectations(t)
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Valid token
	u := &model.User{ID: 7, Role: "admin"}
	secret := []byte("topsecret")
	tok, _ := auth.GenerateToken(u, secret, time.Minute)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tok)

	// Mock store returns not blacklisted
	store := new(authmocks.MockSessionStore)
	store.On("IsBlacklisted", mock.Anything, tok).Return(false, nil)

	m := auth.AuthMiddleware(secret, store)
	m(c)

	assert.False(t, c.IsAborted())
	// context keys
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	assert.Equal(t, int64(7), userID)
	assert.Equal(t, "admin", role)
	store.AssertExpectations(t)
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("role", "user")

	auth.RequireRole("admin")(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Success path
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("role", "admin")
	auth.RequireRole("admin")(c)
	assert.False(t, c.IsAborted())
}
