package models

type MockDB struct {
	Prefs     *Preferences
	GetErr    error
	CreateErr error
	DeleteErr error
	PatchErr  error
}

func (mdb *MockDB) GetPrefs(userID int) (*GlobalPrefs, error) {
	return mdb.Prefs.Global, mdb.GetErr
}

func (mdb *MockDB) GetPrefsConv(userID, conversationID int) (*ConversationPrefs, error) {
	return mdb.Prefs.Conversation[0], mdb.GetErr
}

func (mdb *MockDB) CreatePrefs(prefs *Preferences) error {
	prefs.ID = mdb.Prefs.ID
	return mdb.CreateErr
}

func (mdb *MockDB) CreatePrefsConv(userID int, convPrefs *ConversationPrefs) error {
	return mdb.CreateErr
}

func (mdb *MockDB) DeletePrefs(userID int) error {
	return mdb.DeleteErr
}

func (mdb *MockDB) DeletePrefsConv(userID, conversationID int) error {
	return mdb.DeleteErr
}

func (mdb *MockDB) PatchPrefs(userID int, prefs *GlobalPrefs) error {
	return mdb.PatchErr
}

func (mdb *MockDB) PatchPrefsConv(
	userID,
	conversationID int,
	prefs *ConversationPrefs,
) error {
	return mdb.PatchErr
}
