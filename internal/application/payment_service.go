package application

import (
	"context"

	"github.com/ride4Low/contracts/events"
)

// paymentService implements PaymentService interface
type paymentService struct {
	provider   PaymentProvider
	publisher  EventPublisher
	repository TripRepository
}

// NewPaymentService creates a new payment service with the given provider, publisher, and repository
func NewPaymentService(provider PaymentProvider, publisher EventPublisher, repository TripRepository) PaymentService {
	return &paymentService{
		provider:   provider,
		publisher:  publisher,
		repository: repository,
	}
}

// CreatePaymentSession creates a payment session using the payment provider
func (s *paymentService) CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) error {
	metadata := map[string]string{
		"trip_id":   tripID,
		"user_id":   userID,
		"driver_id": driverID,
	}

	sessionID, err := s.provider.CreatePaymentSession(ctx, amount, currency, metadata)
	if err != nil {
		return err
	}

	msg := &PaymentSessionCreatedEvent{
		UserID: userID,
		PaymentEventSessionCreatedData: events.PaymentEventSessionCreatedData{
			TripID:    tripID,
			SessionID: sessionID,
			Amount:    float64(amount) / 100.0,
			Currency:  currency,
		},
	}

	// Publish the event from application layer (business logic decides when to publish)
	if err := s.publisher.PublishPaymentSessionCreated(ctx, msg); err != nil {
		return err
	}

	return nil
}
