package dto

import "time"

type ProductResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Category  string     `json:"category"`
	Stock     int        `json:"stock"`
	Price     float64    `json:"price"`
	Discount  float64    `json:"discount"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	CreatedBy string     `json:"created_by"`
}

type CreateProductRequest struct {
	Name      string  `json:"name" binding:"required"`
	Category  string  `json:"category" binding:"required"`
	Stock     int     `json:"stock" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Discount  float64 `json:"discount" binding:"required"`
	CreatedBy string  `json:"created_by" binding:"required"`
}

type UpdateProductRequest struct {
	Name     string  `json:"name" binding:"required"`
	Category string  `json:"category" binding:"required"`
	Stock    int     `json:"stock" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Discount float64 `json:"discount" binding:"required"`
}
