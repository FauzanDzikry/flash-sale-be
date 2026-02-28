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

func setupProductsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory", uuid.New().String())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
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
	return db
}

func TestProductsRepository_Create_GetById(t *testing.T) {
	db := setupProductsTestDB(t)
	repo := NewProductsRepository(db)

	userID := uuid.New()
	product := &domain.Product{
		ID:        uuid.New(),
		Name:      "Test Product",
		Category:  "Electronics",
		Stock:     10,
		Price:     99.99,
		Discount:  10,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(product)
	require.NoError(t, err)

	found, err := repo.GetById(product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.Name, found.Name)
	assert.Equal(t, 10, found.Stock)

	_, err = repo.GetById(uuid.New())
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductsRepository_DecrementStock(t *testing.T) {
	db := setupProductsTestDB(t)
	repo := NewProductsRepository(db)

	userID := uuid.New()
	product := &domain.Product{
		ID:        uuid.New(),
		Name:      "Stock Product",
		Category:  "Test",
		Stock:     10,
		Price:     50,
		Discount:  0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	err := db.Transaction(func(tx *gorm.DB) error {
		affected, err := repo.DecrementStock(tx, product.ID, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		return nil
	})
	require.NoError(t, err)

	var updated domain.Product
	require.NoError(t, db.First(&updated, "id = ?", product.ID).Error)
	assert.Equal(t, 7, updated.Stock)
}

func TestProductsRepository_DecrementStock_InsufficientStock(t *testing.T) {
	db := setupProductsTestDB(t)
	repo := NewProductsRepository(db)

	userID := uuid.New()
	product := &domain.Product{
		ID:        uuid.New(),
		Name:      "Low Stock",
		Category:  "Test",
		Stock:     2,
		Price:     10,
		Discount:  0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	err := db.Transaction(func(tx *gorm.DB) error {
		affected, err := repo.DecrementStock(tx, product.ID, 5)
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected)
		return nil
	})
	require.NoError(t, err)

	var updated domain.Product
	require.NoError(t, db.First(&updated, "id = ?", product.ID).Error)
	assert.Equal(t, 2, updated.Stock)
}
