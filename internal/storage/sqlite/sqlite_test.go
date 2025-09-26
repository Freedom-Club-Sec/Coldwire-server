package sqlite

import (
	"bytes"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
	"testing"
)

func TestNewUserAndPublicKeyRetrieve(t *testing.T) {
	store, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Db.Ping(); err != nil {
		t.Fatal(err)
	}

	userId, err := utils.RandomUserId()
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := utils.SecureRandomBytes(2592)
	if err != nil {
		t.Fatal(err)
	}

	err = store.SaveUser(userId, publicKey)
	if err != nil {
		t.Fatal(err)
	}

	fetchedPublicKey, err := store.GetUserPublicKeyById(userId)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(fetchedPublicKey, publicKey) {
		t.Fatalf("fetchedPublicKey does not equal publicKey. fetchedPublicKey: %v \n\n\n publicKey: %v \n\n", fetchedPublicKey, publicKey)
	}

	// Invalid ID
	nilPublicKey, err := store.GetUserPublicKeyById("0")
	if err != nil {
		t.Fatal(err)
	}

	if nilPublicKey != nil {
		t.Fatalf("nilPublicKey is not nil: %v", nilPublicKey)
	}
}
