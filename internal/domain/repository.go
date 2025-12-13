package domain

import (
	"context"

	"github.com/ride4Low/contracts/types"
)

// TripRepository is the port interface for payment persistence
// This is a secondary/driven port - implemented by infrastructure adapters (e.g., MongoDB)
type TripRepository interface {
	GetTripByID(ctx context.Context, tripID string) (*types.Trip, error)
}
