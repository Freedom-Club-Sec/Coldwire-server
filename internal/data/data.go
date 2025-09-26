package data

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Freedom-Club-Sec/Coldwire-server/internal/config"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/constants"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/storage"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/storage/mysql"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/storage/redis"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/storage/sqlite"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/types"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
)

type DataService struct {
	Store     storage.DataStorage
	Cfg       *config.Config
	UserStore storage.UserStorage
}

func NewDataService(cfg *config.Config, userStore storage.UserStorage) (*DataService, error) {
	var s storage.DataStorage
	switch cfg.DataStorage {
	case "internal", "sqlite":
		sqliteStore, err := sqlite.New(constants.SQLITE_DB_NAME)
		if err != nil {
			return nil, err
		}
		s = sqliteStore

	case "mysql", "sql", "mariadb":
		sqlCfg := mysql.SQLDSN{
			User:                 cfg.SQL.DBUser,
			Passwd:               cfg.SQL.DBPassword,
			Net:                  "tcp",
			Addr:                 fmt.Sprintf("%s:%d", cfg.SQL.Host, cfg.SQL.Port),
			DBName:               cfg.SQL.DBName,
			ParseTime:            false,
			AllowNativePasswords: true,
			Collation:            "utf8mb4_unicode_ci",
		}

		sqlStore, err := mysql.New(sqlCfg)
		if err != nil {
			return nil, err
		}
		s = sqlStore

	case "redis":
		portString := strconv.FormatUint(uint64(cfg.Redis.Port), 10)
		dbInt := int(cfg.Redis.DB)
		redisStore, err := redis.New(cfg.Redis.Host, portString, cfg.Redis.Password, dbInt)
		if err != nil {
			return nil, err
		}
		s = redisStore
	default:
		return nil, fmt.Errorf("Unknown DataStorage type (%s)", cfg.DataStorage)
	}

	return &DataService{Store: s, Cfg: cfg, UserStore: userStore}, nil
}

func (svc *DataService) GetLatestData(userId string) ([]byte, error) {
	return svc.Store.GetLatestData(userId)
}

func (svc *DataService) DeleteAck(userId string, acks []string) error {
	var err error
	args := make([][]byte, len(acks))
	for i, v := range acks {
		args[i], err = base64.RawURLEncoding.DecodeString(v)
		if err != nil {
			return err
		}
	}

	return svc.Store.DeleteAck(userId, args)

}

func (svc *DataService) InsertData(data []byte, senderId string, recipientId string) error {
	if utils.IsAllDigits(recipientId) {
		if len(recipientId) != 16 {
			return errors.New("Recipient is of invalid length")
		}

		exists, err := svc.UserStore.CheckUserIdExists(recipientId)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Recipient (%s) does not exist!", recipientId)
		}

		senderIdBytes := []byte(senderId)

		if bytes.Contains(senderIdBytes, []byte{constants.COLDWIRE_DATA_SEP}) {
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

		ackId, err := utils.SecureRandomBytes(32)
		if err != nil {
			return err
		}

		return svc.Store.InsertData(newDataBlob, ackId, recipientId)

		// Max DNS length is 253, 16 for recipient user ID, and 1 for `@`
	} else if len(recipientId) > 253+16+1 || len(recipientId) <= 17 {
		return errors.New("Invalid recipient ID or address")

	} else {
		if !svc.Cfg.FederationEnabled {
			return errors.New("Federation support is disabled on this server.")
		}

		recipientSplit := strings.SplitN(recipientId, "@", 2)
		if len(recipientSplit) != 2 {
			return errors.New("Invalid recipient format")
		}

		if !utils.IsAllDigits(recipientSplit[0]) || len(recipientSplit[0]) != 16 {
			return errors.New("Invalid recipient ID")
		}

		url := strings.TrimSpace(recipientSplit[1])
		url = strings.ToLower(url)

		if url == svc.Cfg.DomainOrIP {
			// If user sends to a recipient with same address as our server, we simply remove the address and treat it as normal data insert.
			return svc.InsertData(data, senderId, recipientSplit[0])

		} else {
			ourPrivateKeyCasted, err := crypto.PrivateKeyFromBytes(svc.Cfg.DSAPrivateKey)
			if err != nil {
				return err
			}

			dataToSign := []byte(url + recipientSplit[0] + senderId)
			dataToSign = append(dataToSign, data...)

			signature, err := crypto.CreateSignature(ourPrivateKeyCasted, dataToSign, nil)
			if err != nil {
				return err
			}

			metadataToSend := types.FederationSendRequest{
				Sender:    senderId,
				Recipient: recipientSplit[0],
				Url:       svc.Cfg.DomainOrIP,
			}

			blobToSend := append(signature, data...)

			err = sendToServer("https://"+url, metadataToSend, blobToSend)
			if err != nil {
				err = sendToServer("http://"+url, metadataToSend, blobToSend)
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}

func sendToServer(url string, metadata types.FederationSendRequest, blob []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	jsonPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return err
	}
	jsonPart.Write(jsonBytes)

	part, err := writer.CreateFormFile("blob", "blob.bin")
	if err != nil {
		return err
	}
	part.Write(blob)

	writer.Close()

	req, err := http.NewRequest("POST", url+"/federation/send", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error %s %d %s", url, resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (svc *DataService) FederationProcessor(senderId string, recipientId string, url string, data_blob []byte) error {
	if len(data_blob) <= constants.ML_DSA_87_SIGN_LEN {
		return errors.New("Malformed signature and blob")
	}

	exists, err := svc.UserStore.CheckUserIdExists(recipientId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Recipient (%s) does not exist!", recipientId)
	}

	publicKey, refetchDate, err := svc.GetServerInfo(url)
	if err != nil {
		return err
	}

	if publicKey == nil {
		publicKey, refetchDate, err = svc.FetchAndSaveServerInfo(url)
		if err != nil {
			return err
		}
	}

	refetchUTC, err := time.Parse("2006-01-02", refetchDate)
	if err != nil {
		return err
	}

	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)

	// Refetch keys if we are past the refetch date
	if !todayUTC.Before(refetchUTC) {
		publicKey, refetchDate, err = svc.FetchAndSaveServerInfo(url)
		if err != nil {
			return err
		}
	}

	signature := data_blob[:constants.ML_DSA_87_SIGN_LEN]
	blob := data_blob[constants.ML_DSA_87_SIGN_LEN:]

	signatureData := []byte(svc.Cfg.DomainOrIP + recipientId + senderId)
	signatureData = append(signatureData, blob...)

	isValidSignature := crypto.VerifySignature(publicKey, signatureData, nil, signature)
	if !isValidSignature {
		return fmt.Errorf("Invalid signature, while processing federation request.")
	}

	senderIdBytes := []byte(senderId + "@" + url)

	if bytes.Contains(senderIdBytes, []byte{constants.COLDWIRE_DATA_SEP}) {
		return fmt.Errorf("Sender Id (%s) has the COLDWIRE_DATA_SEP in it!", senderId)
	}

	var newDataBlob []byte

	newDataBlob = append(newDataBlob, senderIdBytes...)
	newDataBlob = append(newDataBlob, constants.COLDWIRE_DATA_SEP)
	newDataBlob = append(newDataBlob, blob...)

	newDataBlob, err = PrependLengthPrefix(newDataBlob, constants.COLDWIRE_LEN_OFFSET)
	if err != nil {
		return err
	}

	ackId, err := utils.SecureRandomBytes(32)
	if err != nil {
		return err
	}
	return svc.Store.InsertData(newDataBlob, ackId, recipientId)
}

func (svc *DataService) FetchAndSaveServerInfo(url string) (*mldsa87.PublicKey, string, error) {
	resp, err := http.Get("https://" + url + "/federation/info")
	if err != nil {
		resp, err = http.Get("http://" + url + "/federation/info")
		if err != nil {
			return nil, "", err
		}
	}
	defer resp.Body.Close()

	var result types.FederationInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", err
	}

	if len(result.PublicKey) != constants.ML_DSA_87_PK_LEN {
		return nil, "", fmt.Errorf("PublicKey has invalid length (%d), we expected %d", len(result.PublicKey), constants.ML_DSA_87_PK_LEN)
	}

	if len(result.Signature) != constants.ML_DSA_87_SIGN_LEN {
		return nil, "", fmt.Errorf("Signature has invalid length (%d), we expected %d", len(result.Signature), constants.ML_DSA_87_SIGN_LEN)
	}

	signatureData := []byte(url + result.RefetchDate)
	publicKeyCasted, err := crypto.PublicKeyFromBytes(result.PublicKey)
	if err != nil {
		return nil, "", err
	}

	isValidSignature := crypto.VerifySignature(publicKeyCasted, signatureData, nil, result.Signature)
	if !isValidSignature {
		return nil, "", fmt.Errorf("Invalid signature, while fetching for server (%s) info", url)
	}

	err = svc.UserStore.SaveServerInfo(url, result.PublicKey, result.RefetchDate)
	if err != nil {
		return nil, "", err
	}

	return publicKeyCasted, result.RefetchDate, nil
}

func (svc *DataService) GetServerInfo(url string) (*mldsa87.PublicKey, string, error) {
	publicKey, refetchDate, err := svc.UserStore.GetServerInfo(url)
	if err != nil {
		return nil, "", err
	}

	if publicKey == nil {
		return nil, "", nil
	}

	publicKeyCasted, err := crypto.PublicKeyFromBytes(publicKey)
	if err != nil {
		return nil, "", err
	}

	return publicKeyCasted, refetchDate, nil
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
