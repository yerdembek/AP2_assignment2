package domain

import "time"

// Order is the core domain entity.
type Order struct {
	ID        string
	UserID    string
	ProductID string
	Quantity  int
	Amount    float64
	Currency  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	StatusPending    = "PENDING"
	StatusPaid       = "PAID"
	StatusProcessing = "PROCESSING"
	StatusShipped    = "SHIPPED"
	StatusDelivered  = "DELIVERED"
	StatusCancelled  = "CANCELLED"
	StatusFailed     = "FAILED"
)

// CreateOrderRequest carries input data for order creation.
type CreateOrderRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency"`
}
