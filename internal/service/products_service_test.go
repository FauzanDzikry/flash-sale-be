package service

import (
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/mocks"
	"flash-sale-be/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestProductsService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productsRepo.EXPECT().
		GetByName("New Product", uuid.Nil).
		Return(nil, repository.ErrProductNotFound)
	productsRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(p *domain.Product) error {
			assert.Equal(t, "New Product", p.Name)
			assert.Equal(t, 10, p.Stock)
			assert.Equal(t, 99.99, p.Price)
			return nil
		})

	resp, err := svc.Create(&dto.CreateProductRequest{
		Name:      "New Product",
		Category:  "Electronics",
		Stock:     10,
		Price:     99.99,
		Discount:  0,
		CreatedBy: uuid.New().String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "New Product", resp.Name)
	assert.Equal(t, 10, resp.Stock)
}

func TestProductsService_Create_ProductAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productsRepo.EXPECT().
		GetByName("Existing Product", uuid.Nil).
		Return(&domain.Product{ID: uuid.New(), Name: "Existing Product"}, nil)

	_, err := svc.Create(&dto.CreateProductRequest{
		Name:      "Existing Product",
		Category:  "Test",
		Stock:     5,
		Price:     10,
		CreatedBy: uuid.New().String(),
	})
	require.ErrorIs(t, err, ErrProductAlreadyExists)
}

func TestProductsService_Create_InvalidStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productsRepo.EXPECT().
		GetByName("Product", uuid.Nil).
		Return(nil, repository.ErrProductNotFound)

	_, err := svc.Create(&dto.CreateProductRequest{
		Name:      "Product",
		Category:  "Test",
		Stock:     -1,
		Price:     10,
		Discount:  0,
		CreatedBy: uuid.New().String(),
	})
	require.ErrorIs(t, err, ErrProductStockInvalid)
}

func TestProductsService_Create_InvalidDiscount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productsRepo.EXPECT().
		GetByName("Product", uuid.Nil).
		Return(nil, repository.ErrProductNotFound)

	_, err := svc.Create(&dto.CreateProductRequest{
		Name:      "Product",
		Category:  "Test",
		Stock:     5,
		Price:     10,
		Discount:  150,
		CreatedBy: uuid.New().String(),
	})
	require.ErrorIs(t, err, ErrProductDiscountInvalid)
}

func TestProductsService_GetById_AccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	productsRepo.EXPECT().
		GetById(productID).
		Return(&domain.Product{
			ID:        productID,
			Name:      "Product",
			CreatedBy: ownerID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

	_, err := svc.GetById(productID.String(), otherUserID.String())
	require.ErrorIs(t, err, ErrProductAccessDenied)
}

func TestProductsService_GetById_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewProductsService(productsRepo)

	productID := uuid.New()
	userID := uuid.New()

	productsRepo.EXPECT().
		GetById(productID).
		Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.GetById(productID.String(), userID.String())
	require.Error(t, err)
}
