package sqlite

import (
	"database/sql"
    "log/slog"
	"errors"
    "strings"
	"fmt"
    isqlite "modernc.org/sqlite"
    isqlitelib "modernc.org/sqlite/lib"

)

type SQLiteStorage struct {
	Db *sql.DB
}

func New(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open sqlite db: %w", err)
	}

    
    if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
        return nil, err
    }

    if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
        return nil, err
    }


	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            public_key BLOB NOT NULL UNIQUE
        )`,
		`CREATE TABLE IF NOT EXISTS servers (
            url TEXT PRIMARY KEY,
            public_key BLOB UNIQUE NOT NULL,
            refetch_date TEXT NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS challenges (
            challenge BLOB PRIMARY KEY,
            id TEXT,
            public_key BLOB
        )`,
		`CREATE TABLE IF NOT EXISTS data (
            id INTEGER PRIMARY KEY,
            ack_id BLOB NOT NULL,
            recipient TEXT NOT NULL,
            data_blob MEDIUMBLOB NOT NULL
        )`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("failed to exec statement %q: %w", stmt, err)
		}
	}

	return &SQLiteStorage{Db: db}, nil
}

// Implement UserStorage interface
func (s *SQLiteStorage) SaveUser(id string, publicKey []byte) error {
	_, err := s.Db.Exec(`INSERT INTO users (id, public_key) VALUES (?, ?)`, id, publicKey)
	return err
}

func (s *SQLiteStorage) GetUserPublicKeyById(id string) ([]byte, error) {
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

func (s *SQLiteStorage) SaveChallenge(challenge []byte, id interface{}, publicKey interface{}) error {
	_, err := s.Db.Exec(`INSERT INTO challenges (challenge, id, public_key) VALUES (?, ?, ?)`, challenge, id, publicKey)
	return err
}

func (s *SQLiteStorage) SaveServerInfo(url string, publicKey []byte, refetchDate string) error {
	_, err := s.Db.Exec(`INSERT INTO servers (url, public_key, refetch_date) VALUES (?, ?, ?)`, url, publicKey, refetchDate)
	if err != nil {
		_, err = s.Db.Exec(`UPDATE servers SET public_key = ?, refetch_date = ? WHERE url = ?`, publicKey, refetchDate, url)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *SQLiteStorage) GetServerInfo(url string) ([]byte, string, error) {
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

func (s *SQLiteStorage) GetChallengeData(challenge []byte) ([]byte, string, error) {
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

func (s *SQLiteStorage) CleanupChallenges() error {
	_, err := s.Db.Exec(`DELETE FROM challenges`)
	return err
}


func isSQLiteBusy(err error) bool {
    var se *isqlite.Error
    if errors.As(err, &se) {
        if se.Code() == isqlitelib.SQLITE_BUSY {
            slog.Debug("SQLite database is locked.", "error", err)
            return true
        }
    }
    return false
}

// / Implements DataStorage interface
func (s *SQLiteStorage) GetLatestData(userId string) ([]byte, error) {
	rows, err := s.Db.Query("SELECT data_blob, ack_id FROM data WHERE recipient = ? ORDER BY id", userId)
	if err != nil {
        if isSQLiteBusy(err) {
            return nil, nil
        }
		return nil, err
	}
	defer rows.Close()

	var allData []byte
	for rows.Next() {
		var (
            data  []byte
            ackId []byte
        )

		if err := rows.Scan(&data, &ackId); err != nil {
			return nil, err
		}

        data = append(ackId, data...)
		allData = append(allData, data...)
	}

	if err := rows.Err(); err != nil {
        if isSQLiteBusy(err) {
            return nil, nil
        }
		return nil, err
	}


	return allData, nil
}


func (s *SQLiteStorage) DeleteAck(userId string, acks [][]byte) error {
    placeholders := make([]string, len(acks))
    args := make([]interface{}, len(acks))
    for i, v := range acks {
        placeholders[i] = "?"
        args[i] = v
    }

    args = append([]any{userId}, args...)

    var err error

    for {
        query := fmt.Sprintf("DELETE FROM data WHERE recipient = ? AND ack_id IN (%s)", strings.Join(placeholders, ","))
        _, err = s.Db.Exec(query, args...)
        if isSQLiteBusy(err) {
            continue
        }
        break
    }
    return err
}

func (s *SQLiteStorage) InsertData(dataBlob []byte, ackId []byte, recipientId string) error {
    var err error
    for {
        _, err = s.Db.Exec(`INSERT INTO data (recipient, ack_id, data_blob) VALUES (?, ?, ?)`, recipientId, ackId, dataBlob)
        if isSQLiteBusy(err) {
            continue
        }
        break
    }
	return err
}

// Shared methods by UserStorage and DataStorage

func (s *SQLiteStorage) CheckUserIdExists(id string) (bool, error) {
	var exists bool
	row := s.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SQLiteStorage) ExitCleanup() error {
	return s.Db.Close()
}
