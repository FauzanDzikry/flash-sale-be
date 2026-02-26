package repository

import (
	"errors"
	"flash-sale-be/internal/domain"

	"gorm.io/gorm"
)

var (
	ErrCheckoutNotFound = errors.New("checkout not found")
)

type CheckoutRepository interface {
	Create(checkout *domain.Checkout) error
	CreateWithTransaction(tx *gorm.DB, checkout *domain.Checkout) error
	CreateWithTx(tx *gorm.DB, checkout *domain.Checkout) error
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
