package domain

import "time"

// Payment represents the core payment entity (domain model).
type Payment struct {
	ID          string
	OrderID     string
	UserID      string
	Amount      float64
	Currency    string
	Status      string
	ProcessedAt time.Time
	Message     string
}

const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusPending = "PENDING"
)
