package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Datastore interface {
	GetPrefs(string, string) (*Preferences, error)
	CreatePrefs(Preferences) (string, error)
}

type DB struct {
	*mongo.Client
}

func NewDB(dataSourceName string) (*DB, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(dataSourceName)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}
	return &DB{client}, nil
}
