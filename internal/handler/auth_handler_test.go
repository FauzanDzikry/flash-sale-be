package handler

import (
	"bytes"
	"encoding/json"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/mocks"
	"flash-sale-be/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupAuthRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func TestAuthHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := mocks.NewMockAuthService(ctrl)
	blacklist := mocks.NewMockTokenBlacklist(ctrl)
	h := NewAuthHandler(authSvc, blacklist)

	authSvc.EXPECT().
		Register(gomock.Any()).
		DoAndReturn(func(req *dto.RegisterRequest) (*dto.UserResponse, error) {
			assert.Equal(t, "test@example.com", req.Email)
			return &dto.UserResponse{
				ID:    "uuid-1",
				Email: "test@example.com",
				Name:  "Test User",
			}, nil
		})

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupAuthRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp dto.UserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test@example.com", resp.Email)
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := mocks.NewMockAuthService(ctrl)
	h := NewAuthHandler(authSvc, nil)

	authSvc.EXPECT().
		Register(gomock.Any()).
		Return(nil, service.ErrEmailAlreadyExists)

	body, _ := json.Marshal(map[string]string{
		"email":    "existing@example.com",
		"password": "password123",
		"name":     "Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupAuthRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := mocks.NewMockAuthService(ctrl)
	h := NewAuthHandler(authSvc, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupAuthRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := mocks.NewMockAuthService(ctrl)
	h := NewAuthHandler(authSvc, nil)

	authSvc.EXPECT().
		Login(gomock.Any()).
		Return(&dto.LoginResponse{
			AccessToken: "token-123",
			TokenType:   "Bearer",
			ExpiresIn:   86400,
			User:        dto.UserResponse{ID: "1", Email: "u@example.com", Name: "User"},
		}, nil)

	body, _ := json.Marshal(map[string]string{
		"email":    "u@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupAuthRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dto.LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "token-123", resp.AccessToken)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := mocks.NewMockAuthService(ctrl)
	h := NewAuthHandler(authSvc, nil)

	authSvc.EXPECT().
		Login(gomock.Any()).
		Return(nil, service.ErrInvalidCredentials)

	body, _ := json.Marshal(map[string]string{
		"email":    "wrong@example.com",
		"password": "wrongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupAuthRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
