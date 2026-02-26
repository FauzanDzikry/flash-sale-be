package service

import (
	"context"
	"errors"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/queue"
	"flash-sale-be/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCheckoutNotFound          = errors.New("checkout not found")
	ErrCheckoutProductNotFound    = errors.New("product not found")
	ErrCheckoutInsufficientStock  = errors.New("insufficient stock")
)

type CheckoutService interface {
	EnqueueCheckout(ctx context.Context, userID string, req *dto.CheckoutRequest) (jobID string, err error)
	ProcessCheckoutJob(ctx context.Context, job *queue.CheckoutJob) (*dto.CheckoutResponse, error)
}

type checkoutService struct {
	checkoutRepo 	repository.CheckoutRepository
	productsRepo 	repository.ProductsRepository
	queue 			queue.Queue
	productService 	ProductsService
	db 				*gorm.DB
}

func NewCheckoutService(
	checkoutRepo repository.CheckoutRepository,
	productsRepo repository.ProductsRepository,
	q queue.Queue,
	db *gorm.DB,
) CheckoutService {
	return &checkoutService{
		checkoutRepo: checkoutRepo,
		productsRepo: productsRepo,
		queue:        q,
		db:           db,
	}
}

func (s *checkoutService) EnqueueCheckout(ctx context.Context, userID string, req *dto.CheckoutRequest) (jobID string, err error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	productUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return "", fmt.Errorf("invalid product id: %w", err)
	}
	if req.Quantity <= 0 {
		return "", fmt.Errorf("quantity must be greater than 0")
	}
	_, err = s.productsRepo.GetById(productUUID)
	if err != nil {
		return "", ErrCheckoutNotFound
	}
	job := queue.CheckoutJob{
		JobID:     uuid.New().String(),
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	if err := s.queue.EnqueueCheckout(ctx, job); err != nil {
		return "", fmt.Errorf("enqueueing checkout job: %w", err)
	}
	return job.JobID, nil
}

func (s *checkoutService) ProcessCheckoutJob(ctx context.Context, job *queue.CheckoutJob) (*dto.CheckoutResponse, error) {
	userID, err := uuid.Parse(job.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid job user_id: %w", err)
	}
	productID, err := uuid.Parse(job.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid job product_id: %w", err)
	}

	product, err := s.productsRepo.GetById(productID)
	if err != nil || product == nil {
		return nil, ErrCheckoutProductNotFound
	}
	if product.Stock < job.Quantity {
		return nil, ErrCheckoutInsufficientStock
	}

	var checkout *domain.Checkout
	err = s.db.Transaction(func(tx *gorm.DB) error {
		affected, err := s.productsRepo.DecrementStock(tx, productID, job.Quantity)
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrCheckoutInsufficientStock
		}
		subTotal := product.Price * float64(job.Quantity)
		discountAmount := subTotal * (product.Discount / 100)
		totalPrice := subTotal - discountAmount
		checkout = &domain.Checkout{
			UserID:     userID,
			ProductID:  productID,
			Quantity:   job.Quantity,
			Price:      product.Price,
			Discount:   product.Discount,
			TotalPrice: totalPrice,
		}
		return s.checkoutRepo.CreateWithTx(tx, checkout)
	})
	if err != nil {
		return nil, err
	}
	return &dto.CheckoutResponse{
		ID:         checkout.ID.String(),
		ProductID:  checkout.ProductID.String(),
		Quantity:   checkout.Quantity,
		Price:      checkout.Price,
		Discount:   checkout.Discount,
		TotalPrice: checkout.TotalPrice,
		CreatedAt:  checkout.CreatedAt,
		UpdatedAt:  checkout.UpdatedAt,
		DeletedAt:  checkout.DeletedAt,
	}, nil
}

