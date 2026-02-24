package service

import (
	"errors"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductAccessDenied    = errors.New("you do not have access to this product")
	ErrProductAlreadyExists   = errors.New("product already exists")
	ErrProductStockInvalid    = errors.New("product stock is invalid")
	ErrProductPriceInvalid    = errors.New("product price is invalid")
	ErrProductDiscountInvalid = errors.New("product discount is invalid")
)

type ProductsService interface {
	Create(req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	Update(id string, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	GetById(id string, createdBy string) (*dto.ProductResponse, error)
	GetAll(createdBy string) ([]*dto.ProductResponse, error)
	Delete(id string, createdBy string) error
}

type productsService struct {
	productsRepo repository.ProductsRepository
}

func NewProductsService(productsRepo repository.ProductsRepository) ProductsService {
	return &productsService{productsRepo: productsRepo}
}

func (s *productsService) Create(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created by: %w", err)
	}
	req.Name = strings.TrimSpace(req.Name)
	existing, err := s.productsRepo.GetByName(req.Name, uuid.Nil)
	if err != nil && !errors.Is(err, repository.ErrProductNotFound) {
		return nil, fmt.Errorf("checking product name: %w", err)
	}
	if existing != nil {
		return nil, ErrProductAlreadyExists
	}
	if req.Stock < 0 {
		return nil, ErrProductStockInvalid
	}
	if req.Price < 0 {
		return nil, ErrProductPriceInvalid
	}
	if req.Discount < 0 || req.Discount > 100 {
		return nil, ErrProductDiscountInvalid
	}
	product := &domain.Product{
		ID:        uuid.New(),
		Name:      req.Name,
		Category:  req.Category,
		Stock:     req.Stock,
		Price:     req.Price,
		Discount:  req.Discount,
		CreatedBy: createdBy,
	}
	if err := s.productsRepo.Create(product); err != nil {
		return nil, fmt.Errorf("creating product: %w", err)
	}
	return &dto.ProductResponse{
		ID:        product.ID.String(),
		Name:      product.Name,
		Category:  product.Category,
		Stock:     product.Stock,
		Price:     product.Price,
		Discount:  product.Discount,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		DeletedAt: product.DeletedAt,
		CreatedBy: product.CreatedBy.String(),
	}, nil
}

func (s *productsService) Update(id string, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	product, err := s.productsRepo.GetById(productID)
	if err != nil {
		return nil, fmt.Errorf("getting product: %w", err)
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	product.Name = strings.TrimSpace(req.Name)
	existing, err := s.productsRepo.GetByName(product.Name, product.ID)
	if err != nil && !errors.Is(err, repository.ErrProductNotFound) {
		return nil, fmt.Errorf("checking product name: %w", err)
	}
	if existing != nil && existing.ID != product.ID {
		return nil, ErrProductAlreadyExists
	}
	if req.Stock < 0 {
		return nil, ErrProductStockInvalid
	}
	if req.Price < 0 {
		return nil, ErrProductPriceInvalid
	}
	if req.Discount < 0 || req.Discount > 100 {
		return nil, ErrProductDiscountInvalid
	}
	product.Name = req.Name
	product.Category = req.Category
	product.Stock = req.Stock
	product.Price = req.Price
	product.Discount = req.Discount
	product.UpdatedAt = time.Now()
	if err := s.productsRepo.Update(product); err != nil {
		return nil, fmt.Errorf("updating product: %w", err)
	}
	return &dto.ProductResponse{
		ID:        product.ID.String(),
		Name:      product.Name,
		Category:  product.Category,
		Stock:     product.Stock,
		Price:     product.Price,
		Discount:  product.Discount,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		DeletedAt: product.DeletedAt,
		CreatedBy: product.CreatedBy.String(),
	}, nil
}

func (s *productsService) GetById(id string, createdBy string) (*dto.ProductResponse, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	createdByUUID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created by: %w", err)
	}
	product, err := s.productsRepo.GetById(productID)
	if err != nil {
		return nil, fmt.Errorf("getting product: %w", err)
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	if product.CreatedBy != createdByUUID {
		return nil, ErrProductAccessDenied
	}
	return &dto.ProductResponse{
		ID:        product.ID.String(),
		Name:      product.Name,
		Category:  product.Category,
		Stock:     product.Stock,
		Price:     product.Price,
		Discount:  product.Discount,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		DeletedAt: product.DeletedAt,
		CreatedBy: product.CreatedBy.String(),
	}, nil
}

func (s *productsService) GetAll(createdBy string) ([]*dto.ProductResponse, error) {
	createdByUUID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created by: %w", err)
	}
	products, err := s.productsRepo.GetAll(createdByUUID)
	if err != nil {
		return nil, fmt.Errorf("getting products: %w", err)
	}
	productResponses := make([]*dto.ProductResponse, len(products))
	for _, product := range products {
		productResponses = append(productResponses, &dto.ProductResponse{
			ID:        product.ID.String(),
			Name:      product.Name,
			Category:  product.Category,
			Stock:     product.Stock,
			Price:     product.Price,
			Discount:  product.Discount,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
			DeletedAt: product.DeletedAt,
			CreatedBy: product.CreatedBy.String(),
		})
	}
	return productResponses, nil
}

func (s *productsService) Delete(id string, createdBy string) error {
	productID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}
	createdByUUID, err := uuid.Parse(createdBy)
	if err != nil {
		return fmt.Errorf("invalid created by: %w", err)
	}
	product, err := s.productsRepo.GetById(productID)
	if err != nil {
		return fmt.Errorf("getting product: %w", err)
	}
	if product == nil {
		return ErrProductNotFound
	}
	if product.CreatedBy != createdByUUID {
		return ErrProductAccessDenied
	}
	if err := s.productsRepo.Delete(productID); err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}
	return nil
}
