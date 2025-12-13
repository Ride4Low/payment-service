package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ride4Low/contracts/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TripsCollection = "trips"
)

// MongoConfig holds MongoDB connection configuration
type MongoConfig struct {
	URI      string
	Database string
}

// NewMongoDefaultConfig creates a default MongoDB configuration from environment variables
func NewMongoDefaultConfig() *MongoConfig {
	return &MongoConfig{
		URI:      env.GetString("MONGODB_URI", ""),
		Database: env.GetString("MONGODB_DATABASE", ""),
	}
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mongodb database is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
}

// GetDatabase returns the database instance
func GetDatabase(client *mongo.Client, database string) *mongo.Database {
	return client.Database(database)
}
