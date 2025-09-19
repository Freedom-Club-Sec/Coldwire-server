package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	gmysql "github.com/go-sql-driver/mysql"
)

type SQLStorage struct {
	Db *sql.DB
}

type SQLDSN = gmysql.Config

func New(dsn SQLDSN) (*SQLStorage, error) {
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open sql db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(16) PRIMARY KEY,
            public_key VARBINARY(2592) NOT NULL UNIQUE
        )`,
		`CREATE TABLE IF NOT EXISTS servers (
            url VARCHAR(512) PRIMARY KEY,
            public_key VARBINARY(2592) UNIQUE NOT NULL,
            refetch_date VARCHAR(16) NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS challenges (
            challenge BINARY(64) PRIMARY KEY,
            id VARCHAR(16),
            public_key VARBINARY(2592)
        )`,
		`CREATE TABLE IF NOT EXISTS data (
            id INTEGER AUTO_INCREMENT PRIMARY KEY,
            recipient VARCHAR(529),
            data_blob MEDIUMBLOB
        )`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("failed to exec statement %q: %w", stmt, err)
		}
	}

	return &SQLStorage{Db: db}, nil
}

// Implement UserStorage interface
func (s *SQLStorage) SaveUser(id string, publicKey []byte) error {
	_, err := s.Db.Exec(`INSERT INTO users (id, public_key) VALUES (?, ?)`, id, publicKey)
	return err
}

func (s *SQLStorage) GetUserPublicKeyById(id string) ([]byte, error) {
	var publicKey []byte

	err := s.Db.QueryRow("SELECT public_key FROM users WHERE id = ?", id).Scan(&publicKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return publicKey, nil
}

func (s *SQLStorage) SaveChallenge(challenge []byte, id interface{}, publicKey interface{}) error {
	_, err := s.Db.Exec(`INSERT INTO challenges (challenge, id, public_key) VALUES (?, ?, ?)`, challenge, id, publicKey)
	return err
}

func (s *SQLStorage) SaveServerInfo(url string, publicKey []byte, refetchDate string) error {
	_, err := s.Db.Exec(`INSERT INTO servers (url, public_key, refetch_date) VALUES (?, ?, ?)`, url, publicKey, refetchDate)
	if err != nil {
		_, err = s.Db.Exec(`UPDATE servers SET public_key = ?, refetch_date = ? WHERE url = ?`, publicKey, refetchDate, url)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *SQLStorage) GetServerInfo(url string) ([]byte, string, error) {
	var (
		publicKey   []byte
		refetchDate string
	)
	err := s.Db.QueryRow("SELECT public_key, refetch_date FROM servers WHERE url = ?", url).Scan(&publicKey, &refetchDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", err
	}

	return publicKey, refetchDate, nil
}

func (s *SQLStorage) SaveCh(challenge []byte, id interface{}, publicKey interface{}) error {
	_, err := s.Db.Exec(`INSERT INTO challenges (challenge, id, public_key) VALUES (?, ?, ?)`, challenge, id, publicKey)
	return err
}

func (s *SQLStorage) GetChallengeData(challenge []byte) ([]byte, string, error) {
	var (
		publicKey []byte
		userId    sql.NullString
	)

	err := s.Db.QueryRow("SELECT id, public_key FROM challenges WHERE challenge = ?", challenge).Scan(&userId, &publicKey)
	if err != nil {
		return nil, "", err
	}

	if userId.Valid {
		fetchedPublicKey, err := s.GetUserPublicKeyById(userId.String)
		if err != nil {
			return nil, "", err
		}

		return fetchedPublicKey, userId.String, nil
	} else if publicKey != nil {
		return publicKey, "", nil
	} else {
		return nil, "", errors.New("Both userId and publicKey are null! This is a bug, if you see this message, please open an issue on Github")
	}
}

func (s *SQLStorage) CleanupChallenges() error {
	_, err := s.Db.Exec(`DELETE FROM challenges`)
	return err
}

// / Implements DataStorage interface
func (s *SQLStorage) GetLatestData(userId string) ([]byte, error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	rows, err := tx.Query("SELECT data_blob FROM data WHERE recipient = ? ORDER BY id", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allData []byte
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		allData = append(allData, data...)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	_, err = tx.Exec("DELETE FROM data WHERE recipient = ?", userId)
	if err != nil {
		return nil, err
	}

	return allData, tx.Commit()
}

func (s *SQLStorage) InsertData(dataBlob []byte, recipientId string) error {
	_, err := s.Db.Exec(`INSERT INTO data (recipient, data_blob) VALUES (?, ?)`, recipientId, dataBlob)
	return err
}

// Shared methods by UserStorage and DataStorage

func (s *SQLStorage) CheckUserIdExists(id string) (bool, error) {
	var exists bool
	row := s.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SQLStorage) ExitCleanup() error {
	return s.Db.Close()
}
