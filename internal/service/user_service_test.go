package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	authMocks "github.com/enson89/user-service-go/internal/auth/mocks"
	"github.com/enson89/user-service-go/internal/model"
	"github.com/enson89/user-service-go/internal/service"
	repoMocks "github.com/enson89/user-service-go/internal/service/mocks"
)

func TestSignUp_Success(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	mr.On("GetByEmail", mock.Anything, "user@x.com").Return(nil, nil)
	mr.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	u, err := svc.SignUp(t.Context(), "user@x.com", "pwd1234")
	assert.NoError(t, err)
	assert.Equal(t, "user@x.com", u.Email)
	assert.Equal(t, "user", u.Role)
	mr.AssertExpectations(t)
}

func TestSignUp_Duplicate(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	mr.On("GetByEmail", mock.Anything, "user@x.com").
		Return(&model.User{Email: "user@x.com"}, nil)

	u, err := svc.SignUp(t.Context(), "user@x.com", "pwd1234")
	assert.Error(t, err)
	assert.Nil(t, u)
	mr.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	mr.On("GetByEmail", mock.Anything, "user@x.com").
		Return(&model.User{ID: 7, Email: "user@x.com", PasswordHash: string(hash), Role: "user"}, nil)

	token, err := svc.Login(t.Context(), "user@x.com", "correct")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mr.AssertExpectations(t)
}

func TestLogin_Invalid(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	mr.On("GetByEmail", mock.Anything, "user@x.com").
		Return(nil, errors.New("not found"))

	token, err := svc.Login(t.Context(), "user@x.com", "pwd")
	assert.Error(t, err)
	assert.Empty(t, token)
	mr.AssertExpectations(t)
}

func TestGetProfile(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	expected := &model.User{ID: 3, Email: "a@b.com", Role: "admin"}
	mr.On("GetByID", mock.Anything, int64(3)).Return(expected, nil)

	u, err := svc.GetProfile(t.Context(), 3)
	assert.NoError(t, err)
	assert.Equal(t, expected, u)
	mr.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("sec"), time.Hour)

	mr.On("Delete", mock.Anything, int64(5)).Return(nil)

	err := svc.DeleteUser(t.Context(), 5)
	assert.NoError(t, err)
	mr.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("secret"), time.Hour)

	existing := &model.User{ID: 1, Name: "Old"}
	mr.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
	mr.On("Update", mock.Anything, existing).Return(nil)

	u, err := svc.UpdateUser(t.Context(), 1, "New")
	assert.NoError(t, err)
	assert.Equal(t, "New", u.Name)

	mr.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	mr := new(repoMocks.MockUserRepository)
	ms := new(authMocks.MockSessionStore)
	svc := service.NewUserService(mr, ms, []byte("secret"), time.Hour)

	mr.On("GetByID", mock.Anything, int64(2)).Return(nil, errors.New("not found"))

	u, err := svc.UpdateUser(t.Context(), 2, "New")
	assert.Error(t, err)
	assert.Nil(t, u)

	mr.AssertExpectations(t)
}
