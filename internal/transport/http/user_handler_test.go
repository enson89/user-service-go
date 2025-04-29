package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/enson89/user-service-go/internal/model"
	httptransport "github.com/enson89/user-service-go/internal/transport/http"
	httphandlermocks "github.com/enson89/user-service-go/internal/transport/http/mocks"
)

func setupRouter(mockSvc *httphandlermocks.MockUserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return httptransport.NewRouter(mockSvc, []byte("test-secret"), nil)
}

func TestHandler_HealthCheck(t *testing.T) {
	router := setupRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestHandler_SignUp(t *testing.T) {
	mockSvc := new(httphandlermocks.MockUserService)
	router := setupRouter(mockSvc)

	// prepare request body
	body := map[string]string{"email": "new@x.com", "password": "pw1234"}
	buf, _ := json.Marshal(body)

	// expect mock call
	mockSvc.
		On("SignUp", mock.Anything, "new@x.com", "pw1234").
		Return(&model.User{ID: 10, Email: "new@x.com", Role: "user"}, nil)

	// perform request
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10), resp["id"])
	assert.Equal(t, "new@x.com", resp["email"])

	mockSvc.AssertExpectations(t)
}

func TestHandler_Login(t *testing.T) {
	mockSvc := new(httphandlermocks.MockUserService)
	router := setupRouter(mockSvc)

	body := map[string]string{"email": "ok@x.com", "password": "pw"}
	buf, _ := json.Marshal(body)

	mockSvc.
		On("Login", mock.Anything, "ok@x.com", "pw").
		Return("token123", nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "token123", resp["token"])

	mockSvc.AssertExpectations(t)
}

func TestHandler_Profile(t *testing.T) {
	mockSvc := new(httphandlermocks.MockUserService)
	handler := httptransport.NewHandler(mockSvc)

	mockSvc.
		On("GetProfile", mock.Anything, int64(10)).
		Return(&model.User{ID: 10, Email: "u@x.com", Role: "user"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", int64(10))

	handler.Profile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10), resp["id"])
	assert.Equal(t, "u@x.com", resp["email"])

	mockSvc.AssertExpectations(t)
}

func TestHandler_DeleteUser(t *testing.T) {
	mockSvc := new(httphandlermocks.MockUserService)
	handler := httptransport.NewHandler(mockSvc)

	mockSvc.
		On("DeleteUser", mock.Anything, int64(10)).
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(10, 10)}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_UpdateProfile(t *testing.T) {
	// Create mock service
	mockSvc := new(httphandlermocks.MockUserService)
	handler := httptransport.NewHandler(mockSvc)

	// Prepare request
	body := map[string]string{"name": "Alice"}
	buf, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(buf))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", int64(1))

	// Expect service.UpdateUser
	updated := &model.User{ID: 1, Email: "x@x.com", Name: "Alice", Role: "user"}
	mockSvc.
		On("UpdateUser", mock.Anything, int64(1), "Alice").
		Return(updated, nil)

	// Call handler
	handler.UpdateProfile(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.User
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, updated, &resp)

	mockSvc.AssertExpectations(t)
}
