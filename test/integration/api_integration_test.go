//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flash-sale-be/internal/config"
	"flash-sale-be/internal/domain"
	"flash-sale-be/internal/queue"
	"flash-sale-be/internal/repository"
	"flash-sale-be/internal/router"
	"flash-sale-be/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flash-sale-be/test/testutil"
)

func TestAPI_FullFlow_RegisterLoginCreateProductCheckout(t *testing.T) {
	db, cleanupDB := testutil.SetupTestDB(t)
	defer cleanupDB()

	rdb, cleanupRedis := testutil.SetupTestRedis(t)
	defer cleanupRedis()

	cfg := &config.Config{
		JWTKey:        "integration-test-secret",
		JWTExpireHour: 24,
	}

	q := queue.NewRedisQueue(rdb)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)
	checkoutSvc := service.NewCheckoutService(checkoutRepo, productsRepo, q, db)

	r := router.New(router.Deps{
		DB:              db,
		Cfg:             cfg,
		CheckoutService: checkoutSvc,
		Redis:           rdb,
	})

	go runCheckoutWorker(context.Background(), q, checkoutSvc)

	// 1. Register
	registerBody, _ := json.Marshal(map[string]string{
		"email":    "flow@example.com",
		"password": "password123",
		"name":     "Flow User",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "flow@example.com",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.AccessToken)
	userID := loginResp.User.ID

	// 3. Create Product
	productBody, _ := json.Marshal(map[string]interface{}{
		"name":       "Flash Product",
		"category":   "Electronics",
		"stock":      10,
		"price":     100,
		"discount":  10,
		"created_by": userID,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/products/", bytes.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var productResp struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &productResp))
	productID := productResp.ID
	require.NotEmpty(t, productID)

	// 4. Checkout (enqueue)
	checkoutBody, _ := json.Marshal(map[string]interface{}{
		"product_id": productID,
		"quantity":   2,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/checkouts/", bytes.NewReader(checkoutBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	time.Sleep(2 * time.Second)

	pid, _ := uuid.Parse(productID)
	var product domain.Product
	require.NoError(t, db.Where("id = ?", pid).First(&product).Error)
	assert.Equal(t, 8, product.Stock)

	var count int64
	db.Model(&domain.Checkout{}).Where("product_id = ?", pid).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}

func runCheckoutWorker(ctx context.Context, q queue.Queue, svc service.CheckoutService) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job, err := q.DequeueCheckout(ctx)
			if err != nil {
				if errors.Is(err, queue.ErrEmptyQueue) {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				continue
			}
			_, _ = svc.ProcessCheckoutJob(ctx, job)
		}
	}
}
