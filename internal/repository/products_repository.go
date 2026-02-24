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
	GetAllNotDeleted() ([]*domain.Product, error)
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
	var list []domain.Product
	if err := r.db.Where("deleted_at IS NULL").Where("created_by = ?", createdBy).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Product, 0, len(list))
	for i := range list {
		p := new(domain.Product)
		*p = list[i]
		out = append(out, p)
	}
	return out, nil
}

// GetAllNotDeleted returns all products from all users; excludes soft-deleted (deleted_at IS NULL).
func (r *productsRepository) GetAllNotDeleted() ([]*domain.Product, error) {
	var list []domain.Product
	if err := r.db.Where("deleted_at IS NULL").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Product, 0, len(list))
	for i := range list {
		p := new(domain.Product)
		*p = list[i]
		out = append(out, p)
	}
	return out, nil
}

func (r *productsRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Product{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
