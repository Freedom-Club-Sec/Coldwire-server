package sqlite

import (
    "fmt"
    "errors"
    "database/sql"
    _ "modernc.org/sqlite"

)

type SQLiteStorage struct {
    Db *sql.DB
}

func New(path string) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("failed to open sqlite db: %w", err)
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
        `CREATE TABLE IF NOT EXISTS our_keys (
            id INTEGER PRIMARY KEY,
            private_key BLOB UNIQUE NOT NULL,
            public_key BLOB UNIQUE NOT NULL
        )`,
        `CREATE TABLE IF NOT EXISTS challenges (
            challenge BLOB PRIMARY KEY,
            id TEXT,
            public_key BLOB
        )`,
        `CREATE TABLE IF NOT EXISTS data (
            id INTEGER PRIMARY KEY,
            recipient TEXT,
            data_blob MEDIUMBLOB
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

/// Implements DataStorage interface
func (s *SQLiteStorage) GetLatestData(userId string) ([]byte, error) {
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
        if err := rows.Scan(&data); err != nil { return nil, err }
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


func (s *SQLiteStorage) InsertData(dataBlob []byte, recipientId string) error {
    _, err := s.Db.Exec(`INSERT INTO data (recipient, data_blob) VALUES (?, ?)`, recipientId, dataBlob)
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


