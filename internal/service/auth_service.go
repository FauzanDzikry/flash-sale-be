package service

import (
	"errors"
	"flash-sale-be/internal/config"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req *dto.LoginRequest) (string, error)
	ForgotPassword(req *dto.ForgotPasswordRequest) error
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, config *config.Config) AuthService {
	return &authService{userRepo: userRepo, config: config}
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.UserResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, errors.New("email is required")
	}

	existing, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("Checking email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		if isDuplicateError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return &dto.UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (s *authService) Login(req *dto.LoginRequest) (string, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return "", errors.New("email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return "", errors.New("password is required")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("finding user: %w", err)
	}
	if user.DeactivatedAt != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateJwt(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("generating JWT: %w", err)
	}
	return token, nil
}

func (s *authService) ForgotPassword(req *dto.ForgotPasswordRequest) error {
	return nil
}

func isDuplicateError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

func (s *authService) generateJwt(userID uuid.UUID, email string) (string, error) {
	exp := time.Now().Add(time.Duration(s.config.JWTExpireHour) * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     exp.Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTKey))
}
