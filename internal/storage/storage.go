package storage

type UserStorage interface {
	SaveUser(id string, publicKey []byte) error
	CheckUserIdExists(id string) (bool, error)
	GetUserPublicKeyById(id string) ([]byte, error)
	SaveChallenge(challenge []byte, id interface{}, publicKey interface{}) error
	SaveServerInfo(url string, publicKey []byte, refetchDate string) error
	GetServerInfo(url string) ([]byte, string, error)
	GetChallengeData(challenge []byte) ([]byte, string, error)
	ExitCleanup() error
	CleanupChallenges() error
}

type DataStorage interface {
	GetLatestData(userId string) ([]byte, error)
	InsertData(data []byte, recipientId string) error
	ExitCleanup() error
}
