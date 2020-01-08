package models

import "fmt"

type GlobalPrefs struct {
	Invitation   bool `json:"invitation,omitempty"`
	TextEntered  bool `json:"text_entered,omitempty"`
	TextModified bool `json:"text_modified,omitempty"`
	Tag          bool `json:"tag,omitempty"`
	Role         bool `json:"role,omitempty"`
}

type ConversationPrefs struct {
	ConversationID string `json:"conversation_id,omitempty"`
	TextEntered    bool   `json:"text_entered,omitempty"`
	TextModified   bool   `json:"text_modified,omitempty"`
	Tag            bool   `json:"tag,omitempty"`
	Role           bool   `json:"role,omitempty"`
}

type Preferences struct {
	UserID       string               `json:"user_id,omitempty"`
	Global       *GlobalPrefs         `json:"global,omitempty"`
	Conversation []*ConversationPrefs `json:"conversation,omitempty"`
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
