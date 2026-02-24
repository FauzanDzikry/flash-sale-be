package repository

import (
	"errors"
	"flash-sale-be/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductsRepository interface {
	Create(product *domain.Product) error
	Update(product *domain.Product) error
	GetById(id uuid.UUID) (*domain.Product, error)
	GetByName(name string, id uuid.UUID) (*domain.Product, error)
	GetAll(createdBy uuid.UUID) ([]*domain.Product, error)
	Delete(id uuid.UUID) error
}

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return &productsRepository{db: db}
}

func (r *productsRepository) Create(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *productsRepository) Update(product *domain.Product) error {
	return r.db.Save(product).Error
}

func (r *productsRepository) GetById(id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Where("deleted_at IS NULL").Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productsRepository) GetByName(name string, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	db := r.db.Where("deleted_at IS NULL").Where("name = ?", name)
	if id != uuid.Nil {
		db = db.Where("id != ?", id)
	}
	if err := db.First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productsRepository) GetAll(createdBy uuid.UUID) ([]*domain.Product, error) {
	var products []*domain.Product
	if err := r.db.Where("deleted_at IS NULL").Where("created_by = ?", createdBy).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productsRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Product{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
