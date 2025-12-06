package messaging

import (
	"context"
	"encoding/json"

	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/payment-service/internal/application"
)

// MessagePublisher is the interface for publishing messages (allows mocking in tests)
type MessagePublisher interface {
	PublishMessage(ctx context.Context, routingKey string, message events.AmqpMessage) error
}

// RabbitMQPublisher implements application.EventPublisher using RabbitMQ
type RabbitMQPublisher struct {
	publisher MessagePublisher
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher
func NewRabbitMQPublisher(publisher MessagePublisher) *RabbitMQPublisher {
	return &RabbitMQPublisher{publisher: publisher}
}

// PublishPaymentSessionCreated publishes a payment session created event
func (p *RabbitMQPublisher) PublishPaymentSessionCreated(ctx context.Context, event *application.PaymentSessionCreatedEvent) error {

	payloadBytes, err := json.Marshal(event.PaymentEventSessionCreatedData)
	if err != nil {
		return err
	}

	return p.publisher.PublishMessage(
		ctx,
		events.PaymentEventSessionCreated,
		events.AmqpMessage{
			OwnerID: event.UserID,
			Data:    payloadBytes,
		},
	)
}
