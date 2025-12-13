package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ride4Low/contracts/types"
)

// mockPaymentProvider is a mock implementation of PaymentProvider for testing
type mockPaymentProvider struct {
	sessionID string
	err       error
}

func (m *mockPaymentProvider) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.sessionID, nil
}

// mockEventPublisher is a mock implementation of EventPublisher for testing
type mockEventPublisher struct {
	called bool
	event  PaymentSessionCreatedEvent
	err    error
}

func (m *mockEventPublisher) PublishPaymentSessionCreated(ctx context.Context, event *PaymentSessionCreatedEvent) error {
	m.called = true
	m.event = *event
	return m.err
}

type mockTripRepository struct {
	getByIDCalled bool
	getByIDErr    error
	trip          *types.Trip
}

func (m *mockTripRepository) GetTripByID(ctx context.Context, id string) (*types.Trip, error) {
	m.getByIDCalled = true
	return m.trip, m.getByIDErr
}

func TestNewPaymentService(t *testing.T) {
	provider := &mockPaymentProvider{}
	publisher := &mockEventPublisher{}
	tripRepository := &mockTripRepository{}
	svc := NewPaymentService(provider, publisher, tripRepository)

	if svc == nil {
		t.Fatal("expected non-nil PaymentService")
	}
}

func TestPaymentService_CreatePaymentSession_Success(t *testing.T) {
	provider := &mockPaymentProvider{sessionID: "cs_test_session_123"}
	publisher := &mockEventPublisher{}
	tripRepository := &mockTripRepository{}
	svc := NewPaymentService(provider, publisher, tripRepository)

	ctx := context.Background()
	err := svc.CreatePaymentSession(ctx, "trip-1", "user-1", "driver-1", 1000, "usd")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify publisher was called
	if !publisher.called {
		t.Error("expected publisher to be called")
	}
}

func TestPaymentService_CreatePaymentSession_ProviderError(t *testing.T) {
	providerErr := errors.New("stripe api error")
	provider := &mockPaymentProvider{err: providerErr}
	publisher := &mockEventPublisher{}
	tripRepository := &mockTripRepository{}
	svc := NewPaymentService(provider, publisher, tripRepository)

	ctx := context.Background()
	err := svc.CreatePaymentSession(ctx, "trip-1", "user-1", "driver-1", 1000, "usd")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, providerErr) {
		t.Errorf("expected error to be '%v', got '%v'", providerErr, err)
	}
}

func TestPaymentService_CreatePaymentSession_PublisherError(t *testing.T) {
	publisherErr := errors.New("rabbitmq error")
	provider := &mockPaymentProvider{sessionID: "cs_test_session_123"}
	publisher := &mockEventPublisher{err: publisherErr}
	tripRepository := &mockTripRepository{}

	svc := NewPaymentService(provider, publisher, tripRepository)

	ctx := context.Background()
	err := svc.CreatePaymentSession(ctx, "trip-1", "user-1", "driver-1", 1000, "usd")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

}

func TestPaymentService_CreatePaymentSession_EventData(t *testing.T) {
	provider := &mockPaymentProvider{sessionID: "cs_test_session_456"}
	publisher := &mockEventPublisher{}
	tripRepository := &mockTripRepository{}

	svc := NewPaymentService(provider, publisher, tripRepository)

	ctx := context.Background()
	err := svc.CreatePaymentSession(ctx, "trip-123", "user-456", "driver-789", 2500, "eur")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify event data is correct
	if publisher.event.UserID != "user-456" {
		t.Errorf("expected event UserID 'user-456', got '%s'", publisher.event.UserID)
	}

	if publisher.event.TripID != "trip-123" {
		t.Errorf("expected event TripID 'trip-123', got '%s'", publisher.event.TripID)
	}

	if publisher.event.SessionID != "cs_test_session_456" {
		t.Errorf("expected event SessionID 'cs_test_session_456', got '%s'", publisher.event.SessionID)
	}

	// 2500 cents = 25.00 dollars
	expectedAmount := 25.0
	if publisher.event.Amount != expectedAmount {
		t.Errorf("expected event Amount %f, got %f", expectedAmount, publisher.event.Amount)
	}

	if publisher.event.Currency != "eur" {
		t.Errorf("expected event Currency 'eur', got '%s'", publisher.event.Currency)
	}
}
