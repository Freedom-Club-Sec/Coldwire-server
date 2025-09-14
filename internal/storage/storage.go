package storage

type UserStorage interface {
    SaveUser(id string, publicKey []byte) error
    CheckUserIdExists(id string) (bool, error)
    GetUserPublicKeyById(id string) ([]byte, error)
    SaveChallenge(challenge []byte, id interface{}, publicKey interface{}) error
    GetChallengeData(challenge []byte) ([]byte, string, error) 
    ExitCleanup() error
    CleanupChallenges() error

}

type DataStorage interface {
    GetLatestData(recipientId string) ([]byte, error)
    InsertData(data []byte, recipientId string) error
    CheckUserIdExists(id string) (bool, error)
    ExitCleanup() error
}
