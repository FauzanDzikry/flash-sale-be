package service

import (
	"flash-sale-be/internal/config"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	cfg := &config.Config{JWTKey: "test-secret", JWTExpireHour: 24}
	svc := NewAuthService(userRepo, cfg)

	userRepo.EXPECT().
		GetByEmail("test@example.com").
		Return(nil, gorm.ErrRecordNotFound)
	userRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(u *domain.User) error {
			assert.NotEmpty(t, u.ID)
			assert.Equal(t, "test@example.com", u.Email)
			assert.NotEmpty(t, u.Password)
			return nil
		})

	resp, err := svc.Register(&dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.Name)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	cfg := &config.Config{JWTKey: "test-secret"}
	svc := NewAuthService(userRepo, cfg)

	userRepo.EXPECT().
		GetByEmail("existing@example.com").
		Return(&domain.User{ID: uuid.New(), Email: "existing@example.com"}, nil)

	_, err := svc.Register(&dto.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test",
	})
	require.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestAuthService_Register_EmailRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewAuthService(userRepo, &config.Config{})

	_, err := svc.Register(&dto.RegisterRequest{
		Email:    "",
		Password: "password123",
		Name:     "Test",
	})
	require.Error(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	cfg := &config.Config{JWTKey: "test-secret", JWTExpireHour: 24}
	svc := NewAuthService(userRepo, cfg)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo.EXPECT().
		GetByEmail("login@example.com").
		Return(&domain.User{
			ID:       userID,
			Email:    "login@example.com",
			Password: string(hashedPassword),
			Name:     "Login User",
		}, nil)

	resp, err := svc.Login(&dto.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "login@example.com", resp.User.Email)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewAuthService(userRepo, &config.Config{})

	userRepo.EXPECT().
		GetByEmail("nonexistent@example.com").
		Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Login(&dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpass",
	})
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

