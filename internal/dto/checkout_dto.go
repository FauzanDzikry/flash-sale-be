package dto

import "time"

type CheckoutRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type CheckoutResponse struct {
	ID         string     `json:"id"`
	ProductID  string     `json:"product_id"`
	Quantity   int        `json:"quantity"`
	Price      float64    `json:"price"`
	Discount   float64    `json:"discount"`
	TotalPrice float64    `json:"total_price"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
