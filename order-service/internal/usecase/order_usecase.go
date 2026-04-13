package usecase

import (
	"context"
	"fmt"
	"time"

	// Исправлен путь к пакету domain
	"github.com/yerdembek/AP2_assignment2/order-service/internal/domain"
)

// OrderRepository defines the persistence contract used by the use case.
type OrderRepository interface {
	Save(ctx context.Context, o *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	ListenForUpdates(ctx context.Context, orderID string) (<-chan string, error)
}

// PaymentClient defines the gRPC client contract used by the use case.
type PaymentClient interface {
	ProcessPayment(ctx context.Context, orderID, userID, currency string, amount float64) (string, string, error)
}

// OrderUseCase holds all business logic for orders.
type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient}
}

// CreateOrder creates an order and immediately processes payment via gRPC.
func (uc *OrderUseCase) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	order := &domain.Order{
		ID:        fmt.Sprintf("ord_%d", time.Now().UnixNano()),
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if order.Currency == "" {
		order.Currency = "USD"
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	// Call Payment Service over gRPC.
	_, payStatus, err := uc.paymentClient.ProcessPayment(ctx, order.ID, order.UserID, order.Currency, order.Amount)
	if err != nil {
		// Mark the order as failed but do NOT roll back — keep the record.
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed)
		order.Status = domain.StatusFailed
		return order, fmt.Errorf("payment failed: %w", err)
	}

	// Transition to PAID on successful payment.
	newStatus := domain.StatusPaid
	if payStatus != "SUCCESS" {
		newStatus = domain.StatusFailed
	}
	if err := uc.repo.UpdateStatus(ctx, order.ID, newStatus); err != nil {
		return nil, fmt.Errorf("update order status: %w", err)
	}
	order.Status = newStatus
	return order, nil
}

// GetOrder returns an order by its ID.
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.FindByID(ctx, id)
}

// UpdateOrderStatus allows manual status transitions (e.g. from admin endpoints).
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}

// SubscribeToOrderUpdates returns a channel of status strings backed by the DB listener.
func (uc *OrderUseCase) SubscribeToOrderUpdates(ctx context.Context, orderID string) (<-chan string, error) {
	return uc.repo.ListenForUpdates(ctx, orderID)
}
