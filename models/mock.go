package models

type MockDB struct {
	Prefs     *Preferences
	ID        string
	GetErr    error
	CreateErr error
}

func (mdb *MockDB) GetPrefs(userID int) (*GlobalPrefs, error) {
	return mdb.Prefs.Global, mdb.GetErr
}

func (mdb *MockDB) GetPrefsConv(userID, conversationID int) (*ConversationPrefs, error) {
	return mdb.Prefs.Conversation[0], mdb.GetErr
}

func (mdb *MockDB) CreatePrefs(prefs Preferences) (string, error) {
	return mdb.ID, mdb.CreateErr
}
