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

type GlobalPrefs struct {
	Invitation   bool `json:"invitation,omitempty" bson:"invitation,omitempty"`
	TextEntered  bool `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified bool `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag          bool `json:"tag,omitempty" bson:"tag,omitempty"`
	Role         bool `json:"role,omitempty" bson:"role,omitempty"`
}

type ConversationPrefs struct {
	ConversationID int  `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	TextEntered    bool `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified   bool `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag            bool `json:"tag,omitempty" bson:"tag,omitempty"`
	Role           bool `json:"role,omitempty" bson:"role,omitempty"`
}

type GlobalPrefsPatch struct {
	Invitation   *bool `json:"invitation,omitempty" bson:"global.invitation,omitempty"`
	TextEntered  *bool `json:"text_entered,omitempty" bson:"global.text_entered,omitempty"`
	TextModified *bool `json:"text_modified,omitempty" bson:"global.text_modified,omitempty"`
	Tag          *bool `json:"tag,omitempty" bson:"global.tag,omitempty"`
	Role         *bool `json:"role,omitempty" bson:"global.role,omitempty"`
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
	filter := bson.D{{"user_id", userID}}
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
	filter := bson.D{{"user_id", userID}}
	opts := options.FindOne().SetProjection(bson.D{{
		"conversation",
		bson.D{{"$elemMatch", bson.D{{"conversation_id", conversationID}}}},
	}})
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

	if len(prefs.Conversation) == 0 {
		return nil, mongo.ErrNoDocuments
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

func (db *DB) CreatePrefsConv(userID int, convPrefs *ConversationPrefs) error {
	if _, err := db.GetPrefsConv(userID, convPrefs.ConversationID); err != nil && err != mongo.ErrNoDocuments {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		return err
	} else if err == nil {
		log.Printf(
			"conversation (%d) preferences for user (%d) already exists",
			convPrefs.ConversationID,
			userID,
		)
		return ErrPrefsExists
	}

	filter := bson.D{{"user_id", userID}}
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
	filter := bson.D{{"user_id", userID}}
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
	filter := bson.D{{"user_id", userID}}
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

func newTruePtr() *bool {
	b := true
	return &b
}

func (db *DB) PatchPrefs(userID int, prefs *GlobalPrefsPatch) error {
	if _, err := db.GetPrefs(userID); err != nil {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		return err
	}

	filter := bson.D{{"user_id", userID}}

	// Set `set` and `unset` values where the fields in `set` that are set to
	// true will be added to the global document and the fields that are true
	// in `unset` will be removed.
	set := GlobalPrefsPatch{}
	unset := GlobalPrefsPatch{}
	if prefs.Invitation != nil && *prefs.Invitation {
		set.Invitation = newTruePtr()
	} else if prefs.Invitation != nil {
		unset.Invitation = newTruePtr()
	}
	if prefs.Role != nil && *prefs.Role {
		set.Role = newTruePtr()
	} else if prefs.Role != nil {
		unset.Role = newTruePtr()
	}
	if prefs.Tag != nil && *prefs.Tag {
		set.Tag = newTruePtr()
	} else if prefs.Tag != nil {
		unset.Tag = newTruePtr()
	}
	if prefs.TextEntered != nil && *prefs.TextEntered {
		set.TextEntered = newTruePtr()
	} else if prefs.TextEntered != nil {
		unset.TextEntered = newTruePtr()
	}
	if prefs.TextModified != nil && *prefs.TextModified {
		set.TextModified = newTruePtr()
	} else if prefs.TextModified != nil {
		unset.TextModified = newTruePtr()
	}

	update := bson.D{}
	if (GlobalPrefsPatch{}) != set {
		update = append(update, bson.E{"$set", set})
	}
	if (GlobalPrefsPatch{}) != unset {
		update = append(update, bson.E{"$unset", unset})
	}

	collection := db.Database("pest-control").Collection("prefs")
	if _, err := collection.UpdateOne(context.TODO(), filter, update); err != nil {
		log.Printf(
			"failed to update preferences (%+v) in MongoDB collection: %s",
			filter,
			err.Error(),
		)
		return err
	}

	return nil
}
