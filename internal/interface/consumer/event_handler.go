package consumer

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/payment-service/internal/application"
)

// EventHandler handles incoming RabbitMQ messages for payment events
type EventHandler struct {
	paymentSvc application.PaymentService
}

// NewEventHandler creates a new event handler with the given payment service
func NewEventHandler(paymentSvc application.PaymentService) *EventHandler {
	return &EventHandler{paymentSvc: paymentSvc}
}

// Handle processes incoming AMQP messages
func (h *EventHandler) Handle(ctx context.Context, msg amqp091.Delivery) error {
	var message events.AmqpMessage

	if msg.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	if err := sonic.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	switch msg.RoutingKey {
	case events.PaymentCmdCreateSession:
		return h.handleCreateSession(ctx, message)
	default:
		return fmt.Errorf("unknown routing key: %s", msg.RoutingKey)
	}
}

func (h *EventHandler) handleCreateSession(ctx context.Context, message events.AmqpMessage) error {
	var payload events.PaymentSelectCardData
	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Call application layer - publishing is handled there
	err := h.paymentSvc.CreatePaymentSessionWithCard(
		ctx,
		payload.TripID,
		payload.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment session: %w", err)
	}
	return nil
}
