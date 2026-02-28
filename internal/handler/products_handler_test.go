package handler

import (
	"bytes"
	"encoding/json"
	"flash-sale-be/internal/dto"
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

func setupProductsRouter(h *ProductsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/products", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.CreateProduct(c)
	})
	r.PUT("/products/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.UpdateProduct(c)
	})
	r.GET("/products/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.GetProductById(c)
	})
	r.DELETE("/products/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.DeleteProduct(c)
	})
	return r
}

func TestProductsHandler_CreateProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsSvc := mocks.NewMockProductsService(ctrl)
	h := NewProductsHandler(productsSvc)

	productsSvc.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
			assert.Equal(t, "New Product", req.Name)
			return &dto.ProductResponse{
				ID:       uuid.New().String(),
				Name:     "New Product",
				Category: "Electronics",
				Stock:    10,
				Price:    99.99,
			}, nil
		})

	body, _ := json.Marshal(map[string]interface{}{
		"name":      "New Product",
		"category":  "Electronics",
		"stock":     10,
		"price":     99.99,
		"discount":  0,
		"created_by": uuid.New().String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupProductsRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestProductsHandler_CreateProduct_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsSvc := mocks.NewMockProductsService(ctrl)
	h := NewProductsHandler(productsSvc)

	productsSvc.EXPECT().
		Create(gomock.Any()).
		Return(nil, service.ErrProductAlreadyExists)

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "Existing",
		"category":   "Test",
		"stock":      5,
		"price":      10,
		"discount":   0,
		"created_by": uuid.New().String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := setupProductsRouter(h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestProductsHandler_GetProductById_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsSvc := mocks.NewMockProductsService(ctrl)
	h := NewProductsHandler(productsSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/products/:id", h.GetProductById)

	req := httptest.NewRequest(http.MethodGet, "/products/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProductsHandler_GetProductById_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productsSvc := mocks.NewMockProductsService(ctrl)
	h := NewProductsHandler(productsSvc)

	productID := uuid.New().String()
	productsSvc.EXPECT().
		GetById(productID, "user-123").
		Return(nil, service.ErrProductAccessDenied)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/products/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		h.GetProductById(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/products/"+productID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}
