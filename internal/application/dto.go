package application

import "github.com/ride4Low/contracts/events"

// PaymentSessionCreatedEvent represents the event data when a payment session is created
type PaymentSessionCreatedEvent struct {
	UserID string
	events.PaymentEventSessionCreatedData
}
