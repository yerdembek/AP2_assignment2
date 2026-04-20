package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yerdembek/AP2_assignment2/payment-service/internal/domain"
)

// PaymentRepository defines the persistence contract for the use case.
type PaymentRepository interface {
	Save(ctx context.Context, p *domain.Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	GetStats(ctx context.Context) (total, auth, decl int64, sumCents int64, err error)
}

// PaymentUseCase contains all business rules for payment processing.
type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) GetPaymentStats(ctx context.Context) (total, auth, decl int64, sumCents int64, err error) {
	return uc.repo.GetStats(ctx)
}

// ProcessPayment validates and processes a payment request.
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, req *domain.Payment) (*domain.Payment, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("invalid amount: must be greater than zero")
	}
	if req.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	req.ID = fmt.Sprintf("pay_%d", time.Now().UnixNano())
	req.Status = domain.StatusSuccess
	req.ProcessedAt = time.Now().UTC()
	req.Message = fmt.Sprintf("Payment of %.2f %s processed successfully", req.Amount, req.Currency)

	if err := uc.repo.Save(ctx, req); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}
	return req, nil
}
