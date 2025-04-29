package http

import (
	"github.com/enson89/user-service-go/internal/auth"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter sets up routes and middleware
func NewRouter(svc UserService, jwtSecret []byte, sessionStore auth.SessionStore) *gin.Engine {
	h := NewHandler(svc)
	r := gin.Default()

	v1 := r.Group("/v1")
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public
	v1.GET("/health", h.HealthCheck)
	v1.POST("/signup", h.SignUp)
	v1.POST("/login", h.Login)

	// Protected
	authGroup := v1.Group("/")
	authGroup.Use(auth.AuthenticationMiddleware(jwtSecret, sessionStore))
	{
		authGroup.GET("/profile", h.Profile)
		authGroup.PUT("/profile", h.UpdateProfile)

		// Admin-only
		admin := authGroup.Group("/")
		admin.Use(auth.RequireRole("admin"))
		admin.DELETE("/user/:id", h.DeleteUser)
	}
	return r
}
