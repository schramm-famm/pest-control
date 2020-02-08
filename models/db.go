package models

import (
	"context"
	"crypto/tls"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Datastore interface {
	GetPrefs(int) (*GlobalPrefs, error)
	GetPrefsConv(int, int) (*ConversationPrefs, error)
	CreatePrefs(*Preferences) error
	CreatePrefsConv(int, *ConversationPrefs) error
	DeletePrefs(int) error
	DeletePrefsConv(int, int) error
	PatchPrefs(int, *GlobalPrefs) error
	PatchPrefsConv(int, int, *ConversationPrefs) error
}

type DB struct {
	*mongo.Client
}

func NewDB(dataSourceName string, tlsConfig *tls.Config) (*DB, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(dataSourceName)
	if tlsConfig != nil {
		clientOptions.SetTLSConfig(tlsConfig)
	}

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
