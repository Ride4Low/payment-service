package mongodb

import (
	"context"
	"fmt"

	"github.com/ride4Low/contracts/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TripRepository is the MongoDB implementation of types.TripRepository
type TripRepository struct {
	collection *mongo.Collection
}

// NewTripRepository creates a new MongoDB trip repository
func NewTripRepository(db *mongo.Database) *TripRepository {
	return &TripRepository{
		collection: db.Collection(TripsCollection),
	}
}

func (r *TripRepository) GetTripByID(ctx context.Context, tripID string) (*types.Trip, error) {
	var trip types.Trip
	err := r.collection.FindOne(ctx, bson.M{"_id": tripID}).Decode(&trip)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("trip not found: %s", tripID)
		}
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	return &trip, nil
}
