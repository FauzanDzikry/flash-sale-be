package handler

import (
	"bytes"
	"encoding/json"
	"flash-sale-be/internal/mocks"
	"flash-sale-be/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupCheckoutRouter(h *CheckoutHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/checkouts", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.Checkout(c)
	})
	return r
}

func setupCheckoutRouterUnauth(h *CheckoutHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/checkouts", h.Checkout)
	return r
}

func TestCheckoutHandler_Checkout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	checkoutSvc := mocks.NewMockCheckoutService(ctrl)
	h := NewCheckoutHandler(checkoutSvc)

	checkoutSvc.EXPECT().
		EnqueueCheckout(gomock.Any(), "user-123", gomock.Any()).
		DoAndReturn(func(_ interface{}, userID string, req interface{}) (string, error) {
			assert.Equal(t, "user-123", userID)
			return "job-id-456", nil
		})

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": uuid.New().String(),
		"quantity":   2,
	})
	req := httptest.NewRequest(http.MethodPost, "/checkouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupCheckoutRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "job-id-456", resp["job_id"])
}

func TestCheckoutHandler_Checkout_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	checkoutSvc := mocks.NewMockCheckoutService(ctrl)
	h := NewCheckoutHandler(checkoutSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	req := httptest.NewRequest(http.MethodPost, "/checkouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupCheckoutRouterUnauth(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCheckoutHandler_Checkout_ProductNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	checkoutSvc := mocks.NewMockCheckoutService(ctrl)
	h := NewCheckoutHandler(checkoutSvc)

	checkoutSvc.EXPECT().
		EnqueueCheckout(gomock.Any(), "user-123", gomock.Any()).
		Return("", service.ErrCheckoutProductNotFound)

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	req := httptest.NewRequest(http.MethodPost, "/checkouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupCheckoutRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestCheckoutHandler_Checkout_InsufficientStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	checkoutSvc := mocks.NewMockCheckoutService(ctrl)
	h := NewCheckoutHandler(checkoutSvc)

	checkoutSvc.EXPECT().
		EnqueueCheckout(gomock.Any(), "user-123", gomock.Any()).
		Return("", service.ErrCheckoutInsufficientStock)

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": uuid.New().String(),
		"quantity":   100,
	})
	req := httptest.NewRequest(http.MethodPost, "/checkouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupCheckoutRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckoutHandler_Checkout_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	checkoutSvc := mocks.NewMockCheckoutService(ctrl)
	h := NewCheckoutHandler(checkoutSvc)

	req := httptest.NewRequest(http.MethodPost, "/checkouts", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupCheckoutRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
