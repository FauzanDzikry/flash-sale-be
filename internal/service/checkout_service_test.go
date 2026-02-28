package service

import (
	"context"
	"errors"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/mocks"
	"flash-sale-be/internal/queue"
	"flash-sale-be/internal/repository"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func setupCheckoutServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory", uuid.New().String())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, password_hash TEXT, name TEXT, created_at DATETIME, updated_at DATETIME, deactivated_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, category TEXT, stock INTEGER, price REAL, discount REAL, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, created_by TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE checkouts (id TEXT PRIMARY KEY, user_id TEXT, product_id TEXT, quantity INTEGER, price REAL, discount REAL, total_price REAL, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error)
	return db
}

func TestCheckoutService_EnqueueCheckout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	queueMock := mocks.NewMockQueue(ctrl)
	svc := NewCheckoutService(nil, productsRepo, queueMock, nil)

	userID := uuid.New().String()
	productID := uuid.New().String()

	productsRepo.EXPECT().
		GetById(gomock.Any()).
		DoAndReturn(func(id uuid.UUID) (*domain.Product, error) {
			assert.Equal(t, productID, id.String())
			return &domain.Product{ID: id, Stock: 10}, nil
		})
	queueMock.EXPECT().
		EnqueueCheckout(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, job queue.CheckoutJob) error {
			assert.Equal(t, userID, job.UserID)
			assert.Equal(t, productID, job.ProductID)
			assert.Equal(t, 2, job.Quantity)
			return nil
		})

	jobID, err := svc.EnqueueCheckout(context.Background(), userID, &dto.CheckoutRequest{
		ProductID: productID,
		Quantity:  2,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, jobID)
}

func TestCheckoutService_EnqueueCheckout_ProductNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewCheckoutService(nil, productsRepo, nil, nil)

	productsRepo.EXPECT().
		GetById(gomock.Any()).
		Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.EnqueueCheckout(context.Background(), uuid.New().String(), &dto.CheckoutRequest{
		ProductID: uuid.New().String(),
		Quantity:  1,
	})
	require.ErrorIs(t, err, ErrCheckoutNotFound)
}

func TestCheckoutService_EnqueueCheckout_InvalidQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsRepo := mocks.NewMockProductsRepository(ctrl)
	svc := NewCheckoutService(nil, productsRepo, nil, nil)

	_, err := svc.EnqueueCheckout(context.Background(), uuid.New().String(), &dto.CheckoutRequest{
		ProductID: uuid.New().String(),
		Quantity:  0,
	})
	require.Error(t, err)
}

func TestCheckoutService_ProcessCheckoutJob_Success(t *testing.T) {
	db := setupCheckoutServiceTestDB(t)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	product := &domain.Product{
		ID:        productID,
		Name:      "Test",
		Category:  "Test",
		Stock:     10,
		Price:     100,
		Discount:  10,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, productsRepo.Create(product))

	svc := NewCheckoutService(checkoutRepo, productsRepo, nil, db)

	resp, err := svc.ProcessCheckoutJob(context.Background(), &queue.CheckoutJob{
		JobID:     uuid.New().String(),
		UserID:    userID.String(),
		ProductID: productID.String(),
		Quantity:  3,
	})
	require.NoError(t, err)
	assert.Equal(t, productID.String(), resp.ProductID)
	assert.Equal(t, 3, resp.Quantity)
	assert.Equal(t, float64(270), resp.TotalPrice)

	var updated domain.Product
	require.NoError(t, db.First(&updated, "id = ?", productID).Error)
	assert.Equal(t, 7, updated.Stock)
}

func TestCheckoutService_ProcessCheckoutJob_InsufficientStock(t *testing.T) {
	db := setupCheckoutServiceTestDB(t)
	productsRepo := repository.NewProductsRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	product := &domain.Product{
		ID:        productID,
		Name:      "Low Stock",
		Category:  "Test",
		Stock:     2,
		Price:     10,
		Discount:  0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, productsRepo.Create(product))

	svc := NewCheckoutService(repository.NewCheckoutRepository(db), productsRepo, nil, db)

	_, err := svc.ProcessCheckoutJob(context.Background(), &queue.CheckoutJob{
		JobID:     uuid.New().String(),
		UserID:    userID.String(),
		ProductID: productID.String(),
		Quantity:  5,
	})
	require.ErrorIs(t, err, ErrCheckoutInsufficientStock)

	var updated domain.Product
	require.NoError(t, db.First(&updated, "id = ?", productID).Error)
	assert.Equal(t, 2, updated.Stock)
}

func TestCheckoutService_ProcessCheckoutJob_RaceCondition(t *testing.T) {
	db := setupCheckoutServiceTestDB(t)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	product := &domain.Product{
		ID:        productID,
		Name:      "Flash Sale",
		Category:  "Test",
		Stock:     10,
		Price:     100,
		Discount:  0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, productsRepo.Create(product))

	svc := NewCheckoutService(checkoutRepo, productsRepo, nil, db)

	concurrentWorkers := 20
	quantityPerJob := 3
	results := make(chan error, concurrentWorkers)

	for i := 0; i < concurrentWorkers; i++ {
		go func() {
			_, err := svc.ProcessCheckoutJob(context.Background(), &queue.CheckoutJob{
				JobID:     uuid.New().String(),
				UserID:    userID.String(),
				ProductID: productID.String(),
				Quantity:  quantityPerJob,
			})
			results <- err
		}()
	}

	successCount := 0
	for i := 0; i < concurrentWorkers; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			assert.True(t, errors.Is(err, ErrCheckoutInsufficientStock) || errors.Is(err, ErrCheckoutProductNotFound),
				"expected insufficient stock or product not found, got: %v", err)
		}
	}

	assert.LessOrEqual(t, successCount, 3, "max 3 can succeed (10 stock / 3 qty)")
	assert.GreaterOrEqual(t, successCount, 1)

	var finalProduct domain.Product
	require.NoError(t, db.First(&finalProduct, "id = ?", productID).Error)
	assert.GreaterOrEqual(t, finalProduct.Stock, 0)

	var totalCheckouts int64
	db.Model(&domain.Checkout{}).Where("product_id = ?", productID).Count(&totalCheckouts)
	assert.LessOrEqual(t, int(totalCheckouts*int64(quantityPerJob)), 10)
}

func TestCheckoutService_ProcessCheckoutJob_RaceCondition_Parallel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race test in short mode")
	}

	db := setupCheckoutServiceTestDB(t)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)

	userID := uuid.New()
	productID := uuid.New()
	product := &domain.Product{
		ID:        productID,
		Name:      "Race Product",
		Category:  "Test",
		Stock:     5,
		Price:     50,
		Discount:  0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, productsRepo.Create(product))

	svc := NewCheckoutService(checkoutRepo, productsRepo, nil, db)

	var wg sync.WaitGroup
	mu := sync.Mutex{}
	successCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.ProcessCheckoutJob(context.Background(), &queue.CheckoutJob{
				JobID:     uuid.New().String(),
				UserID:    userID.String(),
				ProductID: productID.String(),
				Quantity:  1,
			})
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 5, successCount)
}
