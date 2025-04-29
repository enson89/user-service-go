package service

import (
	"context"
	"errors"
	"time"

	"github.com/enson89/user-service-go/internal/auth"
	"github.com/enson89/user-service-go/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, u *model.User) error
}

type SessionStore interface {
	BlacklistToken(ctx context.Context, token string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

type UserService struct {
	repo      UserRepository
	Store     SessionStore  // exported for middleware
	Secret    []byte        // exported for middleware
	jwtExpire time.Duration // used internally for token expiry
}

func NewUserService(repo UserRepository, store SessionStore, secret []byte, expire time.Duration) *UserService {
	return &UserService{repo: repo, Store: store, Secret: secret, jwtExpire: expire}
}

func (s *UserService) SignUp(ctx context.Context, email, password string) (*model.User, error) {
	if existing, _ := s.repo.GetByEmail(ctx, email); existing != nil {
		return nil, errors.New("email already in use")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Email: email, PasswordHash: string(hash), Role: "user"}
	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil || u == nil {
		return "", errors.New("invalid credentials")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	token, err := auth.GenerateToken(u, s.Secret, s.jwtExpire)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) GetProfile(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, newName string) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil || u == nil {
		return nil, errors.New("user not found")
	}
	u.Name = newName
	if err = s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
