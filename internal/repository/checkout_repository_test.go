package repository

import (
	"flash-sale-be/internal/domain"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCheckoutTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory", uuid.New().String())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deactivated_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		category TEXT NOT NULL,
		stock INTEGER NOT NULL,
		price REAL NOT NULL,
		discount REAL NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME,
		created_by TEXT NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE checkouts (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		product_id TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		price REAL NOT NULL,
		discount REAL NOT NULL,
		total_price REAL NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	return db
}

func TestCheckoutRepository_Create(t *testing.T) {
	db := setupCheckoutTestDB(t)
	repo := NewCheckoutRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	checkout := &domain.Checkout{
		ID:         uuid.New(),
		UserID:     userID,
		ProductID:  productID,
		Quantity:   2,
		Price:      100,
		Discount:   10,
		TotalPrice: 180,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(checkout)
	require.NoError(t, err)

	var found domain.Checkout
	require.NoError(t, db.First(&found, "id = ?", checkout.ID).Error)
	assert.Equal(t, checkout.UserID, found.UserID)
	assert.Equal(t, checkout.Quantity, found.Quantity)
}

func TestCheckoutRepository_CreateWithTx(t *testing.T) {
	db := setupCheckoutTestDB(t)
	repo := NewCheckoutRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	checkout := &domain.Checkout{
		ID:         uuid.New(),
		UserID:     userID,
		ProductID:  productID,
		Quantity:  1,
		Price:     50,
		Discount:  0,
		TotalPrice: 50,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return repo.CreateWithTx(tx, checkout)
	})
	require.NoError(t, err)

	var found domain.Checkout
	require.NoError(t, db.First(&found, "id = ?", checkout.ID).Error)
	assert.Equal(t, checkout.TotalPrice, found.TotalPrice)
}
