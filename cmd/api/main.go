package main

import (
	"github.com/redis/go-redis/v9"
	"log"
	"user-service/internal/cache"
	"user-service/internal/config"
	"user-service/internal/db"
	"user-service/internal/repository"
	"user-service/internal/service"
	"user-service/internal/transport/http"

	_ "user-service/docs"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Initialize Postgres connection
	pgConn, err := db.NewPostgres(db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	repo := repository.NewUserRepository(pgConn)

	// 3. Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	store := cache.NewSessionStore(rdb, cfg.JWT.ExpireHours)

	// 4. Create the service layer
	svc := service.NewUserService(repo, store, []byte(cfg.JWT.Secret), cfg.JWT.ExpireHours)

	// 5. Wire up HTTP transport and start server
	router := http.NewRouter(svc, []byte(cfg.JWT.Secret), store)
	log.Printf("starting server on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
