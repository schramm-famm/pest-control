package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Option string

type GeneralPrefs struct {
	TextEntered  Option `json:"text_entered,omitempty" bson:"text_entered,omitempty"`
	TextModified Option `json:"text_modified,omitempty" bson:"text_modified,omitempty"`
	Tag          Option `json:"tag,omitempty" bson:"tag,omitempty"`
	Role         Option `json:"role,omitempty" bson:"role,omitempty"`
}

type GlobalPrefs struct {
	Invitation    Option `json:"invitation,omitempty" bson:"invitation,omitempty"`
	*GeneralPrefs `bson:"inline"`
}

type ConversationPrefs struct {
	ConversationID int `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	*GeneralPrefs  `bson:"inline"`
}

type Preferences struct {
	ID           string               `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID       int                  `json:"user_id,omitempty" bson:"user_id"`
	Global       *GlobalPrefs         `json:"global,omitempty" bson:"global"`
	Conversation []*ConversationPrefs `json:"conversation,omitempty" bson:"conversation"`
}

const (
	None    Option = "none"
	Email   Option = "email"
	Browser Option = "browser"
	All     Option = "all"
)

var (
	ErrPrefsExists     = errors.New("user preferences already exists")
	ErrPrefsConvExists = errors.New("user preferences for conversation already exists")
	ErrPrefsDNE        = errors.New("user preferences does not exist")
	ErrPrefsConvDNE    = errors.New("user preferences for conversation does not exist")
)

func NewGlobalPrefs() *GlobalPrefs {
	return &GlobalPrefs{
		All,
		&GeneralPrefs{
			TextEntered:  All,
			TextModified: All,
			Tag:          All,
			Role:         All,
		},
	}
}

func NewConversationPrefs() *ConversationPrefs {
	return &ConversationPrefs{
		0,
		&GeneralPrefs{
			TextEntered:  All,
			TextModified: All,
			Tag:          All,
			Role:         All,
		},
	}
}

func NewPreferences() *Preferences {
	return &Preferences{
		Global:       NewGlobalPrefs(),
		Conversation: []*ConversationPrefs{},
	}
}

func (g *GeneralPrefs) String() string {
	return fmt.Sprintf("%+v", *g)
}

func (g *GlobalPrefs) String() string {
	return fmt.Sprintf("%+v", *g)
}

func (g *GlobalPrefs) UnmarshalJSON(data []byte) error {
	type Aux GlobalPrefs
	var s *Aux = (*Aux)(g)
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	if s == nil || s.GeneralPrefs == nil {
		return nil
	}

	invalidVal := []string{}

	switch s.Invitation {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "invitation")
	}

	switch s.Role {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "role")
	}

	switch s.Tag {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "tag")
	}

	switch s.TextEntered {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "text_entered")
	}

	switch s.TextModified {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "text_modified")
	}

	if len(invalidVal) > 0 {
		return errors.New(fmt.Sprintf("invalid value for %v", invalidVal))
	}

	return nil
}

func (c *ConversationPrefs) String() string {
	return fmt.Sprintf("%+v", *c)
}

func (c *ConversationPrefs) UnmarshalJSON(data []byte) error {
	type Aux ConversationPrefs
	var s *Aux = (*Aux)(c)
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	if s == nil || s.GeneralPrefs == nil {
		return nil
	}

	invalidVal := []string{}

	switch s.Role {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "role")
	}

	switch s.Tag {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "tag")
	}

	switch s.TextEntered {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "text_entered")
	}

	switch s.TextModified {
	case All, Email, Browser, None, "":
		break
	default:
		invalidVal = append(invalidVal, "text_modified")
	}

	if len(invalidVal) > 0 {
		return errors.New(fmt.Sprintf("invalid value for %v", invalidVal))
	}

	return nil
}

func (db *DB) GetPrefs(userID int) (*GlobalPrefs, error) {
	filter := bson.D{{"user_id", userID}}
	opts := options.FindOne().SetProjection(bson.D{{"global", 1}})
	collection := db.Database("pest-control").Collection("prefs")
	singleResult := collection.FindOne(context.TODO(), filter, opts)
	if singleResult.Err() != nil {
		if singleResult.Err() == mongo.ErrNoDocuments {
			return nil, ErrPrefsDNE
		}
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
		if singleResult.Err() == mongo.ErrNoDocuments {
			return nil, ErrPrefsDNE
		}
		return nil, singleResult.Err()
	}

	prefs := &Preferences{}
	if err := singleResult.Decode(prefs); err != nil {
		log.Printf("failed to decode retrieved data (%+v): %s", singleResult, err.Error())
		return nil, err
	}

	if len(prefs.Conversation) == 0 {
		return nil, ErrPrefsConvDNE
	}
	return prefs.Conversation[0], nil
}

func (db *DB) CreatePrefs(prefs *Preferences) error {
	if _, err := db.GetPrefs(prefs.UserID); err != nil && err != ErrPrefsDNE {
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

	prefs.ID = insertResult.InsertedID.(primitive.ObjectID).Hex()

	return nil
}

func (db *DB) CreatePrefsConv(userID int, convPrefs *ConversationPrefs) error {
	if _, err := db.GetPrefsConv(userID, convPrefs.ConversationID); err != nil && err != ErrPrefsConvDNE {
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
		return ErrPrefsConvExists
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
		return ErrPrefsConvDNE
	}
	return nil
}

func createUpdateBSON(prefs interface{}, prefix string) ([]byte, error) {
	bytes, err := bson.Marshal(prefs)
	if err != nil {
		log.Printf("failed to marshal prefs to bson: %s", err.Error())
		return nil, err
	}

	prefsMap := map[string]string{}
	if err = bson.Unmarshal(bytes, &prefsMap); err != nil {
		log.Printf("failed to unmarshal bson to map: %s", err.Error())
		return nil, err
	}

	// An empty JSON was provided
	if len(prefsMap) == 0 {
		return nil, nil
	}

	newPrefsMap := map[string]string{}

	for key, value := range prefsMap {
		newPrefsMap[prefix+key] = value
	}

	update := bson.D{{"$set", newPrefsMap}}
	updateBytes, err := bson.Marshal(update)
	if err != nil {
		log.Printf("failed to marshal update to bson: %s", err.Error())
		return nil, err
	}

	return updateBytes, nil
}

func (db *DB) PatchPrefs(userID int, prefs *GlobalPrefs) error {
	if _, err := db.GetPrefs(userID); err != nil {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		if err == mongo.ErrNoDocuments {
			return ErrPrefsDNE
		}
		return err
	}

	filter := bson.D{{"user_id", userID}}
	update, err := createUpdateBSON(prefs, "global.")
	if err != nil {
		log.Printf("failed to create bson for update object: %s", err.Error())
		return err
	} else if update == nil {
		return nil
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

func (db *DB) PatchPrefsConv(
	userID,
	conversationID int,
	prefs *ConversationPrefs,
) error {
	if _, err := db.GetPrefsConv(userID, conversationID); err != nil {
		log.Printf(
			"failed to get preferences from MongoDB collection: %s",
			err.Error(),
		)
		if err == mongo.ErrNoDocuments {
			return ErrPrefsConvDNE
		}
		return err
	}

	filter := bson.D{
		{"user_id", userID},
		{"conversation.conversation_id", conversationID},
	}
	update, err := createUpdateBSON(prefs, "conversation.$.")
	if err != nil {
		log.Printf("failed to create bson for update object: %s", err.Error())
		return err
	} else if update == nil {
		return nil
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
