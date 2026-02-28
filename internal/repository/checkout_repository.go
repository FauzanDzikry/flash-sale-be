package repository

import (
	"errors"
	"flash-sale-be/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCheckoutNotFound = errors.New("checkout not found")
)

type CheckoutRepository interface {
	Create(checkout *domain.Checkout) error
	CreateWithTransaction(tx *gorm.DB, checkout *domain.Checkout) error
	CreateWithTx(tx *gorm.DB, checkout *domain.Checkout) error
	GetAllByUserID(userID uuid.UUID) ([]*domain.Checkout, error)
}

type checkoutRepository struct {
	db *gorm.DB
}

func NewCheckoutRepository(db *gorm.DB) CheckoutRepository {
	return &checkoutRepository{db: db}
}

func (r *checkoutRepository) Create(checkout *domain.Checkout) error {
	return r.db.Create(checkout).Error
}

func (r *checkoutRepository) CreateWithTransaction(tx *gorm.DB, checkout *domain.Checkout) error {
	return tx.Create(checkout).Error
}

func (r *checkoutRepository) CreateWithTx(tx *gorm.DB, checkout *domain.Checkout) error {
	return tx.Create(checkout).Error
}

func (r *checkoutRepository) GetAllByUserID(userID uuid.UUID) ([]*domain.Checkout, error) {
	var list []domain.Checkout
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Checkout, 0, len(list))
	for i := range list {
		out = append(out, &list[i])
	}
	return out, nil
}
