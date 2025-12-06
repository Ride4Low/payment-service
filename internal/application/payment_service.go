package application

import (
	"context"
	"time"

	"github.com/ride4Low/payment-service/internal/domain"
)

// paymentService implements PaymentService interface
type paymentService struct {
	provider PaymentProvider
}

// NewPaymentService creates a new payment service with the given provider
func NewPaymentService(provider PaymentProvider) PaymentService {
	return &paymentService{provider: provider}
}

// CreatePaymentSession creates a payment session using the payment provider
func (s *paymentService) CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) (*domain.PaymentIntent, error) {
	metadata := map[string]string{
		"trip_id":   tripID,
		"user_id":   userID,
		"driver_id": driverID,
	}

	sessionID, err := s.provider.CreatePaymentSession(ctx, amount, currency, metadata)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentIntent{
		TripID:          tripID,
		UserID:          userID,
		DriverID:        driverID,
		Amount:          amount,
		Currency:        currency,
		StripeSessionID: sessionID,
		CreatedAt:       time.Now(),
	}, nil
}
