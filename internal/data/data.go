package data

import (
    "fmt"
    "bytes"
    "errors"

    "github.com/Freedom-Club-Sec/Coldwire-server/internal/config"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/constants"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/storage/sqlite"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/storage"
)

type DataService struct {
    Store storage.DataStorage
    Cfg *config.Config
}

func NewDataService(cfg *config.Config) (*DataService, error) {
    var s storage.DataStorage
    switch cfg.DataStorage {
    case "internal", "sqlite":
        sqliteStore, err := sqlite.New(constants.SQLITE_DB_NAME)
        if err != nil {
            return nil, err
        }
        s = sqliteStore
    default:
        return nil, fmt.Errorf("Unknown DataStorage type (%s)", cfg.DataStorage)
    }

    return &DataService{Store: s, Cfg: cfg}, nil
}



func (svc *DataService) GetLatestData(userId string) ([]byte, error) {
    return svc.Store.GetLatestData(userId)
}

func (svc *DataService) InsertData(data []byte, senderId string, recipientId string) error {
    exists, err := svc.Store.CheckUserIdExists(recipientId)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("Recipient (%s) does not exist!", recipientId)
    }

    senderIdBytes := []byte(senderId)

    if bytes.Contains([]byte(senderIdBytes), []byte{constants.COLDWIRE_DATA_SEP}) {
        return fmt.Errorf("Sender Id (%s) has the COLDWIRE_DATA_SEP in it!", senderId)
    }


    var newDataBlob []byte

    newDataBlob = append(newDataBlob, senderIdBytes...)
    newDataBlob = append(newDataBlob, constants.COLDWIRE_DATA_SEP)
    newDataBlob = append(newDataBlob, data...)

    newDataBlob, err = PrependLengthPrefix(newDataBlob, constants.COLDWIRE_LEN_OFFSET)
    if err != nil {
        return err
    }

    return svc.Store.InsertData(newDataBlob, recipientId)
}

func PrependLengthPrefix(payload []byte, lengthBytes int) ([]byte, error) {
    if lengthBytes <= 0 || lengthBytes > 8 {
        return nil, errors.New("lengthBytes must be between 1 and 8")
    }

    length := len(payload)
    lengthPrefix := make([]byte, lengthBytes)

    // Fill the length prefix in big-endian
    for i := 0; i < lengthBytes; i++ {
        shift := uint((lengthBytes - i - 1) * 8)
        lengthPrefix[i] = byte(length >> shift)
    }

    return append(lengthPrefix, payload...), nil
}
