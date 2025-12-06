package application

import (
	"context"

	"github.com/ride4Low/payment-service/internal/domain"
)

// PaymentService is the application service port (use cases)
type PaymentService interface {
	CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) (*domain.PaymentIntent, error)
}

// PaymentProvider is the port interface for payment providers (Stripe, PayPal, etc.)
// This is a secondary/driven port - implemented by infrastructure adapters
type PaymentProvider interface {
	CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error)
}
