package models

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPrefsExists = errors.New("preferences for user already exists")
	ErrPrefsDNE    = errors.New("preferences for user does not exist")
)

type PrefsFilter struct {
	UserID         int `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ConversationID int `json:"conversation_id,omitempty" bson:"conversation.conversation_id,omitempty"`
}

type GlobalPrefs struct {
	UserID       int  `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Invitation   bool `json:"invitation,omitempty" bson:"invitation,omitempty"`
	TextEntered  bool `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified bool `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag          bool `json:"tag,omitempty" bson:"tag,omitempty"`
	Role         bool `json:"role,omitempty" bson:"role,omitempty"`
}

type ConversationPrefs struct {
	UserID         int  `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ConversationID int  `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	TextEntered    bool `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified   bool `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag            bool `json:"tag,omitempty" bson:"tag,omitempty"`
	Role           bool `json:"role,omitempty" bson:"role,omitempty"`
}

type Preferences struct {
	ID           string               `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID       int                  `json:"user_id,omitempty" bson:"user_id,omitempty"`
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
		TextEntered:  true,
		TextModified: true,
		Tag:          true,
		Role:         true,
	}
}

func NewPreferences() *Preferences {
	return &Preferences{
		Global: NewGlobalPrefs(),
	}
}

func (g *GlobalPrefs) String() string {
	return fmt.Sprintf("%+v", *g)
}

func (c *ConversationPrefs) String() string {
	return fmt.Sprintf("%+v", *c)
}

func (db *DB) GetPrefs(userID int) (*GlobalPrefs, error) {
	filter, err := bson.Marshal(PrefsFilter{UserID: userID})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return nil, err
	}

	opts := options.FindOne().SetProjection(bson.D{{"global", 1}})
	collection := db.Database("pest-control").Collection("prefs")
	singleResult := collection.FindOne(context.TODO(), filter, opts)
	if singleResult.Err() != nil {
		return nil, singleResult.Err()
	}

	prefs := &Preferences{}
	if err := singleResult.Decode(prefs); err != nil {
		log.Printf("failed to decode retrieved data (%+v): %s", singleResult, err.Error())
		return nil, err
	}
	return prefs.Global, nil
}

func (db *DB) GetPrefsConv(userID, conversationID int) (*ConversationPrefs, error) {
	filter, err := bson.Marshal(PrefsFilter{
		UserID:         userID,
		ConversationID: conversationID,
	})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return nil, err
	}

	opts := options.FindOne().SetProjection(bson.D{{"conversation", 1}})
	collection := db.Database("pest-control").Collection("prefs")
	singleResult := collection.FindOne(context.TODO(), filter, opts)
	if singleResult.Err() != nil {
		return nil, singleResult.Err()
	}

	prefs := &Preferences{}
	if err := singleResult.Decode(prefs); err != nil {
		log.Printf("failed to decode retrieved data (%+v): %s", singleResult, err.Error())
		return nil, err
	}
	return prefs.Conversation[0], nil
}

func (db *DB) CreatePrefs(prefs *Preferences) error {
	if _, err := db.GetPrefs(prefs.UserID); err != nil && err != mongo.ErrNoDocuments {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		return err
	} else if err == nil {
		log.Printf("preferences for user (%d) already exists", prefs.UserID)
		return ErrPrefsExists
	}

	collection := db.Database("pest-control").Collection("prefs")
	insertResult, err := collection.InsertOne(context.TODO(), prefs)
	if err != nil {
		log.Printf(
			"failed to insert preferences (%+v) into MongoDB collection: %s",
			prefs,
			err.Error(),
		)
		return err
	}

	prefs.ID = insertResult.InsertedID.(primitive.ObjectID).String()

	return nil
}

func (db *DB) CreatePrefsConv(convPrefs *ConversationPrefs) error {
	if _, err := db.GetPrefsConv(convPrefs.UserID, convPrefs.ConversationID); err != nil && err != mongo.ErrNoDocuments {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		return err
	} else if err == nil {
		log.Printf(
			"conversation (%d) preferences for user (%d) already exists",
			convPrefs.ConversationID,
			convPrefs.UserID,
		)
		return ErrPrefsExists
	}

	filter, err := bson.Marshal(PrefsFilter{UserID: convPrefs.UserID})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return err
	}
	convPrefs.UserID = 0
	update := bson.D{{"$push", bson.D{{Key: "conversation", Value: convPrefs}}}}
	collection := db.Database("pest-control").Collection("prefs")
	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf(
			"failed to insert conversation preferences (%+v) into MongoDB collection: %s",
			convPrefs,
			err.Error(),
		)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		return ErrPrefsDNE
	}

	return nil
}

func (db *DB) DeletePrefs(userID int) error {
	filter, err := bson.Marshal(PrefsFilter{UserID: userID})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return err
	}

	collection := db.Database("pest-control").Collection("prefs")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Printf(
			"failed to delete preferences (%+v) from MongoDB collection: %s",
			filter,
			err.Error(),
		)
		return err
	}

	// No preferences were deleted which means that the user did not have any
	// preferences to begin with
	if deleteResult.DeletedCount == 0 {
		return ErrPrefsDNE
	}
	return nil
}

func (db *DB) DeletePrefsConv(userID, conversationID int) error {
	filter, err := bson.Marshal(PrefsFilter{
		UserID: userID,
	})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return err
	}
	update := bson.D{{
		"$pull",
		bson.D{{
			Key:   "conversation",
			Value: bson.D{{Key: "conversation_id", Value: conversationID}},
		}},
	}}

	collection := db.Database("pest-control").Collection("prefs")
	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf(
			"failed to delete preferences (%+v) from MongoDB collection: %s",
			filter,
			err.Error(),
		)
		return err
	}

	// No preferences were deleted which means that the user did not have any
	// preferences to begin with
	if updateResult.ModifiedCount == 0 {
		return ErrPrefsDNE
	}
	return nil
}

func (db *DB) PatchPrefs(prefs *GlobalPrefs) error {
	filter, err := bson.Marshal(PrefsFilter{
		UserID: prefs.UserID,
	})
	if err != nil {
		log.Printf("failed to create query filter: %s", err.Error())
		return err
	}

	prefs.UserID = 0
	update := bson.D{{
		"$set",
		bson.D{{
			Key:   "global",
			Value: prefs,
		}},
	}}
	collection := db.Database("pest-control").Collection("prefs")
	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf(
			"failed to update preferences (%+v) in MongoDB collection: %s",
			filter,
			err.Error(),
		)
		return err
	}

	// No preferences were deleted which means that the user did not have any
	// preferences to begin with
	if updateResult.ModifiedCount == 0 {
		return ErrPrefsDNE
	}
	return nil
}
