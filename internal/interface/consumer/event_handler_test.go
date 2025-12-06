package consumer

import (
	"context"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ride4Low/contracts/events"
)

// mockPaymentService is a mock implementation of application.PaymentService
type mockPaymentService struct {
	err    error
	called bool
}

func (m *mockPaymentService) CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestNewEventHandler(t *testing.T) {
	mockSvc := &mockPaymentService{}
	handler := NewEventHandler(mockSvc)

	if handler == nil {
		t.Fatal("expected non-nil EventHandler")
	}

	if handler.paymentSvc == nil {
		t.Error("expected non-nil paymentSvc")
	}
}

func TestEventHandler_Handle_NilBody(t *testing.T) {
	mockSvc := &mockPaymentService{}
	handler := NewEventHandler(mockSvc)

	msg := amqp091.Delivery{
		Body: nil,
	}

	err := handler.Handle(context.Background(), msg)

	if err == nil {
		t.Fatal("expected error for nil body")
	}

	if err.Error() != "message body is nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestEventHandler_Handle_InvalidJSON(t *testing.T) {
	mockSvc := &mockPaymentService{}
	handler := NewEventHandler(mockSvc)

	msg := amqp091.Delivery{
		Body: []byte("invalid json"),
	}

	err := handler.Handle(context.Background(), msg)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestEventHandler_Handle_UnknownRoutingKey(t *testing.T) {
	mockSvc := &mockPaymentService{}
	handler := NewEventHandler(mockSvc)

	message := events.AmqpMessage{
		OwnerID: "user-123",
	}
	body, _ := sonic.Marshal(message)

	msg := amqp091.Delivery{
		Body:       body,
		RoutingKey: "unknown.routing.key",
	}

	err := handler.Handle(context.Background(), msg)

	if err == nil {
		t.Fatal("expected error for unknown routing key")
	}
}

func TestEventHandler_Handle_CreateSessionRoutingKey(t *testing.T) {
	mockSvc := &mockPaymentService{}
	handler := NewEventHandler(mockSvc)

	message := events.AmqpMessage{
		OwnerID: "user-123",
	}
	body, _ := sonic.Marshal(message)

	msg := amqp091.Delivery{
		Body:       body,
		RoutingKey: events.PaymentCmdCreateSession,
	}

	err := handler.Handle(context.Background(), msg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
