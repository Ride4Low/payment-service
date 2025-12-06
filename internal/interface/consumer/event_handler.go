package consumer

import (
	"context"
	"fmt"
	"log"

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
	// TODO: Extract tripID, userID, driverID, amount, currency from message.Data
	// For now, placeholder implementation
	log.Printf("Received create session request: %+v", message)

	var payload events.PaymentTripResponseData
	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	paymentSession, err := h.paymentSvc.CreatePaymentSession(
		ctx,
		payload.TripID,
		payload.UserID,
		payload.DriverID,
		int64(payload.Amount),
		payload.Currency,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment session: %w", err)
	}

	log.Printf("Created payment session: %+v", paymentSession)

	return nil
}
