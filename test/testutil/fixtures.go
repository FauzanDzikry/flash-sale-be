package testutil

import (
	"flash-sale-be/internal/domain"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB, email, password, name string) (uuid.UUID, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashed),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(user).Error; err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func SeedProduct(db *gorm.DB, createdBy uuid.UUID, stock int, price, discount float64) (uuid.UUID, error) {
	product := &domain.Product{
		ID:        uuid.New(),
		Name:      "Test Product " + uuid.New().String()[:8],
		Category:  "Test",
		Stock:     stock,
		Price:     price,
		Discount:  discount,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(product).Error; err != nil {
		return uuid.Nil, err
	}
	return product.ID, nil
}

func CleanTables(db *gorm.DB) error {
	tables := []string{"checkouts", "products", "otps", "users"}
	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			return err
		}
	}
	return nil
}
