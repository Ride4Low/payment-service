package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/payment-service/internal/application"
)

// mockMessagePublisher is a mock implementation of MessagePublisher for testing
type mockMessagePublisher struct {
	called     bool
	routingKey string
	message    events.AmqpMessage
	err        error
}

func (m *mockMessagePublisher) PublishMessage(ctx context.Context, routingKey string, message events.AmqpMessage) error {
	m.called = true
	m.routingKey = routingKey
	m.message = message
	return m.err
}

func TestNewRabbitMQPublisher(t *testing.T) {
	mockPub := &mockMessagePublisher{}
	publisher := NewRabbitMQPublisher(mockPub)

	if publisher == nil {
		t.Fatal("expected non-nil RabbitMQPublisher")
	}
}

func TestRabbitMQPublisher_PublishPaymentSessionCreated_Success(t *testing.T) {
	mockPub := &mockMessagePublisher{}
	publisher := NewRabbitMQPublisher(mockPub)

	event := &application.PaymentSessionCreatedEvent{
		UserID: "user-123",
		PaymentEventSessionCreatedData: events.PaymentEventSessionCreatedData{
			TripID:    "trip-456",
			SessionID: "cs_session_789",
			Amount:    10.50,
			Currency:  "usd",
		},
	}

	err := publisher.PublishPaymentSessionCreated(context.Background(), event)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockPub.called {
		t.Error("expected PublishMessage to be called")
	}

	if mockPub.routingKey != events.PaymentEventSessionCreated {
		t.Errorf("expected routing key '%s', got '%s'", events.PaymentEventSessionCreated, mockPub.routingKey)
	}

	if mockPub.message.OwnerID != "user-123" {
		t.Errorf("expected OwnerID 'user-123', got '%s'", mockPub.message.OwnerID)
	}

	// Verify payload data is present
	if len(mockPub.message.Data) == 0 {
		t.Error("expected non-empty message Data")
	}
}

func TestRabbitMQPublisher_PublishPaymentSessionCreated_Error(t *testing.T) {
	publishErr := errors.New("connection failed")
	mockPub := &mockMessagePublisher{err: publishErr}
	publisher := NewRabbitMQPublisher(mockPub)

	event := &application.PaymentSessionCreatedEvent{
		UserID: "user-123",
		PaymentEventSessionCreatedData: events.PaymentEventSessionCreatedData{
			TripID:    "trip-456",
			SessionID: "cs_session_789",
			Amount:    10.50,
			Currency:  "usd",
		},
	}

	err := publisher.PublishPaymentSessionCreated(context.Background(), event)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, publishErr) {
		t.Errorf("expected error '%v', got '%v'", publishErr, err)
	}
}
