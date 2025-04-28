package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"user-service/internal/model"
)

type UserService interface {
	SignUp(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, id int64) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateUser(ctx context.Context, id int64, newName string) (*model.User, error)
}

type Handler struct {
	svc UserService
}

// NewHandler binds a UserService to HTTP handlers.
func NewHandler(svc UserService) *Handler {
	return &Handler{svc: svc}
}

// HealthCheck Health godoc
// @Summary      Health check
// @Description  Returns OK if service is up
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SignUp godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      http.SignUpRequest  true  "Signup payload"
// @Success      201      {object}  model.User
// @Failure      400      {object}  map[string]string
// @Router       /signup [post]
func (h *Handler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.SignUp(getContext(c), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login godoc
// @Summary      Authenticate user
// @Description  Log in a user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      http.LoginRequest  true  "Login payload"
// @Success      200      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.Login(getContext(c), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Profile godoc
// @Summary      Get user profile
// @Description  Fetch the profile data for the authenticated user
// @Tags         users
// @Produce      json
// @Success      200      {object}  model.User
// @Failure      401      {object}  map[string]string
// @Router       /profile [get]
// @Security     ApiKeyAuth
func (h *Handler) Profile(c *gin.Context) {
	idVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userID := idVal.(int64)
	user, err := h.svc.GetProfile(getContext(c), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user by ID (admin only)
// @Tags         users
// @Param        id       path      int  true  "User ID"
// @Success      200      "No Content"
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /user/{id} [delete]
// @Security     ApiKeyAuth
func (h *Handler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	// let parsing errors surface as 400
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	if err := h.svc.DeleteUser(getContext(c), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// UpdateProfile godoc
// @Summary      Update my profile
// @Description  Update the authenticated user's name
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        payload  body      http.UpdateProfileRequest  true  "New name"
// @Success      200      {object}  model.User
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /profile [put]
// @Security     ApiKeyAuth
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetInt64("userID")
	updated, err := h.svc.UpdateUser(getContext(c), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// getContext safely retrieves the request context.
func getContext(c *gin.Context) context.Context {
	if c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context()
	}
	return context.Background()
}
