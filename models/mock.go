package models

type MockDB struct {
	Prefs     *Preferences
	ID        string
	GetErr    error
	CreateErr error
}

func (mdb *MockDB) GetPrefs(userID, conversationID int) (*Preferences, error) {
	return mdb.Prefs, mdb.GetErr
}

func (mdb *MockDB) CreatePrefs(prefs Preferences) (string, error) {
	return mdb.ID, mdb.CreateErr
}
