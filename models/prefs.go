package models

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrPrefsExists = errors.New("preferences for user already exists")

type PrefsFilter struct {
	UserID         string `bson:"user_id,omitempty"`
	ConversationID string `bson:"conversation.conversation_id,omitempty"`
}

type GlobalPrefs struct {
	Invitation   bool `json:"invitation,omitempty" bson:"invitation,omitempty"`
	TextEntered  bool `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified bool `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag          bool `json:"tag,omitempty" bson:"tag,omitempty"`
	Role         bool `json:"role,omitempty" bson:"role,omitempty"`
}

type ConversationPrefs struct {
	ConversationID string `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	TextEntered    bool   `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified   bool   `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag            bool   `json:"tag,omitempty" bson:"tag,omitempty"`
	Role           bool   `json:"role,omitempty" bson:"role,omitempty"`
}

type Preferences struct {
	ID           string               `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID       string               `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Global       *GlobalPrefs         `json:"global,omitempty" bson:"global,omitempty"`
	Conversation []*ConversationPrefs `json:"conversation,omitempty" bson:"conversation,omitempty"`
}

func NewGlobalPrefs() *GlobalPrefs {
	return &GlobalPrefs{
		Invitation:   true,
		TextEntered:  true,
		TextModified: true,
		Tag:          true,
		Role:         true,
	}
}

func NewConversationPrefs() *ConversationPrefs {
	return &ConversationPrefs{
		ConversationID: "",
		TextEntered:    true,
		TextModified:   true,
		Tag:            true,
		Role:           true,
	}
}

func NewPreferences() *Preferences {
	return &Preferences{
		UserID: "",
		Global: NewGlobalPrefs(),
	}
}

func (g *GlobalPrefs) String() string {
	return fmt.Sprintf("%+v", *g)
}

func (c *ConversationPrefs) String() string {
	return fmt.Sprintf("%+v", *c)
}

func (db *DB) GetPrefs(userID string, conversationID string) (*Preferences, error) {
	filter, err := bson.Marshal(PrefsFilter{
		UserID:         userID,
		ConversationID: conversationID,
	})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return nil, err
	}

	collection := db.Database("pest-control").Collection("prefs")
	singleResult := collection.FindOne(context.TODO(), filter)
	if singleResult.Err() != nil {
		return nil, singleResult.Err()
	}

	prefs := &Preferences{}
	if err := singleResult.Decode(prefs); err != nil {
		log.Printf("failed to decode retrieved data (%+v): %s", singleResult, err.Error())
		return nil, err
	}
	return prefs, nil
}

func (db *DB) CreatePrefs(prefs Preferences) (string, error) {
	if _, err := db.GetPrefs(prefs.UserID, ""); err != nil && err != mongo.ErrNoDocuments {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		return "", err
	} else if err == nil {
		log.Printf("preferences for user (%s) already exists", prefs.UserID)
		return "", ErrPrefsExists
	}

	collection := db.Database("pest-control").Collection("prefs")
	insertResult, err := collection.InsertOne(context.TODO(), prefs)
	if err != nil {
		log.Printf(
			"failed to insert preferences (%+v) into MongoDB collection: %s",
			prefs,
			err.Error(),
		)
		return "", err
	}
	return insertResult.InsertedID.(primitive.ObjectID).String(), nil
}
