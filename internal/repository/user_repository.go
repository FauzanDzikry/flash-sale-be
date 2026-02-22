package repository

import (
	"flash-sale-be/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetById(id uuid.UUID) (*domain.User, error)
	Update(user *domain.User) error
	UpdatePassword(id uuid.UUID, password string) error
	Deactivate(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetById(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, password string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("password_hash", password).Error
}

func (r *userRepository) Deactivate(id uuid.UUID) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("deactivated_at", time.Now()).Error
}
