package handler

import (
	"errors"
	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckoutHandler struct {
	checkoutService service.CheckoutService
}

func NewCheckoutHandler(checkoutService service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{checkoutService: checkoutService}
}

// Checkout enqueues a checkout job and returns 202 with job_id.
// POST /api/v1/checkouts
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	var req dto.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request", "error": err.Error()})
		return
	}
	jobID, err := h.checkoutService.EnqueueCheckout(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCheckoutProductNotFound), errors.Is(err, service.ErrCheckoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found", "error": err.Error()})
		case errors.Is(err, service.ErrCheckoutInsufficientStock):
			c.JSON(http.StatusBadRequest, gin.H{"message": "Insufficient stock", "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to enqueue checkout", "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Checkout accepted", "job_id": jobID})
}
