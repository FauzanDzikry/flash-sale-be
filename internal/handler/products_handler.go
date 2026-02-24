package handler

import (
	"errors"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductsHandler struct {
	productsService service.ProductsService
}

func NewProductsHandler(productsService service.ProductsService) *ProductsHandler {
	return &ProductsHandler{productsService: productsService}
}

// Create product endpoint
// POST /api/v1/products
func (h *ProductsHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request", "error": err.Error()})
		return
	}
	product, err := h.productsService.Create(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"message": "Product with this name already exists", "error": err.Error()})
		case errors.Is(err, service.ErrProductStockInvalid), errors.Is(err, service.ErrProductPriceInvalid), errors.Is(err, service.ErrProductDiscountInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product data", "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create product", "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, product)
}

// Update product endpoint
// PUT /api/v1/products/:id
func (h *ProductsHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request", "error": err.Error()})
		return
	}
	product, err := h.productsService.Update(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found", "error": err.Error()})
		case errors.Is(err, service.ErrProductAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"message": "Product with this name already exists", "error": err.Error()})
		case errors.Is(err, service.ErrProductStockInvalid), errors.Is(err, service.ErrProductPriceInvalid), errors.Is(err, service.ErrProductDiscountInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product data", "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update product", "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

// Get product by id endpoint
// GET /api/v1/products/:id
func (h *ProductsHandler) GetProductById(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	id := c.Param("id")
	product, err := h.productsService.GetById(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found", "error": err.Error()})
		case errors.Is(err, service.ErrProductAccessDenied):
			c.JSON(http.StatusForbidden, gin.H{"message": "You do not have access to this product", "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get product", "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

// GetAllProductsByUser returns only products owned by the logged-in user.
// GET /api/v1/products
func (h *ProductsHandler) GetAllProductsByUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	products, err := h.productsService.GetAllByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get products", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GetAllProducts returns all products (all users).
// GET /api/v1/products/all
func (h *ProductsHandler) GetAllProducts(c *gin.Context) {
	products, err := h.productsService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get all products", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// Delete product endpoint
// DELETE /api/v1/products/:id
func (h *ProductsHandler) DeleteProduct(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	id := c.Param("id")
	if err := h.productsService.Delete(id, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found", "error": err.Error()})
		case errors.Is(err, service.ErrProductAccessDenied):
			c.JSON(http.StatusForbidden, gin.H{"message": "You do not have access to this product", "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete product", "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
