//go:build integration

package integration

import (
	"context"
	"errors"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/queue"
	"flash-sale-be/internal/repository"
	"flash-sale-be/internal/service"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flash-sale-be/test/testutil"
)

func TestCheckout_RaceCondition_ConcurrentProcessCheckoutJob(t *testing.T) {
	db, cleanupDB := testutil.SetupTestDB(t)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupTestRedis(t)
	defer cleanupRedis()

	q := queue.NewRedisQueue(rdb)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)
	checkoutSvc := service.NewCheckoutService(checkoutRepo, productsRepo, q, db)

	userID, err := testutil.SeedUser(db, "race@example.com", "pass123", "Race User")
	require.NoError(t, err)

	productID, err := testutil.SeedProduct(db, userID, 10, 100, 0)
	require.NoError(t, err)

	for i := 0; i < 20; i++ {
		err := q.EnqueueCheckout(context.Background(), queue.CheckoutJob{
			JobID:     uuid.New().String(),
			UserID:    userID.String(),
			ProductID: productID.String(),
			Quantity:  3,
		})
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emptyCount := 0
			for emptyCount < 5 {
				job, err := q.DequeueCheckout(context.Background())
				if err != nil {
					if errors.Is(err, queue.ErrEmptyQueue) {
						emptyCount++
						time.Sleep(100 * time.Millisecond)
						continue
					}
					return
				}
				emptyCount = 0
				_, _ = checkoutSvc.ProcessCheckoutJob(context.Background(), job)
			}
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	var product domain.Product
	require.NoError(t, db.Where("id = ?", productID).First(&product).Error)
	assert.GreaterOrEqual(t, product.Stock, 0, "stock must not be negative")

	var totalQty int
	db.Model(&domain.Checkout{}).Where("product_id = ?", productID).Select("COALESCE(SUM(quantity), 0)").Scan(&totalQty)
	assert.LessOrEqual(t, totalQty, 10, "total sold must not exceed stock")
	assert.Equal(t, 10-totalQty, product.Stock)
}
