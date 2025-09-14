package authenticate

import (
    "encoding/base64"
   
    "fmt"

    "github.com/Freedom-Club-Sec/Coldwire-server/internal/config"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/constants"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/storage/sqlite"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/storage"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
)

type UserService struct {
    Store storage.UserStorage
    Cfg *config.Config
}

func NewUserService(cfg *config.Config) (*UserService, error) {
    var s storage.UserStorage
    switch cfg.UserStorage {
    case "internal", "sqlite":
        sqliteStore, err := sqlite.New(constants.SQLITE_DB_NAME)
        if err != nil {
            return nil, err
        }
        s = sqliteStore
    default:
        return nil, fmt.Errorf("Unknown UserStorage type (%s)", cfg.UserStorage)
    }

    return &UserService{Store: s, Cfg: cfg}, nil
}

// Authentication initialization processor
func (svc *UserService) AuthenticateInitProcessor(payload *types.AuthenticateInitRequest) (string, error) {
    challengeBytes, err := utils.SecureRandomBytes(constants.CHALLENGE_LEN)
    if err != nil {
        return "", err
    }

    if payload.PublicKey != "" {
        decodedPublicKey, err := base64.StdEncoding.DecodeString(payload.PublicKey)
        if err != nil {
            return "", err
        }

        if len(decodedPublicKey) != constants.ML_DSA_87_PK_LEN {
            return "", fmt.Errorf("Public-Key length (%d) does not match ML-DSA-87 public-key standard NIST length (%d)!", len(decodedPublicKey), constants.ML_DSA_87_PK_LEN)
        }

        err = svc.Store.SaveChallenge(challengeBytes, nil, decodedPublicKey)
    } else {
        err = svc.Store.SaveChallenge(challengeBytes, payload.UserID, nil)
    }

    if err != nil {
        return "", err
    }


    challengeEncoded := base64.StdEncoding.EncodeToString(challengeBytes)

    return challengeEncoded, nil
}

// Authentication verification processor
func (svc *UserService) AuthenticateVerificationProcessor(payload *types.AuthenticateVerificationRequest) (string, []byte, bool, error) {
    decodedSignature, err := base64.StdEncoding.DecodeString(payload.Signature)
    if err != nil {
        return "", nil, false, err
    }

    if len(decodedSignature) != constants.ML_DSA_87_SIGN_LEN {
        return "", nil, false, fmt.Errorf("Signature length (%d) does not match ML-DSA-87 signature standard NIST length (%d)!", len(decodedSignature), constants.ML_DSA_87_SIGN_LEN)
    }

    decodedChallenge, err := base64.StdEncoding.DecodeString(payload.Challenge)
    if err != nil {
        return "", nil, false, err
    
    }

    if len(decodedChallenge) != constants.CHALLENGE_LEN {
        return "", nil, false, fmt.Errorf("Challenge length (%d) does not match our defined length (%d)!", len(decodedChallenge), constants.CHALLENGE_LEN)
    }

    publicKey, userId, err := svc.Store.GetChallengeData(decodedChallenge)
    if err != nil {
        return "", nil, false, err
    }

    publicKeyParsed, err := crypto.PublicKeyFromBytes(publicKey)
    if err != nil {
        return "", nil, false, err
    }

    return userId, publicKey, crypto.VerifySignature(publicKeyParsed, decodedChallenge, nil, decodedSignature), nil
}


func (svc *UserService) RegisterNewUser(publicKey []byte) (string, error) {
    var (
        userId string
        err    error
    )

    for {
        userId, err = utils.RandomUserId()
        if err != nil {
            return "", err
        }

        exists, err := svc.Store.CheckUserIdExists(userId)
        if err != nil {
            return "", err
        }
        if !exists {
            break
        }
    }
    
    err = svc.Store.SaveUser(userId, publicKey)
    return userId, err
}

