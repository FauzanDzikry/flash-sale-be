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

func setupTestDB(t *testing.T) *gorm.DB {
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
	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashed",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(user)
	require.NoError(t, err)

	var found domain.User
	err = db.First(&found, "id = ?", user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "getbyemail@example.com",
		Password:  "hashed",
		Name:      "Get By Email",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(user).Error)

	found, err := repo.GetByEmail("getbyemail@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)

	_, err = repo.GetByEmail("nonexistent@example.com")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_GetById(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "getbyid@example.com",
		Password:  "hashed",
		Name:      "Get By ID",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(user).Error)

	found, err := repo.GetById(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)

	_, err = repo.GetById(uuid.New())
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
